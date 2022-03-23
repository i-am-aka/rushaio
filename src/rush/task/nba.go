package task

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	// "os"
	// "github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	// "log"
	"regexp"
	"time"
	"math/rand"
	"net/url"
	"strings"
	"github.com/avast/retry-go"
	// "os"
)

type NBASize struct {
	ItemId string `json:"itemId"`
	Size string `json:"size"`
	Available bool `json:"available"`
}
type NBASpecificProductData struct {
	Sizes []NBASize `json:"sizes"`
	Title string `json:"title"`
	ProductId string `json:"productId"`
}
type NBAProductData struct {
	SpecificProductData NBASpecificProductData `json:"pdp"`
}
type NBAPlatformData struct {
	ProductData NBAProductData `json:"pdp-data"`
}

type NBABalanceAmount struct {
	BaseAmount float64 `json:"baseCurrencyValue"`
}

type NBAPaymentInfo struct {
	BalanceAmount NBABalanceAmount `json:"balanceAmount"`
}
type NBAChargeSummary struct {
	Value float64 `json:"userCurrencyValue"`
}
type NBACartSummary struct {
	GrandTotal NBAChargeSummary `json:"grandTotal"`
}
type NBACart struct {
	Id string `json:"cartId"`
	Summary NBACartSummary `json:"summary"`
	PaymentInfo NBAPaymentInfo `json:"paymentInfo"`
}

type NBACartData struct {
	Cart NBACart `json:"cart"`
}

type NBACheckoutSession struct {
	Data NBACartData `json:"data"`
}

type NBAOrderResultMessage struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Code int `json:"code"`
}
type NBAOrderResultHeader struct {
	Status string `json:"status"`
	Messages []NBAOrderResultMessage `json:"messages"`
}
type NBAOrderResult struct {
	Header NBAOrderResultHeader `json:"header"`
}
func FindNBAProductUrlByKeywords(pageBody string, keywords []string) *url.URL {
	r := regexp.MustCompile(`<h4 class="product-card-title">(.+?)<\/h4>`)
	matches := r.FindAllString(pageBody, -1)
	// log.Printf("%d", len(matches))
	for _, match := range matches {
		// log.Printf("match")
		z := html.NewTokenizer(strings.NewReader(match))
		done := false
		for !done {
	    tt := z.Next()
	    switch tt {
	    case html.ErrorToken:
        done = true
	    case html.StartTagToken:
        t := z.Token()
        switch t.Data {
        case "a":
        	href := ""
        	for _, a := range t.Attr {
      			if a.Key == "href" {
      				href = "https://store.nba.com" + a.Val
      				break
      			}
      		}
      		if href == "" {
      			continue
      		}
          z.Next()
          t = z.Token()
          aTitleTokens := strings.Split(t.Data, " ")
          allKwsMatch := All(keywords, func(kw string) bool {
          	return Any(aTitleTokens, func(tok string) bool {
          		// log.Printf("%s %s", kw, tok)
          		return strings.ToLower(kw) == strings.ToLower(tok)
          	})
          })
          if allKwsMatch {
          	productUrl, err := url.Parse(href)
          	if err != nil {
          		// log.Printf("%+v", err)
          		return nil
          	} else {
          		return productUrl
          	}
          }
        }
	    }
	  }
	}
	return nil
}

func (t *CheckoutTask) NbaCheckout() error {
	var err error
	t.UserAgent, err = t.getUserAgent()
  if err != nil {
    return err
  }

	return t.RetryTask(func() error {
		t.initClient()

		var productUrl *url.URL
		if len(t.Keywords) > 0 {
			// log.Printf("%+v %d", t.Keywords, len(t.Keywords))
			t.StatusCh <- Status{Message: "Searching for Keywords"}
			for {
				resp, err := t.doReq(t.client, t.makeGetReq(t.Url, nil))
				if err != nil {
					return err
				}
				body, err := readBody(resp)
				if err != nil {
					return err
				}
				productUrl = FindNBAProductUrlByKeywords(body, t.Keywords)
				if productUrl != nil {
					break
				}
				time.Sleep(time.Duration(1)*time.Second)
				// productUrl = kwUrl
			}
		} else {
			productUrl = t.Url
		}
		t.StatusCh <- Status{Message: "Waiting for Product"}
		resp, err := t.doReq(t.client, t.makeGetReq(productUrl, nil))
		if err != nil {
			return err
		}
		prodBody, err := readBody(resp)
		if err != nil {
			return err
		}

		atcHeaders := [][2]string{
			{"content-type", "application/json"},
		  {"x-xsrf-token", getCookieValue("xsrft", productUrl, t.client.Jar)},
		  {"referer", productUrl.String()},
		}

		pdss := strings.SplitN(prodBody, "<script>var __platform_data__=", 2)
		if len(pdss) < 2 {
			return ErrRetrying
		}
		pdPlusExtra := pdss[1]
		pdStr := strings.SplitN(pdPlusExtra, "</script>", 2)[0]
		var pd NBAPlatformData
		err = json.Unmarshal([]byte(pdStr), &pd)
		if err != nil {
			return err
		}

		sizes := pd.ProductData.SpecificProductData.Sizes
		var selectedSize *NBASize = nil

		rand.Shuffle(len(sizes), func(i, j int) {
		  sizes[i], sizes[j] = sizes[j], sizes[i] })

		for _, nbaSize := range sizes {
			if nbaSize.Available && (len(t.Sizes) == 0 || Any(t.Sizes, func(size string) bool { return nbaSize.Size == size })) {
				selectedSize = &nbaSize
				break
			}
		}

		if selectedSize == nil {
			return errors.New("Out of Stock")
		}

		atcBodyMap := map[string]interface{}{
	    "customProduct": nil,
	    "enableCustomInput": false,
	    "isPreconfigured": false,
	    "itemId": selectedSize.ItemId,
	    "productId": pd.ProductData.SpecificProductData.ProductId,
	    "quantity": 1,
		}

		atcBody, err := json.Marshal(atcBodyMap)
		if err != nil {
			return err
		}
		atcUrl, err := url.Parse("https://store.nba.com/api/experience")
		if err != nil {
			return err
		}
		t.StatusCh <- Status{Message: "Adding to Cart"}
		atcResp, err := t.doReq(t.client, t.makeReq("POST", atcUrl, &atcHeaders, nil, &atcBody))
		if err != nil {
			if atcResp != nil && atcResp.StatusCode == 410 {
				return errors.New(fmt.Sprintf("OOS (selectedSize %v productId %s)", selectedSize, pd.ProductData.SpecificProductData.ProductId))
			} else {
				return err
			}
		}
		atcRespBodyBytes, err := readBodyBytes(atcResp)
		if err != nil {
			return err
		}
		var checkoutSesh NBACheckoutSession
		err = json.Unmarshal(atcRespBodyBytes, &checkoutSesh)
		if err != nil {
			return err
		}
		t.StartCheckoutTime = timeMillis()
		shipUrl, err := url.Parse("https://store.nba.com/api/experience/address-payment")
		if err != nil {
			return err
		}

		shipHeaders := [][2]string{
			{"content-type", "application/json"},
			{"x-frg-pt", "SHPG"},
		  {"x-xsrf-token", getCookieValue("xsrft", productUrl, t.client.Jar)},
		  {"referer", "https://store.nba.com/shipping"},
		}
		shipBodyMap := map[string]interface{} {
	    "cartId": checkoutSesh.Data.Cart.Id,
	    "email": t.Profile.Email,
	    "shippingAddress": map[string]interface{} {
	        "addressLine1": t.Profile.ShippingAddress.Address1,
	        "addressLine2": t.Profile.ShippingAddress.Address2,
	        "city": t.Profile.ShippingAddress.City,
	        "country": t.Profile.ShippingAddress.Country,
	        "email": t.Profile.Email,
	        "firstName": t.Profile.ShippingAddress.FirstName,
	        "isAutoComplete": false,
	        "lastName": t.Profile.ShippingAddress.LastName,
	        "phone": t.Profile.ShippingAddress.Phone,
	        "phoneNumber": t.Profile.ShippingAddress.Phone,
	        "postalCode": t.Profile.ShippingAddress.Zip,
	        "state": t.Profile.ShippingAddress.State,
	    },
	    "subscriptionPreferences": []interface{}{},
		}
		shipBody, err := json.Marshal(shipBodyMap)
		if err != nil {
			return err
		}
		t.StatusCh <- Status{Message: "Shipping Info"}
		shipResp, err := t.doReq(t.client, t.makeReq("POST", shipUrl, &shipHeaders, nil, &shipBody))
		if err != nil {
			return err
		}
		shipRespBytes, err := readBodyBytes(shipResp)
		if err != nil {
			return err
		}
		err = json.Unmarshal(shipRespBytes, &checkoutSesh)
		if err != nil {
			return err
		}

		// checkoutUrl, err := url.Parse("https://store.nba.com/api/cart/checkout")
		// if err != nil {
		// 	return err
		// }
		// t.StatusCh <- Status{Message: "Starting Checkout"}
		// checkoutResp, err := t.doReq(t.client, t.makeReq("POST", checkoutUrl, &shipHeaders, nil, &[]byte{}))
		// if err != nil {
		// 	return err
		// }
		// checkoutBodyBytes, err := readBodyBytes(checkoutResp)
		// if err != nil {
		// 	return err
		// }
		// err = json.Unmarshal(checkoutBodyBytes, &checkoutSesh)
		// if err != nil {
		// 	return err
		// }

		payKeyHeaders := [][2]string {
			{"x-xsrf-token", getCookieValue("xsrft", productUrl, t.client.Jar)},
			{"referer", "https://store.nba.com/payment"},
		}
		payKeyUrl, err := url.Parse("https://store.nba.com/api/payment-gateway/keys")
		t.StatusCh <- Status{Message: "Encrypting Card (1)"}
		payKeyResp, err := t.doReq(t.client, t.makeReq("POST", payKeyUrl, &payKeyHeaders, nil, &[]byte{}))
		payKeyBytes, err := readBodyBytes(payKeyResp)
		if err != nil {
			return err
		}
		var payKeys CSPayKeys
		err = json.Unmarshal(payKeyBytes, &payKeys)
		if err != nil {
			return err
		}
		t.StatusCh <- Status{Message: "Encrypting Card (2)"}
		csToken, err := t.GetCsToken(payKeys, t.Profile.Card, "https://store.nba.com")
		if err != nil {
			return err
		}
		t.StartPaymentTime = timeMillis()
		orderUrl, err := url.Parse("https://store.nba.com/api/experience/create-order")
		if err != nil {
			return err
		}
		orderHeaders := [][2]string {
			{"content-type", "application/json"},
			{"x-pmt-type", "credit_card"},
			{"x-xsrf-token", getCookieValue("xsrft", productUrl, t.client.Jar)},
			{"referer", "https://store.nba.com/payment"},
		}
		orderBodyMap := map[string]interface{} {
	    "cartId": checkoutSesh.Data.Cart.Id,
	    "cartPaymentInstruments": []map[string]interface{} {
	        map[string]interface{} {
	            "amount": fmt.Sprintf("%.2f", checkoutSesh.Data.Cart.Summary.GrandTotal.Value),
	            "paymentAction": "add",
	            "paymentInstrument": map[string]interface{} {
	                "creditCard": map[string]interface{} {
	                    "tokenInfo": map[string]interface{} {
	                        "address": map[string]string {
	                            "addressLine1": t.Profile.BillingAddress.Address1,
	                            "addressLine2": t.Profile.BillingAddress.Address2,
	                            "city": t.Profile.BillingAddress.City,
	                            "country": t.Profile.BillingAddress.Country,
	                            "email": t.Profile.Email,
	                            "firstName": t.Profile.BillingAddress.FirstName,
	                            "lastName": t.Profile.BillingAddress.LastName,
	                            "postalCode": t.Profile.BillingAddress.Zip,
	                            "state": t.Profile.BillingAddress.State,
	                            "type": "billing",
	                        },
	                        "cvv": t.Profile.Card.Cvc,
	                        "expirationDate": fmt.Sprintf("%02d/%d", t.Profile.Card.ExpMonth, t.Profile.Card.ExpYear),
	                        "isExplicitSave": false,
	                        "isPrimary": false,
	                        "issuingBankBIN": csToken.IssuingBankBin(),
	                        "lastFourDigits": t.Profile.Card.LastFour,
	                        "paymentCardType": csToken.CardType,
	                        "token": csToken.Token,
	                        "tokenProvider": "CyberSource",
	                    },
	                },
	                "paymentInstrumentType": "credit_card",
	            },
	        },
	    },
		}
		orderBody, err := json.Marshal(orderBodyMap)
		if err != nil {
			return err
		}
		t.StatusCh <- Status{Message: "Finalizing Order"}
		orderResp, err := t.doReq(t.client, t.makeReq("POST", orderUrl, &orderHeaders, nil, &orderBody))
		if err != nil && (orderResp == nil || orderResp.StatusCode != 400) {
			bb, _ := readBody(orderResp)
			t.LogDebug("%s %+v", bb, checkoutSesh)
			return err
		}
		orderRespBody, err := readBodyBytes(orderResp)
		if err != nil {
			return err
		}
		var orderResult NBAOrderResult
		err = json.Unmarshal(orderRespBody, &orderResult)
		if err != nil {
			return err
		}
		if orderResult.Header.Status == "SUCCESS" {
			t.StatusCh <- Status{Message: "COOKED!!!", Level: 9999} // , "size", cartAdd.Size, "orderJson", paymentRespJson}
			t.SuccessWebhook("NBA Store", pd.ProductData.SpecificProductData.Title, selectedSize.Size)
			return nil
		} else if orderResult.Header.Status == "FAILURE" {
			t.StatusCh <- Status{Message: fmt.Sprintf("Payment failed: %v", orderResult.Header.Messages)}
			t.DeclineWebhook("NBA Store", pd.ProductData.SpecificProductData.Title, selectedSize.Size)
			return nil
		} else {
			// log.Printf("%s", string(orderRespBody))
			return errors.New(fmt.Sprintf("Unknown order status: %s", orderResult.Header.Status))
		}
		return nil
	}, retry.Attempts(1e6), retry.LastErrorOnly(true))
}

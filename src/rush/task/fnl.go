package task

import (
	// "strconv"
	// "os"
	"bytes"
	"io"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"github.com/avast/retry-go"
	"time"
	"rush/net/http"
	"strings"
	"github.com/pkg/errors"
)

var (
	MAX_RETRIES uint = 1024
	FNL_APP_API_HOST = "https://prodmobloy2.finishline.com"
	JD_APP_API_HOST = "https://prodmobloy2.jdsports.com"
	FNL_APP_API_HOST_URL, _ = url.Parse(FNL_APP_API_HOST)
	JD_APP_API_HOST_URL, _ = url.Parse(JD_APP_API_HOST)
)

func (t *CheckoutTask) fnlGetApiHost() string {
	if t.Url.Host == "www.finishline.com" {
		return FNL_APP_API_HOST
	} else if t.Url.Host == "www.jdsports.com" {
		return JD_APP_API_HOST
	} else {
		t.LogDebug("WARNING: UNKNOWN FNL HOST %s", t.Url.Host)
		return FNL_APP_API_HOST
	}
}

func (t *CheckoutTask) fnlGetApiHostUrl() *url.URL {
	if t.Url.Host == "www.finishline.com" {
		return FNL_APP_API_HOST_URL
	} else if t.Url.Host == "www.jdsports.com" {
		return JD_APP_API_HOST_URL
	} else {
		t.LogDebug("WARNING: UNKNOWN FNL HOST %s", t.Url.Host)
		return FNL_APP_API_HOST_URL
	}
}

func (t *CheckoutTask) fnlGetConfig() error {
  url_, err := url.Parse(fmt.Sprintf("%s/api/config", t.fnlGetApiHost()))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "Host",
    "User-Agent",
    "Accept-Encoding",
    "Connection",
    "Accept",
    "X-Api-Version",
    "welove",
    "Content-Type",
    "Content-Length",
    "RiskifiedID",
    "Riskified-User-Agent",
  }
  headers := [][2]string {
    {"Connection", "keep-alive"},
    {"Accept", "*/*"},
    {"X-Api-Version", "3.0"},
    {"welove", "maltliquor"},
    {"Content-Type", "application/json; charset=utf-8"},
    {"Content-Length", "0"},
    {"accept-encoding", "gzip, deflate"},
  }
  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
  if resp != nil {
  	io.Copy(ioutil.Discard, resp.Body)
  	resp.Body.Close()
  	if resp.StatusCode == 403 {
  		t.StatusCh <- StatusAkaBanned
  		return retry.Unrecoverable(ErrAkamaiBanned)
  	}
  }
  return err
}

func (t *CheckoutTask) fnlGetAccountGuest() (error) {
  url_, err :=  url.Parse(fmt.Sprintf("%s/api/account/guest", t.fnlGetApiHost()))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "Host",
    "User-Agent",
    "Accept-Encoding",
    "Connection",
    "Accept",
    "X-acf-sensor-data",
    "X-Api-Version",
    "welove",
    "Content-Type",
    "Content-Length",
    "RiskifiedID",
    "Riskified-User-Agent",
    "Cookie",
  }
  headers := [][2]string {
    {"Connection", "keep-alive"},
    {"Accept", "*/*"},
    {"X-Api-Version", "3.0"},
    {"X-acf-sensor-data", t.Config["x-acf-sensor-data"]},
    {"welove", "maltliquor"},
    {"Content-Type", "application/json; charset=utf-8"},
    {"Content-Length", "0"},
    {"accept-encoding", "gzip, deflate"},
  }
  resp, err := t.FnlDoQueueAwareReq(t.makeReq("GET", url_, &headers, &headerOrder, nil), "Starting Session")
  if resp != nil {
  	io.Copy(ioutil.Discard, resp.Body)
  	resp.Body.Close()
  }
  if err != nil {
  	return err
  }
  if resp.Header.Get("X-Session") != "" {
  	t.Config["x-session"] = resp.Header.Get("X-Session")
  }
  if resp.Header.Get("X-Cookie") != "" {
  	t.Config["x-cookie"] = resp.Header.Get("X-Cookie")
  }
  if resp.Header.Get("AkamaiCookie") != "" {
  	t.Config["AkamaiCookie"] = resp.Header.Get("AkamaiCookie")
  }
  return err
}

func (t *CheckoutTask) fnlGetDrops() error {
	url_, err := url.Parse(fmt.Sprintf("%s/api/shop/drops", t.fnlGetApiHost()))
	if err != nil {
	  return err
	}
	headerOrder := []string {
	  "Host",
	  "User-Agent",
	  "Accept-Encoding",
	  "Connection",
	  "Accept",
	  "X-acf-sensor-data",
	  "X-Api-Version",
	  "x-cookie",
	  "welove",
	  "Content-Type",
	  "x-session",
	  "Content-Length",
	  "RiskifiedID",
	  "Riskified-User-Agent",
	  "AkamaiCookie",
	  "Cookie",
	}
	headers := [][2]string {
	  {"Connection", "keep-alive"},
	  {"Accept", "*/*"},
	  {"X-Api-Version", "3.0"},
	  {"welove", "maltliquor"},
	  {"accept-encoding", "gzip, deflate"},
	  {"Content-Type", "application/json; charset=utf-8"},
	  {"Content-Length", "0"},
	}
	headers = append(headers, t.fnlGetHeaders()...)

	resp, err := t.FnlDoQueueAwareReq(t.makeReq("GET", url_, &headers, &headerOrder, nil), "Waiting for Product")
	if err != nil {
		return err
	}
	body, err := readBodyBytes(resp)
	if err != nil {
		return err
	}
	t.LogDebug(string(body))
	return nil
}


func (t *CheckoutTask) fnlGetProduct(pid string) (FnlProduct, error) {
	var product FnlProduct
	err := t.Retry(func() error {
	  url_, err := url.Parse(fmt.Sprintf("%s/api/products/%s", t.fnlGetApiHost(), pid))
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "Host",
	    "User-Agent",
	    "Accept-Encoding",
	    "Connection",
	    "Accept",
	    "X-acf-sensor-data",
	    "X-Api-Version",
	    "x-cookie",
	    "welove",
	    "Content-Type",
	    "x-session",
	    "Content-Length",
	    "RiskifiedID",
	    "Riskified-User-Agent",
	    "AkamaiCookie",
	    "Cookie",
	  }
	  headers := [][2]string {
	    {"Connection", "keep-alive"},
	    {"Accept", "*/*"},
	    {"X-Api-Version", "3.0"},
	    {"welove", "maltliquor"},
	    {"accept-encoding", "gzip, deflate"},
	    {"Content-Type", "application/json; charset=utf-8"},
	    {"Content-Length", "0"},
	  }
	  headers = append(headers, t.fnlGetHeaders()...)

	  resp, err := t.FnlDoQueueAwareReq(t.makeReq("GET", url_, &headers, &headerOrder, nil), "Waiting for Product")
	  if err != nil {
	  	return err
	  }
	  body, err := readBodyBytes(resp)
	  if err != nil {
	  	return err
	  }
	  err = json.Unmarshal(body, &product)
	  if err != nil {
	  	return err
	  }
	  if len(product.Colorways) == 0 {
	  	return errOutOfStock
	  }
	  return nil
	}, retry.OnRetry(func(n uint, err error) {
		if t.Debug {
			t.LogDebug("#%d: %s\n", n, err)
		}
	}))
  return product, err
}

type FnlSku struct {
	SkuId string `json:"skuId"`
	Size string `json:"size"`
	QuantityAvailable int `json:"quantityAvailable"`
	OutOfStock bool `json:"outOfStock"`
}

type FnlColorway struct {
	Description string `json:"colorDescription"`
	ColorId string `json:"colorId"`
	StyleId string `json:"styleId"`
	Skus []FnlSku `json:"skus"`
}

type FnlProduct struct {
	ProductId string `json:"productId"`
	Name string `json:"displayName"`
	Colorways []FnlColorway `json:"colorways"`
}

func (t *CheckoutTask) fnlGetHeaders() [][2]string {
	return [][2]string {
		{"x-cookie", t.Config["x-cookie"]},
		{"x-session", t.Config["x-session"]},
		{"AkamaiCookie", t.Config["AkamaiCookie"]},
		{"X-acf-sensor-data", t.Config["x-acf-sensor-data"]},
	}
}

type FnlOrderItem struct {
	ProductId string `json:"productId"`
	Name string `json:"displayName"`
	Quantity int `json:"quantity"`
	SkuId string `json:"skuId"`
}

type FnlOrder struct {
	Id string `json:"orderId"`
	Items []FnlOrderItem `json:"items"`
	Reason string `json:"reason"`
	StatusCode int `json:"statusCode"`
	Message string `json:"message"`
	Success bool `json:"success"`
}

func (t *CheckoutTask) fnlGetCart() (*FnlAppCart, error) {
	url_, err := url.Parse(fmt.Sprintf("%s/api/cart", t.fnlGetApiHost()))
	if err != nil {
	  return nil, err
	}
	headerOrder := []string {
	  "Host",
	  "User-Agent",
	  "Accept-Encoding",
	  "Connection",
	  "Accept",
	  "X-acf-sensor-data",
	  "X-Api-Version",
	  "x-cookie",
	  "welove",
	  "Content-Type",
	  "x-session",
	  "Content-Length",
	  "RiskifiedID",
	  "Riskified-User-Agent",
	  "AkamaiCookie",
	  "Cookie",
	}
	headers := [][2]string {
	  {"Connection", "keep-alive"},
	  {"Accept", "*/*"},
	  {"X-Api-Version", "3.0"},
	  {"welove", "maltliquor"},
	  {"accept-encoding", "gzip, deflate"},
	  {"Content-Type", "application/json; charset=utf-8"},
	  {"Content-Length", "0"},
	}
	headers = append(headers, t.fnlGetHeaders()...)

	resp, err := t.FnlDoQueueAwareReq(t.makeReq("GET", url_, &headers, &headerOrder, nil), "Check Cart")
	if err != nil {
		return nil, err
	}
	body, err := readBodyBytes(resp)
	if err != nil {
		return nil, err
	}
	var cart FnlAppCart
	err = json.Unmarshal(body, &cart)
	if err != nil {
		return nil, err
	}
	return &cart, err
}

type FnlAppCart struct {
	OrderID       string `json:"orderId"`
	ShippingGroup struct {
		ID             string      `json:"id"`
		ShippingState  int         `json:"shippingState"`
		LocationID     interface{} `json:"locationId"`
		ShippingMethod string      `json:"shippingMethod"`
		TrackingNumber interface{} `json:"trackingNumber"`
		PriceInfo      struct {
			CurrencyCode     string `json:"currencyCode"`
			AmountCents      int    `json:"amountCents"`
			AmountIsFinal    bool   `json:"amountIsFinal"`
			Discounted       bool   `json:"discounted"`
			RawShippingCents int    `json:"rawShippingCents"`
		} `json:"priceInfo"`
		Description     string `json:"description"`
		ShippingAddress struct {
			MiddleName  interface{} `json:"middleName"`
			LastName    interface{} `json:"lastName"`
			OwnerID     interface{} `json:"ownerId"`
			State       interface{} `json:"state"`
			Address1    interface{} `json:"address1"`
			Address2    interface{} `json:"address2"`
			Address3    interface{} `json:"address3"`
			Country     interface{} `json:"country"`
			City        interface{} `json:"city"`
			PostalCode  interface{} `json:"postalCode"`
			PhoneNumber interface{} `json:"phoneNumber"`
			County      interface{} `json:"county"`
			Email       interface{} `json:"email"`
			FirstName   interface{} `json:"firstName"`
		} `json:"shippingAddress"`
	} `json:"shippingGroup"`
	PaymentGroup  interface{} `json:"paymentGroup"`
	CartDiscounts struct {
		CouponAmountCents int           `json:"couponAmountCents"`
		CouponCode        string        `json:"couponCode"`
		RewardNumber      string        `json:"rewardNumber"`
		RewardAmountCents int           `json:"rewardAmountCents"`
		GiftCards         []interface{} `json:"giftCards"`
	} `json:"cartDiscounts"`
	Items []struct {
		ProductID   string `json:"productId"`
		DisplayName string `json:"displayName"`
		PriceInfo   struct {
			QuantityDiscountedCents int    `json:"quantityDiscountedCents"`
			OnSale                  bool   `json:"onSale"`
			CurrencyCode            string `json:"currencyCode"`
			AmountCents             int    `json:"amountCents"`
			Discountable            bool   `json:"discountable"`
			RawTotalPriceCents      int    `json:"rawTotalPriceCents"`
			ListPriceCents          int    `json:"listPriceCents"`
			AmountIsFinal           bool   `json:"amountIsFinal"`
			Discounted              bool   `json:"discounted"`
			PriceDetails            []struct {
				CurrencyCode           string `json:"currencyCode"`
				DetailedUnitPriceCents int    `json:"detailedUnitPriceCents"`
				AmountCents            int    `json:"amountCents"`
				TaxCents               int    `json:"taxCents"`
				AmountIsFinal          bool   `json:"amountIsFinal"`
				Discounted             bool   `json:"discounted"`
				Quantity               int    `json:"quantity"`
			} `json:"priceDetails"`
			SalePriceCents int `json:"salePriceCents"`
		} `json:"priceInfo"`
		Quantity    int    `json:"quantity"`
		MaxOrderQty int    `json:"maxOrderQty"`
		SkuID       string `json:"skuId"`
		ID          string `json:"id"`
	} `json:"items"`
	PriceInfo struct {
		TotalCents          int    `json:"totalCents"`
		CurrencyCode        string `json:"currencyCode"`
		DiscountAmountCents int    `json:"discountAmountCents"`
		AmountCents         int    `json:"amountCents"`
		ShippingCents       int    `json:"shippingCents"`
		TaxCents            int    `json:"taxCents"`
		AmountIsFinal       bool   `json:"amountIsFinal"`
		Discounted          bool   `json:"discounted"`
		RawSubTotalCents    int    `json:"rawSubTotalCents"`
	} `json:"priceInfo"`
	TaxPriceInfo struct {
		CurrencyCode     interface{} `json:"currencyCode"`
		CountryTaxCents  int         `json:"countryTaxCents"`
		StateTaxCents    int         `json:"stateTaxCents"`
		CityTaxCents     int         `json:"cityTaxCents"`
		DistrictTaxCents int         `json:"districtTaxCents"`
		CountyTaxCents   int         `json:"countyTaxCents"`
		AmountCents      int         `json:"amountCents"`
		AmountIsFinal    bool        `json:"amountIsFinal"`
		Discounted       bool        `json:"discounted"`
	} `json:"taxPriceInfo"`
	ProfileID       string `json:"profileId"`
	ItemCount       int    `json:"itemCount"`
	OrderTotalCents int    `json:"orderTotalCents"`
	LaunchItems     bool   `json:"launchItems"`
}

func (t *CheckoutTask) fnlAtc(product FnlProduct, colorWay FnlColorway, sku FnlSku) (error) {
	err := t.Retry(func() error {
		if cart, _ := t.fnlGetCart(); cart != nil {
			if len(cart.Items) > 0 {
				return nil
			}
		}
	  url_, err := url.Parse(fmt.Sprintf("%s/api/cart/add", t.fnlGetApiHost()))
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "Host",
	    "User-Agent",
	    "Accept-Encoding",
	    "Connection",
	    "Accept",
	    "X-acf-sensor-data",
	    "X-Api-Version",
	    "x-cookie",
	    "welove",
	    "X-JWT",
	    "X-REC",
	    "Content-Type",
	    "x-session",
	    "Content-Length",
	    "RiskifiedID",
	    "Riskified-User-Agent",
	    "AkamaiCookie",
	    "Cookie",
	  }
	  headers := [][2]string {
	    {"Connection", "keep-alive"},
	    {"Accept", "*/*"},
	    {"X-Api-Version", "3.0"},
	    {"X-JWT", ""},
	    {"X-REC", "null"}, // TODO real
	    {"welove", "maltliquor"},
	    {"accept-encoding", "gzip, deflate"},
	    {"Content-Type", "application/json; charset=utf-8"},
	  }
	  headers = append(headers, t.fnlGetHeaders()...)

	  bodyStructure := map[string]interface{} {
		  "colorId": colorWay.ColorId,
		  "productId": product.ProductId,
		  "quantity": 1,
		  "skuId": sku.SkuId,
		  "styleId": colorWay.StyleId,
		}
	  body, err := json.Marshal(bodyStructure)
	  if err != nil {
	    return err
	  }
	  resp, err := t.FnlDoQueueAwareReq(t.makeReq("POST", url_, &headers, &headerOrder, &body), "Add to Cart")
		if err != nil {
			return err
		}
		if resp.StatusCode == 499 {
			return retry.Unrecoverable(errors.New("Product Uncartable"))
		}
		body, err = readBodyBytes(resp)
		if err != nil {
			return err
		}
		t.LogDebug(string(body))
		var cart FnlAppCart
		err = json.Unmarshal(body, &cart)
		if len(cart.Items) == 0 {
			return ErrCartJacked
		}
		return nil
	})
	return err
}


func (t *CheckoutTask) fnlStartCheckout(sku FnlSku) (FnlOrder, error) {
	var order FnlOrder
	err := t.Retry(func() error {
	  url_, err := url.Parse(fmt.Sprintf("%s/api/Checkout/Guest", t.fnlGetApiHost()))
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "Host",
	    "User-Agent",
	    "Accept-Encoding",
	    "Connection",
	    "Accept",
	    "X-acf-sensor-data",
	    "X-Api-Version",
	    "x-cookie",
	    "welove",
	    "Content-Type",
	    "x-session",
	    "Content-Length",
	    "RiskifiedID",
	    "Riskified-User-Agent",
	    "AkamaiCookie",
	    "Cookie",
	  }
	  headers := [][2]string {
	    {"Accept-Encoding", "gzip, deflate"},
	    {"Connection", "keep-alive"},
	    {"Accept", "*/*"},
	    {"X-Api-Version", "3.0"},
	    {"welove", "maltliquor"},
	    {"Content-Type", "application/json; charset=utf-8"},
	  }
	  headers = append(headers, t.fnlGetHeaders()...)
	  bodyStructure := map[string]interface{} {
		  "inStorePickup": false,
		  "items": []interface{} {
		    map[string]interface{} {
          "quantity": 1,
          "skuId": sku.SkuId,
        },
		  },
		}
	  body, err := json.Marshal(bodyStructure)
	  if err != nil {
	    return err
	  }
	  resp, err := t.FnlDoQueueAwareReq(t.makeReq("POST", url_, &headers, &headerOrder, &body), "Start Checkout")
		if err != nil {
			return err
		}
		body, err = readBodyBytes(resp)
		if err != nil {
			return err
		}
		err = json.Unmarshal(body, &order)
		if err != nil {
			return err
		}
		if len(order.Items) == 0 {
			return ErrCartJacked
		}
		return nil
	}, retry.Attempts(1))
  return order, err
}

func (t *CheckoutTask) fnlDoCheckout(sku FnlSku, colorWay FnlColorway, fnlOrder FnlOrder) (error) {
	var err error
	t.StatusCh <- Status{Message:"Checkout 1/4: Email"}
	order := fnlOrder
  url_, err := url.Parse(fmt.Sprintf("%s/api/Checkout/Guest/%s/false", t.fnlGetApiHost(), t.Profile.Email))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "Host",
    "User-Agent",
    "Accept-Encoding",
    "Connection",
    "Accept",
    "X-acf-sensor-data",
    "X-Api-Version",
    "x-cookie",
    "welove",
    "Content-Type",
    "x-session",
    "Content-Length",
    "RiskifiedID",
    "Riskified-User-Agent",
    "AkamaiCookie",
    "Cookie",
  }
  headers := [][2]string {
    {"Accept-Encoding", "gzip, deflate"},
    {"Connection", "keep-alive"},
    {"Accept", "*/*"},
    {"X-Api-Version", "3.0"},
    {"welove", "maltliquor"},
    {"Content-Type", "application/json; charset=utf-8"},
    {"Content-Length", "0"},
  }
  headers = append(headers, t.fnlGetHeaders()...)
  resp, err := t.FnlDoQueueAwareReq(t.makeReq("POST", url_, &headers, &headerOrder, nil), "Checkout 1/4: Email")
	if err != nil {
		return err
	}
	body, err := readBodyBytes(resp)
	if err != nil {
		return err
	}
	order = FnlOrder{}
	err = json.Unmarshal(body, &order)
	if err != nil {
		return err
	}
	if order.Reason != "" {
		t.StatusCh <- Status{Message: "Checkout Error: " + order.Reason}
		return errors.New(fmt.Sprintf("Error: %s", order.Reason))
	}
	if len(order.Items) == 0 {
		t.StatusCh <- StatusCartJacked
		return ErrCartJacked
	}

	t.StatusCh <- Status{Message:"Checkout 2/4: Shipping"}
	url_, err = url.Parse(fmt.Sprintf("%s/api/Checkout/Guest/AddShippingGroup", t.fnlGetApiHost()))
	if err != nil {
	  return err
	}
  bodyStructure := map[string]interface{} {
	  "address1": t.Profile.ShippingAddress.Address1,
	  "address2": t.Profile.ShippingAddress.Address2,
	  "city": t.Profile.ShippingAddress.City,
	  "country": "US",
	  "email": t.Profile.Email,
	  "firstName": t.Profile.ShippingAddress.FirstName,
	  "lastName": t.Profile.ShippingAddress.LastName,
	  "phoneNumber": t.Profile.ShippingAddress.NordPhone(),
	  "postalCode": t.Profile.ShippingAddress.Zip,
	  "shippingMethod": "Economy",
	  "state": t.Profile.ShippingAddress.State,
	}
  body, err = json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  headers = [][2]string {
    {"Accept-Encoding", "gzip, deflate"},
    {"Connection", "keep-alive"},
    {"Accept", "*/*"},
    {"X-Api-Version", "3.0"},
    {"welove", "maltliquor"},
    {"Content-Type", "application/json; charset=utf-8"},
  }
  headers = append(headers, t.fnlGetHeaders()...)
  resp, err = t.FnlDoQueueAwareReq(t.makeReq("POST", url_, &headers, &headerOrder, &body), "Checkout 2/4: Shipping")
  body, err = readBodyBytes(resp)
  if err != nil {
  	return err
  }
 	order = FnlOrder{}
	err = json.Unmarshal(body, &order)
	if err != nil {
		return err
	}
	if order.Reason != "" {
		t.StatusCh <- Status{Message: "Checkout Error: " + order.Reason}
		return errors.New(fmt.Sprintf("Error: %s", order.Reason))
	}
	if len(order.Items) == 0 {
		t.StatusCh <- StatusCartJacked
		return ErrCartJacked
	}
  t.StatusCh <- Status{Message:"Checkout 3/4: Billing"}
  url_, err = url.Parse(fmt.Sprintf("%s/api/Checkout/Billing/NewAddressNewCard", t.fnlGetApiHost()))
	if err != nil {
	  return err
	}

	cardType := ""
	if t.Profile.Card.Type == "MASTER" {
		cardType = "Mastercard"
	} else if t.Profile.Card.Type == "VISA" {
		cardType = "Visa"
	} else if t.Profile.Card.Type == "DISCOVER" {
		cardType = "Discover"
	} else if t.Profile.Card.Type == "AMEX" {
		cardType = "Amex"
	} else {
		return errors.New(fmt.Sprintf("Unsupported card type: %s", t.Profile.Card.Type))
	}
  bodyStructure = map[string]interface{} {
  "address": map[string]interface{} {
    "address1": t.Profile.BillingAddress.Address1,
    "address2": t.Profile.BillingAddress.Address2,
    "address3": "",
    "city": t.Profile.BillingAddress.City,
    "emailAddress": t.Profile.Email,
    "firstName": t.Profile.BillingAddress.FirstName,
    "lastName": t.Profile.BillingAddress.LastName,
    "nickName": "",
    "phoneNumber": t.Profile.BillingAddress.NordPhone(),
    "postalCode": t.Profile.BillingAddress.Zip,
    "primary": false,
    "state": t.Profile.BillingAddress.State,
  },
  "creditCardExpirationMonth": fmt.Sprintf("%02d", t.Profile.Card.ExpMonth),
  "creditCardExpirationYear": fmt.Sprintf("%d", t.Profile.Card.ExpYear),
  "creditCardNickname": fmt.Sprintf("card-%d", timeMillis()),
  "creditCardNumber": t.Profile.Card.Number,
  "creditCardType": cardType,
  "defaultCreditCard": false,
  "newCreditCardVerificationNumber": t.Profile.Card.Cvc,
  "saveCardToProfile": false,
  "setDefault": false,
}
  body, err = json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  headers = [][2]string {
    {"Accept-Encoding", "gzip, deflate"},
    {"Connection", "keep-alive"},
    {"Accept", "*/*"},
    {"X-Api-Version", "3.0"},
    {"welove", "maltliquor"},
    {"Content-Type", "application/json; charset=utf-8"},
  }
  headers = append(headers, t.fnlGetHeaders()...)
  resp, err = t.FnlDoQueueAwareReq(t.makeReq("POST", url_, &headers, &headerOrder, &body), "Checkout 3/4: Billing")
  if err != nil {
  	return err
  }
  body, err = readBodyBytes(resp)
  if err != nil {
  	return err
  }
  order = FnlOrder{}
	err = json.Unmarshal(body, &order)
	if err != nil {
		return err
	}
	if order.Reason != "" {
		t.StatusCh <- Status{Message: "Checkout Error: " + order.Reason}
		return errors.New(fmt.Sprintf("Error: %s", order.Reason))
	}
	if len(order.Items) == 0 {
		t.StatusCh <- StatusCartJacked
		return ErrCartJacked
	}

  t.StartPaymentTime = timeMillis()

  t.StatusCh <- Status{Message:"Checkout 4/4: Pay"}
  url_, err = url.Parse(fmt.Sprintf("%s/api/Checkout/CommitOrder", t.fnlGetApiHost()))
	if err != nil {
	  return err
	}
  bodyStructure = map[string]interface{} {
    "donation": false,
    "isResidential": true,
  }
  body, err = json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  headers = [][2]string {
    {"Accept-Encoding", "gzip, deflate"},
    {"Connection", "keep-alive"},
    {"Accept", "*/*"},
    {"X-Api-Version", "3.0"},
    {"welove", "maltliquor"},
    {"Content-Type", "application/json; charset=utf-8"},
  }
  headers = append(headers, t.fnlGetHeaders()...)
  resp, err = t.FnlDoQueueAwareReq(t.makeReq("POST", url_, &headers, &headerOrder, &body), "Checkout 4/4: Pay")
  if err != nil {
  	return err
  }
  body, err = readBodyBytes(resp)
  if err != nil {
  	return err
  }
  statusCode := 0
  if resp != nil {
  	statusCode = resp.StatusCode
  }
  go t.SendTelemetry(map[string]interface{} {
  	"event": "fnl_app_order_response",
  	"fnl_app_order_response": string(body),
  	"fnl_app_order_response_code": statusCode,
  })
  order = FnlOrder{}
	err = json.Unmarshal(body, &order)
	if err != nil {
		return err
	}

  whFields := [][2]string{
		{"Website", t.Url.Host + " (APP)"},
		{"Product", fnlOrder.Items[0].Name},
		{"Size", sku.Size},
		{"Color", colorWay.Description},
	}

  if order.Reason == "FailedCreditCardAuth" {
  	t.StatusCh <- StatusCardDeclined
  	go t.DeclineWebhookFields(whFields)
  	return retry.Unrecoverable(retry.Unrecoverable(errCardDeclined))
  } else if order.Reason != "" {
		t.StatusCh <- Status{Message: "Order Error: " + order.Reason}
  	return errors.New(order.Reason)
  } else if order.Success {
  	t.StatusCh <- StatusCooked
  	go t.SuccessWebhookFields(whFields)
  	return nil
  } else if len(order.Items) == 0 {
  	t.StatusCh <- StatusOosOnOrder
  	return ErrCartJacked
  }
	t.StatusCh <- Status{Message:"Unknown checkout error"}
	return errors.New("Unknown checkout error")
}

func (t *CheckoutTask) getSensor() (string, error) {
	var sensor string
	err := t.Retry(func() error {
		client := &http.Client{}
		url_, err := url.Parse(fmt.Sprintf("%s/mobile?package=com.winnerscircle", apiHost()))
		if err != nil {
			return err
		}
		req := t.makeGetReq(url_, nil)
		req.Header = map[string][]string {
			"Authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
		}
		resp, err := t.doReq(client, req)
		if err != nil {
			return err
		}
		sensor, err = readBody(resp)
		if err != nil {
			return err
		}
		// fmt.Println(sensor)
		sbja := []string{}
		if err = json.Unmarshal([]byte(sensor), &sbja); err == nil {
			sensor = sbja[0]
			t.Config["rawsensor"] = sbja[1]
		}

		return nil
	}, retry.Attempts(1e6))
	return sensor, err
}

type FnlErrorResponse struct {
	Message string `json:"message"`
	Reason string `json:"reason"`
}

func (t *CheckoutTask) FnlDoQueueAwareReq(req *http.Request, description string) (*http.Response, error) {
	var resp *http.Response
	var err error
	retries := 0
	// ret403 := 0

	err = t.Retry(func() error {
		t.GenericQueueBypassUrl(t.fnlGetApiHostUrl())
		t.GenericQueueBypassUrl(req.URL)

		resp, err = t.doReq(t.client, req)
		if resp != nil {
			body, berr := ioutil.ReadAll(resp.Body)
			if berr != nil {
			  return berr
			}
			bodystr := strings.ToLower(string(body))
			resp.Body = ioutil.NopCloser(bytes.NewReader(body))

			if resp.StatusCode != 403 && strings.Contains(bodystr, "Unexpected Request") {
				t.StatusCh <- Status{Message:fmt.Sprintf("In Queue (%s)", description)}
				DiscardResp(resp)
				return ErrWaitingForPermaQueue
			}
		}
		if resp != nil && resp.StatusCode != 200 && resp.StatusCode != 499 {
			if resp.StatusCode == 403 {
				// TODO try fresh sensor
				// sensor, err := t.getSensor()
				// if err != nil {
				// 	return err
				// }

				// t.RemoveCookieUrl("akavpau_FL-MobApp", t.fnlGetApiHostUrl())
				// t.RemoveCookieUrl("ak_bmsc", t.BaseDomainUrl)
				// t.Config["x-acf-sensor-data"] = sensor
				// for idx, h := range req.RawHeader[:] {
				// 	if h[0] == "x-acf-sensor-data" {
				// 		req.RawHeader[idx][1] = sensor
				// 		break
				// 	}
				// }
				// ret403 += 1
				// if ret403 > 10 {
				t.StatusCh <- StatusSessionBanRestart
				return TaskErrUnrecoverable(errors.New("Banned: Try a different proxy"))
				// }
			} else if resp.StatusCode == 206 {
				t.StatusCh <- Status{Message:fmt.Sprintf("In Queue (%s)", description)}
			} else {
				t.StatusCh <- Status{Message:fmt.Sprintf("%s (%d)", description, resp.StatusCode)}
			}
			return errors.New("Waiting for Session")
		} else if err != nil {
			return err
		}
		// return errors.New("retry for cookie bloat")
		return nil
	}, retry.OnRetry(func(n uint, err error) {
		retries += 1
		if t.Debug {
			t.LogDebug("#%d: %s\n", n, err)
		}
	}), retry.Attempts(2048), retry.Delay(2 * time.Second), retry.DelayType(retry.FixedDelay))

	if retries > 32 && err != nil {
		err = TaskErrUnrecoverable(err)
	}
	return resp, err
}

var (
	ErrFnlProductNeedsSpecificity = errors.New("URL must include product ID, styleId, and colorId")
)

type FnlDesiredProduct struct {
	ProductId string
	StyleId string
	ColorId string
}

func (t *CheckoutTask) FnlEnsureUrlHasProductSpecificity() (*FnlDesiredProduct, error) {
	urlSplit := strings.Split(t.Url.Path, "/")
	pid := urlSplit[len(urlSplit) - 1]
	q := t.Url.Query()
	styleId := q.Get("styleId")
	colorId := q.Get("colorId")

	if len(pid) == 0 || len(styleId) == 0 || len(colorId) == 0 {
		t.StatusCh <- StatusFnlProductNeedsSpecificity
		return nil, ErrFnlProductNeedsSpecificity
	} else {
		return &FnlDesiredProduct{
			ProductId: pid,
			StyleId: styleId,
			ColorId: colorId,
		}, nil
	}
}

func (t *CheckoutTask) FnlAppCheckout() error {
	desiredProduct, err := t.FnlEnsureUrlHasProductSpecificity()
	if err != nil {
		return err
	}

	// u := t.fnlGetApiHostUrl()
	// t.UpdateDnsCache(u.Host)


	t.UserAgent = "Finish Line/2.7.4 (Android 10; Build/PKQ1.180716.001)"
	t.HeaderBlacklist = map[string]bool {
		"sec-fetch-site": true,
		"sec-fetch-mode": true,
		"sec-fetch-dest": true,
		"accept-language": true,
	}

	var lastErr error = nil
	var attempts uint = 0
	// var retryStr string
	return t.RetryTask(func() error {
		// if lastErr != nil {
		// 	retryStr = fmt.Sprintf("(Retrying %d/%d)", attempts, MAX_RETRIES)
		// }
		t.initClient()
		t.StatusCh <- Status{Message:fmt.Sprintf("Starting Session (Initializing)")}
		err := t.fnlGetConfig()
		if err != nil {
			return err

		}
		t.StatusCh <- Status{Message:fmt.Sprintf("Starting Session (Generating Cookie)")}
		sensor, err := t.getSensor()
		if err != nil {
			return err
		}

		t.Config["x-acf-sensor-data"] = sensor

		t.StatusCh <- Status{Message:"Starting Session (Sending Cookie)"}
		err = t.fnlGetAccountGuest()
		if err != nil {
			t.LogDebug("\n\nBMPFAIL")
			fmt.Println(t.Config["rawsensor"])
			// t.LogDebug(t.Config["rawsensor"])
			return err
		}

		t.fnlGetDrops()

		var product FnlProduct
		var colorWay *FnlColorway
		var sku *FnlSku
		return t.Retry(func() error {
			var err error
			t.StatusCh <- Status{Message:"Waiting for Product"}
			product, err = t.fnlGetProduct(desiredProduct.ProductId)
			if err != nil {
				return err
			}
			t.LogDebug("%+v", product)


			for _, cw := range product.Colorways {
				if cw.StyleId != desiredProduct.StyleId || cw.ColorId != desiredProduct.ColorId {
					continue
				}
				rand.Shuffle(len(cw.Skus), func(i, j int) {
				  cw.Skus[i], cw.Skus[j] = cw.Skus[j], cw.Skus[i] })

				for _, _sku := range cw.Skus {
					if _sku.QuantityAvailable > 0 {
						if len(t.Sizes) == 0 || Any(t.Sizes, func(size string) bool { return size == strings.ToLower(_sku.Size) || size + ".0" == strings.ToLower(_sku.Size) }) {
							sku = &_sku
							colorWay = &cw
							break
						}
					}
				}
				if sku != nil {
					break
				}
			}

			if sku == nil || colorWay == nil {
				t.StatusCh <- Status{Message:"Waiting for Product (Out of Stock)"}
				return errOutOfStock
			}

			t.StartAtcTime = timeMillis()
			t.StatusCh <- Status{Message:"Add to Cart"}
			err = t.fnlAtc(product, *colorWay, *sku)
			if err != nil {
				return err
			}

			if rand.Float32() < 0.1 {
				t.StatusTelemetryEnabled = true
			}

			t.StartCheckoutTime = timeMillis()
			t.StatusCh <- Status{Message:"Start Checkout"}
			fnlOrder, err := t.fnlStartCheckout(*sku)
			if err != nil {
				return err
			}

			return t.fnlDoCheckout(*sku, *colorWay, fnlOrder)
		})
	}, retry.Attempts(MAX_RETRIES), retry.MaxDelay(15 * time.Second), retry.Delay(1 * time.Second), retry.DelayType(retry.DefaultDelayType), retry.LastErrorOnly(true), retry.OnRetry(func(attempts_ uint, err_ error) {
		attempts = attempts_ + 1
		lastErr = err_
    t.client.CloseIdleConnections()
    // if rt, ok := t.client.Transport.(*roundTripper); ok {
	   //  log.Println("CLOSE ALL")
	   //  t.LogDebug("%+v", reflect.TypeOf(rt.transport))
    // 	if h2trs, ok := rt.transport.(*http2.Transport); ok {
    // 		log.Println("CLOSE ALL")
	   //  	h2trs.ConnPool.CloseAll()
	   //  }
    // }
	}))
}

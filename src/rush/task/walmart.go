package task

import (
	"context"
	"sync"
	"bytes"
	"io"
	"encoding/base32"
	"io/ioutil"
	"math/rand"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	// "log"
	"net/url"
	"strconv"
	"strings"
	"fmt"
	"rush/net/http"
	"time"
	"regexp"
	"os"
)

var (
	errDecline = errors.New("Card Declined")
	WmtErrThirdPartySeller = errors.New("Waiting for product: third-party seller detected.")
	WmtStatusThirdPartySeller = Status{Message: "Waiting for product: third-party seller detected."}
	WMT_SELLER_ID = "F55CDC31AB754BB68FE0B39041159D63"
)

type WalmartSession struct {
	lastPxGenAt time.Time
	*CheckoutTask
}

func (t *CheckoutTask) WmtSetDefaultShipping() error {
	return t.Retry(func() error {
		su, _ := url.Parse("https://www.walmart.com/account/deliveryaddresses")
		headerOrder := []string {
		  "upgrade-insecure-requests",
		  "user-agent",
		  "accept",
		  "sec-fetch-site",
		  "sec-fetch-mode",
		  "sec-fetch-dest",
		  "accept-encoding",
		  "accept-language",
		}
		headers := [][2]string {
		  {"upgrade-insecure-requests", "1"},
		  {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		  {"sec-fetch-site", "none"},
		  {"sec-fetch-mode", "navigate"},
		  {"sec-fetch-dest", "document"},
		  {"accept-encoding", "gzip, deflate, br"},
		  {"accept-language", "en-US,en;q=0.9"},
		}
		// t.Url.Query().Set("ts", fmt.Sprintf("%d", timeMillis()))
		resp, err := t.doReq(t.client, t.makeReq("GET", su, &headers, &headerOrder, nil))
		if err != nil {
			return err
		}
		jwt := resp.Header.Get("x-csrf-jwt")

	  url_, err := url.Parse("https://www.walmart.com/account/electrode/account/api/customer/:CID/shipping-address")
	  if err != nil {
	    return err
	  }
	  headerOrder = []string {
	    "content-length",
	    "accept",
	    "x-csrf-jwt",
	    "inkiru_precedence",
	    "user-agent",
	    "wm_vertical_id",
	    "content-type",
	    "origin",
	    "sec-fetch-site",
	    "sec-fetch-mode",
	    "sec-fetch-dest",
	    "referer",
	    "accept-encoding",
	    "accept-language",
	    "cookie",
	  }
	  headers = [][2]string {
	    {"accept", "application/json, text/javascript, */*; q=0.01"},
	    {"inkiru_precedence", "false"},
	    {"x-csrf-jwt", jwt},
	    {"wm_vertical_id", "0"},
	    {"content-type", "application/json"},
	    {"origin", "https://www.walmart.com"},
	    {"sec-fetch-site", "same-origin"},
	    {"sec-fetch-mode", "cors"},
	    {"sec-fetch-dest", "empty"},
	    {"referer", "https://www.walmart.com/checkout/"},
	    {"accept-encoding", "gzip, deflate, br"},
	    {"accept-language", "en-US,en;q=0.9"},
	  }
	  bodyStructure := map[string]interface{} {
	  "addressLineOne": t.Profile.ShippingAddress.Address1,
	  "addressLineTwo": t.Profile.ShippingAddress.Address2,
	  "city": strings.ToUpper(t.Profile.ShippingAddress.City),
	  "firstName": t.Profile.ShippingAddress.FirstName,
	  "lastName": t.Profile.ShippingAddress.LastName,
	  "phone": t.Profile.ShippingAddress.NordPhone(),
	  "addressType": "RESIDENTIAL",
	  "validationStatus": "USER_ENTERED",
	  "isDefault": true,
	  "postalCode": t.Profile.ShippingAddress.Zip,
	  "state": t.Profile.ShippingAddress.State,
	  "countryCode": "USA",
	}
	  body, err := json.Marshal(bodyStructure)
	  if err != nil {
	    return err
	  }
	  req := t.makeReq("POST", url_, &headers, &headerOrder, &body)
	  req.IgnoreBody = true
	  req.IgnoreBodyExceptStatus = []int{412}
		resp, err = t.doReq(t.client, req)
	  defer DiscardResp(resp)
	  if resp != nil {
	  	if resp != nil && resp.StatusCode == 412 {
		  		if resp.StatusCode == 412 {
						if bbb, err := readBodyBytes(resp); err == nil {
  if err := t.WmtSolveCaptcha(false, bbb); err != nil {
            return err
          }
        }
						return ErrRetrying
	  			}
		  	}
	  	io.Copy(ioutil.Discard, resp.Body)
	  	resp.Body.Close()
	  }
	  if resp.StatusCode != 200 {
	  	err = ErrRetrying
	  }
	  return err
	}, retry.OnRetry(func (n uint, err error) {
	 		t.StatusCh <- Status{Message: fmt.Sprintf("Waiting to Set Shipping (%v)", err)}
	 	}), retry.Delay(t.Delay), retry.Attempts(8))
}

func (s *CheckoutTask) WmtGetProduct(quiet bool, waitForStock bool) (*WmtProduct, error) {
	var product WmtProduct
	retries := 0
	var solved bool
	t := s
	var lastProdJson []byte
	err  := t.Retry(func() error {
		if rand.Float32() < 0.15 {
			// if ckies, err := t.WmtGenPxCookie(); err == nil {
			// 	for _, ckie := range ckies {
			// 		t.SetCookie(ckie)
			// 	}
			// } else {
			// 	t.LogDebug("px gen err %+v", err)
			// }
		}
		if !quiet  {
			if waitForStock {
				t.StatusCh <- StatusWaitStock
			} else {
				t.StatusCh <- StatusVisitProduct
			}
		}
		defer func() {
			retries += 1
		}()
	  headerOrder := []string {
	    "upgrade-insecure-requests",
	    "user-agent",
	    "accept",
	    "sec-fetch-site",
	    "sec-fetch-mode",
	    "sec-fetch-dest",
	    "accept-encoding",
	    "accept-language",
	  }
	  headers := [][2]string {
	    {"upgrade-insecure-requests", "1"},
	    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	    {"sec-fetch-site", "none"},
	    {"sec-fetch-mode", "navigate"},
	    {"sec-fetch-dest", "document"},
	    {"accept-encoding", "gzip, deflate, br"},
	    {"accept-language", "en-US,en;q=0.9"},
	  }
	  t.Url.Query().Set("ts", fmt.Sprintf("%d", timeMillis()))
	  resp, err := t.doReq(t.client, t.makeReq("GET", t.Url, &headers, &headerOrder, nil))
	  if resp != nil {
	  	if resp.StatusCode == 412 {
	  		if resp.StatusCode == 412 {
				if bbb, err := readBodyBytes(resp); err == nil {
  if err := t.WmtSolveCaptcha(false, bbb); err != nil {
            return err
          }
        }
				return ErrRetrying
	  	}
	  	}
	  	if resp.StatusCode == 307 {
	  		if solved {
	  			return retry.Unrecoverable(ErrFailedToSolveCaptcha)
	  		}
	  		solved = false
	  		// if bbb, err := readBodyBytes(resp); err == nil {
  			pxBlock := DEFAULT_BR_STRUCT
  			pxBlock.RedirectURL = resp.Header.Get("location")
  			if br, err := json.Marshal(pxBlock); err == nil {
		  		if err := t.WmtSolveCaptcha(false, br); err != nil {
						return err
					} else {
						solved = true
					}
  			}
	  		return ErrRetrying
	  	}
	  }
	  if err != nil {
	  	DiscardResp(resp)
	  	return err
	  }
	  if !waitForStock {
	  	DiscardResp(resp)
	  	return nil
	  }
	  body, err := readBody(resp)
	  if err != nil {
	  	return err
	  }
	  re := regexp.MustCompile(`type="application/json">({"item":.+?)<\/script>`)
	  matches := re.FindStringSubmatch(body)
	  if len(matches) != 2 {
	  	return errors.New("Malformed Product Page")
	  }
	  prodJson := []byte(matches[1])
	  // log.Println(string(prodJson))
	  if len(lastProdJson) != len(prodJson) {
	  	defer func() {
	  		lastProdJson = prodJson
	  	}()
		  // var data interface{}
		  // err = json.Unmarshal(prodJson, &data)
		  // prettyJSON, _ := json.MarshalIndent(data, "", "    ")
		  // fmt.Printf("%s\n", string(prettyJSON))

		  // log.Printf("%+v", string( prodJson))
		  err = json.Unmarshal(prodJson, &product)
		  if err != nil {
		  	// log.Printf("%+v", err)
		  	return err
		  }
		  if len(product.Item.Product.BuyBox.Products) < 1 {
		  	// log.Printf("missing offerid")
		  	return errors.New("Product Page Missing OfferID")
		  }
		  var hasWmtSeller bool
		  var wmtProd WmtProductItem
		  for _, prod_ := range product.Item.Product.BuyBox.Products {
		  	t.LogDebug("%+v", prod_)
		  	wmtProd = prod_
		  	if prod_.CatalogSellerID == "0" {
		  		hasWmtSeller = true
		  		break
		  	}
		  }

		  if  waitForStock {
			  if !hasWmtSeller && os.Getenv("WMT_THIRD_PARTY") != "1" {
			  	t.StatusCh <- WmtStatusThirdPartySeller
			  	return WmtErrThirdPartySeller
			  }
			  if !wmtProd.Shippable || (os.Getenv("OOS_SIM") == "1" && os.Getenv("MOBILE") == "") {
			  	return ErrOutOfStock
			  }
			}
		  return nil
		} else {
			return ErrOutOfStock
		}
	 }, retry.Delay(t.Delay*2), retry.Attempts(32))
	return &product, err
  // re := regexp.MustCompile(`"offerId":"(.+?)"`)
}

var ReWmtPacUrl = regexp.MustCompile(`'(.+?)'`)

func (t *CheckoutTask) WmtAtcAlt(offerId string, quantity string) error {
	return t.Retry(func() error {
		url_, err := url.Parse("https://www.walmart.com/product/plugins/atc/ssatc")
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "content-length",
	    "accept",
	    "user-agent",
	    "content-type",
	    "origin",
	    "sec-fetch-site",
	    "sec-fetch-mode",
	    "sec-fetch-dest",
	    "referer",
	    "accept-encoding",
	    "accept-language",
	    "cookie",
	  }
	  headers := [][2]string {
	    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	    {"content-type", "application/x-www-form-urlencoded"},
	    {"origin", "https://www.walmart.com"},
	    {"sec-fetch-site", "same-origin"},
	    {"sec-fetch-mode", "same-origin"},
	    {"sec-fetch-dest", "empty"},
	    {"referer", t.Url.String()},
	    {"accept-encoding", "gzip, deflate, br"},
	    {"accept-language", "en-US,en;q=0.9"},
	  }
	  atcParams := [][2]string {
	  	{"offerId", offerId},
	  	{"storeIds", "5939,2516,3098,2317,2325"},
	  	{"quantity", quantity},
	  }
	  var atcKvs [][]byte
	  for _, kv := range atcParams {
	  	atcKvs = append(atcKvs, []byte(fmt.Sprintf("%s=%s", url.QueryEscape(kv[0]), url.QueryEscape(kv[1]))))
	  }
	  atcBody := bytes.Join(atcKvs, []byte("&"))
	  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &atcBody))
	  defer DiscardResp(resp)
	  if resp != nil {
	  	defer resp.Body.Close()
	  }
  	if resp != nil && resp.StatusCode == 307 {

pxBlock := DEFAULT_BR_STRUCT
  			pxBlock.RedirectURL = resp.Header.Get("location")
  			if br, err := json.Marshal(pxBlock); err == nil {
		  		if err := t.WmtSolveCaptcha(false, br); err != nil {
						return err
					}
  			}

  		return ErrRetrying
  	}
	  if err != nil {
	  	if resp.StatusCode == 429 {
	  		t.StatusCh <- Status{Message: "Rate Limited/Proxy Banned, Backing Off"}
	  		time.Sleep(time.Duration(rand.Intn(120))*time.Second)
	  		return ErrRetrying
	  	}
	  	return err
	  }
	  body, err := readBody(resp)
	  if resp != nil  {

			if resp.StatusCode == 412 {
				// DiscardResp(resp)
			if resp.StatusCode == 412 {
						if bbb, err := readBodyBytes(resp); err == nil {
  if err := t.WmtSolveCaptcha(false, bbb); err != nil {
            return err
          }
        }
						return ErrRetrying
			  	}
			}

	  	if resp.StatusCode == 200 || resp.StatusCode == 201 {
	  		io.Copy(ioutil.Discard, resp.Body)
	  		resp.Body.Close()
	  	} else {
		  	var errResp WmtErrorResponse
		  	err := readRespJsonDst(resp, &errResp)
		  	// log.Printf("%+v %+v", err, errResp)
		  	if err == nil {
		  		return errors.New(errResp.Message)
		  	} else if resp.StatusCode == 444 {
	  			t.StatusCh <- StatusProxyBanned
	  			return retry.Unrecoverable(err)
		  		// }
		  		// return t.WmtAtcAlt(offerId, quantity)
		  	}
		  }
	  }
	  matches := ReWmtPacUrl.FindStringSubmatch(body)
	  if len(matches) != 2 {
	  	return ErrRetrying
	  }
	  if gurl, err := url.Parse(matches[1]); err == nil {
	  	resp, _ := t.doReq(t.client, t.makeGetReq(gurl, nil))
	  	DiscardResp(resp)
	  	if t.GetCookieValue("CRT") != "" {
	  		return nil
	  	}
	  }
	  return ErrRetrying
	})
}

type WmtCart struct {
	Items []struct {
		Id string `json:"id"`
		Seller struct {Name string `json:"name"`} `json:"seller"`
		Quantity int `json:"quantity"`

	} `json:"items"`
}
func (t *CheckoutTask) WmtAtcAlt2(qty int) error {
	ps := strings.Split(t.Url.Path, "/")
	pid := ps[len(ps) - 1]

	url_, err := url.Parse("https://affil.walmart.com/cart/buynow?items=" + pid + "%7C" + fmt.Sprintf("%d", qty))
  if err != nil {
    return err
  }
  req := t.makeGetReq(url_, nil)
  req.IgnoreBody = true
  CloseH2Conns(t.client)
  t.RemoveCookie("CRT")
  t.RemoveCookie("hasCRT")
  t.RemoveCookie("CXO_CART_ST")
  t.RemoveCookie("cart-item-count")
  resp, err := t.doReq(t.client, req)
  defer DiscardResp(resp)
  if err != nil {
  	return err
  }
  if resp == nil {
  	return ErrRetrying
  }
  for _, ckie := range resp.Header["Set-Cookie"]  {
  	t.LogDebug("SCK %s", ckie)
  	if kv := strings.Split(strings.Split(ckie, "; ")[0], "="); len(kv) == 2 {
  		t.LogDebug("sck %s %s", kv[0], kv[1])
  		t.SetCookie(&http.Cookie{Name: kv[0], Value: kv[1]})

  	}
  }

  if t.GetCookieValue("CRT") != "" {
  	if qty > 0 {
  	// if qty > 1 {
  		if req, err := t.MakeChlsRequest(WMT_GETCART_CHLSJ, nil); err == nil {
  			resp, err := t.doReq(t.client, req.Request)
  			if err != nil {
  				return err
  			}
  			if resp != nil {
  				if resp.StatusCode == 412 {
  								if bbb, err := readBodyBytes(resp); err == nil {
  				  				if err := t.WmtSolveCaptcha(false, bbb); err != nil {
  				            return retry.Unrecoverable(ErrRetrying)
  				          }
  				        }
  								return ErrRetrying
  					  	}
  				var cc WmtCart
  				readRespJsonDst(resp, &cc)
  				if len(cc.Items) == 1 && cc.Items[0].Quantity == qty {
  					for _, it := range cc.Items {
  						if it.Seller.Name != "Walmart.com" {
  							return ErrRetrying
  						}
  					}
	  				return nil
  				} else {
  					return ErrRetrying
  				}
  			}
  		}

  	  return ErrRetrying
  	}
  	return ErrRetrying
  }
  return ErrRetrying

}

func (t *CheckoutTask) WalmartLogin(email string, pw string) (error) {
	t.StatusCh <- StatusLoggingIn
	// pw = "fake"
	return t.Retry(func() error {
		resp, err := t.DoChlsRequest(WMT_LOGIN_CHLSJ, map[*regexp.Regexp]string {
			regexp.MustCompile(`eddie_rath@gmail.com`): email,
			regexp.MustCompile(`task_single_5842`): pw,
		})
		if resp != nil {
			if resp.StatusCode == 412 {
							if bbb, err := readBodyBytes(resp); err == nil {
			  				if err := t.WmtSolveCaptcha(false, bbb); err != nil {
			            return retry.Unrecoverable(ErrRetrying)
			          }
			        }
							return ErrRetrying
				  	}
			if bb, err := readBody(resp); err == nil {
				if strings.Contains(bb, "user_auth_fail") {
					t.StatusCh <- Status{Message: "Login failed, please check password in vault"}
					return retry.Unrecoverable(retry.Unrecoverable(ErrRetrying))
				} else if strings.Contains(bb, "user_locked") {
					t.StatusCh <- Status{Message: "Account locked, please reset password and update vault, or restart task in some minutes"}
					return retry.Unrecoverable(retry.Unrecoverable(ErrRetrying))
				}
			}
		}
		if t.GetCookieValue("CID") == "" {
			return ErrRetrying
		}
		// TODO call site response handler, for captcha etc, dedupe this file
		DiscardResp(resp)
		gu, _ := url.Parse("https://www.walmart.com/account?r=yes&action=SignIn&rm=true")
		t.doReq(t.client, t.makeGetReq(gu, nil))

		return err
	}, retry.Attempts(8))
}

var NonAlphaReg = regexp.MustCompile("[^a-zA-Z0-9]+")

func (t *CheckoutTask) WmtAtc(offerId string, quantity int64, product *WmtProductItem) (error) {
	last := time.Now()
	ret := 0
	return t.Retry(func() error {
		if rand.Float32() < 0.5 || offerId == "" {
			return t.WmtAtcAlt2(int(quantity))
		}
		defer func() {
			ret += 1
		}()

		// t.LogDebug("%+v %+v %+v", ret, t.Delay, last)
		if ret > 0 && time.Now().Sub(last) < t.Delay {
			time.Sleep(t.Delay - (time.Now().Sub(last)))
		}
		last = time.Now()
	  url_, err := url.Parse("https://www.walmart.com/api/v3/cart/guest/:CID/items")
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "content-length",
	    "accept",
	    "user-agent",
	    "content-type",
	    "origin",
	    "sec-fetch-site",
	    "sec-fetch-mode",
	    "sec-fetch-dest",
	    "referer",
	    "accept-encoding",
	    "accept-language",
	    "cookie",
	  }
	  headers := [][2]string {
	    {"accept", "application/json"},
	    {"content-type", "application/json"},
	    {"origin", "https://www.walmart.com"},
	    {"sec-fetch-site", "same-origin"},
	    {"sec-fetch-mode", "cors"},
	    {"sec-fetch-dest", "empty"},
	    {"referer", t.Url.String()},
	    {"accept-encoding", "gzip, deflate, br"},
	    {"accept-language", "en-US,en;q=0.9"},
	  }
	  city := strings.Split(t.Profile.ShippingAddress.City, " ")[0]
	  city = NonAlphaReg.ReplaceAllString(city, "")

		  bodyStructure := map[string]interface{} {
		  "offerId": offerId,
		  "quantity": quantity,
		  // TODO need location api?
		  "location": map[string]interface{} {
		    "postalCode": t.Profile.ShippingAddress.Zip,
		    "city": city,
		    "state": t.Profile.ShippingAddress.State,
		    "isZipLocated": true,
		  },
		  "shipMethodDefaultRule": "SHIP_RULE_1",
		  "storeIds": []interface{} {
		  	1988,1881,5192,4202,3090,
		  },
		}
	  if len(product.InflexibleKitGroupComponents.GroupComponents) > 0 {
	  	bodyStructure["groupId"] = bodyStructure["offerId"]
	  	delete(bodyStructure, "offerId")
	  	bodyStructure["groupComponents"] = product.InflexibleKitGroupComponents.GroupComponents
		}
	  body, err := json.Marshal(bodyStructure)
	  if err != nil {
	    return err
	  }
	  t.RemoveCookie("CRT")
	  t.RemoveCookie("hasCRT")
	  t.RemoveCookie("CXO_CART_ST")
	  t.RemoveCookie("cart-item-count")
		resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
		defer DiscardResp(resp)
		if t.GetCookieValue("CRT") != "" {
			DiscardResp(resp)
			return nil
		}
	  if resp != nil  {
			if resp.StatusCode == 444 {
	  		return t.WmtAtcAlt2(int(quantity))
	  	}
	  	if resp.StatusCode == 429 {
	  		t.StatusCh <- Status{Message: "Rate Limited/Proxy Banned, Backing Off"}
	  		time.Sleep(time.Duration(rand.Intn(120))*time.Second)
	  		return ErrRetrying
	  	}
			if resp.StatusCode == 412 {
				if bbb, err := readBodyBytes(resp); err == nil {
  				if err := t.WmtSolveCaptcha(false, bbb); err != nil {
            return err
          }
        }
				return ErrRetrying
	  	}
	  	if resp.StatusCode == 200 || resp.StatusCode == 201 {

	  	} else {
		  	var errResp WmtErrorResponse
		  	err := readRespJsonDst(resp, &errResp)
		  	// log.Printf("%+v %+v", err, errResp)
		  	if err == nil {
		  		return errors.New(errResp.Message)
		  	}
		  }
	  }
	  if err == nil {
	  	err = ErrRetrying
	  }
  	return err
	}, retry.OnRetry(func (n uint, err error) {
		// log.Printf("%+v", err)
		t.StatusCh <- Status{Message: fmt.Sprintf("Waiting to Add to Cart (Out of Stock)")}
	}), retry.Delay(t.Delay), retry.Attempts(256))
}

func (t *CheckoutTask) WmtMobileGetProduct(pid string) (*WalmartMobileProduct, error) {
	var product WalmartMobileProduct
	var retries = 0
	err := t.Retry(func() error {
		defer func() {
			retries += 1
		}()
	  url_, err := url.Parse("https://www.walmart.com/terra-firma/graphql?v=2&options=timing%2Cnonnull%2Cerrors%2Ccontext&id=FullProductRoute-android")
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "wm_consumer.id",
	    "wm_site_mode",
	    "itemid",
	    "user-agent",
	    "did",
	    "x-px-authorization",
	    "content-type",
	    "content-length",
	    "accept-encoding",
	    "cookie",
	  }
	  t.HeaderBlacklist["sec-fetch-dest"] = true
	  t.HeaderBlacklist["sec-fetch-site"] = true
	  defer func() {
	  	t.HeaderBlacklist["sec-fetch-dest"] = false
	  	t.HeaderBlacklist["sec-fetch-site"] = false
	  }()
	  headers := [][2]string {
	    {"wm_consumer.id", uuid.New().String()},
	    {"wm_site_mode", "0"},
	    {"itemid", pid},
	    {"did", uuid.New().String()},
	    {"x-px-authorization", t.GetCookieValue("_px3")},
	    {"user-agent", "Android v20.36.1"},
	    {"content-type", "application/json; charset=utf-8"},
	    {"accept-encoding", "gzip"},
	  }
	  bodyStructure := map[string]interface{} {
	  "variables": fmt.Sprintf("{\"casperSlots\":{\"fulfillmentType\":\"ACC\",\"reservationType\":\"SLOTS\"},\"postalAddress\":{\"addressType\":\"RESIDENTIAL\",\"countryCode\":\"USA\",\"postalCode\":\"%s\",\"stateOrProvinceCode\":\"%s\",\"zipLocated\":true},\"storeFrontIds\":[{\"distance\":1.83,\"inStore\":false,\"preferred\":false,\"storeId\":\"1988\",\"storeUUID\":null,\"usStoreId\":1988},{\"distance\":2.65,\"inStore\":false,\"preferred\":false,\"storeId\":\"3090\",\"storeUUID\":null,\"usStoreId\":3090},{\"distance\":3.93,\"inStore\":false,\"preferred\":false,\"storeId\":\"3587\",\"storeUUID\":null,\"usStoreId\":3587},{\"distance\":4.53,\"inStore\":false,\"preferred\":false,\"storeId\":\"4202\",\"storeUUID\":null,\"usStoreId\":4202},{\"distance\":5.42,\"inStore\":false,\"preferred\":false,\"storeId\":\"5979\",\"storeUUID\":null,\"usStoreId\":5979}],\"productId\":\"%s\",\"selected\":false}",
	  		t.Profile.ShippingAddress.Zip, t.Profile.ShippingAddress.State, pid),
	}
	  body, err := json.Marshal(bodyStructure)
	  if err != nil {
	    return err
	  }
	  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
	  defer DiscardResp(resp)
	  if err != nil {
	  	// DiscardResp(resp)
	  	if resp != nil && resp.StatusCode == 412 {
	  					if bbb, err := readBodyBytes(resp); err == nil {
  if err := t.WmtSolveCaptcha(false, bbb); err != nil {
            return err
          }
        }
	  					return ErrRetrying
	  	}
	    return err
	  }

	  err = readRespJsonDst(resp, &product)
	  if err != nil {
	    return err
	  }
	  offers := product.Data.ProductByProductID.OfferList
	  if len(offers) == 0 {
	  	return ErrProductUnavailable
	  }

	  if offers[0].Fulfillment.Shippable && offers[0].SellerID == WMT_SELLER_ID && os.Getenv("OOS_SIM") == "" { // walmart.com
	  	return nil
	  } else {
	  	if maxret, err := strconv.Atoi(os.Getenv("OOS_SIM")); err == nil {
	  		if retries >= maxret {
	  			return nil
	  		}
	  	}
	  	return ErrProductUnavailable
	  }
	 })
	 return &product, err
}

var WmtCheckoutUrl, _ = url.Parse("https://www.walmart.com/checkout/")

func (t *CheckoutTask) WmtStartCheckout() (*WmtCheckoutSession, error) {
	var sesh WmtCheckoutSession
	err := t.Retry(func() error {
		// resp, err := t.doReq(t.client, t.makeGetReq(WmtCheckoutUrl, nil))
		// DiscardResp(resp)
		// if err != nil {
		// 	return err
		// }
		if t.GetCookieValue("CRT") == "" {
			return retry.Unrecoverable(ErrCartJacked)
		}

	  url_, err := url.Parse("https://www.walmart.com/api/checkout/v3/contract?page=CHECKOUT_VIEW")
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "content-length",
	    "accept",
	    "user-agent",
	    "wm_vertical_id",
	    "content-type",
	    "origin",
	    "sec-fetch-site",
	    "sec-fetch-mode",
	    "sec-fetch-dest",
	    "referer",
	    "accept-encoding",
	    "accept-language",
	    "cookie",
	  }
	  headers := [][2]string {
	    {"accept", "application/json, text/javascript, */*; q=0.01"},
	    {"wm_vertical_id", "0"},
	    {"content-type", "application/json"},
	    {"origin", "https://www.walmart.com"},
	    {"sec-fetch-site", "same-origin"},
	    {"sec-fetch-mode", "cors"},
	    {"sec-fetch-dest", "empty"},
	    {"referer", "https://www.walmart.com/checkout/"},
	    {"accept-encoding", "gzip, deflate, br"},
	    {"accept-language", "en-US,en;q=0.9"},
	  }
	  bodyStructure := map[string]interface{} {
	  "storeList": []interface{} {
	  },
	  "postalCode": t.Profile.ShippingAddress.Zip,
	  "city": t.Profile.ShippingAddress.City,
	  "state": t.Profile.ShippingAddress.State,
	  "isZipLocated": true,
	  "crt:CRT": "",
	  "customerId:CID": "",
	  "customerType:type": "",
	  "affiliateInfo:com.wm.reflector": "",
	}
	  body, err := json.Marshal(bodyStructure)
	  if err != nil {
	    return err
	  }
	  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
	  defer DiscardResp(resp)
	  if err != nil {
	  	if resp != nil && resp.StatusCode == 412 {
				if bbb, err := readBodyBytes(resp); err == nil {
  if err := t.WmtSolveCaptcha(false, bbb); err != nil {
            return err
          }
        }
				return ErrRetrying
	  	}
	  	return err
	  }
	  err = readRespJsonDst(resp, &sesh)
	  if (resp == nil || resp.StatusCode != 201) || sesh.ID == "" {
	  	return ErrRetrying
	  }
	  return err
	 }, retry.Delay(t.Delay), retry.Attempts(32))
	 return &sesh, err
}

 func (t *CheckoutTask) WmtStartCheckout2() error {
 	return t.Retry(func() error {
	 	url_, err := url.Parse("https://www.walmart.com/api/checkout/v3/contract/:PCID?page=CHECKOUT_VIEW")
	 	if err != nil {
	 	  return err
	 	}
	 	headerOrder := []string {
	 	  "accept",
	 	  "wm_cvv_in_session",
	 	  "user-agent",
	 	  "wm_vertical_id",
	 	  "content-type",
	 	  "sec-fetch-site",
	 	  "sec-fetch-mode",
	 	  "sec-fetch-dest",
	 	  "referer",
	 	  "accept-encoding",
	 	  "accept-language",
	 	  "cookie",
	 	}
	 	headers := [][2]string {
	 	  {"accept", "application/json, text/javascript, */*; q=0.01"},
	 	  {"wm_cvv_in_session", "true"},
	 	  {"wm_vertical_id", "0"},
	 	  {"content-type", "application/json"},
	 	  {"sec-fetch-site", "same-origin"},
	 	  {"sec-fetch-mode", "cors"},
	 	  {"sec-fetch-dest", "empty"},
	 	  {"referer", "https://www.walmart.com/checkout/"},
	 	  {"accept-encoding", "gzip, deflate, br"},
	 	  {"accept-language", "en-US,en;q=0.9"},
	 	}
	 	req := t.makeReq("GET", url_, &headers, &headerOrder, nil)
	 	req.IgnoreBody = true
	 	req.IgnoreBodyExceptStatus = []int{307,401}
	 	resp, err := t.doReq(t.client, req)
	 	if resp != nil {
	 		defer DiscardResp(resp)
	 		if resp.StatusCode == 307 {
	 			  			pxBlock := DEFAULT_BR_STRUCT
	 			  			pxBlock.RedirectURL = resp.Header.Get("location")
	 			  			if br, err := json.Marshal(pxBlock); err == nil {
	 					  		if err := t.WmtSolveCaptcha(false, br); err != nil {
	 									return err
	 								}
	 			  			}
	 		} else if resp.StatusCode == 401 {
				if lpw := t.VaultGet(t.Url.Host, t.Profile.Email); lpw != nil {
					if err := t.WalmartLogin(t.Profile.Email, *lpw); err == nil {
						t.WmtStartCheckout()
					}
				}
				return ErrRetrying
	 		} else if resp.StatusCode == 200 {
	 			return nil
	 		}
	 	}
	 	if err != nil {
	 		return err
	 	} else {
	 		return ErrRetrying
	 	}
	 }, retry.Attempts(16))
}
type WmtCheckoutSession struct {
	DbSessionTokenMap struct {
	} `json:"dbSessionTokenMap"`
	ID               string `json:"id"`
	CheckoutFlowType string `json:"checkoutFlowType"`
	CartID           string `json:"cartId"`
	Items            []struct {
		ID               string  `json:"id"`
		OfferID          string  `json:"offerId"`
		ProductID        string  `json:"productId"`
		ProductName      string  `json:"productName"`
		ItemID           int     `json:"itemId"`
		SellerID         string  `json:"sellerId"`
		ThumbnailURL     string  `json:"thumbnailUrl"`
		LegacySellerID   int     `json:"legacySellerId"`
		ProductClassType string  `json:"productClassType"`
		Quantity         int     `json:"quantity"`
		UnitPrice        float32 `json:"unitPrice"`
		Type             string  `json:"type"`
		Price            float32 `json:"price"`
		UnitOfMeasure    string  `json:"unitOfMeasure"`
		HasCarePlan      bool    `json:"hasCarePlan"`
		Brand            string  `json:"brand"`
		Discount         struct {
		} `json:"discount"`
		RhPath                      string `json:"rhPath"`
		IsWarrantyEligible          bool   `json:"isWarrantyEligible"`
		Category                    string `json:"category"`
		PrimaryCategory             string `json:"primaryCategory"`
		IsCarePlan                  bool   `json:"isCarePlan"`
		IsEgiftCard                 bool   `json:"isEgiftCard"`
		IsAssociateDiscountEligible bool   `json:"isAssociateDiscountEligible"`
		IsShippingPassEligible      bool   `json:"isShippingPassEligible"`
		ShippingTier                string `json:"shippingTier"`
		IsTwoDayShippingEligible    bool   `json:"isTwoDayShippingEligible"`
		MeetsSLA                    bool   `json:"meetsSla"`
		ClassID                     string `json:"classId"`
		MaxQuantityPerOrder         int    `json:"maxQuantityPerOrder"`
		IsSubstitutable             bool   `json:"isSubstitutable"`
		IsInstaWatch                bool   `json:"isInstaWatch"`
		IsAlcoholic                 bool   `json:"isAlcoholic"`
		IsSnapEligible              bool   `json:"isSnapEligible"`
		IsAgeRestricted             bool   `json:"isAgeRestricted"`
		IsSubstitutionsAllowed      bool   `json:"isSubstitutionsAllowed"`
		FulfillmentSelection        struct {
			FulfillmentOption string `json:"fulfillmentOption"`
			ShipMethod        string `json:"shipMethod"`
			AvailableQuantity int    `json:"availableQuantity"`
		} `json:"fulfillmentSelection"`
		ServicePlanType string        `json:"servicePlanType"`
		Errors          []interface{} `json:"errors"`
		WfsEnabled      bool          `json:"wfsEnabled"`
	} `json:"items"`
	Shipping struct {
		PostalCode string `json:"postalCode"`
		City       string `json:"city"`
		State      string `json:"state"`
	} `json:"shipping"`
	Summary struct {
		SubTotal               float32 `json:"subTotal"`
		ShippingIsEstimate     bool    `json:"shippingIsEstimate"`
		TaxTotal               float32 `json:"taxTotal"`
		TaxIsEstimate          bool    `json:"taxIsEstimate"`
		GrandTotal             float32 `json:"grandTotal"`
		QuantityTotal          int     `json:"quantityTotal"`
		AmountOwed             float32 `json:"amountOwed"`
		MerchandisingFeesTotal int     `json:"merchandisingFeesTotal"`
		ShippingCosts          []struct {
			Label  string  `json:"label"`
			Type   string  `json:"type"`
			Cost   float32 `json:"cost"`
			Method string  `json:"method"`
		} `json:"shippingCosts"`
		ShippingTotal    float32 `json:"shippingTotal"`
		HasSurcharge     bool    `json:"hasSurcharge"`
		PreTaxTotal      float32 `json:"preTaxTotal"`
		PromotionSummary struct {
		} `json:"promotionSummary"`
		AddOnServicesTotal int     `json:"addOnServicesTotal"`
		ItemsSubTotal      float32 `json:"itemsSubTotal"`
	} `json:"summary"`
	PickupPeople []interface{} `json:"pickupPeople"`
	Email        string        `json:"email"`
	Buyer        struct {
		CustomerAccountID      string `json:"customerAccountId"`
		IsGuestSignupRequired  bool   `json:"isGuestSignupRequired"`
		IsGuest                bool   `json:"isGuest"`
		IsAssociate            bool   `json:"isAssociate"`
		ApplyAssociateDiscount bool   `json:"applyAssociateDiscount"`
	} `json:"buyer"`
	AllowedPaymentTypes []struct {
		Type        string `json:"type"`
		CvvRequired bool   `json:"cvvRequired"`
	} `json:"allowedPaymentTypes"`
	Registries                []interface{} `json:"registries"`
	Payments                  []interface{} `json:"payments"`
	CardsToDisable            []interface{} `json:"cardsToDisable"`
	AllowedPaymentPreferences []interface{} `json:"allowedPaymentPreferences"`
	IsRCFEligible             bool          `json:"isRCFEligible"`
	IsMarketPlaceItemsExist   bool          `json:"isMarketPlaceItemsExist"`
	Version                   string        `json:"version"`
	ShippingCategory          struct {
		ShippingGroups []struct {
			ItemIds              []string `json:"itemIds"`
			Seller               string   `json:"seller"`
			ShippingTier         string   `json:"shippingTier"`
			MeetSLA              bool     `json:"meetSla"`
			DefaultSelection     bool     `json:"defaultSelection"`
			FulfillmentOption    string   `json:"fulfillmentOption"`
			ShippingGroupOptions []WmtShipGroupOption  `json:"shippingGroupOptions"`
			IsEdelivery      bool          `json:"isEdelivery"`
			HasWFSItem       bool          `json:"hasWFSItem"`
			ItemSellerGroups []interface{} `json:"itemSellerGroups"`
		} `json:"shippingGroups"`
	} `json:"shippingCategory"`
	BalanceToReachThreshold float32       `json:"balanceToReachThreshold"`
	EntityErrors            []interface{} `json:"entityErrors"`
	OneDaySelected          bool          `json:"oneDaySelected"`
	PaymentWithBagFee       bool          `json:"paymentWithBagFee"`
	GiftDetails             struct {
		GiftOrder           bool `json:"giftOrder"`
		HasGiftEligibleItem bool `json:"hasGiftEligibleItem"`
	} `json:"giftDetails"`
	CanApplyDetails []interface{} `json:"canApplyDetails"`
	Jwt             string        `json:"jwt"`
}

type WmtShipGroupOption struct {
				Method                string  `json:"method"`
				MethodDisplay         string  `json:"methodDisplay"`
				Selected              bool    `json:"selected"`
				Charge                float32 `json:"charge"`
				DeliveryDate          int64   `json:"deliveryDate"`
				AvailableDate         int64   `json:"availableDate"`
				FulfillmentOption     string  `json:"fulfillmentOption"`
				OnlineStoreID         int64     `json:"onlineStoreId"`
				IsThresholdShipMethod bool    `json:"isThresholdShipMethod"`
			}

type WmtErrorResponse struct {
	Message    string `json:"message"`
}

func (t *CheckoutTask) WmtFulfillment(sesh *WmtCheckoutSession, shipMethod string) (error) {
	return t.Retry(func() error {
	  url_, err := url.Parse("https://www.walmart.com/api/checkout/v3/contract/:PCID/fulfillment")
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "content-length",
	    "accept",
	    "inkiru_precedence",
	    "user-agent",
	    "wm_vertical_id",
	    "content-type",
	    "origin",
	    "sec-fetch-site",
	    "sec-fetch-mode",
	    "sec-fetch-dest",
	    "referer",
	    "accept-encoding",
	    "accept-language",
	    "cookie",
	  }
	  headers := [][2]string {
	    {"accept", "application/json, text/javascript, */*; q=0.01"},
	    {"inkiru_precedence", "false"},
	    {"wm_vertical_id", "0"},
	    {"content-type", "application/json"},
	    {"origin", "https://www.walmart.com"},
	    {"sec-fetch-site", "same-origin"},
	    {"sec-fetch-mode", "cors"},
	    {"sec-fetch-dest", "empty"},
	    {"referer", "https://www.walmart.com/checkout/"},
	    {"accept-encoding", "gzip, deflate, br"},
	    {"accept-language", "en-US,en;q=0.9"},
	  }
	  t.LogDebug("%+v", sesh.ShippingCategory.ShippingGroups[0].ShippingGroupOptions)
	  var shipOption WmtShipGroupOption
	  for _, sg := range sesh.ShippingCategory.ShippingGroups {
	  	for _, sgo := range sg.ShippingGroupOptions {
		  	if sgo.Method == shipMethod {
		  		shipOption = sgo
		  		break
		  	}
	  	}
	  }
	  if shipOption.Method == "" {
	  	t.StatusCh <- StatusShipOptionUnavailable
	  	return ErrShipOptionUnavailable
	  }
	  t.LogDebug("%+v", shipOption)

	  bodyStructure := map[string]interface{} {
	  "groups": []interface{} {
	    map[string]interface{} {
	          "fulfillmentOption": shipOption.FulfillmentOption,
	          "itemIds": []interface{} {
	            sesh.Items[0].ID,
	          },
	          "shipMethod": shipOption.Method,
	        },
	  },
	}
	  body, err := json.Marshal(bodyStructure)
	  if err != nil {
	    return err
	  }
	  req := t.makeReq("POST", url_, &headers, &headerOrder, &body)
	  req.IgnoreBody = true
	  req.IgnoreBodyExceptStatus = []int{412}
		resp, err := t.doReq(t.client, req)
		defer DiscardResp(resp)
	  if resp != nil && resp.StatusCode != 200 {
	  	if resp != nil && resp.StatusCode == 412 {
		  		// DiscardResp(resp)
					if bbb, err := readBodyBytes(resp); err == nil {
  if err := t.WmtSolveCaptcha(false, bbb); err != nil {
            return err
          }
        }
					return ErrRetrying
		  	}

	  	var errResp WmtErrorResponse
	  	err := readRespJsonDst(resp, &errResp)
	  	if err == nil {
	  		return errors.New(errResp.Message)
	  	}
	  }
	  return err
	}, retry.OnRetry(func (n uint, err error) {
		t.StatusCh <- Status{Message: fmt.Sprintf("Waiting to Start Checkout (%v)", err)}
	}), retry.Delay(t.Delay), retry.Attempts(32))
}

func (t *CheckoutTask) WmtSetShipping() (error) {
	return t.Retry(func() error {

	  url_, err := url.Parse("https://www.walmart.com/api/checkout/v3/contract/:PCID/shipping-address")
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "content-length",
	    "accept",
	    "inkiru_precedence",
	    "user-agent",
	    "wm_vertical_id",
	    "content-type",
	    "origin",
	    "sec-fetch-site",
	    "sec-fetch-mode",
	    "sec-fetch-dest",
	    "referer",
	    "accept-encoding",
	    "accept-language",
	    "cookie",
	  }
	  headers := [][2]string {
	    {"accept", "application/json, text/javascript, */*; q=0.01"},
	    {"inkiru_precedence", "false"},
	    {"wm_vertical_id", "0"},
	    {"content-type", "application/json"},
	    {"origin", "https://www.walmart.com"},
	    {"sec-fetch-site", "same-origin"},
	    {"sec-fetch-mode", "cors"},
	    {"sec-fetch-dest", "empty"},
	    {"referer", "https://www.walmart.com/checkout/"},
	    {"accept-encoding", "gzip, deflate, br"},
	    {"accept-language", "en-US,en;q=0.9"},
	  }
	  bodyStructure := map[string]interface{} {
	  "addressLineOne": t.Profile.ShippingAddress.Address1,
	  "addressLineTwo": t.Profile.ShippingAddress.Address2,
	  "addressType": "RESIDENTIAL",
	  "city": strings.ToUpper(t.Profile.ShippingAddress.City),
	  "firstName": t.Profile.ShippingAddress.FirstName,
	  "lastName": t.Profile.ShippingAddress.LastName,
	  "phone": t.Profile.ShippingAddress.NordPhone(),
	  "email": t.Profile.Email,
	  "marketingEmailPref": false,
	  "postalCode": t.Profile.ShippingAddress.Zip,
	  "state": t.Profile.ShippingAddress.State,
	  "countryCode": "USA",
	  "changedFields": []interface{} {
	  	7,
	  },
	  "storeList": []interface{} {

	  },
	}
	  body, err := json.Marshal(bodyStructure)
	  if err != nil {
	    return err
	  }
	  req := t.makeReq("POST", url_, &headers, &headerOrder, &body)
	  req.IgnoreBody = true
	  req.IgnoreBodyExceptStatus = []int{412}
		resp, err := t.doReq(t.client, req)
		defer DiscardResp(resp)
	  if resp != nil {
	  	if resp != nil && resp.StatusCode == 412 {
		  		if resp.StatusCode == 412 {
						if bbb, err := readBodyBytes(resp); err == nil {
  if err := t.WmtSolveCaptcha(false, bbb); err != nil {
            return err
          }
        }
						return ErrRetrying
	  			}
		  	}
	  	io.Copy(ioutil.Discard, resp.Body)
	  	resp.Body.Close()
		  if resp.StatusCode != 200 {
		  	err = ErrRetrying
		  }
	  }
	  return err
	}, retry.OnRetry(func (n uint, err error) {
	 		t.StatusCh <- Status{Message: fmt.Sprintf("Waiting to Set Shipping (%v)", err)}
	 	}), retry.Delay(t.Delay), retry.Attempts(32))
}

type WmtPayKey struct {
	KeyId string `json:"key_id"`
	L int
	E int
	K string
	Phase int
}

func (t *CheckoutTask) WmtGetPayKey() (*WmtPayKey, error) {
	client, err := t.newHttpClient()
	if err != nil {
		return nil, err
	}
	var payKey WmtPayKey
	err = t.Retry(func() error {
	  url_, err := url.Parse(fmt.Sprintf("https://securedataweb.walmart.com/pie/v1/wmcom_us_vtg_pie/getkey.js?bust=%d", timeMillis()))
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "Host",
	    "Connection",
	    "User-Agent",
	    "Accept",
	    "Sec-Fetch-Site",
	    "Sec-Fetch-Mode",
	    "Sec-Fetch-Dest",
	    "Referer",
	    "Accept-Encoding",
	    "Accept-Language",
	    "Cookie",
	  }
	  headers := [][2]string {
	    {"Connection", "keep-alive"},
	    {"Accept", "*/*"},
	    {"Sec-Fetch-Site", "same-site"},
	    {"Sec-Fetch-Mode", "no-cors"},
	    {"Sec-Fetch-Dest", "script"},
	    {"Referer", "https://www.walmart.com/checkout/"},
	    {"Accept-Encoding", "gzip, deflate, br"},
	    {"Accept-Language", "en-US,en;q=0.9"},
	  }
	  req := t.makeReq("GET", url_, &headers, &headerOrder, nil)
	  req.Close = true
	  resp, err := t.doReq(client, req)
	  if err != nil {
	  	if resp != nil && resp.StatusCode == 412 {
	  		DiscardResp(resp)
	  		t.StatusCh <- StatusPxBanned
	  		return ErrPxBanned
	  	}
	  	return err
	  }
	  body, err := readBody(resp)
	  if err != nil {
	  	return err
	  }
	  re := regexp.MustCompile(`PIE\.(.+?) = (.+?);`)
	  matches := re.FindAllStringSubmatch(body, -1)
	  if matches == nil {
	  	return errors.New("Missing payment key fields")
	  }

	  for _, match := range matches {
	  	if match[1] == "L" {
	  		payKey.L, _ = strconv.Atoi(match[2])
	  	} else if match[1] == "E" {
	  		payKey.E, _ = strconv.Atoi(match[2])
	  	} else if match[1] == "K" {
	  		payKey.K = strings.Replace(match[2], "\"", "", -1)
	  	} else if match[1] == "key_id" {
	  		payKey.KeyId = strings.Replace(match[2], "\"", "", -1)
	  	} else if match[1] == "phase" {
	  		payKey.Phase, _ = strconv.Atoi(match[2])
	  	}
	  }
	  return nil
	}, retry.Attempts(5), retry.Delay(t.Delay))

  return &payKey, err
}

func (t *CheckoutTask) WmtSetCard(cardToken *WmtCardToken, payKey *WmtPayKey) (error) {
  url_, err := url.Parse("https://www.walmart.com/api/checkout-customer/:CID/credit-card")
  if err != nil {
    return err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "user-agent",
    "content-type",
    "origin",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "application/json"},
    {"content-type", "application/json"},
    {"origin", "https://www.walmart.com"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", "https://www.walmart.com/checkout/"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  bodyStructure := map[string]interface{} {
  "encryptedPan": cardToken.NumEnc,
  "encryptedCvv": cardToken.CvcEnc,
  "integrityCheck": cardToken.Integrity,
  "keyId": payKey.KeyId,
  "phase": fmt.Sprintf("%d", payKey.Phase),
  "state": t.Profile.BillingAddress.State,
  "postalCode": t.Profile.BillingAddress.Zip,
  "addressLineOne": t.Profile.BillingAddress.Address1,
  "addressLineTwo": t.Profile.BillingAddress.Address2,
  "city": strings.ToUpper(t.Profile.BillingAddress.City),
  "firstName": t.Profile.BillingAddress.FirstName,
  "lastName": t.Profile.BillingAddress.LastName,
  "expiryMonth": fmt.Sprintf("%02d", t.Profile.Card.ExpMonth),
  "expiryYear": fmt.Sprintf("%d", t.Profile.Card.ExpYear),
  "phone": t.Profile.BillingAddress.NordPhone(),
  "cardType": t.WmtCardType(),
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  req := t.makeReq("POST", url_, &headers, &headerOrder, &body)
  req.Close = false
	resp, err := t.doReq(t.client, req)
	defer DiscardResp(resp)
  if resp != nil {
  	if resp.StatusCode == 403 {
  		if lpw := t.VaultGet(t.Url.Host, t.Profile.Email); lpw != nil {
  			err = t.WalmartLogin(t.Profile.Email, *lpw)
  			if err != nil {
  				// t.StatusCh <- Status{Message: "Error logging in"}
  				return err
  			}
  		}
  		return ErrRetrying
  	}
  	if resp.StatusCode == 412 {
			if resp.StatusCode == 412 {
				if bbb, err := readBodyBytes(resp); err == nil {
  if err := t.WmtSolveCaptcha(false, bbb); err != nil {
            return err
          }
        }
				return ErrRetrying
	  	}
		}
		// if resp.StatusCode !=
  	io.Copy(ioutil.Discard, resp.Body)
  	resp.Body.Close()
  }
  return err
}

func (t *CheckoutTask) WmtCardType() string {
	if t.Profile.Card.Type == "MASTER" {
		return "MASTERCARD"
	} else {
		return t.Profile.Card.Type
	}
}

func (t *CheckoutTask) WmtPayment(cardToken *WmtCardToken, payKey *WmtPayKey) (error) {
	return t.Retry(func() error {
	  url_, err := url.Parse("https://www.walmart.com/api/checkout/v3/contract/:PCID/payment")
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "content-length",
	    "accept",
	    "inkiru_precedence",
	    "wm_cvv_in_session",
	    "user-agent",
	    "wm_vertical_id",
	    "content-type",
	    "origin",
	    "sec-fetch-site",
	    "sec-fetch-mode",
	    "sec-fetch-dest",
	    "referer",
	    "accept-encoding",
	    "accept-language",
	    "cookie",
	  }
	  headers := [][2]string {
	    {"accept", "application/json, text/javascript, */*; q=0.01"},
	    {"inkiru_precedence", "false"},
	    {"wm_cvv_in_session", "true"},
	    {"wm_vertical_id", "0"},
	    {"content-type", "application/json"},
	    {"origin", "https://www.walmart.com"},
	    {"sec-fetch-site", "same-origin"},
	    {"sec-fetch-mode", "cors"},
	    {"sec-fetch-dest", "empty"},
	    {"referer", "https://www.walmart.com/checkout/"},
	    {"accept-encoding", "gzip, deflate, br"},
	    {"accept-language", "en-US,en;q=0.9"},
	  }
	  bodyStructure := map[string]interface{} {
	  "payments": []interface{} {
	    map[string]interface{} {
	          "paymentType": "CREDITCARD",
	          "cardType": t.WmtCardType(),
	          "firstName": t.Profile.BillingAddress.FirstName,
	          "lastName": t.Profile.BillingAddress.LastName,
	          "addressLineOne": t.Profile.BillingAddress.Address1,
	          "addressLineTwo": t.Profile.BillingAddress.Address2,
	          "city": t.Profile.BillingAddress.City,
	          "state": t.Profile.BillingAddress.State,
	          "postalCode": t.Profile.BillingAddress.Zip,
	          "expiryMonth": fmt.Sprintf("%02d", t.Profile.Card.ExpMonth),
	          "expiryYear": fmt.Sprintf("%d", t.Profile.Card.ExpYear),
	          "email": t.Profile.Email,
	          "phone": t.Profile.BillingAddress.NordPhone(),
	          "encryptedPan": cardToken.NumEnc,
	          "encryptedCvv": cardToken.CvcEnc,
	          "integrityCheck": cardToken.Integrity,
	          "keyId": payKey.KeyId,
	          "phase": fmt.Sprintf("%d", payKey.Phase),
	          "piHash": "", // todo need?
	          // "piHash": "PIH.ccdb.VISA.CREDITCARD.1123177.4242",
	        },
	  },
	}
	  body, err := json.Marshal(bodyStructure)
	  if err != nil {
	    return err
	  }
	  t.EnablePrivateHttpTrace = true
	  defer func() {
	  	t.EnablePrivateHttpTrace = false
	  }()
	  req := t.makeReq("POST", url_, &headers, &headerOrder, &body)
	  req.Close = false
	  req.IgnoreBody = true
	  req.IgnoreBodyExceptStatus = []int{412}
	  resp, err := t.doReq(t.client, req)
	  defer DiscardResp(resp)
	  if resp != nil && resp.StatusCode == 403 {
  		if lpw := t.VaultGet(t.Url.Host, t.Profile.Email); lpw != nil {
  			err = t.WalmartLogin(t.Profile.Email, *lpw)
  			if err != nil {
  				// t.StatusCh <- Status{Message: "Error logging in"}
  				return err
  			}
  		}
  		return ErrRetrying
  	}
	  if resp != nil {
	  	if resp.StatusCode != 200 {
	  		err = ErrRetrying
	  	}
	  	if resp.StatusCode == 412 {
	  		if resp.StatusCode == 412 {
	  						if bbb, err := readBodyBytes(resp); err == nil {
  if err := t.WmtSolveCaptcha(false, bbb); err != nil {
            return err
          }
        }
	  						return ErrRetrying
	  			  	}
	  	}
	  	resp.Body.Close()
	  }
	  return err
	}, retry.OnRetry(func (n uint, err error) {
	 		t.StatusCh <- Status{Message: fmt.Sprintf("Waiting to Set Billing (%v)", err)}
	 	}), retry.Delay(t.Delay), retry.DelayType(retry.FixedDelay), retry.Attempts(32))
}

func (t *CheckoutTask) WmtOrder(cardToken *WmtCardToken, payKey *WmtPayKey) (error) {
	return t.Retry(func() error {
	  url_, err := url.Parse("https://www.walmart.com/api/checkout/v3/contract/:PCID/order")
	  if err != nil {
	    return err
	  }
	   headerOrder := []string {
    "content-length",
    "accept",
    "inkiru_precedence",
    "wm_cvv_in_session",
    "user-agent",
    "wm_vertical_id",
    "content-type",
    "origin",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "application/json, text/javascript, */*; q=0.01"},
    {"inkiru_precedence", "false"},
    {"wm_cvv_in_session", "true"},
    {"wm_vertical_id", "0"},
    {"content-type", "application/json"},
    {"origin", "https://www.walmart.com"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", "https://www.walmart.com/checkout/"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
	   bodyStructure := map[string]interface{} {
	   	"cvvInSession": true,
	  "voltagePayments": []interface{} {
	    map[string]interface{} {
	          "paymentType": "CREDITCARD",
	          "encryptedPan": cardToken.NumEnc,
	          "encryptedCvv": cardToken.CvcEnc,
	          "integrityCheck": cardToken.Integrity,
	          "keyId": payKey.KeyId,
	          "phase": fmt.Sprintf("%d", payKey.Phase),
	        },
	  },
	}
	  body, err := json.Marshal(bodyStructure)
	  if err != nil {
	    return err
	  }

	  if t.OneCheckoutPerProf && t.CheckHit != nil && t.CheckHit(t.Url.Host, t.Profile.Name) > 0 {
	  	t.StatusCh <- Status{Message: "Stopping Task, Already Hit with Profile"}
	  	return retry.Unrecoverable(retry.Unrecoverable(errors.New("stopping task, already hit")))
	  }

	  req := t.makeReq("PUT", url_, &headers, &headerOrder, &body)
	  req.Close = false
	  resp, err := t.doReq(t.client, req)
	  if resp != nil {
		  if resp.StatusCode == 403 {
	  		if lpw := t.VaultGet(t.Url.Host, t.Profile.Email); lpw != nil {
	  			err = t.WalmartLogin(t.Profile.Email, *lpw)
	  			if err != nil {
	  				// t.StatusCh <- Status{Message: "Error logging in"}
	  				return err
	  			}
	  		}
	  		return ErrRetrying
	  	}
	  	if resp.StatusCode == 444 {
	  		return retry.Unrecoverable(ErrRetrying)
	  	}
	  	if resp.StatusCode == 412 {
	  		if resp.StatusCode == 412 {
				if bbb, err := readBodyBytes(resp); err == nil {
  if err := t.WmtSolveCaptcha(false, bbb); err != nil {
            return err
          }
        }
				return ErrRetrying
	  	}
	  	}
	  	body, _ := readBody(resp)
	  	if resp.StatusCode == 400 {
	  		if strings.Contains(body, "out_of_stock") {
	  			t.StatusCh <- StatusOutOfStock
	  			return retry.Unrecoverable(ErrOutOfStock)
	  		}
	  		if strings.Contains(body, "payment_service_invalid_account_no") || strings.Contains(body, "payment_service_authorization_decline") || strings.Contains(body, "invalid_credit_card_auth_info_input") {
	  			return retry.Unrecoverable(errDecline)
	  		} else {
	  			go t.SendTelemetry(map[string]interface{} {
	  				"event": "walmart_order_decline",
	  				"body": body,
	  			})
	  		}
	  	} else if strings.Contains(body, "orderId") && (resp.StatusCode == 200 || resp.StatusCode == 201) {
	  		if t.AddHit != nil {
	  			t.AddHit(t.Url.Host, t.Profile.Name)
	  		}
	  		return nil
	  	}
		  go t.SendTelemetry(map[string]interface{} {
		  	"event": "wmt_unknown_checkout_resp",
		  	"wmt_checkout_resp": body,
		  	"status_code": resp.StatusCode,
		  })
	  }
	  return ErrUnknownCheckoutError
	 }, retry.OnRetry(func (n uint, err error) {
	 		t.StatusCh <- Status{Message: fmt.Sprintf("Waiting to Order (%v)", err)}
	 	}), retry.Delay(t.Delay), retry.Attempts(32))
}

func (t *CheckoutTask) WmtGenPxCookie() ([]*http.Cookie, error) {

	var proxy string
	if t.Proxy != nil {
		proxy = t.Proxy.String()
	}
	// ckies, err := GetPxCookie(Px3Request{
	return t.GetPxCookieApi(Px3Request{
		Proxy: proxy,
		Cookie: t.CookiesStr(),
		Url: t.Url.String(),
		// TODO cookies. can regenning help avoid 8min limit on checkout poll?
		AppId: "PXu6b0qd2S",
		Vid: t.GetCookieValue("_pxvid"),
		Uuid: "",
		Host: "",
		JsSrc: "https://www.walmart.com/px/PXu6b0qd2S/init.js",
		CapJsSrc: "",
	})

	// log.Printf("%+v (error=%+v)", ckies, err)
	// if err != nil {
	// 	return err
	// }
	// return ckies, err

	// for _, ckie := range ckies {
	// 	// t.SetCookie(ckie)
	// 	if ckie.Name != "pxsid" {
	// 		t.SetCookie(ckie)
	// 	}
	// 	// if ckie.Name == "_px3" {
	// 	// 	// ckie.Value = "d0cbc921fda806178d742f83ed5a2c6615329da751770db4dd25bfd6acad7723:y5WxzFxCYDRjAW8aAw1EaDTYiHPGaksK+U2iBF9wx3I2CYN/pGMG+2u+H/ARTn3oM1a0H+DFP7k3n0UJbXvCVg==:1000:yYCpOwtvUSED7O9KgP76TEqHb/09MyUqPKO9IGA77TKh6sjoeKqRv791qjy1BiUQyidxGEWXcqBc9q+i3ndBS5yllYIx+y0mBxtJ91Rtk8adcb432dp7IhVM4fW6s+XzNxfHkIMd72BuC/m1L+JnexdCPJDpoiOgkNLpBgDCpl4="
	// 	// 	t.SetCookie(ckie)
	// 	// }
	// }

	// return nil
}

func (t *CheckoutTask) WmtSolveCaptchaCkies(quiet bool, blockResponse []byte) ([]*http.Cookie, error) {
	if !quiet {
		t.StatusCh <- StatusSolvingCap
	}
	bu, _ := url.Parse(fmt.Sprintf("https://www.walmart.com/blocked?url=%s&uuid=vid=&g=a", base32.StdEncoding.EncodeToString([]byte(t.Url.Path))))
	var err error
	var recapToken string
	ckies := []*http.Cookie{}
	if recapToken, err = t.SolveCaptcha("6Lcj-R8TAAAAABs3FrRPuQhLMbp5QrHsHufzLf7b", bu.String()); err == nil {
		// if solves, err := strconv.Atoi(t.Config["solves"]); err == nil {
		// 	solves += 1
		// 	if solves >  {
		// 		t.StatusCh <- Status{Message: "Stuck in captcha loop, backing off"}
		// 		time.Sleep(60*time.Second)
		// 		t.Config["solves"] = "0"
		// 		return TaskErrUnrecoverable(ErrFailedToSolveCaptcha)
		// 	}
		// 	t.Config["solves"] = fmt.Sprintf("%d", solves)
		// } else {
		// 	t.Config["solves"] = "1"
		// }

		// log.Printf("recap token %s", recapToken)
		err = t.Retry(func() error {
			var pxBlock PxBlockResponse
			var err error
			json.Unmarshal(blockResponse, &pxBlock)
			if ckies, err = t.GetPxCookieApi(Px3Request{
				Proxy: t.ProxyStr,
				Cookie: t.CookiesStr(),
				Url: "https://" + t.Url.Host + pxBlock.RedirectURL,
				AppId: pxBlock.AppID,
				Vid: pxBlock.Vid,
				Uuid: pxBlock.UUID,
				Host: pxBlock.HostURL,
				JsSrc: "https://" + t.Url.Host + pxBlock.JsClientSrc,
				CapJsSrc: "https://" + t.Url.Host + pxBlock.BlockScript,
				RecapToken: recapToken,
			}); err == nil {
				t.LogDebug("got cookies %+v", ckies)
				return nil
				// for _, ckie := range ckies {
				// 	t.SetCookie(ckie)
				// }
				// if !quiet {
				// 	t.StatusCh <- StatusCapSolved
				// }
				// return nil
			} else {
				t.LogDebug("gen err %+v", err)
			}
			return err
		}, retry.Attempts(4))
		if err == nil {
			return ckies, nil
		}
	}

	if !quiet {
		t.StatusCh <- StatusCaptchaSolveFailedRetrying
	}
	return ckies, err
}

func (t *CheckoutTask) WmtSolveCaptcha(quiet bool, blockResponse []byte) error {
	t.RemoveBadSession()
	if !quiet {
		t.StatusCh <- StatusSolvingCap
	}
	bu, _ := url.Parse(fmt.Sprintf("https://www.walmart.com/blocked?url=%s&uuid=vid=&g=a", base32.StdEncoding.EncodeToString([]byte(t.Url.Path))))
	var err error
	var recapToken string
	if recapToken, err = t.SolveCaptcha("6Lcj-R8TAAAAABs3FrRPuQhLMbp5QrHsHufzLf7b", bu.String()); err == nil {
		// if solves, err := strconv.Atoi(t.Config["solves"]); err == nil {
		// 	solves += 1
		// 	if solves >  {
		// 		t.StatusCh <- Status{Message: "Stuck in captcha loop, backing off"}
		// 		time.Sleep(60*time.Second)
		// 		t.Config["solves"] = "0"
		// 		return TaskErrUnrecoverable(ErrFailedToSolveCaptcha)
		// 	}
		// 	t.Config["solves"] = fmt.Sprintf("%d", solves)
		// } else {
		// 	t.Config["solves"] = "1"
		// }

		// log.Printf("recap token %s", recapToken)
		err = t.Retry(func() error {
			var pxBlock PxBlockResponse
			var err error
			json.Unmarshal(blockResponse, &pxBlock)
			ckies := []*http.Cookie{}
			if ckies, err = t.GetPxCookieApi(Px3Request{
				Proxy: t.ProxyStr,
				Cookie: t.CookiesStr(),
				Url: "https://" + t.Url.Host + pxBlock.RedirectURL,
				AppId: pxBlock.AppID,
				Vid: pxBlock.Vid,
				Uuid: pxBlock.UUID,
				Host: pxBlock.HostURL,
				JsSrc: "https://" + t.Url.Host + pxBlock.JsClientSrc,
				CapJsSrc: "https://" + t.Url.Host + pxBlock.BlockScript,
				RecapToken: recapToken,
			}); err == nil {
				t.LogDebug("got cookies %+v", ckies)
				for _, ckie := range ckies {
					t.SetCookie(ckie)
				}
				// t.FtlSendSessionSnapshot()
				if !quiet {
					t.StatusCh <- StatusCapSolved
				}
				return nil
			} else {
				t.LogDebug("gen err %+v", err)
			}
			return err
		}, retry.Attempts(4))
		if err == nil {
			return nil
		}
	}
	if !quiet {
		t.StatusCh <- StatusSolvingCap
	}
	return err
}
var DEFAULT_BR = []byte(`{
			"redirectUrl": "/blocked?url=L2lwL0JsYWNrLU1vdW50YWluLVByb2R1Y3RzLVJlc2lzdGFuY2UtQmFuZC1TZXQtd2l0aC1Eb29yLUFuY2hvci1BbmtsZS1TdHJhcC1FeGVyY2lzZS1DaGFydC1hbmQtUmVzaXN0YW5jZS1CYW5kLUNhcnJ5aW5nLUNhc2UvMjM0NzE2MzU=&uuid=&vid=&g=a",
			"appId": "PXu6b0qd2S",
			"jsClientSrc": "/px/PXu6b0qd2S/init.js",
			"firstPartyEnabled": true,
			"vid": "",
			"uuid": "",
			"hostUrl": "/px/PXu6b0qd2S/xhr",
			"blockScript": "/px/PXu6b0qd2S/captcha/captcha.js?a=c\u0026m=0\u0026u=\u0026v=\u0026g=a"
		}`)
var DEFAULT_BR_STRUCT = PxBlockResponse{
	RedirectURL: "/blocked?url=L2lwL0JsYWNrLU1vdW50YWluLVByb2R1Y3RzLVJlc2lzdGFuY2UtQmFuZC1TZXQtd2l0aC1Eb29yLUFuY2hvci1BbmtsZS1TdHJhcC1FeGVyY2lzZS1DaGFydC1hbmQtUmVzaXN0YW5jZS1CYW5kLUNhcnJ5aW5nLUNhc2UvMjM0NzE2MzU=&uuid=&vid=&g=a",
	AppID: "PXu6b0qd2S"	,
	JsClientSrc: "/px/PXu6b0qd2S/init.js",
	FirstPartyEnabled: true,
	HostURL: "/px/PXu6b0qd2S/xhr",
	BlockScript: "/px/PXu6b0qd2S/captcha/captcha.js?a=c&m=0&u=&v=&g=a",
}

func (t *CheckoutTask) WmtCreateAccount() error {
	return t.Retry(func() error {
		t.StatusCh <- StatusCreatingAccount

		pw := randHex(10)
		req, err := t.MakeChlsRequest(WMT_CREATE_ACC_CHLSJ, map[*regexp.Regexp]string {
			regexp.MustCompile("Iam"): t.Profile.ShippingAddress.FirstName,
			regexp.MustCompile("Aka"): t.Profile.ShippingAddress.LastName,
			regexp.MustCompile("rrrrrr@gmail.com"): t.Profile.Email,
			regexp.MustCompile("\"zzzzzzzzz\""): "\"" + pw + "\"", // TODO pw gen, store
		})
		t.LogDebug("%+v %+v", req, err)
		// b, _ := ioutil.ReadAll(req.Request.Body)
		// t.LogDebug("%s", string(b))
		resp, err := t.doReq(t.client, req.Request)
		if err != nil {
			return err
		}
					if resp != nil && resp.StatusCode == 412 {
						return retry.Unrecoverable(ErrRetrying)
						if bbb, err := readBodyBytes(resp); err == nil {
		  				if err := t.WmtSolveCaptcha(false, bbb); err != nil {
		            return retry.Unrecoverable(ErrRetrying)
		          }
		        }
						return ErrRetrying
			  	}
		bb, err := readBodyBytes(resp)
		if err != nil {
			return err
		}

		for _, rcand := range *req.Responses {
			if ourresp := rcand.MatchJsonRootKeys(bb); ourresp != nil {
				if resp.StatusCode == 200 {
					t.VaultPut(t.Url.Host, t.Profile.Email, pw)

					gu, _ := url.Parse("https://www.walmart.com/account?r=yes&action=SignIn&rm=true")
					reqq := t.makeGetReq(gu, nil)
					reqq.IgnoreBody = true
					t.doReq(t.client, reqq)

					return nil
				} else {
					if err, ok := (*ourresp)["error"].(map[string]interface{}); ok {
						t.StatusCh <- Status{Message: "Error creating account:" + (err["message"].(string))[:48]}
						return retry.Unrecoverable(retry.Unrecoverable(errors.New(err["message"].(string))))
					} else if err, ok := (*ourresp)["message"].(string); ok {
						t.StatusCh <- Status{Message: "Error creating account: " + err[:48]}
						return retry.Unrecoverable(retry.Unrecoverable(errors.New(err)))
					}
				}
			}
		}
		return retry.Unrecoverable(ErrUnknownResponse)
	}, retry.Attempts(4))
}

func (t *CheckoutTask) WmtCheckoutCache() (*map[string]interface{}, error) {
  url_, err := url.Parse("https://www.walmart.com/api/checkout/v3/checkout-cache")
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "accept",
    "omitcsrfjwt",
    "wm_qos.correlation_id",
    "User-Agent",
    "credentials",
    "content-type",
    "Origin",
    "Sec-Fetch-Site",
    "Sec-Fetch-Mode",
    "Sec-Fetch-Dest",
    "Referer",
    "Accept-Encoding",
    "Accept-Language",
    "Cookie",
    "Content-Length",
  }
  headers := [][2]string {
    {"Connection", "keep-alive"},
    {"accept", "application/json, text/javascript, */*; q=0.01"},
    {"omitcsrfjwt", "true"},
    {"wm_qos.correlation_id", "e43a0941-fbf0-4675-862d-9beee29ca718"},
    {"credentials", "include"},
    {"content-type", "application/json"},
    {"Origin", "https://www.walmart.com"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", "https://www.walmart.com/cart"},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }
  bodyStructure := map[string]interface{} {
  "requestType": "REGULAR_CACHE",
  "customerType": "CUSTOMER",
  "cacheIdType": "CUSTOMERID",
  "crt:CRT": "",
  "customerId:CID": "",
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  defer DiscardResp(resp)
  if err != nil {
    return nil, err
  }

  rs := map[string]interface{}{}
  err = readRespJsonDst(resp, &rs)
  return &rs, err
}

func (t *CheckoutTask) WmtCheckout() error {
	t.dialTimeout = time.Duration(10+rand.Intn(50))*time.Second
	t.UserAgent = DD_UAS[rand.Intn(len(DD_UAS))]
	if os.Getenv("PX_TEST") == "1" {
		ckies, err := t.WmtGenPxCookie()
		t.LogDebug("%+v %+v", ckies, err)
	}

	if os.Getenv("PX_TEST_CAP") == "1" {

		var pxBlock PxBlockResponse
		json.Unmarshal(DEFAULT_BR, &pxBlock)

		recapToken, err := t.SolveCaptcha("6Lcj-R8TAAAAABs3FrRPuQhLMbp5QrHsHufzLf7b", "https://" + t.Url.Host + pxBlock.RedirectURL)
		if err != nil {
			panic(err);
		}
		// log.Printf("recap token %s", recapToken)

		if ckies, err := t.GetPxCookieApi(Px3Request{
			Proxy: t.ProxyStr,
			Cookie: t.CookiesStr(),
			Url: "https://" + t.Url.Host + pxBlock.RedirectURL,
			AppId: pxBlock.AppID,
			Vid: pxBlock.Vid,
			Uuid: pxBlock.UUID,
			Host: pxBlock.HostURL,
			JsSrc: "https://" + t.Url.Host + pxBlock.JsClientSrc,
			CapJsSrc: "https://" + t.Url.Host + pxBlock.BlockScript,
			RecapToken: recapToken,
		}); err == nil {
			for _, ckie := range ckies {
				t.SetCookie(ckie)
			}
			t.StatusCh <- StatusCapSolved
			return ErrRetrying
		}
	}


	// t.ElgoogStartTrustBuilder()

	var err error
	ps := strings.Split(t.Url.Path, "/")
	pid := ps[len(ps) - 1]

	qstr := t.Url.Query().Get("qty")
	if qstr == "" {
		qstr = os.Getenv("QTY")
	}
	var quantity int64 = 1
	if qtymaybe, err := strconv.Atoi(qstr); err == nil {
		quantity = int64(qtymaybe)
	}
	presolve := os.Getenv("PRESOLVE") == "1"
	shipMethod := t.Url.Query().Get("ship_method")
	if os.Getenv("SHIPMETHOD") != "" {
		shipMethod = os.Getenv("SHIPMETHOD")
	}

	enableAccount := t.Url.Query().Get("account") == "1" || os.Getenv("ACCOUNT") == "1"
	t.Url.RawQuery = ""

	t.LogDebug(pid)


	// t.ApplyDelayInDoReq = true

	// needsPx := false
	// pxckies := []*http.Cookie{}
	// t.StatusCh <- StatusTestingProxy

	// err = t.WalmartLogin()
	s := &WalmartSession{
		time.Unix(0,0),
		t,
	}
	// if errors.Is(err, ErrPxBanned) {
	//   needsPx = true
		// t.StatusCh <- StatusPxGen
		// t.initClient()
		// t.UserAgent, err = t.getUserAgent()
	 //  if err != nil {
	 //    return err
	 //  }
	 //  if pxckies, err = t.WmtGenPxCookie(); err != nil {
	 //    return err
	 //  }
	// }


	var cardToken *WmtCardToken
	var payKey *WmtPayKey
	// var product *WmtProduct
	origDelay := t.Delay
	err = t.Retry(func() error {
		t.Delay = 1*time.Second
		defer func() {
			t.Delay = origDelay
		}()
		t.StatusCh <- Status{Message: "Pre-Encrypting Card"}
		payKey, err = t.WmtGetPayKey()
		if err != nil {
			pbak := t.Proxy
			t.Proxy = nil
			payKey, err = t.WmtGetPayKey()
			t.Proxy = pbak

			if err != nil {
				payKey = &WmtPayKey{
					L: 6,
					E: 4,
					K: "928510B220E997E3AD79D944507FC2B2",
					KeyId: "cfffb880",
					Phase: 1,
				}
		  }

		}
		cardToken, err = GetWmtToken(payKey, t.Profile.Card.Number, t.Profile.Card.Cvc)
	  return err
	})
	if err != nil {
		return err
	}

	return t.RetryTask(func() error {
		// t.Config["solves"] = "0"
		err = t.initClient()
		if err != nil {
		  return err
		}
		isSessionStoreEnabled := os.Getenv("SESH") != "0"

		var tsesh *TaskSession
		if isSessionStoreEnabled && len(t.ProxyStr) > 0 {
		  if resp, err := t.ApiPost("/sesh/get", t.GetBareSessionBytes()); err == nil {
		  	// log.Println(string(resp))
		    var s TaskSession
		    if err = json.Unmarshal(resp, &s); err == nil && len(s.UserAgent) > 0 {
		      tsesh = &s
		      t.LogDebug("using sesh: %+v", tsesh)
		    } else {
		    	t.LogDebug("failed to unmarshal sesh \n\n%+v\n\n %s",err)
		    	// log.Println(string(resp))
		    }
		  } else {
		    t.LogDebug("sesh get err %+v", err)
		  }
		}



		if tsesh != nil  {
		  t.ApplySession(tsesh)
		}

		t.StatusCh <- StatusDdEstablishingTrust
		if ckies, err := t.WmtGenPxCookie(); err == nil {
			for _, ckie := range ckies {
				t.SetCookie(ckie)
			}
			// t.FtlSendSessionSnapshot()
		}
		if ckies, err := t.WmtGenPxCookie(); err == nil {
			for _, ckie := range ckies {
				t.SetCookie(ckie)
			}
			// t.FtlSendSessionSnapshot()
		}

		pxCtx, pxCancel := context.WithCancel(t.ClientCtx)
		defer pxCancel()
		go func() {
			time.Sleep(30*time.Second)
			t.Retry(func() error {
				if ckies, err := t.WmtGenPxCookie(); err == nil {
					for _, ckie := range ckies {
						t.SetCookie(ckie)
					}
					// t.`FtlSend`SessionSnapshot()
				}
				select {
				case <-pxCtx.Done():
					return nil
				case <-time.After(time.Duration(rand.Intn(180))*time.Second):
					break
				}
				return ErrRetrying
			})
		}()


		// t.StatusCh <- StatusDdEstablishingTrust
		// // if ckies, err := t.WmtGenPxCookie(); err == nil {
		// // 	for _, ckie := range ckies {
		// // 		t.SetCookie(ckie)
		// // 	}
		// // }


		// if ckies, err := t.WmtGenPxCookie(); err == nil {
		// 	for _, ckie := range ckies {
		// 		t.SetCookie(ckie)
		// 	}
		// 	t.FtlSendSessionSnapshot()
		// }

		// err = t.WmtSetDefaultShipping()
		t.RemoveCookie("com.wm.reflector")
		t.RemoveCookie("xpa") // TODO why are these blank in applied sesh?
		t.RemoveCookie("hasCRT")
		t.RemoveCookie("CRT")
		t.RemoveCookie("ACID")
		t.RemoveCookie("hasACID")
		t.RemoveCookie("PCID")
		t.RemoveCookie("hasPCID")
		t.RemoveCookie("auth")
		t.RemoveCookie("rtoken")
		t.RemoveCookie("SPID")
		t.RemoveCookie("CID")
		t.RemoveCookie("hasCID")
		t.RemoveCookie("customer")
		t.RemoveCookie("type")
		t.RemoveCookie("WMP")
		t.RemoveCookie("bm_sv")
		// t.LogDebug("%+v",GetAllClientCookies(t.client))


		var offerId string
		product, err := t.WmtGetProduct(false, false)


		if os.Getenv("WMTEMAIL") != "" {
			t.Profile.Email = os.Getenv("WMTEMAIL")
		}
		if enableAccount {
			if lpw := t.VaultGet(t.Url.Host, t.Profile.Email); lpw != nil {
				err = t.WalmartLogin(t.Profile.Email, *lpw)
				if err != nil {
					// t.StatusCh <- Status{Message: "Error logging in"}
					return err
				}
			} else {
				err = t.WmtCreateAccount()
				if err != nil {
					// t.StatusCh <- Status{Message: "Error creating account"}
					return err
				}
			}
		}

		// err = t.WmtSetDefaultShipping()
		t.RemoveCookie("xpa") // TODO why are these blank in applied sesh?
		t.RemoveCookie("hasCRT")
		t.RemoveCookie("CRT")
		t.RemoveCookie("ACID")
		t.RemoveCookie("hasACID")
		t.RemoveCookie("PCID")
		t.RemoveCookie("hasPCID")
		t.RemoveCookie("auth")
		t.RemoveCookie("rtoken")
		t.RemoveCookie("SPID")
		t.RemoveCookie("CID")
		t.RemoveCookie("hasCID")
		t.RemoveCookie("customer")
		t.RemoveCookie("type")
		t.RemoveCookie("WMP")
		t.RemoveCookie("bm_sv")
		// t.LogDebug("%+v",GetAllClientCookies(t.client))
		// if err != nil {
		// 	return err
		// }



		// fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_%d_%d) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.67 Safari/537.36", 14 + rand.Intn(2), 4 + rand.Intn()
		// t.UserAgent, err = "Mozilla/5.0 (Windows NT 10.0; Win32; x32) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36"
	 //  if err != nil {
	 //    return err
	 //  }


		if os.Getenv("MOBILE") == "1" {
	    product, err = t.WmtGetProduct(false, false)
	    if err != nil {
	      return err
	    }
	    t.StatusCh <- StatusPxGen
			if _, err := t.WmtGenPxCookie(); err == nil {
				// px3 := "" // todo get from ckies arr
			  if prodMobile, err := s.WmtMobileGetProduct(pid); err == nil {
			  	offerId = prodMobile.Data.ProductByProductID.OfferList[0].OfferInfo.OfferID
			  }
			}
		}



		if presolve {
			t.WmtGetProduct(false, false)
			if err := t.WmtSolveCaptcha(false, []byte(`{
				"appId": "PXu6b0qd2S",
				"jsClientSrc": "/px/PXu6b0qd2S/init.js",
				"firstPartyEnabled": true,
				"vid": "",
				"uuid": "",
				"hostUrl": "/px/PXu6b0qd2S/xhr",
				"blockScript": "/px/PXu6b0qd2S/captcha/captcha.js?a=c\u0026m=0\u0026v=\u0026g=a"
			}`)); err != nil {
					return err
				}

				t.WmtGetProduct(false, false)
				if err := t.WmtSolveCaptcha(false, []byte(`{
					"appId": "PXu6b0qd2S",
					"jsClientSrc": "/px/PXu6b0qd2S/init.js",
					"firstPartyEnabled": true,
					"vid": "",
					"uuid": "",
					"hostUrl": "/px/PXu6b0qd2S/xhr",
					"blockScript": "/px/PXu6b0qd2S/captcha/captcha.js?a=c\u0026m=0\u0026v=\u0026g=a"
				}`)); err != nil {
						return err
					}


			if os.Getenv("PRESOLVE_HALT") == "1" {
				t.LogDebug("%+v", t.TaskSessionSnapshot())
				// t.FtlSendSessionSnapshot()
				return nil
			}
			// 	return nil
				// panic("Saved presolve")
			// 	// time.Sleep(1800*time.Second)
			// }
		}

		// t.StatusCh <- StatusDdEstablishingTrust
		// // if ckies, err := t.WmtGenPxCookie(); err == nil {
		// // 	for _, ckie := range ckies {
		// // 		t.SetCookie(ckie)
		// // 	}
		// // }


		// if ckies, err := t.WmtGenPxCookie(); err == nil {
		// 	for _, ckie := range ckies {
		// 		t.SetCookie(ckie)
		// 	}
		// 	t.FtlSendSessionSnapshot()
		// }


		pmut := sync.Mutex{}
		var released bool
		presolveckies := [][]*http.Cookie{}
		// var waitedForRelease bool
		if bb, err := t.S3Get("wrels.json", 3); err == nil {
			t.LogDebug("%s", string(bb))
			var wrels map[string]int64
			if err := json.Unmarshal(bb, &wrels); err == nil {
				t.LogDebug("%v", wrels)
				if rlepoch, ok := wrels[pid]; ok {
					reltime := time.Unix(rlepoch, 0)

					if os.Getenv("TEST_RELEASE_30SEC") == "1" {
					  reltime = time.Now().Add(30*time.Second)
					}
					// reltime = time.Now().Add(30*time.Second)
					released = reltime.Sub(time.Now()) < (10*time.Second)
					t.LogDebug("%v", reltime)

						// TODO farm

						// reltime = time.Now().Add(10*time.Second)
						// reltime = time.Now().Add(120*time.Second)
						if !released && (t.GetCookieValue("_px3") == "" || rand.Float32() < 0.25) {
							go func() {
								// if tsesh == nil || rand.Float32() < 0.25 {
									defer func() {
	                    recover()
	                  }()
									t.Retry(func() error {
										pmut.Lock()
										if len(presolveckies) > 0 || released {
											pmut.Unlock()
											return nil
										}
										pmut.Unlock()

										if ckies, err := t.WmtSolveCaptchaCkies(true, DEFAULT_BR); err == nil {
											pmut.Lock()
											presolveckies = append(presolveckies, ckies)
											t.LogDebug("PRESOLVE %+v %+v", presolveckies, err)
											  if len(presolveckies) > 0 {
											  	pckies := presolveckies[len(presolveckies) - 1]
											  	if len(presolveckies) < 2 {
											  		presolveckies = nil
											  	} else {
											  		presolveckies = presolveckies[:len(presolveckies) - 1]
											  	}
												  t.RemoveCookie("_px3")
												  t.RemoveCookie("pxsid")
												  t.RemoveCookie("_pxvid")
												  t.RemoveCookie("_pxde")
												  t.RemoveCookie("_pxff_rf")
												  t.RemoveCookie("_pxff_fp")
												  for _, ckie := range pckies {
												  	t.SetCookie(ckie)
												  }
												}
											pmut.Unlock()

											t.WmtGetProduct(true, false)
											return nil
										}
										return ErrRetrying
									}, retry.Attempts(196), retry.Delay(time.Duration(rand.Intn(10))*time.Second))
								// }
							}()
						}
						wt := reltime.Sub(time.Now()) + (time.Duration(-10 + rand.Intn(10)) * time.Second)
						// wt = 0
						t.LogDebug("waiting %v", wt)

						for !released {
							t.StatusCh <- Status{Message: fmt.Sprintf("Waiting for Release (%s EST)", reltime.In(ESTZ).Format("15:04"))}
							select {
							case <-time.After(wt):
								released = true
								continue
							case <-t.Ctx.Done():
								return t.Ctx.Err()
							}
							time.Sleep(1*time.Second)
						}
					}
			}
		}


	var pprod WmtProductItem
		return t.Retry(func() error {
			if !released && !pprod.Shippable {
				if rand.Float32() < 1.2 {
					// err := t.WmtAtcAlt2(int(quantity))
					// if err != nil {
					// 	return err
					// }
				} else {
					product, err = t.WmtGetProduct(false, true)
							for _, prod_ := range product.Item.Product.BuyBox.Products {
								if (prod_.CatalogSellerID == "0" || os.Getenv("WMT_THIRD_PARTY") == "1") || (offerId != "" && prod_.OfferID == offerId) {
									pprod = prod_
									break
								}
							}
					    offerId = pprod.OfferID
							if pprod.OfferID == "" {
								// TODO grab from api
								return ErrProductUnavailable
							}

							if pprod.MaxQuantity > 0 && quantity > pprod.MaxQuantity {
								quantity = pprod.MaxQuantity
							}

							t.LogDebug("OFFERID %s", offerId)
					if err == nil {
							for _, prod_ := range product.Item.Product.BuyBox.Products {
								if (prod_.CatalogSellerID == "0" || os.Getenv("WMT_THIRD_PARTY") == "1") || (offerId != "" && prod_.OfferID == offerId) {
									pprod = prod_
									break
								}
							}
					    offerId = pprod.OfferID
					  }
					}
			}
			if released && t.Delay > 0 {
				t.Delay = time.Duration(rand.Intn(250+750))*time.Millisecond
				defer func() {
					t.Delay = origDelay
				}()
			}
			// t.SetCookie(&http.Cookie{Name: "akavpau_p0", Value: t.GetCookieValue("akavpau_p8"), Path: "/" })
			// t.SetCookie(&http.Cookie{Name: "akavpau_p8", Value: "", Path: "/", MaxAge: -1, Expires: time.Unix(0,0) })
			cl := 1
			x := 0
			cl, err = strconv.Atoi(os.Getenv("CART_LOOP"))
			if cl == 0 {
				cl = 1
			}



				for x < cl {
					// t.RemoveCookie("CRT")
					// t.RemoveCookie("hasCRT")
					// t.RemoveCookie("ACID")
					// t.RemoveCookie("hasACID")
					// t.RemoveCookie("CXO_CART_ST")
					x += 1
					  if len(presolveckies) > 0 {
					  	pckies := presolveckies[len(presolveckies) - 1]
					  	if len(presolveckies) < 2 {
					  		presolveckies = nil
					  	} else {
					  		presolveckies = presolveckies[:len(presolveckies) - 1]
					  	}
						  t.RemoveCookie("_px3")
						  t.RemoveCookie("pxsid")
						  t.RemoveCookie("_pxvid")
						  t.RemoveCookie("_pxde")
						  t.RemoveCookie("_pxff_rf")
						  t.RemoveCookie("_pxff_fp")
						  for _, ckie := range pckies {
						  	t.SetCookie(ckie)
						  }
						  // t.`FtlSend`SessionSnapshot()
						}
					t.StatusCh <- Status{Message: "Adding to Cart"}
					err = t.WmtAtcAlt2(int(quantity))
				  if err != nil {
				  	if serr, ok :=err.(*StatusCodeRejectedError); ok {
				  		if serr.StatusCode == 444 {
								if err = t.WmtAtcAlt(offerId, fmt.Sprintf("%d", quantity)); err != nil {
					  			t.StatusCh <- StatusProxyBanned
					  			return retry.Unrecoverable(err)
					  		}
				  		} else {
				  			return err
				  		}
				  	} else {
				    	return err
				    }
				  }
				}


		  // t.EnableHttpTrace = true
		  // defer func() {
		  // 	t.EnableHttpTrace = false
		  // }()

		 //  if needsPx {
			// 	if ckies, err := t.WmtGenPxCookie(); err == nil {
			// 		for _, ckie := range ckies {
			// 			t.SetCookie(ckie)
			// 		}
			// 	}
			// }

		// if ckies, err := t.WmtGenPxCookie(); err == nil {
		// 		for _, ckie := range ckies {
		// 			t.SetCookie(ckie)
		// 		}
		// 	}

			t.Delay = time.Duration(rand.Intn(333))*time.Millisecond
			defer func() {
				t.Delay = origDelay
			}()
			t.DefaultReqClose = true
			defer func() {
				t.DefaultReqClose = false
			}()
			CloseH2Conns(t.client)


		  t.StartCheckoutTime = timeMillis()
		  // t.WmtCheckoutCache()
		  // t.StatusCh <- Status{Message: "Start Checkout"}

		  checkoutSesh, err := t.WmtStartCheckout()
		  if err != nil {
		    return err
		  }
		  err = t.WmtStartCheckout2()
		  if err != nil {
		    return err
		  }


		  // if ckies, err := t.WmtGenPxCookie(); err == nil {
				// 	for _, ckie := range ckies {
				// 		t.SetCookie(ckie)
				// 	}
				// }
		  if shipMethod != "" {
			  t.StatusCh <- Status{Message: "Set Shipping Method"}
			  err = t.WmtFulfillment(checkoutSesh, shipMethod)
			  if err != nil {
			    return err
			  }
			}

		  t.StatusCh <- Status{Message: "Set Shipping Address"}
		  err = t.WmtSetShipping()
		  if err != nil {
		    return err
		  }

		  // time.Sleep(5*time.Second)

		  // TODO can we modify cart during checkout?
		  // err = t.WmtAtc(product.Item.Product.BuyBox.Products[0].OfferID)
		  // if err != nil {
		  //   return err
		  // }

		  // t.StatusCh <- Status{Message: "Set Card"}
		  // err = t.WmtSetCard(cardToken, payKey)
		  // if err != nil {
		  //   return err
		  // }
		  // if ckies, err := t.WmtGenPxCookie(); err == nil {
				// 	for _, ckie := range ckies {
				// 		t.SetCookie(ckie)
				// 	}
				// }
		  t.StatusCh <- Status{Message: "Set Billing Address"}
		  err = t.WmtPayment(cardToken, payKey)
		  if err != nil {
		    return err
		  }


		  if len(presolveckies) > 0 {
		  	pckies := presolveckies[0]
		  	if len(presolveckies) < 2 {
		  		presolveckies = nil
		  	} else {
		  		presolveckies = presolveckies[1:]
		  	}
			  t.RemoveCookie("_px3")
			  t.RemoveCookie("pxsid")
			  t.RemoveCookie("_pxvid")
			  t.RemoveCookie("_pxde")
			  for _, ckie := range pckies {
			  	t.SetCookie(ckie)
			  }
			}

		  t.StartPaymentTime = timeMillis()
		  t.StatusCh <- Status{Message: "Sending Order"}
		  err = t.WmtOrder(cardToken, payKey)

		  whFields := [][2]string{
		  	{"Website", t.Url.Host},
		  	{"Product", checkoutSesh.Items[0].ProductName},
		  	{"Quantity", fmt.Sprintf("%d", quantity)},
		  	// {"Size", prodMobile.Size},
		  	// {"Style", product.Style},
		  }
		  if len(pprod.Images) > 0 {
		  	whFields = append([][2]string{ {"image", pprod.Images[0].URL }}, whFields...)
		  }
		  if err == nil {
		  	t.LogDebug("%+v", t.TaskSessionSnapshot())
		  	t.StatusCh <- Status{Message: "COOKED!!!!!"}
		  	go t.SuccessWebhookFields(whFields)
		  	t.FtlSendSessionSnapshotBlk()
		  	return nil
		  } else if errors.Is(err, errDecline) {
		  	t.LogDebug("%+v", t.TaskSessionSnapshot())
		  	t.StatusCh <- StatusCardDeclined
		  	go t.DeclineWebhookFields(whFields)
		  	t.FtlSendSessionSnapshotBlk()
		  	return retry.Unrecoverable(retry.Unrecoverable(err))
		  }
		  return err
		})
	 }, retry.Delay(t.Delay), retry.DelayType(retry.DefaultDelayType), retry.Attempts(1e6), retry.OnRetry(func (_ uint, err error) {
	 	// t.StatusCh <- Status{Message: "Task Failed - Waiting to Restart"}
    t.LogDebug("retry %s\n", err.Error())
    CloseH2Conns(t.client)
  })) // TODO backoff for proxy bans
}


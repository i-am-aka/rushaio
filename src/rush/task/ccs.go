package task

import (
	"encoding/json"
	"github.com/avast/retry-go"
	// "time"
	"fmt"
	// "bytes"
	"log"
	// "math/rand"
	"strings"
	"rush/net/http"
	"net/url"
	"regexp"
	"github.com/pkg/errors"
	// "github.com/PuerkitoBio/goquery"
)

type CcsProductAttributeOption struct {
	Id string `json:"id"`
	Label string `json:"label"`
}

type CcsProductAttributeSet struct {
	Options []CcsProductAttributeOption `json:"options"`
}

type CcsProductAttributes struct {
	Color CcsProductAttributeSet `json:"554"`
	Size CcsProductAttributeSet `json:"454"`
}

type CcsProduct struct {
	Pid string `json:"productId"`
	Name string
	Attributes CcsProductAttributes `json:"attributes"`
}

func (t *CheckoutTask) ccsGetProduct() (CcsProduct, error) {
	var product CcsProduct
  headerOrder := []string {
    "Host",
    "Connection",
    "Upgrade-Insecure-Requests",
    "User-Agent",
    "Accept",
    "Sec-Fetch-Site",
    "Sec-Fetch-Mode",
    "Sec-Fetch-Dest",
    "Accept-Encoding",
    "Accept-Language",
  }
  headers := [][2]string {
    {"Connection", "keep-alive"},
    {"Upgrade-Insecure-Requests", "1"},
    {"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"Sec-Fetch-Site", "none"},
    {"Sec-Fetch-Mode", "navigate"},
    {"Sec-Fetch-Dest", "document"},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }
  resp, err := t.doReq(t.client, t.makeReq("GET", t.Url, &headers, &headerOrder, nil))
  if err != nil {
  	return product, err
  }
  prodBody, err := readBody(resp)
  if err != nil {
  	return product, err
  }

  prodConfigRe := regexp.MustCompile(`new Product.Config\((.+?)\);`)
  matches := prodConfigRe.FindStringSubmatch(prodBody)
  if len(matches) != 2 {
  	if strings.Contains(prodBody, "OptionsPrice") {
  		pidRe := regexp.MustCompile(`input type="hidden" name="product" value="(.+?)"`)
  		pidMatches := pidRe.FindStringSubmatch(prodBody)
  		if len(pidMatches) == 2 {
  			product.Pid = pidMatches[1]
  		} else {
  			return product, errors.New("Missing product config (Unknwon Page Format)")
  		}
  	} else {
  		return product, errors.New("Missing product config (Incap Blocked)")
  	}
  } else {
  	json.Unmarshal([]byte(matches[1]), &product)
  }

  titleRe := regexp.MustCompile(`<title>(.+?)<\/title>`)
  titleMatches := titleRe.FindStringSubmatch(prodBody)
  if len(titleMatches) == 2 {
  	product.Name = titleMatches[1]
  }

  return product, err
}

func (t *CheckoutTask) ccsAtc(pid string, colorway string, size string) (error) {
	return t.Retry(func() error {
		  url_, err := url.Parse("https://shop.ccs.com/checkout/cart/addAjax/")
		  if err != nil {
		    return err
		  }
		  headerOrder := []string {
		    "Host",
		    "Connection",
		    "Content-Length",
		    "Accept",
		    "X-Prototype-Version",
		    "X-Requested-With",
		    "User-Agent",
		    "Content-type",
		    "Origin",
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
		    {"Accept", "text/javascript, text/html, application/xml, text/xml, */*"},
		    {"X-Prototype-Version", "1.7"},
		    {"X-Requested-With", "XMLHttpRequest"},
		    {"Content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
		    {"Origin", "https://shop.ccs.com"},
		    {"Sec-Fetch-Site", "same-origin"},
		    {"Sec-Fetch-Mode", "cors"},
		    {"Sec-Fetch-Dest", "empty"},
		    {"Referer", t.Url.String()},
		    {"Accept-Encoding", "gzip, deflate, br"},
		    {"Accept-Language", "en-US,en;q=0.9"},
		  }
		  params := [][2]string {
		    {"product", pid},
		    {"related_product", ""},
		  }
		  if colorway != "" {
		     params = append(params, [2]string{"super_attribute[554]", colorway})
		  }
		  if size != "" {
		     params = append(params, [2]string{"super_attribute[454]", size})
		  }
		  params = append(params, [2]string{"qty", "1"})

		  body := urlEncodeParams(params)
		  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
		  if err != nil {
		  	return err
		  }
		  body, err = readBodyBytes(resp)
		  if err != nil {
		  	return err
		  }
		  var cartResp map[string]interface{}
		  err = json.Unmarshal(body, &cartResp)
		  if err != nil {
		  	return err
		  }
		  return nil
	 }, retry.Delay(0))
}

func (t *CheckoutTask) ccsGetIncapCookie() (error) {
	return t.Retry(func() error {
	  url_, err := url.Parse("https://shop.ccs.com/")
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "Host",
	    "Connection",
	    "Upgrade-Insecure-Requests",
	    "User-Agent",
	    "Accept",
	    "Sec-Fetch-Site",
	    "Sec-Fetch-Mode",
	    "Sec-Fetch-User",
	    "Sec-Fetch-Dest",
	    "Accept-Encoding",
	    "Accept-Language",
	  }
	  headers := [][2]string {
	    {"Connection", "keep-alive"},
	    {"Upgrade-Insecure-Requests", "1"},
	    {"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	    {"Sec-Fetch-Site", "none"},
	    {"Sec-Fetch-Mode", "navigate"},
	    {"Sec-Fetch-User", "?1"},
	    {"Sec-Fetch-Dest", "document"},
	    {"Accept-Encoding", "gzip, deflate, br"},
	    {"Accept-Language", "en-US,en;q=0.9"},
	  }
	  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
	  if err != nil {
	  	return err
	  }
	  body, err := readBody(resp)
	  if err != nil {
	  	return err
	  }
	  if strings.Contains(body, "Request unsuccessful.") {
	  	return retry.Unrecoverable(errors.New("Error: Incap splash bypass not implemented"))
	  }
	  jsRe := regexp.MustCompile(`src="(/_Incapsula_Resource\?SWJIYLWA=.+?)"`)
	  jsMatches := jsRe.FindStringSubmatch(body)
	  if len(jsMatches) != 2 {
	  	log.Println("no incap js, assuming has cookie")
	  	return nil
	  }
	  jsUrl := "https://shop.ccs.com" + jsMatches[1]
	  cookies := []string{}
	  for _, c := range t.client.Jar.Cookies(url_) {
	  	cookies = append(cookies, fmt.Sprintf("%s=%s", c.Name, c.Value))
	  }
	  proxyStr := ""
	  if t.Proxy != nil {
	  	proxyStr = t.Proxy.String()
	  }
	  chalSoln, err := SolveIncapChallenge(proxyStr, jsUrl, strings.Join(cookies, "; "))
	  if err != nil {
	  	return err
	  }
	  // url_, err = url.Parse("https://shop.ccs.com/skin/frontend/enterprise/shopccs/js/jquery.cycle.all.js")
	  // if err != nil {
	  //   return err
	  // }
	  // headerOrder = []string {
	  //   "Host",
	  //   "Connection",
	  //   "User-Agent",
	  //   "Accept",
	  //   "Sec-Fetch-Site",
	  //   "Sec-Fetch-Mode",
	  //   "Sec-Fetch-Dest",
	  //   "Referer",
	  //   "Accept-Encoding",
	  //   "Accept-Language",
	  //   "Cookie",
	  // }
	  // headers = [][2]string {
	  //   {"Connection", "keep-alive"},
	  //   {"Accept", "*/*"},
	  //   {"Sec-Fetch-Site", "same-origin"},
	  //   {"Sec-Fetch-Mode", "no-cors"},
	  //   {"Sec-Fetch-Dest", "script"},
	  //   {"Referer", "https://shop.ccs.com/"},
	  //   {"Accept-Encoding", "gzip, deflate, br"},
	  //   {"Accept-Language", "en-US,en;q=0.9"},
	  // }
	  // resp, err = t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
	  // resp.Body.Close()
	  // if err != nil {
	  // 	return err
	  // }

	  // url_, err = url.Parse("https://shop.ccs.com/skin/frontend/enterprise/shopccs/js/bootstrap.min.js")
	  // if err != nil {
	  //   return err
	  // }
	  // headerOrder = []string {
	  //   "Host",
	  //   "Connection",
	  //   "User-Agent",
	  //   "Accept",
	  //   "Sec-Fetch-Site",
	  //   "Sec-Fetch-Mode",
	  //   "Sec-Fetch-Dest",
	  //   "Referer",
	  //   "Accept-Encoding",
	  //   "Accept-Language",
	  //   "Cookie",
	  // }
	  // headers = [][2]string {
	  //   {"Connection", "keep-alive"},
	  //   {"Accept", "*/*"},
	  //   {"Sec-Fetch-Site", "same-origin"},
	  //   {"Sec-Fetch-Mode", "no-cors"},
	  //   {"Sec-Fetch-Dest", "script"},
	  //   {"Referer", "https://shop.ccs.com/"},
	  //   {"Accept-Encoding", "gzip, deflate, br"},
	  //   {"Accept-Language", "en-US,en;q=0.9"},
	  // }
	  // resp, err = t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
	  // resp.Body.Close()
	  // if err != nil {
	  // 	return err
	  // }

	  // url_, err = url.Parse("https://shop.ccs.com/skin/frontend/enterprise/shopccs/css/bootstrap.min.css")
	  // if err != nil {
	  //   return err
	  // }
	  // headerOrder = []string {
	  //   "Host",
	  //   "Connection",
	  //   "User-Agent",
	  //   "Accept",
	  //   "Sec-Fetch-Site",
	  //   "Sec-Fetch-Mode",
	  //   "Sec-Fetch-Dest",
	  //   "Referer",
	  //   "Accept-Encoding",
	  //   "Accept-Language",
	  //   "Cookie",
	  // }
	  // headers = [][2]string {
	  //   {"Connection", "keep-alive"},
	  //   {"Accept", "*/*"},
	  //   {"Sec-Fetch-Site", "same-origin"},
	  //   {"Sec-Fetch-Mode", "no-cors"},
	  //   {"Sec-Fetch-Dest", "script"},
	  //   {"Referer", "https://shop.ccs.com/"},
	  //   {"Accept-Encoding", "gzip, deflate, br"},
	  //   {"Accept-Language", "en-US,en;q=0.9"},
	  // }
	  // resp, err = t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
	  // resp.Body.Close()
	  // if err != nil {
	  // 	return err
	  // }


	  url_, err = url.Parse(chalSoln.ImgEndpoint)
	  if err != nil {
	    return err
	  }
	  t.client.Jar.SetCookies(url_, []*http.Cookie{ &http.Cookie{ Name: "___utmvc", Value: chalSoln.Cookie, Path: "/_Incapsula_Resource" }})
	  headerOrder = []string {
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
	  headers = [][2]string {
	    {"Connection", "keep-alive"},
	    {"Accept", "image/webp,image/apng,image/*,*/*;q=0.8"},
	    {"Sec-Fetch-Site", "same-origin"},
	    {"Sec-Fetch-Mode", "no-cors"},
	    {"Sec-Fetch-Dest", "image"},
	    {"Referer", "https://shop.ccs.com/"},
	    {"Accept-Encoding", "gzip, deflate, br"},
	    {"Accept-Language", "en-US,en;q=0.9"},
	  }
	  resp, err = t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
	  if err != nil {
	  	return err
	  }
	  body, err = readBody(resp)
	  if err != nil {
	  	return err
	  }
	  if body != "1" {
	  	status := Status{Message: "Invalid incap response. Resetting cookies and waiting 30 seconds"}
	  	t.StatusCh <- status
	  	return errors.New(status.Message)
	  }
	  return nil
	}, retry.OnRetry(func(n uint, err error) {
		if t.Debug {
			log.Printf("#%d: %s\n", n, err)
		}
	}))
}

func (t *CheckoutTask) ccsStartCheckout() (string, error) {
	var formKey string
	url_, err := url.Parse("https://shop.ccs.com/checkout/onepage/")
  if err != nil {
    return formKey, err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Upgrade-Insecure-Requests",
    "User-Agent",
    "Accept",
    "Sec-Fetch-Site",
    "Sec-Fetch-Mode",
    "Sec-Fetch-User",
    "Sec-Fetch-Dest",
    "Referer",
    "Accept-Encoding",
    "Accept-Language",
    "Cookie",
  }
  headers := [][2]string {
    {"Connection", "keep-alive"},
    {"Upgrade-Insecure-Requests", "1"},
    {"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "navigate"},
    {"Sec-Fetch-User", "?1"},
    {"Sec-Fetch-Dest", "document"},
    {"Referer", t.Url.String()},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }
  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
  if err != nil {
  	return formKey, err
  }
  body, err := readBody(resp)
  if err != nil {
  	return formKey, err
  }
  formKeyRe := regexp.MustCompile(`input name="form_key" type="hidden" value="(.+?)"`)
  formKeyMatches := formKeyRe.FindStringSubmatch(body)
  if len(formKeyMatches) != 2 {
  	return formKey, errors.New("Missing form key")
  }
  formKey = formKeyMatches[1]
  return formKey, nil
}

func (t *CheckoutTask) ccsUseGuest() (error) {
  url_, err := url.Parse("https://shop.ccs.com/checkout/onepage/saveMethod/")
  if err != nil {
    return err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Content-Length",
    "Accept",
    "X-Prototype-Version",
    "X-Requested-With",
    "User-Agent",
    "Content-type",
    "Origin",
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
    {"Accept", "text/javascript, text/html, application/xml, text/xml, */*"},
    {"X-Prototype-Version", "1.7"},
    {"X-Requested-With", "XMLHttpRequest"},
    {"Content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
    {"Origin", "https://shop.ccs.com"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", "https://shop.ccs.com/checkout/onepage/"},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }
  params := [][2]string {
    {"method", "guest"},
  }
  body := urlEncodeParams(params)
  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  if resp != nil {
  	resp.Body.Close()
  }
  return err
}

func (t *CheckoutTask) ccsSetShipping() (error) {
  url_, err := url.Parse("https://shop.ccs.com/checkout/onepage/saveShipping/")
  if err != nil {
    return err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Content-Length",
    "Accept",
    "X-Prototype-Version",
    "X-Requested-With",
    "User-Agent",
    "Content-type",
    "Origin",
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
    {"Accept", "text/javascript, text/html, application/xml, text/xml, */*"},
    {"X-Prototype-Version", "1.7"},
    {"X-Requested-With", "XMLHttpRequest"},
    {"Content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
    {"Origin", "https://shop.ccs.com"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", "https://shop.ccs.com/checkout/onepage/"},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }
  params := [][2]string {
    {"shipping[address_id]", "12577130"},
    {"shipping[country_id]", "US"},
    {"shipping[firstname]", t.Profile.ShippingAddress.FirstName},
    {"shipping[lastname]", t.Profile.ShippingAddress.LastName},
    {"shipping[street][]", t.Profile.ShippingAddress.Address1},
    {"shipping[street][]", t.Profile.ShippingAddress.Address2},
    {"shipping[city]", t.Profile.ShippingAddress.City},
    {"shipping[region_id]", "16"},
    {"shipping[region]", ""},
    {"shipping[postcode]", t.Profile.ShippingAddress.Zip},
    {"shipping[save_in_address_book]", "1"},
    {"shipping[telephone]", t.Profile.ShippingAddress.NordPhone()},
    {"shipping[email]", t.Profile.Email},
    {"shipping[same_as_billing]", "0"},
  }
  body := urlEncodeParams(params)
    resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  if resp != nil {
  	resp.Body.Close()
  }
  return err
}

func (t *CheckoutTask) ccsSaveShipMethod(formKey string) (error) {
  url_, err := url.Parse("https://shop.ccs.com/checkout/onepage/saveShippingMethod/")
  if err != nil {
    return err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Content-Length",
    "Accept",
    "X-Prototype-Version",
    "X-Requested-With",
    "User-Agent",
    "Content-type",
    "Origin",
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
    {"Accept", "text/javascript, text/html, application/xml, text/xml, */*"},
    {"X-Prototype-Version", "1.7"},
    {"X-Requested-With", "XMLHttpRequest"},
    {"Content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
    {"Origin", "https://shop.ccs.com"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", "https://shop.ccs.com/checkout/onepage/"},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }
  params := [][2]string {
    {"shipping_method", "freeshipping_freeshipping"},
    {"form_key", formKey},
  }
  body := urlEncodeParams(params)
    resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  if resp != nil {
  	resp.Body.Close()
  }
  return err
}

func (t *CheckoutTask) ccsSaveBilling() (error) {
  url_, err := url.Parse("https://shop.ccs.com/checkout/onepage/saveBilling/")
  if err != nil {
    return err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Content-Length",
    "Accept",
    "X-Prototype-Version",
    "X-Requested-With",
    "User-Agent",
    "Content-type",
    "Origin",
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
    {"Accept", "text/javascript, text/html, application/xml, text/xml, */*"},
    {"X-Prototype-Version", "1.7"},
    {"X-Requested-With", "XMLHttpRequest"},
    {"Content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
    {"Origin", "https://shop.ccs.com"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", "https://shop.ccs.com/checkout/onepage/"},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }

  var ccType string
  if t.Profile.Card.Type == "VISA" {
  	ccType = "VI"
  } else if t.Profile.Card.Type == "MASTER" {
  	ccType = "MC"
  } else if t.Profile.Card.Type == "AMEX" {
  	ccType = "AE"
  } else if t.Profile.Card.Type == "DISCOVER" {
  	ccType = "DI"
  } else {
  	return errors.New(fmt.Sprintf("Unsupported card type %s", t.Profile.Card.Type))
  }

  params := [][2]string {
    {"billing[address_id]", "12577129"},
    {"billing[same_as_shipping]", "1"},
    {"billing[country_id]", "US"},
    {"billing[firstname]", t.Profile.BillingAddress.FirstName},
    {"billing[lastname]", t.Profile.BillingAddress.LastName},
    {"billing[street][]", t.Profile.BillingAddress.Address1},
    {"billing[street][]", t.Profile.BillingAddress.Address2},
    {"billing[city]", t.Profile.BillingAddress.City},
    {"billing[region_id]", "16"},
    {"billing[region]", ""},
    {"billing[postcode]", t.Profile.BillingAddress.Zip},
    {"billing[save_in_address_book]", "1"},
    {"billing[email]", t.Profile.Email},
    {"billing[telephone]", t.Profile.BillingAddress.NordPhone()},
    {"payment[method]", "md_cybersource"},
    {"payment[cc_type]", ccType},
    {"payment[cc_number]", t.Profile.Card.Number},
    {"payment[cc_exp_month]", fmt.Sprintf("%d", t.Profile.Card.ExpMonth)}, // TODO leading zeroes?
    {"payment[cc_exp_year]", fmt.Sprintf("%d", t.Profile.Card.ExpYear)},
    {"payment[cc_cid]", t.Profile.Card.Cvc},
    {"payment[save_card]", "1"},
    {"coupon_code", ""},
    {"remove", "0"},
    {"giftcard_code", ""},
  }
  body := urlEncodeParams(params)
  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  if resp != nil {
  	resp.Body.Close()
  }
  return err
}

type CcsPayResponse struct {
	Success bool `json:"success"`
	Error bool `json:"error"`
	Message string `json:"error_messages"`
}

func (t *CheckoutTask) ccsPay() (CcsPayResponse, error) {
	t.StartPaymentTime = timeMillis()

  var payResp CcsPayResponse
  url_, err := url.Parse("https://shop.ccs.com/checkout/onepage/saveOrder/")
  if err != nil {
    return payResp, err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Content-Length",
    "Accept",
    "X-Prototype-Version",
    "X-Requested-With",
    "User-Agent",
    "Content-type",
    "Origin",
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
    {"Accept", "text/javascript, text/html, application/xml, text/xml, */*"},
    {"X-Prototype-Version", "1.7"},
    {"X-Requested-With", "XMLHttpRequest"},
    {"Content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
    {"Origin", "https://shop.ccs.com"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", "https://shop.ccs.com/checkout/onepage/"},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }

  var ccType string
  if t.Profile.Card.Type == "VISA" {
  	ccType = "VI"
  } else if t.Profile.Card.Type == "MASTER" {
  	ccType = "MC"
  } else if t.Profile.Card.Type == "AMEX" {
  	ccType = "AE"
  } else if t.Profile.Card.Type == "DISCOVER" {
  	ccType = "DI"
  } else {
  	return payResp, errors.New(fmt.Sprintf("Unsupported card type %s", t.Profile.Card.Type))
  }

  params := [][2]string {
    {"payment[method]", "md_cybersource"},
 		{"payment[cc_type]", ccType},
    {"payment[cc_number]", t.Profile.Card.Number},
    {"payment[cc_exp_month]", fmt.Sprintf("%d", t.Profile.Card.ExpMonth)}, // TODO leading zeroes?
    {"payment[cc_exp_year]", fmt.Sprintf("%d", t.Profile.Card.ExpYear)},
    {"payment[cc_cid]", t.Profile.Card.Cvc},
    {"payment[save_card]", "1"},
  }
  body := urlEncodeParams(params)
  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  if err != nil {
  	return payResp, err
  }
  body, err = readBodyBytes(resp)
  if err != nil {
  	return payResp, err
  }
  err = json.Unmarshal(body, &payResp)
  return payResp, err
}

func (t *CheckoutTask) CcsCheckout() error {
	t.StatusCh <- Status{Message: "Generating Incap Cookie"}
	err := t.ccsGetIncapCookie()
	if err != nil {
		return err
	}

	t.StatusCh <- Status{Message: "Carting Cheap Product"}
	// TODO many products
	err = t.ccsAtc("471261", "", "3884")
	if err != nil {
		return err
	}

	t.StatusCh <- Status{Message: "Start Checkout Preload"}
  formKey, err := t.ccsStartCheckout()
  if err != nil {
    return err
  }

  t.StatusCh <- Status{Message: "Checkout Preload 1/4"}
  err = t.ccsUseGuest()
  if err != nil {
    return err
  }
  t.StatusCh <- Status{Message: "Checkout Preload 2/4"}
  err = t.ccsSetShipping()
  if err != nil {
    return err
  }
  t.StatusCh <- Status{Message: "Checkout Preload 3/4"}
  err = t.ccsSaveShipMethod(formKey)
  if err != nil {
    return err
  }
  t.StatusCh <- Status{Message: "Checkout Preload 4/4"}
  err = t.ccsSaveBilling()
  if err != nil {
    return err
  }

  // t.StatusCh <- Status{Message: "4/4"}
	product, err := t.ccsGetProduct()
	log.Printf("%+v", product)

	var colorway CcsProductAttributeOption
	var size CcsProductAttributeOption
	for _, opt := range product.Attributes.Color.Options {
		colorway = opt
		break
	}
	for _, opt := range product.Attributes.Size.Options {
		size = opt
		break
	}
	log.Printf("%+v %+v %+v", product, colorway, size)

	t.StatusCh <- Status{Message: "Adding Desired Product to Cart"}
	t.StartCheckoutTime = timeMillis()
	t.ccsAtc(product.Pid, colorway.Id, size.Id)

	t.StatusCh <- Status{Message: "Fast Checkout 1/2"}
  formKey, err = t.ccsStartCheckout()
  if err != nil {
    return err
  }

  t.StatusCh <- Status{Message: "Fast Checkout 2/2"}
  payResp, err := t.ccsPay()
  if err != nil {
    return err
  }
  if payResp.Success {
  	t.StatusCh <- Status{Message: "COOKED!!!"}
		go t.SuccessWebhook("CCS", product.Name, size.Label)
		return nil
  } else {
  	if strings.Contains(payResp.Message, "decline") {
  		go t.DeclineWebhook("CCS", product.Name, size.Label)
  		return retry.Unrecoverable(errors.New("Card Declined"))
  	} else {
  		return errors.New(fmt.Sprintf("Unknown payment error: %s", payResp.Message))
  	}
  }

	return err
}

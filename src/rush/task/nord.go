package task

import (
	"bytes"
  "fmt"
  "github.com/avast/retry-go"
  "time"
  "os"
  "regexp"
  "github.com/pkg/errors"
  "io/ioutil"
	"encoding/json"
	"rush/net/http"
	"github.com/PuerkitoBio/goquery"
	"net/url"
  // "log"
  "strings"
)

func (t *CheckoutTask) nordGetProduct(productUrl *url.URL) (*http.Response, error) {
  headerOrder := []string {
    "x-y8s6k3db-f",
     "x-y8s6k3db-d",
     "accept-language",
     "x-akamai-config-log-detail",
     "x-forwarded-for",
     "nord-country-code",
     "x-forwarded-proto",
     "x-akamai-edgescape",
     "true-client-ip",
     "pragma",
     "x-y8s6k3db-b",
     "x-y8s6k3db-c",
     "ak_sh_tcip_hash",
     "ak_sh_tcip",
     "x-y8s6k3db-a",
     "upgrade-insecure-requests",
     "user-agent",
     "accept",
     "x-cloud-trace-context",
     "x-edgeconnect-session-id",
     "istl-infinite-loop",
     "akamai-origin-hop",
     "x-y8s6k3db-z",
     "sec-fetch-site",
     "sec-fetch-mode",
     "sec-fetch-dest",
     "referer",
     "accept-encoding",
  }
  headers := [][2]string {
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "none"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
  }
  resp, err := t.doReq(t.client, t.makeReq("GET", productUrl, &headers, &headerOrder, nil))
  if err != nil {
  	return nil, err
  }
  body, err := readBodyBytes(resp)
  if err != nil {
  	return nil, err
  }

  doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
  if err != nil {
  	return nil, err
  }

  akScript := doc.Find("script").Eq(1).Text()

  AK_HEADER_BLACKLIST := []string {
    "accept-charset",
    "accept-encoding",
    "access-control-request-headers",
    "access-control-request-method",
    "connection",
    "content-length",
    "cookie",
    "cookie2",
    "date",
    "dnt",
    "expect",
    "host",
    "keep-alive",
    "origin",
    "referer",
    // added
    "accept-language",
    "sec-fetch-user",
    "upgrade-insecure-requests",
    "accept",
    "sec-fetch-site",
    "set-fetch-user",
    "sec-fetch-mode",
    "sec-fetch-dest",
    // end added
    "te",
    "trailer",
    "transfer-encoding",
    "upgrade",
    "user-agent",
    "via",
  }
  akHeaderRe := regexp.MustCompile(`var originalHeaders = \[(.+?),\]`)
  akHeaderMatches := akHeaderRe.FindStringSubmatch(akScript)
  if len(akHeaderMatches) != 2 {
    return nil, errors.New("Missing Edge Headers")
  }
  var akHeaderArr []string
  err = json.Unmarshal([]byte("[" + akHeaderMatches[1] + "]"), &akHeaderArr)
  if err != nil {
    return nil, err
  }
  akHeaders := make([][2]string, 0)
  akKey := ""
  for idx, s := range akHeaderArr {
    if idx % 2 == 0 {
      akKey = s
    } else if !Includes(AK_HEADER_BLACKLIST, strings.ToLower(akKey)) {
      akHeaders = append(akHeaders, [2]string{akKey, s})
    }
  }

  znScript := doc.Find("script").Eq(0).Text()
  znScript = "export default function(window, document, dispatchEvent) {" + znScript + "}"
  tmpfile, err := ioutil.TempFile("", "")
  if err != nil {
    return nil, err
  }
  defer os.Remove(tmpfile.Name()) // clean up

  if _, err := tmpfile.Write([]byte(znScript)); err != nil {
    return nil, err
  }
  if err := tmpfile.Close(); err != nil {
    return nil, err
  }
  znHeaders, err := GetZeronaughtHeaders(tmpfile.Name())

  headers = [][2]string {
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"ISTL-INFINITE-LOOP", "1"},
    {"referer", productUrl.String()},
  }
  headers = append(headers, akHeaders...)
  headers = append(headers, znHeaders...)

  resp, err = t.doReq(t.client, t.makeReq("GET", productUrl, &headers, &headerOrder, nil))
  if err != nil {
    return nil, err
  }
  redir := resp.Header.Get("istl-redirect-to")
  if redir != "" {
    if redir == "/not-found" {
      return nil, errors.New("Product not found")
    } else {
      redirUrl, err := url.Parse(resp.Header.Get("istl-redirect-to"))
      if err != nil {
        return nil, err
      }
      return t.doReq(t.client, t.makeReq("GET", redirUrl, &headers, &headerOrder, nil))
    }
  } else {
    return resp, err
  }
}

func (t *CheckoutTask) nordShopperId() (string, error) {
  nordCookie := getCookieValue("nordstrom", t.Url, t.client.Jar)
  vals, err := url.ParseQuery(nordCookie)
  if err != nil {
    return "", err
  }
  if shopperId, ok := vals["shopperid"]; ok {
    return shopperId[0], nil
  } else {
    return "", errors.New("Missing shopper id")
  }
}

func (t *CheckoutTask) nordDoCheckout(productId string, sku int) error {
  var err error
  jar := t.client.Jar
  t.client, err = t.newHttpClient()
  if err != nil {
    return err
  }
  t.client.Jar = jar
  t.StatusCh <- Status{Message:"Adding to Cart"}
  _, err = t.nordAtc(productId, sku)
  if err != nil {
    return err
  }
  t.StatusCh <- Status{Message:"Fetching Cart"}
  order, err := t.nordStartOrder()
  if err != nil {
    return err
  }
  _, err = t.nordCheckoutShipping(*order)
  return err
}

func (t *CheckoutTask) nordCheckoutShipping(order NordOrder) (*http.Response, error) {
  t.StatusCh <- Status{Message:"Checkout 1/4"}
  url_, err := url.Parse("https://secure.nordstrom.com/api/checkout/web/checkout/v1/order?postalcode=" + t.Profile.ShippingAddress.Zip)
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Content-Length",
    "X-y8S6k3DB-f",
    "X-y8S6k3DB-b",
    "x-forter-token",
    "X-y8S6k3DB-a",
    "x-nor-appiid",
    "User-Agent",
    "experiments",
    "Content-Type",
    "X-y8S6k3DB-d",
    "X-y8S6k3DB-c",
    "nord-country-code",
    "x-shopper-id",
    "X-y8S6k3DB-z",
    "tracecontext",
    "x-shopper-token",
    "Accept",
    "Origin",
    "Sec-Fetch-Site",
    "Sec-Fetch-Mode",
    "Sec-Fetch-Dest",
    "Referer",
    "Accept-Encoding",
    "Accept-Language",
    "Cookie",
  }
  shopperId, err := t.nordShopperId()
  if err != nil {
    return nil, err
  }
  shopperToken, err := t.nordShopperToken()
  if err != nil {
    return nil, err
  }
  commonHeaders := [][2]string {
    {"Connection", "keep-alive"},
    {"Content-Type", "application/json"},
    {"nord-country-code", "US"},
    {"x-shopper-id", shopperId},
    {"x-shopper-token", shopperToken},
    {"Accept", "*/*"},
    {"Origin", "https://secure.nordstrom.com"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", "https://secure.nordstrom.com/checkout"},
  }
  ckHeaders, ckCookies, err := GetCkZeronaughtHeaders(t.Proxy.String())
  if err != nil {
    return nil, err
  }
  headers := append(commonHeaders, ckHeaders...)
  for key, value := range ckCookies {
    t.client.Jar.SetCookies(url_, []*http.Cookie{ &http.Cookie{ Name: key, Value: value, }})
  }

  bodyStructure := map[string]interface{} {
  "id": order.Id,
  "items": order.Items,
  "contact": map[string]interface{} {
      "email": "",
      "phone": "",
      "marketing": false,
    },
    "pickupPerson": map[string]interface{} {
      "firstName": "",
      "lastName": "",
    },
    "employee": map[string]interface{} {

    },
    "shippingAddress": map[string]interface{} {
      "storeNumber": "",
      "firstName": "",
      "middleName": "",
      "lastName": "",
      "addressLine1": "",
      "addressLine2": "",
      "city": "",
      "state": "",
      "postalCode": "",
      "countryCode": "US",
    },
    "payment": map[string]interface{} {
      "creditCard": map[string]interface{} {
        "id": nil,
        "type": "",
        "lastFour": "",
        "lowValueToken": "",
        "expirationMonth": "",
        "expirationYear": "",
        "isExpired": false,
        "saveToAccount": false,
        "saveAsDefault": false,
        "billingAddress": map[string]interface{} {
          "firstName": "",
          "middleName": "",
          "lastName": "",
          "addressLine1": "",
          "addressLine2": "",
          "city": "",
          "state": "",
          "postalCode": "",
          "countryCode": "US",
          "default": true,
        },
      },
      "payPal": nil,
      "giftCards": []interface{} {

      },
      "manualNotes": []interface{} {

      },
      "systematicNotes": []interface{} {

      },
    },
    "rewards": map[string]interface{} {
      "applyBonusPoints": false,
    },
    "promoCodes": []interface{} {

    },
    "gwp": []interface{} {

    },
    "samples": []interface{} {

    },
    "metadata": []interface{} {

    },
    "validate": []interface{} {
      "none",
    },
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  _, err = t.doReq(t.client, t.makeReq("PUT", url_, &headers, &headerOrder, &body))
  if err != nil {
    return nil, err
  }

  ckHeaders, _, err = GetCkZeronaughtHeaders(t.Proxy.String())
  if err != nil {
    return nil, err
  }
  headers = append(commonHeaders, ckHeaders...)

  t.StatusCh <- Status{Message:"Checkout 2/4 (Shipping)"}
  bodyStructure = map[string]interface{} {
  "id": order.Id,
  "items": order.Items,
   "contact": map[string]interface{} {
     "email": "",
     "phone": "",
     "marketing": false,
   },
   "pickupPerson": map[string]interface{} {
     "firstName": "",
     "lastName": "",
   },
    "employee": map[string]interface{} {

    },
    "shippingAddress": map[string]interface{} {
      "storeNumber": "",
      "id": "",
      "firstName": t.Profile.ShippingAddress.FirstName,
      "middleName": "",
      "lastName": t.Profile.ShippingAddress.LastName,
      "addressLine1": t.Profile.ShippingAddress.Address1,
      "addressLine2": t.Profile.ShippingAddress.Address2,
      "city": t.Profile.ShippingAddress.City,
      "state": t.Profile.ShippingAddress.State,
      "postalCode": t.Profile.ShippingAddress.Zip,
      "countryCode": t.Profile.ShippingAddress.Country,
      "default": false,
      "saveAsDefault": false,
    },
    "payment": map[string]interface{} {
      "creditCard": map[string]interface{} {
        "id": "",
        "type": "",
        "firstSix": "",
        "lastFour": "",
        "lowValueToken": "",
        "expirationMonth": "",
        "expirationYear": "",
        "billingAddress": map[string]interface{} {
          "id": "",
          "firstName": "",
          "middleName": "",
          "lastName": "",
          "addressLine1": "",
          "addressLine2": "",
          "city": "",
          "state": "",
          "postalCode": "",
          "countryCode": "US",
          "default": true,
          "saveAsDefault": false,
        },
        "isExpired": false,
        "default": false,
        "saveAsDefault": false,
      },
      "payPal": nil,
      "giftCards": []interface{} {

      },
      "manualNotes": []interface{} {

      },
      "systematicNotes": []interface{} {

      },
    },
    "rewards": map[string]interface{} {
      "applyBonusPoints": false,
    },
    "promoCodes": []interface{} {

    },
    "gwp": []interface{} {

    },
    "samples": []interface{} {

    },
    "metadata": []interface{} {

    },
    "validate": []interface{} {
      "none",
    },
  }
  body, err = json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  resp, err := t.doReq(t.client, t.makeReq("PUT", url_, &headers, &headerOrder, &body))
  if err != nil {
    return nil, err
  }
  body, err = readBodyBytes(resp)
  if err != nil {
    return nil, err
  }

  ckHeaders, _, err = GetCkZeronaughtHeaders(t.Proxy.String())
  if err != nil {
    return nil, err
  }
  headers = append(commonHeaders, ckHeaders...)

  t.StatusCh <- Status{Message:"Checkout 3/4 (Billing)"}

  yearStr := fmt.Sprintf("%d", t.Profile.Card.ExpYear)
  bodyStructure = map[string]interface{} {
    "id": order.Id,
    "items": order.Items,
    "contact": map[string]interface{} {
      "email": t.Profile.Email,
      "phone": t.Profile.ShippingAddress.NordPhone(),
      "marketing": false,
    },
    "pickupPerson": map[string]interface{} {
      "firstName": t.Profile.ShippingAddress.FirstName,
      "lastName": t.Profile.ShippingAddress.LastName,
    },
    "employee": map[string]interface{} {

    },
    "shippingAddress": map[string]interface{} {
      "id": "",
      "default": false,
      "saveAsDefault": false,
      "storeNumber": "",
      "firstName": t.Profile.ShippingAddress.FirstName,
      "middleName": "",
      "lastName": t.Profile.ShippingAddress.LastName,
      "addressLine1": t.Profile.ShippingAddress.Address1,
      "addressLine2": t.Profile.ShippingAddress.Address2,
      "city": t.Profile.ShippingAddress.City,
      "state": t.Profile.ShippingAddress.State,
      "postalCode": t.Profile.ShippingAddress.Zip,
      "countryCode": t.Profile.ShippingAddress.Country,
    },
    "payment": map[string]interface{} {
        "creditCard": map[string]interface{} {
          "id": "",
          "type": "MC", // TODO what are other types
          "firstSix": t.Profile.Card.Number[0:6],
          "lastFour": t.Profile.Card.Number[len(t.Profile.Card.Number)-4:],
          "lowValueToken": "",
          "expirationMonth": fmt.Sprintf("%02d", t.Profile.Card.ExpMonth),
          "expirationYear": yearStr[len(yearStr)-2:],
          "isExpired": false,
          "saveAsDefault": false,
          "billingAddress": map[string]interface{} {
            "id": "",
            "firstName": t.Profile.BillingAddress.FirstName,
            "middleName": "",
            "lastName": t.Profile.BillingAddress.LastName,
            "addressLine1": t.Profile.BillingAddress.Address1,
            "addressLine2": t.Profile.BillingAddress.Address2,
            "city": t.Profile.BillingAddress.City,
            "state": t.Profile.BillingAddress.State,
            "postalCode": t.Profile.BillingAddress.Zip,
            "countryCode": t.Profile.BillingAddress.Country,
            "default": true,
            "saveAsDefault": false,
          },
          "default": false,
          "securityCode": t.Profile.Card.Cvc,
        },
        "payPal": nil,
        "giftCards": []interface{} {

        },
        "manualNotes": []interface{} {

        },
        "systematicNotes": []interface{} {

        },
      },
    "rewards": map[string]interface{} {
      "applyBonusPoints": false,
    },
    "promoCodes": []interface{} {

    },
    "gwp": []interface{} {

    },
    "samples": []interface{} {

    },
    "metadata": []interface{} {

    },
    "validate": []interface{} {
      "items",
      "shippingAddress",
      "contact",
      "phone",
      "email",
      "notes",
      "giftCards",
      "systematicNotes",
      "estimatedTotal",
    },
  }
  body, err = json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  _, err = t.doReq(t.client, t.makeReq("PUT", url_, &headers, &headerOrder, &body))
  if err != nil {
    return nil, err
  }

  ckHeaders, _, err = GetCkZeronaughtHeaders(t.Proxy.String())
  if err != nil {
    return nil, err
  }
  headers = append(commonHeaders, ckHeaders...)

  t.StatusCh <- Status{Message:"Encrypting Card"}
  url_, err = url.Parse("https://secure.nordstrom.com/api/checkout/web/checkout/v1/order/submit?postalcode=" + t.Profile.ShippingAddress.Zip)
  if err != nil {
    return nil, err
  }
  cardToken, err := GetEprotectToken(t.Profile.Card.Number, t.Profile.Card.Cvc)
  if err != nil {
    return nil, err
  }
  t.StatusCh <- Status{Message:"Checkout 4/4 (Payment)"}
  bodyStructure = map[string]interface{} {
    "id": order.Id,
    "items": order.Items,
    "contact": map[string]interface{} {
      "email": t.Profile.Email,
      "phone": t.Profile.ShippingAddress.NordPhone(),
      "marketing": false,
    },
    "pickupPerson": map[string]interface{} {
      "firstName": t.Profile.ShippingAddress.FirstName,
      "lastName": t.Profile.ShippingAddress.LastName,
    },
    "employee": map[string]interface{} {

    },
    "shippingAddress": map[string]interface{} {
      "id": "",
      "default": false,
      "saveAsDefault": false,
      "storeNumber": "",
      "firstName": t.Profile.ShippingAddress.FirstName,
      "middleName": "",
      "lastName": t.Profile.ShippingAddress.LastName,
      "addressLine1": t.Profile.ShippingAddress.Address1,
      "addressLine2": t.Profile.ShippingAddress.Address2,
      "city": t.Profile.ShippingAddress.City,
      "state": t.Profile.ShippingAddress.State,
      "postalCode": t.Profile.ShippingAddress.Zip,
      "countryCode": t.Profile.ShippingAddress.Country,
    },
    "payment": map[string]interface{} {
      "creditCard": map[string]interface{} {
        "id": "",
        "type": "MC", // TODO what are other types
        "firstSix": t.Profile.Card.Number[0:6],
        "lastFour": t.Profile.Card.Number[len(t.Profile.Card.Number)-4:],
        "lowValueToken": cardToken,
        "expirationMonth": fmt.Sprintf("%02d", t.Profile.Card.ExpMonth),
        "expirationYear": yearStr[len(yearStr)-2:],
        "isExpired": false,
        "saveAsDefault": false,
        "billingAddress": map[string]interface{} {
          "id": "",
          "firstName": t.Profile.BillingAddress.FirstName,
          "middleName": "",
          "lastName": t.Profile.BillingAddress.LastName,
          "addressLine1": t.Profile.BillingAddress.Address1,
          "addressLine2": t.Profile.BillingAddress.Address2,
          "city": t.Profile.BillingAddress.City,
          "state": t.Profile.BillingAddress.State,
          "postalCode": t.Profile.BillingAddress.Zip,
          "countryCode": t.Profile.BillingAddress.Country,
          "default": true,
          "saveAsDefault": false,
        },
        "default": false,
        "securityCode": t.Profile.Card.Cvc,
      },
      "payPal": nil,
      "giftCards": []interface{} {

      },
      "manualNotes": []interface{} {

      },
      "systematicNotes": []interface{} {

      },
    },
    "rewards": map[string]interface{} {
      "applyBonusPoints": false,
    },
    "promoCodes": []interface{} {

    },
    "gwp": []interface{} {

    },
    "samples": []interface{} {

    },
    "metadata": []interface{} {
      map[string]interface{} {
            "name": "orderSource",
            "value": "standard",
          },
    },
  }
  body, err = json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  resp, err = t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  var orderResp NordOrderResponse
  if resp != nil {
    body, err := readBodyBytes(resp)
    if err != nil {
      return nil, err
    }
    err = json.Unmarshal(body, &orderResp)
    if err != nil {
      return nil, err
    }
  }
  var skuDetail NordSkuDetail = orderResp.SkuDetails[order.Items[0]["rmsSku"].(string)]
  if err != nil {
    if resp.StatusCode == 400 {
      if len(orderResp.Exceptions) > 0 {
        if orderResp.Exceptions[0].Code == "SubmitPaymentAuthDecline" {
          go t.DeclineWebhook("Nordstrom",
            skuDetail.Name,
            skuDetail.Style.Size)
          t.StatusCh <- Status{Message: orderResp.Exceptions[0].Message}
          return resp, nil
        } else {
          return nil, errors.New(orderResp.Exceptions[0].Message)
        }
      }
    }
    return nil, err
  } else {
    go t.SuccessWebhook("Nordstrom",
      skuDetail.Name,
      skuDetail.Style.Size)
    return resp, err
  }
}


func (t *CheckoutTask) nordStartOrder() (*NordOrder, error) {
  url_, err := url.Parse("https://secure.nordstrom.com/api/checkout/web/checkout/v1/shopping-bag/checkout?postalcode=" + t.Profile.ShippingAddress.Zip)
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "x-forter-token",
    "x-nor-appiid",
    "User-Agent",
    "experiments",
    "Content-Type",
    "nord-country-code",
    "x-shopper-id",
    "x-shopper-token",
    "tracecontext",
    "Accept",
    "Sec-Fetch-Site",
    "Sec-Fetch-Mode",
    "Sec-Fetch-Dest",
    "Referer",
    "Accept-Encoding",
    "Accept-Language",
    "Cookie",
  }
  shopperId, err := t.nordShopperId()
  if err != nil {
    return nil, err
  }
  shopperToken, err := t.nordShopperToken()
  if err != nil {
    return nil, err
  }
  headers := [][2]string {
    {"Connection", "keep-alive"},
    {"Content-Type", "application/json"},
    {"nord-country-code", "US"},
    {"x-shopper-id", shopperId},
    {"x-shopper-token", shopperToken},
    {"Accept", "*/*"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", "https://secure.nordstrom.com/checkout"},
  }
  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
  if err != nil {
    return nil, err
  }
  body, err := readBodyBytes(resp)
  if err != nil {
    return nil, err
  }
  var orderResp NordOrderResponse
  err = json.Unmarshal(body, &orderResp)
  if err != nil {
    return nil, err
  }
  if orderResp.Order.Id != "" {
    return &orderResp.Order, nil
  } else {
    return nil, errors.New("Missing order ID")
  }
}


type NordOrder struct {
  Id string `json:"id"`
  Items []map[string]interface{} `json:"items"`
}

type NordOrderException struct {
  Code string `json:"code"`
  Message string `json:"message"`
}
type NordSkuStyle struct {
  Color string `json:"color"`
  Size string `json:"size"`
}
type NordSkuDetail struct {
  Name string `json:"name"`
  Style NordSkuStyle `json:"style"`
}
type NordOrderResponse struct {
  Order NordOrder `json:"order"`
  SkuDetails map[string]NordSkuDetail `json:"skuDetails"`
  Exceptions []NordOrderException `json:"exceptions"`
}

func (t *CheckoutTask) nordShopperToken() (string, error) {
  ckSplit := strings.Split(getCookieValue("shoppertoken", t.Url, t.client.Jar), "shopperToken=")
  if len(ckSplit) != 2 {
    return "", errors.New("Missing shopper token")
  }
  return ckSplit[1], nil
}

func (t *CheckoutTask) nordAtc(productId string, sku int) (*http.Response, error) {
  url_, err := url.Parse("https://secure.nordstrom.com/api/checkout/web/checkout/v1/shopping-bag/items?postalcode=" + t.Profile.ShippingAddress.Zip)
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Content-Length",
    "X-y8S6k3DB-f",
    "x-forter-token",
    "identified-bot",
    "experiments",
    "nord-country-code",
    "x-shopper-id",
    "tracecontext",
    "X-y8S6k3DB-b",
    "X-y8S6k3DB-c",
    "X-y8S6k3DB-a",
    "X-y8S6k3DB-d",
    "User-Agent",
    "x-nor-appiid",
    "Content-Type",
    "X-y8S6k3DB-z",
    "x-shopper-token",
    "Accept",
    "Origin",
    "Sec-Fetch-Site",
    "Sec-Fetch-Mode",
    "Sec-Fetch-Dest",
    "Referer",
    "Accept-Encoding",
    "Accept-Language",
    "Cookie",
  }


  shopperId, err := t.nordShopperId()
  if err != nil {
    return nil, err
  }
  shopperToken, err := t.nordShopperToken()
  if err != nil {
    return nil, err
  }
  headers := [][2]string {
    {"Connection", "keep-alive"},
    {"identified-bot", "False"},
    {"nord-country-code", "US"},
    {"x-shopper-id", shopperId},
    {"Content-Type", "application/json"},
    {"x-shopper-token", shopperToken},
    {"Accept", "*/*"},
    {"Origin", "https://www.nordstrom.com"},
    {"Sec-Fetch-Site", "same-site"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", t.Url.String()},
  }
  ckHeaders, ckCookies, err := GetCkZeronaughtHeaders(t.Proxy.String())
  if err != nil {
    return nil, err
  }
  headers = append(headers, ckHeaders...)
  for key, value := range ckCookies {
    t.client.Jar.SetCookies(url_, []*http.Cookie{ &http.Cookie{ Name: key, Value: value, }})
  }
  bodyStructure := []interface{} {
  map[string]interface{} {
    "rmsSku": fmt.Sprintf("%d", sku),
    "id": productId, // root sku
    "quantity": 1,
    "metadata": []interface{} {
      map[string]interface{} {
        "name": "Breadcrumb",
        "value": "Home/All Results",
      },
    },
  },
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  return t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
}


type NordSku struct {
  Id int `json:"id"`
  RmsSkuId int `json:"rmsSkuId"`
  Size string `json:"sizeId"`
  Width string `json:"widthId"`
  Available bool `json:"isAvailable"`
  ColorId int `json:"colorId"`
}

type NordProdSkus struct {
  ById map[string]NordSku `json:"byId"`
}
type NordProdStyle struct {
  Skus NordProdSkus `json:"skus"`
}

type NordProdStyles struct {
  Data map[string]NordProdStyle `json:"data"`
}

type NordProdConfig struct {
  Styles NordProdStyles `json:"stylesById"`
}

func (t *CheckoutTask) NordCheckout() error {
  // return t.Retry(func() error {
    t.Retry(func() error { fmt.Printf("%v\n", time.Second); return nil }, retry.Attempts(1))
    t.StatusCh <- Status{Message:"Waiting for Product"}
    resp, err := t.nordGetProduct(t.Url)
    if err != nil {
    	return err
    }
    prodRespBody, err := readBody(resp)
    if err != nil {
      return err
    }
    urlSplit := strings.Split(t.Url.Path, "/")
    productId := urlSplit[len(urlSplit)-1]

    prodConfigRe := regexp.MustCompile(`<script>window.__INITIAL_CONFIG__ = (.+?)</script>`)
    prodConfigMatches := prodConfigRe.FindStringSubmatch(prodRespBody)
    if len(prodConfigMatches) != 2 {
      return errors.New("Rejected")
    }
    var prodConfig NordProdConfig
    err = json.Unmarshal([]byte(prodConfigMatches[1]), &prodConfig)
    if err != nil {
      return err
    }

    skus := []NordSku{}
    for _, style := range prodConfig.Styles.Data {
      for _, v := range style.Skus.ById {
        skus = append(skus, v)
      }
    }
    // TODO size selection
    if len(skus) == 0 {
      return errors.New("Out of stock")
    }
  	return t.nordDoCheckout(productId, skus[0].RmsSkuId)
  // }, retry.Delay(256 * time.Millisecond), retry.Attempts(1e9), retry.LastErrorOnly(true))
}

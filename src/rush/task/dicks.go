package task

import (
  // "bytes"
  "os"
  "fmt"
  "regexp"
  "time"
  // "rush/net/http/cookiejar"
  "encoding/json"
  "github.com/pkg/errors"
  "rush/net/http"
  "log"
  "strings"
  "math/rand"
  "net/url"
  "github.com/avast/retry-go"
)

func (t *CheckoutTask) DsgLoginAuth0() (*http.Response, error) {
  t.DefaultProto = "HTTP/1.1"
  // t.HeaderCaseFn = strings.Title
  t.HeaderBlacklist = map[string]bool {
    "sec-fetch-site": true,
    "sec-fetch-mode": true,
    "sec-fetch-dest": true,
  }
  url_, err := url.Parse("https://sso.dickssportinggoods.com/oauth/token")
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "Accept-Language",
    "Auth0-Client",
    "Content-Type",
    "Content-Length",
    "Host",
    "Connection",
    "Accept-Encoding",
    "User-Agent",
  }
  headers := [][2]string {
    {"Accept-Language", "en_US"},
    {"Auth0-Client", "eyJuYW1lIjoiQXV0aDAuQW5kcm9pZCIsImVudiI6eyJhbmRyb2lkIjoiMjgifSwidmVyc2lvbiI6IjEuMjMuMCJ9"},
    {"Content-Type", "application/json; charset=utf-8"},
    {"user-agent", "okhttp/2.7.5"},
    {"Connection", "Keep-Alive"},
    {"Accept-Encoding", "gzip"},
  }
  bodyStructure := map[string]interface{} {
  "password": "1FastBot!",
  "audience": "dsg-jwts",
  "grant_type": "http://auth0.com/oauth/grant-type/password-realm",
  "scope": "read:identity offline_access",
  "realm": "dsg-username-password-auth",
  "client_id": "a39zfGCrJx7WQ5i2mH2Xz7kxdLsm6hKG",
  "username": "andyh2@me.com",
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  if err != nil {
    return nil, err
  }
  var authresp = make(map[string]interface{})
  err = readRespJsonDst(resp, &authresp)
  if err != nil {
    return nil, err
  }
  tok := authresp["access_token"].(string)

  t.DsgAppGetUserInfo(tok)
  return nil, nil
}

func (t *CheckoutTask) DsgAppGetUserInfo(accessToken string) (*http.Response, error) {
  t.DefaultProto = "HTTP/2"
  t.HeaderBlacklist = map[string]bool {
    "sec-fetch-site": true,
    "sec-fetch-mode": true,
    "sec-fetch-dest": true,
    "accept-language": true,
  }
  t.UserAgent = "DSG Android App/4.7.9 Dalvik/2.1.0 (Linux; U; Android 9; ONEPLUS A5010 Build/PKQ1.180716.001)"

  url_, err := url.Parse("https://move.dickssportinggoods.com/user/api/v2/info")
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "partner_key",
    "secret_key",
    "authorization",
    "accept",
    "user-agent",
    "x-acf-sensor-data",
    "accept-encoding",
  }
  s, _ := t.getSensor()
  headers := [][2]string {
    {"partner_key", "mobile_android"},
    {"secret_key", "ZtjmnXy3RLb8P64n17KQ"},
    {"authorization", fmt.Sprintf("Bearer %s", accessToken)},
    {"accept", "application/json"},
    {"x-acf-sensor-data", s},
    {"accept-encoding", "gzip"},
  }
  return t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
}


var (
  DICKS_BASE = "https://www.dickssportinggoods.com/"
)

func (t *CheckoutTask) hasValidAbckDicks(abck string) bool {
  return strings.Contains(abck, "==~-1~-1~-1")
}

type DicksOrderItem struct {
  OrderItemId string `json:"orderItemId"`
}
type DicksCart struct {
  OrderId string `json:"orderId"`
  OrderItem []DicksOrderItem `json:"orderItem"`
}

type DicksSkuDetail struct {
  Id string `json:"id"`
  Name string `json:"name"`
  Value string `json:"value"`
}
type DicksSkus struct {
  PartNumbers []string `json:"skuIdList"`
  Skus [][]DicksSkuDetail `json:"availableCombinations"`
}
type DicksAddress struct {
  AddressLines []string `json:"addressLines"`
  City string `json:"city"`
  Country string `json:"country"`
  State string `json:"state"`
  ZipCode string `json:"zipCode"`
}
type DicksAvsResult struct {
  Decision string `json:"decision"`
  ReviewNeeded bool `json:"reviewNeeded"`
}
type DicksVerifyAddressResult struct {
  AvsResult DicksAvsResult `json:"avsResult"`
  UnverifiedAddress DicksAddress `json:"unverifiedAddress"`
  VerifiedAddress DicksAddress `json:"verifiedAddress"`
}
func (t *CheckoutTask) dicksVerifyAddress(orderId string, address Address) (DicksAddress, error) {
  verifyUrl, err := url.Parse("https://avs.dickssportinggoods.com/avs/verifyAddress")
  if err != nil {
    return DicksAddress{}, err
  }

  client, err := t.newHttpClient()
  if err != nil {
    return DicksAddress{}, err
  }

  client.Jar.SetCookies(verifyUrl, t.client.Jar.Cookies(verifyUrl))
  if err != nil {
    return DicksAddress{}, err
  }
  headers := [][2]string {
    {"accept", "application/json"},
    {"content-type", "application/json"},
    {"origin", "https://www.dickssportinggoods.com"},
  }
  addressLines := []string{address.Address1,}
  if address.Address2 != "" {
    addressLines = append(addressLines, address.Address2)
  }
  bodyMap := map[string]interface{} {
    "originIdentifier": "CHECKOUT-UI",
    "referenceCode": orderId,
    "unverifiedAddress": DicksAddress {
      AddressLines: addressLines,
      City: address.City,
      Country: address.Country,
      State: address.State,
      ZipCode: address.Zip,
    },
  }
  body, err := json.Marshal(bodyMap)
  if err != nil {
    return DicksAddress{}, err
  }
  resp, err := t.doReq(client, t.makeReq("POST", verifyUrl, &headers, nil, &body))
  respBytes, err := readBodyBytes(resp)
  if err != nil {
    return DicksAddress{}, err
  }
  var result DicksVerifyAddressResult
  err = json.Unmarshal(respBytes, &result)
  if err != nil {
    return DicksAddress{}, err
  }
  if result.AvsResult.Decision != "ACCEPT" {
    return DicksAddress{}, errors.New("Address Rejected")
  }
  return result.VerifiedAddress, nil
}


func (t *CheckoutTask) DsgHandshake(handshake string, value string) (error) {
  url_, err := url.Parse(fmt.Sprintf("https://www.dickssportinggoods.com/api/v1/handshakes?%s=%s", handshake, value))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Content-Length",
    "X-INSTANA-T",
    "HS-2",
    "HS-0",
    "X-INSTANA-L",
    "elastic-apm-traceparent",
    "HS-3",
    "accept-language",
    "Accept",
    "res",
    "HS-1",
    "User-Agent",
    "X-INSTANA-S",
    "HS-4",
    "Origin",
    "Sec-Fetch-Site",
    "Sec-Fetch-Mode",
    "Sec-Fetch-Dest",
    "Referer",
    "Accept-Encoding",
    "Cookie",
  }
  headers := [][2]string {
    {"Connection", "keep-alive"},
    {"content-length", "0"},
    {"X-INSTANA-T", "935441e66e70284b"},
    {"HS-2", "47ee"},
    {"HS-0", "518cc9b5"},
    {"X-INSTANA-L", "1,correlationType=web;correlationId=935441e66e70284b"},
    {"elastic-apm-traceparent", "00-ce48475248f273f8e52f080f771e246b-8fea67d8631e9f91-01"},
    {"HS-3", "a481"},
    {"accept-language", "en-US,de-CH"},
    {"Accept", "application/json, text/plain, */*"},
    {"res", "20716718"},
    {"HS-1", "fc46"},
    {"X-INSTANA-S", "935441e66e70284b"},
    {"HS-4", "1b0b98ff9c0b"},
    {"Origin", "https://www.dickssportinggoods.com"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", t.Url.String()},
    {"Accept-Encoding", "gzip, deflate, br"},
  }
  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, nil))
  DiscardResp(resp)
  return err
}
func (t *CheckoutTask) DsgAtc(pid string) (error) {
  url_, err := url.Parse("https://www.dickssportinggoods.com/api/v1/carts/contents/" + pid + "?qty=1")
  if err != nil {
    return err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Content-Length",
    "Accept",
    "User-Agent",
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
    {"Accept", "*/*"},
    {"Origin", "https://www.dickssportinggoods.com"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", t.Url.String()},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,de-CH"},
  }
  resp, err := t.DicksDoRequest(t.makeReq("PUT", url_, &headers, &headerOrder, nil))
  DiscardResp(resp)
  if err != nil {
    return err
  }
  if resp.StatusCode == 200 && strings.ToLower(resp.Header.Get("Server")) != "akamaighost" {
    return nil
  } else {
    return ErrAtcFailed
  }
}

type DsgErrorResponse struct {
  Timestamp   int64  `json:"timestamp"`
  Status      int    `json:"status"`
  Error       string `json:"error"`
  Message     string `json:"message"`
  Path        string `json:"path"`
  UIMessaging struct {
    Identifier string `json:"identifier"`
    Header     string `json:"header"`
    Message    string `json:"message"`
    Button     struct {
      Text        string `json:"text"`
      RedirectURI string `json:"redirect_uri"`
      Action      string `json:"action"`
    } `json:"button"`
  } `json:"ui_messaging"`
}

type DsgCheckoutSession struct {
  CheckoutKey    string    `json:"checkout_key"`
  CheckoutSchema string    `json:"checkout_schema"`
  Channel        string    `json:"channel"`
  Store          string    `json:"store"`
  Created        time.Time `json:"created"`
  Cart           struct {
    Items []struct {
      Sku            string   `json:"sku"`
      Ecode          string   `json:"ecode"`
      UnitPrice      string   `json:"unit_price"`
      Quantity       int      `json:"quantity"`
      TotalUnitPrice string   `json:"total_unit_price"`
      ListPrice      string   `json:"list_price"`
      Description    []string `json:"description"`
      Manufacturer   string   `json:"manufacturer"`
      Images         []string `json:"images"`
      Analytics      struct {
        OrderDiscount    string `json:"order_discount"`
        ItemDiscount     string `json:"item_discount"`
        ShippingDiscount string `json:"shipping_discount"`
        ShippingPaid     string `json:"shipping_paid"`
        TaxPaid          string `json:"tax_paid"`
      } `json:"analytics"`
      ExternalKey string `json:"external_key"`
    } `json:"items"`
    ExternalKey    string `json:"external_key"`
    ExternalSource string `json:"external_source"`
    PaypalEligible bool   `json:"paypal_eligible"`
  } `json:"cart"`
  Promotion struct {
    Cart struct {
      Promotions []struct {
        Description string `json:"description"`
        ExternalKey string `json:"external_key"`
      } `json:"promotions"`
      ReferenceKey   string `json:"reference_key"`
      ExternalSource string `json:"external_source"`
    } `json:"cart"`
  } `json:"promotion"`
  Pricing struct {
    Currency       string `json:"currency"`
    TotalUnit      string `json:"total_unit"`
    TotalShipping  string `json:"total_shipping"`
    TotalTax       string `json:"total_tax"`
    TotalDiscounts string `json:"total_discounts"`
    Total          string `json:"total"`
    ExternalKey    string `json:"external_key"`
    ExternalSource string `json:"external_source"`
  } `json:"pricing"`
  OrderSummary []struct {
    Type          string `json:"type"`
    DisplayText   string `json:"display_text,omitempty"`
    DisplayAmount string `json:"display_amount,omitempty"`
    Removable     bool   `json:"removable"`
    Informational bool   `json:"informational"`
  } `json:"order_summary"`
}

var ErrInvalidCheckoutSession = errors.New("Invalid checkout session")

func (t *CheckoutTask) DsgStartCheckout() (*DsgCheckoutSession, error) {
  // url_, _ :=  url.Parse("https://www.dickssportinggoods.com/OrderItemDisplay?calculationUsageId=-1&calculationUsageId=-2&calculationUsageId=-3&calculationUsageId=-7&updatePrices=1&doConfigurationValidation=Y&catalogId=12301&orderId=.&langId=-1&storeId=15108")
  // cartResp, err := t.DicksDoRequest(t.makeGetReq(url_, nil))
  // if err != nil{
  //   return nil, err
  // }
  // cartRespBody, err := readBody(cartResp)
  // if err != nil {
  //   return nil, err
  // }
  // log.Println(cartRespBody)

  // t.EnsureValidAbck(t.Config["akUrl"], t.hasValidAbckDicks)

  // url_, err = url.Parse("https://www.dickssportinggoods.com/api/v1/checkouts/order-summary")
  // if err != nil {
  //   return nil, err
  // }
  // headerOrder := []string {
  //   "Host",
  //   "Connection",
  //   "Content-Length",
  //   "Accept",
  //   "User-Agent",
  //   "Content-Type",
  //   "Origin",
  //   "Sec-Fetch-Site",
  //   "Sec-Fetch-Mode",
  //   "Sec-Fetch-Dest",
  //   "Referer",
  //   "Accept-Encoding",
  //   "Accept-Language",
  //   "Cookie",
  // }
  // headers := [][2]string {
  //   {"Connection", "keep-alive"},
  //   {"Accept", "application/json, text/plain, */*"},
  //   {"Content-Type", "application/json"},
  //   {"Origin", "https://www.dickssportinggoods.com"},
  //   {"Sec-Fetch-Site", "same-origin"},
  //   {"Sec-Fetch-Mode", "cors"},
  //   {"Sec-Fetch-Dest", "empty"},
  //   {"Referer", "https://www.dickssportinggoods.com/"},
  //   {"Accept-Encoding", "gzip, deflate, br"},
  //   {"Accept-Language", "en-US,en;q=0.9"},
  // }
  // bodyStructure := map[string]interface{} {
  //   "headers": map[string]interface{} {
  //     "normalizedNames": map[string]interface{} {

  //     },
  //     "lazyUpdate": nil,
  //   },
  //   "withCredentials": true,
  // }
  // body, err := json.Marshal(bodyStructure)
  // if err != nil {
  //   return nil, err
  // }
  // resp, err := t.DicksDoRequest(t.makeReq("POST", url_, &headers, &headerOrder, &body))
  // if err != nil && resp == nil {
  //   return nil, err
  // }


  // rbytes, err := readBodyBytes(resp)
  // if err != nil {
  //   return nil, err
  // }

  // go t.SendTelemetry(map[string]interface{} {
  //   "event": "dsg_start_checkout_resp",
  //   "dsg_start_checkout_resp": string(rbytes),
  // })

  // var storeId string = "15108"
  // var catalogId string = "12301"
  billingUrl, _ := url.Parse("https://www.dickssportinggoods.com/DSGBillingAddressView?calculationUsageId=-1&calculationUsageId=-2&calculationUsageId=-3&calculationUsageId=-7&updatePrices=1&doConfigurationValidation=Y&catalogId=12301&orderId=.&langId=-1&storeId=15108")
  billingResp, err := t.DicksDoRequest(t.makeReq("GET", billingUrl, nil, nil, nil))
  DiscardResp(billingResp)
  if err != nil {
    return nil, err
  }

  t.EnsureValidAbck(t.Config["akUrl"], t.hasValidAbckDicks)

  t.SetCookie(&http.Cookie{ Name: "check", Value: "true" })
  t.SetCookie(&http.Cookie{ Name: "air", Value: "1" })
  t.client.Jar.SetCookies(t.BaseDomainUrl, []*http.Cookie { &http.Cookie { Name: "test", Value: "cookie" }, })
  t.SetCookie(&http.Cookie{ Name: "test", Value: "cookie" })
  t.SetCookie(&http.Cookie{ Name: "yellow", Value: "1" })
  t.SetCookie(&http.Cookie{ Name: "red", Value: "1" })
  t.SetCookie(&http.Cookie{ Name: "blue", Value: "1" })
  t.SetCookie(&http.Cookie{ Name: "peach", Value: "1" })
  t.client.Jar.SetCookies(t.BaseUrl,[]*http.Cookie { &http.Cookie { Name: "test", Value: "me", } })
  // t.SetCookie(&http.Cookie{ Name: "test", Value: "me" })

  // startChkoutUrl, _ := url.Parse(fmt.Sprintf("https://www.dickssportinggoods.com/api/v1/checkouts/%s", "99815666-fa29-49a2-a1a1-d89bda05540b=1357924680"))
  // resp, err := t.DicksDoRequest(t.makeReq("GET", startChkoutUrl, nil, nil, nil))
  // DiscardResp(resp)
  // if err != nil && (resp == nil || resp.StatusCode != 400) {
  //   return nil, err
  // }

  var checkoutSesh DsgCheckoutSession
  startCheckoutUrl, err := url.Parse("https://www.dickssportinggoods.com/api/v1/checkouts")
  headerOrder := []string {
    "Host",
    "Connection",
    "Content-Length",
    "X-INSTANA-T",
    "Pragma",
    "X-INSTANA-L",
    "elastic-apm-traceparent",
    "Content-Type",
    "Access-Control-Allow-Origin",
    "Accept",
    "Cache-Control",
    "User-Agent",
    "X-INSTANA-S",
    "Expires",
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
    {"Content-Length", "0"},
    {"Pragma", "no-cache"},
    {"Content-Type", "application/json"},
    {"Access-Control-Allow-Origin", "*"},
    {"Accept", "application/json"},
    {"Cache-Control", "no-cache,no-store,must-revalidate,max-age=0"},
    {"Expires", "Sat, 01 Jan 2000 00:00:00 GMT"},
    {"Origin", "https://www.dickssportinggoods.com"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", billingUrl.String()},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }
  startCheckoutResp, err := t.DicksDoRequest(t.makeReq("POST", startCheckoutUrl, &headers, &headerOrder, nil))
  if err != nil {
    return nil, err
  }
  err = readRespJsonDst(startCheckoutResp, &checkoutSesh)
  return &checkoutSesh, err

  // checkoutId := getCookieValue("X-DCSG-CHECKOUT", productUrl, t.client.Jar)


  // var checkoutSesh DsgCheckoutSession
  // err = json.Unmarshal(rbytes, &checkoutSesh)
  // if len(checkoutSesh.CheckoutKey) == 0 {
  //   var derr DsgErrorResponse
  //   err = json.Unmarshal(rbytes, &derr)
  //   if derr.UIMessaging.Message != "" {
  //     return nil, errors.New(derr.UIMessaging.Message)
  //   } else {
  //     return nil, ErrInvalidCheckoutSession
  //   }
  // }
  // return &checkoutSesh, err
}

func (t *CheckoutTask) DicksPidLookup(start int, end int) error {
  t.Config["akUrl"] = "https://www.dickssportinggoods.com/assets/d4dd35deui1791386793d54097de02"
  for start < end {
    start += 1
    t.initClient()

    var err error
    t.UserAgent, err = t.getUserAgent()
    if err != nil {
      continue
    }

    // TODO abck for atc?
    // for !hasValidAbckDicks(t.client.Jar) {
    //   t.updateSensorData(t.client, akUrl, productUrl)
    // }

    // t.StatusCh <- Status{Message: "Visiting Product"}
    // productUrl := t.Url

    // prodResp, err := t.DicksDoRequest(t.makeGetReq(productUrl, nil))
    // if err != nil {
    //   return err
    // }
    // prodRespBody, err := readBody(prodResp)
    // if err != nil {
    //   return err
    // }

    // productTitleRe := regexp.MustCompile(`<meta property=og:title content="(.+?)">`)
    // productTitle := productTitleRe.FindStringSubmatch(prodRespBody)[1]
    // storeIdRe := regexp.MustCompile(`storeId=([0-9]+)`)
    // storeId := storeIdRe.FindStringSubmatch(prodRespBody)[1]
    // catalogIdRe := regexp.MustCompile(`catalogId=([0-9]+)`)
    // catalogId := catalogIdRe.FindStringSubmatch(prodRespBody)[1]
    // catEntryIdRe := regexp.MustCompile(`catentryId&q;:([0-9]+)`)
    // catEntryId := catEntryIdRe.FindStringSubmatch(prodRespBody)[1]
    // skusRe := regexp.MustCompile(`skuIdList&q;:.+?}]]`)
    // skusJson := "{\"" + strings.Replace(skusRe.FindString(prodRespBody), "&q;", "\"", -1) + "}"
    // log.Printf("%+v", skusJson)
    // var skus DicksSkus

    // err = json.Unmarshal([]byte(skusJson), &skus)
    // if err != nil {
    //   return err
    // }
    // var selectedSkuIdx int = -1

    // rand.Shuffle(len(skus.Skus), func(i, j int) {
    //   skus.Skus[i], skus.Skus[j] = skus.Skus[j], skus.Skus[i] })

    // if len(t.Sizes) == 0 {
    //   selectedSkuIdx = rand.Intn(len(skus.Skus))
    // } else {
    //   for idx, sku := range skus.Skus {
    //     for _, skuDetail := range sku {
    //       if Any(t.Sizes, func(size string) bool { return skuDetail.Value == size || skuDetail.Value == size + ".0" || skuDetail.Value == "M" + size || strings.HasPrefix(skuDetail.Value, "M" + size + "/") || strings.HasPrefix(skuDetail.Value, "M" + size + ".0") }) {
    //         selectedSkuIdx = idx
    //         break
    //       }
    //     }
    //     if selectedSkuIdx != -1 {
    //       break
    //     }
    //   }
    // }
    // if selectedSkuIdx == -1 {
    //   return retry.Unrecoverable(errors.New("Out of Stock"))
    // } else {
    //   log.Printf("selected sku %+v", skus.Skus[selectedSkuIdx])
    // }

    // var selectedVariant string
    // selectedSize := skus.Skus[selectedSkuIdx][len(skus.Skus[selectedSkuIdx])-1].Value
    // for _, skuDetail := range skus.Skus[selectedSkuIdx] {
    //   if skuDetail.Value == selectedSize {
    //     continue
    //   }
    //   selectedVariant = selectedVariant + skuDetail.Value + " "
    // }

    // t.StatusCh <- Status{Message: "Adding to Cart"}
    // partNum := skus.PartNumbers[selectedSkuIdx]
    // log.Printf("variant %+v", selectedVariant)
    // log.Printf("partnum %+v", start)
    // log.Printf("catEntryId %+v", catEntryId)

    err = t.DsgAtc(fmt.Sprintf("%d", start))
    if err != nil {
      continue
    }

    t.StartCheckoutTime = timeMillis()
    checkoutSesh, err := t.DsgStartCheckout()
    if checkoutSesh != nil {
      log.Printf("%d == %s", start, strings.Join(checkoutSesh.Cart.Items[0].Description, ";"))
    } else if err != nil {
      errstr := fmt.Sprintf("%v", err)
      if strings.Contains(errstr, "a restricted product") {
        log.Printf("%d == %s", start, errstr)
      }
    }
  }
  return nil
}

type DsgRushApiProduct struct {
  ReleaseTime int   `json:"reltime"`
  SkusBySize   map[string]string `json:"skus_by_size"`
}

func (t *CheckoutTask) DicksDoRequest(req *http.Request) (*http.Response, error) {
  var resp *http.Response
  var err error
  err = t.Retry(func() error {
    resp, err = t.doReqTC(req)
    if resp != nil {
      var passedc string
      for _, c := range t.client.Jar.Cookies(req.URL) {
        if strings.Contains(c.Name, "akavpwr") {
          DiscardResp(resp)
          t.StatusCh <- StatusInQueue
          t.GenericQueueBypass()
          // https://www.dickssportinggoods.com/api/v1/carts/contents/21587362?qty=1
          return errInQueue
        }
        if strings.Contains(c.Name, "_WonLottery") {
          passedc = c.Name
        }
      }
      if passedc != "" {
        t.SetCookie(&http.Cookie{ Name: strings.Replace(passedc, "WonLottery", "FailedLottery", -1), Value:"", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/"} )
      }
      if resp.StatusCode == 403 {
        if req.Method != "GET" && t.Config["akUrl"] != "" {
          t.EnsureValidAbck(t.Config["akUrl"], t.hasValidAbckDicks)
          return ErrRetrying
        }
      }
    }

    return retry.Unrecoverable(err)
  }, retry.Attempts(64), retry.Delay(1*time.Second))
  return resp, err
}



func (t *CheckoutTask) DicksCheckout() error {
  if os.Getenv("PID_LOOKUP") == "1" {
    t.DicksPidLookup(21587366, 21600000)
  }
  // t.DicksPidLookup(20740629, 20800000)

  // t.DicksPidLookup(20590253) // sz 14 latest live j34
  // t.DicksPidLookup(20716667) // j12 unigold
  t.StatusTelemetryEnabled = true

  return t.Retry(func() error {
    t.initClient()

    var err error
    t.UserAgent, err = t.getUserAgent()
    if err != nil {
      return err
    }

    homeResp, err := t.DicksDoRequest(t.makeGetReq(t.BaseUrl, nil))
    if err != nil{
      return err
    }
    homeRespBody, err := readBody(homeResp)
    if err != nil {
      return err
    }
    akUrl, err := akUrlFromPageBody(DICKS_BASE, homeRespBody)
    if err != nil {
     return err
    }
    t.Config["akUrl"] = akUrl
    log.Printf("akurl %s", akUrl)

    // t.DicksPidLookup(21543096)

    t.StatusCh <- Status{Message: "Initializing"}
    productUrl := t.Url

    urlSp := strings.Split(t.Url.Path, "/")
    productEcode := urlSp[len(urlSp)-1]


    var partNum string
    var storeId string = "15108"
    var catalogId string = "12301"
    var apiProducts map[string]DsgRushApiProduct
    var releaseAt int64 = 0
    err = t.GetJson("https://rushaio.s3.amazonaws.com/dsgpid.json", &apiProducts)
    if apiProduct, ok := apiProducts[productEcode]; ok {
      releaseAt = int64(apiProduct.ReleaseTime)
      if len(t.Sizes) > 0 {
        rand.Shuffle(len(t.Sizes), func(i, j int) {
            t.Sizes[i], t.Sizes[j] = t.Sizes[j], t.Sizes[i] })
        for skusize, sku := range apiProduct.SkusBySize {
          if Any(t.Sizes, func(size string) bool { return skusize == size || skusize == size + ".0" || skusize == "M" + size || strings.HasPrefix(skusize, "M" + size + "/") || strings.HasPrefix(skusize, "M" + size + ".0") }) {
            partNum = sku
            break
          }
        }
        if partNum == "" {
          return ErrNoSizesAvailable
        }
      } else {
        skus := []string{}
        for _, sku := range apiProduct.SkusBySize {
          skus = append(skus, sku)
        }
        partNum = skus[rand.Intn(len(skus))]
      }
    } else {
      prodResp, err := t.DicksDoRequest(t.makeGetReq(productUrl, nil))
      if err != nil {
        return err
      }
      prodRespBody, err := readBody(prodResp)
      if err != nil {
        return err
      }
      var selectedSize string
      if strings.Contains(prodRespBody, "hasSingleSKU&q;:true") {
        skuRe := regexp.MustCompile(`skuIdList&q;:\[&q;([0-9]+)&q;\]`)
        partNum = skuRe.FindStringSubmatch(prodRespBody)[1]
      } else {
        prodResp, err := t.DicksDoRequest(t.makeGetReq(productUrl, nil))
        if err != nil {
          return err
        }
        prodRespBody, err := readBody(prodResp)
        if err != nil {
          return err
        }
        skusRe := regexp.MustCompile(`skuIdList&q;:.+?}]]`)
        skusJson := "{\"" + strings.Replace(skusRe.FindString(prodRespBody), "&q;", "\"", -1) + "}"
        // log.Printf("%+v", skusJson)
        var skus DicksSkus

        err = json.Unmarshal([]byte(skusJson), &skus)
        if err != nil {
          return err
        }
        var selectedSkuIdx int = -1

        rand.Shuffle(len(skus.Skus), func(i, j int) {
          skus.Skus[i], skus.Skus[j] = skus.Skus[j], skus.Skus[i] })

        if len(t.Sizes) == 0 {
          selectedSkuIdx = rand.Intn(len(skus.Skus))
        } else {
          for idx, sku := range skus.Skus {
            for _, skuDetail := range sku {
              if Any(t.Sizes, func(size string) bool { return skuDetail.Value == size || skuDetail.Value == size + ".0" || skuDetail.Value == "M" + size || strings.HasPrefix(skuDetail.Value, "M" + size + "/") || strings.HasPrefix(skuDetail.Value, "M" + size + ".0") }) {
                selectedSkuIdx = idx
                break
              }
            }
            if selectedSkuIdx != -1 {
              break
            }
          }
        }
        if selectedSkuIdx == -1 {
          return retry.Unrecoverable(errors.New("Out of Stock"))
        } else {
          log.Printf("selected sku %+v", skus.Skus[selectedSkuIdx])
        }

        selectedSize = skus.Skus[selectedSkuIdx][len(skus.Skus[selectedSkuIdx])-1].Value
        for _, skuDetail := range skus.Skus[selectedSkuIdx] {
          if skuDetail.Value == selectedSize {
            continue
          }
          // selectedVariant = selectedVariant + skuDetail.Value + " "
        }
        partNum = skus.PartNumbers[selectedSkuIdx]
      }
    }

    t.StatusCh <- Status{Message: "Adding to Cart"}
    // partNum := "20716675"
    // log.Printf("variant %+v", selectedVariant)
    log.Printf("partnum %+v", partNum)

    t.SetCookie(&http.Cookie{ Name: "air", Value: "1" })
    t.SetCookie(&http.Cookie{ Name: "check", Value: "true" })
    t.client.Jar.SetCookies(t.BaseDomainUrl, []*http.Cookie { &http.Cookie { Name: "test", Value: "cookie" }, })
    t.SetCookie(&http.Cookie{ Name: "test", Value: "cookie" })
    t.SetCookie(&http.Cookie{ Name: "yellow", Value: "1" })
    t.SetCookie(&http.Cookie{ Name: "red", Value: "1" })
    t.SetCookie(&http.Cookie{ Name: "blue", Value: "1" })
    t.SetCookie(&http.Cookie{ Name: "peach", Value: "1" })
    t.client.Jar.SetCookies(t.BaseUrl,[]*http.Cookie { &http.Cookie { Name: "test", Value: "me", } })

    var checkoutSesh *DsgCheckoutSession
    err = t.Retry(func() error {
      t.EnsureValidAbck(t.Config["akUrl"], t.hasValidAbckDicks)
      err := t.DsgAtc(partNum)
      if err != nil {
        return err
      }

      t.EnsureValidAbck(t.Config["akUrl"], t.hasValidAbckDicks)

      err = t.DsgHandshake("8bfadb12-ec53-4adf-99b1-f9914c48f67a", "1357924680")
      if err != nil {
        return err
      }

      checkoutSesh, err = t.DsgStartCheckout()
      log.Printf("%+v", err)

      return err
    }, retry.Delay(500*time.Millisecond))
    if err != nil {
      return err
    }
    t.StartCheckoutTime = timeMillis()

    // TODO visit billingaddressview "youre headed to checkout"
    // https://www.dickssportinggoods.com/DSGBillingAddressView

    t.EnsureValidAbck(t.Config["akUrl"], t.hasValidAbckDicks)

    t.StatusCh <- Status{Message: "Verifying Addresses"}
    verifiedShipAddress, err := t.dicksVerifyAddress(checkoutSesh.CheckoutKey, t.Profile.ShippingAddress)
    if err != nil {
      return err
    }
    verifiedBillAddress, err := t.dicksVerifyAddress(checkoutSesh.CheckoutKey, t.Profile.BillingAddress)
    if err != nil {
      return err
    }



    apiCheckoutUrlStr := fmt.Sprintf("https://www.dickssportinggoods.com/api/v1/checkouts/%s", checkoutSesh.CheckoutKey)
    setBillAddressUrl, err := url.Parse(fmt.Sprintf("%s/addresses?type=billing", apiCheckoutUrlStr))
    if err != nil {
      return err
    }
    setAddressHeaders := [][2]string {
      {"accept", "application/json"},
      {"content-type", "application/json"},
      {"origin", "https://www.dickssportinggoods.com"},
      {"sec-fetch-dest", "empty"},
      {"sec-fetch-mode", "cors"},
      {"sec-fetch-site", "same-origin"},
      {"referer", fmt.Sprintf("https://www.dickssportinggoods.com/DSGBillingAddressView")},
    }
    /*
    AddressLines []string `json:"addressLines"`
    City string `json:"city"`
    Country string `json:"country"`
    State string `json:"state"`
    ZipCode string `json:"zipCode"`
    */
    setBillAddressBodyMap := map[string]string {
        "address": strings.Join(verifiedBillAddress.AddressLines, " "),
        "city": verifiedBillAddress.City,
        "country": verifiedBillAddress.Country,
        "email": t.Profile.Email,
        "first_name": t.Profile.BillingAddress.FirstName,
        "last_name": t.Profile.BillingAddress.LastName,
        "phone": t.Profile.BillingAddress.DicksPhone(),
        "state": verifiedBillAddress.State,
        "zipcode": verifiedBillAddress.ZipCode,
    }
    setBillAddressBody, err := json.Marshal(setBillAddressBodyMap)
    if err != nil {
      return err
    }
    t.StatusCh <- Status{Message: "Sending Billing Info"}
    _, err = t.DicksDoRequest(t.makeReq("PUT", setBillAddressUrl, &setAddressHeaders, nil, &setBillAddressBody))
    if err != nil {
      return err
    }

    setShipAddressUrl, err := url.Parse(fmt.Sprintf("%s/addresses?type=shipping", apiCheckoutUrlStr))
    if err != nil {
      return err
    }
    // setAddressHeaders := [][2]string {
    //  {"accept", "application/json"},
    //  {"content-type", "application/json"},
    //  {"origin", "https://www.dickssportinggoods.com"},
    // }
    /*
    AddressLines []string `json:"addressLines"`
    City string `json:"city"`
    Country string `json:"country"`
    State string `json:"state"`
    ZipCode string `json:"zipCode"`
    */
    setShipAddressBodyMap := map[string]string {
        "address": strings.Join(verifiedShipAddress.AddressLines, " "),
        "city": verifiedShipAddress.City,
        "country": verifiedShipAddress.Country,
        "email": t.Profile.Email,
        "first_name": t.Profile.ShippingAddress.FirstName,
        "last_name": t.Profile.ShippingAddress.LastName,
        "phone": t.Profile.ShippingAddress.DicksPhone(),
        "state": verifiedShipAddress.State,
        "zipcode": verifiedShipAddress.ZipCode,
    }
    setShipAddressBody, err := json.Marshal(setShipAddressBodyMap)
    if err != nil {
      return err
    }
    t.StatusCh <- Status{Message: "Sending Shipping Info"}
    _, err = t.DicksDoRequest(t.makeReq("PUT", setShipAddressUrl, &setAddressHeaders, nil, &setShipAddressBody))
    if err != nil {
      return err
    }

    finishAddressesUrl, err := url.Parse(fmt.Sprintf("%s/addresses", apiCheckoutUrlStr))
    if err != nil {
      return err
    }
    finishAddressesResp, err := t.DicksDoRequest(t.makeReq("POST", finishAddressesUrl, &setAddressHeaders, nil, nil))
    if err != nil {
      return err
    }
    finishAddressesResp.Body.Close()


    paymentPageUrl, _ := url.Parse(fmt.Sprintf("https://www.dickssportinggoods.com/DSGPaymentViewCmd?catalogId=%slangId=-1&storeId=%s", catalogId, storeId))
    paymentRespp, err := t.DicksDoRequest(t.makeReq("GET", paymentPageUrl, nil, nil, nil))
    DiscardResp(paymentRespp)
    if err != nil {
      return err
    }

    payKeyHeaders := [][2]string {
      {"accept", "application/json"},
      {"content-type", "application/json"},
      {"origin", "https://www.dickssportinggoods.com"},
      {"sec-fetch-dest", "empty"},
      {"sec-fetch-mode", "cors"},
      {"sec-fetch-site", "same-origin"},
      {"referer", paymentPageUrl.String()},
    }
    payKeyUrl, err := url.Parse(fmt.Sprintf("%s/payment/payment-processor/cybersource/public-key", apiCheckoutUrlStr))
    if err != nil {
      return err
    }
    t.StatusCh <- Status{Message: "Encrypting Card (1)"}
    payKeyResp, err := t.DicksDoRequest(t.makeReq("GET", payKeyUrl, &payKeyHeaders, nil, nil))
    payKeyBytes, err := readBodyBytes(payKeyResp)
    if err != nil {
      return err
    }
    var payKeys CSPayKeys
    err = json.Unmarshal(payKeyBytes, &payKeys)
    if err != nil {
      return err
    }
    cardNumberEnc, err := payKeys.EncryptCard(t.Profile.Card.Number)
    if err != nil {
      return err
    }
    csTokenUrl, err := url.Parse("https://flex.cybersource.com/cybersource/flex/v1/tokens")
    if err != nil {
      return err
    }
    csTokenHeaders := [][2]string {
      {"accept", "application/json"},
      {"origin", "https://www.dickssportinggoods.com"},
      {"content-type", "application/json"},
      {"sec-fetch-dest", "empty"},
      {"sec-fetch-mode", "cors"},
      {"sec-fetch-site", "same-origin"},
      {"referer", "https://flex.cybersource.com/"},
    }
    csTokenBodyMap := map[string]interface{} {
        "cardInfo": map[string]string {
            "cardExpirationMonth": fmt.Sprintf("%02d", t.Profile.Card.ExpMonth),
            "cardExpirationYear": fmt.Sprintf("%d", t.Profile.Card.ExpYear),
            "cardNumber": cardNumberEnc,
            "cardType": CSCardTypeNumberFromName(t.Profile.Card.Type),
        },
        "keyId": payKeys.DicksKeyId,
    }
    csTokenBody, err := json.Marshal(csTokenBodyMap)
    if err != nil {
      return err
    }
    t.StatusCh <- Status{Message: "Encrypting Card (2)"}
    csClient, err := t.newHttpClient()
    if err != nil {
      return err
    }
    csTokenResp, err := t.doReq(csClient, t.makeReq("POST", csTokenUrl, &csTokenHeaders, nil, &csTokenBody))
    if err != nil {
      return err
    }
    csTokenBytes, err := readBodyBytes(csTokenResp)
    if err != nil {
      return err
    }
    var csToken map[string]interface{}
    err = json.Unmarshal(csTokenBytes, &csToken)
    if err != nil {
      return err
    }

    t.StartPaymentTime = timeMillis()

    // PUT
    paymentUrl, err := url.Parse(fmt.Sprintf("%s/payment/payment-processor", apiCheckoutUrlStr))
    if err != nil {
      return err
    }
    paymentHeaders := [][2]string {
      {"accept", "application/json"},
      {"content-type", "application/json"},
      {"sec-fetch-dest", "empty"},
      {"sec-fetch-mode", "cors"},
      {"sec-fetch-site", "same-origin"},
      {"origin", "https://www.dickssportinggoods.com"},
      {"referer", fmt.Sprintf("https://www.dickssportinggoods.com/DSGPaymentViewCmd?catalogId=%slangId=-1&storeId=%s", catalogId, storeId)},
    }
    paymentBodyMap := map[string]interface{} {
      "accountDisplay": nil,
      "cardType": CSCardTypeNumberFromName(t.Profile.Card.Type),
      "cvv": t.Profile.Card.Cvc,
      "decisionParameter": "", // TODO real
      "expireMonth": fmt.Sprintf("%02d", t.Profile.Card.ExpMonth),
      "expireYear": fmt.Sprintf("%d", t.Profile.Card.ExpYear),
      "flexPublicKey": payKeys,
      "token": csToken,
    }

    paymentReqBody, err := json.Marshal(paymentBodyMap)
    if err != nil {
      return err
    }
    t.StatusCh <- Status{Message: "Sending Encrypted Card"}
    paymentResp, err := t.DicksDoRequest(t.makeReq("PUT", paymentUrl, &paymentHeaders, nil, &paymentReqBody))
    if err != nil {
      return err
    }
    paymentResp.Body.Close()


    checkoutUrl, err := url.Parse(apiCheckoutUrlStr)
    if err != nil {
      return err
    }
    checkoutHeaders := [][2]string {
      {"accept", "application/json"},
      {"content-type", "application/json"},
      {"sec-fetch-dest", "empty"},
      {"sec-fetch-mode", "cors"},
      {"sec-fetch-site", "same-origin"},
      {"origin", "https://www.dickssportinggoods.com"},
      {"referer", fmt.Sprintf("https://www.dickssportinggoods.com/DSGPaymentViewCmd?catalogId=%slangId=-1&storeId=%s", catalogId, storeId)},
    }
    t.StatusCh <- Status{Message: "Finalizing Order"}

    if os.Getenv("DSG_NOWAIT") != "1" && releaseAt > 0 {
      t.StatusCh <- Status{Message: "Waiting to Order"}
      // time.Sleep(600 * time.Second)
      waitsecs := releaseAt - time.Now().Unix()
      if waitsecs > 0 {
        log.Printf("waiting %d secs till release", waitsecs)
        time.Sleep(time.Duration(waitsecs) * time.Second)
      }
    }

    return t.Retry(func() error {
      checkoutResp, err := t.DicksDoRequest(t.makeReq("POST", checkoutUrl, &checkoutHeaders, nil, nil))
      if err != nil {
        bb, _ := readBodyBytes(checkoutResp)
        if len(bb)>0 {
          var bbm map[string]interface{}
          json.Unmarshal(bb, &bbm)

          if checkoutResp != nil && checkoutResp.StatusCode == 402 {
            t.StatusCh <- Status{Message: fmt.Sprintf("Payment failed: %v", bbm["message"])}
            go t.DeclineWebhook("DICK'S", strings.Join(checkoutSesh.Cart.Items[0].Description, "; "), "")
            return TaskErrUnrecoverable(errCardDeclined)
          }
        }
        return err
      }
      checkoutResp.Body.Close()
      t.StatusCh <- Status{Message: "COOKED!!!", Level: 9999}
      go t.SuccessWebhook("DICK'S", strings.Join(checkoutSesh.Cart.Items[0].Description, "; "), "")

      return nil
    })
  }, retry.Attempts(16), retry.LastErrorOnly(true), retry.DelayType(retry.FixedDelay), retry.Delay(1*time.Second), retry.OnRetry(func(_ uint, err error) {
    log.Printf("%+v", err)
  }))
}

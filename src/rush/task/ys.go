package task

import (
  "strconv"
  "hash/crc32"
  "sync"
  "encoding/json"
  "fmt"
  "github.com/avast/retry-go"
  "github.com/getsentry/sentry-go"
  "github.com/pkg/errors"
  "rush/net/http/cookiejar"
  "log"
  "math/rand"
  "net/url"
  "os"
  "rush/common"
  "rush/net/http"
  "strings"
  "bytes"
  // "context"
  "io/ioutil"
  "time"
)

var YsGetBasketsUrl, _ = url.Parse(YS_BASE + "api/checkout/customer/baskets")
func (t *CheckoutTask) YsGetCart() (*Cart, error) {
  headers := [][2]string {
    {"referer", t.Url.String()},
    {"checkout-authorization", t.Config["checkout-authorization"]},
    {"content-type", "application/json"},
    {"accept", "*/*"},
    {"sec-fetch-mode", "cors"},
  }
  resp, err := t.doReq(t.client, t.makeGetReq(YsGetBasketsUrl, &headers))
  defer DiscardResp(resp)
  if err != nil {
    DiscardResp(resp)
    return nil, err
  }
  if resp.Header.Get("authorization") != "" {
    t.Config["checkout-authorization"] = resp.Header.Get("authorization")
  } else {
    t.Config["checkout-authorization"] = "null"
  }
  var cart Cart
  err = readRespJsonDst(resp, &cart)
  return &cart, err
}

func (t *CheckoutTask) YsGetU(u string) {
  headers := make([][2]string, 0)
  headers = addHeader(headers, "sec-fetch-site", "none")
  headers = addHeader(headers, "sec-fetch-mode", "navigate")
  headers = addHeader(headers, "sec-fetch-user", "?1")
  headers = addHeader(headers, "upgrade-insecure-requests", "1")
  if urll, _ := url.Parse(u); urll != nil {
    homeResp, _ := t.doReq(t.client, t.makeGetReq(urll,&headers))
    defer DiscardResp(homeResp)
  }
}
func (t *CheckoutTask) YsGetXhrU(u string) (error) {
  is := randHex(8)
  headers := [][2]string {
    {"referer", t.Url.String()},
    {"checkout-authorization", t.Config["checkout-authorization"]},
    {"content-type", "application/json"},
    {"accept", "*/*"},
    {"origin", "https://www.yeezysupply.com"},
    {"sec-fetch-mode", "cors"},
    {"x-instana-t", is},
    {"x-instana-s", is},
    {"x-instana-l", "1,correlationType=web;correlationId=" + is},
  }
  if urll, _ := url.Parse(u); urll != nil {
    resp, err := t.doReq(t.client, t.makeGetReq(urll, &headers))
    defer DiscardResp(resp)
    if err != nil {
      DiscardResp(resp)
      return err
    }
    if resp.Header.Get("authorization") != "" {
      t.Config["checkout-authorization"] = resp.Header.Get("authorization")
    } else {
      t.Config["checkout-authorization"] = "null"
    }
    DiscardResp(resp)
    return err
  }
  return ErrRetrying
}

func (t *CheckoutTask) YsGenCookie(akUrl string) error {
  var tries = 0
  gotChal := false

  defer func() {
    t.Config["stage"] = ""
  }()
  if t.Config["stage"] != "" {
    if aji, err := strconv.Atoi(t.Config["aj_indx"]); err == nil {
      t.Config["aj_indx"] = fmt.Sprintf("%d", aji+1)
    }
    t.updateSensorData(t.client,t.ClientCtx,t.UserAgent,akUrl,t.Url)
    if strings.Contains(t.GetCookieValue("_abck"), "||1") {
      t.updateSensorData(t.client,t.ClientCtx,t.UserAgent,akUrl,t.Url)
      tries += 1
    }
    return nil
  }
  t.Config["aj_indx"] = "0"
  t.Config["stage"] = "onload"
  t.updateSensorData(t.client,t.ClientCtx,t.UserAgent,akUrl,t.Url)
  t.Config["aj_indx"] = "1"
  t.Config["stage"] = "fp"
  t.updateSensorData(t.client,t.ClientCtx,t.UserAgent,akUrl,t.Url)

  // t.Config["aj_indx"] = "2"
  // t.updateSensorData(t.client,t.ClientCtx,t.UserAgent,akUrl,t.Url)
  // for strings.Contains(t.GetCookieValue("_abck"), "||1") && tries < 6 {
  //   t.Config["stage"] = "chalsolve"
  //   tries += 1
  // }
  return nil

  // return nil

  for tries < 12 {
    body, err := t.updateSensorData(t.client,t.ClientCtx,t.UserAgent,akUrl,t.Url)
    if err != nil && !t.shouldRetry(err) {
      return err
    }
    if strings.Contains(body, "true") && err == nil && ((gotChal && !strings.Contains(t.GetCookieValue("_abck"), "||")) || tries > 3) {
      return nil
    }
    t.StatusCh <- StatusGeneratingCookie // , "tries", tries)
    tries = tries + 1
    if strings.Contains(t.GetCookieValue("_abck"), "||") {
      gotChal = true
    } else {
      time.Sleep(time.Duration(int(1500 * rand.Float32())) * time.Millisecond)
    }

  }
  return ErrFailedToGenCookie
}

func (t *CheckoutTask) YsCheckout(productUrl *url.URL, profile Profile, taskConfig TaskConfig) error {

  productStyleCode := productUrl.Path[strings.LastIndex(productUrl.Path, "/")+1:]
  var sizeChoice string
  // var err error
  return t.RetryTask(func() error {
    t.Config["checkout-authorization"] = "null"
    headers := make([][2]string, 0)

    // t.StatusCh <- Status{Message: "Rotating Fingerprint"}
    // if os.Getenv("FF_FP") == "1" {
    //   // t.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:83.0) Gecko/20100101 Firefox/8.0"
    //   os.Setenv("FF_FP", "1")
    // } else {
    // }
    var err error
    t.UserAgent, err = t.getUserAgent()
    if err != nil {
      return err
    }
    t.SetTelemetryContext("ua", t.UserAgent)

    err = t.initClient()
    if err != nil {
      return err
    }


    // resp, err := t.ApiPost("/sesh/get", t.GetBareSessionBytes())
    // if err == nil {
    //   var s TaskSession
    //   if err = json.Unmarshal(resp, &s); err == nil {
    //     t.ApplySession(&s)
    //     t.LogDebug("using sesh: %+v", s)
    //   }
    // }

    if strings.Contains(t.UserAgent, "Firefox") {
      t.HeaderOrder = []string{
        "user-agent",
        "accept",
        "accept-encoding",
        "accept-language",
      }
    }

    t.LogDebug("ua: %s", t.UserAgent)

    t.DelTelemetryContext("ys_visithome")
    if rand.Float32() < 0.05 {
      t.SetTelemetryContext("ys_visithome", "1")
      t.Config["genurl"] = t.BaseUrl.String()
      t.StatusCh <- Status{Message: "Visit Homepage", Level: 0}
      headers = addHeader(headers, "sec-fetch-site", "none")
      headers = addHeader(headers, "sec-fetch-mode", "navigate")
      headers = addHeader(headers, "sec-fetch-user", "?1")
      headers = addHeader(headers, "sec-fetch-dest", "document")
      headers = addHeader(headers, "upgrade-insecure-requests", "1")
      homeUrl := t.BaseUrl
      homeUrl.Path = ""
      homeUrl.RawQuery = ""
      homeResp, err := t.doReq(t.client, t.makeGetReq(homeUrl,&headers))
      defer DiscardResp(homeResp)
      if err != nil {

        errf := fmt.Sprintf("Failed to fetch homepage: %+v", err)
        t.StatusCh <- Status{Message: errf}
        return err
      }
      var akUrl string
      if homeBody, err := readBody(homeResp); err == nil {
        akUrl, err = akUrlFromPageBody(YS_BASE, homeBody)
        if aup, err := url.Parse(akUrl); err == nil {
          headers = [][2]string {
            {"accept", "*/*"},
            {"accept-language", "en-US,en;q=0.5"},
            {"accept-encoding", "gzip, deflate, br"},
            {"referer", t.BaseUrl.String()},
          }
          resp, _ := t.doReq(t.client, t.makeGetReq(aup, &headers))
          DiscardResp(resp)
          t.Config["akUrl"] = akUrl
          go func() {
            time.Sleep(time.Duration(rand.Intn(1000))*time.Millisecond)
            defer func() {
              recover()
            }()
            t.YsGetXhrU("https://www.yeezysupply.com/api/yeezysupply/products/bloom")
          }()
          t.YsGenCookie(akUrl)
        }
      }

      time.Sleep(1*time.Second)
    }
    // newRecapCookie(taskConfig, YS_BASE + productEndpoint, t.client)
    /*
    body, err := t._PageBody(client, taskConfig, productUrl)
    if err != nil {
      return "", err
    }
    return akUrlFromPageBody(YS_BASE, body)
    */


    // akUrl, err := t.waitForProductAk(taskConfig, productUrl, t.client)
    // if err != nil {
    //   t.StatusCh <- Status{Message: fmt.Sprintf("%v", err)}
    //   return err
    // }

    // tries := 1
    // url, _ := url.Parse(YS_BASE + "/")
    // t.client.Jar.SetCookies(url, []*http.Cookie{ &http.Cookie{ Name: "_abck", Value: "", MaxAge: -1, Expires: time.Now().Add(-100 * time.Hour), Secure: true, Path: "/"}})
    // t.LogDebug("cookie %s", getCookieValue("_abck", t.client.Jar))
    // akJar,_ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

    // if err != nil {
    //  log.Fatal(err)
    // }

    // t.client.CloseIdleConnections()

    // t.client := t.client
    // newRecapCookieSync(taskConfig, YS_BASE + productEndpoint, t.client)
    // t.client.Jar.SetCookies(url, []*http.Cookie{ &http.Cookie{ Name: "PH0ENIX", Value: "false", }})
    failuresByStage := map[string]uint{}

    n403 := 0
    retatc := 0
    return t.Retry(func() error {
      defer CloseH2Conns(t.client)
      t.Config["genurl"] = productUrl.String()
      defer func() {
        t.resetClient(false)
      }()
      if n403 > 12 {
        return retry.Unrecoverable(ErrRetrying)
      }
      defer func() {
        retatc += 1
      }()
      if os.Getenv("HALT") == "1" && retatc > 0{
        t.FtlSendSessionSnapshot()
        time.Sleep(3*time.Second)
        return nil
      }
      body, err := t.getProductPageBody(t.client, taskConfig, productUrl)
      if err != nil {
        return err
      }
      akUrl, err := akUrlFromPageBody(YS_BASE, body)
      if err != nil {
        akUrl = t.ApiConfig.SensorEndpoints[t.Url.Host]
      }

      t.StatusCh <- Status{Message: "Waiting for Stock", Level: 5}
      // t.StatusCh <- StatusGeneratingCookie
      // t.LogDebug(akUrl)
      // t.LogDebug(t.GetCookieValue("_abck"))
      if aup, err := url.Parse(akUrl); err == nil {
        headers = [][2]string {
          {"accept", "*/*"},
          {"accept-language", "en-US,en;q=0.5"},
          {"accept-encoding", "gzip, deflate, br"},
          {"referer", productUrl.String()},
        }
        resp, _ := t.doReq(t.client, t.makeGetReq(aup, &headers))
        DiscardResp(resp)
      }
      t.Config["stage"] = ""
      go func() {
        defer func(){
          recover()
        }()
        time.Sleep(time.Duration(rand.Intn(500))*time.Millisecond)
        t.YsGetApiProduct()
      }()
      t.YsGenCookie(akUrl)

      // if err != nil && !t.shouldRetry(err) {
      //   return err
      // }

      t.EnableHttpTrace = rand.Float32() < 0.01



      // if rand.Float32() < 0.5 {
      //   t.SetTelemetryContext("ys_get_cart", "1")
      //   t.YsGetCart()
      // }


      var cartAddResp *http.Response
      var cart *Cart
      var cartAdd CartAdd
      // common.Any([]string{}, func(s string) bool { return true })
      var availability Availability
      is := randHex(8)
      skuStr := ""
      av := "https://www.yeezysupply.com/api/products/%s/availability"
      // if retatc > 0 && rand.Float32() < 0.75 {
      //   av = "https://www.yeezysupply.com/hpl/content/availability-v2/yeezy-supply/US/%s.json"
      // }
      availUrl, err  := url.Parse(fmt.Sprintf(av, productStyleCode))
      if err != nil {
        return retry.Unrecoverable(retry.Unrecoverable(errors.New(fmt.Sprintf("Product unavailable: %+v", err))))
      }
      headers = [][2]string {
        {"accept", "*/*"},
        {"accept-language", "en-US,en;q=0.5"},
        {"accept-encoding", "gzip, deflate, br"},
        {"referer", productUrl.String()},
        {"content-type", "application/json"},
        {"x-instana-t", is},
        {"x-instana-s", is},
        {"x-instana-l", "1,correlationType=web;correlationId=" + is},
      }

      availResp, err := t.doReq(t.client, t.makeReq("GET", availUrl, &headers, nil, nil))
      defer DiscardResp(availResp)
      if err != nil && os.Getenv("YSATC") != "1"{
        t.StatusCh <- Status{Message: "Product unavailable, retrying"}
        return errors.New(fmt.Sprintf("Product unavailable: %+v", err))
      }
      if os.Getenv("YSATC") == "1" {
        DiscardResp(availResp)
        sizeChoice = "LOL"
        skuStr = "DONGS"
      } else {
        availabilityBytes, err := readBodyBytes(availResp)
        if err != nil {
          return err
        }
        if (strings.Contains(av, "availability-v2")) {
          var a2 YsAvailabilityV2
          if err := json.Unmarshal(availabilityBytes, &a2); err == nil {
            for skuId, sku := range a2.Skus  {
              /*
              type SizeAvailability struct {
                Sku string `json:"sku"`
                Stock int `json:"availability"`
                Status string `json:"availability_status"`
                Size string `json:"size"`
              }
              */
              hasstock := sku.HypeAvailability > 0
              var status string = "IN_STOCK"
              if !hasstock {
                status = "OUT_OF_STOCK"
              }
              availability.Sizes = append(availability.Sizes, SizeAvailability{
                Sku: skuId,
                Stock: 1,
                Status: status,
                Size: sku.DisplaySize,
              })
            }
          }
        } else {
          json.Unmarshal(availabilityBytes, &availability)
        }
        t.LogDebug("AVAILABILITY %s %v", string(availabilityBytes), availability)
        if len(availability.Sizes) == 0 {
          return errors.New("Product unavailable: out of stock")
        }
      }


      // t.EnableHttpTrace = true

      rand.Shuffle(len(availability.Sizes), func(i, j int) {
        availability.Sizes[i], availability.Sizes[j] = availability.Sizes[j], availability.Sizes[i] })

      for _, availSize := range availability.Sizes {
        log.Printf("%+v %+v", t.Sizes, availSize)
        if availSize.Status == "IN_STOCK" && (len(t.Sizes) == 0 || common.Any(t.Sizes, func(size string) bool { return availSize.Size == size })) {
          skuStr = availSize.Sku
          sizeChoice = availSize.Size
          break
        }
      }


      if sizeChoice == "" || skuStr == "" {
        // no sizes in stock
        t.StatusCh <- StatusOutOfStock
        if os.Getenv("YSATC") == "1" {
          sizeChoice = "LOL"
          skuStr = "DONGS"
        } else {
          return ErrOutOfStock
        }
      }


      // t.SetCookie(&http.Cookie{Name: "NujTerVZWUM68Dv", Value: "exp=1603548365~acl=%2f*~data=46384636444343323436383233413931454245324133363642434533374134423037394432464336324430324142364137444345304530313446384438424136~hmac=6796bb603e3b49e9dd9adb69f9b2f74ecb363fd6b5de1020e6080bac5e62a0f0"})
      //  t.StatusCh <- StatusGeneratingCookie
      // t.LogDebug(akUrl)
      // t.LogDebug(t.GetCookieValue("_abck"))
      // err = t.YsGenCookie(akUrl)
      // if err != nil && !t.shouldRetry(err) {
      //   return err
      // }
      // t.SetCookie(&http.Cookie{Name: "xhwUqgFqfW88H50_u" , Value: "exp=1603548365~acl=%2f*~data=2~hmac=e57e5608f094777734d5dec33685d90f03d170de2b2e6c555b7978ea31b0bb37"})
      // t.SetCookie(&http.Cookie{Name: "akavpau_ys_us", Value: "1603548365~id=45c7094af7334151ed45ae32f644fdef"})
      // todo can we gen cookie earlier / attempt pregen?
      // for {
      //  t.StatusCh <- StatusGeneratingCookie
      // }

      t.StatusCh <- Status{Message: "Adding to cart", Level: 6}
      is = randHex(8)
       headerOrder := []string {
        "user-agent",
        "accept",
        "accept-language",
        "accept-encoding",
        "referer",
        "content-type",
        "checkout-authorization",
        "x-instana-t",
        "x-instana-s",
        "x-instana-l",
        "origin",
        "content-length",
        "cookie",
        "te",
      }

      headers := [][2]string {
        {"accept", "*/*"},
        {"accept-language", "en-US,en;q=0.5"},
        {"accept-encoding", "gzip, deflate, br"},
        {"referer", productUrl.String()},
        {"content-type", "application/json"},
        {"checkout-authorization", "null"},
        {"x-instana-t", is},
        {"x-instana-s", is},
        {"x-instana-l", "1,correlationType=web;correlationId=" + is},
        {"origin", "https://www.yeezysupply.com"},
        // {"te", "trailers"},
      }
      cartAdd = CartAdd {
        Product_id: productStyleCode,
        Product_variation_sku: skuStr,
        ProductId:  skuStr,
        Quantity: 1,
        Size: sizeChoice,
        DisplaySize: sizeChoice,
      }
      cartAddJson, _ := json.Marshal([]CartAdd { cartAdd, })
      cartAddUrl, err := url.Parse(fmt.Sprintf("%sapi/checkout/baskets/-/items", YS_BASE))
      if err != nil {
        return err
      }
      // t.SetCookie(&http.Cookie{Name: "UserSignUpAndSaveOverlay", Value: "1"})
      // t.SetCookie(&http.Cookie{Name: "_fbp", Value: "fb.1.1607671953162.1511214515"})
      // t.SetCookie(&http.Cookie{Name: "s_pers", Value: "%20s_vnum%3D1609488000570%2526vn%253D2%7C1609488000570%3B%20s_invisit%3Dtrue%7C1607698868802%3B"})
      // t.SetCookie(&http.Cookie{Name: "utag_main", Value: "v_id:017650b735730010361cca1d105601077001c06f00c48$_sn:2$_se:35$_ss:0$_st:1607698868799$ses_id:1607696921787%3Bexp-session$_pn:5%3Bexp-session$_prevpage:CHECKOUT%7CPAYMENT%3Bexp-1607700648315"})
      // // t.SetCookie(&http.Cookie{Name: "RT", Value: "\"z", Value: "1&dm", Value: "yeezysupply.com&si", Value: "70832ca4-9d7a-4418-a990-b25b06257d14&ss", Value: "kikd8nkd&sl", Value: "7&tt", Value: "1438&bcn", Value: "%2F%2F173e2509.akstat.io%2F&ld", Value: "2vg3\""})
      // t.SetCookie(&http.Cookie{Name: "restoreBasketUrl", Value: "%2Fon%2Fdemandware.store%2FSites-ys-US-Site%2Fen_US%2FCart-UpdateItems%3Fpid_0%3DFZ4994_660%26qty_0%3D1%26"})
      // t.SetCookie(&http.Cookie{Name: "persistentBasketCount", Value: "1"})
      // t.SetCookie(&http.Cookie{Name: "userBasketCount", Value: "1"})
      // t.SetCookie(&http.Cookie{Name: "_ga", Value: "GA1.2.1728050020.1607671955"})
      // t.SetCookie(&http.Cookie{Name: "_gcl_au", Value: "1.1.509916260.1607671953"})
      // t.SetCookie(&http.Cookie{Name: "_gid", Value: "GA1.2.1108058676.1607671955"})
      // t.SetCookie(&http.Cookie{Name: "s_cc", Value: "true"})
      // t.SetCookie(&http.Cookie{Name: "UserSignUpAndSave", Value: "5"})
      // t.SetCookie(&http.Cookie{Name: "default_searchTerms_CustomizeSearch", Value: "%5B%5D"})
      // t.SetCookie(&http.Cookie{Name: "geoRedirectionAlreadySuggested", Value: "false"})
      // t.SetCookie(&http.Cookie{Name: "wishlist", Value: "%5B%5D"})
      // t.SetCookie(&http.Cookie{Name: "pagecontext_cookies", Value: ""})
      // t.SetCookie(&http.Cookie{Name: "pagecontext_secure_cookies", Value: ""})
      // t.SetCookie(&http.Cookie{Name: "AMCVS_7ADA401053CCF9130A490D4C%40AdobeOrg", Value: "1"})
      // t.SetCookie(&http.Cookie{Name: "AMCV_7ADA401053CCF9130A490D4C%40AdobeOrg", Value: "-227196251%7CMCIDTS%7C18608%7CMCMID%7C19619260385922531791834449715236847933%7CMCAAMLH-1608276752%7C9%7CMCAAMB-1608301722%7C6G1ynYcLPuiQxYZrsz_pkqfLG9yMXBpb2zX5dvJdYQJzPXImdj0y%7CMCOPTOUT-1607704122s%7CNONE%7CMCAID%7CNONE"})
      // t.SetCookie(&http.Cookie{Name: "PH0ENIX", Value: "false"})
      // t.SetCookie(&http.Cookie{Name: "AKA_A2", Value: "A"})
      // t.SetCookie(&http.Cookie{Name: "geo_country", Value: "US"})
      t.LogDebug("cart abck %s", t.GetCookieValue("_abck"))
      var authorization string

        time.Sleep(time.Duration(rand.Intn(3000))*time.Millisecond)
        t.Config["stage"] = "click"
        t.YsGenCookie(akUrl)

        if retatc > 0 && rand.Float32() < 0.5 {
          if ccart, err := t.YsGetCart(); err == nil {
            cart = ccart
          }
        }
        if cart == nil || (cart.TotalProductCount <1  && len(cart.ShipmentList) < 1){
          // t.Config["stage"] = "click"
          // t.YsGenCookie(akUrl)


          if err != nil && !t.shouldRetry(err) {
            return err
          }
          cartAddResp, err = t.doReq(t.client, t.makeReq("POST", cartAddUrl, &headers, &headerOrder, &cartAddJson))
          defer DiscardResp(cartAddResp)
          var cartJson []byte
          var cartRespStatus int
          if cartAddResp != nil {
            if auth := cartAddResp.Header.Get("authorization"); auth != "" {
              authorization = auth
              t.Config["checkout-authorization"] = auth
            }
            var berr error
            cartJson, berr = readBodyBytes(cartAddResp)
            if berr != nil {
              return berr
            }
            cartRespStatus = cartAddResp.StatusCode
          }
          t.LogDebug("%s", string(cartJson[:]))

          t.LogDebug("cart_resp_ua %s %d", t.UserAgent, cartRespStatus)



          go t.SendTelemetry(map[string]interface{} {
            "event": "ys_cart_response",
            "status_code": cartRespStatus,
            "ys_cart_response": string(cartJson[:]),
          })
          if cartRespStatus == 403 {
            n403 += 1
            t.StatusCh <- Status{Message: "Add to Cart Forbidden - Retrying"}
            return retry.Unrecoverable(ErrRetrying)
          }
          if err != nil {
            if serr, ok := err.(*StatusCodeRejectedError); ok {
              if serr.StatusCode == 403 {
              }
            }
            t.StatusCh <- Status{Message: fmt.Sprintf("Add to cart failure: %+v %s", err, string(cartJson[:]))}
            return err
          }


          err = json.Unmarshal(cartJson, &cart)
          if err != nil {

            t.StatusCh <- Status{Message: fmt.Sprintf("Add to cart failure: %+v", err)}
            return err
          }
          if len(cart.MessageList)>0 {
            if cart.MessageList[len(cart.MessageList)-1].Type == "InvalidProductItemException" {
              t.StatusCh <- Status{Message: fmt.Sprintf("Add to cart error: Product no longer available (Unrecoverable")}
              return errors.New("Product no longer available")
            } else if cart.MessageList[len(cart.MessageList)-1].Type == "ProductsForbidden" {
              t.StatusCh <- Status{Message: fmt.Sprintf("Add to cart error: Product not yet available. WR is probably not online. Start task again")}
              return TaskErrUnrecoverable(errors.New("Product not yet available"))
            } else if cart.MessageList[len(cart.MessageList)-1].Type == "ProductItemNotAvailableException"  {
              msgg := fmt.Sprintf("Add to cart error: Size %s unavailable", cartAdd.Size)
              err = errors.New(msgg)
              t.StatusCh <- Status{Message: msgg}
              return err
            }
          }
          if cart.TotalProductCount < 1 || len(cart.ShipmentList) == 0 {
            err = errors.New("Add to cart error: cart has no products" + string(cartJson[:]))

            t.StatusCh <- Status{Message: fmt.Sprintf("Add to cart failed \"%s\". Retrying...", cart.Message)}
            return err
          }
        }
        // return nil

      t.SetCookie(&http.Cookie{Name: "restoreBasketUrl", Value: "%2Fon%2Fdemandware.store%2FSites-ys-US-Site%2Fen_US%2FCart-UpdateItems%3Fpid_0%3D" + skuStr + "%26qty_0%3D1%26"})
      t.SetCookie(&http.Cookie{Name: "userBasketCount", Value: "1"})
      t.SetCookie(&http.Cookie{Name: "persistentBasketCount", Value: "1"})
      // t.SetCookie(&http.Cookie{Name: "pagecontext_cookies", Value: ""})

      t.EnableHttpTrace = rand.Float32() < 0.2

      var checkoutInitResp *http.Response
      var cret = 0
      t.Config["stage"] = "click"
      t.YsGenCookie(akUrl)
      t.YsGetCart()
      return t.Retry(func() error {
        time.Sleep(time.Duration(rand.Intn(6000))*time.Millisecond)
        defer func() {
          cret += 1
        }()
        defer CloseH2Conns(t.client)
        t.Config["genurl"] = "https://www.yeezysupply.com/delivery"
        t.Config["stage"] = ""
        if cret > 0 {
          t.YsGetU("https://www.yeezysupply.com/delivery")
          if aup, err := url.Parse(akUrl); err == nil {
            headers = [][2]string {
              {"accept", "*/*"},
              {"accept-language", "en-US,en;q=0.5"},
              {"accept-encoding", "gzip, deflate, br"},
              {"referer", "https://www.yeezysupply.com/delivery"},
            }
            resp, _ := t.doReq(t.client, t.makeGetReq(aup, &headers))
            DiscardResp(resp)
            t.Config["akUrl"] = akUrl
            go func() {
              defer func() {
                recover()
              }()
              time.Sleep(time.Duration(rand.Intn(500))*time.Millisecond)
              t.YsGetXhrU("https://www.yeezysupply.com/api/checkout/baskets/" + cart.BasketId + "/shipping_methods")
            }()
            t.YsGenCookie(akUrl)
          }
        } else {
          t.Config["stage"] = "click"
          t.YsGenCookie(akUrl)
          t.YsGetXhrU("https://www.yeezysupply.com/api/checkout/baskets/" + cart.BasketId + "/shipping_methods")
        }


        t.LogDebug("Checking out size %s", cartAdd.Size)

        t.StartCheckoutTime = timeMillis()
        t.StatusCh <- Status{Message: fmt.Sprintf("Checking out size %s", cartAdd.Size), Level: 7}


        t.Config["stage"] = "click"
        t.YsGenCookie(akUrl)

        is = randHex(8)
        headers = [][2]string{
          {"checkout-authorization", authorization},
          {"content-type", "application/json"},
          {"accept", "*/*"},
          {"origin", YS_BASE[:len(YS_BASE)-1]},
          {"sec-fetch-mode", "cors"},
          {"referer", YS_BASE + "delivery"},
          {"x-instana-t", is},
          {"x-instana-s", is},
          {"x-instana-l", "1,correlationType=web;correlationId=" + is},
        }
        checkoutInit := CheckoutInit {
          Customer: Customer {
            Email: profile.Email,
            ReceiveSmsUpdates: false,
          },
          MethodList: []ShippingMethod{
            ShippingMethod {
              Id: cart.ShipmentList[0].ShippingLineItem.Id,
              ShipmentId: cart.ShipmentList[0].ShipmentId,
              CollectionPeriod: "",
              DeliveryPeriod: "",
            },
          },
          ShippingAddress: profile.ShippingAddress,
          BillingAddress: profile.BillingAddress,
          NewsletterSubscription: false,
        }
        checkoutInitJson, _ := json.Marshal(checkoutInit)
        checkoutInitUrl, err := url.Parse(fmt.Sprintf("%sapi/checkout/baskets/%s", YS_BASE, cart.BasketId))
        if err != nil {
          return err
        }
        t.EnablePrivateHttpTrace = true

        checkoutInitResp, err = t.doReq(t.client, t.makeReq("PATCH",
          checkoutInitUrl, &headers, nil, &checkoutInitJson))
        // defer DiscardResp(checkoutInitResp)
        t.EnablePrivateHttpTrace = false
        if checkoutInitResp != nil && checkoutInitResp.StatusCode == 400 {
          checkoutJson, _ := readBody(checkoutInitResp)
          go t.SendTelemetry(map[string]interface{} {
            "event": "ys_checkout_response",
            "status_code": checkoutInitResp.StatusCode,
            "ys_checkout_response": string(checkoutJson),
          })
          err := errors.New(fmt.Sprintf("Checkout failed: invalid profile %s", checkoutJson))
          return err
        }
        if err != nil {
          if checkoutInitResp != nil && checkoutInitResp.Body != nil {
            checkoutInitResp.Body.Close()
          }
          // checkoutJson, _ := readBody(checkoutInitResp)

          t.StatusCh <- Status{Message: fmt.Sprintf("Checkout failed: %+v", err)} // , "body", checkoutJson}
          return err
        }
        checkoutJson, err := readBodyBytes(checkoutInitResp)
        // TODO check status code, retries, ... harden
        if err != nil {

          t.StatusCh <- Status{Message: fmt.Sprintf("Checkout failed: %+v", err)}
          return err
        }
        go t.SendTelemetry(map[string]interface{} {
          "event": "ys_checkout_response",
          "ys_checkout_response": string(checkoutJson),
        })
        err = json.Unmarshal(checkoutJson, &cart)
        if err != nil {

          t.StatusCh <- Status{Message: fmt.Sprintf("Checkout failed: %+v", err)}
          return err
        }
        if len(cart.MessageList) < 1 || cart.MessageList[0].Type != "PaymentMethodRequired" {
          sentry.CaptureMessage("Checkout failure: unexpected cart message list" + string(checkoutJson[:]))
          t.StatusCh <- Status{Message: "Checkout failure: unexpected cart message list" + string(checkoutJson[:])}
          return errors.New("Checkout failure: unexpected cart message list" + string(checkoutJson[:]))
        }
        if checkoutInitResp ==nil {
          return ErrRetrying
        }


        t.Retry(func() error {
          cart, err := t.YsGetCart()
          if err == nil && cart.BasketId != "" && cart.TotalProductCount < 1 {
            return retry.Unrecoverable(retry.Unrecoverable(ErrCartJacked))
          }
          t.LogDebug("Get cart: %v", cart)
          return err
        }, retry.Attempts(3))

      var alreadyDeclined bool
      var pret = 0
      t.Config["stage"] = "click"
      t.YsGenCookie(akUrl)
      return t.Retry(func() error {

        defer func() {
          pret += 1
        }()
        t.Config["stage"] = ""
        t.Config["genurl"] = "https://www.yeezysupply.com/payment"
        defer CloseH2Conns(t.client)

        if pret > 0 {
          t.YsGetU("https://www.yeezysupply.com/payment")
          if aup, err := url.Parse(akUrl); err == nil {
            headers = [][2]string {
              {"accept", "*/*"},
              {"accept-language", "en-US,en;q=0.5"},
              {"accept-encoding", "gzip, deflate, br"},
              {"referer", "https://www.yeezysupply.com/payment"},
            }
            resp, _ := t.doReq(t.client, t.makeGetReq(aup, &headers))
            DiscardResp(resp)
            t.Config["akUrl"] = akUrl
            go func() {
              defer func() {
                recover()
              }()
              time.Sleep(time.Duration(rand.Intn(500))*time.Millisecond)
              t.YsGetXhrU("https://www.yeezysupply.com/api/checkout/baskets/" + cart.BasketId + "/payment_methods")
            }()
            t.YsGenCookie(akUrl)
          }
        } else {
          t.YsGetXhrU("https://www.yeezysupply.com/api/checkout/baskets/" + cart.BasketId + "/payment_methods")
        }


        time.Sleep(time.Duration(rand.Intn(5000))*time.Millisecond)

        t.Config["stage"] = "click"
        t.YsGenCookie(akUrl)

        t.LogDebug("Submitting payment")
        t.StatusCh <- Status{Message: "Submitting Payment", Level: 8}
        is = randHex(8)
        headers = [][2]string{
          {"checkout-authorization", authorization},
          {"content-type", "application/json"},
          {"accept", "*/*"},
          {"origin", YS_BASE[:len(YS_BASE)-1]},
          {"sec-fetch-mode", "cors"},
          {"referer", YS_BASE + "payment"},
          {"x-instana-t", is},
          {"x-instana-s", is},
          {"x-instana-l", "1,correlationType=web;correlationId=" + is},
        }
        adyenEnc, err := adyenEncrypt(profile.Card, cart.ModifiedDate)
        if err != nil {

          t.StatusCh <- Status{Message: fmt.Sprintf("Payment failed: %+v", err)}
          return err
        }
        if os.Getenv("DEBUG") != "" {
          t.LogDebug("adyenEnc %s\n", adyenEnc)
        }
        adyenFp, err := t.getAdyenFp(0)
        if err != nil {

          t.StatusCh <- Status{Message: fmt.Sprintf("Payment failed: %+v", err)}
          return err
        }

        payment := Payment {
          BasketId: cart.BasketId,
          EncryptedInstrument: adyenEnc,
          Fingerprint: adyenFp,
          PaymentInstrument: PaymentInstrument{
            CardType: profile.Card.Type,
            ExpMonth: profile.Card.ExpMonth,
            ExpYear: profile.Card.ExpYear,
            Holder: profile.Card.Name,
            LastFour: profile.Card.LastFour,
            PaymentMethodId: "CREDIT_CARD",
          },
        }
        paymentJson, _ := json.Marshal(payment)
        t.LogDebug("pay json %s", string(paymentJson))
        paymentUrl, err := url.Parse(fmt.Sprintf("%sapi/checkout/orders", YS_BASE))
        if err != nil {

          return err
        }
        t.StartPaymentTime = timeMillis()
        // t.YsGenCookie(akUrl)
        t.EnablePrivateHttpTrace = true


        // ys common req func
          // gen clicks/keys before any POST
        // add missing checkout reqs
        // add token sharing via farm cache
          // global token sharing for those w/o farm
        // pixel ff

        paymentResp, err := t.doReq(t.client, t.makeReq("POST",
          paymentUrl, &headers, nil, &paymentJson))
        defer DiscardResp(paymentResp)
        if paymentResp == nil {
          if err != nil {
            return err
          } else {
            return ErrRetrying
          }
        }
        t.EnablePrivateHttpTrace = false

        if paymentResp != nil && paymentResp.StatusCode == 403 {
          if _, ok := failuresByStage["pay"]; ok {
            failuresByStage["pay"] += 1
          } else {
            failuresByStage["pay"] = 1
          }
          if failuresByStage["pay"] > 3 {
            t.StatusCh <- Status{Message: "Shadow banned at ATC. Restarting task..."}
            return TaskErrUnrecoverable(retry.Unrecoverable(retry.Unrecoverable(ErrRetrying)))
          }
          t.YsGenCookie(akUrl)
          return ErrRetrying
        }
        paymentRespJson, err := readBody(paymentResp)
        if err != nil {

          return err
        }
        go t.SendTelemetry(map[string]interface{} {
          "event": "ys_payment_response",
          "status_code": paymentResp.StatusCode,
          "ys_payment_response": string(paymentRespJson),
        })
        if paymentResp.StatusCode == 201 {
          t.StatusCh <- Status{Message: "COOKED!!!", Level: 9999} // , "size", cartAdd.Size, "orderJson", paymentRespJson}
          t.LogDebug(fmt.Sprintf("COOKED"))
          go t.SuccessWebhook("YS", productStyleCode, sizeChoice)
          t.FtlSendSessionSnapshot()
          return nil
        } else {
            var paymentError AdyenPaymentError
            err = json.Unmarshal([]byte(paymentRespJson), &paymentError)
            if err != nil {
              t.StatusCh <- Status{Message: fmt.Sprintf("Payment failed: %s", paymentResp.Status)}
              return err
            }
            go t.SendTelemetry(map[string]interface{} {
              "event": "ys_payment_response",
              "status_code": paymentResp.StatusCode,
              "response": paymentError,
            })
            t.StatusCh <- Status{Message: fmt.Sprintf("Payment failed: %s", paymentError.Message)}
            fmt.Printf("Payment failed: %s", paymentRespJson)
            err = errors.New(fmt.Sprintf("Payment failure: %+v", err))
            if strings.Contains(paymentError.Message, "declined") && !alreadyDeclined {
              alreadyDeclined = true
              go t.DeclineWebhook("YS", productStyleCode, sizeChoice)
              t.FtlSendSessionSnapshot()
              return nil
            }
            if strings.Contains(paymentError.Message, "basket") {
              return retry.Unrecoverable(err)
            }

            return err
        }
          // var paymentError AdyenPaymentError
          // err = json.Unmarshal([]byte(paymentRespJson), &paymentError)
          // if err == nil {
          //   t.StatusCh <- Status{Message: fmt.Sprintf("Payment failed: %s", paymentError.Message)}
          // } else {

          //   t.StatusCh <- Status{Message: fmt.Sprintf("Payment failed: %s", paymentResp.Status)}
          // }
          // // TODO: replace most log.Fatal with retry
          // fmt.Sprintf("Payment failed: %s", paymentRespJson)
          // return err
        return err
      }, retry.Attempts(16), retry.Delay(time.Duration(rand.Intn(10000))*time.Millisecond))
      }, retry.Attempts(16), retry.Delay(time.Duration(rand.Intn(10000))*time.Millisecond))
    }, retry.Attempts(16), retry.Delay(time.Duration(rand.Intn(10000))*time.Millisecond))
  }, retry.LastErrorOnly(true))
}

func (t *CheckoutTask) newRecapCookie(taskConfig TaskConfig, urlStr string, client *http.Client) error {
  return t.Retry(func () error {
    t.LogDebug("Refreshing Recaptcha...")
    tclient := http.Client{Timeout:30*time.Second}
    tokenReq := map[string]string {
      "url": urlStr,
      "apiKey": taskConfig.RcApiKey,
      "action": taskConfig.RcAction,
    }
    tokenReqJson, err := json.Marshal(tokenReq)
    if err != nil {
      return retry.Unrecoverable(err)
    }

    req, err := http.NewRequest("POST", "http://localhost:42424", ioutil.NopCloser(bytes.NewReader(tokenReqJson)))
    resp, err := tclient.Do(req)
    if err != nil {

      t.LogDebug("Error refreshing recaptcha: %+v", err)
      return err
    }
    tokenBytes, err := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    if err != nil {

      t.LogDebug("Error refreshing recaptcha: %+v", err)
      return err
    }
    if resp.StatusCode != 200 {
      err = errors.New(fmt.Sprintf("Error refreshing recaptcha: invalid status %d", resp.StatusCode))
      t.LogDebug("Error refreshing recaptcha: %+v", err)
      return err
    }
    token := string(tokenBytes[:])
    t.LogDebug("Set RC cookie %s", token)

    client.Jar.SetCookies(t.BaseUrl,
      []*http.Cookie{
        &http.Cookie{ Name: taskConfig.RcCookie, Value: token, },
      },
    )
    return nil
  }, retry.Attempts(1e3), retry.Delay(3e3 * time.Millisecond))
}

type YsRcCookie struct {
  Name string `json:"rc_cookie"`
  Expires uint64 `json:"expires"`
  Values []string `json:"rc_cookie_vals"`
}

var ErrNoCookieAvailable = errors.New("No Cookie Available")
var YsCkieIndexUrl, _ = url.Parse("https://rushaio.s3.amazonaws.com/ysckieindex.txt")
var YsStaticCookieUrl, _ = url.Parse("https://rushaio.s3.amazonaws.com/ys_cookies.txt")
var Crc32Q = crc32.MakeTable(0xD5828281)

func (t *CheckoutTask) YsSetStaticCookies() error {
  resp, err := t.doReq(t.s3client, t.makeGetReq(YsStaticCookieUrl, nil))
  defer DiscardResp(resp)
  if err != nil {
    DiscardResp(resp)
    return err
  }
  var ckies [][2]string
  err = readRespJsonDst(resp, &ckies)
  if err != nil {
    return err
  }
  t.LogDebug("YS Static Cookies: %+v", ckies)
  for _, c := range ckies {
    if t.GetCookieValue(c[0]) == "" {
      t.SetCookie(&http.Cookie{Name: c[0], Value: c[1], })
    }
  }
  return nil
}

func (t *CheckoutTask) YsLiveGenToken() (*YsRcCookie, error) {
  client := &http.Client{Timeout:10*time.Second}
  su, _ := url.Parse("http://127.0.0.1:9999/")
  resp, err := t.doReq(client, t.makeReq("GET", su, &[][2]string{ {"url", t.Url.String()}, {"enterprise", "1"}, { "proxy", t.ProxyStr }}, nil, nil))
  defer DiscardResp(resp)
  if err != nil {
    return nil, err
  }
  if resp.StatusCode != 200 {
    return nil, ErrRetrying
  }
  if tok, err := readBody(resp); err == nil && len(tok) > 32 {
    t.SetCookie(&http.Cookie{Name: "xhwUqgFqfW88H50", Value: tok, Expires: time.Now().Add(120*time.Second), })
    return &YsRcCookie{Name: "xhwUqgFqfW88H50", Values: []string{tok}}, nil
  }
  return nil, ErrFailedToGenCookie
}

func (t *CheckoutTask) YsPullCloudHarvesterCookie() (*YsRcCookie, error) {
  rckie := YsRcCookie{}
  ckieslots, err := t.doReq(t.s3client, t.makeGetReq(YsCkieIndexUrl, nil))
  defer DiscardResp(ckieslots)
  if err != nil {
    DiscardResp(ckieslots)
    return nil, err
  }
  slotsLines, err := readBody(ckieslots)
  if err != nil {
    return nil, err
  }
  slots := strings.Split(slotsLines, "\n")
  if len(slots) == 0 || len(slots) == 1 && slots[0] == "" {
    return nil, ErrNoCookieAvailable
  }

  var slot uint32
  // t.TelemetryContextMut.Lock()
  // if t.TelemetryContext["ys_use_rand_token"] == "1" {
  slot = uint32(rand.Intn(len(slots)))
  // } else {
  //   slot = crc32.ChecksumIEEE([]byte(t.Id))
  // }
  // t.TelemetryContextMut.Unlock()

  // slots = append([]string{"https://rushaio.s3.amazonaws.com/cookie1.txt"}, slots...)
  // if host, ok := t.DnsCacheActiveHost[t.Url.Host]; ok {
  //   slot = crc32.ChecksumIEEE([]byte(host))
  //   // t.LogDebug("host (%s) slot=%+v", host, slot)
  // } else {
  //   // t.LogDebug("tid (%s) slot=%+v", t.Id, slot)
  // }

  slotUrl, err := url.Parse(strings.TrimSpace(slots[slot % uint32(len(slots))]))
  t.LogDebug("%+v %+v %+v %+v", len(slots), slots, slotUrl, err)
  if err != nil {
    return nil, err
  }
  slotresp, err := t.doReq(t.s3client, t.makeGetReq(slotUrl, nil))
  if err != nil {
    DiscardResp(slotresp)
    return nil, err
  }
  err = readRespJsonDst(slotresp, &rckie)
  return &rckie, err
}

type YsWaitingRoomConfig struct {
  EnableAvailabilityGrid bool     `json:"enableAvailabilityGrid"`
  EnableRecaptchaV3      bool     `json:"enableRecaptchaV3"`
  UseWrSlim              bool     `json:"useWrSlim"`
  YeezySupplyWrMessage   interface{} `json:"yeezySupplyWrMessage"`
  YsStatusMessageKey     string   `json:"ysStatusMessageKey"`
}

var YS_SALE_HOLD_STATUSES = []string{"welcome", "sold_out", "sale_on_hold"}

func (c *YsWaitingRoomConfig) IsSaleActive() bool {
  return !common.Any(YS_SALE_HOLD_STATUSES, func(s string) bool { return c.YsStatusMessageKey == s || c.YeezySupplyWrMessage == s  })
}


var YsWrConfigUrl, _ = url.Parse("https://www.yeezysupply.com/hpl/content/yeezy-supply/config/US/waitingRoomConfig.json")

func (t *CheckoutTask) YsGetWaitingRoomConfig() (*YsWaitingRoomConfig, error) {
  var config YsWaitingRoomConfig
  headers := [][2]string {
    {"referer", t.Url.String()},
    {"sec-fetch-dest", "empty"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-site", "same-origin"},
    {"accept", "application/json, text/plain, */*"},
  }
  req := t.makeGetReq(YsWrConfigUrl, &headers)
  resp, err := t.doReq(t.client, req)
  defer DiscardResp(resp)
  if err != nil {
    return nil, err
  }
  if err = readRespJsonDst(resp, &config); err == nil {
    return &config, nil
  } else {
    return nil, err
  }
}

var StatusInWaitingRoomSaleStopped = Status{Message: "Waiting Room (Sale Not Started)"}
var StatusInWaitingRoomSaleStarted = Status{Message: "Waiting Room (Sale Started)"}

func (t *CheckoutTask) YsHandleWaitingRoomSession() error {
  rcCkie := &YsRcCookie{Name: "xhwUqgFqfW88H50"}

  mut := sync.Mutex{}
  lastCookiePulledAt := time.Time{}
  t.StatusCh <- StatusInWaitingRoom
  if t.StatusTelemetryEnabled {
    go t.SendTelemetry(map[string]interface{} {
      "event": "ys_wr_thrownback",
      "jar": GetAllClientCookies(t.client),
    })
  }
  lastJar := &[]cookiejar.Entry{}
  // todo pull cookie parallel to session
  var wrConfig YsWaitingRoomConfig
  if wrConfig_, _ := t.YsGetWaitingRoomConfig(); wrConfig_ != nil {
    wrConfig = *wrConfig_
  }
  ret := 0
  wrpoll := 3 + rand.Intn(15)
  return t.Retry(func() error {
    ret += 1
    if ret % wrpoll == 0 {
      if wrConfig_, _ := t.YsGetWaitingRoomConfig(); wrConfig_ != nil {
        wrConfig = *wrConfig_
      }
    }
    // wrConfig.YsStatusMessageKey = "welcome"
    t.LogDebug("wrConfig %+v", wrConfig)
    if !wrConfig.IsSaleActive() {
      if rcCkie != nil && len(rcCkie.Name) > 0 {
        t.RemoveCookie(rcCkie.Name)
      }
      t.StatusCh <- StatusInWaitingRoomSaleStopped
      return ErrRetrying
    }
    t.StatusCh <- StatusInWaitingRoomSaleStarted
    lastJar = GetAllClientCookies(t.client)
    var rcCkieStr string
    mut.Lock()
    if time.Now().Sub(lastCookiePulledAt) > 115*time.Second && (wrConfig.EnableRecaptchaV3 || len(wrConfig.YsStatusMessageKey) == 0)  {
      // t.RemoveCookie(rcCkie.Name)
      lastCookiePulledAt = time.Now()
      mut.Unlock()
      go func() {
        if ckie, err := t.YsLiveGenToken(); err == nil {
          t.LogDebug("token %+v", ckie)
        } else {
          if rcCkie_, err := t.YsPullCloudHarvesterCookie(); rcCkie_ != nil && err == nil{
            log.Printf("%+v", rcCkie)
            rcCkie = rcCkie_
            if len(rcCkie.Values) > 0 {
              mut.Lock()
              defer mut.Unlock()
              t.SetCookie(&http.Cookie{Name: "xhwUqgFqfW88H50", Value: rcCkie.Values[0], Expires: time.Now().Add(120*time.Second), })
            } else {
              lastCookiePulledAt = time.Time{}
            }
          } else {
            mut.Lock()
            defer mut.Unlock()
            lastCookiePulledAt = time.Time{}
          }
        }
      }()


    } else {
      mut.Unlock()
    }

    t.SetCookie(&http.Cookie{ Name: "akavpwr_ys_us", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "", })


    headers := make([][2]string, 0)
    headers = addHeader(headers, "referer", t.Url.String())
    headers = addHeader(headers, "sec-fetch-dest", "empty")
    headers = addHeader(headers, "sec-fetch-mode", "cors")
    headers = addHeader(headers, "sec-fetch-site", "same-origin")
    headers = addHeader(headers, "accept", "application/json, text/plain, */*")
    headerOrder := []string{"accept", "user-agent"}
    req := t.makeReq("GET", t.ApiConfig.slimUrl(), &headers, &headerOrder, nil)
    req.Close = false
    resp, rerr := t.doReq(t.client, req)
    defer DiscardResp(resp)

    t.SetCookie(&http.Cookie{ Name: "akavpwr_ys_us", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "", })

    if resp != nil {
      body, _ := readBody(resp)
      if os.Getenv("DEBUG") != "" {
        t.LogDebug("%s: %s", t.ApiConfig.slimUrl().String(), body)
      }
      if strings.Contains(t.GetCookieValue("xhwUqgFqfW88H50_u"), "1~hmac") {
        t.RemoveCookie("xhwUqgFqfW88H50_u")
        return retry.Unrecoverable(ErrRetrying)
      }
      if resp.StatusCode == 200 {
        setckies := resp.Header.Values("set-cookie")
        if len(setckies) > 0 && rcCkie != nil && len(rcCkie.Name) > 0 {
          for _, ckie := range setckies {
            if strings.Contains(ckie, rcCkie.Name + "=_pls_remove_me_")  {
              t.RemoveCookie(rcCkie.Name)
              break
            }
          }
        }
        t.RemoveCookie("xhwUqgFqfW88H50")
        go t.SendTelemetry(map[string]interface{} {
          "event": "ys_wr_pass_cookies",
          "ys_wr_pass_cookies": map[string]interface{} {
            "status_code": resp.StatusCode,
            "rctoken": rcCkieStr,
            "lastJar_all": lastJar,
            "currentJar_all": GetAllClientCookies(t.client),
            "resp_header": resp.Header,
          },
          "__sendnow__": "1",
        })
        t.EnableHttpTrace = rand.Float32() < 0.02
        t.StatusTelemetryEnabled = true
        return nil
      } else {
        return ErrRetrying
      }
    } else {
      return rerr
    }
    return ErrRetrying
  }, retry.DelayType(retry.FixedDelay), retry.Delay(time.Duration(3e3) * time.Millisecond))
}

func (t *CheckoutTask) waitForProductAk(taskConfig TaskConfig, productUrl *url.URL, client *http.Client) (string, error) {
  body, err := t.getProductPageBody(client, taskConfig, productUrl)
  if err != nil {
    return "", err
  }
  return akUrlFromPageBody(YS_BASE, body)
}

func (t *CheckoutTask) YsGetApiProduct() (*YsProduct, error) {
  productStyleCode := t.Url.Path[strings.LastIndex(t.Url.Path, "/")+1:]

  var prod YsProduct
  is := randHex(8)
  headers := [][2]string {
    {"referer", t.Url.String()},
    {"sec-fetch-dest", "empty"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-site", "same-origin"},
    {"content-type", "application/json"},
    {"accept", "*/*"},
    {"origin", "https://www.yeezysupply.com"},
    {"x-instana-t", is},
    {"x-instana-s", is},
    {"x-instana-l", "1,correlationType=web;correlationId=" + is},
  }
  prodApiUrl, err := url.Parse(fmt.Sprintf("https://www.yeezysupply.com/api/products/%s", productStyleCode))
  if err != nil {
    return nil, err
  }
  req := t.makeGetReq(prodApiUrl, &headers)
  resp, err := t.doReq(t.client, req)
  defer DiscardResp(resp)
  if err != nil {
    return nil, err
  }
  if err = readRespJsonDst(resp, &prod); err == nil {
    return &prod, nil
  } else {
    return nil, err
  }
}

var UTC, _ = time.LoadLocation("UTC")

func (t *CheckoutTask) getProductPageBody(client *http.Client, taskConfig TaskConfig, productUrl *url.URL) (string, error) {
  var body string
  var retries int = 0
  headers := [][2]string {
    // {}
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"accept-language", "en-US,en;q=0.5"},
    {"accept-encoding", "gzip, deflate, br"},
    {"upgrade-insecure-requests", "1"},
    // {"referer", t.BaseUrl.String()},
  }
  var releaseTime time.Time

  hasVisitedWr := false

  err := t.Retry(func() error {
    t.StatusCh <- StatusVisitProduct
    resp, err := t.doReq(client, t.makeGetReq(productUrl, &headers))
    defer DiscardResp(resp)
    if err != nil {
      return err
    }
    if retries == 0 {
      t.YsSetStaticCookies()
    }
    // log.Println("past")
    retries += 1
    body, err = readBody(resp)
    // log.Println(body)
    if err != nil {
      return err
    }
    if strings.Contains(body, "wrgen_orig_assets") {
      hasVisitedWr = true
      t.YsHandleWaitingRoomSession()
      return ErrRetrying
    } else if strings.Contains(body, "You don't have permission to access") {
      t.StatusCh <- Status{Message: "Product page forbidden. Restarting task"}
      return TaskErrUnrecoverable(ErrAkamaiBanned)
    } else {
      if !hasVisitedWr {
        if releaseTime.IsZero() {
          if prod, _ := t.YsGetApiProduct(); prod != nil {
            if pttime, err := time.ParseInLocation("2006-01-02T15:04:05.000Z", prod.AttributeList.PreviewTo, UTC); err == nil {
              releaseTime = pttime
            }
          }
        }
        if os.Getenv("RELEASE") != "0" && !releaseTime.IsZero() {
          t.LogDebug("previewTo: %v", releaseTime)
          dtRelease := releaseTime.Sub(time.Now())
          if dtRelease > 0 {
            t.LogDebug("waiting for %v", dtRelease)
            t.StatusCh <- StatusWaitingForRelease
            go func() {
              defer func() {
                recover()
              }()
              if akUrl, err := akUrlFromPageBody(YS_BASE, body); err == nil {
                if aup, err := url.Parse(akUrl); err == nil {
                  headers = [][2]string {
                    {"accept", "*/*"},
                    {"accept-language", "en-US,en;q=0.5"},
                    {"accept-encoding", "gzip, deflate, br"},
                    {"referer", t.BaseUrl.String()},
                  }
                  resp, _ := t.doReq(t.client, t.makeGetReq(aup, &headers))
                  DiscardResp(resp)
                  t.Config["akUrl"] = akUrl
                  t.YsGenCookie(akUrl)
                }
              }
            }()
            time.Sleep(dtRelease)
            return ErrRetrying
          }
        }
      }
      return nil
    }
  }, retry.Delay(1*time.Second))

  return body, err
}

type SizeAvailability struct {
  Sku string `json:"sku"`
  Stock int `json:"availability"`
  Status string `json:"availability_status"`
  Size string `json:"size"`
}
type Availability struct {
  Id string `json:"id"`
  Status string `json:"availability_status"`
  Sizes []SizeAvailability `json:"variation_list"`
}

type YsV2Size struct {
  DisplaySize      string `json:"displaySize"`
  HypeAvailability int    `json:"hypeAvailability"`
}
type YsAvailabilityV2 struct {
  Availability string `json:"availability"`
  Skus         map[string]YsV2Size `json:"skus"`
}
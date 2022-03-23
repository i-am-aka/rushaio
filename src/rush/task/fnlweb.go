package task

import (
  "os"
  "bytes"
  "encoding/base64"
  "encoding/json"
  "fmt"
  "github.com/avast/retry-go"
  "github.com/pkg/errors"
  "github.com/PuerkitoBio/goquery"
  "io"
  "io/ioutil"
  "math/rand"
  "net/url"
  "regexp"
  "rush/net/http"
  "strings"
  "time"
)

var FNL_SITEKEY = "6LcPD74ZAAAAAFtuJvnpIV5VjKE16Ma7oRCcENIN"
var JD_SITEKEY = "6LdpAb4ZAAAAAKTaBdqCW12WwD3BWp63gKQtNMcC"
var (
  ErrStartCheckoutFailed = errors.New("Start checkout failed")
	errCardDeclined = errors.New("Card Declined")
	errAtcFailed = errors.New("ATC Failed")
  errUnknownCheckoutResult = errors.New("Unknown Checkout Result")
)

type FnlWebProduct struct {
	ColorId string
	Name string
	Pid string
	Skus []FnlWebSku
	StyleId string
  Color string
}

type FnlWebSku struct {
  ProductId string `json:"productId"`
	SkuId string `json:"sku"`
	StockLevel string `json:"stockLevel"`
	Size string `json:"sizeValue"`
}

type FnlWebSession struct {
  Task *CheckoutTask
  AkUrl string
  abckRetries uint
  EnableRecap bool
}
// FNL WEB (DESKTOP): https://" + s.Task.Url.Host + "
// FNL APP: 					app://" + s.Task.Url.Host + "


func (s *FnlWebSession) FnlLiveGenToken(action string) (string, error) {
  s.Task.StatusCh <- Status{Message: "Refreshing Recaptcha (" + action + ")"}
  var tok string
  err := s.Task.Retry(func() error {
    client := &http.Client{Timeout:10*time.Second}
    su, _ := url.Parse("http://127.0.0.1:9999/")

      req := s.Task.makeReq("GET", su, &[][2]string{ { "enterprise", "1" }, { "proxy", s.Task.ProxyStr }, {"url", s.Task.BaseUrl.String()}, {"action", action}, {"sitekey", s.SiteKey()} }, nil, nil)
      var b bytes.Buffer
      req.Write(&b)
      s.Task.LogDebug("%s", string(b.Bytes()))
    resp, err := s.Task.doReq(client, s.Task.makeReq("GET", su, &[][2]string{ { "enterprise", "1" }, { "proxy", s.Task.ProxyStr }, {"url", s.Task.BaseUrl.String()}, {"action", action}, {"sitekey", s.SiteKey()} }, nil, nil))
    if err != nil {
      s.Task.StatusCh <- StatusFarmUnreachable
      return err
    }
    if resp.StatusCode != 200 {
      return ErrRetrying
    }
    if tok, err = readBody(resp); err == nil {
      return nil
    }
    return ErrRetrying
  }, retry.Delay(500*time.Millisecond), retry.Attempts(16))
  return tok, err
}

func (s *FnlWebSession) SiteKey() string {
  if s.Task.Url.Host == "www.finishline.com" {
    return FNL_SITEKEY
  } else {
    return JD_SITEKEY
  }
}
func (s *FnlWebSession) DoReq(req *http.Request, stage string) (*http.Response, error) {
  var err error
  var resp *http.Response

  // if s.Task.Config["last_valid_abck"] != "" {
  //   s.Task.client.Jar.SetCookies(s.Task.BaseUrl, []*http.Cookie{ &http.Cookie{ Name: "_abck", Value: s.Task.Config["last_valid_abck"], Path: "/", }})
  // }

  // if req.Method == "POST" || req.Method == "PUT" {
  //   s.Task.StatusCh <- Status{Message: fmt.Sprintf("Generating Cookie (%s)", stage)}
  //   s.Task.EnsureValidAbck(s.AkUrl, s.hasValidAbck)
  // }

  genfails := 0

  err = s.Task.Retry(func() error {
    resp, err = s.Task.doReq(s.Task.client, req)
    if resp != nil {
      bodyraw, err := ioutil.ReadAll(resp.Body)
      if err != nil {
        return err
      }
      rr := &http.Response{Body: ioutil.NopCloser(bytes.NewReader(bodyraw)), Header: resp.Header}
      body, err := readBodyBytes(rr)
      bodystr := strings.ToLower(string(body))
      if strings.Contains(bodystr, "slow down") {
        time.Sleep(30 * time.Second)
        return ErrRateLimited
      }
      // s.Task.LogDebug("WAITING: %+v", strings.Contains(bodystr, "<title>WAITING</title>"))
      // log.Println(bodystr)
      jdUrl, _ := url.Parse("https://www.jdsports.com/")
      fnlUrl, _ := url.Parse("https://www.finishline.com/")
      for _, c := range s.Task.client.Jar.Cookies(jdUrl) {
        if strings.Contains(c.Name, "akavpwr") {
          s.Task.StatusCh <- StatusInQueue
          s.Task.GenericQueueBypassUrl(jdUrl)
          return errInQueue
        }
      }
      for _, c := range s.Task.client.Jar.Cookies(fnlUrl) {
        if strings.Contains(c.Name, "akavpwr") {
          s.Task.StatusCh <- StatusInQueue
          s.Task.GenericQueueBypassUrl(fnlUrl)
          return errInQueue
        }
      }

      if strings.Contains(bodystr, "<title>WAITING</title>") {
        // TODO what to do when no akavpwr
        // no akavpwr, restart task?
        return errInQueue
      }
      for _, c := range s.Task.client.Jar.Cookies(s.Task.Url) {
        if strings.Contains(c.Name, "akavpwr") {
          s.Task.StatusCh <- StatusInQueue
          s.Task.GenericQueueBypass()
          return errInQueue
        }
      }
      if strings.Contains(bodystr, "little crowded right now") {
        s.Task.StatusCh <- StatusInQueue
        return errInQueue
      }
      resp.Body = ioutil.NopCloser(bytes.NewReader(bodyraw))

      if resp.StatusCode == 403 {
        DiscardResp(resp)
        // `if strings.HasSuffix(s.Task.GetCookieValue("_abck"), "~0~-1~-1") {
                //   s.Task.StatusCh <- StatusAkaBanned
                //   return TaskErrUnrecoverable(ErrAkamaiBanned)
                // }`
        var hasabck bool
        for _, ckie := range resp.Header["Set-Cookie"]  {
          if strings.Contains(ckie, "_abck=") {
            hasabck = true
            break
          }
        }
        if hasabck && s.AkUrl != "" {
          s.Task.StatusCh <- Status{Message: fmt.Sprintf("Generating Cookie (%s)", stage)}
          s.Task.EnsureValidAbck(s.AkUrl, s.hasValidAbck)
          genfails += 1
          if genfails > 10 {
            return TaskErrUnrecoverable(ErrFailedToGenCookie)
          }
          return ErrRetrying
        } else {
          s.Task.StatusCh <- StatusAkaBanned
          return TaskErrUnrecoverable(ErrAkamaiBanned) // TODO auto rotate prox when shared prox distribution implemented
        }
      } else if resp.StatusCode == 429 {
        DiscardResp(resp)
        s.Task.StatusCh <- StatusProxyBanned
        return retry.Unrecoverable(retry.Unrecoverable(ErrProxyBanned))
      }
    }
    return err
  }, retry.Attempts(1e3), retry.Delay(3*time.Second))
  s.Task.LogDebug("DoReq err %+v", err)
  return resp, err
}

func (s *FnlWebSession) FnlWebGetProductSkus(product *FnlDesiredProduct) ([]FnlWebSku, error) {
	skus := []FnlWebSku{}

  url_, err := url.Parse("https://" + s.Task.Url.Host + "/store/browse/json/productSizesJson.jsp?productId=" + product.ProductId + "&styleId=" + product.StyleId + "&colorId=" + product.ColorId)
  if err != nil {
    return skus, err
  }
  headerOrder := []string {
    "accept",
    "user-agent",
    "x-requested-with",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "*/*"},
    {"x-requested-with", "XMLHttpRequest"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", s.Task.Url.String()},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  s.Task.Config["stage"] = "Stock Check"
  resp, err := s.DoReq(s.Task.makeReq("GET", url_, &headers, &headerOrder, nil), "Stock Check")
  // s.Task.LogDebug("%+v %+v", resp, err)
  if err != nil {
    return skus, err
  }
  var respJson map[string][]FnlWebSku
  err = readRespJsonDst(resp, &respJson)
  if err != nil {
    return skus, err
  }
  desiredProductId := fmt.Sprintf("%s-%s", product.StyleId, product.ColorId)
  for _, skus_ := range respJson {
  	for _, sku := range skus_[:] {
      if sku.ProductId != desiredProductId {
        continue
      }
		 	skuId, _ := base64.StdEncoding.DecodeString(sku.SkuId)
		 	stockLevel, _ := 	base64.StdEncoding.DecodeString(sku.StockLevel)
		 	sku.SkuId = string(skuId)
		 	sku.StockLevel = string(stockLevel)
      // s.Task.LogDebug("%+v", sku)
		 	if sku.StockLevel != "0" && sku.StockLevel != "" {
		 		skus = append(skus, sku)
		 	}
  	}
  	break
  }
  rand.Shuffle(len(skus), func(i, j int) {
    skus[i], skus[j] = skus[j], skus[i] })

  return skus, nil
}

func (s *FnlWebSession) GetProduct(desiredProduct *FnlDesiredProduct) (*FnlWebProduct, error) {
  var product FnlWebProduct

  err := s.Task.Retry(func() error {
    product.Pid = desiredProduct.ProductId
    product.StyleId = desiredProduct.StyleId
    product.ColorId = desiredProduct.ColorId

    headerOrder := []string {
      "upgrade-insecure-requests",
      "user-agent",
      "accept",
      "sec-fetch-site",
      "sec-fetch-mode",
      "sec-fetch-user",
      "sec-fetch-dest",
      "referer",
      "accept-encoding",
      "accept-language",
      "cookie",
    }
    headers := [][2]string {
      {"upgrade-insecure-requests", "1"},
      {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
      {"sec-fetch-site", "same-origin"},
      {"sec-fetch-mode", "navigate"},
      {"sec-fetch-user", "?1"},
      {"sec-fetch-dest", "document"},
      {"accept-encoding", "gzip, deflate, br"},
      {"accept-language", "en-US,en;q=0.9"},
    }

    surl, err := url.Parse("https://" + s.Task.Url.Host + "/store/browse/productDetail.jsp?productId=" + product.Pid + "&styleId=" + product.StyleId + "&colorId=" + product.ColorId)
      if err != nil {
        return retry.Unrecoverable(err)
      }
    s.Task.Config["stage"] = "Visiting Product"
    resp, err := s.DoReq(s.Task.makeReq("GET", surl, &headers, &headerOrder, nil), "Visiting Product")
    // if resp.StatusCode == 403 {
    //   return retry.Unrecoverable(errForbidden)
    // }
    s.Task.LogDebug("%+v %+v", resp, err)
    if err != nil {
      if resp != nil {
        io.Copy(ioutil.Discard, resp.Body)
        resp.Body.Close()
      }
      return err
    }
    bodyBytes, err := readBodyBytes(resp)
    if err != nil {
      s.Task.LogDebug("%+v %+v", resp, err)
      return err
    }

    s.AkUrl, err = akUrlFromPageBody("https://" + s.Task.Url.Host + "/", string(bodyBytes))
    if err != nil {
      s.AkUrl = s.Task.ApiConfig.SensorEndpoints[s.Task.Url.Host]
    }
    s.Task.Config["akUrl"] = s.AkUrl

    doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
    if err != nil {
      s.Task.LogDebug("%+v %+v", resp, err)
      return err
    }

    product.Name = doc.Find("meta[property='og:title']").AttrOr("content", "")
    product.Color = doc.Find(fmt.Sprintf("div#stylecolor_%s-%s span.description", product.StyleId, product.ColorId)).Text()
    doc.Find(".sizeOptions").Each(func(_ int, s *goquery.Selection) {
      if strings.Contains(s.AttrOr("class", ""), "disable") {
        return
      }
      skuId, _ := base64.StdEncoding.DecodeString(s.AttrOr("data-sku", ""))
      sku := FnlWebSku {
        ProductId: fmt.Sprintf("%s-%s", product.StyleId, product.ColorId),
        SkuId: string(skuId),
        Size: s.AttrOr("data-size", ""),
      }
      product.Skus = append(product.Skus, sku)
    })

    return nil
  }, retry.Attempts(8))

  return &product, err
}

func (s *FnlWebSession) GetHome() (error) {
  err := s.Task.Retry(func() error {
    headerOrder := []string {
      "upgrade-insecure-requests",
      "user-agent",
      "accept",
      "sec-fetch-site",
      "sec-fetch-mode",
      "sec-fetch-user",
      "sec-fetch-dest",
      "referer",
      "accept-encoding",
      "accept-language",
      "cookie",
    }
    headers := [][2]string {
      {"upgrade-insecure-requests", "1"},
      {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
      {"sec-fetch-site", "same-origin"},
      {"sec-fetch-mode", "navigate"},
      {"sec-fetch-user", "?1"},
      {"sec-fetch-dest", "document"},
      {"accept-encoding", "gzip, deflate, br"},
      {"accept-language", "en-US,en;q=0.9"},
    }
    req := s.Task.makeReq("GET", s.Task.BaseUrl, &headers, &headerOrder, nil)
    s.Task.Config["stage"] = "Visiting Homepage"
    resp, err := s.DoReq(req, "Visiting Homepage")
    // if resp.StatusCode == 403 {
    //   return retry.Unrecoverable(errForbidden)
    // }
    fmt.Printf("visited %+v %+v", resp, err)
    if err != nil {
      if resp != nil {
        io.Copy(ioutil.Discard, resp.Body)
        resp.Body.Close()
      }
      return err
    }
    fmt.Printf("%+v\n", GetAllClientCookies(s.Task.client))
    return nil
    bodyBytes, err := readBodyBytes(resp)
    if err != nil {
      return err
    }

    s.AkUrl, err = akUrlFromPageBody("https://" + s.Task.Url.Host + "/", string(bodyBytes))
    if err != nil {
      return err
    }

    return nil
  })

  return err
}


type FnlWebCart struct {

}

func (s *FnlWebSession) HasCart() bool {
  return s.Task.GetCookieValue("fl_shoppingcart_items") != ""
}

func (s *FnlWebSession) StartSession() error {
  return s.FnlWebAtcGet(fmt.Sprintf("https://%s/store/storepickup/json/storePickUpSession.jsp", s.Task.Url.Host))
}

func (s *FnlWebSession) FnlWebAtc(product *FnlDesiredProduct, sku *FnlWebSku) (*FnlWebCart, error) {
  var loc string
  var cart FnlWebCart
  err := s.Task.Retry(func() error {
    url_, err := url.Parse("https://" + s.Task.Url.Host + "/store/browse/productDetail.jsp?productId=" + product.ProductId + "&_DARGS=/store/browse/productDetailDisplay.jsp.flAddToCart")
    if err != nil {
      return err
    }
    headerOrder := []string {
      "content-length",
      "accept",
      "x-requested-with",
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
      {"accept", "*/*"},
      {"x-requested-with", "XMLHttpRequest"},
      {"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
      {"origin", "https://" + s.Task.Url.Host + ""},
      {"sec-fetch-site", "same-origin"},
      {"sec-fetch-mode", "cors"},
      {"sec-fetch-dest", "empty"},
      {"referer", s.Task.Url.String()},
      {"accept-encoding", "gzip, deflate, br"},
      {"accept-language", "en-US,en;q=0.9"},
    }
    var rctoken string
    if s.EnableRecap {
      rctoken, err = s.FnlLiveGenToken("ADD_TO_CART_EVENT")
      if err != nil {
        return err
      }
    }
    s.Task.LogDebug(rctoken)
    params := [][2]string {
      {"_dyncharset", "UTF-8"},
      {"requiresSessionConfirmation", "false"},
      {"/atg/commerce/order/purchase/CartModifierFormHandler.addItemToOrderSuccessURL", "/store/browse/productDetail.jsp?productId=" + product.ProductId},
      {"_D:/atg/commerce/order/purchase/CartModifierFormHandler.addItemToOrderSuccessURL", " "},
      {"/atg/commerce/order/purchase/CartModifierFormHandler.addItemToOrderErrorURL", "/store/browse/productDetail.jsp?productId=" + product.ProductId + "&dontCachePDP=true"},
      {"_D:/atg/commerce/order/purchase/CartModifierFormHandler.addItemToOrderErrorURL", " "},
      {"catalogRefIds", sku.SkuId},
      {"_D:catalogRefIds", " "},
      {"productId", product.ProductId},
      {"_D:productId", " "},
      {"items", ""},
      {"_D:items", " "},
      {"quantity", "1"},
      {"_D:quantity", " "},
      {"/atg/commerce/order/purchase/CartModifierFormHandler.dimensionId", ""},
      {"_D:/atg/commerce/order/purchase/CartModifierFormHandler.dimensionId", " "},
      {"addToCartExecutionToken", rctoken},
      {"Add To Cart", "Add+To+Cart"},
      {"_D:Add To Cart", " "},
      {"_DARGS", "/store/browse/productDetailDisplay.jsp.flAddToCart"},
    }
    body := urlEncodeParams(params)


    s.Task.Config["stage"] = "Add to Cart"

    // s.Task.AbckJar = [][]*http.Cookie {

    //   []*http.Cookie {
    //       &http.Cookie{
    //       Name: "_abck",
    //       Value: os.Getenv("ABCK"),
    //       Secure: true,
    //       HttpOnly: false,
    //       MaxAge: 1e6,
    //       Path: "/",
    //     // Domain: s.Task.BaseDomainUrl.Host,
    //    },
    //  },
    // }
    // s.Task.RemoveCookie("_abck")
    // s.Task.client.Jar.SetCookies(s.Task.BaseDomainUrl, s.Task.AbckJar[0])
    // s.Task.SetCookie(s.Task.AbckJar[0][0])

    s.Task.EnsureValidAbck(s.AkUrl, s.hasValidAbck)

    resp, err := s.DoReq(s.Task.makeReq("POST", url_, &headers, &headerOrder, &body), "Add to Cart")
    if resp != nil {
      if resp.Body != nil {
        io.Copy(ioutil.Discard, resp.Body)
        resp.Body.Close()
      }
    }
    if err != nil {
      return err
    }
    // if resp.Get()
    // for key, vals := range resp.Header {
    //   s.Task.LogDebug("%s=%+v", key, vals)
    // }

    // if !s.HasCart() {
    //   return errAtcFailed
    // }
    // s.Task.LogDebug("header %+v", resp.Header)
    loc = resp.Header.Get("location")
    if !strings.Contains(loc, "productDetail") {
    	return errAtcFailed
    }

    return nil
  }, retry.Attempts(3))

  if err != nil {
    return nil, err
  }

  err = s.FnlWebAtcGet(loc)
  if err != nil {
    return nil, err
  }
  return &cart, nil
}

func (s *FnlWebSession) FnlWebGetCheckoutShipping(shipUrl string) (*FnlWebProduct, [][2]string, error) {
  inps := [][2]string{}
  url_, err := url.Parse(shipUrl)
  if err != nil {
    return nil, inps, err
  }
  headerOrder := []string {
    "cache-control",
    "upgrade-insecure-requests",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-user",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }

  headers := [][2]string {
    {"cache-control", "max-age=0"},
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"referer", s.Task.Url.String()},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := s.DoReq(s.Task.makeReq("GET", url_, &headers, &headerOrder, nil), "Checkout (Shipping)")
  if err != nil {
    if resp != nil {
      DiscardResp(resp)
    }
    return nil, inps, err
  }
  if resp != nil {
    bodyBytes, err := readBodyBytes(resp)
   if err != nil {
     return nil, inps, err
   }

   doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
   if err != nil {
     return nil, inps, err
   }

   var hasshipmethod bool
   doc.Find("#shipping_form input").Each(func(_ int, s *goquery.Selection) {
    if name, ok := s.Attr("name"); ok {
      if name == "shippingMethod" && hasshipmethod {
        return
      }
     if value, ok := s.Attr("value"); ok {
       inps = append(inps, [2]string{name,value})
       if name == "shippingMethod" {
        hasshipmethod = true;
       }
     }
    }
   })

   return &FnlWebProduct{
    /*
    type FnlWebProduct struct {
      ColorId string
      Name string
      Pid string
      Skus []FnlWebSku
      StyleId string
      Color string
    }
    */
    Name: doc.Find("strong.productName").Text(),
    Color: doc.Find("span.colorDescription").Text(),
   }, inps, nil
	}
  return nil, inps, err
}

func (s *FnlWebSession) FnlWebShipping(shipUrl string) (*FnlWebProduct, string, error) {
  product, inps, err := s.FnlWebGetCheckoutShipping(shipUrl)
  if err != nil {
    return nil, "", err
  }

  s.Task.Config["stage"] = "Shipping"
  // s.Task.EnsureValidAbck(s.AkUrl, s.hasValidAbck)

  url_, err := url.Parse("https://" + s.Task.Url.Host + "/store/checkout/shipping.jsp?_DARGS=/store/checkout/shipping.jsp")
  if err != nil {
    return nil, "", err
  }
  headerOrder := []string {
    "content-length",
    "cache-control",
    "upgrade-insecure-requests",
    "origin",
    "content-type",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-user",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"cache-control", "max-age=0"},
    {"upgrade-insecure-requests", "1"},
    {"origin", "https://" + s.Task.Url.Host + ""},
    {"content-type", "application/x-www-form-urlencoded"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"referer", shipUrl},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }

  cparams := [][2]string {
    {"firstName", s.Task.Profile.ShippingAddress.FirstName},
    {"lastName", s.Task.Profile.ShippingAddress.LastName},
    {"address1", s.Task.Profile.ShippingAddress.Address1},
    {"address2", s.Task.Profile.ShippingAddress.Address2},
    {"city", s.Task.Profile.ShippingAddress.City},
    {"state", s.Task.Profile.ShippingAddress.State},
    {"zip", s.Task.Profile.ShippingAddress.Zip},
    {"phone", s.Task.Profile.ShippingAddress.HibPhone()},
    {"email", s.Task.Profile.Email},
    {"country", "US"},
    {"shippingMethod", "Economy"},
    {"receiveEmail", "false"},
    {"shipToAddressName", ""},
    {"copyNewAddrToBilling", "false"},
    {"upsValidation", "false"},
  }
  params := [][2]string{}
  for _, inp := range inps {
    key := inp[0]
    value := inp[1]
    s.Task.LogDebug("%s=%s", key, value)
    for _, p := range cparams {
      if p[0] == key {
        value = p[1]
        break
      }
    }
    params = append(params, [2]string{key, value})
  }
  for _, cp := range cparams {
    var pidx int = -1
    for idx, p := range params {
      if p[0] == cp[0] {
        pidx = idx
        break
      }
    }

    if pidx == -1 {
      var insIdx int
      for idx, p := range params {
        if p[0] == "_D:" + cp[0] {
          insIdx = idx
          break
        }
      }
      params = append(params[:insIdx], append([][2]string{ cp }, params[insIdx:]...)...)
    }
  }
  body := urlEncodeParams(params)
  resp, err := s.DoReq(s.Task.makeReq("POST", url_, &headers, &headerOrder, &body), "Shipping")
  if resp != nil {
	  io.Copy(ioutil.Discard, resp.Body)
	  resp.Body.Close()
	}
  if err != nil {
    return nil, "", err
  }
  if !strings.Contains(resp.Header.Get("Location"), "billing") {
    return nil, "", ErrRetrying
  }
  return product, resp.Header.Get("Location"), nil
}

func (s *FnlWebSession) FnlWebGetCheckoutBilling(billUrl string) ([][2]string, error) {
  inps := [][2]string{}
  url_, err := url.Parse(billUrl)
  if err != nil {
    return inps, err
  }
  headerOrder := []string {
    "cache-control",
    "upgrade-insecure-requests",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-user",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }

  headers := [][2]string {
    {"cache-control", "max-age=0"},
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"referer", s.Task.Url.String()},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := s.DoReq(s.Task.makeReq("GET", url_, &headers, &headerOrder, nil), "Checkout (Billing)")
  if err != nil {
    DiscardResp(resp)
    return inps, err
  }
  if resp != nil {
    bodyBytes, err := readBodyBytes(resp)
    if err != nil {
     return inps, err
    }

    doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
    if err != nil {
     return  inps, err
    }

    doc.Find("#billing_form input").Each(func(_ int, s *goquery.Selection) {
     if name, ok := s.Attr("name"); ok {
      if value, ok := s.Attr("value"); ok {
        inps = append(inps, [2]string{name,value})
      }
     }
    })
    return inps, nil
  }
  return inps, err
}

func (s *FnlWebSession) FnlWebBilling(billingUrl string) (string, error) {
  inps, err := s.FnlWebGetCheckoutBilling(billingUrl)
  if err != nil {
    return "", err
  }

  s.Task.Config["stage"] = "Billing"
  s.Task.EnsureValidAbck(s.AkUrl, s.hasValidAbck)

  url_, err := url.Parse("https://" + s.Task.Url.Host + "/store/checkout/billing.jsp?_DARGS=/store/checkout/billing.jsp")
  if err != nil {
    return "", err
  }
  headerOrder := []string {
    "content-length",
    "cache-control",
    "upgrade-insecure-requests",
    "origin",
    "content-type",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-user",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"cache-control", "max-age=0"},
    {"upgrade-insecure-requests", "1"},
    {"origin", "https://" + s.Task.Url.Host + ""},
    {"content-type", "application/x-www-form-urlencoded"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"referer", billingUrl},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }


  cardType := s.Task.Profile.Card.Type
  if cardType == "MASTER" {
  	cardType += "CARD"
  }


  cparams := [][2]string {
    {"creditCardNumberTemp", fmt.Sprintf("%s %s %s %s", s.Task.Profile.Card.Number[0:4], s.Task.Profile.Card.Number[4:8], s.Task.Profile.Card.Number[8:12], s.Task.Profile.Card.Number[12:])}, // todo split every 4 chars?
    {"expirationMonthTemp", fmt.Sprintf("%02d", s.Task.Profile.Card.ExpMonth)},
    {"expirationYearTemp", fmt.Sprintf("%02d", s.Task.Profile.Card.ExpYear)},
    {"securityCodeTemp", s.Task.Profile.Card.Cvc},
    {"firstName", s.Task.Profile.BillingAddress.FirstName},
    {"lastName", s.Task.Profile.BillingAddress.LastName},
    {"address1", s.Task.Profile.BillingAddress.Address1},
    {"address2", s.Task.Profile.BillingAddress.Address2},
    {"city", s.Task.Profile.BillingAddress.City},
    {"state", s.Task.Profile.BillingAddress.State},
    {"zip", s.Task.Profile.BillingAddress.Zip},
    {"phone", s.Task.Profile.BillingAddress.HibPhone()},
    {"email", s.Task.Profile.Email},
    {"country", s.Task.Profile.BillingAddress.Country},
    {"creditCardTypeTemp", cardType},
    {"transactionType", "CC"},
  }
  params := [][2]string{}
  for _, inp := range inps {
    key := inp[0]
    value := inp[1]
    s.Task.LogDebug("%s=%s", key, value)
    for _, p := range cparams {
      if p[0] == key {
        value = p[1]
        break
      }
    }

    params = append(params, [2]string{key, value})
  }
  for _, cp := range cparams {
    var pidx int = -1
    for idx, p := range params {
      if p[0] == cp[0] {
        pidx = idx
        break
      }
    }

    if pidx == -1 {
      var insIdx int
      for idx, p := range params {
        if p[0] == "_D:" + cp[0] {
          insIdx = idx
          break
        }
      }
      params = append(params[:insIdx], append([][2]string{ cp }, params[insIdx:]...)...)
    }
  }
  body := urlEncodeParams(params)
  resp, err := s.DoReq(s.Task.makeReq("POST", url_, &headers, &headerOrder, &body), "Billing")
  if resp != nil {
	  io.Copy(ioutil.Discard, resp.Body)
	  resp.Body.Close()
	}
  if err != nil {
    return "", err
  }
  if strings.Contains(resp.Header.Get("Location"), "billing") {
    return "", ErrRetrying
  }
  if !strings.Contains(resp.Header.Get("Location"), "review") {
    return "", ErrRetrying
  }
  // TODO check loc - billing is being rejected and redir to self
  return resp.Header.Get("Location"), nil
}

func (s *FnlWebSession) FnlWebGetCheckoutOrder(reviewUrl string) ([][2]string, error) {
  inps := [][2]string{}
  url_, err := url.Parse(reviewUrl)
  if err != nil {
    return inps, err
  }

  headerOrder := []string {
    "cache-control",
    "upgrade-insecure-requests",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-user",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }

  headers := [][2]string {
    {"cache-control", "max-age=0"},
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"referer", s.Task.Url.String()},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := s.DoReq(s.Task.makeReq("GET", url_, &headers, &headerOrder, nil), "Checkout (Ordering)")
  if err != nil {
    DiscardResp(resp)
    return inps, err
  }
  if resp != nil {
    bodyBytes, err := readBodyBytes(resp)
   if err != nil {
     return inps, err
   }

   doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
   if err != nil {
     return  inps, err
   }

    doc.Find("#review_form input").Each(func(_ int, s *goquery.Selection) {
     if name, ok := s.Attr("name"); ok {
      if value, ok := s.Attr("value"); ok {
        inps = append(inps, [2]string{name,value})
      }
     }
    })
    return inps, nil
  }
  return inps, err
}


func (s *FnlWebSession) FnlWebOrder(reviewUrl string) (error) {
  return s.Task.Retry(func() error {
    inps, err := s.FnlWebGetCheckoutOrder(reviewUrl)
    if err != nil {
      return err
    }
    s.Task.Config["stage"] = "Order"

    s.Task.EnsureValidAbck(s.AkUrl, s.hasValidAbck)
    var rctoken string
    if s.EnableRecap {
      rctoken, err = s.FnlLiveGenToken("CONFIRM_EVENT")
      if err != nil {
        return err
      }
    }

    url_, err := url.Parse(reviewUrl)
    if err != nil {
      return err
    }
    url_.RawQuery = "_DARGS=/store/checkout/review.jsp"

    if rctoken == "" {
      rctoken = "default"
    }
    pv := map[string]string {
      "submitOrderExecutionToken": rctoken,
    }
    qparams := [][2]string{}

    for _, inp := range inps {
      key := inp[0]
      value := inp[1]
      if pv[key] != "" {
        value = pv[key]
      }
      qparams = append(qparams, [2]string{key, value})
    }

    // q := url_.Query()
    // for _, qp := range qparams {
    //   q.Add(qp[0], qp[1])
    // }
    // url_.RawQuery = q.Encode()
    headerOrder := []string {
      "accept",
      "user-agent",
      "x-requested-with",
      "sec-fetch-site",
      "sec-fetch-mode",
      "sec-fetch-dest",
      "referer",
      "accept-encoding",
      "accept-language",
      "cookie",
    }
    headers := [][2]string {
      {"cache-control", "max-age=0"},
      {"upgrade-insecure-requests", "1"},
      {"origin", "https://" + s.Task.Url.Host + ""},
      {"content-type", "application/x-www-form-urlencoded"},
      {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
      {"sec-fetch-site", "same-origin"},
      {"sec-fetch-mode", "navigate"},
      {"sec-fetch-user", "?1"},
      {"sec-fetch-dest", "document"},
      {"referer", reviewUrl},
      {"accept-encoding", "gzip, deflate, br"},
      {"accept-language", "en-US,en;q=0.9"},
    }
    body := urlEncodeParams(qparams)
    resp, err := s.DoReq(s.Task.makeReq("POST", url_, &headers, &headerOrder, &body), "Ordering")


    if resp != nil {
  	  io.Copy(ioutil.Discard, resp.Body)
  	  resp.Body.Close()

    	loc := resp.Header.Get("Location")
    	if strings.Contains(loc, "checkout/billing") {
        // return errCardDeclined

        // todo fix, getting 403 on this get
        urlll, err := url.Parse(loc)
        if err != nil {
          return err
        }
        headerOrder := []string {
          "upgrade-insecure-requests",
          "user-agent",
          "accept",
          "sec-fetch-site",
          "sec-fetch-mode",
          "sec-fetch-user",
          "sec-fetch-dest",
          "referer",
          "accept-encoding",
          "accept-language",
          "cookie",
        }
        headers := [][2]string {
          {"upgrade-insecure-requests", "1"},
          {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
          {"sec-fetch-site", "same-origin"},
          {"sec-fetch-mode", "navigate"},
          {"sec-fetch-user", "?1"},
          {"sec-fetch-dest", "document"},
          {"accept-encoding", "gzip, deflate, br"},
          {"accept-language", "en-US,en;q=0.9"},
        }
        cresp, err := s.DoReq(s.Task.makeReq("GET", urlll, &headers, &headerOrder, nil), "Check Order Status")
        if err != nil {
          DiscardResp(cresp)
          return err
        }
        if cresp != nil {
          body, _ := readBody(cresp)
          if strings.Contains(body, "There Was an Issue Processing Your Order") || strings.Contains(body, "Your transaction could not be authorized.") {
            s.Task.StatusCh <- StatusCardDeclined
      		  return retry.Unrecoverable(errCardDeclined)
          } else if strings.Contains(body, "One or more items in your cart are no longer available.") {
            s.Task.StatusCh <- StatusItemRemovedFromCart
            return retry.Unrecoverable(errUnknownCheckoutResult)
          } else {
            // oos on order?
            go s.Task.SendTelemetry(map[string]interface{} {
              "event": "fnl_unexpected_billing_response",
              "response": map[string]interface{} {
                // "headers": todo
                "location": loc,
                "post_status_code": resp.StatusCode,
                "get_status_code": cresp.StatusCode,
                "body": body,
              },
            })
            return errUnknownCheckoutResult
          }
        } else {
          return errUnknownCheckoutResult
        }
    	} else if strings.Contains(loc, "checkout/confirm") {
        go s.Task.SendTelemetry(map[string]interface{} {
          "event": "fnl_success_checkout_response",
          "response": map[string]interface{} {
            // "headers": todo
            "location": loc,
            "post_status_code": resp.StatusCode,
          },
        })
        _, err := s.FnlWebGetCheckoutOrder(loc)
        return err
    	} else {
        go s.Task.SendTelemetry(map[string]interface{} {
          "event": "fnl_unexpected_checkout_response",
          "response": map[string]interface{} {
            // "headers": todo
            "location": loc,
            "post_status_code": resp.StatusCode,
          },
        })
      }
      return errUnknownCheckoutResult
    }
    return errUnknownCheckoutResult
  }, retry.Attempts(3))
}

// func (s *FnlWebSession) FnlWebOrderStatus() (error) {
//   url_, err := url.Parse("https://" + s.Task.Url.Host + "/store/checkout/billing.jsp?_requestid=498215")
//   if err != nil {
//     return err
//   }
//   headerOrder := []string {
//     "cache-control",
//     "upgrade-insecure-requests",
//     "user-agent",
//     "accept",
//     "sec-fetch-site",
//     "sec-fetch-mode",
//     "sec-fetch-user",
//     "sec-fetch-dest",
//     "referer",
//     "accept-encoding",
//     "accept-language",
//     "cookie",
//   }
//   headers := [][2]string {
//     {"cache-control", "max-age=0"},
//     {"upgrade-insecure-requests", "1"},
//     {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
//     {"sec-fetch-site", "same-origin"},
//     {"sec-fetch-mode", "navigate"},
//     {"sec-fetch-user", "?1"},
//     {"sec-fetch-dest", "document"},
//     {"referer", "https://" + s.Task.Url.Host + "/store/checkout/review.jsp?_requestid=498203"},
//     {"accept-encoding", "gzip, deflate, br"},
//     {"accept-language", "en-US,en;q=0.9"},
//   }
//   resp, err := s.Task.doReq(s.Task.client, s.Task.makeReq("GET", url_, &headers, &headerOrder, nil))
//   if err != nil {
//     return err
//   }
//   loc := resp.Header.Get("location")
//   if strings.Contains(loc, "checkout/billing") {
//   	return errCardDeclined
//   } // else if strings.Contains(loc, )

//   // body, err := readBody(resp)
//   // if err != nil {
//   //   return err
//   // }

//   return nil
// }

func (s *FnlWebSession) FnlWebGetProductMiniCart() (error) {
  url_, err := url.Parse("https://" + s.Task.Url.Host + "/store/cart/gadgets/miniCart.jsp")
  if err != nil {
    return err
  }
  headerOrder := []string {
    "accept",
    "user-agent",
    "x-requested-with",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "text/html, */*; q=0.01"},
    {"x-requested-with", "XMLHttpRequest"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", s.Task.Url.String()},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := s.DoReq(s.Task.makeReq("GET", url_, &headers, &headerOrder, nil), "Get Cart")
  if resp != nil {
	  io.Copy(ioutil.Discard, resp.Body)
	  resp.Body.Close()
	}
  if err != nil {
    return err
  }
  return nil
}

func (s *FnlWebSession) FnlWebCheckoutLogin(loginUrl string) (string, error) {
  url_, err := url.Parse(loginUrl)
  if err != nil {
    return "", err
  }
  headerOrder := []string {
    "cache-control",
    "upgrade-insecure-requests",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-user",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"cache-control", "max-age=0"},
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"referer", s.Task.Url.String()},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := s.DoReq(s.Task.makeReq("GET", url_, &headers, &headerOrder, nil), "Starting Checkout")
  if resp != nil {
	  io.Copy(ioutil.Discard, resp.Body)
	  resp.Body.Close()
	}
  if err != nil {
    return "", err
  }
  if !strings.Contains(resp.Header.Get("Location"), "shipping.jsp") {
    return "", ErrRetrying
  }
  return resp.Header.Get("Location"), nil
}

func (s *FnlWebSession) FnlWebCartSlide(skuId string) (string, error) {
  s.Task.StatusCh <- Status{Message: "Starting Checkout (1/3)"}

  url_, err := url.Parse("https://" + s.Task.Url.Host + "/store/cart/cartSlide.jsp?stage=pdp")
  if err != nil {
    return "", err
  }
  headerOrder := []string {
    "accept",
    "user-agent",
    "x-requested-with",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "text/html, */*; q=0.01"},
    {"x-requested-with", "XMLHttpRequest"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", s.Task.Url.String()},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
	resp, err := s.DoReq(s.Task.makeReq("GET", url_, &headers, &headerOrder, nil), "Starting Checkout (1/3)")
	if err != nil {
		return "", err
	}
	rbody, err := readBody(resp)
  re := regexp.MustCompile(`name="_dynSessConf" value="(.+?)"`)
  matches := re.FindStringSubmatch(rbody)
  if len(matches) != 2 {
  	return "", errAtcFailed
  }
  csrf := matches[1]

  err = s.FnlWebGetProductMiniCart()
  if err != nil {
    return csrf, err
  }

  var rctoken string
  if s.EnableRecap {
    rctoken, err = s.FnlLiveGenToken("MOVE_TO_CHECKOUT_EVENT")
    if err != nil {
      return csrf, err
    }
  }
  s.Task.StatusCh <- Status{Message: "Starting Checkout (2/3)"}

  url_, err = url.Parse("https://" + s.Task.Url.Host + "/store/cart/cartSlide.jsp?_DARGS=/store/cart/cartInner.jsp.cartform")
  if err != nil {
    return  csrf, err
  }
  headerOrder = []string {
    "content-length",
    "cache-control",
    "upgrade-insecure-requests",
    "origin",
    "content-type",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-user",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers = [][2]string {
    {"cache-control", "max-age=0"},
    {"upgrade-insecure-requests", "1"},
    {"origin", "https://" + s.Task.Url.Host + ""},
    {"content-type", "application/x-www-form-urlencoded"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"referer", s.Task.Url.String()},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  params := [][2]string {
    {"_dyncharset", "UTF-8"},
    {"_dynSessConf", csrf},
    {skuId, "1"},
    {"/atg/commerce/order/purchase/CartModifierFormHandler.shippingInfoURL", "/store/checkout/shipping.jsp"},
    {"_D:/atg/commerce/order/purchase/CartModifierFormHandler.shippingInfoURL", " "},
    {"/atg/commerce/order/purchase/CartModifierFormHandler.moveToPurchaseInfoErrorURL", "/store/cart/cart.jsp"},
    {"_D:/atg/commerce/order/purchase/CartModifierFormHandler.moveToPurchaseInfoErrorURL", " "},
    {"/atg/commerce/order/purchase/CartModifierFormHandler.loginDuringCheckoutURL", "/store/checkout/login.jsp"},
    {"_D:/atg/commerce/order/purchase/CartModifierFormHandler.loginDuringCheckoutURL", " "},
    {"/atg/commerce/order/purchase/CartModifierFormHandler.moveToPurchaseInfo", "Submit"},
    {"_D:/atg/commerce/order/purchase/CartModifierFormHandler.moveToPurchaseInfo", " "},
    {"cartFormExecutionToken", rctoken},
    {"_D:PayPal", " "},
    {"_DARGS", "/store/cart/cartInner.jsp.cartform"},
  }
  body := urlEncodeParams(params)
  resp, err = s.DoReq(s.Task.makeReq("POST", url_, &headers, &headerOrder, &body), "Start Checkout")
  if resp != nil {
	  io.Copy(ioutil.Discard, resp.Body)
	  resp.Body.Close()
	}
  if err != nil {
    return  "", err
  }
  if !strings.Contains(resp.Header.Get("location"), "login.jsp") {
    return "", ErrStartCheckoutFailed
  }
  return  resp.Header.Get("location"), nil
}

func (s *FnlWebSession) hasValidAbck(abck string) bool {
  // return true

	if abck == "" || len(abck) < 64 {
		return false
	}
  s.abckRetries += 1
  if s.abckRetries == uint(3+rand.Intn(4)) {
    s.abckRetries = 0
    return true
  } else {
    return false
  	// return (s.Task.Url.Host == "www.finishline.com" && strings.HasSuffix(abck, "~-1~-1~-1") && !strings.HasSuffix(abck, "=~-1~-1~-1")) || (s.Task.Url.Host == "www.jdsports.com" && !strings.HasSuffix(abck, "=~-1~-1~-1") && strings.HasSuffix(abck, "~-1~-1~-1"))
  }
}

func (s *FnlWebSession) RefreshConfig() {
  if bb, err := s.Task.S3Get("fnl_web_config.json", 10); err == nil {
    var config map[string]interface{}
    json.Unmarshal(bb, &config)
    if enable_recap, ok := config["enable_recap"].(bool); ok {
      s.EnableRecap = enable_recap
    }
  }
  s.EnableRecap = (s.EnableRecap || os.Getenv("RECAP") == "1") && os.Getenv("RECAP") != "0"
}

func (s *FnlWebSession) FnlWebAtcGet(urlstr string) (error) {
  url_, err := url.Parse(urlstr)
  if err != nil {
    return err
  }
  headerOrder := []string {
    "accept",
    "x-requested-with",
    "user-agent",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "*/*"},
    {"x-requested-with", "XMLHttpRequest"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", s.Task.Url.String()},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  return s.Task.Retry(func() error {
    req := s.Task.makeReq("GET", url_, &headers, &headerOrder, nil)
    req.IgnoreBody = true
    resp, err := s.DoReq(req, "Get Cart")
    DiscardResp(resp)
    return err
  }, retry.Attempts(3))
}

func (s *FnlWebSession) Checkout() error {
  desiredProduct, err := s.Task.FnlEnsureUrlHasProductSpecificity()
  if err != nil {
    return err
  }
  s.Task.AbckJarMaxUsage = 1

  defer s.Task.PhoneHomeUnusedAbck()
  return s.Task.RetryTask(func() error {
    s.Task.StatusCh <- Status{Message: "Initializing"}
    s.Task.initClient()
    s.RefreshConfig()

    var err error
    // s.Task.UserAgent = "Mozilla/5.0 (Windows NT 6.2; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.92 Safari/537.36"
    s.Task.UserAgent, err = s.Task.getUserAgent()
    if err != nil {
      return err
    }


    s.Task.StatusCh <- StatusStartingSession
    err = s.StartSession()
    if err != nil {
      return err
    }


    var sku *FnlWebSku
    err = s.Task.Retry(func() error {
      s.Task.StatusCh <- Status{Message: "Waiting for Stock"}
      pproduct, err := s.GetProduct(desiredProduct)
      if err != nil {
        return err
      }
      s.Task.LogDebug("%+v", pproduct)

      rand.Shuffle(len(pproduct.Skus), func(i, j int) {
        pproduct.Skus[i], pproduct.Skus[j] = pproduct.Skus[j], pproduct.Skus[i] })

      for _, _sku := range pproduct.Skus {
        if len(s.Task.Sizes) == 0 || Any(s.Task.Sizes, func(size string) bool { return size == strings.ToLower(_sku.Size) || size + ".0" == strings.ToLower(_sku.Size) }) {
          sku = &_sku
          break
        }
      }
      if sku == nil {
        // s.Task.StatusCh <- StatusOutOfStock
        return errOutOfStock
      }
      return nil
    }, retry.Attempts(128))
    if err != nil {
      return err
    }
    // err = s.GetHome()
    // if err != nil {
    //   return err
    // }



    // TODO pregen before visit product ?
    // TODO queue bypass to gen?
    // if false && os.Getenv("NO_PREGEN") != "1" && (rand.Float32() < 0.5 || os.Getenv("PREGEN") == "1") {
    //   s.Task.TelemetryContext["fnl_pregen"] = "1"
    //   cancel := s.Task.PopulateAbckJar(6, s.hasValidAbck) // todo api parameterize / higher for initial drop
    //   defer cancel()
    //   s.Task.WaitForAbck(3*time.Second)

    //   // TODO fix abck (4)03s
    //   // time.Sleep(2*time.Second)
    //   // TODO wait for abck
    // }

    // // s.Task.StatusCh <- Status{Message: "Visit Homepage"}
    // // product := &FnlWebProduct{}
    // // if err != nil {
    //   return err
    // }

    s.RefreshConfig()


    if os.Getenv("ABCK_UPFRONT") == "1" {
      s.Task.AbckGenLoop(s.AkUrl, s.hasValidAbck)
    }


    // err = s.Task.EnsureValidAbck(s.AkUrl, s.hasValidAbck)
    // if err != nil {
    //  return err
    // }

    // TODO cache ?

    // err = s.Task.EnsureValidAbck(s.AkUrl, s.hasValidAbck)
    // if err != nil {
    //  return err
    // }

    // err = s.FnlWebGetProductMiniCart()
    // if err != nil {
    //   return err
    // }

    return s.Task.Retry(func() error {
      s.Task.StartAtcTime = timeMillis()

      s.Task.StatusCh <- Status{Message: "Adding to Cart"}
      s.Task.RemoveCookie("fl_shoppingcart_items")
      _, err = s.FnlWebAtc(desiredProduct, sku)

      if err != nil {
        return err
      }

      s.Task.StartCheckoutTime = timeMillis()

      return s.Task.Retry(func() error {
        if !s.HasCart() {
          return retry.Unrecoverable(ErrCartJacked)
        }

        loginUrl, err := s.FnlWebCartSlide(sku.SkuId)
        if err != nil {
          return err
        }

        if !s.HasCart() {
          return retry.Unrecoverable(ErrCartJacked)
        }

        s.Task.StatusCh <- Status{Message: "Starting Checkout (3/3)"}
        shipUrl, err := s.FnlWebCheckoutLogin(loginUrl)
        if err != nil {
          return err
        }

        if !s.HasCart() {
          return retry.Unrecoverable(ErrCartJacked)
        }

        s.Task.StatusCh <- Status{Message: "Checkout 1/3: Shipping"}
        product, billingUrl, err := s.FnlWebShipping(shipUrl)
        if err != nil {
          return err
        }

        if !s.HasCart() {
          return retry.Unrecoverable(ErrCartJacked)
        }

        s.Task.EnsureValidAbck(s.AkUrl, s.hasValidAbck)
        s.Task.StatusCh <- Status{Message: "Checkout 2/3: Billing"}
        reviewUrl, err := s.FnlWebBilling(billingUrl)
        if err != nil {
          return err
        }

        if !s.HasCart(){
          return retry.Unrecoverable(ErrCartJacked)
        }

        s.Task.StartPaymentTime = timeMillis()
        s.Task.StatusCh <- Status{Message: "Checkout 3/3: Order"}
        err = s.FnlWebOrder(reviewUrl)
        whFields := [][2]string{
          {"Website", s.Task.Url.Host + " (DESKTOP)"},
          {"Product", product.Name},
          {"Size", sku.Size},
          {"Color", product.Color},
        }
        if errors.Is(err, errCardDeclined) {
          s.Task.StatusCh <- StatusCardDeclined
          go s.Task.DeclineWebhookFields(whFields)
          return retry.Unrecoverable(retry.Unrecoverable(retry.Unrecoverable(err)))
        } else if err == nil {
          s.Task.StatusCh <- StatusCooked
          go s.Task.SuccessWebhookFields(whFields)
        }
        s.Task.LogDebug("%+v", err)
        // err = s.Task.FnlWebOrderStatus()
        return err
      })
    }, retry.Attempts(8))
  }, retry.DelayType(retry.DefaultDelayType), retry.MaxDelay(33*time.Second), retry.Delay(2*time.Second), retry.LastErrorOnly(true),
  retry.OnRetry(func(_ uint, err error) {
    s.Task.client.CloseIdleConnections()
    CloseH2Conns(s.Task.client)
  }))
}

func (t *CheckoutTask) FnlWebCheckout() error {
  sesh := FnlWebSession{Task: t}

	return sesh.Checkout()
}

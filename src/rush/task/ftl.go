package task

import (
  "os"
  "context"
  "sort"
  "encoding/json"
  "fmt"
  "github.com/google/uuid"
  // "rush/common"
  "github.com/pkg/errors"
  // "io"
  "math/rand"
  "net/url"
  "regexp"
  "rush/net/http"
  "strconv"
  "strings"
  "bytes"
  "github.com/avast/retry-go"
  "io/ioutil"
  "time"
)

var (
	ftlErrOutOfStock = errors.New("Out of Stock")
  errInvalidCard = errors.New("Card Rejected (Invalid)")
  errUnexpectedCheckoutError = errors.New("Checkout Error")
)
var FTL_SITEKEY = "6LccSjEUAAAAANCPhaM2c-WiRxCZ5CzsjR_vd8uX"
var LOW_DELAY_TOUT = 5*time.Minute

func (t *CheckoutTask) FtlCheckOrder(orderId string, email string) ()  {
  sesh, _ := t.FtlGetSession()

  req, err := t.MakeChlsRequest(FTL_ORDERCHK_CHLS, map[*regexp.Regexp]string {
    regexp.MustCompile(`V2001624425`): orderId,
    regexp.MustCompile(`HeatherJackson1@recsemail.com`): email,
    regexp.MustCompile(`79e6d93c-1d4c-4088-8732-d9c9e82aefb2`): sesh.Data.CsrfToken,
  })
  req.Request.Close = true
  t.RemoveCookie("datadome")
  resp, err := t.FtlWebDoRequest(req.Request, "Order Check")
  if err != nil {
    return
  }
  if b, _ := readBody(resp); len(b)>0 {
    fmt.Println(b)
  }
}

func (t *CheckoutTask) FtlGetSession() (*FtlSession, error) {
  url_, err := url.Parse(fmt.Sprintf(t.FtlBase() + "api/v3/session?timestamp=%d", timeMillis()))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "accept",
    "user-agent",
    "x-fl-request-id",
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
    {"x-fl-request-id", uuid.New().String()},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", t.FtlBase() + ""},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }

  resp, err := t.FtlWebDoRequest(t.makeReq("GET", url_, &headers, &headerOrder, nil), "Get Session")
  if err != nil {
  	return nil, err
  }
  var ftlSession FtlSession
  if err := readRespJsonDst(resp, &ftlSession); err != nil {
  	return nil, err
  }
  ftlSession.Time = time.Now()
  return &ftlSession, nil
}


func (t *CheckoutTask) FtlGetProductApi() (*FtlProductPdpResponse, error) {
  // t.FastlyJig()
  var pdpResp FtlProductPdpResponse
  // t.DefaultReqClose = true
  err := t.Retry(func() error {
    pSplit := strings.Split(t.Url.Path, "/")
    pid := strings.Split(pSplit[len(pSplit) - 1], ".")[0]

    // apiUrl, err := url.Parse(fmt.Sprintf("%sapi/products/pdp/%s?tsz=%d&products=%s", t.FtlBase(), pid,  timeMillis(), pid))
    apiUrl, err := url.Parse(fmt.Sprintf("%sapi/products/pdp/%s?product=%s", t.FtlBase(), pid, pid))
    if err != nil {
      return err
    }
    // if rand.Float32() < 0.5 {
    //   apiUrl.Path = strings.Replace(apiUrl.Path, "api", "api", -1)
    //   t.SetTelemetryContext("stock_path", apiUrl.Path)
    // }
    // fmt.Println(apiUrl.String())

    headerOrder := []string {
      "accept",
      "user-agent",
      "x-fl-request-id",
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
      {"origin", t.FtlOrigin()},
      {"pragma", "no-cache"},
      {"cache-control", "no-cache"},
      {"sec-fetch-site", "same-origin"},
      {"sec-fetch-mode", "cors"},
      {"sec-fetch-dest", "empty"},
      {"referer", t.Url.String()},
      {"accept-encoding", "gzip, deflate, br"},
      {"accept-language", "en-US,en;q=0.9"},
    }

    if !t.UseIpv6 {
      t.DnsCacheMut.Lock()
      t.DnsCache = nil
      t.DnsCacheMut.Unlock()
    }
    t.resetClient(false)

    resp, err := t.FtlWebDoRequest(t.makeReq("GET", apiUrl, &headers, &headerOrder, nil), "Stock Check")
    if err != nil {
      DiscardResp(resp)
      return err
    }

    err = readRespJsonDst(resp, &pdpResp)
    if err != nil {
      DiscardResp(resp)
      return err
    }
    if len(pdpResp.SellableUnits) == 0 {
      return ErrInvalidProductResponse
    }

    if os.Getenv("DISCOVER") == "1" {
      t.LogDebug("x-cache=%s,%v", resp.Header.Get("x-cache"), t.DnsCacheActiveHost)
      t.LogDebug("x-served-by=%s", resp.Header.Get("x-served-by"))
    }

    var code string
    for _, variant := range pdpResp.VariantAttributes[:] {
      if variant.Sku == pid {
        code = variant.Code
        break
      }
    }
    if code == "" {
      return ErrProductUnavailable
    }

    for _, opt := range pdpResp.SellableUnits[:] {
      for _, attr := range opt.Attributes[:] {
        if attr.Type == "style" && attr.ID == code {
          opt.Name = pdpResp.Name
          pdpResp.Skus = append(pdpResp.Skus, opt)
        }
      }
    }

    return nil
  }, retry.Delay(t.Delay), retry.Attempts(1))
  // fmt.Printf("ret %+v %+v\n", skus,err)

  return &pdpResp,err
}

func (t *CheckoutTask) FtlGetHomepage() error {
  headerOrder := []string {
    "upgrade-insecure-requests",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "`accept-encoding`",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "none"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-dest", "document"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }

  req := t.makeReq("GET", t.BaseUrl, &headers, &headerOrder, nil)
  req.Close = false
  req.IgnoreBody = true
  resp, err := t.FtlWebDoRequest(req, "Visit Homepage")
   DiscardResp(resp)

  return err
}

var prodPageRe = regexp.MustCompile(`window.footlocker.STATE_FROM_SERVER = (.+?});`)

func (t *CheckoutTask) FtlGetProduct() (*FtlProductPageState, error) {
  var productResp FtlProductPageState

  pSplit := strings.Split(t.Url.Path, "/")
  pid := strings.Split(pSplit[len(pSplit) - 1], ".")[0]
  ret := 0
  t.StatusCh <- StatusWaitStock
  err := t.Retry(func() error {
    defer func() {
      ret += 1
    }()
    if t.GetCookieValue("waiting_room") == "" {
      cache := t.GetTelemetryContext("cache")
      if cache == "1" || (cache != "0" && (rand.Float32() > 0.2)) {
        if body, err := t.ApiGetRetries(fmt.Sprintf("/hcache?k=pdp_%s", pid), 2); err == nil {
          // fmt.Println(string(body))
          if err := json.Unmarshal(body, &productResp); err == nil {
            t.LogDebug("USING CACHED PRODUCT %+v", productResp)
            return nil
          }
        }
      }
    }
    headerOrder := []string {
      "upgrade-insecure-requests",
      "user-agent",
      "accept",
      "sec-fetch-site",
      "sec-fetch-mode",
      "sec-fetch-dest",
      "accept-encoding",
      "accept-language",
      "cookie",
    }
    headers := [][2]string {
      {"upgrade-insecure-requests", "1"},
      {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
      {"sec-fetch-site", "none"},
      {"sec-fetch-mode", "navigate"},
      {"sec-fetch-dest", "document"},
      {"accept-encoding", "gzip, deflate, br"},
      {"accept-language", "en-US,en;q=0.9"},
    }

    // pSplit := strings.Split(t.Url.Path, "/")
    // prodpage := pSplit[len(pSplit) - 1]

    // url_, err := url.Parse(fmt.Sprintf("%sproduct/~/%s", t.FtlBase(), prodpage))
    req := t.makeReq("GET", t.Url, &headers, &headerOrder, nil)
    req.IgnoreBody = true
    resp, err := t.FtlWebDoRequest(req, "Get Product")
    DiscardResp(resp)

    // fmt.Println(fmt.Sprintf("ERR  %v", err))
    if err != nil {
      if errors.Is(err, errInQueue) {
      }
      if serr, ok := err.(*StatusCodeRejectedError); ok {
        return retry.Unrecoverable(serr)
      }
      if !strings.Contains(fmt.Sprintf("%v", err), "body close") {
        t.SetTelemetryContext("jig", "0")
      	return err
      }
      return err
    }
    t.StatusCh <- StatusWaitStock


    if pp, err := t.FtlGetProductApi(); err == nil {
      sizes := []FtlPdpSize{}
      for _, sku := range pp.Skus {
        sizes = append(sizes, FtlPdpSize{
          Name: sku.Attributes[0].Value,
          Code: sku.Code,
          IsDisabled: sku.StockLevelStatus != "inStock",
        })
      }

      productResp.Details.Sizes = map[string][]FtlPdpSize{
        "s": sizes,
      }
      productResp.Details.Product = map[string]FtlProductPageProductSummary {"s":FtlProductPageProductSummary{Name: pp.Name}}
      var style string
      var styleSku string
      if len(pp.Skus[0].Attributes) > 1 {
        style = pp.Skus[0].Attributes[1].Value
        styleSku = pp.Skus[0].Attributes[1].ID
      }
      sp := FtlProductPageSelectedProduct{Style: style, StyleSku: styleSku}
      if sld := (pp.ReleaseTime().Sub(time.Unix(0,0))).Milliseconds(); sld > 1607769811000 {
        sp.Launch.SkuLaunchDate = sld
      }
      productResp.Details.Selected = map[string]FtlProductPageSelectedProduct{"s": sp}
      if len(*(productResp.Sizes())) > 0 {
        go func() {
          defer func() {
            recover()
          }()
          if pj, err := json.Marshal(productResp); err == nil {
            t.ApiPost(fmt.Sprintf("/hcache?k=pdp_%s", pid), pj)
            // if styleSku != pid {
            //   t.ApiPost(fmt.Sprintf("/hcache?k=pdp_%s", styleSku), pj)
            // }
          }
        }()
      }
      return nil
    }
      // time.Sleep(t.Delay)
  	return errors.New("Missing Product Details")
    // var ss map[string]interface{}
    // json.Unmarshal([]byte(matches[1]), &ss)
    // sse, _ := json.MarshalIndent(ss, "", "  ")
    // fmt.Println(string(sse))
  }, retry.Attempts(1e6), retry.Delay(t.Delay / 4), retry.DelayType(retry.FixedDelay), retry.LastErrorOnly(true))

  return &productResp, err
}

func (t *CheckoutTask) FtlBase() string {
  return "https://" + t.Url.Host + "/"
}

func (t *CheckoutTask) FtlOrigin() string {
  return "https://" + t.Url.Host
}

func prettyPrint(i interface{}) string {
    s, _ := json.MarshalIndent(i, "", "\t")
    return string(s)
}

func (t *CheckoutTask) FtlWebDoRequest(req *http.Request, stage string) (*http.Response, error) {
  // if req.RawHeader != nil {
  //   req.RawHeader = append(req.RawHeader, FtlStaticHeaders...)
  // }
  // quiet := t.Config["capquiet"] == "1"
  quiet := false
  var err error
  var resp *http.Response

  // if t.Config["last_valid_abck"] != "" {
  //   t.SetCookie(&http.Cookie{ Name: "_abck", Value: t.Config["last_valid_abck"], Path: "/", })
  //   t.LogDebug("using last_valid_abck: %s (%s)", t.GetCookieValue("_abck"), t.Config["last_valid_abck"])

  //   // for _, ck := range t.client.Jar.Cookies(req.URL) {
  //   //   if ck.Name == "_abck" {
  //   //     fmt.Printf("using cached abck %s\n", prettyPrint(ck))
  //   //   }
  //   // }
  // }

  // req.RawHeader = append(req.RawHeader, [2]string{"Fastly-Debug", randHex(8)})
  // req.RawHeader = append(req.RawHeader, [2]string{"Fastly-No-Shield", randHex(8)})

  // req.RawHeader = append(req.RawHeader, [2]string{"auth-user-id", })
  // req.RawHeader = append(req.RawHeader, [2]string{"fastly-force-shield", "0"})
  // req.RawHeader = append(req.RawHeader, [2]string{"fastly-force-shield", "0"})

  if t.UseIpv6 || rand.Float32() < 1 {
    // orh := req.RawHeader
    spp := ""
    for i := 0; i < rand.Intn(32); i++ {
      spp += " "
    }

    // req.RawHeader = append(req.RawHeader, [2]string{"Accept-Encoding", fmt.Sprintf("br, compress, deflate, identity, %s", spp)})
    // req.RawHeader = append(req.RawHeader, [2]string{"Accept-Encoding", fmt.Sprintf("identity")})
    // req.RawHeader = append(req.RawHeader, [2]string{"Fastly-Client-IP", randHex(8)})
    // req.RawHeader = append(req.RawHeader, [2]string{"Vary", "Accept-Encoding, accept-encoding"})
    // req.RawHeader = append(req.RawHeader, [2]string{"Vary", "accept-encoding, Accept-Encoding"})
    // req.RawHeader = append(req.RawHeader, [2]string{"Vary", "Fastly-Debug-Digest"})//, Cookie, Date, Age, Expires, X-Cache-Hits, X-Served-By, " + strings.ToLower("Accept-Encoding, Fastly-Debug-Digest, Cookie, Date, Age, Expires, X-Cache-Hits, X-Served-By")})
    // req.RawHeader = append(req.RawHeader, [2]string{"Vary", "fastly-debug-digest"})

    // req.RawHeader = append(req.RawHeader, [2]string{"Cache-Control", randHex(8)})
    // req.RawHeader = append(req.RawHeader, [2]string{"Surrogate-Control", randHex(8)})
    // req.RawHeader = append(req.RawHeader, [2]string{"", randHex(8)})
    // req.RawHeader = append(req.RawHeader, [2]string{"Surrogate-Key", randHex(8)})
    // t.HeaderBlacklist["accept-encoding"] = true
  }
  t.LogDebug("%+v", req.RawHeader)


  if !strings.Contains(req.URL.String(), "carts/current?") {
    // t.FtlGetCart()
  }
  if !req.IgnoreBody {
    req.IgnoreBody = true
    req.IgnoreBodyExceptStatus = []int{200,201,403,400}
  }

  if req.Method != "GET" {
    if t.GetTelemetryContext("low_data") == "1" {
      // TODO wait time should be function of total tasks running on product

      wait := time.Duration(30000 + rand.Intn(777777)) * time.Millisecond
      select {
      case <-time.After(wait):
        break
      case <-t.WakeUp:
        t.Metrics.Incr("wakeups")
        t.SetTelemetryContext("last_wakeup", time.Now().Unix())
        t.LogDebug("WAKEUP")
        break
      }
    }
  }

  var retries = 0
  // var backoff = 1
  var solvedCap = false

  err = t.Retry(func() error {
    if resp != nil {
      DiscardResp(resp)
    }
    // t.FastlyJig()
    defer func() {
      retries += 1
    }()
    // reqabck := t.GetCookieValue("_abck")
    t.LogDebug(req.URL.String())
    // req.RawHeader = append(req.RawHeader, [2]string{"Fastly-Client-IP", "174.193.130.100"})
    // req.RawHeader = append(req.RawHeader, [2]string{"X-Forwarded-For", "174.193.130.100"})
    resp, err = t.doReq(t.client, req)
    t.LogDebug("%+v %+v", resp, err)
    if resp != nil {
      cap_cookie := t.GetCookieValue("datadome")



      // if strings.ToLower(resp.Header.Get("server")) == "akamaighost" && strings.Contains(bodystr, "<h1>access denied</h1>") {
      //   return TaskErrUnrecov
      //erable(ErrAkamaiBanned)
      // }
      body, berr := ioutil.ReadAll(resp.Body)
      // if err != nil && !req.IgnoreBody {
      if berr != nil && !req.IgnoreBody {
        return berr
      }
      bodystr := string(body)
      resp.Body = ioutil.NopCloser(bytes.NewReader(body))
      if solvedCap && resp.StatusCode != 403 {
        t.Metrics.Incr("dd:cap_pass")
        go func() {
          defer func() {
                    recover()
                  }()
          t.ApiPost("/sesh/pushcap", []byte(cap_cookie))
        }()
        // go func() {
        //   defer recover()
        //   // t.FtlSendSessionSnapshot()
        // }()
      }

      if resp.StatusCode == 503 && resp.Header.Get("server") == "Varnish" &&  t.GetCookieValue("waiting_room") != "" {
        if !strings.Contains(bodystr, "Guru Meditation") {
          DiscardResp(resp)
          t.StatusCh <- StatusInQueue
          t.Metrics.Incr("queue:waiting")

          time.Sleep(15*time.Second)
          return retry.Unrecoverable(errInQueue)
        }
      }
      // if req.IgnoreBody {
      //   req.IgnoreBody
      // }

      if strings.Contains(req.URL.String(), "carts/current") && req.Method == "POST" {
        defer t.resetClient(false)
      }

      t.LogDebug("cache: %s %s %s %v", resp.Header.Get("x-cache"), resp.Header.Get("x-served-by"), req.URL.String(), t.DnsCache[t.Url.Host])
      if (strings.Contains(req.URL.String(), "carts/current") && (strings.Contains(resp.Header.Get("x-cache"), "HIT") || strings.Contains(resp.Header.Get("X-Cache"), "HIT"))) {
        t.resetClient(false)
        t.FastlyJig()
        if strings.Contains(req.URL.String(), "carts/current") {
          // t.RemoveCookie("JSESSIONID")
          // t.FtlGetSession()
           t.Metrics.Incr("ftl_atc_hit")
            // age := resp.Header.Get("age")
            // // fmt.Printf("Age: %s\n",age)
            // if agei, err := strconv.Atoi(age); err == nil && agei > 0 {
            //   sl := time.Duration(180-agei)*time.Second - t.Delay
            //   if sl < 3*time.Second && sl > 0 {
            //     t.LogDebug("time until exp %v\n", sl)
            //     // time.Sleep(sl)
            //     return ErrRetrying
            //   }
            // }
        }
      } else if strings.Contains(resp.Header.Get("x-cache"), "MISS") || strings.Contains(resp.Header.Get("X-Cache"), "MISS"){
        // realip := resp.Header.Get("fastly-client-ip")
        // if realip == "" {
        //   realip = resp.Header.Get("true-client-ip")
        // }
        // if realip != "" {
        //   t.RealIp = realip
        // }
        if strings.Contains(req.URL.String(), "carts/current") {
          if resp.Header.Get("x-cache") != "MISS" && resp.Header.Get("x-cache") != "MISS, MISS"{
            t.FastlyJig()
          }
           t.Metrics.Incr("ftl_atc_miss")
        }
      } else if resp.StatusCode == 302  {
        t.FastlyJig()
        return ErrRetrying
      } else if rand.Float32() < 0.01 {
        t.FastlyJig()
      }


      if resp.StatusCode == 302 && !strings.Contains(resp.Header.Get("location"), t.Url.Host) {
        t.FastlyJig()
        return ErrRetrying
      }

      if resp.StatusCode == 403 {
        // t.FastlyJig()
        t.RemoveBadSession()
        t.resetClient(false)
        t.Metrics.Incr("dd:block")
        if t.GetTelemetryContext("dd_last") == "403" && t.GetTelemetryContext("rotate") == "1" {
          t.RotateProxy()
          return ErrRetrying
        }
        t.Config["dd_last"] = "403"
        if solvedCap {
          t.Metrics.Incr("dd:cap_fail")
          if rand.Float32() < 0.5 {
            t.RemoveCookie("datadome")
          }

          solvedCap = false
          return ErrRetrying
        }

        if strings.Contains(bodystr, "captcha") { // todo less jank
          // t.

          // if resp, err := t.ApiPost("/sesh/get", t.GetBareSessionBytes()); err == nil {
          //   var s TaskSession
          //   if err = json.Unmarshal(resp, &s); err == nil && len(s.UserAgent) > 0 {
          //     t.ApplySession(&s)
          //     t.LogDebug("using sesh: %+v", s)
          //     return ErrRetrying
          //   }
          // } else {
          //   t.LogDebug("sesh get err %+v", err)
          // }
          if os.Getenv("NOCAP") == "1" {
            return retry.Unrecoverable(err)
          }
          if t.GetCookieValue("datadome") == "" && !t.UseIpv6 {
            if rand.Float32() < 0.1 {
              return TaskErrUnrecoverable(ErrRetrying)
            } else {
              return ErrRetrying
            }
          }

          // t.StatusCh <- Status{Message: "DD Block, Rotating proxy"}//StatusRotatingProxy

          // if err := t.RotateProxy(); err != nil {
          //   time.Sleep(time.Duration(rand.Intn(16))*time.Second)
          // } else {
          //   go func() {
          //     defer func() {
          //       recover()
          //     }()
          //     if resp, err := t.ApiPost("/sesh/get", t.GetBareSessionBytes()); err == nil {
          //       var s TaskSession
          //       if err = json.Unmarshal(resp, &s); err == nil && len(s.UserAgent) > 0 {
          //         t.ApplySession(&s)
          //         t.LogDebug("using sesh: %+v", s)
          //       }
          //     }
          //   }()
          // }

          // t.RemoveCookie("datadome")
          // t.RemoveCookie("waiting_room")

          // if os.Getenv("PRESOLVE") != "0" {
          //   if presolve, err := t.ApiGetRetries("/sesh/getcap", 1); err == nil {
          //     if len(presolve) > 5{
          //       t.LogDebug("got presolved %+v", presolve)
          //       t.SetCookie(&http.Cookie{
          //             Name: "datadome",
          //             Value: string(presolve),
          //             Secure: true,
          //             HttpOnly: false,
          //             MaxAge: 1e6,
          //             Path: "/",
          //           // Domain: t.BaseDomainUrl.Host,
          //          },)
          //       return ErrRetrying
          //     }
          //   }
          // }
          DiscardResp(resp)
          bs := map[string]string{}
          json.Unmarshal([]byte(bodystr), &bs)
          t.LogDebug(string(bodystr))
          if solvurl, ok := bs["url"]; ok && !t.UseIpv6 {
            // TODO prevent cap thrash loop when perma-captcha'd
            if !quiet {
              t.StatusCh <- StatusSolvingCap
            }
            // TODO sitekey from api
            // CTxqD0Z3-4Cl5SRFdVktuT77fje3QWZhTLfwT7Bt4LfigpXLNuEvV06S-RhjCXpMA3TxBBd4hTDEMk~tpf-Vg-eP-aCZX9B5ajLW2f04-~
            if cerr := t.DdCapVerify(solvurl, t.GetCookieValue("datadome")); cerr == nil {
              t.Metrics.Incr("dd:cap_solve")
              solvedCap = true
              if !quiet {
                t.StatusCh <- StatusCapSolved
              }
              if derr := t.DatadomeSendPayload(req.URL.String(), t.Url.String(), t.GetCookieValue("datadome")); derr != nil {
                t.Metrics.Incr("dd:gen_fail")
                t.LogDebug("DatadomeSendPayload err=%+v", derr)
                // fmt.Printf("%+v\n", derr)
                // panic("fuark")
                // return derr
              } else {
                t.Metrics.Incr("dd:gen_success")
              }
              return ErrRetrying
            } else {
              if errors.Is(cerr, ErrDdHardBan) {
                t.StatusCh <- StatusDdHardBan
                // t.StatusCh <- StatusRotatingProxy
                orig := t.ProxyStr
                t.resetClient(false)
                t.RotateProxy()
                t.RemoveCookie("datadome")
                if orig == t.ProxyStr {
                  time.Sleep(time.Duration(rand.Intn(30))*time.Second)
                } else {
                  // t.RemoveCookie("waiting_room")
                }
                return ErrRetrying
              }
              if !quiet {
                t.StatusCh <- StatusCaptchaSolveFailedRetrying
              }

              t.RotateProxy()
              return ErrRetrying
              t.LogDebug("cap solve err=%+v", cerr)
            }
          } else {
            t.resetClient(false)
            t.RemoveCookie("datadome")
            req.Close = true
          }
          return ErrDdBanned
        } else {
          t.StatusCh <- Status{Message: "Access Denied (" + stage + ")"}
        }
      } else if resp.StatusCode == 429 {
        DiscardResp(resp)
        CloseH2Conns(t.client)
        t.client.CloseIdleConnections()

        // if !t.UseIpv6 && t.GetTelemetryContext("jig") != "0" {
        //   t.SetTelemetryContext("jig", "0")
        //   t.DnsCacheMut.Lock()
        //   t.DnsCache = nil
        //   t.DnsCacheMut.Unlock()
        //   // req.Close = true
        //   return ErrRateLimited
        // }
        if !quiet {
          t.StatusCh <- Status{Message: "IP/Subnet Rate Limited"}
          // CloseH2Conns(t.client)
          // t.client.CloseIdleConnections()
          // req.Close = true
          // t.StatusCh <- StatusRotatingProxy
          // t.RotateProxy()
          // t.RemoveCookie("datadome")
          // t.RemoveCookie("waiting_room")
        }
        return ErrRateLimited
      } else if resp.StatusCode >= 200 && resp.StatusCode <= 210 && req.Method == "GET" && strings.Contains(resp.Header.Get("content-type"), "text/html") && resp.Header.Get("via") == "1.1 varnish" {
        t.StatusCh <- StatusDdEstablishingTrust
        // todo error handle
        go func() {
          if t.UseIpv6{return}
          defer func() {
                    recover()
                  }()
                 t.TelemetryContextMut.Lock()
          var ref string
          if ld, ok := t.TelemetryContext["dd_last_gen_page"].(string); ok {
            ref = ld
          }
          t.TelemetryContextMut.Unlock()
          if derr := t.DatadomeSendPayload(req.URL.String(), ref, t.GetCookieValue("datadome")); derr != nil && rand.Float32() < 0.25 {
            t.Metrics.Incr("dd:gen_fail")
            t.LogDebug("DatadomeSendPayload err=%+v", derr)
            // fmt.Printf("%+v\n", derr)
            // panic("fuark")
            // return derr
          } else {
            t.Metrics.Incr("dd:gen_success")
          }
          t.TelemetryContextMut.Lock()
          t.TelemetryContext["dd_last_gen_page"] = req.URL.String()
          t.TelemetryContextMut.Unlock()
        }()
      } else if resp.Header.Get("x-datadome") == "protected" {
        if solvedCap || rand.Float32() < 0.03 {
          go func() {
            defer func() {
                    recover()
                  }()
            t.FtlSendSessionSnapshot()
          }()
        }
        t.TelemetryContextMut.Lock()
        t.TelemetryContext["dd_last"] = "allow"
        t.TelemetryContextMut.Unlock()
        t.Metrics.Incr("dd:allow")
      }
    }
    if err != nil {
      return retry.Unrecoverable(err)
    } else {
      return nil
    }
  }, retry.Attempts(6), retry.Delay(t.Delay))
  // if serr, ok := err.(*StatusCodeRejectedError); ok {

  // }
  if resp != nil && err != nil {
    DiscardResp(resp)
  }
  go func() {
    defer func() {
      recover()
    }()
    time.Sleep(5*time.Second)
    _, err = resp.Body.Read(nil)
    if err == nil {
      t.LogDebug("unread_resp %s", req.URL.String())
      t.Metrics.Incr("unread_resp")
      DiscardResp(resp)
    }
  }()
  return resp, err
}

func (t *CheckoutTask) FtlAtc(pid string, csrf string, quantity int) (*FtlCart, error) {
  // if _, ok := t.GetTelemetryContext("dd_last").(string); !ok {
  //   t.`FastlyJig`()
  // }
  var cart FtlCart
  // t.RemoveCookie("cart-guid")
  // t.SetCookie(&http.Cookie{Name: "cart-guid", Value: randHex(4)})
  bp := (t.GetTelemetryContext("cj") == "1") && t.GetCookieValue("cart-guid") != ""
  // err := t.Retry(func() error {
    // url_, err := url.Parse(fmt.Sprintf(t.FtlBase() + "api/users/carts/current/entries?timestamp=%d", timeMillis() + uint64(rand.Intn(86400*3*1e3))))
    url_, err := url.Parse(fmt.Sprintf(t.FtlBase() + "api/users/carts/current/entries?timestamp=%d", timeMillis()+ uint64(rand.Intn(86400*7*1e3))))
  if bp {
    url_, err = url.Parse(fmt.Sprintf(t.FtlBase() + "api/users/carts/current/entries/?timestamp=%d", timeMillis()+ uint64(rand.Intn(86400*7*1e3))))
    if err != nil {
      return nil, err
    }
  }
  // TODO re enable; but use unique dd cookies
  // if rand.Float32() < 0.75 {
  //   url_.Path = strings.Replace(url_.Path, "api", "api", -1)
  //   t.SetTelemetryContext("cart_path", url_.Path)
  // }


    headerOrder := []string {
      "content-length",
      "accept",
      "x-csrf-token",
      "traceparent",
      "x-fl-productid",
      "newrelic",
      "user-agent",
      "x-fl-request-id",
      "content-type",
      "tracestate",
      "origin",
      "sec-fetch-site",
      "sec-fetch-mode",
      "sec-fetch-dest",
      "sec-gpc",
      "referer",
      "accept-encoding",
      "accept-language",
      "cookie",
    }
    bb := randHex(8)
    bpid := pid
    if bp {
      bpid = randHex(4)
    }
    headers := [][2]string {
      {"accept", "application/json"},
      {"x-csrf-token", csrf},
      {"traceparent", "xx-" + bb + "-" + bb + "-xx"},
      // {"breadcrumbid", "xx-" + bb + "-" + bb + "-xx"},
      // {"x-fl-productid", pid},
      {"x-fl-productid", bpid},
      {"newrelic", strings.Repeat(bb, 14)},
      {"x-fl-request-id", uuid.New().String()},
      {"content-type", "application/json"},
      {"tracestate", fmt.Sprintf("%d@nr=2-3-%d-%d-", rand.Intn(65535), rand.Intn(65535), rand.Intn(65535)) + bb + "----" + fmt.Sprintf("%d", timeMillis() + uint64(rand.Intn(50000)))},
      {"origin", t.FtlOrigin()},
      {"sec-fetch-site", "same-origin"},
      // {"fastly-force-shield", "no"},
      {"sec-fetch-mode", "cors"},
      {"sec-fetch-dest", "empty"},
      {"referer", "https://www.footlocker.com/cart"},
      {"accept-encoding", "gzip, deflate, br"},
      {"accept-language", "en-US,en;q=0.9"},
    }
    bodyStructure := map[string]interface{} {
  	  "productQuantity": quantity,
  	  "productId": pid,
      // "z": timeMillis(),
  	}
    body, err := json.Marshal(bodyStructure)
    if err != nil {
      return nil, err
    }
    if bp  {
      body = []byte(fmt.Sprintf(`{"productQuantity": %d, "productId": "%s", "productId": "%s"}`, quantity, bpid, pid))
    }

    t.LogDebug("atc %s\n", string(body))
    req := t.makeReq("POST", url_, &headers, &headerOrder, &body)
    // req.Close = false // t.UseIpv6 && rand.Float32() < 0.5
    req.Close = false // t.GetTelemetryContext("dd_last") == "403" // rand.Float32() < 0.5
    if t.UseIpv6 {
      t.RemoveCookie("datadome")
    }
    req.IgnoreBody = true
    req.IgnoreBodyExceptStatus = []int{200,403}
    // t.DnsCache = nil
    // t.FtlGetCart()
    resp, err := t.FtlWebDoRequest(req, "Add to Cart")
    if resp != nil {
      if resp.StatusCode == 531 || resp.StatusCode == 403 {
        t.LogDebug("x-cache=%s,%v", resp.Header.Get("x-cache"), t.DnsCacheActiveHost)
        t.LogDebug("x-served-by=%s", resp.Header.Get("x-served-by"))
      }
      // age := resp.Header.Get("age")
      // fmt.Printf("Age: %s\n",age)
      // if agei, err := strconv.Atoi(age); err == nil && agei > 0 {
      //   sl := time.Duration(180-agei)*time.Second - t.Delay
      //   fmt.Printf("sleeping %v\n", sl)
      //   time.Sleep(sl)
      // }
      if resp.StatusCode == 403 {
        t.resetClient(false)
        t.RemoveCookie("datadome")
      }
      if resp.StatusCode == 531 {
        DiscardResp(resp)
        // TODO refactor to return to main loop, get product page, try again
    		return nil, ftlErrOutOfStock
      }
    }

    // var cart FtlCart
    if err := readRespJsonDst(resp, &cart); err != nil {
      DiscardResp(resp)
    	return nil, err
    }
    if len(cart.Entries) > 0 {
      t.SetTelemetryContext("ftl_cart_cache", resp.Header.Get("x-served-by"))
      t.SetTelemetryContext("ftl_cart_dns", t.DnsCacheActiveHost[t.Url.Host])
    }

  // })
  cart.Time = time.Now()

  return &cart, err
}

func (t *CheckoutTask) FtlSetBilling(csrf string) (error) {
  url_, err := url.Parse(fmt.Sprintf(t.FtlBase() + "api/users/carts/current/set-billing?timestamp=%d", timeMillis()))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-csrf-token",
    "user-agent",
    "x-fl-request-id",
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
    {"x-csrf-token", csrf},
    {"x-fl-request-id", uuid.New().String()},
    {"content-type", "application/json"},
    {"origin", t.FtlOrigin()},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", t.FtlBase() + "checkout"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }


  bodyStructure := map[string]interface{} {
	  "setAsBilling": false,
	  "country": map[string]interface{} {
	    "isocode": t.Profile.BillingAddress.Country,
	    "name": "United States",
	  },
	  "id": nil,
	  "region": map[string]interface{} {
      "isocode": fmt.Sprintf("%s-%s", t.Profile.BillingAddress.Country, t.Profile.BillingAddress.State),
	  },
	  "phone": t.Profile.BillingAddress.NordPhone(),
	  "type": "default",
	  "firstName": t.Profile.BillingAddress.FirstName,
	  "lastName": t.Profile.BillingAddress.LastName,
	  "line1": t.Profile.BillingAddress.Address1,
	  "postalCode": t.Profile.BillingAddress.Zip,
	  "town": strings.ToUpper(t.Profile.BillingAddress.City),
	  "regionFPO": nil,
	  "recordType": "S",
	  "shippingAddress": true,
	}
  if t.Profile.BillingAddress.Address2 != "" {
    bodyStructure["line2"] = t.Profile.BillingAddress.Address2
  }
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  resp, err := t.FtlWebDoRequest(t.makeReq("POST", url_, &headers, &headerOrder, &body), "Billing")
  if err != nil {
    return err
  }
  DiscardResp(resp)
  return nil
}

func (t *CheckoutTask) FtlSetShipping(csrf string) (error) {
  url_, err := url.Parse(fmt.Sprintf(t.FtlBase() + "g?timestamp=%d", timeMillis()))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-csrf-token",
    "user-agent",
    "x-fl-request-id",
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
    {"x-csrf-token", csrf},
    {"x-fl-request-id", uuid.New().String()},
    {"content-type", "application/json"},
    {"origin", t.FtlOrigin()},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", t.FtlBase() + "checkout"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  bodyStructure := map[string]interface{} {
  "shippingAddress": map[string]interface{}{
    "setAsDefaultBilling": false,
    "setAsDefaultShipping": false,
    "id": nil,
    "setAsBilling": false,
    "type": "default",
    "LoqateSearch": "",
    "regionFPO": nil,
    "shippingAddress": true,
    "recordType": " ",
    "country": map[string]interface{} {
      "isocode": strings.TrimSpace(t.Profile.ShippingAddress.Country),
      "name": "United States",
    },
    "firstName": t.Profile.ShippingAddress.FirstName,
    "lastName": t.Profile.ShippingAddress.LastName,
    "line1": t.Profile.ShippingAddress.Address1,
    "email": false,
    "phone": t.Profile.ShippingAddress.NordPhone(),
    "postalCode": t.Profile.ShippingAddress.Zip,
    "region": map[string]interface{} {
      "countryIso": t.Profile.ShippingAddress.Country,
      "isocode": fmt.Sprintf("%s-%s", strings.TrimSpace(t.Profile.ShippingAddress.Country), t.Profile.ShippingAddress.State),
      "isocodeShort": t.Profile.ShippingAddress.State,
      "name": "District of Columbia",
    },
    "town": strings.ToUpper(t.Profile.ShippingAddress.City),
  },
}
  if t.Profile.ShippingAddress.Address2 != "" {
    bodyStructure["shippingAddress"].(map[string]interface{})["line2"] = t.Profile.ShippingAddress.Address2
  }
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  t.LogDebug(string(body))
  resp, err := t.FtlWebDoRequest(t.makeReq("POST", url_, &headers, &headerOrder, &body), "Shipping")
  DiscardResp(resp)
  if err != nil {
    return err
  }
  if resp != nil && resp.StatusCode == 500 {
    t.StatusCh <- Status{Message: "Shipping Address Rejected (UPS Verification Failed)"}
    return retry.Unrecoverable(retry.Unrecoverable(retry.Unrecoverable(ErrShippingAddressRejected)))
  }

  return nil
}

func (t *CheckoutTask) FtlGetCartIgnore() (*FtlCart, error) {

  url_, err := url.Parse(fmt.Sprintf(t.FtlBase() + "api/users/carts/current?checkInventory=true&timestamp=%d", timeMillis()))
  if err != nil {
    return nil, err
  }
  // if rand.Float32() < 0.5 {
  //   url_.Path = strings.Replace(url_.Path, "api", "api", -1)
  //   // t.SetTelemetryContext("stock_path", apiUrl.Path)
  // }
  headerOrder := []string {
    "accept",
    "user-agent",
    "x-fl-request-id",
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
    {"x-fl-request-id", uuid.New().String()},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", t.FtlBase() + "checkout"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  req := t.makeReq("GET", url_, &headers, &headerOrder, nil)
  req.IgnoreBody = true
  resp, err := t.FtlWebDoRequest(req, "Get Cart")
  DiscardResp(resp)
  if err != nil {
    return nil, err
  }
  // fmt.Printf("%+v", ftlCart)
  return nil, nil
}

func (t *CheckoutTask) FtlGetCart() (*FtlCart, error) {

  url_, err := url.Parse(fmt.Sprintf(t.FtlBase() + "api/users/carts/current?checkInventory=true&timestamp=%d", timeMillis()))
  if err != nil {
    return nil, err
  }

  headerOrder := []string {
    "accept",
    "user-agent",
    "x-fl-request-id",
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
    {"x-fl-request-id", uuid.New().String()},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", t.FtlBase() + "checkout"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := t.FtlWebDoRequest(t.makeReq("GET", url_, &headers, &headerOrder, nil), "Get Cart")
  if err != nil {
    return nil, err
  }
  var ftlCart FtlCart
  if err := readRespJsonDst(resp, &ftlCart); err != nil {
  	return nil, err
  }
  // fmt.Printf("%+v", ftlCart)
  return &ftlCart, nil
}

func (t *CheckoutTask) FtlSetPayment(csrf string, billingAddressId string) (error) {
  return t.Retry(func() error {
  	token, err := t.FtlJigCard(csrf)
  	if err != nil {
  	  return err
  	}

    url_, err := url.Parse(fmt.Sprintf(t.FtlBase() + "api/users/carts/current/payment-detail?timestamp=%d", timeMillis()))
    if err != nil {
      return err
    }
    headerOrder := []string {
      "content-length",
      "accept",
      "x-csrf-token",
      "user-agent",
      "x-fl-request-id",
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
      {"x-csrf-token", csrf},
      {"x-fl-request-id", uuid.New().String()},
      {"content-type", "application/json"},
      {"origin", t.FtlOrigin()},
      {"sec-fetch-site", "same-origin"},
      {"sec-fetch-mode", "cors"},
      {"sec-fetch-dest", "empty"},
      {"referer", t.FtlBase() + "checkout"},
      {"accept-encoding", "gzip, deflate, br"},
      {"accept-language", "en-US,en;q=0.9"},
    }
    bodyStructure := map[string]interface{} {
    "gateway": "Credit",
    "saved": false,
    "cardType": map[string]interface{} {
      "code": strings.ToLower(t.Profile.Card.Type), // TODO test all card types manual / find in js
    },
    "expiryMonth": fmt.Sprintf("%02d", t.Profile.Card.ExpMonth),
    "expiryYear": fmt.Sprintf("%d", t.Profile.Card.ExpYear),
    "flApiCCNumber": token.Token, //
    "billingAddress": map[string]interface{} {
      "id": billingAddressId,
    },
  }
    body, err := json.Marshal(bodyStructure)
    if err != nil {
      return err
    }
    resp, err := t.FtlWebDoRequest(t.makeReq("POST", url_, &headers, &headerOrder, &body), "Payment")
    if err != nil {
      return err
    }
    DiscardResp(resp)
    return nil
  }, retry.Attempts(5))
}

type FtlErrorResp struct {
  Code string `json:"code"`
	Errors []FtlError `json:"errors"`
}

type FtlError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    interface{}    `json:"code"`
}

func (t *CheckoutTask) FtlSubmit(csrf string, cart *FtlCart) (error) {
  url_, err := url.Parse(fmt.Sprintf(t.FtlBase() + "api/users/orders?timestamp=%d", timeMillis()))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-csrf-token",
    "user-agent",
    "x-fl-request-id",
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
    {"x-csrf-token", csrf},
    {"x-fl-request-id", uuid.New().String()},
    {"content-type", "application/json"},
    {"origin", t.FtlOrigin()},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", t.FtlBase() + "checkout"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  bodyStructure := map[string]interface{} {
	  "preferredLanguage": "en",
	  "securityCode": t.Profile.Card.Cvc,
	  "deviceId": "",
	  "cartId": cart.GUID,
	}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return err
  }

  resp, err := t.FtlWebDoRequest(t.makeReq("POST", url_, &headers, &headerOrder, &body), "Order")
  if err != nil && resp == nil {
    return err
  }

  bb, _ := readBodyBytes(resp)
  var sresp FtlErrorResp
  if err := json.Unmarshal(bb, &sresp); err != nil {
  	return err
  }

  go t.SendTelemetry(map[string]interface{} {
    "event": "ftl_order_response",
    "ftl_order_response": bb,
  })

  if len(sresp.Errors) > 0 {
  	if strings.Contains(sresp.Errors[0].Message, "different card") || strings.Contains(sresp.Errors[0].Message, "another card") {
  	  return errCardDeclined
  	} else {
      return errors.New(sresp.Errors[0].Message)
    }
  } else if resp.StatusCode == 200 {
  	return nil
  } else if resp.StatusCode == 403 {
    return ErrProxyBanned
  } else {
    return errUnexpectedCheckoutError
  }
}

func (t *CheckoutTask) FtlStartCheckout(csrf string) (error) {

  url_, err := url.Parse(fmt.Sprintf("%sapi/users/carts/current/email/%s?timestamp=%d", t.FtlBase(), t.Profile.Email, timeMillis()))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-csrf-token",
    "user-agent",
    "x-fl-request-id",
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
    {"x-csrf-token", csrf},
    {"x-fl-request-id", uuid.New().String()},
    {"origin", t.FtlOrigin()},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", t.FtlBase() + "checkout"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := t.FtlWebDoRequest(t.makeReq("PUT", url_, &headers, &headerOrder, &[]byte{}), "Start Checkout")
  if err != nil {
    return err
  }
  DiscardResp(resp)
  return nil
}

func (t *CheckoutTask) hasValidAbckFtl() func(abck string) bool {
  var stage = 0
  return func(abck string) bool {
    stage += 1
    return stage > rand.Intn(5)
  }
}

func (t *CheckoutTask) hasValidAbckFtlClient(client *http.Client) func() bool {
  return func() bool {
    abck := getCookieValue("_abck", t.BaseUrl, client.Jar)
    // t.LogDebug("%+v", abck)
    if abck == "" {
      return false
    }
    return strings.Contains(abck[:64], "~0~") //  || strings.Contains(abck, "||")
  }
}

func (t *CheckoutTask) FtlEnsureValidAbck() error {
  return nil
  // fmt.Println("last valid", t.Config["last_valid_abck"])
  // if t.Config["last_valid_abck"] != "" {
  //   if ckie := getCookie("_abck", t.BaseUrl, t.client.Jar); ckie != nil {
  //     ckie.Secure = true
  //     ckie.HttpOnly = false
  //     ckie.MaxAge = 1e6
  //     ckie.Domain = t.BaseDomainUrl.Host
  //     ckie.Value = t.Config["last_valid_abck"]
  //     t.client.Jar.SetCookies(t.BaseDomainUrl, []*http.Cookie { ckie })
  //     return nil
  //   }
  // }
  return t.EnsureValidAbck(t.Config["akUrl"], t.hasValidAbckFtl())
}

type FtlCardToken struct {
	Token        string `json:"token"`
	IsSuccessful bool   `json:"isSuccessful"`
	Error        string `json:"error"`
}

func (t *CheckoutTask) FtlJigCard(csrf string) (*FtlCardToken, error) {
  url_, err := url.Parse(fmt.Sprintf(t.FtlBase() + "paygate/cct?timestamp=%d", timeMillis()))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-csrf-token",
    "user-agent",
    "x-fl-request-id",
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
    {"x-csrf-token", csrf},
    {"x-fl-request-id", uuid.New().String()},
    {"content-type", "application/json"},
    {"origin", t.FtlOrigin()},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", t.FtlBase() + "checkout"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  bodyStructure := map[string]interface{} {
	  "cardNumber": t.Profile.Card.Number,
	}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }

  resp, err := t.FtlWebDoRequest(t.makeReq("POST", url_, &headers, &headerOrder, &body), "Card Encryption")
  if err != nil {
    return nil, err
  }
  // if resp.StatusCode == 429 {
  //   return nil, retry.Unrecoverable(errInvalidCard)
  // }

  var token FtlCardToken
  if err := readRespJsonDst(resp, &token); err != nil {
  	return nil, err
  }
  return &token, nil
}

func (t *CheckoutTask) FtlPaypalFinishOrder(presp *PaypalCheckoutSuccess, csrf string) (error) {
  err := t.FtlApplyPaypalToCart(presp, csrf, "PayPalAccount")
  if err != nil {
    return err
  }

  // t.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36"
  url_, err := url.Parse(fmt.Sprintf(t.FtlBase() + "api/users/orders?timestamp=%d", timeMillis()))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-csrf-token",
    "user-agent",
    "x-fl-request-id",
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
    {"x-csrf-token", csrf},
    {"x-fl-request-id", uuid.New().String()},
    {"content-type", "application/json"},
    {"origin", "https://" + t.Url.Host + ""},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", "https://" + t.Url.Host + "/checkout"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }

  bodyStructure := map[string]interface{} {
  "preferredLanguage": "en",
  "securityCode": "",
  "deviceId": "0400JapG4txqVP4Nf94lis1ztrjQCvk297SBnrp/XmcfWoVVgr+Rt2dAZIo7BJIRIWDNtjiuvPP9Vk+xH1ZPRIwM6njw/ujAyYdbGKZt5JLThTvosS1xgSAgNfLEMokGoGJx1x7Vc9CP0wq4vIwXj7xMGGA73rYZtIsVyoA80pXCaXaHarNYAXR7cNC1UJgA1BQZDi8R6wVNV9r6qoxEqm2gQ71/Y9e8eiB47caT1pbMLWowNvvGpkrbfUjXzk77ILZ/eUsQ7RNrLro1kTKIs1496YkpIh3A707lm2e25SQbo1Nahvu1nnvSP6dYfFIGMH4xTn3Cs8SNRaHmY1XpDGBtyepPcth3j49FqfEF5rn/fvH/R7UZRBPWKNGAAm/mcNZvBO2H/Tlr05hcVLXZTPsAuL6tSoymLcas/dbeJBd5ABDb8RzWdzZK4RpNGfQbU0JlfhyrSbxA8f904hiAcL/dBjV9nBJW1Y3JLACqjtcRnAt8qjUOoCPake9RJsn4u3PdHf7WUtiodkQZ2kH+5AQDR8Ee0JyrVTVozBRxN5nDyWopQs/EGzmZLzHZ809aE2mlMXvg7Fjyu7Q2qKge+Ey13gQxSxzPuRtycb0H5IjHkCcKSs4KVYKB4p8ICkH/4CLq+Tp+QmlT5YNHDo0vYYDC3xqkbjiKwBZfdLDs+1JC1RHeqz7TtH33U5qH9vja1PCVFgzzHSahZXK8hZCSHGBUGQbz44XiMf/eY92hNSmr+k28EssPaX5qs/6UpGGAjQt+Ep48DVRl0oBkPCJz7Ln+v5KTIBvOIMzuTzEBdNLiXsPV6scyeMvquKgBCHz1/QUZbZ+NCYNO3AHcIII2ayAqc5YwJVsFx3E/58zGq/hZ0ZoyseD0G1/ThxEIvsJRqGfF9spFd7LTpxiIVb7/dLR/aeq024GYThb6+Z4baa9a0Da8VDdLGBfK9PH/LgnAtMTg/gz07uVSINAgzJ8FJa/++5dWoUSzxwidETiC3QRoEk/+wksDLHOftRZrvUUNEWo8zkk+RyyTuzxje3kFUHvz2oJ9ekNnn5cNz38wBbph49eozb7yCE8lGHNhkYsv4tGGTu1D5NehnYRMkiGsGf2WQMx2HuDoTGzXRgEqSehE5i2Mped1MhppWAoPjvWMTtkeBqVHq2r3awv7xVOwtmVnfQfJ/zck+QEC91vDzq9GxxVn+WvCwoViHULgwe4ejOQ9IF9wOHABPWPenBG5oxEOsA==",
  "cartId": t.GetCookieValue("cart-guid"),
  }
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  return t.Retry(func() error {
    // t.FtlEnsureValidAbck()

    t.StartPaymentTime = timeMillis()
    resp, err := t.FtlWebDoRequest(t.makeReq("POST", url_, &headers, &headerOrder, &body), "Ordering")
    if resp == nil || resp.Body == nil || (err != nil && !t.shouldRetry(err) && (resp == nil || ErrIsTaskUnrecoverable(err))) {
      DiscardResp(resp)
      if err == nil {
        err = ErrRetrying
      }
      return err
    }
    body, err := readBodyBytes(resp)
    if err != nil {
      return err
    }
    // resp.Body = ioutil.NopCloser(bytes.NewReader(body))


    var sresp FtlErrorResp
    if err := json.Unmarshal(body, &sresp); err != nil {
      return err
    }
    go t.SendTelemetry(map[string]interface{} {
      "event": "ftl_paypal_order",
      "ftl_paypal_order": map[string]interface{} {
        "body": string(body),
        "status": resp.Status,
        "err": fmt.Sprintf("%+v", err),
      },
    })
    if len(sresp.Errors) > 0 {
      site_err := sresp.Errors[0]
      errmsg := sresp.Errors[0].Message
      if strings.Contains(errmsg, "We suggest you check again or try another card") || strings.Contains(errmsg, "Invalid Payment details") || strings.Contains(errmsg, "We suggest you try again or use another payment method") {
        return retry.Unrecoverable(errCardDeclined)
      } else if strings.Contains(errmsg, "has to be removed from cart") || strings.Contains(errmsg, "Cart is empty") || strings.Contains(errmsg, "Product Reservation has expired") {
        // t.FtlGetCart()
        t.StatusCh <- StatusCartJacked
        return retry.Unrecoverable(retry.Unrecoverable(ErrCartJacked))
      } else if strings.Contains(errmsg, "Token does not match session") {
        t.StatusCh <- StatusCartExpired
        return retry.Unrecoverable(ErrSessionExpired)
      } else if strings.Contains(errmsg, "Item in your Cart is limited to 1 per Customer") {
        // TODO profile rotation
        t.StatusCh <- StatusProfileAlreadyCheckedOutItem
        return TaskErrUnrecoverable(ErrSessionExpired)
      } else if strings.Contains(errmsg, "Product Reservation has expired")  {
        t.StatusCh <- StatusCartExpired
        // TODO dont fully restart task, go back to atc
        return TaskErrUnrecoverable(ErrSessionExpired)
      } else if strings.Contains(errmsg, "Validation error") && site_err.Code == 0 {
        t.StatusCh <- StatusCartExpired
        // TODO dont fully restart task, go back to atc
        return TaskErrUnrecoverable(ErrSessionExpired)
      } else {
        return ErrRetrying
      }
    } else {
      var order FtlOrderSummary
      if err := json.Unmarshal(body, &order); err == nil {
        if len(order.Order.Code) > 0 {
          t.SetTelemetryContext("ordernum", order.Order.Code)
        }
      }
      return nil
    }
    return errUnexpectedCheckoutError
  }, retry.Attempts(2), retry.Delay(0*time.Second))
}

func (t *CheckoutTask) FtlAdyenOrder(sesh *FtlSession) (error) {
  cartguid := t.GetCookieValue("cart-guid")
  // t.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36"
  url_, err := url.Parse(fmt.Sprintf(t.FtlBase() + "api/v2/users/orders?timestamp=%d", timeMillis()))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-csrf-token",
    "user-agent",
    "x-fl-request-id",
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
  return t.Retry(func() error {
    // t.FtlEnsureValidAbck()
    headers := [][2]string {
      {"accept", "application/json"},
      {"x-csrf-token", sesh.Data.CsrfToken},
      {"x-fl-request-id", uuid.New().String()},
      {"content-type", "application/json"},
      {"origin", "https://" + t.Url.Host + ""},
      {"sec-fetch-site", "same-origin"},
      {"sec-fetch-mode", "cors"},
      {"sec-fetch-dest", "empty"},
      {"referer", "https://" + t.Url.Host + "/checkout"},
      {"accept-encoding", "gzip, deflate, br"},
      {"accept-language", "en-US,en;q=0.9"},
    }

    fp, _ := t.getAdyenFp(2)
    ay, err := AdyenMultiEncrypt(t.Profile.Card, fp)
    if err != nil {
      return err
    }
    bodyStructure := map[string]interface{} {
    "preferredLanguage": "en",
    "termsAndCondition": false,
    "deviceId": "0400tyDoXSFjKeoNf94lis1zts15tY4EpA9p2F1gxe1tA7ncJ5IZ4FjqbSNzzHW9ahduH3RduePjg2hIs5Leau5ALOTV5e47uAYJR+BRm4sTbEImINsJ9VlMYGTX23I2zuDzEYngAz99aSMxPsziLTvCI6V5TrKFtoyvmGODXWbkrLSLP/wT3etan16/FINOuYS5LH/sQ/sXqk0yDUQkBm21+3hdrSzenwv2iK7pMCw9zdQDHhpZVFJIfuTQHjy7H5XhLHCHAGY3J0jpLqVrSVzsI1LB6M5fl2EoMEAoEitEttFUo1DRkUtPGBDt+bV6wpIPYoFkLaHHs+V4VShbY+EVBD6upMsx8e+4a0Yy5JnHYpNW+4LcpZ3IMAt004u7280RJYmuM4HtzKHHImOg7rLXZjfbG4qRc28WV67rFsW5+713xBfNPgHEB9/KUsiAoCUdJndkm9obPrK94iG+Obwalwdhg13I6YIZvZJVRTsOxKGM1eyAwSCEbXPCKc6iDsD88+g/mDVDCxHwEBUtnxfjempApfFWA+gzp61dl4+5GiVhO1Q2jmVgfOlzR3iofwxf87p3+HmNUhCTEEZ0CdteMyaUGABVKw5X0ikb/3h1d0Im6Fy7sPS5unsowb8ZPzurChB0xK1ryt0RBN5Cs8UsEjFdh3l37CKXAwbSyBXB7HMibAjilOCqFlrfKCZtmOT1S4e4f0GE7S28oxo5VEoPmW5VK3YFpCSU+fEAAdzX30Ad/g3ekijXJBAY2IHpPa0cZS0HPjgc2Tff+x4pw6W9esbGEpq7CoRYvFQ3SxgXyvTx/y4JwLTE4Nxv4KowoTq9FkM8/lgfSJMvEqDnvZw2o0zI+mPNi65X5WFPYlDG9N0Lbh5nOj3u3DXqRCiKCUrsEkMt8z9fxO+ITycV+QprneG5CIm4d8hq7oc5h1YbsSjz0paGQAofAE8xAXTS4l7DQtPYEf0PHxnrIWh2D/DqupPnFVyYldsigJnjRx+vsOpmpE0Tgww+85Nze+FmXgdeB+juvc3V9mSuB3VV17Gw6vt8UE3KYJYJUl9OoTOizNgLYgctQgLMmnLnwcg4b56noI4kZo8o34BauODchk2WrDaU9I8RMbbaYeTcLtcejyCxPk2UkqDvdYEYDfmfBwBYRhc3f5urJ1aotlycMEBL1IzyVekwhI5/tPvXfQEg+KLRVsDNUznM+4IlRQO86SD0OIS7LH591kg7Q92FGKueu33EPopga4/ZIoOpEIX1KkX4xPDcpR57NEFoJWeLyJF7VYK/kbdnQGRROPHMSFFvUNyYaft0eHAV9N1yOZDRsMMLpiAB+gLOiw==",
    "cartId": t.GetCookieValue("cart-guid"),
    "encryptedCardNumber": ay.Number, // "adyenjs_0_1_18$bWr+wZgszfrj2/P5tJfcqA5GcZtxRf+IafvnJqk8Fh2v9AXh4ubQBh4A8HtY75YfaQw4mRAJIZAcbQqYEIZNu3V2FAopgYly3W60y6fyqRNnHrljNLTP6bRff9so/O9edRontUKRX0t+eVd0j9nyiEHw9hxBFBdbCJoVv6mIhTZPGwCpwUSyCBYvF9FI4CD+hAfJoI3BSOgg8uoKSFloi9O6tGDY1jrqWiTarvkzOAOfTflYCQjmYjOzvzXQ71eBBA+0199551WEiDniRtNCsRoaWfuampLTLJdCgXfaIoyF+xUE98LglZUwFGT92xmwmJtI0rZPJfysm2AYuZqBsg==$iiiSVLN+IheR9JZPSnSslVYMh6QtVvmR8EBAZLfaoXlmj9YXWGncpKjWKAQkTDGXhqmdXufwA4jYw+xa9WsGLw28yUD6O0KBpxvUH2DJtqhJaYJoqTtMmGl90vGlcIuT+jbFGO9vY3VrqtN1SfqrScigjD8M1t42qxopofskoN7Bbdda/BVAxTp8gcPgloyZQDAsABdNfMxUSmfZ4VC6m6f773VPkbicBfh0p0LWecKaNpwUWjkyCrVrVXb2pZFZr+VuRJQRM17PLMJrKWskJ9aUsIOEEMjp7if6jFAa321luwCrtvpNw2vE1Lt2UI3BDPGQvFPtC0HCMaoCJqijYIeqoFAl+Z6QkGa8FGjx6u+mxKKzrhOVH6YMgsKhYJRLMssmMJoYLo6GqtjaKSrFdqMRxFmxZ7/UTDKBN092OZdkKCI5usw2JjKL",
    "encryptedExpiryMonth": ay.ExpMonth, // "adyenjs_0_1_18$e9Z78jxlaFVeBC8VIsTnYwb1av6oCbfpXsQ/qmgZ0RUmfcoHhgVhmXfk+oWHwAIDrWLqUWz71sVEcJGJCDbeG51HEkgLaFZPnebs0HK5ide941EvK+X+nk8oFrSNANscimJQoWryRrnrNo+cJ6ja/eRLsNTGXLIqtEvaOVzVdJTfCLLolTCffdOwtO+NGPC/NHK1b3i0ME4j6OKdmLPhLMwEjSe0m+t4DOEJc6v6NsjlL0YGS/Yk2rqT3eQzVNzxoqzle+YO/5LCXv5wJPAiMxGE9w8UkDOtD3HaaURE3/XREV6leMc/apryBJMD8gEPHCB+PyBvKTsbUht8Kr+j+Q==$RvKTjdCTiB/IZdtWA5bYCvIx9tMkAfBPqjo3bl22Y1EyAsBrhkf4+POYxd6fEsliQmmLlOjS6v8/3ZXUurmUrpCr4CXtrgUBU04ScCOXklHDK1DUwywpvdtsPPrd7dmbfpBp0ybFdO6PQw+wjv/E/v0z1gFdONhhDwRCKMx0UIG0G85JctYrQpZUD0Yx1TP4K4MdIA/LDJEUO+BJyWOY3xcfMmKKZNJlAh6HGfcWXXrGCis0+/L3x65shht7yhNrHtfeOZljSFVI3Cc5RP2oXiKCKEHp7wZrP5T6laZyj2mvvcbNaMhIpDiIOoxjBoMJ7RfO2WWgLOfD3S/hf9s5x6A=",
    "encryptedExpiryYear": ay.ExpYear, // "adyenjs_0_1_18$HNo/WDle3J4Ptb8I0Z8pm9cPoL4ceOI+hOay8wzkqymx//wxvLKc6XQCWcC7cQvXpAwzMiEssXra3v8GztYXg6O9UMkyfaD8VL0di7oHnxIxU99cDu3aTFnBnZRB8aHcIwI/3sUqeAWMPCr/K1fNu7Fo1D8FNNay2AV/VtCRavBYU9SF5WAqJSN6B4XQ0lR3nLRzx/37IFFPAxvbABdnS1oy+Q6hte65PJ+l3G3314kQ3tSAKsOpR1ZGjJATyF7RLLrLmU7hM6ed1uoNgPoHYd0gStEnVcDXTho5KOTT8WucYdbOf0V+RRA9EMqCfjsJ2XIEP8fvzVeyY1VZWW5FSw==$AdLiLNxywlwYyJ+xjiVCi66dsBlsUbrr4ukH6t0Eq4tcopfmJ5ZwnXFaza9S27g5NVpIAE2tYSB1DX85RK4/Lu9pIS7rEKBLo3gvu4aFy8M2pM3nk2dTa7JDLfyrog/MU3O/YDnzghyjd3i9MtOcsBT5QIpQYJEIe3CqBhu/IWOGUDIK6g+6pyN0LwOQy676dXYfMWWlUcsh9xkasigY2PENbQ0Ras0iw3sk85AHY9/MXVgd4T2YyTpMJleukxZaS0Rzz3xz26rWqdP4KOzWf0xyvMKXuOXzXBQdB2Dp+UcUrkeQ+aLzlhRJeNUPxuk7Vih5ewEeZEbi9TrI0HDLinWm",
    "encryptedSecurityCode": ay.Cvc, // "adyenjs_0_1_18$GoDC07jlgsn3U3wIEkVhIAWz65pGWVXNMdZEcni/0DuoCfnGGuL9UQeadQMX9mz4rzVEWYX0jJH4M6aWfXeWfXRy4a4d4UAw5sVw4qKZkriXx0zRyA75C0CBZGhHrV0zDWbEB1MSgS4obrN2OByWJ8Y2gYV97idSw3ib9LXBlqiEJwVqpM2g+5prOoGykLlETp7nQ+9HsqOQyvt8fAIe7zHcXuD4YYLIr8xFiORiAVeYzUciAu9t3yXkAb94oBryWGyqG9Iqi6cAHqpXlV+Q3x5c3uEg1vb6/Lw6B2AbvU1Y/bdBvHx/TkYrRJy3WxwZY+/+K+hjqfpnvawEoRr06w==$Sbn/ZxYzCaMnTN/+LdgkEaR321GCWtLXkuSFww5dZYSa3JgqVT5PzJvO3VTrxzZoQizH1EAwvYCE9Wpa9FXo2rFyH+Z6VPzcKq6eBbjVn74b641XkXUjbJ3GUwhlqgkawJDVmDiQ484CzXcZH+92Ho36khL1FbQxkC5pYQ8HYrk4PxRSJS2sWly983EGB6ITQTSv0TWBpBG9hv0yAzI35RCR1NSFNtduUTIGz4znobgSB7119DJMOL5+mei0tA79BR3yehRLl2qhT0p0j3W2pSP4SqRxcIl2O1MkFiKkw00QtTz1sQTaKD5hkkTfcXC4DFyiJz/h9fiQHQ==",
    "returnUrl": fmt.Sprintf("https://%s/adyen/checkout", t.Url.Host),
    "paymentMethod": "CREDITCARD",
    "browserInfo": map[string]interface{} {
      "screenWidth": 1366, // TODO real
      "screenHeight": 768,
      "colorDepth": 24,
      "userAgent": t.UserAgent,
      "timeZoneOffset": 420,
      "language": "en-US",
      "javaEnabled": false,
      },
    }
    reqbody, err := json.Marshal(bodyStructure)
    if err != nil {
      return err
    }

    if t.OneCheckoutPerProf && t.CheckHit != nil && t.CheckHit(t.Url.Host, t.Profile.Name) > 0 {
      t.StatusCh <- Status{Message: "Stopping Task, Already Hit with Profile"}
      return retry.Unrecoverable(retry.Unrecoverable(retry.Unrecoverable(retry.Unrecoverable(errors.New("stopping task, already hit")))))
    }

    resp, err := t.FtlWebDoRequest(t.makeReq("POST", url_, &headers, &headerOrder, &reqbody), "Ordering")
    // fmt.Printf("%+v\n", err)
    if resp == nil || resp.Body == nil || (err != nil && !t.shouldRetry(err) && (resp == nil || ErrIsTaskUnrecoverable(err))) {
      DiscardResp(resp)
      if err == nil {
        err = ErrRetrying
      }
      return err
    }

    body, berr := readBodyBytes(resp)
    if berr != nil {
      return berr
    }
    t.LogDebug("%+v", resp)
    t.LogDebug("BODY %s", string(body))
    resp.Body = ioutil.NopCloser(bytes.NewReader(body))

    var sresp FtlErrorResp
    if jerr := json.Unmarshal(body, &sresp); jerr != nil {
      t.LogDebug("JERR %v", jerr)
      return jerr
    }
    go t.SendTelemetry(map[string]interface{} {
      "event": "ftl_adyen_order",
      "ftl_adyen_order": map[string]interface{} {
        "body": string(body),
        "status": resp.Status,
        "err": fmt.Sprintf("%+v", err),
        "cart_id": cartguid,
        "fp": fp,
      },
    })


    if len(sresp.Errors) > 0 {
      site_err := sresp.Errors[0]
      errmsg := sresp.Errors[0].Message
      if strings.Contains(errmsg, "We suggest you check again or try another card") || strings.Contains(errmsg, "Invalid Payment details") || strings.Contains(errmsg, "We suggest you try again or use another payment method") {
        go func() {
          defer func() {
            recover()
          }()
          if cart, _ := t.FtlGetCart(); cart != nil && len(cart.Entries) > 0 {
            if cj, err := json.Marshal(cart); err == nil {
              t.ApiPost(fmt.Sprintf("/hcache?k=cart%s%s", "%2F", cart.Entries[0].Product.BaseProduct), cj)
            }
          }
        }()

        return retry.Unrecoverable(errCardDeclined)
      } else if strings.Contains(errmsg, "has to be removed from cart") || strings.Contains(errmsg, "Cart is empty") || strings.Contains(errmsg, "Product Reservation has expired") {
        if strings.Contains(errmsg, "Reservation") {
          t.SetTelemetryContext("cj", "0")
        }
        t.RemoveCookie("cart-guid")
        t.StatusCh <- StatusCartJacked
        return retry.Unrecoverable(retry.Unrecoverable((ErrCartJacked)))
      } else if strings.Contains(errmsg, "Token does not match session") {
        t.StatusCh <- StatusCartExpired
        return retry.Unrecoverable(ErrSessionExpired)
      } else if strings.Contains(errmsg, "Item in your Cart is limited to 1 per Customer") {
        // TODO profile rotation
        t.StatusCh <- StatusProfileAlreadyCheckedOutItem
        return TaskErrUnrecoverable(ErrSessionExpired)
      } else if strings.Contains(errmsg, "Product Reservation has expired")  {
        t.StatusCh <- StatusCartExpired
        // TODO dont fully restart task, go back to atc
        return TaskErrUnrecoverable(ErrSessionExpired)
      } else if strings.Contains(errmsg, "Validation error") && site_err.Code == 0 {
        t.StatusCh <- StatusCartExpired
        // TODO dont fully restart task, go back to atc
        return TaskErrUnrecoverable(ErrSessionExpired)
      } else {
        return ErrRetrying
      }
    } else {
      var order FtlOrderSummary
      if err := json.Unmarshal(body, &order); err == nil {
        if len(order.Order.Code) > 0 {
          if cart, _ := t.FtlGetCart(); cart != nil && len(cart.Entries) > 0 {
            go func() {
              defer func() {
                recover()
              }()
              if cj, err := json.Marshal(cart); err == nil {
                t.ApiPost(fmt.Sprintf("/hcache?k=cart%s%s", "%2F", cart.Entries[0].Product.BaseProduct), cj)
              }
            }()
          }
          t.SetTelemetryContext("ordernum", order.Order.Code)
        }
      }
    }
    return err
  }, retry.Attempts(2), retry.Delay(0*time.Second))
}


func (t *CheckoutTask) FtlGetPayToken() (*string, error) {
  url_, err := url.Parse(t.FtlBase() + fmt.Sprintf("api/payment/token?timestamp=%d", timeMillis()))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "accept",
    "user-agent",
    "x-fl-request-id",
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
    {"x-fl-request-id", uuid.New().String()},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer",t.FtlBase() + "cart"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := t.FtlWebDoRequest(t.makeReq("GET", url_, &headers, &headerOrder, nil), "Get Payment Token")
  if err != nil {
    return nil, err
  }
  var respMap map[string]string
  err = readRespJsonDst(resp, &respMap)
  if err != nil {
    return nil, err
  }
  if paytoken, ok := respMap["paymentToken"]; ok {
    return &paytoken, nil
  } else {
    return nil, errors.New("response missing payment token")
  }
}

func (t *CheckoutTask) FtlApplyPaypalToCart(ppresp *PaypalCheckoutSuccess, csrf string, typ string) (error) {
  url_, err := url.Parse(fmt.Sprintf("https://" + t.Url.Host + "/api/users/carts/current/paypal?timestamp=%d", timeMillis()))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-csrf-token",
    "user-agent",
    "x-fl-request-id",
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
    {"x-csrf-token", csrf},
    {"x-fl-request-id", uuid.New().String()},
    {"content-type", "application/json"},
    {"origin", "https://www.footlocker.com"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", "https://www.footlocker.com/cart"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  bodyStructure := map[string]interface{} {
  "checkoutType": "EXPRESS",
  "nonce": ppresp.Nonce,
  "reservationTTLSeconds": 300,
  "details": map[string]interface{} {
    "reservationTTLSeconds": 300,
    "email": t.Profile.Email,
    "firstName": t.Profile.ShippingAddress.FirstName,
    "lastName": t.Profile.ShippingAddress.LastName,
    "payerId": ppresp.PayerId,
    "shippingAddress": map[string]interface{} {
      "recipientName": t.Profile.ShippingAddress.FirstName + " " + t.Profile.ShippingAddress.LastName,
      "line1": t.Profile.ShippingAddress.Address1,
      "line2": t.Profile.ShippingAddress.Address2,
      "city":  t.Profile.ShippingAddress.City,
      "state":t.Profile.ShippingAddress.State,
      "postalCode": t.Profile.ShippingAddress.Zip,
      "countryCode": t.Profile.ShippingAddress.Country,
      "countryCodeAlpha2": t.Profile.ShippingAddress.Country,
      "locality": t.Profile.ShippingAddress.City,
      "region": t.Profile.ShippingAddress.State,
      "streetAddress": t.Profile.ShippingAddress.Address2,
      "extendedAddress": t.Profile.ShippingAddress.Address2,
    },
    "phone": t.Profile.ShippingAddress.NordPhone(),
    "countryCode": t.Profile.ShippingAddress.Country,
    "billingAddress": map[string]interface{} {
     "recipientName": t.Profile.BillingAddress.FirstName + " " + t.Profile.BillingAddress.LastName,
      "line1": t.Profile.BillingAddress.Address1,
      "line2": t.Profile.BillingAddress.Address2,
      "city":  t.Profile.BillingAddress.City,
      "state":t.Profile.BillingAddress.State,
      "postalCode": t.Profile.BillingAddress.Zip,
      "countryCode": t.Profile.BillingAddress.Country,
      "countryCodeAlpha2": t.Profile.BillingAddress.Country,
      "locality": t.Profile.BillingAddress.City,
      "region": t.Profile.BillingAddress.State,
      "streetAddress": t.Profile.BillingAddress.Address2,
      "extendedAddress": t.Profile.BillingAddress.Address2,
    },
  },
  "type": typ,
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  if t.OneCheckoutPerProf && t.CheckHit != nil && t.CheckHit(t.Url.Host, t.Profile.Name) > 0 {
    t.StatusCh <- Status{Message: "Stopping Task, Already Hit with Profile"}
    return retry.Unrecoverable(retry.Unrecoverable(errors.New("stopping task, already hit")))
  }
  req := t.makeReq("POST", url_, &headers, &headerOrder, &body)
  req.IgnoreBody = true
  req.IgnoreBodyExceptStatus = []int{403}
  resp, err := t.FtlWebDoRequest(req, "Ordering")
  DiscardResp(resp)
  if resp != nil && resp.StatusCode != 200 {
    return ErrPaypalCheckoutFailed
  }
  return err
}

func (t *CheckoutTask) FtlPaypalOrder(sesh *FtlSession, amount float64) (*PaypalCheckoutSuccess, error) {
  origDelay := t.Delay
  t.Delay = 0
  defer func() {
    t.Delay = origDelay
  }()

  // t.StatusCh <- StatusStartingPpCheckout

  t.StartCheckoutTime = timeMillis()
  // fmt.Printf("%+v\n", cart)
  var cartUrl, _ = url.Parse(t.FtlBase() + "cart")
  token, err := t.FtlGetPayToken()
  if err != nil {
    return nil, err
  }
  psesh := NewPaypalCheckoutSession(*token, t, cartUrl)
  psesh.Amount = amount
  presp, err := psesh.Checkout()
  t.LogDebug("%+v\n", err)
  if err != nil {
    return nil, err
  }

  return presp, nil
}

var FtlStaticCookies = []*http.Cookie {
  &http.Cookie {
    Name: "at_check",
    Value: "true",
  },
}

var FtlStaticHeaders = [][2]string {
  // {"sec-ch-ua", `"\\Not;A\"Brand";v="99", "Google Chrome";v="85", "Chromium";v="85"`},
  // {"sec-ch-ua-mobile", "?1"},
}

var FtlCheckoutCookies = []string{
  "JSESSIONID",
  "waiting_room",
  "cart-guid",
}

func (t *CheckoutTask) ApplySession(sesh *TaskSession) {
  t.LogDebug("%+v", sesh)
  t.UserAgent = sesh.UserAgent
  t.LastSessionSnapshot = sesh

  ck := []*http.Cookie{}
  for _, ckie := range FtlCheckoutCookies {
    if cv := getCookie(ckie, t.BaseUrl, t.client.Jar); cv != nil {
      ck = append(ck, cv)
    }
  }

  t.LogDebug("%+v", sesh.Cookies)
  t.client.Jar.SetEntries(sesh.Cookies)
  for _, cv := range ck {
    // if !common.Any(SESSION_CKIE_BLACKLIST, func(s string) bool { return s == cv.Name}) {
    t.SetCookie(cv)
    // }
  }
  t.LogDebug("%+v", t.client.Jar)
}

func (t *CheckoutTask) GetBareSessionBytes() []byte {
 bareTaskSession := TaskSession{Host: t.Url.Host, Proxy: t.ProxyStr}
 bareTaskSessionBytes, _ := json.Marshal(bareTaskSession)
 return bareTaskSessionBytes
}
func (t *CheckoutTask) GetSessionStats() (*RushTaskSessionStats, error) {
  bareTaskSessionBytes := t.GetBareSessionBytes()

  var err error
  var cjb []byte
  if cjb, err = t.ApiPost("/sesh/count", bareTaskSessionBytes); err == nil {
    var seshStats RushTaskSessionStats
    if err = json.Unmarshal(cjb, &seshStats); err == nil {
      return &seshStats, nil
    }
  }
  return nil, err
}

func (t *CheckoutTask) RemoveBadSession() {
  go func() {
    defer func() {
                    recover()
                  }()
    if t.LastSessionSnapshot != nil {
      if seshb, err := json.Marshal(*t.LastSessionSnapshot); err == nil {
        t.LastSessionSnapshot = nil
        t.ApiPost("/sesh/remove", seshb)
      }
    }
  }()
}

func (t *CheckoutTask) FtlSendSessionSnapshot() {
  go func() {
    defer recover()
    seshsnap := t.TaskSessionSnapshot()
    t.LastSessionSnapshot = &seshsnap
    if seshb, err := json.Marshal(seshsnap); err == nil {
        go t.ApiPost("/sesh/push", seshb)
    }
    seshsnap.Cookies = nil
    if rand.Float32() < 0.05 {
      t.SendTelemetry(map[string]interface{} {
        "event": "ftl_session",
        "ftl_sesh": seshsnap,
      })
    }
  }()
}

func (t *CheckoutTask) FtlSendSessionSnapshotBlk() {
  // go func() {
  //   defer recover()
    seshsnap := t.TaskSessionSnapshot()
    t.LastSessionSnapshot = &seshsnap
    if seshb, err := json.Marshal(seshsnap); err == nil {
        t.ApiPost("/sesh/push", seshb)
    }
  // }()
}
func (t *CheckoutTask) FtlCheckout() error {

  // t.FtlGetSession()
  // for _, ord := range orders {
  //   if false {
  //     t.LogDebug("%s", ord[0])
  //   }
  //   // t.FtlCheckOrder(ord[0], ord[1])
  // }
  // return nil
  // pSplit := strings.Split(t.Url.Path, "/")
  // pid := strings.Split(pSplit[len(pSplit) - 1], ".")[0]

  // t.ElgoogGenNID()
  cfgctx, cancel := context.WithCancel(t.Ctx)
  defer cancel()
  go func() {
    ret := 0
    for {
      wait := time.Duration(rand.Intn(90))*time.Second
      if ret == 0 {
        wait = 0
      }
      select {
      case <-cfgctx.Done():
        return
      case <-time.After(wait):
        if bb, err := t.S3Get("ftl.cfg", 2); err == nil {
          ret += 1
          ftlCfg := map[string]interface{}{}
          if err := json.Unmarshal(bb, &ftlCfg); err == nil {
            for k,v := range ftlCfg {
              if s, ok := t.GetTelemetryContext(k).(string); !ok || len(s) == 0 {
                t.SetTelemetryContext(k,v)
              }
            }
          }
        }
        continue
      }
      time.Sleep(5*time.Second)
    }
  }()
  if os.Getenv("CAP_TEST") == "1" {
    soln, err := t.SolveGeetestWs("https://geo.captcha-delivery.com/captcha/?initialCid=AHrlqAAAAAMABn9uMAGvAC0ArsxS9A==&hash=A55FBF4311ED6F1BF9911EB71931D5&t=fe&s=17434&cid=7NBe-0Y_TMVbaM5v85yJ7N-x-eyEvBX8B86~5p47t0yGVzS7tngdp8fdTmd~wUPH3DejcH0vkFIx99QCi-XUTFxk147XTta55OrFrjMioC")
    fmt.Printf("%+v %+v", soln, err)
    // t.initClient()
    // t.UserAgent = DD_UAS[rand.Intn(len(DD_UAS))]
    // t.ElgoogGenNid()
    // t0 := time.Now()
    // _, err := t.SolveCaptcha(FTL_SITEKEY, "https://geo.captcha-delivery.com/robots.txt")
    // if err != nil {
    //   fmt.Printf("err=%+v\n", err)
    //   return err
    // }
    // fmt.Printf("solvtime=%v\n", time.Now().Sub(t0))
    return nil
  }
  // t.LogDebug("rotate res=%+v", t.RotateProxy())
  // t.Url.Path = strings.ToLower(t.Url.Path)
  qstr := t.Url.Query().Get("qty")
  var qty int = 1
  if qtymaybe, err := strconv.Atoi(qstr); err == nil {
    qty = qtymaybe
  }
  enableWaitForRelease := os.Getenv("RELEASE_TIMER") != "0"
  isPayPalFallbackEnabled := t.Url.Query().Get("pp") == "2" || t.Url.Query().Get("PP") == "2"
  isPayPalEnabled := isPayPalFallbackEnabled || os.Getenv("PAYPAL") == "1" || t.Url.Query().Get("pp") == "1" || t.Url.Query().Get("PP") == "1"
  isFarmEnabled := os.Getenv("FARM") == "1" || t.Url.Query().Get("farm") == "1"

  if os.Getenv("DISCOVER") == "1" {
    for {
      t.FtlGetProductApi()
    }
  }

  cartLoop, _ := strconv.Atoi(os.Getenv("CART_LOOP"))
  isSessionStoreEnabled := os.Getenv("SESH") != "0"
  t.Url.RawQuery = ""
  var hasWaitedForRelease bool
  var tsesh *TaskSession
	var err error
  origDelay := t.Delay

  err = t.RetryTask(func() error {
    defer func() {
      if t.GetTelemetryContext("dd_last") == "allow" {
        t.FtlSendSessionSnapshot()
      } else if t.GetTelemetryContext("dd_last") == "403" {
        t.RemoveBadSession()
      }
    }()
    var specificProduct *FtlPdpSize
    var ppAppliedTo string
    t.initClient()


    // t.UserAgent, err = t.getUserAgent()
    t.UserAgent = DD_UAS[rand.Intn(len(DD_UAS))]
    if os.Getenv("USER_AGENT") != "" {
      t.UserAgent = os.Getenv("USER_AGENT")
    }
    if !t.UseIpv6 {
        t.DnsCacheMut.Lock()
        t.DnsCache = nil
        t.DnsCacheMut.Unlock()
          }
    // t.ElgoogStartTrustBuilder()
    // t.SetCookie(&http.Cookie{Name: "waiting_room", Value: "ZGVjPWFsbG93JmV4cD0xNjA3MzY0MTg0JnVpZD0xNzQuMTkzLjE0MS4xMDEma2lkPWtleTEmc2lnPTB4NDc3NzQzZDRkNDEwY2ZmZjkxODE4MGM1YjAyNTk5YmM0ZTE0YWQzM2E1ZWQxZjI5MmQxNTdkNDQxNDY5MDliZA=="})
    product, err := t.FtlGetProduct()
    if err != nil {
      return err
    }
    // t.SetCookie(&http.Cookie{Name: "waiting_room", Value: "ZGVjPWFsbG93JmV4cD0xNjA3MzY0MTg0JnVpZD0xNzQuMTkzLjE0MS4xMDEma2lkPWtleTEmc2lnPTB4NDc3NzQzZDRkNDEwY2ZmZjkxODE4MGM1YjAyNTk5YmM0ZTE0YWQzM2E1ZWQxZjI5MmQxNTdkNDQxNDY5MDliZA=="})
    // fmt.Printf("%+v\n", product.Sizes())



    if isSessionStoreEnabled && !isFarmEnabled && len(t.ProxyStr) > 0 {

      resp, err := t.ApiPost("/sesh/get", t.GetBareSessionBytes())
      if err == nil {
        var s TaskSession
        if err = json.Unmarshal(resp, &s); err == nil && len(s.UserAgent) > 0 {
          tsesh = &s
          t.LogDebug("using sesh: %+v", tsesh)
        }
      } else {
        t.LogDebug("sesh get err %+v", err)
        if serr, ok := err.(*StatusCodeRejectedError); ok {
          if serr.StatusCode == 401 {
            return ErrRetrying
          }
        }
      }
    }

    if tsesh != nil {
      t.ApplySession(tsesh)
    }


    t.SetCookie(&http.Cookie{Name: "datadome", Value: "Bu7m~WK~8LLYxeZwlyPLcPBrbfqWWYRpJKggF.u~eqTPBr1hic8pizhnxR7YB8d_yqt6vnjFojoBEaX3BFbwnkEjoMoYBecmoWlIQMuPlc"})

    // for {

    if os.Getenv("DDTEST") == "1" {
      if derr := t.DatadomeSendPayload(t.BaseUrl.String(), "", ""); derr != nil {
        t.LogDebug("DatadomeSendPayload err=%+v", derr)
        if !t.shouldRetry(derr) {
          return derr
        }
      }
    }



    go func() {
      if t.UseIpv6 {
        return
      }
    // func() {

      defer func() {
      t.LogDebug("%+v", recover())
    }()
      rr := rand.Float32()
      ref := t.Url.String()
      if 0.33 < rr && rr < 0.66 {
        ref = ""
      } else if rr > 0.66 {
        ref = t.BaseUrl.String()
      }

      if derr := t.DatadomeSendPayload(t.BaseUrl.String(), ref, ""); derr != nil {
        t.LogDebug("DatadomeSendPayload err=%+v", derr)
        if !t.shouldRetry(derr) {
          return
        }
      }

      time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)

      rr = rand.Float32()
      ref = t.Url.String()
      if 0.33 < rr && rr < 0.66 {
        ref = ""
      } else if rr > 0.66 {
        ref = t.BaseUrl.String()
      }
      if derr := t.DatadomeSendPayload(t.Url.String(), ref, ""); derr != nil {
        t.LogDebug("DatadomeSendPayload err=%+v", derr)
        if !t.shouldRetry(derr) {
          return
        }
      }
    }()

    for _, ckie := range FtlStaticCookies {
      t.SetCookie(ckie)
    }

    var cart *FtlCart
    var sesh *FtlSession
    // var pdpResp *FtlProductPdpResponse
    // t.StatusCh <- StatusStartingSession
    sesh, err = t.FtlGetSession()
    if err != nil {
      if os.Getenv("HALTERR") == "1" {
        panic("halt")
      }
      return err
    }


    var paypal bool
    err = t.Retry(func() error {
      err = t.Retry(func() error {
        if (t.GetTelemetryContext("cj") == "1") && t.GetCookieValue("cart-guid") == "" {
          t.FtlAtc(fmt.Sprintf("%d", rand.Intn(9999999)), sesh.Data.CsrfToken, 0)
        }
        if time.Now().Sub(sesh.Time) > 1*time.Minute {
          if ssesh, err := t.FtlGetSession(); err == nil {
            sesh = ssesh
          }
        }

        ddSolvCtx, ddSolvCancel := context.WithCancel(t.ClientCtx)
        defer ddSolvCancel()
        go func() {
          if t.UseIpv6 {
            return
          }
          defer func() {
            recover()
          }()
          var ref = t.Url.String()
          t.DatadomeSendPayload(ref, ref, t.GetCookieValue("datadome"))
          for {
            time.Sleep(3*time.Second)
            if rand.Float32() < 0.5 {
              ref = ""
            }
            td := time.Duration(rand.Intn(666))*time.Second

            select {
            case <-ddSolvCtx.Done():
              return
            case <-time.After(td):
              // TODO randomize url, ref, add context
              if derr := t.DatadomeSendPayload(t.Url.String(), ref, t.GetCookieValue("datadome")); derr != nil {
                t.LogDebug("DatadomeSendPayload err=%+v", derr)
                if !t.shouldRetry(derr) {
                  return
                }
              }
            }
          }
        }()



        // if prodPageState != nil{
        //   if timeUntilRelease > 0 {
        //     time.Sleep(time.Duration(timeUntilRelease) * time.Millisecond)
        //   }
        // }

        // fmt.Printf("%+v\n", prodPageState)
        // if os.Getenv("DELAY") == "" {
        //   t.Delay = 0
        //   defer func() {
        //     if t.Delay == 0 {
        //       t.Delay = origDelay
        //     }
        //   }()
        // }

        // go func() {

        // }()
        t.LogDebug("PROD %+v", product)
        t.LogDebug("RELTIME %+v %+v", product.ReleaseTime())
        if enableWaitForRelease && !hasWaitedForRelease {
          if releaseTime := product.ReleaseTime(); !releaseTime.IsZero() {
            // releaseTime = time.Now().Add(10*time.Second)
            if os.Getenv("TEST_RELEASE_30MIN") == "1" {
              releaseTime = time.Now().Add(1*time.Second)
            }

            timeUntilRelease := releaseTime.Sub(time.Now()) - time.Duration(rand.Intn(600))*time.Second

              t.LogDebug("%+v\n", timeUntilRelease)
            if timeUntilRelease > 0 {

              // TODO above verif tsesh
              // if not using recent prefarmed sesh , loop prod atc till captcha, and solve, with deadline of relase time
              t.StatusCh <- Status{Message: fmt.Sprintf("Waiting Until Auto-Stagger-Start Before %s EST", releaseTime.In(ESTZ).Format("15:04"))}
              // go func() {
              //   ct, _ := context.WithCancel(t.ClientCtx)
              //   defer func() {
              //     recover()
              //   }()
              //   xxx := 0
              //   for {
              //     if ct.Err() != nil {
              //       return
              //     }
              //     if xxx > 64 || hasWaitedForRelease || releaseTime.Sub(time.Now()) < 120*time.Second || t.Metrics.GetUint64("dd:cap_pass") > 0 || t.Metrics.GetUint64("dd:cap_solve") > 0 {
              //       return
              //     }


              //     code := "8107129"
              //     if product != nil && product.Sizes() != nil && len(*product.Sizes()) > 0 {
              //       code = (*product.Sizes())[rand.Intn(len((*product.Sizes())))].Code
              //     }
              //     cart, err := t.FtlAtc(code, sesh.Data.CsrfToken, qty)
              //     t.RemoveCookie("cart-guid")
              //     isRespNot403 := false
              //     if serr, ok := err.(*StatusCodeRejectedError); ok {
              //       if serr.StatusCode != 403 {
              //         isRespNot403 = true
              //       }
              //     }
              //     if cart != nil || errors.Is(err, ftlErrOutOfStock) || isRespNot403 {
              //       t.StatusCh <- Status{Message: fmt.Sprintf("Waiting for Release (%s EST)", releaseTime.In(ESTZ).Format("15:04"))}
              //       t.FtlSendSessionSnapshot()
              //       return
              //     } else if !t.shouldRetry(err) {
              //       return
              //     }
              //     select {
              //       case <-time.After(time.Duration(rand.Intn(30))*time.Second):
              //         continue
              //       case <-ct.Done():
              //         return
              //     }

              //   }
              // }()

              // TODO replace time.Sleep with context aware delay
              for !hasWaitedForRelease && releaseTime.Sub(time.Now()) > 30*time.Second {
                select {
                case <-t.Ctx.Done():
                  return nil
                  break
                case <-time.After(timeUntilRelease):
                  hasWaitedForRelease = true
                  t.StatusCh <- Status{Message: "Preparing for Release"}
                  continue
                  break
                }
                time.Sleep(500*time.Millisecond)
              }
              hasWaitedForRelease = true
            }
          }
        }


        if sesh != nil && (t.GetCookieValue("cart-guid") == "" && rand.Float32() < 1) {
          t.FtlAtc(fmt.Sprintf("%d", rand.Intn(9999999)), sesh.Data.CsrfToken, 0)
          if t.GetCookieValue("cart-guid") != ppAppliedTo && sesh != nil && len(sesh.Data.CsrfToken) > 0 {
            err = t.FtlApplyPaypalToCart(&PaypalCheckoutSuccess{ Nonce: uuid.New().String(), PayerId: "AAAAAAAAAAAAA"}, sesh.Data.CsrfToken, "PayPalAccount")
            if err == nil {
              ppAppliedTo = t.GetCookieValue("cart-guid")
            }
          }
        }

        err = t.Retry(func() error {
          if product != nil && product.Sizes() != nil && len(*product.Sizes()) > 0 {
            return nil
          }
          pproduct, err := t.FtlGetProduct()
          if err == nil {
            if pproduct != nil && pproduct.Sizes() != nil && len(*pproduct.Sizes()) > 0 {
              product = pproduct
              return nil
            }
            return ErrRetrying
          } else {

          }
          return err
        }, retry.Delay(t.Delay), retry.Attempts(1e6))
        if err != nil || product == nil {
          return err
        }

        t.StartAtcTime = timeMillis()
        if t.GetTelemetryContext("dyndelay") == "1" {
          dctx, cartcachecancel := context.WithCancel(t.ClientCtx)
          defer cartcachecancel()
          go func() {
            for dctx.Err() == nil {
              select {
              case <-dctx.Done():
                t.Delay = origDelay
                return
            // case <-time.After(1*time.Minute):
              case <-time.After(time.Duration(rand.Intn(60))*time.Second): // TODO push not pull
                go func() {
                  defer func() {
                    t.LogDebug("recover_dyndelay %+v", recover())
                  }()
                  if cartb, err := t.ApiGet(fmt.Sprintf("/hcache?k=cart%s%s", "%2F", product.Product().StyleSku)); err == nil {
                    var cart FtlCart
                    if err := json.Unmarshal(cartb, &cart); err == nil {
                      t.LogDebug("last cart: %v ago", time.Now().Sub(cart.Time))
                      t.LogDebug("origDelay %v t.Delay %v", origDelay, t.Delay)
                      if t.GetTelemetryContext("low_data") != "1" && !cart.Time.IsZero() && time.Now().Sub(cart.Time) > LOW_DELAY_TOUT {
                        t.StatusCh <- Status{Message: "Switching to Low-Delay"}
                        t.SetTelemetryContext("low_data", "1")
                      } else if time.Now().Sub(cart.Time) < LOW_DELAY_TOUT && !cart.Time.IsZero() {
                        t.DelTelemetryContext("low_data")
                        t.WakeUp <- true
                      }
                    }
                  } else {
                    t.DelTelemetryContext("low_data")
                  }
                }()
              }
              time.Sleep(50*time.Millisecond)
            }
          }()
        }
        return t.Retry(func() error {
          rand.Shuffle(len(*product.Sizes()), func(i, j int) {
                     (*product.Sizes())[i], (*product.Sizes())[j] = (*product.Sizes())[j], (*product.Sizes())[i] })
          sort.SliceStable(*product.Sizes(), func(i, j int) bool {
            if ((*product.Sizes())[i].IsDisabled) || rand.Float32()<0.5 {
              return false
            }
            return true
          })

          for _, sz := range (*product.Sizes())[:] {
                 // if !disableStockCheck && product.StockLevelStatus != "inStock" { continue }
                 size := strings.ToLower(sz.Name)
                 sizeFl, _ := strconv.ParseFloat(size, 64)
                 sizeMatch := (len(t.Sizes) == 0 || Any(t.Sizes, func(desiredSize string) bool {
                         desiredSizeFl, err := strconv.ParseFloat(desiredSize, 64)
                         return (err == nil && fmt.Sprintf("%02.1f", desiredSizeFl) == fmt.Sprintf("%02.1f", sizeFl)) || strings.ToLower(desiredSize) == strings.ToLower(size)
                 }))
                 if sizeMatch && len(sz.Code) > 0 {
          specificProduct = &sz
                         break
                 }
          }
          if specificProduct == nil {
            if len(t.Sizes) == 0 {
              specificProduct = &(*product.Sizes())[rand.Intn(len((*product.Sizes())))]
            } else {
              t.StatusCh <- StatusSizeUnavailable
              pproduct, err := t.FtlGetProduct()
              if err == nil {
                if pproduct != nil && pproduct.Sizes() != nil && len(*pproduct.Sizes()) > 0 {
                  product = pproduct
                }
              }
              return ErrRetrying
            }
          }
          if cartLoop >= 1 {
            if time.Now().Sub(sesh.Time) > 1*time.Minute {
              if ssesh, err := t.FtlGetSession(); err == nil {
                sesh = ssesh
              }
            }
            // code = code + "00"
            cartLoop -= 1
            // t.RemoveCookie("cart-guid")
            if os.Getenv("DD_ATC_FLAG") == "1" {
              t.RemoveCookie("datadome")
            }
          }

          t.StatusCh <- StatusWaitForCart
          if (t.GetTelemetryContext("cj") == "1") && (t.GetCookieValue("cart-guid") == "" || rand.Float32() < 0.1) {
            t.RemoveCookie("cart-guid")
            t.FtlAtc(fmt.Sprintf("%d", rand.Intn(9999999)), sesh.Data.CsrfToken, 0)
          }

          t.LogDebug("%v\n",specificProduct)
        	cart, err = t.FtlAtc(specificProduct.Code, sesh.Data.CsrfToken, qty)
          // if t.Delay == 0 {
          //   t.Delay = origDelay
          // }
          t.Metrics.Incr("cart_attempts")
          // if rand.Float32() > 0.05 {
          //   return errors.New("test")
          // } else {
            if cart != nil && len(cart.Entries) > 0 && t.GetCookieValue("cart-guid") != "" {
              // t.FtlGetCart()
              t.Metrics.Incr("carts")
              t.StatusCh <- StatusCarted

              go func() {
                defer func() {
                      recover()
                    }()
                t.SendTelemetry(map[string]interface{} {
                  "event": "ftl_cart",
                  "ftl_cart": map[string]interface{} {
                    "cart_selected": cart.Entries[0].Product.BaseOptions[0],
                    // "skus": skus,
                    // "respheaders":
                  },
                })
                t.LogDebug("%+v", cart)
                if t.GetTelemetryContext("cj") != "1" {
                if cj, err := json.Marshal(cart); err == nil {
                  t.ApiPost(fmt.Sprintf("/hcache?k=cart%s%s", "%2F", cart.Entries[0].Product.BaseProduct), cj)
                }
                }
              }()
              // if len(cart.Entries) > 1 {
              //   for _, entry := range cart.Entries {

              //   }
              // }
              if cartLoop >= 1 {
                return ErrRetrying
              }
              return nil
            } else {
              // t.RemoveCookie("cart-guid")
              return ErrRetrying
            }
          // }
          return err

        }, retry.Attempts(512), retry.Delay(origDelay), retry.OnRetry(func (_ uint, err error) {
    t.LogDebug("retry %s\n", err.Error())
    t.resetClient(false)
  }))
      }, retry.Delay(origDelay))
      if err != nil {
        return err
      }

      if !t.UseIpv6 {
      t.DnsCacheMut.Lock()
      t.DnsCache = nil
      t.DnsCacheMut.Unlock()
    }
    var ppOrder *PaypalCheckoutSuccess


    if isPayPalEnabled {
      t.StatusCh <- Status{Message: "PayPal Checkout 1/2"}
      ppOrder, err = t.FtlPaypalOrder(sesh, cart.TotalPrice.Value)
      if err != nil {
        go t.SendTelemetry(map[string]interface{} {
          "event": "paypal_checkout_error",
          "paypal_checkout_failure": fmt.Sprintf("%+v", err),
        })
        if errors.Is(err, ErrPaypalCardRejected) {
          t.StatusCh <- StatusPaypalCardRejected
          err = ErrCardRejected
          if !isPayPalFallbackEnabled {
            return retry.Unrecoverable(err)
          }
        } else if t.shouldRetry(err) {
          t.StatusCh <- StatusPaypalCheckoutFailed
          if !isPayPalFallbackEnabled {
            return retry.Unrecoverable(err)
          }
        } else if !isPayPalFallbackEnabled {
          return err
        }
      }
    }

    if os.Getenv("PP_NO_ORDER") == "1" {
      fmt.Println("skipping order")
      return nil
    }

    if isPayPalEnabled && ppOrder != nil {
      // err = t.Retry(func() error {
      t.FtlEnsureValidAbck()
      t.StartCheckoutTime = timeMillis()
      // if cart.TotalPrice.Value != predictedPrice { } //
      t.StatusCh <- Status{Message: "PayPal Checkout 2/2"}
      err = t.FtlPaypalFinishOrder(ppOrder, sesh.Data.CsrfToken)
      if err != nil {
        go t.SendTelemetry(map[string]interface{} {
          "event": "paypal_order_failure",
          "paypal_order_failure": fmt.Sprintf("%+v", err),
        })
        if errors.Is(err, ErrPaypalCardRejected) {
          t.StatusCh <- StatusPaypalCardRejected
          err = ErrCardRejected
          if !isPayPalFallbackEnabled {
            return retry.Unrecoverable(err)
          }
        } else if errors.Is(err, errCardDeclined) {
          t.StatusCh <- Status{Message: "PayPal Declined"}
          paypal = false
        } else if t.shouldRetry(err) {
          t.StatusCh <- StatusPaypalCheckoutFailed
          if !isPayPalFallbackEnabled {
            return retry.Unrecoverable(err)
          }
        } else if !isPayPalFallbackEnabled {
          return err
        }
      } else {
        paypal = true
      }
      // }, retry.Attempts(2))
    } else if isPayPalEnabled && ppOrder == nil && !isPayPalFallbackEnabled {
      t.StatusCh <- StatusPaypalCheckoutFailed
      return retry.Unrecoverable(ErrPaypalCheckoutFailed)
    }

    if !isPayPalEnabled || (!paypal && isPayPalFallbackEnabled) {
      rettt := 0
      t.StartCheckoutTime = timeMillis()
      return t.Retry(func() error {
        defer func() {
          rettt += 1
        }()
      	t.StatusCh <- Status{Message: "Starting Checkout"}

        if rettt > 0  {
          if cart, err := t.FtlGetCart(); cart != nil && err == nil && len(cart.Entries) == 0 {
            t.StatusCh <- StatusCartJacked
            return retry.Unrecoverable(ErrRetrying)
          }
        }
        if ppAppliedTo != t.GetCookieValue("cart-guid") {
          err = t.FtlApplyPaypalToCart(&PaypalCheckoutSuccess{ Nonce: uuid.New().String(), PayerId: "AAAAAAAAAAAAA"}, sesh.Data.CsrfToken, "PayPalAccount")
          if err != nil {
            return err
          }
        }
      	// err = t.FtlStartCheckout(sesh.Data.CsrfToken)
      	// if err != nil {
      	// 	return err
      	// }

        // if os.Getenv("EXPRESS") != "0" {



        // } else {

        // 	t.StatusCh <- Status{Message: "Setting Shipping Address"}
        // 	// shipping
        // 	err = t.FtlSetShipping(sesh.Data.CsrfToken)
        // 	if err != nil {
        // 		return err
        // 	}

        //   // t.FtlEnsureValidAbck()

        // 	t.StatusCh <- Status{Message: "Setting Billing Address"}
        // 	err = t.FtlSetBilling(sesh.Data.CsrfToken)
        // 	if err != nil {
        // 		return err
        // 	}
        // }

        t.StatusCh <- Status{Message: "Ordering"}
        t.StartPaymentTime = timeMillis()
        if t.Ctx.Err() != nil || t.ClientCtx.Err() != nil {
          return retry.Unrecoverable(ErrRetrying)
        }
        err = t.FtlAdyenOrder(sesh)
        if errors.Is(err, errCardDeclined) {
          return retry.Unrecoverable(retry.Unrecoverable(err))
        }
        return err
      }, retry.Attempts(3), retry.OnRetry(func (_ uint, err error) {

          if ssesh, err := t.FtlGetSession(); err == nil {
            sesh = ssesh
          }

    t.LogDebug("retry %s\n", err.Error())
    t.resetClient(false)
  }))
    }
    return ErrRetrying
  })

    t.LogDebug("checkout result: %+v\n", err)
    whFields := [][2]string{
      {"Website", t.Url.Host},
    }
    if product != nil {
      whFields = append(whFields, [2]string{"Product", product.Name()}, [2]string{"Style", product.Style()})
    }
    if specificProduct != nil {
      whFields = append(whFields, [2]string{"Size", specificProduct.Name})
    }
    if paypal {
      whFields = append(whFields, [2]string{"Payment Processor", "PayPal"})
    }
    if errors.Is(err, errCardDeclined) {
      t.StatusCh <- StatusCardDeclined
      go t.DeclineWebhookFields(whFields)
      if os.Getenv("LOOP") != "1" {
        return nil
      }
    } else if err == nil {
      t.StatusCh <- StatusCooked
      go t.SuccessWebhookFields(whFields)
    }
    return err
  }, retry.OnRetry(func (_ uint, err error) {
    t.LogDebug("retry %s\n", err.Error())
    t.resetClient(false)
  }), retry.MaxDelay(origDelay), retry.DelayType(retry.DefaultDelayType), retry.Delay(origDelay), retry.LastErrorOnly(true))
	// t.LogDebug("%+v", cart)
	// billing

  return err
}

package task

import (
  "context"
  "fmt"
  "github.com/pkg/errors"
  "github.com/avast/retry-go"
  "math/rand"
  "net/url"
  "os"
  "log"
  "strings"
  "rush/net/http"
  "time"
  "strconv"
)

var ErrNoCookiesAvailable = errors.New("No cookies available")

func (t *CheckoutTask) EnsureValidAbck(akUrl string, hasValidAbckFn func(abck string) bool) error {
  // var abck string
  // abck, COOKIES = COOKIES[0], COOKIES[1:]
  // t.LogDebug("set %s", abck)
  // t.SetCookie(&http.Cookie{ Name: "_abck", Value: "", MaxAge: -1, Expires: time.Unix(0,0), Path: "/", })
  // t.SetCookie(&http.Cookie{ Name: "_abck", Value: abck, MaxAge: 999999, Expires: time.Now().Add(time.Duration(86400*7)*time.Second), Path: "/", })
  // return nil

  // if hasValidAbckFn(t.GetCookieValue("_abck")) {
  //   return nil
  // }

  // origabck := t.GetCookieValue("_abck")
  // log.Printf("EnsureValidAbck: %s", origabck)

  t.AbckJarMut.Lock()
  if os.Getenv("ABCK") != "" {
    t.AbckJar = [][]*http.Cookie {

      []*http.Cookie {
          &http.Cookie{
          Name: "_abck",
          Value: os.Getenv("ABCK"),
          Secure: true,
          HttpOnly: false,
          MaxAge: 1e6,
          Path: "/",
        // Domain: t.BaseDomainUrl.Host,
       },
     },
    }
    t.RemoveCookie("_abck")
    t.client.Jar.SetCookies(t.BaseDomainUrl, t.AbckJar[0])
    t.SetCookie(t.AbckJar[0][0])
    return nil
  }
  t.LogDebug("usage %v max %v", t.AbckJarUsage, t.AbckJarMaxUsage)
  if t.AbckJarUsage >= t.AbckJarMaxUsage || strings.HasSuffix(t.GetCookieValue("_abck"), "~0~-1~-1") {
    if len(t.AbckJar) > 1 {
      t.AbckJar = t.AbckJar[1:]
    } else {
      t.AbckJar = nil
    }
    t.AbckJarUsage = 0
  }
  if len(t.AbckJar) > 0 {
    // fmt.Printf("entry pre %+v\n", GetCookieEntry("_abck", t.client))
    // fmt.Printf("entry post %+v\n", GetCookieEntry("_abck", t.client))
    // } else {
    // }
    t.AbckJarUsage += 1
    t.LogDebug("set _abck=%+v\n", t.AbckJar[0])
    t.LogDebug("pre set abck: %v", t.GetCookieValue("_abck"))
    for _, ckie := range t.AbckJar[0] {
      // if ckie.Name == "bm_sz" {
      //   oldval := strings.Split(t.GetCookieValue(ckie.Name), "~")
      //   newval := strings.Split(ckie.Value, "~")
      //   t.LogDebug("old/new bm_sz %+v %+v", oldval, newval)
      //   if len(oldval) > 1 && len(newval) > 1 && oldval[0] == newval[0] {
      //     t.LogDebug("skipping bm_sz")
      //     // let task keep its bm_sv which will be more up to date
      //     continue
      //   }
      // } else if ckie.Name == "akavpau_wwwvpfs" {
      //   // oldval := strings.Split(t.GetCookieValue(ckie.Name), "~")
      //   // newval := strings.Split(ckie.Value, "~")
      //   // t.LogDebug("old/new akavpau_wwwvpfs %+v %+v", oldval, newval)
      //   // if len(oldval) > 1 && len(newval) > 1 && strings.Compare(newval[0], oldval[0]) <= 0 {
      //   //   t.LogDebug("skipping akavpau_wwwvpfs")
      //   //   continue
      //   // }
      // }
      t.RemoveCookie(ckie.Name)
      if strings.HasPrefix(ckie.Domain, "www") {
        t.client.Jar.SetCookies(t.BaseUrl, []*http.Cookie { ckie })
      } else {
        t.client.Jar.SetCookies(t.BaseDomainUrl, []*http.Cookie { ckie })
      }
      t.SetCookie(ckie)
    }
    t.LogDebug("post set: %v", t.GetCookieValue("_abck"))
    t.AbckJarMut.Unlock()
    return nil
  }
  t.AbckJarMut.Unlock()

  // t.LogDebug("cloud fallback")
  // if ckie, err := t.GetCloudAbck(t.Url.Host); err == nil {
  //   cckie := [][]*http.Cookie{ { ckie } }

  //   t.RemoveCookie("_abck")
  //   t.client.Jar.SetCookies(t.BaseDomainUrl, cckie[0])
  //   t.SetCookie(cckie[0][0])

  //   t.AbckJarMut.Lock()
  //   if len(t.AbckJar) > 1 {
  //     t.AbckJar = append(cckie, t.AbckJar[1:]...)
  //   } else {
  //     t.AbckJar = cckie
  //   }
  //   t.AbckJarUsage = 1
  //   t.AbckJarMut.Unlock()
  //   return nil
  // }

  if t.Config["abck_validation_disabled"] == "1" {
    return nil
  }

  t.LogDebug("livegen fallback")
  if t.Config["stage"] != "" {
    t.StatusCh <- Status{Message: "Generating Cookie (" + t.Config["stage"] + ")"}
  } else {
    t.StatusCh <- StatusGeneratingCookie
  }
  var err error

  defer func() {
    t.LogDebug("EnsureValidAbck RET: %v %v", t.GetCookieValue("_abck"), err)
  }()

  if _, err = t.GenValidAbck(t.client, t.ClientCtx, t.UserAgent, akUrl, hasValidAbckFn); err == nil {
    if t.AbckJarMaxUsage > 0 {
      if ckie := t.getAbckCookie(t.client); ckie != nil {
        t.LogDebug("fallback GenValidAbck: %#v", ckie)
        cckie := [][]*http.Cookie{ { ckie } }
        t.AbckJarMut.Lock()
        if len(t.AbckJar) > 1 {
          t.AbckJar = append(cckie, t.AbckJar[1:]...)
        } else {
          t.AbckJar = cckie
        }
        t.AbckJarUsage = 1
        t.AbckJarMut.Unlock()
      }
    }
  }
  return err
}

func (t *CheckoutTask) GetCloudAbck(host string) (*http.Cookie, error) {
  if bb, err := t.ApiGetRetries("/abck?host=" + host, 1); err == nil {
    t.LogDebug("cloud abck %s", string(bb))
    return &http.Cookie{
      Name: "_abck",
      Value: string(bb),
      Secure: true,
      HttpOnly: false,
      MaxAge: 1e6,
      Path: "/",
      // Domain: t.BaseDomainUrl.Host,
    }, nil
  } else {
    return nil, err
  }

}
func (t *CheckoutTask) getAbckCookie(client *http.Client) *http.Cookie {
  if ckie := getCookie("_abck", t.BaseDomainUrl, client.Jar); ckie != nil {
    ckie.Secure = true
    ckie.HttpOnly = false
    ckie.MaxAge = 1e6
    ckie.Domain = t.BaseDomainUrl.Host
    return ckie
  }
  return nil
}

var AbckGenCookieBlocklistUrl, _ = url.Parse("https://rushaio.s3.amazonaws.com/abck_gen_cookie_blocklist.txt")
var AbckGenToggleUrl, _ = url.Parse("https://rushaio.s3.amazonaws.com/abck_gen_enabled.txt")
var ErrAbckGenCookieBlocklistEmpty = errors.New("Empty blocklist, use default")

func (t *CheckoutTask) getAbckJarEnabled() (bool) {
  resp, err := t.doReq(t.s3client, t.makeGetReq(AbckGenToggleUrl, nil))
  if err != nil {
    DiscardResp(resp)
    return true
  }
  enabled, err := readBody(resp)
  if err != nil {
    return true
  }
  return strings.TrimSpace(enabled) == "1"
}

func (t *CheckoutTask) GetAbckGenCookieBlocklist() (*[]string, error) {
  if t.Url.Host == "www.finishline.com" || t.Url.Host == "www.jdsports.com" {
    return &DefaultAbckGenCookieBlocklistFnl, nil
  } else {
    return &DefaultAbckGenCookieBlocklistFtl, nil
  }

  resp, err := t.doReq(t.s3client, t.makeGetReq(AbckGenCookieBlocklistUrl, nil))
  if err != nil {
    DiscardResp(resp)
    return nil, err
  }
  slotsLines, err := readBody(resp)
  if err != nil {
    return nil, err
  }
  slots := strings.Split(slotsLines, "\n")
  blocklist := []string{}
  for _, slot := range slots {
    if strings.TrimSpace(slot) != "" {
      blocklist = append(blocklist, strings.TrimSpace(slot))
    }
  }
  if len(blocklist) > 0 {
    return &blocklist, nil
  } else {
    return nil, errors.New("Empty blocklist, use default")
  }
}

var DefaultAbckGenCookieBlocklistFnl = []string{"_abck",}// "bm_sv", "JSESSIONID", "FNL_RISKIFIED_ID"} // "bm_sz"}
var DefaultAbckGenCookieBlocklistFtl = []string{
  "_abck",
  "bm_sv",
  "bm_sz",
  // "ak_bmsc",
  "JSESSIONID",
  // "bm_mi",
  // "JSESSIONID",
}
var CookieListByHost = map[string][]string {
  "www.finishline.com": []string {
    // "_abck",
    "akavpau_wwwvpfs",
  },
  "www.jdsports.com": []string {
    // "_abck",
    "akavpau_wwwvpfs",
    "akavpau_jdwwwvpfs",
  },
}

func (t *CheckoutTask) GenValidAbckOnSesh(parentCtx context.Context, akUrl string, hasValidAbckFn func(abck string) bool) (*[]*http.Cookie, error) {
  client, err := t.newHttpClient()
  if err != nil {
    return nil, err
  }
  client.Jar.InitFrom(t.client.Jar)

  // client.Transport = t.client.Transport

  if len(t.AbckGenCookieBlocklist) == 0 {
    if blocklist, err := t.GetAbckGenCookieBlocklist(); err == nil {
      t.LogDebug("using blocklist: %s", strings.Join(*blocklist, ","))
      t.AbckGenCookieBlocklist = *blocklist
    }
  }
  t.LogDebug("using blocklist: %s", strings.Join(t.AbckGenCookieBlocklist, ","))

  for _, name := range t.AbckGenCookieBlocklist {
    ckie := &http.Cookie{ Name: name, MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: ""}
    client.Jar.SetCookies(t.BaseDomainUrl, []*http.Cookie{ ckie })
    client.Jar.SetCookies(t.BaseUrl, []*http.Cookie{ ckie })
    ckie.Secure = false
    client.Jar.SetCookies(t.BaseDomainUrl, []*http.Cookie{ ckie })
    client.Jar.SetCookies(t.BaseUrl, []*http.Cookie{ ckie })
  }
  // client.Jar.SetCookies(t.BaseDomainUrl, []*http.Cookie{ &http.Cookie{ Name: "_abck", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "" } })
  // client.Jar.SetCookies(t.BaseDomainUrl, []*http.Cookie{ &http.Cookie{ Name: "bm_sz", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "" } })
  // client.Jar.SetCookies(t.BaseDomainUrl, []*http.Cookie{ &http.Cookie{ Name: "cart-guid", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "" } })
  // client.Jar.SetCookies(t.BaseDomainUrl, []*http.Cookie{ &http.Cookie{ Name: "JSESSIONID", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "" } })
  // client.Jar.SetCookies(t.BaseUrl, []*http.Cookie{ &http.Cookie{ Name: "_abck", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "" } })
  // client.Jar.SetCookies(t.BaseUrl, []*http.Cookie{ &http.Cookie{ Name: "bm_sz", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "" } })
  // client.Jar.SetCookies(t.BaseUrl, []*http.Cookie{ &http.Cookie{ Name: "cart-guid", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "" } })
  // client.Jar.SetCookies(t.BaseUrl, []*http.Cookie{ &http.Cookie{ Name: "JSESSIONID", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "" } })
  // TODO whitelist

  ctx, cancel := context.WithCancel(parentCtx)
  defer CloseH2Conns(client)
  defer cancel()
  if err != nil {
    return nil, err
  }
  ua, err := t.getUserAgent()
  if err != nil {
    return nil, err
  }

  if t.GetCookieValueClient("_abck", client) == "" {
    _, err = t.postSensorData([]byte{}, client, ctx, ua, akUrl, t.Url)
    if err != nil {
      t.LogDebug("GenValidAbckOnSesh postSensorData error %+v", err)
      // log.Printf("%+v", err)
      return nil, err
    }
  }

  if _, err = t.GenValidAbck(client, ctx, ua, akUrl, hasValidAbckFn); err == nil {
    ckies := []*http.Cookie{}
    if ckie := t.getAbckCookie(client); ckie != nil {
      ckies = append(ckies, ckie)
    } else {
      return nil, ErrFailedToGenCookie
    }
    if ckwl, ok := CookieListByHost[t.Url.Host]; ok {
      for _, ckname := range ckwl {
        if ckie := getCookie(ckname, t.BaseUrl, client.Jar); ckie != nil {
          ckie.Secure = true
          ckie.HttpOnly = false
          ckie.MaxAge = 1e6
          ckie.Domain = t.BaseUrl.Host
          ckies = append(ckies, ckie)
        }
      }
    } else {
      if ckie := getCookie("bm_sz", t.BaseUrl, client.Jar); ckie != nil {
        ckie.Secure = true
        ckie.HttpOnly = false
        ckie.MaxAge = 1e6
        ckie.Domain = t.BaseDomainUrl.Host
        ckies = append(ckies, ckie)
      }
    }
    // for _, ckie := range client.Jar.Cookies(t.BaseUrl) {
    //   if ckie.Name == "_abck" {
    //     continue
    //   }
    //   ckie.Secure = true
    //   ckie.HttpOnly = false
    //   ckie.MaxAge = 1e6
    //   if ckie.Domain != t.BaseUrl.Host {
    //     ckie.Domain = t.BaseDomainUrl.Host
    //   }
    //   ckies = append(ckies, ckie)
    // }
    return &ckies, nil
  }
  return nil, err
}

func (t *CheckoutTask) GenValidAbckOffSesh(pctx context.Context, akUrl string, hasValidAbckFn func(abck string) bool) (*[]*http.Cookie, error) {
  client, err := t.newHttpClient()
  if err != nil {
    return nil, err
  }
  ctx, cancel := context.WithCancel(pctx)
  defer CloseH2Conns(client)
  defer cancel()
  if err != nil {
    return nil, err
  }
  ua, err := t.getUserAgent()
  if err != nil {
    return nil, err
  }
  _, err = t.postSensorData([]byte{}, client, ctx, ua, akUrl, t.Url)
  if err != nil {
    // log.Printf("%+v", err)
    return nil, err
  }
  if _, err = t.GenValidAbck(client, ctx, ua, akUrl, hasValidAbckFn); err == nil {
    ckies := []*http.Cookie{}
    if ckie := t.getAbckCookie(client); ckie != nil {
      ckies = append(ckies, ckie)
    }
    // good for fnl, bad for foots ?
    // if ckie := getCookie("bm_sz", t.BaseUrl, client.Jar); ckie != nil {
    //   ckie.Secure = true
    //   ckie.HttpOnly = false
    //   ckie.MaxAge = 1e6
    //   ckie.Domain = t.BaseDomainUrl.Host
    //   ckies = append(ckies, ckie)
    // }
    return &ckies, nil
  }
  return nil, err
}

var InitialAbckBackoff = 3

func (t *CheckoutTask) PopulateAbckJar(nck int, hasValidAbckFn func(s string) bool) context.CancelFunc {
  ctx, cancel := context.WithCancel(t.ClientCtx)
  backoff := InitialAbckBackoff

  var delay = 1000*time.Millisecond
  if os.Getenv("ABCK_DELAY") == "0" {
    delay = 0
  }
  go func() {
    // time.Sleep(20*time.Second)
    defer recover()
    t.RetryTask(func() error {
      if ctx.Err() != nil {
        return nil
      }

      t.AbckJarMut.Lock()
      ll := len(t.AbckJar)
      t.AbckJarMut.Unlock()

      if ll >= nck {
        t.LogDebug("GENNED %d ABCK, WAITING", nck)
        time.Sleep(1000*time.Millisecond)
        return ErrRetrying
      }
      // TODO try off session if on session fails

      if ckies, err := t.GenValidAbckOffSesh(ctx, t.Config["akUrl"], hasValidAbckFn); err == nil {
        t.AbckJarMut.Lock()
        t.AbckJar = append(t.AbckJar, *ckies)
        t.AbckJarMut.Unlock()
        t.LogDebug("ABCK JAR %d", len(t.AbckJar))// strings.Join(t.AbckJar, "\n"))
        backoff = InitialAbckBackoff
      } else {
        t.LogDebug("abck gen error %+v\n", err)
        if t.shouldRetry(err) {
          t.LogDebug("abck gen backoff %+v\n", backoff)
          time.Sleep(time.Duration(backoff)*time.Second)
          backoff = backoff * 2
          if backoff > 300 {
            // go t.SendTelemetry(map[string]interface{} {
            //   "event": "halt_abck_pregen",
            //   "last_err": fmt.Sprintf("%+v", err),
            // })
            return retry.Unrecoverable(ErrFailedToGenCookie)
          }
        }
        return err
      }
      return ErrRetrying
    }, retry.Delay(delay), retry.DelayType(retry.FixedDelay))
  }()
  return cancel
}

func (t *CheckoutTask) AbckGenLoop(akUrl string, hasValidAbckFn func(abck string) bool) {
  genchan := make(chan bool)
  WORKERS := 2
  if os.Getenv("WORKERS") != "" {
    WORKERS, _ = strconv.Atoi(os.Getenv("WORKERS"))
  }
  start := time.Now()
  ckies := 0

  for i := 1; i <= WORKERS; i++ {
    go func() {
      for _ = range genchan {
        func() {
          ctx, cancel := context.WithCancel(t.Ctx)
          tt := *t
          tt.Ctx = ctx
          err := tt.initClient()
          defer CloseH2Conns(tt.client)
          defer cancel()
          if err != nil {
            log.Printf("%+v", err)
            return
          }
          client := tt.client
          // _, err = tt.FtlGetProduct()
          tt.UserAgent, _ = tt.getUserAgent()
          log.Println(tt.UserAgent)
          _, err = tt.postSensorData([]byte{}, tt.client, tt.ClientCtx, tt.UserAgent, akUrl, t.Url)
          if err != nil {
            // log.Printf("%+v", err)
            return
          }
          abck, err := tt.GenValidAbck(client, tt.ClientCtx, tt.UserAgent, akUrl, hasValidAbckFn)
          // log.Printf("%s", tt.GetCookieValueClient("_abck", client))
          if err != nil {
            log.Printf("failed: %s", tt.UserAgent)
            // log.Printf("%+v", err)
            return
          } else {
            log.Printf("%s (%s)", abck, tt.UserAgent)
          }

          dt := time.Now().Sub(start).Seconds()
          log.Printf("%v valid cookies/second", int(float64(ckies) / dt))
          ckies += 1
        }()
      }
    }()
    // time.Sleep(250*time.Millisecond)
  }


  for i := 0; i < 1e6; i++ {
    genchan <- true
  }
  time.Sleep(600*time.Second)
}

func (t *CheckoutTask) GenValidAbck(client *http.Client, ctx context.Context, ua string, akUrl string, hasValidAbckFn func(abck string) bool) (string, error) {
  var retries = 0
  var abck string
  var body string
  var hasValidAbck bool
  var err error
  // err := t.Retry(func() error {
  //   if hasValidAbckFn(t.GetCookieValueClient("_abck", client)) {
  //     return nil
  //   }
  // })
  for !hasValidAbck {
    abck = t.GetCookieValueClient("_abck", client)
    if ctx.Err() != nil {
      return abck, ctx.Err()
    }
    retries += 1
    if os.Getenv("UA_TEST") == "1" {
      if ua, err := t.getUserAgent(); err == nil {
        t.UserAgent = ua
      }
    }
    if os.Getenv("PRINT_ABCK") == "1" {
      log.Printf("abck: %s\n", abck)
    }

    if retries > 3 && !strings.Contains(abck, "~0~-1") {
      return abck, nil
    }
    if retries > 7 {
      if rand.Float32() > 0.9 {
        go t.SendTelemetry(map[string]interface{} {
          "event": "abck_gen_failed",
          "user_agent": t.UserAgent,
          "abck_gen_failure": map[string]interface{} {
            "abck": abck,
            "last_body": body,
            "last_err": fmt.Sprintf("%+v", err),
          },
        })
      }
      return abck, TaskErrUnrecoverable(ErrFailedToGenCookie)
      // // t.SetCookie(&http.Cookie{ Name: "ak_bmsc", Value: "", MaxAge: -1, Expires: time.Unix(0,0), })
      // // t.SetCookie(&http.Cookie{ Name: "bm_sv", Value: "", MaxAge: -1, Expires: time.Unix(0,0), })
      // // t.SetCookie(&http.Cookie{ Name: "bm_sz", Value: "", MaxAge: -1, Expires: time.Unix(0,0), })
      // fmt.Printf("abck post reset: %s %s %s %s\n", t.GetCookieValue("_abck"), t.GetCookieValue("ak_bmsc"), t.GetCookieValue("bm_sv"), t.GetCookieValue("bm_sz"))
    }
    start := time.Now()
    body, err = t.updateSensorData(client, ctx, ua, akUrl, t.Url)
    if err != nil {
      time.Sleep(time.Duration(1000 + rand.Intn(1000)) * time.Millisecond)
      continue
    }

    dt := time.Now().Sub(start)
    if t.Debug {
      log.Printf("sensor dt %+v", dt)
    }
    // if dt < 1*time.Second {
    //   time.Sleep(1*time.Second - dt)
    // }

    if body == "{\"success\": false}" || body == "{\"success\":false}" {
      t.Config["abck_validation_disabled"] = "1"
      return "", ErrCookieValidationDisabled
    //   // TODO audit this
    }
    //   return abck, TaskErrUnrecoverable(ErrFailedToGenCookie)
    //
    abck = t.GetCookieValueClient("_abck", client)
    if hasValidAbckFn(abck) {
      err = nil
      hasValidAbck = true
    } else {
      if os.Getenv("ABCK_DELAY") != "" {

      } else {
        time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
      }
    }
  }
  abck = t.GetCookieValueClient("_abck", client)
  if hasValidAbckFn(abck) {
    return abck, nil
  } else {
    return abck, ErrFailedToGenCookie
  }
}

func (t *CheckoutTask) PhoneHomeUnusedAbck() {
  t.AbckJarMut.Lock()
  defer t.AbckJarMut.Unlock()
  if len(t.AbckJar) < 2 {
    return
  }

  ckies := make([]interface{}, 0)
  for _, ckie := range t.AbckJar[1:] {
    ckies = append(ckies, ckie)
  }
  if len(ckies)>0 {
    go t.SendTelemetry(map[string]interface{} {
      "event": "unused_abck",
      "unused_abck": map[string]interface{} {
        "cookies": ckies,
      },
    })
  }
}

func (t *CheckoutTask) WaitForAbck(timeout time.Duration) error {
  return t.Retry(func() error {
    t.AbckJarMut.Lock()
    ll := len(t.AbckJar)
    t.AbckJarMut.Unlock()
    if ll > 0 {
      return nil
    }
    return ErrRetrying
  }, retry.Delay(100*time.Millisecond), retry.DelayType(retry.FixedDelay), retry.Attempts(uint((timeout / time.Millisecond) / 100)))
}
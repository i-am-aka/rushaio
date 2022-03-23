package task

import (
	"regexp"
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"net/url"
	"fmt"
	"strings"
	"rush/net/http"
	"github.com/pkg/errors"
	"os"
	"context"
	"nhooyr.io/websocket"
	"github.com/avast/retry-go"
	"time"
)

var DD_UAS = []string{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36",
}

var ErrDdHardBan = errors.New("Proxy Hard Banned, Cannot Solve Captcha")
var StatusDdHardBan = Status{Message: "Proxy Hard Banned, Cannot Solve Captcha"}

func (t *CheckoutTask) DatadomeSendPayload(urlStr string, ref string, cid string) (error) {
	if t.UseIpv6 {
		return nil
	}
	// var err error
	// var serr error
	return t.DatadomeSendPayloadWs(urlStr, ref, cid)
	//; err == nil {
//		return nil
	//}
	// else {
	// 	if serr = t.DatadomeSendPayloadSubprocess(urlStr, ref, cid); serr == nil {
	// 		return nil
	// 	}
	// }
	//return err
}

var ttRe = regexp.MustCompile(`%22ttst%22%3A.+?%2C`)
var ggRe = regexp.MustCompile(`%22tagpu%22%3A.+?%2C`)

func (t *CheckoutTask) DatadomeSendPayloadWs(urlStr string, ref string, cid string) (error) {
	var err error
	defer func() {
		t.LogDebug("DatadomeSendPayloadWs %+v", err)
	}()
	client, err := t.newHttpClient()
	if err != nil {
		return err
	}
	defer CloseH2Conns(client)
	ctx, cancel := context.WithCancel(t.Ctx)
	defer cancel()

	u, _ := url.Parse("ws://127.0.0.1:25144/")
	c, _, err := websocket.Dial(ctx, u.String(), &websocket.DialOptions{
		HTTPHeader: map[string][]string {
			"x-url": []string{urlStr},
			"x-ref": []string{ref},
			"x-cid": []string{cid},
			"x-ua": []string{t.UserAgent},
		},
	})
	if err != nil {
		return err
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	var fetchReq FetchRequest
	_, msg, err := c.Read(ctx)
	if err != nil {
		return err
	}
	// t.LogDebug(string(msg))
	err = json.Unmarshal(msg, &fetchReq)
	if err != nil {
		return err
	}
	// t.LogDebug("%+v", fetchReq)

	headers := [][2]string{}
	for key, value := range fetchReq.Headers {
		headers = append(headers, [2]string{strings.ToLower(key), value})
	}

	bbytes := []byte(fetchReq.Body)
	tagpu := "%22tagpu%22%3A"+fmt.Sprintf("%.15f", 12.0 + rand.Float64() * 8)+"%2C"
	ttst := "%22ttst%22%3A"+fmt.Sprintf("%.15f", 12.0 + rand.Float64() * 20)+"%2C"
	t.LogDebug("MATCH %+v", ttRe.FindStringSubmatch(string(bbytes)))
	t.LogDebug("%s", string(bbytes))
	t.LogDebug("tagpu %s", tagpu)
	t.LogDebug("ttst %s", ttst)
	bbytes = ttRe.ReplaceAll(bbytes, []byte(ttst))
	bbytes = ggRe.ReplaceAll(bbytes, []byte(tagpu))
	// if form, err := url.ParseQuery(string(bbytes)); err == nil {
	// 	ddpl := map[string]interface{}{}
	// 	jsData := []byte(form.Get("jsData"))

	// 	t.LogDebug(string(jsData))
	// 	if err := json.Unmarshal(jsData, &ddpl); err == nil {
	// 		if _, ok := ddpl["tagpu"]; ok {
	// 			ddpl["tagpu"] = json.Number(fmt.Sprintf("%.15f", 12.0 + rand.Float64() * 8))
	// 		}
	// 		if _, ok := ddpl["ttst"]; ok {
	// 			ddpl["ttst"] = json.Number(fmt.Sprintf("%.15f", 12.0 + rand.Float64() * 20))
	// 		}
	// 		jsDataFixed, _ := json.Marshal(ddpl)
	// 		jdq := append([]byte("jsData="), []byte(url.QueryEscape(string(jsDataFixed)))...)
	// 		jdq = append(jdq, []byte("&")...)
	// 		bbytes = jdRe.ReplaceAll(bbytes, jdq)

	// 		form.Set("jsData", string(jsDataFixed))
	// 		bbytes = []byte(form.Encode())
	// 	}
	// }
	furl, _ := url.Parse(fetchReq.Url)
	hreq := t.makeReq(strings.ToUpper(fetchReq.Method), furl, &headers, nil, &bbytes).WithContext(ctx)
	hreq.Close = true
	resp, err := t.doReq(client, hreq)
	if err != nil {
		DiscardResp(resp)
		// log.Printf("DoReq failed %+v", err)
		return err
	}
	return nil
}


type GeetsSoln struct {
	GeetestChallenge string `json:"geetest_challenge"`
	GeetestSeccode   string `json:"geetest_seccode"`
	GeetestValidate  string `json:"geetest_validate"`
	Score            string `json:"score"`
	SolveTime        string `json:"solveTime"`
}

var GeetsApiUrl, _	= url.Parse("http://127.0.0.1:1996")

func (t *CheckoutTask) SolveGeetestHttp(gurl string) (*GeetsSoln, error) {
	var soln GeetsSoln
	err := t.Retry(func() error {
		c := &http.Client{}
		resp, err := t.doReq(c, t.makeGetReq(GeetsApiUrl, &[][2]string{ {"url", gurl}}))
		if err != nil {
			return err
		}

		return readRespJsonDst(resp, &soln)
	})
	return &soln, err
}


func (t *CheckoutTask) SolveGeetestWs(gurl string) (*GeetsSoln, error) {
	var soln GeetsSoln
	var err error
	defer func() {
		t.LogDebug("DatadomeSendPayloadWs %+v", err)
	}()

	ctx, cancel := context.WithTimeout(t.Ctx, 30*time.Second)
	defer cancel()

	u, _ := url.Parse("wss://api.rushaio.co/gtgen")
		// u, _ := url.Parse("ws://127.0.0.1:1997/")
	c, _, err := websocket.Dial(ctx, u.String(), &websocket.DialOptions{
		HTTPHeader: map[string][]string {
			"url": []string{gurl},
			"x-ua": []string{t.UserAgent},
		},
	})
	if err != nil {
		return nil, err
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	for ctx.Err() == nil {
		_, msg, err := c.Read(ctx)
		if err != nil {
			return nil, err
		}
		t.LogDebug("WS RECV %s", string(msg))
		// fmt.Printf("WS RECV %s\n", string(msg))
		var fetchReq FetchRequest
		err = json.Unmarshal(msg, &fetchReq)
		err = json.Unmarshal(msg, &soln)
		if err == nil && len(fetchReq.Url) > 0 {
			go func() {
				defer func() {
					recover()
				}()
				headers := [][2]string{}
				for key, value := range fetchReq.Headers {
					headers = append(headers, [2]string{strings.ToLower(key), value})
				}
				bbytes := []byte(fetchReq.Body)
				furl, _ := url.Parse(fetchReq.Url)
				client, err := t.newHttpClient()
				if err != nil {
					return
				}
				defer CloseH2Conns(client)
				defer client.CloseIdleConnections()
				t.Retry(func() error {
					hreq := t.makeReq(strings.ToUpper(fetchReq.Method), furl, &headers, nil, &bbytes).WithContext(ctx)
					// hreq.Close = true
					// TODO remove when one client can connect to multiple hosts

					resp, err := t.doReq(client, hreq)
					if err != nil {
						DiscardResp(resp)
						t.LogDebug("DoReq failed %+v", err)
						return ErrPxGenFailed
					}
					body, err := readBodyBytes(resp)
					if err != nil {
						t.LogDebug("read failed %+v", err)
						return ErrPxGenFailed
					}
					jresp := map[string]interface{} {
						"response": map[string]interface{} {
							"url": fetchReq.Url,
							"headers": resp.Header,
							"type": "default",
							"redirected": false,
							"status": resp.StatusCode,
							"statusText": resp.Status,
							"_stream": nil,
						},
						"responseBuffer": base64.StdEncoding.EncodeToString(body),
					}
					if jj, err := json.Marshal(jresp); err == nil {
						c.Write(ctx, websocket.MessageText, jj)
					}
					return nil
				}, retry.Attempts(3), retry.Delay(0*time.Second))
			}()
			// bs := string(body)
			// if strings.Contains(bs, "_px3|") {
			// 	return []*http.Cookie {
			// 			&http.Cookie{
			// 				Name: "_px3",
			// 				Value: strings.Split(bs, "|")[3],
			// 				Path: "/",
			// 		},
			// 	}, nil
			// }
			// log.Printf("WS SEND %s", string(body))
		} else if len(soln.GeetestChallenge) > 0 {
			return &soln, err
		}
	}
	return nil, ErrFailedToSolveCaptcha

}

func (t *CheckoutTask) DdCapVerify(capUrl string, ddCkie string) (error) {
	if strings.Contains(capUrl, "t=bv") {
		return ErrDdHardBan
	}
	if !strings.Contains(capUrl, "&cid=") {
		capUrl += "&cid=" + ddCkie
	}

	gcdClient, err := t.newHttpClient()
	if err != nil {
		return err
	}
	defer CloseH2Conns(gcdClient)
	defer gcdClient.CloseIdleConnections()
	t.LogDebug(capUrl)
	capUrlParsed, err := url.Parse(capUrl)
	// headerOrder := []string {
	//   "Host",
	//   "Connection",
	//   "User-Agent",
	//   "Content-Type",
	//   "Accept",
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
	//   {"Content-Type", "application/x-www-form-urlencoded; charset=UTF-8"},
	//   {"Accept", "*/*"},
	//   {"Sec-Fetch-Site", "same-origin"},
	//   {"Sec-Fetch-Mode", "cors"},
	//   {"Sec-Fetch-Dest", "empty"},
	//   // {"Referer", capUrl},
	//   {"Accept-Encoding", "gzip, deflate, br"},
	//   {"Accept-Language", "en-US,en;q=0.9"},
	// }
	// req := t.makeReq("GET", capUrlParsed, &headers, &headerOrder, nil)
	// req.Close = true
	// if t.UseIpv6 {
	// 	gcdClient, _ = t.newHttpClientProxy(nil)
	// }

	// resp, err := t.doReq(gcdClient, req)
	// if err != nil {
	// 	gcdClient, _ = t.newHttpClientProxy(nil)
	// 	var gerr error
 //  	resp, gerr = t.doReq(gcdClient, req)
 //  	if gerr != nil {
 // 			return err
 // 		}
	// }
	// capUrlBody, err := readBody(resp)
	// if err != nil {
	// 	return err
	// }

	// if strings.Contains(capUrlBody, "hard-block") {
	// 	return ErrDdHardBan
	// }
	// t.LogDebug(capUrlBody)

	// TODO race mode
		// spawn N requests
		// when first finishes, cancel others and continue
	// cap, err := t.SolveGeetestHttp(capUrl)
	// cap, err := t.SolveGeetestWs(capUrl)
	// if err != nil {
	// 	return err
	// }
	recapToken, err := t.SolveCaptcha("6LccSjEUAAAAANCPhaM2c-WiRxCZ5CzsjR_vd8uX", capUrl)
	if err != nil {
		return err
	}
	url_, err := url.Parse("https://geo.captcha-delivery.com/captcha/check")
  if err != nil {
    return err
  }
  chal, err := JsEvalFile(map[string]string{}, "datadome/chal.js", t.UserAgent, ddCkie, "10")
  if err != nil {
    return err
  }
  chal = strings.TrimSpace(chal)

  qparams := [][2]string {
    {"cid", ddCkie},
    {"icid", capUrlParsed.Query().Get("initialCid")},
    {"ccid", "null"},
    {"g-recaptcha-response", recapToken},
    // {"geetest-response-challenge", cap.GeetestChallenge},
    // {"geetest-response-validate", cap.GeetestValidate},
    // {"geetest-response-seccode", cap.GeetestSeccode},
    {"hash", capUrlParsed.Query().Get("hash")},
    {"ua", t.UserAgent},
    {"referer", ""},
    {"parent_url", capUrl},
    {"x-forwarded-for", ""},
    {"captchaChallenge", chal},
    {"s", capUrlParsed.Query().Get("s")},
  }

  for _, qp := range qparams {
  	q := url.Values{}
    q.Add(qp[0], qp[1])
  	url_.RawQuery += q.Encode() + "&"
  }
  url_.RawQuery = url_.RawQuery[:len(url_.RawQuery) - 1]
  origCaseFn := t.HeaderCaseFn
  t.HeaderCaseFn = func(s string) string { return s }
  defer func() {
  	t.HeaderCaseFn = origCaseFn
  }()
  headerOrder := []string {
    "Host",
    "Connection",
    "User-Agent",
    "Content-Type",
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
    {"Content-Type", "application/x-www-form-urlencoded; charset=UTF-8"},
    {"Accept", "*/*"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", capUrl},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }
  req := t.makeReq("GET", url_, &headers, &headerOrder, nil)
  req.Close = true
  resp, err := t.doReq(gcdClient, req)
  if err != nil {
  	if resp != nil && resp.StatusCode == 403 {
  		t.Metrics.Incr("dd_captcha_check_403")
  	}
  	return err
  }
  // TODO close conn
  rc := map[string]string{}
  readRespJsonDst(resp, &rc)
  if _, ok := rc["cookie"]; !ok {
  	return ErrFailedToSolveCaptcha
  }

  ckie := rc["cookie"]
  t.LogDebug(ckie)
	csp := strings.SplitN(ckie, ";", 2)[0]
	csps := strings.SplitN(csp, "=", 2)
	t.SetCookie(&http.Cookie{
        Name: csps[0],
        Value: csps[1],
        Secure: true,
        HttpOnly: false,
        MaxAge: 1e6,
        Path: "/",
      // Domain: t.BaseDomainUrl.Host,
     },)
  return nil

}

func (t *CheckoutTask) DatadomeSendPayloadSubprocess(url string, ref string, cid string) error {
	// var doneChan <-bool
	args := []string{}
	if os.Getenv("C6") == "1" {
		args = append(args, "--cert", "/Users/aka/drive/mitm/charles-ssl-proxying-certificate.pem") // todo remove
	}
	args = append(args, "--allow-all", "datadome/dd.exec.js")
	args = append(args, url)
	args = append(args, ref)
	args = append(args, cid)
	// log.Printf("%+v", args)

	env := map[string]string{
		"NO_PROXY": "js.datadome.co",
	}
	if t.ProxyStr != "" {
		env["HTTPS_PROXY"] = t.ProxyStr
	}

	_, err := JsEvalFile(env, args...)
	// log.Printf("%+v %+v", ret, err)
	return err
}
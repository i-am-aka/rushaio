package task

import (
	"fmt"
	"github.com/avast/retry-go"
	// "log"
	"context"
	"strings"
	"runtime"
	"path/filepath"
	"net/url"
	"encoding/json"
	"rush/net/http"
	"os"
	"time"
	"github.com/pkg/errors"
	"os/exec"
	"nhooyr.io/websocket"
)

func GetPxDenoDir() (string, error) {
	if os.Getenv("DENO_DIR") != "" {
		return os.Getenv("DENO_DIR"), nil
	}

	dir, err := os.Getwd()
	if err != nil {
	  return "", err
	}
	return filepath.Join(dir, "zeronaught", "deno"), nil
}

type Px3Request struct {
	Cookie string
	AppId string
	Vid string
	Uuid string
	Url string
	Host string
	Proxy string
	UserAgent string
	JsSrc string
	CapJsSrc string
	RecapToken string
}

var ErrPxGenFailed = errors.New("PX Gen Failed")

func pxApiHost() string {
	if os.Getenv("PX_API_HOST") != "" {
		return "ws://" +os.Getenv("PX_API_HOST")
	} else {
		return "wss://api.rushaio.co"
	}
}

func pxHttpApiHost() string {
	if os.Getenv("PX_API_HOST") != "" {
		return "http://" +os.Getenv("PX_API_HOST")
	} else {
		return "https://api.rushaio.co"
	}
}


type PxBlockResponse struct {
	RedirectURL       string `json:"redirectUrl"`
	AppID             string `json:"appId"`
	JsClientSrc       string `json:"jsClientSrc"`
	FirstPartyEnabled bool   `json:"firstPartyEnabled"`
	Vid               string `json:"vid"`
	UUID              string `json:"uuid"`
	HostURL           string `json:"hostUrl"`
	BlockScript       string `json:"blockScript"`
}

type FetchRequest struct {
	Headers map[string]string `json:"headers"`
	ReferrerPolicy string `json:"referrerPolicy"`
	Body           string `json:"body"`
	Method         string `json:"method"`
	Mode           string `json:"mode"`
	Url 					 string `json:"url"`
}

func (t *CheckoutTask) GetPxCookieApi(request Px3Request) ([]*http.Cookie, error) {
	t.LogDebug("%+v", request)
	cookies := []*http.Cookie{}
	// return cookies, nil
	ctx, cancel := context.WithTimeout(t.Ctx, 20*time.Second)
	defer cancel()

	opts := websocket.DialOptions {
		HTTPHeader: map[string][]string {
			"Cookie": []string {
				fmt.Sprintf("Authorization=%s", t.GetJwt()),
			},
		},
	}

	u, _ := url.Parse(pxApiHost() + "/gen10")
	c, _, err := websocket.Dial(ctx, u.String(), &opts)
	if err != nil {
		return cookies, err
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	if request.UserAgent == "" {
		request.UserAgent = t.UserAgent
	}
	if reqJson, err := json.Marshal(request); err == nil {
		c.Write(ctx, websocket.MessageBinary, reqJson)
	} else {
		return cookies, err
	}


	// TODO: retry requests

	// TODO: consider intercepting ocaptcha resp

	for ctx.Err() == nil {
		_, msg, err := c.Read(ctx)
		if err != nil {
			return cookies, err
		}
		t.LogDebug("WS RECV %s", string(msg))
		// fmt.Printf("WS RECV %s\n", string(msg))
		var fetchReq FetchRequest
		err = json.Unmarshal(msg, &fetchReq)
		if err == nil {
			headers := [][2]string{}
			for key, value := range fetchReq.Headers {
				headers = append(headers, [2]string{strings.ToLower(key), value})
			}
			bbytes := []byte(fetchReq.Body)
			furl, _ := url.Parse(fetchReq.Url)
			client, err := t.newHttpClient()
			if err != nil {
				return cookies, err
			}
			defer CloseH2Conns(client)
			t.Retry(func() error {
				hreq := t.makeReq(strings.ToUpper(fetchReq.Method), furl, &headers, nil, &bbytes).WithContext(ctx)
				// hreq.Close = true
				// TODO remove when one client can connect to multiple hosts
				resp, err := t.doReq(client, hreq)
				if err != nil {
					DiscardResp(resp)
					t.LogDebug("\n\nDoReq failed %+v\n\n", err)
					return ErrPxGenFailed
				}
				body, err := readBodyBytes(resp)
				if err != nil {
					// log.Printf("DoReq failed %+v", err)
					return ErrPxGenFailed
				}
				jresp := map[string]string {
					"url": fetchReq.Url,
					"text": string(body),
				}
				if jj, err := json.Marshal(jresp); err == nil {
					c.Write(ctx, websocket.MessageText, jj)
				}
				return nil
			}, retry.Attempts(3), retry.Delay(0*time.Second))
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
		} else {
			msgstr := strings.TrimSpace(string(msg))

			if strings.Contains(msgstr, "; ") {
				cookiestrs := strings.Split(msgstr, "; ")
				for _, cs := range cookiestrs {
					sp := strings.SplitN(cs, "=", 2)
					// log.Printf("%+v", sp)
					if len(sp) != 2 {
						continue
					}
					cookies = append(cookies, &http.Cookie{
						Name: strings.TrimSpace(sp[0]),
						Value: strings.TrimSpace(sp[1]),
						Path: "/",
						MaxAge: 999999, Expires: time.Now().Add(time.Duration(86400*7)*time.Second),
					})
				}
				// log.Printf("GOT COOKIES %+v", cookies)
				if len(cookies) == 0 {
					return cookies, ErrPxGenFailed
				}
				return cookies, nil
			}
			return cookies, err
		}
	}

	return cookies, nil
}


func (t *CheckoutTask) GetPxMobileUserAgent() (string, error) {
	if os.Getenv("MOBILE_USER_AGENT") != "" {
		return os.Getenv("MOBILE_USER_AGENT"), nil
	}

	client := &http.Client{Timeout:30*time.Second}
	var userAgent string
	err := t.Retry(
		func() error {
			req, _ := http.NewRequestWithContext(t.Ctx, "GET", pxHttpApiHost() + "/pxmobileua", nil)
			req.Close = true
			req.Header = map[string][]string {
				"authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
			}
			resp, err := t.doReq(client, req)
			if err != nil {
				return err
			} else {
				userAgent, err = readBody(resp)
				return err
			}
		},
	)
	t.LogDebug("ua %s", userAgent)
	return userAgent, err
}

func (t *CheckoutTask) GetPxUserAgent() (string, error) {
	if os.Getenv("USER_AGENT") != "" {
		return os.Getenv("USER_AGENT"), nil
	}

	client := &http.Client{Timeout:30*time.Second}
	var userAgent string
	err := t.Retry(
		func() error {
			req, _ := http.NewRequestWithContext(t.Ctx, "GET", pxHttpApiHost() + "/pxua", nil)
			req.Close = true
			req.Header = map[string][]string {
				"authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
			}
			resp, err := t.doReq(client, req)
			if err != nil {
				return err
			} else {
				userAgent, err = readBody(resp)
				return err
			}
		},
	)
	t.LogDebug("ua %s", userAgent)
	return userAgent, err
}

func GetPxCookie(request Px3Request) ([]*http.Cookie, error) {
	cookies := []*http.Cookie{}
	var deno string
	if runtime.GOOS == "windows" {
	  deno = "external/deno.exe"
	} else {
	  // deno = "external/deno_osx_x64"
	  deno = "deno"
	}
	denoAbspath, err := exec.LookPath(deno)
	if err != nil {
	  return cookies, err
	}
	denoDir, err := GetPxDenoDir()
	if err != nil {
	  return cookies, err
	}
	reqBytes, _ := json.Marshal(request)
	args := []string {
		"run", "--cached-only", "--allow-read", "--allow-write", "--allow-net", "--allow-env",
	}
	cert := os.Getenv("CERT")
	if cert != "" {
		args = append(args, "--cert")
		args = append(args, cert)
	}
	args = append(args, "px/env.js")
	args = append(args, string(reqBytes))
	cmd := exec.Command(denoAbspath, args...)
	cmd.Env = append(os.Environ(), "DENO_DIR=" + denoDir)
	cmd.Env = append(cmd.Env, "HTTPS_PROXY=" + request.Proxy)
	out, err := cmd.Output()
	// log.Printf("%+v %+v", string(out), err)
	if err != nil {
		return cookies, err
	}
  outLines := strings.Split(string(out), "\n")
  cookiestrs := strings.Split(outLines[0], "; ")
  for _, cs := range cookiestrs {
  	sp := strings.SplitN(cs, "=", 2)
  	// log.Printf("%+v", sp)
  	if len(sp) != 2 {
  		continue
  	}
  	cookies = append(cookies, &http.Cookie{
  		Name: sp[0],
  		Value: sp[1],
  		Path: "/",
  		Secure: true,
  		HttpOnly: false,
  		MaxAge: 999999, Expires: time.Now().Add(time.Duration(86400*7)*time.Second),
  	})
  }
  if len(cookies) == 0 {
  	return cookies, ErrPxGenFailed
  }
  return cookies, nil
}

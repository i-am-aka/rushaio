package task

import (
	"math/rand"
	// "log"
	"net"
	"strings"
	"io/ioutil"
	"encoding/json"
	"rush/net/http"
	"fmt"
	"bytes"
	"strconv"
	"os"
	"time"
	"github.com/pkg/errors"
	"github.com/avast/retry-go"
	"context"
	"net/url"
)
var CREATE_CAP_TASK_URL, _ = url.Parse(captchaApiHost() + "/createTask")
var CAP_TASK_STATUS_URL, _ = url.Parse(captchaApiHost() + "/getTaskResult")

func (t *CheckoutTask) SolveCaptcha(siteKey string, capUrl string) (string, error) {
	solvChan := make(chan CapSolution)
	solvCtx, solvCancel := context.WithCancel(t.ClientCtx)
	defer func() {
		go func() {
			time.Sleep(time.Duration(rand.Float32())*time.Second)
			// close(solvChan)
		}()
	}()
	defer solvCancel()


	i := 0
	SOLVES := 1 + rand.Intn(3)
	if os.Getenv("SOLVES") != "" {
		if s, err := strconv.Atoi(os.Getenv("SOLVES")); err == nil {
			SOLVES = s
		}
	}
	for i < SOLVES {
		go func() error {
			// defer solvCancel()

				var err error
				defer func() {
					if err != nil {
						t.LogDebug("cap solve fail %+v", err)
					}
				}()
				defer recover()


				var request CreateCapTaskRequest
				err = json.Unmarshal(BASE_CAP_REQ_JSON, &request)
				if err != nil {
					return err
				}
				request.ClientKey = t.Id
				request.SoftID = rand.Intn(1e6)
				request.LanguagePool = fmt.Sprintf("%s", request.SoftID)
				request.Task.UserAgent = t.UserAgent
				request.Task.WebsiteURL = capUrl
				// request.Task.WebsiteURL = capUrlParsed.Scheme + "://" + capUrlParsed.Host + "/robots.txt"
				request.Task.WebsiteKey = siteKey
				request.Task.Type = "NoCaptchaTaskProxyless"

				// TODO use task proxy
				// use less as more solves complete with prox
				// rotate to avoid bans
				if t.Proxy != nil {
					if pi, err := strconv.Atoi(t.Proxy.Port()); err == nil {
						if os.Getenv("CAPPROXY") == "task" || (t.Proxy.User != nil && rand.Float32() < 0.15) {
							request.Task.ProxyType = "http"
							request.Task.ProxyAddress, _, _ = net.SplitHostPort(t.Proxy.Host)
							request.Task.ProxyPort = pi
							request.Task.ProxyLogin = t.Proxy.User.Username()
							request.Task.ProxyPassword, _  = t.Proxy.User.Password()
						}
					}
				}

				// TODO judge NID quality, solve time,
				// TODO try diff nid strategies
				// TODO save NIDs for reuse, especially the good ones
				if nid := getCookieValue("NID", ElgoogUrl, t.client.Jar); nid != "" {
					request.Task.Cookies = fmt.Sprintf("NID=%s", nid)
					t.Metrics.Incr("nid_uses")
				}
				// 	// nid := NIDS[rand.Intn(len(NIDS))]
				// 	// t.client.Jar.SetCookies(ElgoogUrl, []*http.Cookie{&http.Cookie{Name: "NID", Value: nid}})
				// 	request.Task.Cookies = fmt.Sprintf("NID=%s", nid)
				// // }
				if capps := os.Getenv("CAPPROXY"); capps != "task" && capps != "" {
					if capp, _ := url.Parse(os.Getenv("CAPPROXY")); capp != nil {
						request.Task.ProxyType = "http"
						request.Task.ProxyAddress, _, _ = net.SplitHostPort(capp.Host)
						request.Task.ProxyPort, _ = strconv.Atoi(capp.Port())
						request.Task.ProxyLogin = capp.User.Username()
						request.Task.ProxyPassword, _ = capp.User.Password()
					}
				}

				client := &http.Client{Timeout: 30*time.Second}

				var csreq CapTaskStatusRequest
				csreq.TaskID = 420
				csreq.SiteKey = siteKey
				csb, _ := json.Marshal(csreq)
				var csol CapSolution
				// 	req, _ := http.NewRequestWithContext(t.ClientCtx, "POST", CAP_TASK_STATUS_URL.String(), ioutil.NopCloser(bytes.NewReader(csb)))
				// 	req.Close = true
				// 	req.Header = map[string][]string {
				// 		"Accept": []string{"*/*"},
				// 		"welove": "gerolsteiner",
				// 		"Content-Type": []string{"application/x-www-form-urlencoded"},
				// 		"authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
				// 	}
				// 	req.RawHeader = [][2]string{
				// 	{"Content-Length", fmt.Sprintf("%d", len(csb))},
				// }
				// 	req.ContentLength = int64(len(csb))
				// 	if resp, err := client.Do(req); err == nil {
				// 		rb, _ := readBodyBytes(resp)
				// 		t.LogDebug(string(rb))
				// 		if err = json.Unmarshal(rb, &csol); err == nil {
				// 			if csol.Status == "ready" {
				// 				t.LogDebug("%+v", csol)
				// 				solvChan <- csol
				// 				return nil
				// 			}

				// 		}
				// 	}

				rb, _ := json.Marshal(request)
			  // buffer := &bytes.Buffer{}
			  // encoder := json.NewEncoder(buffer)
			  // encoder.SetEscapeHTML(false)
			  // if err := encoder.Encode(request); err != nil {
			  // 	return err
			  // }
			  // rb := buffer.Bytes()

				t.LogDebug(string(rb))
				ctx, cancel := context.WithCancel(solvCtx)
				defer cancel()

			// var t0 = time.Now()

				getcapDelay := 2*time.Second + time.Duration(rand.Intn(8))*time.Second
				if os.Getenv("NOCAPDELAY") != "" {
					getcapDelay = 1*time.Second
				}

				sentCapRequest := false
				err = t.Retry(func() error {
					defer func() {
						if r := recover(); r != nil {
							t.LogDebug("CAP PANIC %+v", r)
						}
					}()

					if solvCtx.Err() != nil {
							return retry.Unrecoverable(solvCtx.Err())
					}
					// defer func() {
						if solvCtx.Err() == nil && !sentCapRequest {
							req, err := http.NewRequestWithContext(ctx, "POST", CREATE_CAP_TASK_URL.String(), ioutil.NopCloser(bytes.NewReader(rb)))
							if err != nil {
								return err
							}
							req.Header = map[string][]string {
								"Accept": []string{"*/*"},
								"Content-Type": []string{"application/x-www-form-urlencoded"},
								"authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
								"welove": []string{"hits"},
								"walton": []string{"1"},
								"x-host": {t.Url.Host},
								"task_id": {t.Id},
								"uid": {t.UserId},
								"w3": {"1"},
							}
							req.RawHeader = [][2]string{
								{"Content-Length", fmt.Sprintf("%d", len(rb))},
							}
							req.ContentLength = int64(len(rb))
							req.Close = true
							t.LogDebug("%+v", req)
							resp, err := client.Do(req)
							if err == nil {
								var cresp CreateCapTaskResponse
								if err := readRespJsonDst(resp, &cresp); err == nil {
									t.LogDebug("%+v", cresp)
									sentCapRequest = true
								}
							}
							t.LogDebug("%+v %+v", resp, err)
						}
					// }()

					req, _ := http.NewRequestWithContext(ctx, "POST", CAP_TASK_STATUS_URL.String(), ioutil.NopCloser(bytes.NewReader(csb)))
					req.Close = true
					req.Header = map[string][]string {
						"Accept": []string{"*/*"},
						"Content-Type": []string{"application/x-www-form-urlencoded"},
						"authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
						"welove": {"hits"},
						"walton": []string{"1"},
						"x-host": {t.Url.Host},
						"task_id": {t.Id},
						"uid": {t.UserId},
						"w3": {"1"},
					}
					req.RawHeader = [][2]string{
					{"Content-Length", fmt.Sprintf("%d", len(csb))},
				}
					req.ContentLength = int64(len(csb))
					resp, err := client.Do(req)
					if err != nil {
						return err
					}
					rb, _ := readBodyBytes(resp)
					t.LogDebug(string(rb))
					err = json.Unmarshal(rb, &csol)
					if err != nil {
						return err
					}
					t.LogDebug("%+v", csol)
					if csol.ErrorID != 0 {
						t.LogDebug("%+v", csol)
						return retry.Unrecoverable(errors.New("Error solving captcha: " + csol.ErrorDescription))
					}
					if csol.Status == "ready" && !strings.Contains(csol.Solution.Text, "ERROR") && len(csol.Solution.GRecaptchaResponse) > 0 {
						sentCapRequest = true
						t.LogDebug("%+v", csol)
						go func() {
							defer recover()
							if solvCtx.Err() == nil {
								solvChan <- csol
							}
						}()
						return nil
					}
					return ErrRetrying
				}, retry.Attempts(60), retry.Delay(getcapDelay), retry.OnRetry(func(_ uint, err error) {
					t.LogDebug("cap retry %+v", err)
				}))
				t.LogDebug("cap res %+v", err)
				return err
		}()
		i += 1
	}

	var csol CapSolution
	select {
	case csol = <-solvChan:
		break
	case <-solvCtx.Done():
		return "", ErrFailedToSolveCaptcha
	case <-time.After(90*time.Second):
		close(solvChan)
		solvCancel()
		return "", ErrFailedToSolveCaptcha
	}
	t.LogDebug("%+v", csol)
	// if csol.PreSolved.Cookie != "" {
	// 	t.LogDebug("presolved %s", csol.PreSolved.Cookie)
	// 	t.SetCookie(&http.Cookie{
	//         Name: "datadome",
	//         Value: csol.PreSolved.Cookie,
	//         Secure: true,
	//         HttpOnly: false,
	//         MaxAge: 1e6,
	//         Path: "/",
	//       // Domain: t.BaseDomainUrl.Host,
	//      },)
	//   return nil
	// }
	if csol.Solution.GRecaptchaResponse == "" {
		return "", ErrFailedToSolveCaptcha
	}

	return csol.Solution.GRecaptchaResponse, nil
}

type CreateCapTaskRequest struct {
	ClientKey string `json:"clientKey"`
	Task      struct {
		Type          string `json:"type"`
		WebsiteURL    string `json:"websiteURL"`
		WebsiteKey    string `json:"websiteKey"`
		ProxyType     string `json:"proxyType"`
		ProxyAddress  string `json:"proxyAddress"`
		ProxyPort     int    `json:"proxyPort"`
		ProxyLogin    string `json:"proxyLogin"`
		ProxyPassword string `json:"proxyPassword"`
		UserAgent     string `json:"userAgent"`
		Cookies       string `json:"cookies"`
	} `json:"task"`
	SoftID       int    `json:"softId"`
	LanguagePool string `json:"languagePool"`
}

type CreateCapTaskResponse struct {
	ErrorCode int `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
	ErrorID int `json:"errorId"`
	TaskID  int `json:"taskId"`
}

type CapTaskStatusRequest struct {
	ClientKey string `json:"clientKey"`
	SiteKey string `json:"siteKey"`
	TaskID int `json:"taskId"`
}

type CapSolution struct {
	PreSolved struct {
		Cookie string
	}
	ErrorID  int    `json:"errorId"`
	ErrorCode int `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
	Status   string `json:"status"`
	Solution struct {
		Text               string `json:"text"`
		GRecaptchaResponse string `json:"gRecaptchaResponse"`
	} `json:"solution"`
	Cost       string `json:"cost"`
	IP         string `json:"ip"`
	CreateTime int    `json:"createTime"`
	EndTime    int    `json:"endTime"`
	SolveCount string `json:"solveCount"`
}

var BASE_CAP_REQ_JSON = []byte(`{"clientKey": "", "task": {"type": "NoCaptchaTaskProxyless", "websiteURL": "https://recaptcha-demo.appspot.com/recaptcha-v2-checkbox.php", "websiteKey": "6LfW6wATAAAAAHLqO2pb8bDBahxlMxNdo9g947u9"}, "softId": 847, "languagePool": "en"}`)



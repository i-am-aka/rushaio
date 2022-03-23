package main

import (
	"encoding/base64"
	"context"
	"time"
	"log"
	// "path/filepath"

	"io/ioutil"
	"strings"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"

	"github.com/chromedp/cdproto/storage"
	"fmt"
	"net/http"
	"net/url"
	"rush/rcdp"
)

const (
	PageBody = `
<!DOCTYPE html>
<html lang="en">
<head></head>
</html>
`
)


func Intercept(ctx *context.Context, url string) func(event interface{}) {
	return func(event interface{}) {
		switch ev := event.(type) {
		case *fetch.EventRequestPaused:
			go func() {
				c := chromedp.FromContext(*ctx)
				cctx := cdp.WithExecutor(*ctx, c.Target)
				if ev.Request.URL == url {
					bodyB64 := base64.StdEncoding.EncodeToString([]byte(PageBody))
					err := fetch.FulfillRequest(ev.RequestID, 200).WithBody(bodyB64).Do(cctx)
					if err != nil {
						log.Printf("Intercept error: %+v", err)
					}
				// } else if (strings.Contains(ev.Request.URL, ".js") || strings.Contains(ev.Request.URL, "https://www.google.com/recaptcha/api2/")) {
				} else if(strings.Contains(ev.Request.URL, "/recaptcha") || strings.Contains(ev.Request.URL, ".js") || strings.Contains(ev.Request.URL, "woff")) {
					fetch.ContinueRequest(ev.RequestID).Do(cctx)
				} else {
					fetch.ContinueRequest(ev.RequestID).Do(cctx)
					// log.Printf("blocked %+v", ev.Request.URL)
					// fetch.FailRequest(ev.RequestID, network.ErrorReasonBlockedByClient).Do(cctx)
				}
			}()
		}
	}
}

func testProxy(proxyUrl string) {
	request, err := http.NewRequest("GET", "https://ipecho.net/plain", nil)
	if err != nil {
		log.Fatalf("new request failed:%v", err)
	}
	tr := &http.Transport{Proxy: func(req *http.Request) (*url.URL, error) { return url.Parse(proxyUrl) }}
	client := &http.Client{Transport: tr}
	rsp, err := client.Do(request)
	if err != nil {
		log.Fatalf("get rsp failed:%v", err)

	}
	defer rsp.Body.Close()
	data, _ := ioutil.ReadAll(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		log.Fatalf("status %d, data %s", rsp.StatusCode, data)
	}
}

type RCTokenResponse struct {
	Token string
	Err error
}

type RCContext struct {
	cdpCtx *rcdp.CDPContext
}

func (rc RCContext) Close() {
	log.Println("rc.Close")
	rc.cdpCtx.Cancel()
	log.Println("rc.Closed")
}

func NewRCContext() (*RCContext, error) {
	cdpCtx, err := rcdp.NewCDPContext(false)
	if err != nil {
		return nil, err
	}
	return &RCContext{cdpCtx: cdpCtx,}, nil
}

func (rc *RCContext) GetToken(url string, action string, apiKey string, cookies []network.SetCookieParams) (string, error) {
	ctx, cancel := context.WithTimeout(rc.cdpCtx.BrowserCtx, time.Duration(10)*time.Second)
	defer cancel()
	chromedp.ListenTarget(ctx, Intercept(&ctx, url))
	var token string
	foo := ""
	if err := chromedp.Run(ctx,
		fetch.Enable(),
		rcdp.SpoofHeadfulAction(),
		storage.ClearCookies(), chromedp.ActionFunc(func(ctx context.Context) error {
			// log.Println("Doing cookies")
			for _, cookie := range cookies {
				_, err := cookie.Do(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}),
		// chromedp.ActionFunc(func(ctx context.Context) error {
		// 			cookies, err := network.GetAllCookies().Do(ctx)
		// 			if err != nil {
		// 				return err
		// 			}

		// 			for i, cookie := range cookies {
		// 				log.Printf("chrome cookie %d: %+v", i, cookie)
		// 			}

		// 			return nil
		// 		}),
		chromedp.Navigate(url),
		chromedp.Evaluate(fmt.Sprintf(`
			var s = document.createElement('script');
			s.src = "https://www.google.com/recaptcha/api.js?render=%s";
			document.getElementsByTagName("head")[0].appendChild(s);
			s.addEventListener('load', function() {
				window.grecaptcha.ready(function(e) {
					grecaptcha.execute('%s', {action: '%s'}).then(function(t) {
						var g = document.createElement('span');
						g.id = "token";
						g.innerHTML = t;
						document.getElementsByTagName("body")[0].appendChild(g);
					})
				})
			})
			''
		`, apiKey, apiKey, action), &foo),
		chromedp.WaitVisible("#token"),
		chromedp.Evaluate(`document.getElementById('token').innerHTML`, &token)); err != nil {
			log.Printf("%+v", err)
			return "", err
	} else {
		return token, nil
	}
}

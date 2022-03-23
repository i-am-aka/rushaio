package main

import (
	"encoding/json"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/network"
	"rush/rcdp"
	"time"
	// "os"
	// "net/http"
	// "strings"
)

func main() {
	proxy := flag.String("proxy", "", "proxy")
	flag.Parse()
	log.SetOutput(os.Stderr)
	cdpCtx, err := rcdp.NewCDPContext(true)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	cdpCtx.SetUpstreamProxy(*proxy)

	var email string
	var cookies []*network.Cookie
	if err := chromedp.Run(cdpCtx.BrowserCtx,
		rcdp.SpoofHeadfulAction(),
		chromedp.Navigate("https://accounts.google.com/ServiceLogin?hl=en&passive=true&continue=https://www.google.com/"),
		chromedp.WaitReady(`//a[starts-with(@href, 'https://accounts.google.com/SignOutOptions')]`),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error = nil
			cookies, err = network.GetAllCookies().Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
		chromedp.Evaluate(`
			var userBtn = Array.from(document.getElementsByTagName('a')).filter(a => a.href.indexOf('https://accounts.google.com/SignOutOptions')===0)[0];
			var email = userBtn.title.split('(')[1];
			console.log(email, email.slice(0, email.length - 1))
			email.slice(0, email.length - 1)
		`, &email),
		); err != nil {
			fmt.Printf("browser error %+v\n", err)
		} else {
			m := map[string]interface{} {
				"cookies": cookies,
				"email": email,
			}
			b, err := json.Marshal(m)
			if err != nil {
				fmt.Printf("%+v\n", err)
			} else {
				fmt.Printf(string(b[:]))
			}
		}
	chromedp.Cancel(cdpCtx.BrowserCtx)
	cdpCtx.Cancel()
	time.Sleep(60)
}

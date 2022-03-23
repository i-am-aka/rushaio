package task

import (
	"encoding/json"
	"rush/net/http"
	"net/url"
	"strings"
)

func (t *CheckoutTask) SizeVisitProd() (*http.Response, error) {
  url_, err := url.Parse("https://m.size.co.uk/product/black-asics-gel-lyte-iii-womens/16030006/")
  if err != nil {
    return nil, err
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
    "if-none-match",
    "if-modified-since",
  }
  headers := [][2]string {
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"referer", "https://m.size.co.uk/campaign/New+In/?facet:new=latest&sort=latest"},
    {"user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en-GB;q=0.9,en;q=0.8,it;q=0.7"},
    {"if-none-match", "98a626a9792f6416e683f63306d56560"},
    {"if-modified-since", "Tue, 08 Sep 2020 17:32:05 GMT"},
  }
  return t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
}
func (t *CheckoutTask) SizeAtcReq() (*http.Response, error) {
  url_, err := url.Parse("https://m.size.co.uk/cart/16030006.2245058/")
  if err != nil {
    return nil, err
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
    {"content-type", "application/json"},
    {"origin", "https://m.size.co.uk"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36"},
    {"referer", "https://m.size.co.uk/product/black-asics-gel-lyte-iii-womens/16030006/"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en-GB;q=0.9,en;q=0.8,it;q=0.7"},
  }
  bodyStructure := map[string]interface{} {
  "customisations": false,
  "cartPosition": nil,
  "recaptchaResponse": false,
  "cartProductNotification": nil,
  "quantityToAdd": 1,
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  return t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
}


func (t *CheckoutTask) SizeAtc() error {
	t.UserAgent, _ = t.getUserAgent()
	resp, err := t.SizeVisitProd()
	if err != nil {
		return err
	}
	DiscardResp(resp)

	t.GenValidAbck(t.client, t.Ctx, t.UserAgent, "https://m.size.co.uk/assets/9283d004ui26261004e3a7c1fcc801", func (abck string) bool {
		return strings.HasSuffix(abck, "==~-1~-1~-1")
	})

	resp, err = t.SizeAtcReq()
	if err != nil {
		return err
	}
	DiscardResp(resp)
	return nil
}
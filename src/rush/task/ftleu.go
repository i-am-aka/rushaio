package task

import (
	"strings"
	"log"
	"rush/net/http"
	"net/url"
)

func (t *CheckoutTask) FtlEuTest() error {
  t.DefaultReqClose = true
	var err error
	t.UserAgent, err = t.getUserAgent()

	resp, err := t.FtlEuGetProduct()
	if err != nil {
		return err
	}
	body, _ := readBody(resp)
	log.Println(body)
	log.Println()
	log.Println("monitor")

	// t.FtlEuGetBmsv()

	t.EnsureValidAbck("https://www.footlocker.nl/static/2f49bbe0a02ti197c1a0b02706d2e905a", func(abck string) bool {
		log.Println(abck)
		return strings.HasSuffix(abck, "==~-1~-1~-1") && !strings.Contains(abck, "||")
	})
	// t.FtlEuGetBmsv()

	resp, err = t.FtlEuMonitorProduct()
	if err != nil {
		return err
	}
	body, _ = readBody(resp)
	log.Println(body)
	return nil
}

func (t *CheckoutTask) FtlEuGetBmsv() (*http.Response, error) {
  url_, err := url.Parse("https://www.footlocker.nl/INTERSHOP/web/FLE/Footlocker-Footlocker_NL-Site/en_GB/-/EUR/ViewCountryRedirect-GetUserCountryCode")
  if err != nil {
    return nil, err
  }
  qparams := [][2]string {

  }
  q := url_.Query()
  for _, qp := range qparams {
    q.Add(qp[0], qp[1])
  }
  url_.RawQuery = q.Encode()
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
    {"accept", "application/json, text/javascript, */*; q=0.01"},
    {"x-requested-with", "XMLHttpRequest"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", "https://www.footlocker.nl/en/p/puma-future-rider-women-shoes-99258?v=315247072002"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
  DiscardResp(resp)
  return resp, err
}

func (t *CheckoutTask) FtlEuGetProduct() (*http.Response, error) {
  url_, err := url.Parse("https://www.footlocker.nl/nl/startpagina")
  if err != nil {
    return nil, err
  }
  // qparams := [][2]string {
  //   {"v", "315247072002"},
  // }
  // q := url_.Query()
  // for _, qp := range qparams {
  //   q.Add(qp[0], qp[1])
  // }
  // url_.RawQuery = q.Encode()
  headerOrder := []string {
    "cache-control",
    "upgrade-insecure-requests",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-user",
    "sec-fetch-dest",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"cache-control", "max-age=0"},
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "none"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  return t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
}

func (t *CheckoutTask) FtlEuMonitorProduct() (*http.Response, error) {
	// t.RemoveCookie("bm_sz")
  url_, err := url.Parse("https://www.footlocker.nl/INTERSHOP/web/FLE/Footlocker-Footlocker_NL-Site/en_GB/-/EUR/ViewProduct-ProductVariationSelect")
  if err != nil {
    return nil, err
  }
  qparams := [][2]string {
    {"BaseSKU", "315247072002"},
    {"InventoryServerity", "ProductDetail"},
  }
  q := url_.Query()
  for _, qp := range qparams {
    q.Add(qp[0], qp[1])
  }
  url_.RawQuery = q.Encode()
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
    {"accept", "application/json, text/javascript, */*; q=0.01"},
    {"x-requested-with", "XMLHttpRequest"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", "https://www.footlocker.nl/en/p/puma-future-rider-women-shoes-99258?v=315247072002"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  return t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
}
package task

import (
	"bytes"
	// crand "crypto/rand"
	"time"
	"fmt"
	"regexp"
	// "log"
	"encoding/json"
	"math/rand"
	"rush/net/http"
	"net/url"
	"strconv"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"github.com/avast/retry-go"
	"os"
	"github.com/pkg/errors"
)

var (
	HIB_BASE = "https://www.hibbett.com"
	errPxBanned = errors.New("PX Banned (Try Another Proxy)")
)

func (t *CheckoutTask) assertNoPxBan(response *http.Response) error {
	body, err := readBody(response)
	if err != nil {
		return err
	}
	if strings.Contains(body, "id=\"px-captcha\"") {
		return errPxBanned
	}
	return nil
}

var HibSizeIdBySize = map[string]string {
	"XS": "1010",
	"S": "1020",
	"M": "1030",
	"L": "1050",
	"XL": "1060",
	"XXL": "1070",
	"2XL": "1070",
	"XXXL": "1080",
	"3XL": "1080",
	"4XL": "1085",
	"XXXXL": "1085",
	"XST": "1110",
	"ST": "1120",
	"MT": "1130",
	"LT": "1150",
	"XLT": "1160",
	"2XLT": "1170",
	"3XLT": "1180",
	"4XLT": "1185",
	"10-13": "1013",
	"6-9": "0609",
	"1X": "8101",
	"2X": "8102",
	"3X": "8103",
}
func HibSizeId(size string) (string, error) {
	sizeNum, err := strconv.Atoi(strings.Replace(size, "-", "", -1))
	if err != nil {
		sizeNumFl, err := strconv.ParseFloat(strings.Replace(size, "-", "", -1), 32)
		if err != nil {
			if sizeId, ok := HibSizeIdBySize[strings.ToUpper(size)]; ok {
				return sizeId, nil
			} else {
				return "", err
			}
		}
		sizeNum = int(sizeNumFl * 10)
	} else {
		sizeNum = sizeNum * 10
	}
	return fmt.Sprintf("%04d", sizeNum), nil
}

var ErrMissingPxJs = errors.New("Missing PX JS")
func (t *CheckoutTask) PxCaptchaSolve(splashBytes []byte) error {
	splashStr := string(splashBytes)

	jsRe := regexp.MustCompile(`window._pxJsClientSrc = '(.+?)'`)
	jsMatches := jsRe.FindStringSubmatch(splashStr)
	if len(jsMatches) < 2 {
		return errors.New("Missing PX JS")
	}
	jsEndpoint := "https:" + jsMatches[1]

	uuidRe := regexp.MustCompile(`window._pxUuid = '(.+?)'`)
	uuidMatches := uuidRe.FindStringSubmatch(splashStr)
	if len(uuidMatches) < 2 {
		return errors.New("Missing PX UUID")
	}
	uuid := uuidMatches[1]

	vidRe := regexp.MustCompile(`window._pxVid = '(.+?)'`)
	vidMatches := vidRe.FindStringSubmatch(splashStr)
	vid := ""
	if len(vidMatches) == 2 {
		vid = vidMatches[1]
	}

	appIdRe := regexp.MustCompile(`window._pxAppId = '(.+?)'`)
	appIdMatches := appIdRe.FindStringSubmatch(splashStr)
	if len(appIdMatches) < 2 {
		return errors.New("Missing PX AppId")
	}
	appId := appIdMatches[1]

	pxHostRe := regexp.MustCompile(`window._pxHostUrl = '(.+?)'`)
	pxHostMatches := pxHostRe.FindStringSubmatch(splashStr)
	if len(pxHostMatches) < 2 {
		return errors.New("Missing PX Host")
	}
	pxHost := pxHostMatches[1]

	// //captcha.px-cdn.net/PX9Qx3Rve4/captcha.js?a=c&u=cceafd60-d12c-11ea-95e4-035c52fd644f&v=&m=0
	jsUrlRe := regexp.MustCompile(`s.src = '(//.+?)';`)
	jsUrlMatches := jsUrlRe.FindStringSubmatch(splashStr)
	if len(jsUrlMatches) < 2 {
		return ErrMissingPxJs
	}

	t.LogDebug("%+v", Px3Request{
			Proxy: "",
			AppId: appId,
			Vid: vid,
			Uuid: uuid,
			Host: pxHost,
			JsSrc: jsEndpoint,
			CapJsSrc: "",
		})
	return ErrNotImplemented
}

type HibColor struct {
	Id string `json:"id"`
	Label string `json:"label"`
}

type HibSku struct {
	RootPid string
	Pid string `json:"id"`
	Color HibColor `json:"color"`
	Size string `json:"size"`
	IsAvailable bool `json:"isAvailable"`
	SizeId string
	AtcCsrf string
}

type HibProductApiResponse struct {
	Skus []HibSku `json:"skus"`
	LaunchDate string `json:"launchDate"`
	ReleaseTime int64
}

// func (t *CheckoutTask) HibGetRushAvailSkus(rootPid string, colorId string) ([]HibSku, error) {
// 	hibSkus := []HibSku{}
// 	err := t.Retry(func() error {
// 		client := &http.Client{}
// 		url_, err := url.Parse(fmt.Sprintf("%s/hib_skus?rootPid=%s&colorId=%s", apiHost(), rootPid, colorId))
// 		if err != nil {
// 			return err
// 		}
// 		req := t.makeGetReq(url_, nil)
// 		req.Header = map[string][]string {
// 			"Authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
// 		}
// 		resp, err := t.doReq(client, req)
// 		if err != nil {
// 			return err
// 		}
// 		err = readRespJsonDst(resp, &hibSkus)
// 		if err != nil {
// 			return retry.Unrecoverable(err)
// 		}
// 		for idx, sku := range hibSkus {
// 			sku.SizeId, err = HibSizeId(sku.Size)
// 			hibSkus[idx] = sku
// 		}

// 		return nil
// 	})
// 	return hibSkus, err
// }


func (t *CheckoutTask) HibChooseSku(hibSkus []HibSku, sizes []string, colorId string, inStock bool) *HibSku {
	t.LogDebug("%+v %+v %+v %+v", hibSkus, sizes, colorId, inStock)
	var sku *HibSku
	if len(sizes) == 0 && len(hibSkus) > 0 {
		sku = &hibSkus[rand.Intn(len(hibSkus))]
		sku.SizeId, _ = HibSizeId(sku.Size)
	} else {
		rand.Shuffle(len(hibSkus), func(i, j int) {
		  hibSkus[i], hibSkus[j] = hibSkus[j], hibSkus[i] })

		for _, sku_ := range hibSkus {
			if sku != nil { break }

			if inStock && !sku.IsAvailable {
				continue
			}

			if (colorId != "" && sku_.Color.Id != colorId) {
				continue
			}

			t.LogDebug("skupre %+v", sku_)
			sku_.SizeId, _ = HibSizeId(sku_.Size)
			t.LogDebug("skupost %+v", sku_)
			for _, size := range sizes {
				t.LogDebug("%+v", size)
				sizeId, _ := HibSizeId(size)
				t.LogDebug("%+v %+v %+v", sku_.SizeId, sizeId, sku_.SizeId == sizeId)
				if sku_.SizeId == sizeId {
					// todo why do i have to copy sku to prevent an off-by-one error
					sku__ := sku_
					sku = &sku__
					break
				}
			}
		}
	}
	return sku
}

func HibGetRootPid(url_ *url.URL) string {
	pUrlSplit := strings.Split(url_.Path, "/")
	return strings.Split(pUrlSplit[len(pUrlSplit)-1], ".")[0]
}

func HibGetColorId(url_ *url.URL) string {
	rootPid := HibGetRootPid(url_)
	hibColorKey := fmt.Sprintf("dwvar_%s_color", rootPid)
	return url_.Query().Get(hibColorKey)
}

func (t *CheckoutTask) HibVariantLookup(productUrl string, rootPid string, sizeId string, colorId string) (*HibSku, error) {
	getVariantHeaders := [][2]string {
		{"accept", "text/html, */*; q=0.01"},
	  {"accept-language", "en-us"},
		{"referer", productUrl},
	  {"x-requested-with", "XMLHttpRequest"},
	  {"connection", "keep-alive"},
	}
	getVariantHeaderOrder := []string {
		"accept",
		"cookie",
		"referer",
		"host",
		"accept-language",
		"user-agent",
		"accept-encoding",
		"connection",
		"x-requested-with",
	}

	var sku HibSku
	variantUrlStr := fmt.Sprintf(
		"https://www.hibbett.com/product?pid=%s&%s=%s&%s=%s&Quantity=1&format=ajax&productlistid=undefined&pickupOption=ship&showStoreSizeAvailability=false",
		rootPid, fmt.Sprintf("dwvar_%s_size", rootPid), sizeId, fmt.Sprintf("dwvar_%s_color", rootPid), colorId)
	if sizeId == "" {
		variantUrlStr = fmt.Sprintf(
				"https://www.hibbett.com/product?pid=%s&%s=%s&Quantity=1&format=ajax&productlistid=undefined&showStoreSizeAvailability=false",
				rootPid, fmt.Sprintf("dwvar_%s_color", rootPid), colorId)
	}
	variantUrl, err := url.Parse(variantUrlStr)
	if err != nil {
		return nil, err
	}
	variantResp, err := t.doReq(t.client, t.makeReq("GET", variantUrl, &getVariantHeaders, &getVariantHeaderOrder, nil))
	if err != nil {
		return nil, err
	}
	variantBytes, err := readBodyBytes(variantResp)
	if err != nil {
		return nil, err
	}
	if strings.Contains(string(variantBytes), "id=\"px-captcha\"") {
		t.StatusCh <- Status{Message: "PX Banned (PID Lookup)"}
		return nil,errors.New("PX Banned (PID Lookup)")
	}
	variantDoc, err := goquery.NewDocumentFromReader(bytes.NewReader(variantBytes))
	if err != nil {
		return nil,err
	}
	pid := variantDoc.Find("span[itemprop='productID']").Text()
	sku.AtcCsrf, _ = variantDoc.Find("input[name='csrf_token']").Attr("value")
	// if sku.AtcCsrf == "" {
	// 	return nil,errors.New("No CSRF on Product Page")
	// }
	sku.SizeId = sizeId
	sku.Color = HibColor{Id: colorId}
	sku.Pid = pid
	sku.IsAvailable = variantDoc.Find("li.selected").HasClass("selectable")
	return &sku, nil
}

func (t *CheckoutTask) HibAtcClassic(sesh *HibAppSession, url_ *url.URL, sizes []string) (*HibSku, error) {
	t.StatusCh <- Status{Message: "Pre-carting"}
	rootPid := HibGetRootPid(url_)
	product, err := sesh.HibPidLookup(rootPid, t.GetCookieValue("_px3")) // "3")
	if err != nil {
		return nil, err
	}

	colorId := HibGetColorId(url_)
	sku := t.HibChooseSku(product.Skus, sizes, colorId, true)

	if sku == nil {
		t.StatusCh <- Status{Message: "Pre-cart failed (Proxy Banned)"}
		return nil, errors.New("Out of Stock")
	}

	// sku.Color = strings.Title(strings.ToLower(doc.Find("ul.color > li.selected > a > span").Eq(1).Text()))


	ru, _ := url.Parse((strings.Replace(url_.String(), rootPid, sku.Pid, -1)))
	ru.RawQuery = ""
	//

	t.LogDebug("%+v", sku)
	ssku, err := t.HibVariantLookup(ru.String(), rootPid, sku.SizeId, colorId)
	if err != nil {
		return nil, err
	}
	sku.RootPid = rootPid
	sku.AtcCsrf = ssku.AtcCsrf
	sku.Color = HibColor{Id: colorId}

	// sku, _, err := t.HibHttpSkuLookup(url_, sizes)
	// if err != nil {
	// 	return nil, err
	// }
	// t.LogDebug("%+v", sku)
	return t.HibAtcSku(url_, sku)
}

func (t *CheckoutTask) HibAtcSku(productUrl *url.URL, sku *HibSku) (*HibSku, error) {
	t.StatusCh <- Status{Message: "Adding to Cart"}
	atcUrl, err := url.Parse(fmt.Sprintf("%s/on/demandware.store/Sites-Hibbett-US-Site/default/Cart-AddProduct?format=ajax", HIB_BASE))
	if err != nil {
		return nil,err
	}
	atcHeaderOrder := []string {
		"content-length",
		"accept",
		"user-agent",
		"x-requested-with",
		"sec-fetch-site",
		"sec-fetch-mode",
		"sec-fetch-dest",
		"referer",
		"accept-encoding",
		"accept-language",
	}
	ru := productUrl
	if rand.Float32() < 0.5 {
		ru,_ = url.Parse(productUrl.String())
		ru.RawQuery = ""
	}
	atcHeaders := [][2]string {
		{"accept", "*/*"},
		{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
		{"origin", HIB_BASE},
	  {"accept-language", "en-us"},
	  {"x-requested-with", "XMLHttpRequest"},
		{"referer", ru.String()},
	}

	atcParams := [][2]string {
		{"select-Quantity", "1"},
		{"Quantity", "1"},
		{"bopis-ropis-option","ship"},
		{"cartAction", "add"},
		{"pid", sku.Pid},
		{"gtmData", url.QueryEscape("{\"dimension5\":\"Online\"}")},
		{"csrf_token", url.QueryEscape(sku.AtcCsrf)},
	}
	var atcBodyKvs [][]byte
	for _, kv := range atcParams {
		atcBodyKvs = append(atcBodyKvs, []byte(fmt.Sprintf("%s=%s", kv[0], kv[1])))
	}
	atcBody := bytes.Join(atcBodyKvs, []byte("&"))
	atcResp, err := t.doReq(t.client, t.makeReq("POST", atcUrl, &atcHeaders, &atcHeaderOrder, &atcBody))
	if err != nil {
		return nil,err
	}
	if atcResp.StatusCode != 200 {
		t.StatusCh <- StatusAtcFailed
		return nil, ErrAtcFailed
	}
	err = t.assertNoPxBan(atcResp)
	if err != nil {
		t.StatusCh <- StatusPxBanned
		return nil, &ErrSessionBanned{Stage: "ATC", Err: errors.New("PX Banned (ATC)")}
	}
	return sku,nil
}

func (t *CheckoutTask) HibPurchaseNow(pid string) (error) {
	t.StatusCh <- Status{Message: "Move Desired Product To Cart"}

	doc, err := t.HibGetCartDoc("")
	if err != nil {
		return err
	}

	cartRow := doc.Find(fmt.Sprintf(".cart-row[data-pid='%s']", pid))
	href, ok := cartRow.Find(".save-for-later").Attr("href")
	if !ok {
		return errors.New("Missing Purchase Now Link")
	}

	purchaseNowUrl, _ := url.Parse(href)

	// TODO retry i think this is where we need to loop until product is available??
	purchaseNowResp, err := t.doReq(t.client, t.makeGetReq(purchaseNowUrl, nil))
	if err != nil {
		return err
	}
	purchaseNowResp.Body.Close()
	return nil
}

func (t *CheckoutTask) HibGetCartDoc(refUrl string) (*goquery.Document, error) {
	cartUrl, err := url.Parse(fmt.Sprintf("%s/cart", HIB_BASE))
	if err != nil {
		return nil, err
	}
	// cartParams := [][2]string {
	// 	{"products", fmt.Sprintf("[{\"pid\":\"%s\",\"qty\":\"1\"}]", pid)},
	// }
	// t.LogDebug("VISIT CART %+v", cartParams)
	// var cartBodyKvs [][]byte
	// for _, kv := range cartParams {
	// 	cartBodyKvs = append(cartBodyKvs, []byte(fmt.Sprintf("%s=%s", kv[0], kv[1])))
	// }
	// cartBody := bytes.Join(cartBodyKvs, []byte("&"))
	headerOrder := []string {
	  "accept",
	  "cookie",
	  "referer",
	  "user-agent",
	  "accept-language",
	  "accept-encoding",
	  "connection",
	}
	accept := []string{
		// "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	}
	// ua := []string{
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// 	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36",
	// }

	ru := t.Url
	// if rand.Float32() < 0.5 {
	if true {
		ru,_ = url.Parse(ru.String())
		ru.RawQuery = ""
	}
	ref := []string{
		"",
		"",
		"https://www.hibbett.com/",
		"https://www.hibbett.com/cart",
		t.Url.String(),
		ru.String(),
		"https://www.hibbett.com/launch-calendar/",
	}
	headers := [][2]string {
		{"connection", "keep-alive"},
	  {"accept", accept[rand.Intn(len(accept))]},
	  {"accept-encoding", "gzip, deflate, br"},
	  {"accept-language", "en-us"},
	  // {"referer", ref[rand.Intn(len(ref))]},
	}

	r := ref[rand.Intn(len(ref))]
	if refUrl != "" {
		r = refUrl
	}
	if r != "" {
		headers = append(headers, [2]string{"referer", r})
	}
	for _, h := range headers {
		t.LogDebug("%+v=%+v", h[0], h[1])
	}
	req := t.makeReq("GET", cartUrl, &headers, &headerOrder, nil)
	cartResp, err := t.doReq(t.client, req)
	cartBytes, err := readBodyBytes(cartResp)
	if err != nil {
		return nil, err
	}
	if cartResp.StatusCode != 200 {
		return nil, &ErrSessionBanned{Stage: "Visit Cart", Err: errPxBanned}
	}
	if strings.Contains(string(cartBytes), "id=\"px-captcha\"") {
		t.StatusCh <- Status{Message: "PX Banned (Visit Cart)"}
		return nil, &ErrSessionBanned{Stage: "Visit Cart", Err: errPxBanned}
	}
	return goquery.NewDocumentFromReader(bytes.NewReader(cartBytes))
}

func (t *CheckoutTask) HibSaveForLater(pid string) (error) {
	cartDoc, err := t.HibGetCartDoc(t.Url.String())
	if err != nil {
		return err
	}
	saveForLaterUrlStr, ok := cartDoc.Find(".save-for-later").Attr("href")
	if !ok {
		return errors.New("Missing cart save for later link")
	}
	re := regexp.MustCompile(`pid=[0-9]+`)
	pidSaveForLaterUrlStr := re.ReplaceAllString(saveForLaterUrlStr, fmt.Sprintf("pid=%s", pid))
	pidSaveForLaterUrl, err := url.Parse(pidSaveForLaterUrlStr)
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
	  {"referer", "https://www.hibbett.com/cart"},
	  {"accept-encoding", "gzip, deflate, br"},
	  {"accept-language", "en-US,en;q=0.9"},
	}
	resp, err := t.doReq(t.client, t.makeReq("GET", pidSaveForLaterUrl, &headers, &headerOrder, nil))
	if err != nil {
		return err
	}
	readBody(resp)
	// resp.Body.Close()
	return nil
}

func (t *CheckoutTask) HibGetHome() (*http.Response, error) {
  url_, err := url.Parse("https://www.hibbett.com/")
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
    "accept-language",
    "accept-encoding",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-us"},
  }
  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
  if err != nil {
  	return nil, err
  }
  cc, err := readBody(resp)
  if err != nil {
  	return nil, err
  }
  if strings.Contains(string(cc), "id=\"px-captcha\"") {
  	return nil, &ErrSessionBanned{Stage: "Visit Homepage", Err: errPxBanned}
  }
  return resp, nil
}

type HibProductPage struct {
	ReleaseTime int64
}

func (t *CheckoutTask) HibGetProduct() (*HibProductPage, error) {
	var product HibProductPage
  headerOrder := []string {
    "upgrade-insecure-requests",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-user",
    "sec-fetch-dest",
    "accept-language",
    "accept-encoding",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-us"},
    {"referer", "https://www.hibbett.com/"},
  }
  // u := t.Url

  resp, err := t.doReq(t.client, t.makeReq("GET", t.Url, &headers, &headerOrder, nil))
  if err != nil {
  	return nil, err
  }
  cc, err := readBody(resp)
  if err != nil {
  	return nil, err
  }
  if strings.Contains(string(cc), "id=\"px-captcha\"") {
  	return nil, &ErrSessionBanned{Stage: "Visit Product", Err: errPxBanned}
  }
  re := regexp.MustCompile(`var productLaunchDate = '(.+?)';`)
  matches := re.FindStringSubmatch(cc)
  if len(matches) > 1 {
  	// Mon Jan 2 15:04:05 -0700 MST 2006
  	 t, err := time.Parse("Mon Jan 2 15:04:05 MST 2006", matches[1])
  	 if err != nil {
  	 	return nil, err
  	 }
  	 product.ReleaseTime = t.Unix()
  }
  return &product, nil
}

func (t *CheckoutTask) SbxGenPxCookie() error {
	var proxy string
	if t.Proxy != nil {
		proxy = t.Proxy.String()
	}
	t.UserAgent, _ = t.GetPxUserAgent()
	// ckies, err := GetPxCookie(Px3Request{
	ckies, err := t.GetPxCookieApi(Px3Request{
		Proxy: proxy,
		Cookie: " ",
		Url: t.Url.String(),
		// TODO cookies. can regenning help avoid 8min limit on checkout poll?
		AppId: "PXuR63h57Z",
		Vid: "",
		Uuid: "",
		Host: "",
		JsSrc: "https://client.px-cloud.net/PXuR63h57Z/main.min.js",
		CapJsSrc: "",
	})
	for _, ckie := range ckies {
		t.LogDebug("%s=%s", ckie.Name, ckie.Value)
	}
	// t.LogDebug("%+v (error=%+v)", ckies, err)
	if err != nil {
		return err
	}
	return nil
}

func (t *CheckoutTask) HibGenPxCookie() error {
	var proxy string
	if t.Proxy != nil {
		proxy = t.Proxy.String()
	}
	// ckies, err := GetPxCookie(Px3Request{
	ckies, err := t.GetPxCookieApi(Px3Request{
		Proxy: proxy,
		Cookie: t.CookiesStr(),
		Url: t.Url.String(),
		// TODO cookies. can regenning help avoid 8min limit on checkout poll?
		AppId: "PXAJDckzHD",
		Vid: "",
		Uuid: "",
		Host: "",
		JsSrc: "https://www.hibbett.com/px/AJDckzHD/init.js",
		CapJsSrc: "",
	})
	t.LogDebug("%+v (error=%+v)", ckies, err)
	if err != nil {
		return err
	}

	for _, ckie := range ckies {
		// t.SetCookie(ckie)
		if ckie.Name != "pxsid" {
			t.SetCookie(ckie)
		}
		// if ckie.Name == "_px3" {
		// 	// ckie.Value = "d0cbc921fda806178d742f83ed5a2c6615649da751770db4dd25bfd6acad7723:y5WxzFxCYDRjAW8aAw1EaDTYiHPGaksK+U2iBF9wx3I2CYN/pGMG+2u+H/ARTn3oM1a0H+DFP7k3n0UJbXvCVg==:1000:yYCpOwtvUSED7O9KgP76TEqHb/09MyUqPKO9IGA77TKh6sjoeKqRv791qjy1BiUQyidxGEWXcqBc9q+i3ndBS5yllYIx+y0mBxtJ91Rtk8adcb432dp7IhVM4fW6s+XzNxfHkIMd72BuC/m1L+JnexdCPJDpoiOgkNLpBgDCpl4="
		// 	t.SetCookie(ckie)
		// }
	}

	return nil
}

func (t *CheckoutTask) HibGenSidecar() {
	cc := t.ClientCtx
	defer recover()
	t.Retry(func() error {
		defer recover()
		if t.Ctx.Err() != nil || t.ClientCtx.Err() != nil {
			return nil
		}
		err := t.HibGenPxCookie()
		if err != nil {
			// t.
			return err
		} else {
			t.Config["last_cookie_gen_at"] = fmt.Sprintf("%d", time.Now().Unix())
			err = ErrRetrying
		}
    select{
    case <-t.Ctx.Done():
    	return nil
    case <-cc.Done():
    	return nil
    case <-time.After(time.Duration(90 + rand.Intn(90)) * time.Second):
			return err
    }
	}, retry.Delay(time.Duration(2 + rand.Intn(59)) * time.Second))
}

// func (t *CheckoutTask) HibMonitorSidecar() (chan <-bool) {
// 	defer recover()
// 	t.Retry(func() error {
// 		err := t.HibGenPxCookie()
// 		if err != nil {
// 			return err
// 		} else {
// 			err = ErrRetrying
// 		}
//     select{
//     case <-t.Ctx.Done():
//     	return nil
//     case <-time.After(time.Duration(60 + rand.Intn(300)) * time.Second):
// 			return err
//     }
// 	}, retry.Delay(time.Duration(1 + rand.Intn(29)) * time.Second))
// }

var HibPrecartUrls = []string {
	"https://www.hibbett.com/nike-air-force-1-07-grey-white-mens-shoe/M1287.html?dwvar_M1287_color=0256",
	"https://www.hibbett.com/under-armour-mens-tech-2.0-v-neck-tee/6293A.html?dwvar_6293A_color=0001",
	"https://www.hibbett.com/nike-sportswear-mens-t-shirt-black-gold/B4816.html?dwvar_B4816_color=0036",
	"https://www.hibbett.com/under-armour-mens-tech-2.0-v-neck-steel-tee/6296A.html?dwvar_6296A_color=0508",
	"https://www.hibbett.com/nike-mens-everyday-plus-cushion-crew-training-socks---6-pairs/F0309.html?dwvar_F0309_color=0001",
	"https://www.hibbett.com/nike-sportswear-club-mens-graphic-shorts/B4277.html?dwvar_B4277_color=0001",
	"https://www.hibbett.com/never-broke-again-mens-caution-fleece-shorts/0A573.html?dwvar_0A573_color=0001",
	"https://www.hibbett.com/never-broke-again-mens-caution-baby-tee/0A569.html?dwvar_0A569_color=0001",
	"https://www.hibbett.com/nike-sportswear-mens-alumni-shorts-black/B5065.html?dwvar_B5065_color=0036",
	"https://www.hibbett.com/nike-mens-chalk-tee/B5330.html?dwvar_B5330_color=0001",
	"https://www.hibbett.com/nike-sportswear-mens-t-shirt-black-gold/B4816.html?dwvar_B4816_color=0036",
	"https://www.hibbett.com/nike-womens-pro-3-short---black/A1078.html?dwvar_A1078_color=0001",
	"https://www.hibbett.com/nike-womens-dry-tempo-running-shorts/3754Z.html?dwvar_3754Z_color=0001",
	"https://www.hibbett.com/nike-sportswear-essential-t-shirt-glacier-ice/A2329.html?dwvar_A2329_color=4000",
	"https://www.hibbett.com/nike-air-womens-shorts-black/A2198.html?dwvar_A2198_color=0001",
	"https://www.hibbett.com/nike-womens-leg-a-see-just-do-it-leggings/A1931.html?dwvar_A1931_color=0201",

	// "https://www.hibbett.com/nike-sportswear-club-mens-graphic-shorts/B4277.html?dwvar_B4277_color=0001&cgid=men-clothing-shopall#srule=top-sellers&start=3&sz=24",
	// "https://www.hibbett.com/nike-sportswear-club-mens-graphic-shorts/B4277.html?dwvar_B4277_color=0001&cgid=men-clothing-shopall",
	// "https://www.hibbett.com/nike-sportswear-club-mens-graphic-shorts/B4277.html?dwvar_B4277_color=0001",
}
func (t *CheckoutTask) HibCheckout(productUrl *url.URL, size string, profile Profile) error {
	t.HeaderCaseFn = strings.Title
	t.DefaultProto = "HTTP/1.1"
	return t.RetryTask(func() error {
		t.initClient()
		var err error
		t.UserAgent, err = t.GetPxUserAgent() // "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36"
		if err != nil {
			return err
		}


		t.StatusCh <- StatusVisitHomepage
		_, err = t.HibGetHome()
		if err != nil {
			return err
		}


		if os.Getenv("PXGENLOOP") == "1" {
			for {
				t.StatusCh <- StatusGeneratingCookie
		  	err = t.HibGenPxCookie()
		  	if err != nil {
		  		return err
		  	}
	  	}
	  } else {
			// t.StatusCh <- StatusGeneratingCookie
	  	// err = t.HibGenPxCookie()
	  	// if err != nil {
	  	// 	return err
	  	// }
	  }



  // 	t.StatusCh <- StatusVisitProduct
		// presp, err := t.HibGetProduct()
		// if err != nil {
		// 	return err
		// }
		// t.LogDebug("%+v", presp)

		// sesh, err := t.HibAppStartSession(t.GetCookieValue("_px3"), "")
		sesh, err := t.HibAppInitSession()
		t.LogDebug("%+v %+v", sesh, err)
		if err != nil {
			return err
		}

  	socksUrl, _ := url.Parse(HibPrecartUrls[rand.Intn(len(HibPrecartUrls))])
  	socksSku, err := t.HibAtcClassic(sesh, socksUrl, []string{})
  	if err != nil {
  		t.LogDebug("precart failed")
  		return err
  	}



		product, err := sesh.HibPidLookup(HibGetRootPid(productUrl), t.GetCookieValue("_px3"))
		if err != nil {
			return err
		}
		t.LogDebug("product %+v", product)
		sku := t.HibChooseSku(product.Skus, t.Sizes, HibGetColorId(productUrl), false)
		if sku == nil {
			return retry.Unrecoverable(errors.New("Size Unavailable"))
		}
		// t.LogDebug("chose sku: %+v", sku)
		t.StatusCh <- Status{Message: "Saving Product For Later"}
		// _, err = sesh.HibSaveForLater(sku.Pid)
		// if err != nil {
		//   return err
		// }

		// sku.AtcCsrf = socksSku.AtcCsrf
		// t.HibAtcSku(sku)

		err = t.HibSaveForLater(sku.Pid)
		if err != nil {
  		return err
  	}

		t.StatusCh <- Status{Message: "Visiting Cart"}
		cartUrl, err := url.Parse(fmt.Sprintf("%s/checkout", HIB_BASE))
		if err != nil {
			return err
		}

		cartResp, err := t.doReq(t.client, t.makeGetReq(cartUrl, nil))
		if err != nil {
			return err
		}
		cc, err := readBody(cartResp)
		if err != nil {
			return err
		}
		if strings.Contains(string(cc), "id=\"px-captcha\"") {
			t.StatusCh <- Status{Message: "PX Banned (Visit Cart)"}
			return &ErrSessionBanned{Stage: "Visit Cart", Err: errPxBanned}
		}
		checkoutIdRe := regexp.MustCompile(`dwcont=(.+?)"`)

		checkoutIdMatches := checkoutIdRe.FindStringSubmatch(cc)
		if len(checkoutIdMatches) == 0 {
			t.StatusCh <-StatusAtcFailed
			return ErrAtcFailed
		}
		checkoutId := checkoutIdMatches[1]
		csrfRe := regexp.MustCompile(`name="csrf_token" value="(.+?)"`)
		startCheckoutCsrf := csrfRe.FindStringSubmatch(cc)[1]

		t.StatusCh <- Status{Message: "Starting Checkout"}
		startCheckoutUrl, err := url.Parse(fmt.Sprintf("%s/checkout?dwcont=%s", HIB_BASE, checkoutId))
		if err != nil {
			return err
		}
		startCheckoutHeaders := [][2]string {
			{"accept", "*/*"},
			{"content-type", "application/x-www-form-urlencoded"},
			{"origin", HIB_BASE},
			{"sec-fetch-dest", "empty"},
		  {"sec-fetch-mode", "cors"},
		  {"sec-fetch-site", "same-origin"},
			{"referer", cartUrl.String()},
		}
		startCheckoutParams := [][2]string {
			{"dwfrm_login_unregistered", "Checkout as Guest"},
			{"csrf_token", startCheckoutCsrf},
		}
		var startCheckoutBodyKvs [][]byte
		for _, kv := range startCheckoutParams {
			startCheckoutBodyKvs = append(startCheckoutBodyKvs, []byte(fmt.Sprintf("%s=%s", url.QueryEscape(kv[0]), url.QueryEscape(kv[1]))))
		}
		startCheckoutBody := bytes.Join(startCheckoutBodyKvs, []byte("&"))
		startCheckoutResp, err := t.doReq(t.client, t.makeReq("POST", startCheckoutUrl, &startCheckoutHeaders, nil, &startCheckoutBody))
		if err != nil {
			return err
		}
		startCheckoutResp.Body.Close()
		if startCheckoutResp.StatusCode != 302 {
			return errors.New("Retrying checkout due to invalid status code")
		}

		t.StatusCh <- Status{Message: "Checkout 1/4: Shipping"}
		// shipZipRadial, err := t.hibNormalizeZipcode(t.Profile.ShippingAddress)
		// if err != nil {
		// 	return err
		// }
		shipUrl, err := url.Parse(fmt.Sprintf("%s/shipping", HIB_BASE))
		if err != nil {
			return err
		}
		shipResp, err := t.doReq(t.client, t.makeGetReq(shipUrl, nil))
		if err != nil {
			return err
		}
		shipBody, err := readBody(shipResp)
		if err != nil {
			return err
		}
		if strings.Contains(string(shipBody), "id=\"px-captcha\"") {
			t.StatusCh <- Status{Message: "PX Banned (Shipping)"}
			return errors.New("PX Banned (Shipping)")
		}
		shipCsrf := csrfRe.FindStringSubmatch(shipBody)[1]

		shipParams := [][2]string {
			{"dwfrm_singleshipping_shippingAddress_addressFields_firstName", profile.ShippingAddress.FirstName,},
			{"dwfrm_singleshipping_shippingAddress_addressFields_lastName", profile.ShippingAddress.LastName,},
			{"dwfrm_singleshipping_shippingAddress_addressFields_address1", profile.ShippingAddress.Address1,},
			{"dwfrm_singleshipping_shippingAddress_addressFields_address2", profile.ShippingAddress.Address2,},
			{"dwfrm_singleshipping_shippingAddress_addressFields_city", profile.ShippingAddress.City,},
			{"dwfrm_singleshipping_shippingAddress_addressFields_states_state", profile.ShippingAddress.State,},
			{"dwfrm_singleshipping_shippingAddress_addressFields_postal", profile.ShippingAddress.Zip,},
			{"dwfrm_singleshipping_shippingAddress_addressFields_country", profile.ShippingAddress.Country,},
			{"dwfrm_singleshipping_shippingAddress_addressFields_phone", profile.ShippingAddress.HibPhone(),},
			{"dwfrm_singleshipping_shippingAddress_addressFields_radialAVSValidated", "false",},
			{"dwfrm_singleshipping_shippingAddress_useAsBillingAddress", "false",},
			{"dwfrm_singleshipping_narvarPhone", "",},
			{"dwfrm_singleshipping_narvarRequestType", "POST",},
			{"dwfrm_singleshipping_shippingAddress_email_emailAddress", profile.Email,},
			{"dwfrm_singleshipping_shippingAddress_shippingMethodID", "ANY_GND",},
			{"dwfrm_singleshipping_shippingAddress_save", "Continue to Billing",},
			{"csrf_token", shipCsrf,},
			{"dropShipExclusion", "false",},
		}
		billingResp, err := t.hibSubmitForm(fmt.Sprintf("%s/billing", HIB_BASE), "https://www.hibbett.com/shipping", shipParams)
		if err != nil {
			DiscardResp(billingResp)
			return err
		}
		if billingResp.StatusCode != 200 {
			DiscardResp(billingResp)
			if billingResp.StatusCode == 302 {
				t.StatusCh <- StatusShippingAddyRejected
				return retry.Unrecoverable(errors.New("Shipping Address Rejected"))
			} else {
				return errors.New("Retrying checkout due to invalid status code")
			}
		}
		billingBody, err := readBody(billingResp)
		if err != nil {
			return err
		}
		if strings.Contains(string(billingBody), "id=\"px-captcha\"") {
			t.StatusCh <- Status{Message: "PX Banned (Billing 1)"}
			return errors.New("PX Banned (Billing 1)")
		}
		reviewCsrf := csrfRe.FindStringSubmatch(billingBody)[1]

		t.StatusCh <- Status{Message: "Checkout 2/4: Encrypting Card"}
		radialNonce, radialToken, encryptedCvv, err := t.hibGetRadialFields(profile.Card)
		if err != nil {
			t.LogDebug("%+v", err)
			return retry.Unrecoverable(errors.New("Card Rejected Pre-Authorization"))
		}

		t.StatusCh <- Status{Message: "Checkout 3/4: Billing"}
		billZipRadial, err := t.hibNormalizeZipcode(t.Profile.ShippingAddress)
		if err != nil {
			return err
		}
		reviewParams := [][2]string {
			{"dwfrm_billing_billingAddress_addressFields_firstName",profile.BillingAddress.FirstName},
			{"dwfrm_billing_billingAddress_addressFields_lastName",profile.BillingAddress.LastName},
			{"dwfrm_billing_billingAddress_addressFields_address1",profile.BillingAddress.Address1},
			{"dwfrm_billing_billingAddress_addressFields_address2",profile.BillingAddress.Address2},
			{"dwfrm_billing_billingAddress_addressFields_city",profile.BillingAddress.City},
			{"dwfrm_billing_billingAddress_addressFields_states_state",profile.BillingAddress.State},
			{"dwfrm_billing_billingAddress_addressFields_postal",billZipRadial},
			{"dwfrm_billing_billingAddress_addressFields_country",profile.BillingAddress.Country},
			{"dwfrm_billing_billingAddress_addressFields_phone",profile.BillingAddress.HibPhone()},
			{"dwfrm_billing_billingAddress_addressFields_radialAVSValidated","false"},
			{"dwfrm_billing_billingAddress_email_emailAddress",profile.Email},
			{"dwfrm_billing_billingAddress_addToEmailList","false"},
			{"dwfrm_billing_couponCode",""},
			{"dwfrm_billing_giftCertCode",""},
			{"g-recaptcha-response",""},
			{"dwfrm_billing_paymentMethods_selectedPaymentMethodID","CREDIT_CARD"},
			{"dwfrm_billing_paymentMethods_creditCard_owner",profile.Card.Name},
			{"dwfrm_billing_paymentMethods_creditCard_type",profile.Card.HibCardType()},
			{"dwfrm_billing_paymentMethods_creditCard_number", insertInto(profile.Card.Number, 4, ' ')},
			{"dwfrm_billing_paymentMethods_creditCard_encryptedNumber",""},
			{"dwfrm_billing_paymentMethods_creditCard_encryptedCVV",encryptedCvv},
			{"dwfrm_billing_paymentMethods_creditCard_expiration_month",strconv.Itoa(profile.Card.ExpMonth)},
			{"dwfrm_billing_paymentMethods_creditCard_expiration_year",strconv.Itoa(profile.Card.ExpYear)},
			{"dwfrm_billing_paymentMethods_creditCard_cvn","***"},
			{"dwfrm_billing_paymentMethods_creditCard_radialNonce",radialNonce},
			{"dwfrm_billing_paymentMethods_creditCard_radialToken",radialToken},
			{"numberOfAdditionalTokenizatioRetries","1"},
			{"additionalTokenizatioRetriesErroLevel","50000"},
			{"resetRetries", "1"},
			{"dwfrm_billing_paymentMethods_klarna_basketHash",""},
			{"dwfrm_billing_paymentMethods_klarna_client__token",""},
			{"dwfrm_billing_paymentMethods_klarna_session__id",""},
			{"dwfrm_billing_paymentMethods_klarna_payment__methods",""},
			{"dwfrm_billing_paymentMethods_klarna_authorization__token",""},
			{"dwfrm_profile_customer_firstname",""},
			{"dwfrm_profile_customer_lastname",""},
			{"dwfrm_profile_customer_email",""},
			{"dwfrm_billing_password",""},
			{"dwfrm_billing_passwordconfirm",""},
			{"dwfrm_billing_save"," Continue to Review"},
			{"csrf_token", reviewCsrf},
		}
		reviewResp, err := t.hibSubmitForm(fmt.Sprintf("%s/review", HIB_BASE), fmt.Sprintf("%s/review", HIB_BASE), reviewParams)
		if err != nil {
			return err
		}
		reviewBody, err := readBody(reviewResp)
		if err != nil {
			return err
		}
		if strings.Contains(string(reviewBody), "id=\"px-captcha\"") {
			t.StatusCh <- Status{Message: "PX Banned (Billing 2)"}
			return &ErrSessionBanned{Stage: "Billing 2", Err: errPxBanned}
		}
		reviewCsrf = csrfRe.FindStringSubmatch(reviewBody)[1]

		err = t.HibSaveForLater(socksSku.Pid)
		if err != nil {
			return err
		}

		err = t.HibPurchaseNow(sku.Pid)
		t.LogDebug("%+v", err)
		if err != nil {
			return err
		}

		t.StatusCh <- Status{Message: "Checkout 4/4: Ordering"}

		// TODO: api for rel time
		// t.StatusCh <- Status{Message: "Waiting for release"}
		// dt := time.Duration(1596722395 - time.Now().Unix()) * time.Second
		// t.LogDebug("%+v", dt)
		// time.Sleep(dt)

		det1, err := t.hibGetFp()

		monitorPing := make(chan bool)
		orderAttempt := make(chan bool)
		done := make(chan bool)
		defer close(done)
		defer close(orderAttempt)
		defer close(monitorPing)
		jitter := 10*time.Second
		// product.ReleaseTime = time.Now().Unix() + 1200
		isLaunchProduct := product.ReleaseTime > time.Now().Unix()
		if isLaunchProduct {
			go t.HibGenSidecar()

			// product.ReleaseTime = time.Now().Unix() + 5
			jitter = 1*time.Second
			t.StatusCh <- StatusWaitingForRelease
			t.LogDebug("waiting for %+v", time.Duration(product.ReleaseTime - time.Now().Unix()) * time.Second)
	    select{
	    case <-t.Ctx.Done():
	    	return t.Ctx.Err()
	    case <-time.After(time.Duration(product.ReleaseTime - time.Now().Unix()) * time.Second):
	    	break
	    }
		} else {
			go t.HibGenSidecar()

			// TODO ctx for when task loops
			var failures = 0
			go func() {
				for {
					select{
					case <-t.Ctx.Done():
						return
					case <-t.ClientCtx.Done():
						return
					default:
						t.StatusCh <- StatusWaitingForRestock
						var err error
						var mdoc *goquery.Document
						// if _, err := t.HibGetHome(); err == nil {
							if mdoc, err = t.HibGetCartDoc(""); err == nil {
								table := mdoc.Find("#cart-table")
								if table == nil {
									failures += 1
									t.LogDebug("%+v", mdoc)
									t.StatusCh <- Status{Message: "Desktop Monitor Error"}
									break
								}
								avail := table.Find(".product-availability-list")
								failures = 0
								if avail.Find("span").HasClass("is-in-stock") {
									go t.SendTelemetry(map[string]interface{} {
										"event": "hib_desktop_monitor_ping",
										"sku": sku,
									})
									t.StatusCh <- Status{Message: "Monitor Pinged"}
									t.LogDebug("PING")
									monitorPing <- true
									select {
									case <-orderAttempt:
										break
									case <-t.Ctx.Done():
										break
									case <-t.ClientCtx.Done():
										return
									}
								}
							} else {
								failures += 1
								t.LogDebug("%+v", err)
								t.StatusCh <- Status{Message: "Desktop Monitor Error"}
								break
							}
						// } else {
						// 	t.LogDebug("%+v", err)
						// 	t.StatusCh <- Status{Message: "Desktop Monitor Error"}
						// }
						// var msku *HibSku
						// if msku, err = t.HibVariantLookup(HibGetRootPid(productUrl), sku.SizeId, sku.Color.Id); err == nil {
						// 	if msku.IsAvailable {
						// 		go t.SendTelemetry(map[string]interface{} {
						// 			"event": "hib_desktop_monitor_ping",
						// 			"sku": msku,
						// 		})
						// 		t.StatusCh <- Status{Message: "Monitor Pinged"}
						// 		t.LogDebug("PING")
						// 		monitorPing <- true
						// 		select {
						// 		case <-orderAttempt:
						// 			break
						// 		case <-t.Ctx.Done():
						// 			break
						// 		}
						// 	}
						// }
					}
					if failures > 5 {
						done <- true
						return
					}
					time.Sleep(time.Duration(5 + rand.Intn(55))*time.Second)
				}
			}()
			// go func() {
			// 	defer recover()

			// 	for {
			// 		select{
			// 		case <-t.Ctx.Done():
			// 			return
			// 		case <-t.ClientCtx.Done():
			// 			return
			// 		default:
			// 			var err error
			// 			if sesh.ExpTimeUnix < time.Now().Unix() {
			// 				sesh, err = t.HibAppStartSession(t.GetCookieValue("_px3"), "")
			// 				t.LogDebug("%+v %+v", sesh, err)
			// 				if err != nil {
			// 					break
			// 				}
			// 			}
			// 			t.StatusCh <- StatusWaitingForRestock
			// 			if product, err := sesh.HibPidLookup(HibGetRootPid(productUrl), t.GetCookieValue("_px3")); err == nil {
			// 				for _, msku := range product.Skus {
			// 					if msku.Pid == sku.Pid && msku.IsAvailable {
			// 						go t.SendTelemetry(map[string]interface{} {
			// 							"event": "hib_monitor_ping",
			// 							"sku": msku,
			// 						})
			// 						t.StatusCh <- Status{Message: "Monitor Pinged"}
			// 						t.LogDebug("PING")
			// 						monitorPing <- true
			// 						select {
			// 						case <-orderAttempt:
			// 							time.Sleep(30*time.Second)
			// 							break
			// 						case <-t.Ctx.Done():
			// 							break
			// 						case <-t.ClientCtx.Done():
			// 							break
			// 						}
			// 					}
			// 				}
			// 			} else {
			// 				t.StatusCh <- Status{Message: "Mobile Monitor Error"}
			// 			}
			// 		}
			// 		time.Sleep(time.Duration(1 + rand.Intn(3))*time.Second)
			// 	}
			// }()
		}

		return t.Retry(func() error {
			// err = t.HibGenPxCookie()
	  // 	if err != nil {
	  // 		return err
	  // 	}
	  	if !isLaunchProduct {
		  	select{
		  	case <-t.Ctx.Done():
		  		return retry.Unrecoverable(t.Ctx.Err())
		  	case <-monitorPing:
			  	defer func() {
			  		orderAttempt <- true
			  	}()
		  		break
		  	case <-done:
		  		t.LogDebug("DONE")
		  		return retry.Unrecoverable(ErrPxBanned)
		  	}
		  }
			startInstantCheckoutTime := timeMillis()
			orderParams := [][2]string {
				{"det1", det1}, // fp
				{"acceptCharset", "UTF-8"},
				// {"submitOrder", "Place Order"},
				{"csrf_token", reviewCsrf},
				{"radialToken", radialToken},
				{"encryptedCardCVV", encryptedCvv},
				{"encryptedCardNumber", "null"},
				{"RDFUID", getCookieValue("rdf-uuid", productUrl, t.client.Jar)},
			}
			// t.SetCookie(&http.Cookie{ Name: "pxsid", Value: "", MaxAge: -1, Expires: time.Unix(0,0), })
			t.StatusCh <- Status{Message: "Ordering"}
			orderResp, err := t.hibSubmitForm(fmt.Sprintf("%s/order-confirmation", HIB_BASE), "https://www.hibbett.com/review", orderParams)
			if err != nil {
				DiscardResp(orderResp)
				return err
			}
			instantCheckoutTime := fmt.Sprintf("%.02fs", (float64(timeMillis()) - float64(startInstantCheckoutTime)) / 1000.0)

			orderBytes, err := readBodyBytes(orderResp)
			if err != nil {
				return err
			}
			if strings.Contains(string(orderBytes), "id=\"px-captcha\"") {
				t.StatusCh <- Status{Message: "PX Banned (Order Confirm)"}

				return TaskErrUnrecoverable(errors.New("PX Banned (Order Confirm)"))
			} else if strings.Contains(string(orderBytes), "<h1>CSRF Token mismatch!</h1>") {
				t.StatusCh <- Status{Message: "Checkout session expired, restarting Task"}
				return retry.Unrecoverable(ErrRetrying)
			}
			orderDoc, err := goquery.NewDocumentFromReader(bytes.NewReader(orderBytes))
			if err != nil {
				return err
			}
			matches := csrfRe.FindStringSubmatch(string(orderBytes))
			if len(matches) > 1 {
				reviewCsrf = csrfRe.FindStringSubmatch(string(orderBytes))[1]
			}
			productSize := strings.TrimSpace(orderDoc.Find("div[data-attribute='size'] span.value").First().Text())
			errFormSel := orderDoc.Find(".error-form")
			for i := range errFormSel.Nodes {
				errStr := errFormSel.Eq(i).Text()
				if strings.Contains(errStr, "declined") {
					t.StatusCh <- Status{Message: "Checkout Failed: Card Declined"}
					productName := strings.TrimSpace(orderDoc.Find(".mini-cart-name").First().Text())
					whFields := [][2]string{
						{"Website", t.Url.Host},
						{"Product", productName},
						{"Size", productSize},
						{"Color", sku.Color.Label},
						{"Checkout Time", instantCheckoutTime},
					}
					go t.DeclineWebhookFields(whFields)
					return nil
				} else if strings.Contains(errStr, "shopping bag can currently not be ordered") {
					t.StatusCh <- Status{Message: "Waiting For Instant Checkout"}
					return ErrRetrying
				} else {
					go t.SendTelemetry(map[string]interface{} {
						"event": "hib_checkout_failed_unknown_error",
						"response": string(orderBytes),
					})
					t.StatusCh <- Status{Message: "Checkout Failed: Unknown Error."}
					return errors.New("unknown checkout error")
				}
			}

			confSel := orderDoc.Find(".confirmation-message").First()
			for _ = range confSel.Nodes {
				productName := strings.TrimSpace(orderDoc.Find(".product-list-item .name").First().Text())
				whFields := [][2]string{
					{"Website", t.Url.Host},
					{"Product", productName},
					{"Size", productSize},
					{"Color", sku.Color.Label},
					{"Checkout Time", instantCheckoutTime},
				}
				t.StatusCh <- Status{Message: "COOKED!!!"}
				go t.SuccessWebhookFields(whFields)
				return nil
			}

			go t.SendTelemetry(map[string]interface{} {
				"event": "hib_checkout_failed_unknown_response",
				"response": string(orderBytes),
			})
			t.StatusCh <- Status{Message: "Checkout Failed"}
			return errors.New("unknown checkout response")
		}, retry.DelayType(retry.RandomDelay), retry.MaxJitter(jitter))
	}, retry.Delay(3 * time.Second), retry.Attempts(10), retry.DelayType(retry.BackOffDelay), retry.MaxDelay(1800 * time.Second), retry.LastErrorOnly(true), retry.OnRetry(func (_ uint, err error) {
    CloseH2Conns(t.client)
  }))
}

func (t *CheckoutTask) hibGetFp() (string, error) {
	var fp string
	err := t.Retry(func() error {
		client := &http.Client{}
		url_, err := url.Parse(fmt.Sprintf("%s/hib_fp", apiHost()))
		if err != nil {
			return err
		}
		req := t.makeGetReq(url_, nil)
		req.Header = map[string][]string {
			"Authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
		}
		resp, err := t.doReq(client, req)
		if err != nil {
			DiscardResp(resp)
			return err
		}
		fp, err = readBody(resp)
		if err != nil {
			return err
		}
		fpSplit := strings.Split(fp, ";")
		pdt, err := time.LoadLocation("America/Los_Angeles")
		if err != nil {
		  return err
		}
		fpSplit[57] = time.Now().In(pdt).Format("1/2/2006") + strings.Replace(url.QueryEscape(time.Now().In(pdt).Format(",3:04:05 PM")), "+", "%20", -1)
		fp = strings.Join(fpSplit, ";")
		return nil
	})
	return fp, err
}

func (t *CheckoutTask) hibGetRadialFields(card Card) (string, string, string, error) {
	client, err := t.newHttpClient()

	radialNonce, err := t.hibGetRadialNonce()
	if err != nil {
		return "", "", "", err
	}
	pkeyurl, _ := url.Parse(
		fmt.Sprintf(
			"https://hostedpayments.radial.com/hosted-payments/publickey/pan?access_token=%s",
			radialNonce))
	radialPubKeyResp, err := t.doReq(client, t.makeGetReq(pkeyurl, nil))
	if err != nil {
		DiscardResp(radialPubKeyResp)
		return "", "", "", err
	}
	radialPubKeyBytes, err := readBodyBytes(radialPubKeyResp)
	if err != nil {
		return "", "", "", err
	}
	var pkeymap map[string]interface{}
	err = json.Unmarshal(radialPubKeyBytes, &pkeymap)
	if err != nil {
		return "", "", "", err
	}
	radialPubKey := pkeymap["publicKey"].(string)

	cardNumberEnc, err := EncRadial(card.Number, radialPubKey)
	if err != nil {
		return "", "", "", err
	}

	protecturl, _ := url.Parse(
		fmt.Sprintf(
			"https://hostedpayments.radial.com/hosted-payments/pan/protect?access_token=%s",
			radialNonce))
	protectheaders := [][2]string {
			{"accept", "application/json"},
			{"origin", HIB_BASE},
			{"content-type", "application/json"},
			{"sec-fetch-dest", "empty"},
		  {"sec-fetch-mode", "cors"},
		  {"sec-fetch-site", "same-origin"},
			{"referer", "https://hostedpayments.radial.com/"},
	}
	protectmap := map[string]string {
		"encryptedPaymentAccountNumber": cardNumberEnc,
	}
	protectbody, err := json.Marshal(protectmap)
	if err != nil {
		return "", "", "", err
	}
	csTokenResp, err := t.doReq(client, t.makeReq("POST", protecturl, &protectheaders, nil, &protectbody))
	if err != nil {
		return "", "", "", err
	}
	csTokenBytes, err := readBodyBytes(csTokenResp)
	if err != nil {
		return "", "", "", err
	}
	var protectresp map[string]string
	err = json.Unmarshal(csTokenBytes, &protectresp)
	if err != nil {
		return "", "", "", err
	}
	radialToken := protectresp["account_token"]

	cvvCrypted, err := EncRadial(card.Cvc, radialPubKey)
	if err != nil {
		return "", "", "", err
	}

	return radialNonce, radialToken, cvvCrypted, nil
}

func (t *CheckoutTask) hibGetRadialNonce() (string, error) {
	url_, err := url.Parse(fmt.Sprintf("%s/on/demandware.store/Sites-Hibbett-US-Site/default/Radial-GetNonce", HIB_BASE))
	if err != nil {
		return "", err
	}
	resp, err := t.doReq(t.client, t.makeGetReq(url_, nil))
	if err != nil {
		return "", err
	}
	respBytes, err := readBodyBytes(resp)
	if err != nil {
		return "", err
	}
	var result map[string]interface{}
	json.Unmarshal(respBytes, &result)
	nonce, ok := result["nonce"].(string)
	if ok {
		return nonce, nil
	} else {
		return "", errors.New(fmt.Sprintf("Nonce error: %s", string(respBytes)))
	}
}

func insertInto(s string, interval int, sep rune) string {
	var buffer bytes.Buffer
	before := interval - 1
	last := len(s) - 1
	for i, char := range s {
		buffer.WriteRune(char)
		if i%interval == before && i != last {
			buffer.WriteRune(sep)
		}
	}
	return buffer.String()
}

func (t *CheckoutTask) hibGetCsrf(urlStr string) (string, error) {
	url_, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	shipResp, err := t.doReq(t.client, t.makeGetReq(url_, nil))
	if err != nil {
		return "", err
	}
	shipBody, err := readBody(shipResp)
	if err != nil {
		return "", err
	}
	csrfRe := regexp.MustCompile(`name="csrf_token" value="(.+?)"`)
	return csrfRe.FindStringSubmatch(shipBody)[1], nil
}

func (t *CheckoutTask) hibNormalizeZipcode(address Address) (string, error) {
	params := [][2]string {
		{"address1", address.Address1},
		{"address2", address.Address2},
		{"countryCode", address.Country},
		{"stateCode", address.State},
		{"postalCode", address.Zip},
		{"city", address.City},
	}
	queryStr := urlEncodeParams(params)
	url_, err := url.Parse(fmt.Sprintf("%s/on/demandware.store/Sites-Hibbett-US-Site/default/RadialAVS-GetResultCode?%s", HIB_BASE, queryStr))
	if err != nil {
		return "", err
	}
	resp, err := t.doReq(t.client, t.makeGetReq(url_, nil))
	if err != nil {
		return "", err
	}
	respBytes, err := readBodyBytes(resp)
	if err != nil {
		return "", err
	}
	var result map[string]string
	json.Unmarshal(respBytes, &result)
	return result["postalCode"], nil
}

func (t *CheckoutTask) hibSubmitForm(formUrl string, referer string, params [][2]string) (*http.Response, error) {
	_url, err := url.Parse(formUrl)
	if err != nil {
		return nil, err
	}
	  headerOrder := []string {
	  	"host",
    "content-type",
    "origin",
    "accept-encoding",
    "cookie",
    "connection",
    "accept",
    "user-agent",
    "referer",
    "accept-language",
    "content-length",
  }
  headers := [][2]string {
    {"origin", HIB_BASE},
    {"connection", "keep-alive"},
    {"accept-Encoding", "gzip, deflate, br"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
    {"content-Type", "application/x-www-form-urlencoded"},
    {"accept-Language", "en-us"},
    {"referer", referer},
  }
	body := urlEncodeParams(params)
	return t.doReq(t.client, t.makeReq("POST", _url, &headers, &headerOrder, &body))
}

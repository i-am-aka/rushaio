package task

import (
	"bytes"
	"strings"
	// "log"
	"regexp"
	"github.com/avast/retry-go"
	"time"
	"fmt"
	"encoding/json"
	"rush/net/http"
	"net/url"
	"math/rand"
	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

func (t *CheckoutTask) EblensCheckout() error {
	var productTitle string
	var pid string
	var attr1 string
	var attr2 string

	err := t.Retry(func() error {
		t.StatusCh <- Status{Message: "Visiting Product"}
		headers := make([][2]string, 0)
		headers = addHeader(headers, "sec-fetch-site", "none")
		headers = addHeader(headers, "sec-fetch-mode", "navigate")
		headers = addHeader(headers, "sec-fetch-user", "?0")
		headers = addHeader(headers, "upgrade-insecure-requests", "1")
		resp, err := t.doReq(t.client, t.makeGetReq(t.Url,&headers))
		if err != nil {
			return err
		}
		prodBody, err := readBodyBytes(resp)
		if err != nil {
			return err
		}
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(prodBody))
		if err != nil {
			return err
		}

    if strings.Contains(doc.Find(".page-heading").Text(), "currently unavailable or out of stock") {
      return retry.Unrecoverable(retry.Unrecoverable(errOutOfStock))
    }
		titleSplit := strings.Split(doc.Find("title").Text(), "EbLens | ")
		productTitle = titleSplit[len(titleSplit)-1]
		attr1 = doc.Find("#attr1").First().AttrOr("value", "NULL")
		attr2 = doc.Find("#attr2").First().AttrOr("value", "NULL")
		pathParts := strings.Split(t.Url.Path, "-")
		pid = pathParts[len(pathParts)-1]
		return nil
	}, retry.Delay(256 * time.Millisecond), retry.Attempts(1e4), retry.DelayType(retry.FixedDelay), retry.LastErrorOnly(true))
	if err != nil {
		return err
	}
	return t.Retry(func() error {
		resp, err := t.ebCheckStock(t.Url.String(), pid)
		if err != nil {
	    return err
	  }
		body, err := readBodyBytes(resp)
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			return err
		}
		sizes := doc.Find("a.sizeBox")
		availSizeSkus := [][2]string{}
		var taskSku string
		var size string
		if len(sizes.Nodes) == 0 {
			size = "No Size"
			taskSku = "1"
		} else {
			for idx, _ := range sizes.Nodes {
				sel := sizes.Eq(idx)
				if sel.HasClass("sizeBoxInactive") {
					continue
				}
				idSplit := strings.Split(sel.AttrOr("id", ""), "sizeBox")
				if len(idSplit) == 2 {
					sizeSku := idSplit[1]
          if len(t.Sizes) == 0 || Any(t.Sizes, func(size string) bool { return size == sel.Text() }) {
					 availSizeSkus = append(availSizeSkus, [2]string{sizeSku, sel.Text()})
          }
				}
			}
			if len(availSizeSkus) > 0 {
				sku := availSizeSkus[rand.Intn(len(availSizeSkus))]
				taskSku = sku[0]
				size = sku[1]
			}
			if taskSku == "" || size == "" {
				return errors.New("Out of Stock")
			}
		}

    err = t.Retry(func() error {
  		return t.ebAtc1(pid, taskSku, attr1, attr2)
    }, retry.Delay(256 * time.Millisecond))
    if err != nil {
      return err
    }

    return t.Retry(func() error {
  		err := t.ebCheckout()
  		if err != nil {
  			errStr := fmt.Sprintf("%v", err)
  			if strings.Contains(errStr, "Decline") {
  				go t.DeclineWebhook("Eblens", productTitle, size)
  				return retry.Unrecoverable(retry.Unrecoverable(err))
  			} else if strings.Contains(errStr, "INVALID CARD INFO") {
  				return retry.Unrecoverable(retry.Unrecoverable(err))
  			} else {
  				return err
  			}
  		} else {
		    go t.SuccessWebhook("Eblens", productTitle, size)
        return nil
      }
    }, retry.Delay(256 * time.Millisecond))
	},
	retry.Delay(256 * time.Millisecond), retry.Attempts(1e9), retry.DelayType(retry.FixedDelay), retry.LastErrorOnly(true),
	retry.OnRetry(func(n uint, err error) {
		t.StatusCh <- Status{Message: fmt.Sprintf("%v", err)}
	}),)
}

func (t *CheckoutTask) ebCheckStock(productUrl string, pid string) (*http.Response, error) {
	t.StatusCh <- Status{Message: "Checking Stock"}
  url_, err := url.Parse("https://www.eblens.com/functions/purchasing.cfc?method=init&random=97")
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
  }
  headers := [][2]string {
    {"accept", "text/html, */*; q=0.01"},
    {"x-requested-with", "XMLHttpRequest"},
    {"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
    {"origin", "https://www.eblens.com"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", productUrl},
  }

  params := [][2]string {
    {"id", pid},
    {"attr1", "NULL"},
    {"attr2", "NULL"},
    {"qty", "1"},
    {"mode", "FULL"},
    {"package", "NO"},
  }
  body := urlEncodeParams(params)
  return t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
}

func (t *CheckoutTask) ebAtc1(pid string, sku string, attr1 string, attr2 string) error {
	t.StatusCh <- Status{Message: "Adding to Cart"}
  url_, err := url.Parse(fmt.Sprintf(`https://www.eblens.com/index.cfm?action=ADDTOCART&height=470` +
  	`&width=470&style_id=%s&siz=%s&attr1=%s&attr2=%s&qty=1&package=0&_=%d`, pid, sku, attr1, attr2, timeMillis()))
  if err != nil {
    return err
  }
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
    {"referer", t.Url.String()},
  }
  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
  if err != nil {
  	return err
  }
  body, err := readBodyBytes(resp)
  if err != nil {
  	return err
  }
  respJson := map[string]interface{}{}
  err = json.Unmarshal(body, &respJson)
  if err != nil {
  	return err
  }
  status, ok := respJson["STATUS"].(string)
  if !ok || status != "SUCCESS" {
  	msg, ok := respJson["MESSAGE"].(string)
  	if ok {
  		return errors.New(fmt.Sprintf("ATC failed: %s", msg))
  	} else {
  		return errors.New(fmt.Sprintf("ATC failed: %s", string(body)))
  	}
  }

  return nil
}

func (t *CheckoutTask) ebCheckout() error {
	t.StatusCh <- Status{Message: "Checkout 1/4: Shipping"}
  url_, err := url.Parse("https://www.eblens.com/index.cfm?action=processNewCustomer")
  if err != nil {
    return err
  }
  headerOrder := []string {
    "content-length",
    "cache-control",
    "upgrade-insecure-requests",
    "origin",
    "content-type",
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
    {"cache-control", "max-age=0"},
    {"upgrade-insecure-requests", "1"},
    {"origin", "https://www.eblens.com"},
    {"content-type", "application/x-www-form-urlencoded"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"referer", "https://www.eblens.com/new_customer.cfm"},
  }
  params := [][2]string {
    {"first_name", t.Profile.ShippingAddress.FirstName},
    {"last_name", t.Profile.ShippingAddress.LastName},
    {"Email1", t.Profile.Email},
    {"Email2", t.Profile.Email},
    {"phone1", t.Profile.ShippingAddress.Phone},
    {"passwrd", t.Profile.ShippingAddress.Phone},
    {"address1", t.Profile.ShippingAddress.Address1},
    {"address2", t.Profile.ShippingAddress.Address2},
    {"City", t.Profile.ShippingAddress.City},
    {"state", t.Profile.ShippingAddress.State},
    {"ZIP", t.Profile.ShippingAddress.Zip},
    {"country",t.Profile.ShippingAddress.Country},
    {"locality", ""},
    {"returnto", ""},
    {"Username", ""},
    {"Company", ""},
    {"BYEAR", ""},
    {"BDAY", ""},
    {"BMONTH", ""},
    {"CAN_EMAIL", ""},
    {"SubmitInfo", ""},
    {"selectValidate", ""},
  }
  body := urlEncodeParams(params)
  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  if err != nil {
  	return err
  }
  if resp.StatusCode != 302 {
  	return errors.New(fmt.Sprintf("Invalid status code %d", resp.StatusCode))
  }

  payTxId, err := t.ebCheckoutGetTxSetupID()
  if err != nil {
  	return err
  }
  return t.ebPay(payTxId)
}

func (t *CheckoutTask) ebCheckoutGetTxSetupID() (string, error) {
	t.StatusCh <- Status{Message: "Checkout 2/4: Billing"}
  url_, err := url.Parse("https://www.eblens.com/checkout.cfm")
  if err != nil {
    return "", err
  }
  headerOrder := []string {
    "cache-control",
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
    {"cache-control", "max-age=0"},
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"referer", "https://www.eblens.com/new_customer.cfm"},
  }
  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
  if err != nil {
  	return "", err
  }
  body, err := readBodyBytes(resp)
  if err != nil {
  	return "", err
  }
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	attr := doc.Find("iframe").First().AttrOr("src", "")
	paymentSrcSplit := strings.Split(attr, "TransactionSetupID=")
	if len(paymentSrcSplit) < 2 {
		return "", errors.New("Missing payment transaction ID")
	}
	return paymentSrcSplit[1], nil
}

func (t *CheckoutTask) ebPay(payTxId string) error {
	t.StatusCh <- Status{Message: "Checkout 3/4: Payment"}
	client, err := t.newHttpClient()
	if err != nil {
    return err
  }
  url_, err := url.Parse("https://transaction.hostedpayments.com/?TransactionSetupID=" + payTxId)
  if err != nil {
    return err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Upgrade-Insecure-Requests",
    "User-Agent",
    "Accept",
    "Sec-Fetch-Site",
    "Sec-Fetch-Mode",
    "Sec-Fetch-Dest",
    "Referer",
    "Accept-Encoding",
    "Accept-Language",
  }
  headers := [][2]string {
    {"Connection", "keep-alive"},
    {"Upgrade-Insecure-Requests", "1"},
    {"Sec-Fetch-Site", "cross-site"},
    {"Sec-Fetch-Mode", "navigate"},
    {"Sec-Fetch-Dest", "iframe"},
    {"Referer", "https://www.eblens.com/checkout.cfm"},
  }
  resp, err := t.doReq(client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
	if err != nil {
	  return err
	}
	body, err := readBodyBytes(resp)
	if err != nil {
	  return err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}
	input := doc.Find("input")
	params := [][2]string{}
	params = append(params, [2]string{"scriptManager", "upFormHP|processTransactionButton"})
	for idx, _ := range input.Nodes {
		sel := input.Eq(idx)
		name := sel.AttrOr("name", "")
		if name == "" {
			continue
		}
		value := sel.AttrOr("value", "")
		if name == "cardNumber" {
			value = t.Profile.Card.Number
		} else if name == "CVV" {
			value = t.Profile.Card.Cvc
		} else if name == "__EVENTTARGET" {
			value = "processTransactionButton"
		}
		params = append(params, [2]string{ name, value })
	}
	params = append(params, [2]string { "ddlExpirationMonth", fmt.Sprintf("%02d", t.Profile.Card.ExpMonth) })
	params = append(params, [2]string { "ddlExpirationYear", fmt.Sprintf("%d", t.Profile.Card.ExpYear - 2000) })
	params = append(params, [2]string { "__ASYNCPOST", "true" })

  headerOrder = []string {
    "Host",
    "Connection",
    "Content-Length",
    "Cache-Control",
    "X-Requested-With",
    "X-MicrosoftAjax",
    "User-Agent",
    "Content-Type",
    "Accept",
    "Origin",
    "Sec-Fetch-Site",
    "Sec-Fetch-Mode",
    "Sec-Fetch-Dest",
    "Referer",
    "Accept-Encoding",
    "Accept-Language",
    "Cookie",
  }
  headers = [][2]string {
    {"Connection", "keep-alive"},
    {"Cache-Control", "no-cache"},
    {"X-Requested-With", "XMLHttpRequest"},
    {"X-MicrosoftAjax", "Delta=true"},
    {"Content-Type", "application/x-www-form-urlencoded; charset=UTF-8"},
    {"Accept", "*/*"},
    {"Origin", "https://transaction.hostedpayments.com"},
    {"Sec-Fetch-Site", "same-origin"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", url_.String()},
  }
  body = urlEncodeParams(params)
  resp, err = t.doReq(client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  if err != nil {
  	return err
  }
  payRespBody, err := readBodyBytes(resp)
  if err != nil {
  	return err
  }
  payRespBodyStr := string(payRespBody)
  if strings.Contains(payRespBodyStr, "HostedPaymentStatus=Complete") {
  	re := regexp.MustCompile(`redirect\('(.+?)'`)
  	redirectUrl := strings.Replace(re.FindStringSubmatch(payRespBodyStr)[1], " ", "%20", -1)
  	return t.ebFinalizeOrder(payTxId, redirectUrl)
  } else {
	  doc, err = goquery.NewDocumentFromReader(bytes.NewReader(payRespBody))
	  if err != nil {
	  	return err
	  }
	  errMsgs := doc.Find("span.error").First()
	  if len(errMsgs.Nodes) > 0 {
	  	return errors.New(strings.Replace(errMsgs.Text(), "- ", "", -1))
	  }
	  return errors.New("Unknown Payment Response")
	}
}

func (t *CheckoutTask) ebFinalizeOrder(txId string, redirectUrl string) error {
	t.StatusCh <- Status{Message: "Checkout 4/4: Finalize"}
  url_, err := url.Parse(redirectUrl)
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
    {"sec-fetch-site", "cross-site"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "iframe"},
    {"referer", "https://transaction.hostedpayments.com/?TransactionSetupID=" + txId},
  }
  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
  if err != nil {
  	return err
  }
  body, err := readBodyBytes(resp)
  if err != nil {
  	return err
  }
  doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
  if err != nil {
  	return err
  }
  input := doc.Find("input")
  params := [][2]string{}
  for idx, _ := range input.Nodes {
  	sel := input.Eq(idx)
  	name := sel.AttrOr("name", "")
  	value := sel.AttrOr("value", "")
  	if name == "" {
  		continue
  	}
  	params = append(params, [2]string{ name, value })
  }

  url_, err = url.Parse("https://www.eblens.com/process_order.cfm")
  if err != nil {
    return err
  }
  headerOrder = []string {
    "content-length",
    "cache-control",
    "upgrade-insecure-requests",
    "origin",
    "content-type",
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
  headers = [][2]string {
    {"cache-control", "max-age=0"},
    {"upgrade-insecure-requests", "1"},
    {"origin", "https://www.eblens.com"},
    {"content-type", "application/x-www-form-urlencoded"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"referer", redirectUrl},
  }
  body = urlEncodeParams(params)
  resp, err = t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  if err != nil {
    return err
  } else if resp.StatusCode != 302 {
  	return errors.New(fmt.Sprintf("Unknown order confirmation status %d", resp.StatusCode))
  } else if !strings.Contains(resp.Header.Get("location"), "thankyou") {
  	return errors.New(fmt.Sprintf("Unknown order confirmation redirect: %s", resp.Header.Get("location")))
  } else {
	  url_, err := url.Parse("https://www.eblens.com/thankyou.cfm")
	  if err != nil {
	    return err
	  }
	  headerOrder := []string {
	    "cache-control",
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
	    {"cache-control", "max-age=0"},
	    {"upgrade-insecure-requests", "1"},
	    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	    {"sec-fetch-site", "same-origin"},
	    {"sec-fetch-mode", "navigate"},
	    {"sec-fetch-user", "?1"},
	    {"sec-fetch-dest", "document"},
	    {"referer", redirectUrl},
	  }
	  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
	  if err != nil {
	    return err
	  }
	  body, err := readBody(resp)
	  if err != nil {
	    return err
	  }
	  if !strings.Contains(body, "Thank You For Your Order") {
	  	return errors.New("Could not verify order confirmation")
 	  } else {
	  	return nil
	  }
  }
}

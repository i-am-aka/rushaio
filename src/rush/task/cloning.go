package task

import (
	"bytes"
	"rush/net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"net/url"
	"strings"
	"regexp"
	"github.com/pkg/errors"
)

func (t *CheckoutTask) DoChlsRequest(fn string, substitutions map[*regexp.Regexp]string) (*http.Response, error) {
	req, err := t.MakeChlsRequest(fn, substitutions)
	if err != nil {
		return nil, err
	}

	t.LogDebug("%+v %+v", req, err)
	// b, _ := ioutil.ReadAll(req.Request.Body)
	// t.LogDebug("%s", string(b))
	resp, err := t.doReq2(t.client, req.Request)
	if err != nil {
		return nil, err
	}
	bbraw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	resp.Body = ioutil.NopCloser(bytes.NewReader(bbraw))
	bb, err := readBodyBytes(resp)
	resp.Body = ioutil.NopCloser(bytes.NewReader(bbraw))

	for _, rcand := range *req.Responses {
		if ourresp := rcand.MatchJsonRootKeys(bb); ourresp != nil {
			if resp.StatusCode == 200 {
				return resp, nil
			} else {
				if err, ok := (*ourresp)["error"].(map[string]interface{}); ok {
					return resp, errors.New(err["message"].(string))
				} else if err, ok := (*ourresp)["message"].(string); ok {
					return resp, errors.New(err)
				}
			}
		}
	}
	return resp, ErrUnknownResponse
}

type ResponseCloneSample struct {
	Headers [][2]string
	StatusCode int
	Body string
}

func (r *ResponseCloneSample) MatchJsonRootKeys(body []byte) *map[string]interface{} {
	x := map[string]interface{}{}
	y := map[string]interface{}{}
	json.Unmarshal(body, &x)
	json.Unmarshal([]byte(r.Body), &y)

	if len(x) == 0 || len(y) == 0 {
		return nil
	}
	// keys of root object of body must match all keys in root of collected body
	for yk, _ := range y{
		if _, ok := x[yk]; !ok {
			return nil
		}
	}
	return &x
}

type RequestTemplate struct {
	Request *http.Request
	Cookies [][2]string
	Responses *[]ResponseCloneSample
}


type CharlesSession struct {
	sequence []CharlesRequest
}

func (t *CheckoutTask) LoadCharlesSession(fn string) (*CharlesSession, error) {
	var creqs []CharlesRequest
	bb, err := ioutil.ReadFile(fn)
	if err != nil {
		bb = []byte(fn)
	}
	err = json.Unmarshal(bb, &creqs)
	if err != nil {
		return nil, err
	}
	return &CharlesSession{
		sequence: creqs,
	}, nil
}

func (t *CheckoutTask) MakeChlsRequest2(creq *CharlesRequest, substitutions map[*regexp.Regexp]string) (*RequestTemplate, error) {
	rurl, err := url.Parse(fmt.Sprintf("https://%s%s?%s", creq.Host, creq.Path, creq.Query))
	if err != nil {
		return nil, err
	}

	body := []byte(creq.Request.Body.Text)
	if substitutions != nil {
		for find, replace := range substitutions {
			body = find.ReplaceAll(body, []byte(replace))
		}
	}
	headerOrder := []string{}
	headers := [][2]string{}
	cookies := [][2]string{}


	for _, header := range creq.Request.Header.Headers {
		if strings.HasPrefix(header.Name, ":") {
			// todo parameterizee pho
			continue
		}
		if strings.ToLower(header.Name) == "cookie" {
			kv := strings.Split(header.Value, "=")
			if len(kv) > 1 {
				cookies = append(cookies, [2]string{kv[0], kv[1]})
			}
			continue
		}
		headerOrder = append(headerOrder, header.Name)
		if strings.ToLower(header.Name) == "content-length" {
			continue
		} else {
			for find, replace := range substitutions {
				header.Value = string(find.ReplaceAll([]byte(header.Value), []byte(replace)))
			}
			headers = append(headers, [2]string{header.Name, header.Value})
		}
	}

	resps := []ResponseCloneSample{}
	// for _, creq := range creqs {
		var resp ResponseCloneSample
		resp.StatusCode = creq.Response.Status
		resp.Body = creq.Response.Body.Text
		// if gostruct, err := JsEvalFile(nil, "codegen/json-to-go.js", resp.Body); err == nil {
		// 	t.LogDebug("%s", gostruct)
		// }
		respheaders := [][2]string{}
		for _, header := range creq.Response.Header.Headers {
			respheaders = append(respheaders, [2]string{header.Name, header.Value})
		}
		resp.Headers = respheaders
		resps = append(resps, resp)
	// }

	return &RequestTemplate{
		Request: t.makeReq(creq.Method, rurl, &headers, &headerOrder, &body),
		Cookies: cookies,
		Responses: &resps,
	}, nil

}


func (t *CheckoutTask) MakeChlsRequest(fn string, substitutions map[*regexp.Regexp]string) (*RequestTemplate, error) {
	var creqs []CharlesRequest
	bb, err := ioutil.ReadFile(fn)
	if err != nil {
		bb = []byte(fn)
	}
	err = json.Unmarshal(bb, &creqs)
	if err != nil {
		return nil, err
	}
	creq := creqs[0]

	rurl, err := url.Parse(fmt.Sprintf("https://%s%s?%s", creq.Host, creq.Path, creq.Query))
	if err != nil {
		return nil, err
	}

	body := []byte(creq.Request.Body.Text)
	for find, replace := range substitutions {
		body = find.ReplaceAll(body, []byte(replace))
	}
	headerOrder := []string{}
	headers := [][2]string{}
	cookies := [][2]string{}


	for _, header := range creq.Request.Header.Headers {
		if strings.HasPrefix(header.Name, ":") {
			// todo parameterizee pho
			continue
		}
		if strings.ToLower(header.Name) == "cookie" {
			kv := strings.Split(header.Value, "=")
			if len(kv) > 1 {
				cookies = append(cookies, [2]string{kv[0], kv[1]})
			}
			continue
		}
		headerOrder = append(headerOrder, header.Name)
		if strings.ToLower(header.Name) == "content-length" {
			continue
		} else {
			for find, replace := range substitutions {
				header.Value = string(find.ReplaceAll([]byte(header.Value), []byte(replace)))
			}
			headers = append(headers, [2]string{header.Name, header.Value})
		}
	}

	resps := []ResponseCloneSample{}
	for _, creq := range creqs {
		var resp ResponseCloneSample
		resp.StatusCode = creq.Response.Status
		resp.Body = creq.Response.Body.Text
		// if gostruct, err := JsEvalFile(nil, "codegen/json-to-go.js", resp.Body); err == nil {
		// 	t.LogDebug("%s", gostruct)
		// }
		respheaders := [][2]string{}
		for _, header := range creq.Response.Header.Headers {
			respheaders = append(respheaders, [2]string{header.Name, header.Value})
		}
		resp.Headers = respheaders
		resps = append(resps, resp)
	}
	return &RequestTemplate{
		Request: t.makeReq(creq.Method, rurl, &headers, &headerOrder, &body),
		Cookies: cookies,
		Responses: &resps,
	}, nil
}
/*

*/

type CharlesRequest struct {
	Status          string `json:"status"`
	Method          string `json:"method"`
	ProtocolVersion string `json:"protocolVersion"`
	Scheme          string `json:"scheme"`
	Host            string `json:"host"`
	ActualPort      int    `json:"actualPort"`
	Path            string `json:"path"`
	Query           string `json:"query"`
	Tunnel          bool   `json:"tunnel"`
	KeptAlive       bool   `json:"keptAlive"`
	WebSocket       bool   `json:"webSocket"`
	RemoteAddress   string `json:"remoteAddress"`
	ClientAddress   string `json:"clientAddress"`
	ClientPort      int    `json:"clientPort"`
	Times           struct {
		Start           string `json:"start"`
		RequestBegin    string `json:"requestBegin"`
		RequestComplete string `json:"requestComplete"`
		ResponseBegin   string `json:"responseBegin"`
		End             string `json:"end"`
	} `json:"times"`
	Durations struct {
		Total    int `json:"total"`
		DNS      int `json:"dns"`
		Connect  int `json:"connect"`
		Ssl      int `json:"ssl"`
		Request  int `json:"request"`
		Response int `json:"response"`
		Latency  int `json:"latency"`
	} `json:"durations"`
	Speeds struct {
		Overall  int `json:"overall"`
		Request  int `json:"request"`
		Response int `json:"response"`
	} `json:"speeds"`
	TotalSize int `json:"totalSize"`
	Ssl       struct {
		Protocol    string `json:"protocol"`
		CipherSuite string `json:"cipherSuite"`
	} `json:"ssl"`
	Alpn struct {
		Protocol string `json:"protocol"`
	} `json:"alpn"`
	Request struct {
		Sizes struct {
			Handshake int `json:"handshake"`
			Headers   int `json:"headers"`
			Body      int `json:"body"`
		} `json:"sizes"`
		MimeType        string      `json:"mimeType"`
		Charset         interface{} `json:"charset"`
		ContentEncoding interface{} `json:"contentEncoding"`
		Header          struct {
			Headers []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"headers"`
		} `json:"header"`
		Body struct {
			Text    string      `json:"text"`
			Charset interface{} `json:"charset"`
		} `json:"body"`
	} `json:"request"`
	Response struct {
		Status int `json:"status"`
		Sizes  struct {
			Handshake int `json:"handshake"`
			Headers   int `json:"headers"`
			Body      int `json:"body"`
		} `json:"sizes"`
		MimeType        string `json:"mimeType"`
		Charset         string `json:"charset"`
		ContentEncoding string `json:"contentEncoding"`
		Header          struct {
			Headers []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"headers"`
		} `json:"header"`
		Body struct {
			Text    string `json:"text"`
			Charset string `json:"charset"`
			Decoded bool   `json:"decoded"`
		} `json:"body"`
	} `json:"response"`
}
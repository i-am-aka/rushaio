package main

import (
	"bytes"
	"encoding/base64"
	"time"
	"compress/gzip"
	"github.com/andybalholm/brotli"
	"os"
	"strings"
	"context"
	"fmt"
	"log"
	"net"
	"flag"
	"net/http"
	"net/url"
	// "regexp"
	"io"
	"io/ioutil"
	// "bytes"
	// "github.com/elazarl/goproxy/ext/auth"
	"github.com/elazarl/goproxy"
)

// func orPanic(err error) {
// 	if err != nil {
// 		panic(err)
// 	}
// }

// TODO
// test with authed upstream
// implement host override

type ProxyPassThru struct {
	listener net.Listener
	cancel context.CancelFunc
	ctx context.Context
}

var HOST_HIJACK = map[string]bool{
	"recaptcha-demo.appspot.com:443":true,
	"www.finishline.com:443" :true,
	"www.jdsports.com:443":true,
	"www.yeezysupply.com:443":true,
	"score.rushaio.co:443":true,
	"rushaio.co:443":true,

}

func readBodyBytes(resp *http.Response) ([]byte, error) {
	var reader io.ReadCloser
	var err error
	switch resp.Header.Get("Content-Encoding") {
	case "br":
		reader = ioutil.NopCloser(brotli.NewReader(resp.Body))
	case "gzip":
    reader, err = gzip.NewReader(resp.Body)
	default:
	  reader = resp.Body
	}
	if reader != nil && err == nil {
		defer reader.Close()
		bb, err := ioutil.ReadAll(reader)
		if os.Getenv("DEBUG") == "1" && os.Getenv("RESPDUMP") != "0" {
			log.Printf("RESP: %s", string(bb))
		}
		return bb, err
	} else {
		return []byte{}, nil
	}
}


func NewProxyPassThru(ctx context.Context, upstream *url.URL, port string) (*url.URL, error) {
	var err error
	ctx, cancel := context.WithCancel(ctx)
	p := ProxyPassThru{ctx: ctx, cancel: cancel}
	p.listener, err = net.Listen("tcp", "127.0.0.1:" + port)
	if err != nil {
		return nil, err
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = os.Getenv("MITMDEBUG") == "1"

	if upstream != nil {
		proxy.Tr = &http.Transport{Proxy: func(req *http.Request) (*url.URL, error) {
			return upstream, nil
		}}
		proxy.ConnectDial = proxy.NewConnectDialToProxyWithHandler("http://" + upstream.Host, func(req *http.Request) {
			pw, _ := upstream.User.Password()
			SetBasicAuth(upstream.User.Username(), pw, req)
		})
	}

	// proxy.OnRequest().HandleConnect(goproxy.FuncHttpsHandler(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	// 	log.Println(host)
	// 	if _, ok := HOST_HIJACK[host]; ok {
	// 		return goproxy.MitmConnect, host
	// 		// return &goproxy.ConnectAction{Action: goproxy.ConnectHijack, Hijack: func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
	// 		// 	resp := &http.Response{
	// 		// 		Status: "200 OK",
	// 		// 		StatusCode: 200,
	// 		// 		Proto: "HTTP/2",
	// 		// 		Body: ioutil.NopCloser(bytes.NewReader([]byte("<html><head></head><body></body></html>"))),
	// 		// 		Header: http.Header{
	// 		// 			"Content_Type": []string{"text/html"},
	// 		// 			"Date": []string{"Fri, 02 Oct 2020 05:02:37 GMT"},
	// 		// 			"Server": []string{"AmazonS3"},
	// 		// 		},
	// 		// 		// ContentLength: 0,
	// 		// 		ProtoMajor: 2,
	// 		// 		ProtoMinor: 0,
	// 		// 	}
	// 		// 	resp.Write(client)
	// 		// 	client.Close()
	// 		// }}, host
	// 	} else {
	// 		return goproxy.HTTPMitmConnect, host
	// 	}
	// 	// TODO host override for harvester
	// }))
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		// log.Println(req.URL.Host)
		if os.Getenv("MITM") == "0" || (strings.HasPrefix(req.URL.Host, "www.google.com") || strings.HasPrefix(req.URL.Host, "www.gstatic.com")) {
			req.URL.Host = strings.Split(req.URL.Host, ":")[0]
			return req, nil
		} else {
			return nil, goproxy.TextResponse(req, "<html><head></head><body></body></html>")
		}
	})
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		log.Printf("Response %s", resp.Request.URL.String())
		lps := strings.Split(resp.Request.URL.Path, "/")
		lp := lps[len(lps)-1]
		if lp == "payload" || lp == "reload" {
			if borig, err := ioutil.ReadAll(resp.Body); err == nil {
				resp.Body = ioutil.NopCloser(bytes.NewReader(borig))
				if payload, err := readBodyBytes(resp); err == nil {
					ext := ".jpg"
					if lp == "reload" {
						ext = ""
					}
					ioutil.WriteFile(fmt.Sprintf("cap/%s-%d%s", lp, time.Now().UnixNano(), ext), payload, 0644)
				} else {
					log.Printf("read err: %+v")
				}
				resp.Body = ioutil.NopCloser(bytes.NewReader(borig))
			}
		}
		// if resp.Request.URL.Path
		resp.Header.Set("Access-Control-Allow-Origin", "*")
		return resp
	})
	go func() {
    err := http.Serve(p.listener, proxy)
    log.Printf("SERVER ENDED %+v", err)
  }()
	go func() {
		log.Println("START SERV MANAGER")
		select {
			case <-ctx.Done():
				p.listener.Close()
		}
		log.Println("END SERV MANAGER")

	}()

	return url.Parse("http://" + p.listener.Addr().String())
}

func SetBasicAuth(username, password string, req *http.Request) {
	req.Header.Set("Proxy-Authorization", fmt.Sprintf("Basic %s", basicAuth(username, password)))
}

func basicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func main() {
	upstream := flag.String("upstream", "", "upstream")
	proxport := flag.String("proxport", "", "proxport")
	flag.Parse()

	ctx := context.Background()
	if upurl, err := url.Parse(*upstream); err == nil {

		if len(upurl.Host) == 0 {
			upurl = nil
		}
		if purl, err := NewProxyPassThru(ctx, upurl, *proxport); err == nil {
			log.Printf("listening at %+v", purl)
		} else {
			log.Fatalf("%+v", err)
		}
	}

	select {
		case <-ctx.Done():
			return
	}
}
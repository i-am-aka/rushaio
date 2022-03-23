package task


import (
	"encoding/base64"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	// "regexp"
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

func NewProxyPassThru(ctx context.Context, upstream *url.URL) (*url.URL, error) {
	var err error
	ctx, cancel := context.WithCancel(ctx)
	p := ProxyPassThru{ctx: ctx, cancel: cancel}
	p.listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	proxy.Tr = &http.Transport{Proxy: func(req *http.Request) (*url.URL, error) {
		return upstream, nil
	}}
	proxy.ConnectDial = proxy.NewConnectDialToProxyWithHandler("http://" + upstream.Host, func(req *http.Request) {
		pw, _ := upstream.User.Password()
		SetBasicAuth(upstream.User.Username(), pw, req)
	})
	proxy.OnRequest().HandleConnect(goproxy.FuncHttpsHandler(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		// TODO host override for harvester
		return goproxy.OkConnect, host
	}))
	go func() {
		defer recover()
    err := http.Serve(p.listener, proxy)
    log.Printf("SERVER ENDED %+v", err)
  }()
	go func() {
		defer recover()
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
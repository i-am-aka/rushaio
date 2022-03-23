package rcdp

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"log"
	"net"
	"io/ioutil"
	"net/http"
	"net/url"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/page"
	"github.com/elazarl/goproxy"
	"context"
	"reflect"
	"runtime"
)

const (
	ProxyAuthHeader = "Proxy-Authorization"
)

func SetBasicAuth(credentials string, req *http.Request) {
	// log.Println(fmt.Sprintf("Basic %s", basicAuth(credentials)))
	req.Header.Set(ProxyAuthHeader, fmt.Sprintf("Basic %s", basicAuth(credentials)))
}

func basicAuth(credentials string) string {
	return base64.StdEncoding.EncodeToString([]byte(credentials))
}

type CDPContext struct {
	Ctx context.Context
	Cancel context.CancelFunc

	BrowserCtx context.Context
	proxyServer *goproxy.ProxyHttpServer
	proxyHttpServer *http.Server
	upstreamProxy *url.URL

	conns []*net.Conn
}

func NewCDPContext(headful bool) (*CDPContext, error) {
	// log.Println("NewCDPContext")
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	  signal.Notify(c, os.Interrupt)
	  go func() {
	    select {
	    case <-c:
	      cancel()
	    case <-ctx.Done():
	    }
	  }()

	listener, err := net.Listen("tcp4", ":0")
	if err != nil {
		return nil, err
	}
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}

	rc := CDPContext{Ctx: ctx, Cancel: cancel,}
	rc.StartProxyServer(listener.Addr().String())

	defaultOpts := chromedp.DefaultExecAllocatorOptions[:]
	if headful || os.Getenv("HEADFUL") != "" {
		defaultOpts = remove(defaultOpts, indexOfEAOption(chromedp.Headless, defaultOpts))
	}

	// log.Printf(fmt.Sprintf("http://localhost:%d", listener.Addr().(*net.TCPAddr).Port))
	var ua string
	if runtime.GOOS == "windows" {
		ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.92 Safari/537.36"
	} else {
		ua = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.92 Safari/537.36"
	}
	opts := append(defaultOpts,
		// chromedp.DisableGPU, // examine perf with/without
		// chromedp.Flag("auto-open-devtools-for-tabs", true),
		chromedp.UserDataDir(dir),
		chromedp.UserAgent(ua),
		chromedp.ProxyServer(fmt.Sprintf("localhost:%d", listener.Addr().(*net.TCPAddr).Port)),
	)
	if os.Getenv("CHROMIUM") != "" {
		opts = append(opts, chromedp.ExecPath(os.Getenv("CHROMIUM")))
	}

	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)

	// also set up a custom logger
	logEnabled := os.Getenv("RCDEBUG") != ""
	fn := func(format string, a ...interface{}) {
		if logEnabled {
			fmt.Println(fmt.Sprintf(format, a...))
		}
	}

	// start browser
	BrowserCtx, _ := chromedp.NewContext(allocCtx, chromedp.WithErrorf(fn),
		chromedp.WithLogf(fn),
		chromedp.WithDebugf(fn))

	if err := chromedp.Run(BrowserCtx, SpoofHeadfulAction(),); err != nil {
		return nil, err
	}

	rc.BrowserCtx = BrowserCtx
	return &rc, nil
}

func remove(s []chromedp.ExecAllocatorOption, i int) []chromedp.ExecAllocatorOption {
    s[len(s)-1], s[i] = s[i], s[len(s)-1]
    return s[:len(s)-1]
}

func indexOfEAOption(element chromedp.ExecAllocatorOption, data []chromedp.ExecAllocatorOption) (int) {
   for k, v := range data {
       if runtime.FuncForPC(reflect.ValueOf(element).Pointer()).Name() == runtime.FuncForPC(reflect.ValueOf(v).Pointer()).Name() {
           return k
       }
   }
   return -1    //not found.
}

func SpoofHeadfulAction() chromedp.Action {
	const script = `(function(w, n, wn) {
	  // Pass the Webdriver Test.
	  const newProto = wn.__proto__;
	  delete newProto.webdriver;
	  wn.__proto__ = newProto;

	  // Object.defineProperty(n, 'webdriver', {
	  //   get: () => false,
	  // });

	  // // Pass the Plugins Length Test.
	  // // Overwrite the plugins property to use a custom getter.
	  // Object.defineProperty(n, 'plugins', {
	  //   // This just needs to have length > 0 for the current test,
	  //   // but we could mock the plugins too if necessary.
	  //   get: () => [1, 2, 3, 4, 5],
	  // });

	  // // Pass the Languages Test.
	  // // Overwrite the plugins property to use a custom getter.
	  // Object.defineProperty(n, 'languages', {
	  //   get: () => ['en-US', 'en'],
	  // });

	  // // Pass the Chrome Test.
	  // // We can mock this in as much depth as we need for the test.
	  // w.chrome = {
	  //   runtime: {},
	  // };

	  // window.navigator.plugins = [
	  //   'Chrome PDF Plugin',
	  //   'Chrome PDF Viewer',
	  //   'Native Client'
	  // ];

	  // const PERMISSIONS = {
	  //   "geolocation": "prompt",
	  //   "notifications": "prompt",
	  //   "push": {raise: "Foo"},
	  //   "midi": "granted",
	  //   "camera": "prompt",
	  //   "microphone": "prompt",
	  //   "speaker": {raise: "is not a valid enum value of type PermissionName"},
	  //   "device-info": {raise: "is not a valid enum value of type PermissionName"},
	  //   "background-sync": "granted",
	  //   "bluetooth": {raise: "is not a valid enum value of type PermissionName"},
	  //   "persistent-storage": "prompt",
	  //   "ambient-light-sensor": {raise: "Foo"},
	  //   "accelerometer": "granted",
	  //   "gyroscope": "granted",
	  //   "magnetometer": "granted",
	  //   "clipboard": {raise: "is not a valid enum value of type PermissionName"},
	  //   "accessibility-events": {raise: "Foo"},
	  //   "clipboard-read": "prompt",
	  //   "clipboard-write": "granted",
	  //   "payment-handler": "granted",
	  // }

	  // const originalQuery = window.navigator.permissions.query;
	  // window.navigator.permissions.query = (parameters) => {
	  //   if (parameters.name in PERMISSIONS) {
	  //     if (typeof PERMISSIONS[parameters.name] === 'string') {
	  //       return Promise.resolve({state: PERMISSIONS[parameters.name]});
	  //     } else {
	  //       return Promise.reject(new Error(PERMISSIONS[parameters.name].raise));
	  //     }
	  //   } else {
	  //   	console.log('query orig perms', parameters.name)
	  //     return originalQuery(parameters)
	  //   }
	  // };


	})(window, navigator, window.navigator);`

	return chromedp.ActionFunc(func(ctx context.Context) error {
		_, err := page.AddScriptToEvaluateOnNewDocument(script).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	})
}


func (rc *CDPContext) StartProxyServer(addr string) {
	rc.proxyServer = goproxy.NewProxyHttpServer()
	rc.proxyServer.Verbose = os.Getenv("CDPDEBUG") == "1"

	rc.proxyHttpServer = &http.Server{Addr: addr, Handler: rc.proxyServer}
	rc.proxyHttpServer.SetKeepAlivesEnabled(false)
	go rc.proxyHttpServer.ListenAndServe()

	go func() {
		select {
		case <-rc.Ctx.Done():
			// log.Println("rc.ctx.Done. Shutting down proxy server...")
			go rc.proxyHttpServer.Shutdown(rc.Ctx)
		}
	}()
}

func (rc *CDPContext) closeConns() {
	if len(rc.conns) > 0 {
		for _, conn := range rc.conns {
			defer func() {
				if err := recover(); err != nil {
					// log.Printf("SetUpstreamProxy panic: %+v", err)
					// return err
				}
			}()
			if conn != nil {
				(*conn).Close()
			}
		}
	}
}
func (rc *CDPContext) SetUpstreamProxy(upstream string) error {
	if len(rc.conns) > 0 {
		for _, conn := range rc.conns {
			defer func() {
				if err := recover(); err != nil {
					// log.Printf("SetUpstreamProxy panic: %+v", err)
				}
			}()
			if conn != nil {
				(*conn).Close()
			}
		}
	}

	if upstream == "" || upstream == "http://localhost" {
		log.Printf("set local upstream %s", upstream)
		rc.proxyServer = goproxy.NewProxyHttpServer()
		rc.proxyServer.Verbose = os.Getenv("CDPDEBUG") == "1"
		rc.proxyServer.Tr.DisableKeepAlives = true
		connectDial := func(network, addr string) (net.Conn, error) {
			conn, err := net.Dial(network, addr)
			log.Printf("new local conn: %+v", conn)
			rc.conns = append(rc.conns, &conn)
			return conn, err
		}
		rc.proxyServer.ConnectDial = connectDial
		rc.proxyHttpServer.Handler = rc.proxyServer
		return nil
	}

	log.Printf("set proxy upstream %s", upstream)

	upstreamProxyUrl, err := url.Parse(upstream)
	if err != nil {
		return err
	}

	rc.proxyServer = goproxy.NewProxyHttpServer()
	rc.proxyServer.Verbose = os.Getenv("CDPDEBUG") == "1"
	rc.proxyServer.Tr.DisableKeepAlives = true
	rc.proxyServer.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		return upstreamProxyUrl, nil
	}
	connectReqHandler := func(req *http.Request) {
		log.Printf("auth: %s", upstreamProxyUrl.User.String())
		if upstreamProxyUrl.User.String() != "" {
			SetBasicAuth(upstreamProxyUrl.User.String(), req)
		}
	}
	connectDial := func(network, addr string) (net.Conn, error) {
		dial := rc.proxyServer.NewConnectDialToProxyWithHandler("http://" + upstreamProxyUrl.Host, connectReqHandler)
		conn, err := dial(network, addr)
		if err != nil {
			rc.closeConns()
			rc.Cancel()
		}
		log.Printf("new proxy conn: %+v err: %+v", conn, err)
		rc.conns = append(rc.conns, &conn)
		return conn, err
	}
	rc.proxyServer.ConnectDial = connectDial
	rc.proxyServer.OnRequest().Do(goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		log.Printf("auth: %s", upstreamProxyUrl.User.String())
		if upstreamProxyUrl.User.String() != "" {
			SetBasicAuth(upstreamProxyUrl.User.String(), req)
		}
		return req, nil
	}))

	rc.proxyHttpServer.Handler = rc.proxyServer
	return nil
}

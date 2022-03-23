package main

import (
	"encoding/json"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"log"
	"io/ioutil"
	"sync"
	"gopkg.in/src-d/go-vitess.v1/pools"
	"net/http"
	"github.com/chromedp/cdproto/network"
	"time"
	"rush/util"
)

type Session struct {
	Email string `json:"email"`
	Cookies []network.SetCookieParams `json:"cookies"`
}

type RCPool struct {
	pool *pools.ResourcePool
	sessions []Session
	ctx context.Context
	session_idx int
	sessionMut sync.Mutex
	proxies []string
	proxy_idx int
	proxyMut sync.Mutex
}

type TokenRequest struct {
	Url string `json:"url"`
	ApiKey string `json:"apiKey"`
	Action string `json:"action"`
	Proxy string `json:"proxy"`
	Session Session `json:"session"`
}

func (rcpool *RCPool) http_handler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "400: could not read body", 400)
		return
	}
	var tokenReq TokenRequest
	err = json.Unmarshal(body, &tokenReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("400: could not decode json body: %+v", err), 400)
		return
	}

	log.Println("Get")
	res, err := rcpool.pool.Get(rcpool.ctx)
	log.Println("Got")

	if err != nil {
		http.Error(w, "500: could not get RCContext", 500)
		return
	}
	rc := res.(*RCContext)
	if rc.cdpCtx.Ctx.Err() != nil {
		rcpool.pool.Put(nil)
		http.Error(w, "500: could not get token. please try again", 500)
		return
	} else {
		defer rcpool.pool.Put(res)
	}

	var proxy string
	if tokenReq.Proxy == "none" {
		proxy = ""
	} else if tokenReq.Proxy != "" {
		proxy = tokenReq.Proxy
	} else if len(rcpool.proxies) > 0 {
		rcpool.proxyMut.Lock()
		proxy = rcpool.proxies[rcpool.proxy_idx]
		rcpool.proxy_idx++
		if rcpool.proxy_idx >= len(rcpool.proxies) {
			rcpool.proxy_idx = 0
		}
		rcpool.proxyMut.Unlock()
	}

	err = rc.cdpCtx.SetUpstreamProxy(proxy)
	if err != nil {
		http.Error(w, "500: bad proxy", 500)
		return
	}

	var email string
	var cookies []network.SetCookieParams
	if len(tokenReq.Session.Cookies) > 0 {
		cookies = tokenReq.Session.Cookies
	} else if len(rcpool.sessions) > 0 {
		rcpool.sessionMut.Lock()
		session := rcpool.sessions[rcpool.session_idx]
		cookies = session.Cookies
		email = session.Email
		rcpool.session_idx++
		if rcpool.session_idx >= len(rcpool.sessions) {
			rcpool.session_idx = 0
		}
		rcpool.sessionMut.Unlock()
	}

	log.Printf("proxy %s email %s url %s action %s apiKey %s", proxy, email, tokenReq.Url, tokenReq.Action, tokenReq.ApiKey)
	token, err := rc.GetToken(tokenReq.Url, tokenReq.Action, tokenReq.ApiKey, cookies)
	if err != nil {
		http.Error(w, "500: could not get token", 500)
		return
	}
	fmt.Fprintf(w, token)
}

func logWait(t time.Time) {}

func (rcpool *RCPool) WorkerFactory() (pools.Resource, error) {
	ctx, err := NewRCContext()
	return ctx, err
}

func health(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Healthy")
}

func main() {
	ctx := context.Background()
	workers := flag.Int("workers", 1, "number of chrome processes")
	sessionsJson  := flag.String("sessionsJson", "", "string or path to file containing sessions (json encoded array of (array of rclogin outputs))")
	proxiesJson := flag.String("proxiesJson", "", "string or path to file containing json encoded array of proxy strings")
	flag.Parse()
	var sessions []Session
	if *sessionsJson != "" {
		err := json.Unmarshal(util.ReadFileOrStr(*sessionsJson), &sessions)
		if err != nil {
			log.Fatalf("could not load sessions: %+v", err)
		}
	}
	var proxies []string
	if *proxiesJson != "" {
			err := json.Unmarshal(util.ReadFileOrStr(*proxiesJson), &proxies)
		if err != nil {
			log.Fatalf("could not load proxies: %+v", err)
		}
	}
	rcpool := &RCPool{
		ctx: ctx,
		sessions: sessions,
		session_idx: 0,
		proxies: proxies,
		proxy_idx: 0,
	}
	rcpool.pool = pools.NewResourcePool(rcpool.WorkerFactory, *workers, *workers, 120 * time.Second)
	defer rcpool.pool.Close()
	log.Println("Get")
	res, err := rcpool.pool.Get(rcpool.ctx)
	log.Println("Got")
	if err != nil {
		log.Fatalf("%+v", err)
		return
	}
	if os.Getenv("WAIT") == "" {
		rcpool.pool.Put(res)
	}

	httpServerExitDone := &sync.WaitGroup{}
  httpServerExitDone.Add(1)

	addr := ":42424"
	log.Println("listen on", addr)
  srv := &http.Server{Addr: addr}

	http.HandleFunc("/", rcpool.http_handler)
	http.HandleFunc("/health", health)

  go func() {
      defer httpServerExitDone.Done()
      if err := srv.ListenAndServe(); err != http.ErrServerClosed {
          log.Fatalf("ListenAndServe(): %v", err)
      }
  }()

  c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)

    go func() {
      select {
      case <-c:
      	log.Println("rcpool: interrupt")
				srv.Shutdown(context.TODO())
				rcpool.pool.Close()
      case <-ctx.Done():
      }
    }()

		httpServerExitDone.Wait()
}

package main

import (
	// "github.com/getsentry/sentry-go"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
	"fmt"

	// "io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"math"
	"strconv"
	"net/http"
	// "strings"

	"os"
	"os/exec"
	"os/signal"
	"time"
)

var PX_SOCK = "../js/px.sock"
var upgrader = websocket.Upgrader{} // use default options

// type PxParams struct {
// 	AppId string
// 	JsSrc string
// 	Url string
// 	Cookie string
// 	UserAgent string
// 	XhrProxySock string
// }

type Px3Request struct {
	Cookie string
	AppId string
	Vid string
	Uuid string
	Url string
	Host string
	Proxy string
	UserAgent string
	JsSrc string
	CapJsSrc string
	XhrProxySock string
	RecapToken string
}

func (a *Api) LogDebug(fmtstr string, args ...interface{}) {
	if a.Debug {
		log.Printf(fmtstr, args...)
	}
}

func (a *Api) gen(w http.ResponseWriter, r *http.Request) {
	a.LogDebug("%+v", r.URL.String())

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		a.LogDebug("upgrade:", err)
		return
	}
	defer c.Close()

	var px3req Px3Request
	if _, p, err := c.ReadMessage(); err == nil {
		a.LogDebug("recv %s", string(p))
		json.Unmarshal(p, &px3req)
		a.LogDebug("px3req=%+v", px3req)
	}

	if px3req.UserAgent == "" || px3req.JsSrc == "" || px3req.AppId == "" || px3req.Url == "" {
		a.LogDebug("read err %+v", err)
		http.Error(w, fmt.Sprintf("400: invalid params (need ua, jsUrl, appId, url)"), 400)
		return
	}


	stmpf, err := ioutil.TempFile("", "xhrsock")
	if err != nil {
		return
	}
	stmpf.Close()
	os.Remove(stmpf.Name())

	sock, err := net.ListenUnix("unix", &net.UnixAddr{stmpf.Name(), "unix"})

	if err != nil {
		return
	}
	defer sock.Close()



	go func() {
		conn, err := sock.AcceptUnix()
		if err != nil {
			a.LogDebug("sock err %+v", err)
			return
		}
		defer conn.Close()
		a.LogDebug("got conn %+v", conn)

		go func() {
			defer conn.Close()
			for {
				var recvbuf = make([]byte, 32768)
				_, err = conn.Read(recvbuf)
				if err != nil {
					a.LogDebug("sock err %+v", err)
					return
				}
				recvbuf = bytes.Trim(recvbuf, "\x00")
				var req FetchRequest
				err = json.Unmarshal(recvbuf, &req)
				if err != nil {
					a.LogDebug("req unmarshal err %+v", err)
					return
				}
				a.LogDebug("recv request %+v", req)
				err = c.WriteJSON(req)
				if err != nil {
					a.LogDebug("write err %+v", err)
					return
				}
			}
		}()

		for {
			_, p, err := c.ReadMessage()
			if err != nil {
				a.LogDebug("ws err %+v", err)
				return
			}
			buf := make([]byte, len(p) + 32768-len(p))
			a.LogDebug("buflen %d recv from ws: %d %s", len(buf), len(p), string(p))
			copy(buf, p)
			// a.LogDebug("buf %s", string(buf))
			_, err = conn.Write(buf)
			// a.LogDebug("wrote %d", n)
			if err != nil {
				a.LogDebug("sock err %+v", err)
				return
			}
		}
	}()


	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	px3req.XhrProxySock = stmpf.Name()
	pxArgsBytes, err := json.Marshal(px3req)
	if err != nil {
		a.LogDebug("%+v", err)
		return
	}

	var d net.Dialer
	// genc, err := d.DialContext(ctx, "unix", PX_SOCK)
	genc, err := d.DialContext(ctx, "tcp", "127.0.0.1:40400")
	if err != nil {
		a.LogDebug("%+v", err)
		return
	}
	defer genc.Close()
	bs := make([]byte, 4)
	a.LogDebug("len=%d", len(pxArgsBytes))
	binary.LittleEndian.PutUint32(bs, uint32(len(pxArgsBytes)))
	_, err = genc.Write(bs)
	if err != nil {
		a.LogDebug("%+v", err)
		return
	}
	_, err = genc.Write(pxArgsBytes)
	if err != nil {
		a.LogDebug("%+v", err)
		return
	}
	resp, err := ioutil.ReadAll(genc)
	if err != nil {
		a.LogDebug("%+v", err)
		return
	}
	a.LogDebug("%s", string(resp))
	c.WriteMessage(websocket.TextMessage, resp)

	// a.LogDebug("%s", string(pxArgsBytes))

	// cmd := exec.CommandContext(ctx, "deno", "run", "--cached-only", "--allow-all", "--unstable", "env.js", string(pxArgsBytes))
	// if DENO_DIR == "" {
	// 	DENO_DIR = "../../task/zeronaught/deno/"
	// }
	// cmd.Env = append(os.Environ(), "DENO_DIR=" + DENO_DIR)
	// cmd.Env = append(cmd.Env, "XHR_PROXY=" + stmpf.Name())
	// // cmd.Env = append(cmd.Env, "DEBUG=1")
	// // a.LogDebug("%+v", cmd.Env)
	// // cmd.Env = append(os.Environ(), "HTTPS_PROXY=" + )
	// cmd.Dir = "../js"
	// // read, write := io.Pipe()
	// // cmd.Stdout = write
	// // cmd.Stderr = write
 //  // cmd.Start()

 //  // for {
 //  // 	buf := new(bytes.Buffer)
 //  // 	_, err := buf.ReadFrom(read)
 //  // 	if err !=nil {
 //  // 		a.LogDebug("read err %+v", err)
 //  // 		break
 //  // 	}
 //  // 	dat := string(buf.Bytes())
 //  //   a.LogDebug("recv %s", dat)
 //  // }

	// out, err := cmd.Output()
	// if err != nil {
	// 	a.LogDebug("errout %s", string(out))
	//   return
	// }
	// a.LogDebug("out %s", string(out))
	// TODO lock to prevent concurrent writes
	// c.WriteMessage(websocket.TextMessage, out)
}

var denoAbspath, _ = exec.LookPath("deno")
var DENO_DIR = os.Getenv("DENO_DIR")

func get_pxua(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())
	fmt.Fprintf(w, PX_UAS[rand.Intn(len(PX_UAS))])
}

func get_pxmobileua(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())
	fmt.Fprintf(w, PX_MOBILE_UAS[rand.Intn(len(PX_MOBILE_UAS))])
}

func HandleErr(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("500: %+v", err), 500)
}

type FetchRequest struct {
	Headers map[string]string `json:"headers"`
	ReferrerPolicy string `json:"referrerPolicy"`
	Body           string `json:"body"`
	Method         string `json:"method"`
	Mode           string `json:"mode"`
	Url 					 string `json:"url"`
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Onlien")
}

func LaunchGenWorker() error {
	denoAbspath, err := exec.LookPath("deno")
	if err != nil {
	  return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	// DENO_DIR=../../task/zeronaught/deno DEBUG=1 deno run --allow-all --cached-only --unstable --cert ./charles-ssl-proxying-certificate.pem serv.js
	cmd := exec.CommandContext(ctx, denoAbspath, "run", "--allow-all", "--cert", "./charles-ssl-proxying-certificate.pem", "--unstable", "./serv.js")
	// cmd := exec.CommandContext(ctx, denoAbspath, "run", "--allow-all", "--unstable", "./serv.bundle.js")

	cmd.Env = append(os.Environ(), "PX_SOCK=" + PX_SOCK)
	cmd.Env = append(cmd.Env, "DENO_DIR=../../task/zeronaught/deno")
	cmd.Dir = "../js"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err = cmd.Start()
	if err != nil {
		return err
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	go func() {
		err = cmd.Wait()
		log.Fatalf("base gen serv died %+v", err)
	}()
	go func() {
		select {
			case <-c:
			   cancel()
		}
	}()
	return nil
}

type Api struct {
	Debug bool
	MaxClientSemaphore chan struct{}
}

func (api *Api) EnforceMaxClientsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.MaxClientSemaphore <- struct{}{}
		defer func() { <-api.MaxClientSemaphore }()

		next.ServeHTTP(w, r)
	})
}


func main() {
	// err := sentry.Init(sentry.ClientOptions{
	// 	Dsn: "https://cff4b8b7461747e290882beeddd5032d@o366922.ingest.sentry.io/5387282",
	// })
	// if err != nil {
	// 	log.Fatalf("sentry.Init: %s", err)
	// }
	// // Flush buffered events before the program terminates.
	// defer sentry.Flush(2 * time.Second)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	addr := ":" + port
	fmt.Println("listen on", addr)

	if os.Getenv("WORKER") != "0" {
		err := LaunchGenWorker()
		if err != nil {
			log.Fatalf("%+v", err)
		}
	}

	sema := make(chan struct{}, math.MaxInt32)
	if os.Getenv("MAX_CLIENTS") != "" {
		if maxClients, err := strconv.Atoi(os.Getenv("MAX_CLIENTS")); err == nil {
			sema = make(chan struct{}, maxClients)
		}
	}


	r := http.NewServeMux()

	api := Api{MaxClientSemaphore: sema, Debug: os.Getenv("DEBUG") == "1"}
	r.Handle("/gen10", api.EnforceMaxClientsHandler(http.HandlerFunc(api.gen)))
	r.Handle("/pxua", http.HandlerFunc(get_pxua))
	r.Handle("/pxmobileua", http.HandlerFunc(get_pxmobileua))
	r.Handle("/", http.HandlerFunc(home))
	if err := http.ListenAndServe(addr, handlers.CompressHandler(r)); err != http.ErrServerClosed {
	    fmt.Printf("ListenAndServe(): %v", err)
	}
}

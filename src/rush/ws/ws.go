package ws

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/getsentry/sentry-go"
	"time"
	"bytes"
		"crypto/sha256"
		"crypto/tls"
		"crypto/x509"
		"log"
		"net"
		// "net/http"
	// "math"
	"rush/task"
	"net/http"
	// "runtime/pprof"
	"os"
	"runtime"
	// "log"
	"github.com/avast/retry-go"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	// "github.com/rustler47/SecureClient"
)

func timeMillis() uint64 {
	return uint64(time.Now().UnixNano() / int64(time.Millisecond))
}

type AuthStatus struct {
	Jwt string `json:"jwt"`
	ErrorMsg string `json:"error"`
	Error error
}

func cloudSocketHost() string {
  // envCsHost := os.Getenv("RUSH_CSHOST")
  // if envCsHost != "" {
  //   return envCsHost
  // }
  return "wss://auth.rushaio.co/" // ws://localhost:8099
}

// var pinner, err = SecureClient.New([]string{"auth.rushaio.co"}, true, func(p string) { log.Printf("rush 8========D ({deadcode}) %s", p) })
type Dialer func(network, addr string) (net.Conn, error)

func makeDialer(fingerprint []byte, skipCAVerification bool) Dialer {
	return func(network, addr string) (net.Conn, error) {
		c, err := tls.Dial(network, addr, &tls.Config{InsecureSkipVerify: skipCAVerification})
		if err != nil {
			return c, err
		}
		connstate := c.ConnectionState()
		keyPinValid := false
		for _, peercert := range connstate.PeerCertificates {
			der, _ := x509.MarshalPKIXPublicKey(peercert.PublicKey)
			hash := sha256.Sum256(der)
			// log.Println(addr)
			// log.Println(peercert.Issuer)
			// log.Printf("%#v", hash)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			if bytes.Compare(hash[0:], fingerprint) == 0 {
				log.Println("Pinned Key found")
				keyPinValid = true
			}
		}
		if keyPinValid == false {
			go task.DefaultCloudLogger.SendTelemetryNow(map[string]interface{} {
				"event": "authsecping",
				"authsecping": map[string]interface{} {
					"env": os.Environ(),
					"issuer": connstate.PeerCertificates[0].Issuer.String(),
					"addr": c.RemoteAddr().String(),
				},
			})
			return nil, errors.New("hey buddy ")
		}
		return c, nil
	}
}
func AuthPoll(licenseKey string, authStatusCh chan <-AuthStatus, pctx context.Context) {
	defer sentry.Recover()

	userAgent := "RushAIO v" + os.Getenv("VERSION") + " (" + runtime.GOOS + ")"
	var authStatus AuthStatus
	var err error
	// var ce websocket.CloseError
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()



	client := &http.Client{}
	client.Transport = &http.Transport{
		DialTLS: makeDialer([]byte{0x42, 0x1d, 0x96, 0x88, 0xaf, 0x26, 0x20, 0x42, 0xe7, 0x34, 0x54, 0xcd, 0x72, 0x7, 0x23, 0xa7, 0xcc, 0x22, 0x19, 0x22, 0x4d, 0x2b, 0x83, 0x8e, 0x1b, 0xaf, 0x3a, 0xf4, 0xe1, 0x22, 0x80, 0xd0}, false),
	}

	var c *websocket.Conn
	opts := websocket.DialOptions {
		HTTPHeader: map[string][]string {
			"Cookie": []string {
				fmt.Sprintf("key=%s", licenseKey),
			},
			"User-Agent": []string { userAgent, },
			"Propane": []string{ "and propane accessories" },
		},
		HTTPClient: client,
	}

	// log.Printf("AuthPoll %+v", opts)
	cshost := cloudSocketHost()
	retry.Do(func() error {
		// log.Println("authpool_loop")
		defer time.Sleep(1*time.Second)
		c, _, err = websocket.Dial(ctx, cshost, &opts)
		if err != nil {
			authStatus.Error = err
			authStatusCh <- authStatus
			return err
		}

		for {
			err = wsjson.Read(ctx, c, &authStatus)
			if err != nil {
				authStatus.Error = err
				authStatusCh <- authStatus
				return err
			} else {
				authStatus.Error = nil
			}

			authStatusCh <- authStatus
		}
		c.Close(websocket.StatusNormalClosure, "")
		return err
	}, retry.Attempts(1e6), retry.MaxDelay(30 * time.Second), retry.RetryIf(func(err error) bool {
		// log.Println("authpoll_retryif")
		return (ctx.Err() == nil && !errors.Is(err, context.Canceled))
	}))
}

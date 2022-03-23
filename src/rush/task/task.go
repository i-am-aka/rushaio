package task

import (
	"context"
	"bufio"
	"github.com/avast/retry-go"
	"crypto/hmac"
	"crypto/tls"
	"crypto/sha256"
	"crypto/rsa"
	"crypto/x509"
	"sort"
	"golang.org/x/net/publicsuffix"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"rush/common"
	_ "github.com/mattn/go-sqlite3"
	stdhttp "net/http"
	"rush/net/http"
	"database/sql"
	"rush/net/http/cookiejar"
	// "rush/net/http/httptrace"

	"encoding/hex"
	"encoding/json"
	"compress/gzip"
	"strings"
	"crypto/aes"
	"crypto/cipher"
	"strconv"
	"bytes"
	"net/url"
	"runtime/debug"
	"os"
	"math/rand"
	crand "crypto/rand"
	"sync"
	"golang.org/x/net/http2"
	"golang.org/x/net/proxy"
	"time"
	"syreclabs.com/go/faker"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"regexp"
	"github.com/andybalholm/brotli"
	"reflect"
	utls "github.com/refraction-networking/utls"
)

var (
	dialTimeout   = time.Duration(30) * time.Second
	YS_BASE 			= "https://www.yeezysupply.com/"
	errOutOfStock = errors.New("Out of Stock")
	// YS_BASE = "https://localhost:5001/"
)

type Address struct {
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	City string `json:"city"`
	Zip string `json:"zipcode"`
	State string `json:"stateCode"`
	Country string `json:"country"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Phone string `json:"phoneNumber"`
}

func (a *Address) DicksPhone() string {
	re := regexp.MustCompile("[0-9]+")
	phoneNums := strings.Join(re.FindAllString(a.Phone, -1), "")
	if len(phoneNums) > 10 {
		phoneNums = phoneNums[len(phoneNums)-10:]
	}
	return fmt.Sprintf("(%c%c%c) %c%c%c-%c%c%c%c",
	phoneNums[0], phoneNums[1], phoneNums[2],
	phoneNums[3], phoneNums[4], phoneNums[5],
	phoneNums[6], phoneNums[7], phoneNums[8], phoneNums[9])
}
func (a *Address) HibPhone() string {
	re := regexp.MustCompile("[0-9]+")
	phoneNums := strings.Join(re.FindAllString(a.Phone, -1), "")
	if len(phoneNums) > 10 {
		phoneNums = phoneNums[len(phoneNums)-10:]
	}
	return fmt.Sprintf("%c%c%c-%c%c%c-%c%c%c%c",
	phoneNums[0], phoneNums[1], phoneNums[2],
	phoneNums[3], phoneNums[4], phoneNums[5],
	phoneNums[6], phoneNums[7], phoneNums[8], phoneNums[9])
}
func (a *Address) NordPhone() string {
	if len(a.Phone) < 10 {
		// todo error
		return ""
	}

	re := regexp.MustCompile("[0-9]+")
	phoneNums := strings.Join(re.FindAllString(a.Phone, -1), "")
	if len(phoneNums) > 10 {
		phoneNums = phoneNums[len(phoneNums)-10:]
	}
	return fmt.Sprintf("%c%c%c%c%c%c%c%c%c%c",
	phoneNums[0], phoneNums[1], phoneNums[2],
	phoneNums[3], phoneNums[4], phoneNums[5],
	phoneNums[6], phoneNums[7], phoneNums[8], phoneNums[9])
}
type Card struct {
	Type string `json:"cardType"`
	Number string `json:"number"`
	Cvc string `json:"cvc"`
	Name string `json:"holder"`
	ExpMonth int `json:"expirationMonth"`
	ExpYear int `json:"expirationYear"`
	PaymentMethodId string `json:"paymentMethodId"`
	LastFour string `json:"lastFour"`
}

func AesCrypt(plaintext string, key []byte, iv []byte, blockSize int) ([]byte, error) {
	bPlaintext := PKCS5Padding([]byte(plaintext), blockSize, len(plaintext))
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	ciphertext := make([]byte, len(bPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, bPlaintext)
	return append(iv, ciphertext...), nil
}

func PKCS5Padding(ciphertext []byte, blockSize int, after int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func EncRadial(number string, pubKeyB64 string) (string, error) {
	key := make([]byte, 32)
	iv := make([]byte, 16)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	_, err = rand.Read(iv)
	if err != nil {
		return "", err
	}
	enc, err := AesCrypt(number, key, iv, aes.BlockSize)
	if err != nil {
		return "", err
	}
	hmacKey := make([]byte, 32)
	_, err = rand.Read(hmacKey)
	if err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, hmacKey)
	h.Write(enc)
	hmacb64 := base64.StdEncoding.EncodeToString(h.Sum(nil))

	encb64 := base64.StdEncoding.EncodeToString(enc)

	keyConcat := append(key, hmacKey...)

	pkeybytes, err := base64.StdEncoding.DecodeString(pubKeyB64)
	if err != nil {
		return "", err
	}
	pkey, err := x509.ParsePKIXPublicKey(pkeybytes)
	if err != nil {
		return "", err
	}
	penc, err := rsa.EncryptPKCS1v15(crand.Reader, pkey.(*rsa.PublicKey), []byte(base64.StdEncoding.EncodeToString(keyConcat)))
	if err != nil {
		return "", err
	}
	penchex := hex.EncodeToString(penc)

	ciphed, err := RCipher(penchex)
	if err != nil {
		return "", err
	}
	encstr := fmt.Sprintf("$bt4|javascript_1_3_10$%s$%s$%s", ciphed, encb64, hmacb64)
	return encstr, nil
}
// func strwow(str []byte) string {
// 	et := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
// 	e := 0
// 	n := 0
// 	i := ""
// 	for e + 3 <= len(str) {
// 		n = strconv.Atoi("0x" + str[e:e+3])
// 		i += et[n >> 6] + et[63 & n]
// 		e = e + 3
// 	}
// 	for e + 1 {
// 		i += "="
// 	}
// 	return i
// }
func (c *Card) HibCardType() string {
	switch c.Type {
	case "AMEX":
		return "American Express"
	case "VISA":
		return "Visa"
	case "MASTER":
		return "Master Card"
	case "DISCOVER":
		return "Discover"
	}
	return ""
}
type Profile struct {
	Email string

	ShippingAddress Address

	BillingAddress Address

	Card Card

	Name string
}
type Status struct {
	Message string
	Level int
}

type CheckoutTask struct {
	Id string
	Version string
	Debug bool
	DebugCountBytes func(uint8, uint)

	WakeUp chan interface{}
	cloudCache *CloudCache
	dialTimeout time.Duration
	Db *sql.DB
	Url *url.URL
	BaseUrl *url.URL
	BaseDomainUrl *url.URL
	LastUrl *url.URL
	Keywords []string
	Proxy *url.URL
	ProxyGroup string
	ProxyStr string
	Profile Profile
	Sizes []string
	Config map[string]string
	ApiConfig TaskConfig
	StatusCh chan<- Status
	LastStatus Status
	UserAgent string
	Ctx context.Context
	Cancel context.CancelFunc
	ClientCtx context.Context
	ClientCtxCancel context.CancelFunc

	LastClientInitAt time.Time

	StartTime uint64
	StartCheckoutTime uint64
	StartPaymentTime uint64
	StartAtcTime uint64
	HeaderOrder []string
	HeaderBlacklist map[string]bool
	HeaderCaseFn func(string) string
	GetJwt func() string
	Webhook string
	UserId string

	Metrics TaskMetrics

	telemetryClient *http.Client
	StatusTelemetryEnabled bool
	TelemetryContextMut *sync.Mutex
	TelemetryContext map[string]interface{}
	// httpTrace *httptrace.ClientTrace
	EnableHttpTrace bool
	EnablePrivateHttpTrace bool

	DefaultReqClose bool
	DefaultProto string

	Delay time.Duration
	ApplyDelayInDoReq bool

	client *http.Client
	s3client *http.Client
	ApiClient *stdhttp.Client

	AbckJarMut sync.RWMutex
	AbckJar [][]*http.Cookie
	AbckJarUsage uint
	AbckJarMaxUsage uint
	AbckGenCookieBlocklist []string

	ReqRetries uint

	DnsCacheMut *sync.Mutex
	DnsCache map[string][]string
	DnsCacheActiveHost map[string]string
	DnsCacheLookupTime uint64

	UseIpv6 bool
	IpLookupTime uint64
	IpLookupResult map[string]string
	RealIp string

	LastRequestTime time.Time

	AddHit func(string, string)
	CheckHit func(string, string) uint32
	OneCheckoutPerProf bool

	LastSessionSnapshot *TaskSession

	ProxyTestResultsMut sync.Mutex
	ProxyTestResults map[string]bool
}

func (t *CheckoutTask) ProxySafeString() string {
	if t.Proxy == nil {
		return ""
	}

	proxySafe := *t.Proxy
	proxySafe.User = url.User(proxySafe.User.Username())
	return proxySafe.String()
}

var IpLookupUrl64, _ = url.Parse("https://api64.ipify.org/?format=json")
var IpLookupUrl, _ = url.Parse("https://ipinfo.io/json")
var ErrIpUnavailable = errors.New("IP Unavailable")

func (t *CheckoutTask) LookupProxyIp() (error) {
	t0 := time.Now()
	return t.Retry(func() error {
		client, err := t.newHttpClient()
		client.Timeout = 10*time.Second
		req := t.makeGetReq(IpLookupUrl64, nil)
		req = req.WithContext(t.Ctx)
		req.Close = true
		req.Proto = "HTTP/1.1"
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		ipm := map[string]string{}
		if err = readRespJsonDst(resp, &ipm); err == nil {
			if ip, ok := ipm["ip"]; ok {
				t.RealIp = ip
				// if strings.Contains(ip, "::") && t.EnableIpv6 {
				// 	t.UseIpv6 = true
				// }
				t.IpLookupTime = uint64(time.Now().Sub(t0) / time.Millisecond)
				t.IpLookupResult = ipm
				// go t.SendTelemetry(map[string]interface{} {
				// 	"event": "proxy_ip_lookup",
				// 	"proxy_ip_lookup": ipm,
				// })
				return nil
			}
		}
		return ErrIpUnavailable
	}, retry.Attempts(10), retry.DelayType(retry.DefaultDelayType))
}

// var NAMESERVERS = []string{"208.67.222.222", "1.1.1.1", "8.8.8.8", "64.6.64.6", "64.6.65.6", "9.9.9.9"}
var DohEndpoints = []string{"https://cloudflare-dns.com/dns-query","https://dns.google/resolve"}
func (t *CheckoutTask) UpdateDnsCache(host string) error {
	updCtx, updCancel := context.WithCancel(t.Ctx)
	go func() {
		defer func() {
			recover()
		}()
		defer updCancel()
		// t.DnsCacheActiveHost[host] = "23.192.110.132"
		// return nil

		t0 := time.Now()
		defer func() {
			t.DnsCacheLookupTime = uint64(time.Now().Sub(t0) / time.Millisecond)
		}()

		if os.Getenv("DNS_CACHE") == "0" {
			return
		}
		client, err := t.newHttpClient()
		if err != nil {
			return
		}
		// var client *http.Client
		// if t.Proxy != nil {
		// 	client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(t.Proxy)}}
		// } else {
		// 	client = &http.Client{}
		// }
		client.Timeout = 3*time.Second

		ips := []string{}
		dohEndpoint := DohEndpoints[rand.Intn(len(DohEndpoints))]
		var recordType string
		// if ip, err := t.LookupProxyIp(); err == nil {
		// 	t.TelemetryContextMut.Lock()
		// 	t.TelemetryContext["realip"] = ip
		// 	t.TelemetryContextMut.Unlock()
		// 	// t.TelemetryContext["ip_lookup_time"] = int64(t.IpLookupTime / time.Millisecond)
		// 	if strings.Contains(ip, ":") {
		// 		t.UseIpv6 = true
		// 	}
		// }
		if t.UseIpv6 {
			recordType = "AAAA"
		} else {
			recordType = "A"
		}
		// for _, dohEndpoint := range DohEndpoints {
			lurl, err := url.Parse(fmt.Sprintf("%s?name=%s&type=%s", dohEndpoint, host, recordType))
			if err != nil {
				return
			}
			req := t.makeGetReq(lurl, &[][2]string{ { "accept", "application/dns-json" } })
			req = req.WithContext(t.Ctx)
			req.Close = true
			// req.Proto = "HTTP/1.1"
			resp, err := client.Do(req)
			if err != nil {
			  return
			}
			var qResp DnsOverHttpResp
			err = readRespJsonDst(resp, &qResp)
			if err != nil {
				return
			}
			for _, record := range qResp.Answer {
				if record.Type == 1 || record.Type == 28 {
					ips = append(ips, record.Data)
				}
			}
			t.LogDebug("%s -> %+v\n", host, ips)
		// }
		if len(ips) > 0 {
			t.DnsCacheMut.Lock()
			defer t.DnsCacheMut.Unlock()
			t.DnsCache[host] = ips
			delete(t.DnsCacheActiveHost, host)
		}
	}()

	select {
	case <-updCtx.Done():
		break
	case <-time.After(4*time.Second):
		updCancel()
	}
	return nil
	// var resolver *net.Resolver
	// nameserver := NAMESERVERS[rand.Intn(len(NAMESERVERS))]
 //  resolver = &net.Resolver{
 //      PreferGo: true,
 //      Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
 //        d := net.Dialer{}
 //        return d.DialContext(ctx, "udp", net.JoinHostPort(nameserver, "53"))
 //      },
 //  }

	// records, err := resolver.LookupIPAddr(t.Ctx, t.Url.Host)
	// if err != nil {
	// 	return err
	// }
	// rand.Shuffle(len(records), func(i, j int) {
	//   records[i], records[j] = records[j], records[i] })
	// ips := []string{}
	// for _, record := range records {
	// 	ip := record.IP.To4()
	// 	if ip != nil {
	// 		ips = append(ips, ip.String())
	// 	}
	// }
	// log.Printf("%+v", ips)

}

// type DnsResolver struct {
// 	*net.Resolver
// }

func (t *CheckoutTask) dialDnsOverHttps() func (ctx context.Context, network string, addr string) (net.Conn, error) {
	return func (ctx context.Context, network string, addr string) (net.Conn, error) {
		if t.Debug {
			log.Printf("dialDnsOverHttps %+v %+v", network, addr)
			log.Printf("%+v", t.DnsCache)
		}
		var d = net.Dialer{Timeout: t.dialTimeout, KeepAlive: -1*time.Second}

		addrhostorig, addrport, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		addrport_res, _ := net.LookupPort(network, addrport)
		addrhost := addrhostorig

		t.DnsCacheMut.Lock()
		if ips, ok := t.DnsCache[addrhost]; ok {
			host := ips[rand.Intn(len(ips))]
			t.DnsCacheActiveHost[addrhost] = host
			t.TelemetryContextMut.Lock()
			t.TelemetryContext["dns"] = t.DnsCacheActiveHost
			t.TelemetryContextMut.Unlock()
			addrhost = host
			t.DnsCacheMut.Unlock()
		} else {
			t.DnsCacheMut.Unlock()
			addr_res, err := net.LookupHost(addrhost)
			if err != nil {
				return nil, err
			}
			for _, addrstr := range addr_res {
				addr, err := net.ResolveTCPAddr(network, addrstr)
				if err != nil {
					continue
				}
				if t.UseIpv6 && addr.IP.To4() == nil {
					addrhost = addrstr
				} else if !t.UseIpv6 && addr.IP.To4() != nil {
					addrhost = addrstr
				}
			}
		}

		if network == "tcp6" {
			addrhost = "[" + addrhost + "]"
		}
		if t.Debug {
			log.Printf("dial %s %s", network, addrhost + ":" + strconv.Itoa(addrport_res))
		}
		return d.DialContext(ctx, network, addrhost + ":" + strconv.Itoa(addrport_res))
	}
}

var ErrInvalidProxyFormat = errors.New("Invalid proxy format")

func (t *CheckoutTask) dial(proxy *url.URL) func (ctx context.Context, network string, addr string) (net.Conn, error) {
	return func (ctx context.Context, network string, addr string) (net.Conn, error) {
		conn, err := func() (net.Conn, error) {
			// debug.PrintStack()
			t.LogDebug("dial %+s proxy=%s\n", addr, proxy.Host)
			proxyhost, proxyport, err := net.SplitHostPort(proxy.Host)
			t.LogDebug("%s %s %s\n", proxyhost, proxyport, "")
			proxy_res, err := net.LookupHost(proxyhost)
			if len(proxy_res) == 0 {
				return nil, ErrInvalidProxyFormat
			}
			if err != nil {
				return nil, err
			}
			proxyaddr := proxy_res[len(proxy_res)-1]
			if strings.Contains(proxyhost, "::") {
				proxyaddr = "[" + proxyaddr + "]"
			}
			d := net.Dialer{
				Timeout: t.dialTimeout,
				KeepAlive: -1 * time.Second,
			}
			conn, err := d.DialContext(ctx, network, proxyaddr + ":" + proxyport)
			if err != nil {
				t.Metrics.Incr("proxy_conn_fail")
				if t.Metrics.GetUint64("proxy_conn_fail") > 3 {
					t.RotateProxy()
				}
				if t.shouldRetry(err) &&  ctx.Err() == nil {
	    		go func() {
	    			defer func() {
	    				recover()
	    			}()
	    			// t.StatusCh <- StatusProxyConnectionFailed
	    		}()
	    	}
				t.LogDebug("DialContext failed %+v", err)
				return nil, err
			}

			addrhostorig, addrport, err := net.SplitHostPort(addr)
			// addr_res, _ := net.LookupHost(addrhost)
			addrport_res, _ := net.LookupPort(network, addrport)
			// TODO dns through proxy to get optimal edge servers

			// var resolver *net.Resolver
			// nameserver := NAMESERVERS[rand.Intn(len(NAMESERVERS))]
			// if nameserver != "" {
			//     resolver = &net.Resolver{
			//         PreferGo: true,
			//         Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			//             d := net.Dialer{}
			//             return d.DialContext(ctx, "udp", net.JoinHostPort(nameserver, "53"))
			//         },
			//     }
			// } else {
			//     resolver = net.DefaultResolver
			// }

			// records, _ := resolver.LookupIPAddr(ctx, addrhost)
			// rand.Shuffle(len(records), func(i, j int) {
			//   records[i], records[j] = records[j], records[i] })
			// for _, record := range records {
			// 	ip := record.IP.To4()
			// 	if ip != nil {
			// 		addrhost = ip.String()
			// 		break
			// 	}

			// }

			addrhost := addrhostorig
			t.DnsCacheMut.Lock()
			// if ip, ok := t.DnsCacheActiveHost[addrhost]; ok {
			// 	addrhost = ip
			// } else
			if ips, ok := t.DnsCache[addrhost]; ok {
				host := ips[rand.Intn(len(ips))]
				t.DnsCacheActiveHost[addrhost] = host
				t.TelemetryContextMut.Lock()
				t.TelemetryContext["dns"] = t.DnsCacheActiveHost
				t.TelemetryContextMut.Unlock()
				addrhost = host
			} else {
				// t.DnsCacheMut.Unlock()
				// t.UpdateDnsCache(addrhost)
				// t.DnsCacheMut.Lock()
				// if ips, ok := t.DnsCache[addrhost]; ok {
				// 	addrhost = ips[rand.Intn(len(ips))]
				// }
			}
			t.DnsCacheMut.Unlock()
			if os.Getenv("DEBUG") != "" {
				log.Println(addrhost + " port " + strconv.Itoa(addrport_res))
			}
			if strings.Contains(addrhost, "::") {
				addrhost = "[" + addrhost + "]"
			}
			addr_connstr := addrhost + ":" + strconv.Itoa(addrport_res)

			// log.Println(addr_connstr)
			// log.Println(addrhostorig + ":" + strconv.Itoa(addrport_res))

			var connStr = ("CONNECT " + addr_connstr +  " HTTP/1.1\r\n" +
				// "Host: " + addrhost + ":" + strconv.Itoa(addrport_res) + "\r\n" +
				"Host: " + addrhostorig + ":" + strconv.Itoa(addrport_res) + "\r\n" +
	      "Proxy-Connection: keep-alive\r\n")
	      t.LogDebug(connStr)
			if proxy.User.String() != "" {
				userDec, err := url.PathUnescape(proxy.User.String())
				if err != nil {
					userDec = proxy.User.String()
				}
				credentials := base64.StdEncoding.EncodeToString([]byte(userDec))
				connStr += "Proxy-Authorization: Basic " + credentials + "\r\n"
			}
			connStr += "\r\n"
			// log.Println(connStr)
			_, err = conn.Write([]byte(connStr))
			if err != nil {
				t.Metrics.Incr("proxy_conn_fail")
				if t.Metrics.GetUint64("proxy_conn_fail") > 3 {
					if t.shouldRetry(err) &&  t.Ctx.Err() == nil {
		    		go func() {
		    			defer func() {
		    				recover()
		    			}()
		    			// t.StatusCh <- StatusProxyConnectionFailed
		    		}()
		    	}
					t.RotateProxy()
				}
				return nil, err
			}
			br := make([]byte, 1024)
			_, err = conn.Read(br)
			if err != nil {
				t.Metrics.Incr("proxy_conn_fail")
				if t.Metrics.GetUint64("proxy_conn_fail") > 3 {
					t.RotateProxy()
				}
				if t.shouldRetry(err) &&  t.Ctx.Err() == nil {
	    		go func() {
	    			defer func() {
	    				recover()
	    			}()
	    			// t.StatusCh <- StatusProxyConnectionFailed
	    		}()
	    		CloseH2Conns(t.client)
	    		t.client.CloseIdleConnections()
	    	}
				return nil, err
			}
			resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(br)), nil)
	    // log.Printf("%+v %+v", resp, err)
	    if err != nil {
	    	t.Metrics.Incr("proxy_conn_fail")
	    	if t.Metrics.GetUint64("proxy_conn_fail") > 3 {
	    		t.RotateProxy()
		    	if t.shouldRetry(err) && t.Ctx.Err() == nil {
		    		go func() {
		    			defer func() {
		    				recover()
		    			}()
		    			// t.StatusCh <- StatusProxyConnectionFailed
		    		}()
		    	}
	    	}
	      return nil, err
	    }
	    // log.Println(string(br))
	    if resp.StatusCode < 200 || resp.StatusCode > 299 {
	    	// disable dns cache on banned pages
	    	if !t.UseIpv6 {
		    	t.DnsCacheMut.Lock()
		    	t.DnsCacheActiveHost = map[string]string{}
		    	t.DnsCache = nil
		    	t.DnsCacheMut.Unlock()
		    }
	    	// t.CloseH2Conns()
	    	// t.client.CloseIdleConnections()
	    	// t.DnsCache = map[string][]string{}
		    // ss := fmt.Sprintf("Proxy Connection Failed (%s) (%s)", resp.Status, connStr)
		    CloseH2Conns(t.client)
		    t.client.CloseIdleConnections()

		    if t.Metrics.GetUint64("proxy_conn_fail") > 3 {
		    	t.RotateProxy()
			    if t.Ctx.Err() == nil {
			    	go func() {
			    		defer func() {
			    			recover()
			    		}()
			    		// t.StatusCh <- Status{Message: ss}
			    	}()
			    }
		    }
	    	// if resp.StatusCode == 407 {
		    // }

		    err = ErrProxyConnectionFailed
	    	return nil, err
	    }
	    if err == nil {
	    	t.Metrics.Set("proxy_conn_fail", 0)
	    }
			return conn, err
		}()
		return conn, err
	}
}

func keyLog(fn string) (io.Writer, error)  {
	return os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
}

func (t *CheckoutTask) addHeader(headers [][2]string, key string, value string) ([][2]string) {
	return append(headers, [2]string{ t.HeaderCaseFn(key), value, })
}

func addHeader(headers [][2]string, key string, value string) ([][2]string) {
	return append(headers, [2]string{ key, value, })
}

func getHeader(headers *[][2]string, key string) (string) {
	if headers == nil {
		return ""
	}

	for _, h := range *headers {
		if strings.ToLower(h[0]) == strings.ToLower(key) {
			return h[1]
		}
	}
	return ""
}

func getHeaderIdx(headers *[][2]string, key string) (int) {
	if headers == nil {
		return -1
	}

	for idx, h := range *headers {
		if strings.ToLower(h[0]) == strings.ToLower(key) {
			return idx
		}
	}
	return -1
}



var (
	// TODO per browser fp
	DEFAULT_HEADER_PRI = []string{
		"upgrade-insecure-requests",
		"content-length",
		"user-agent",
		"x-requested-with",
		"x-auth-cookie-required",
		"content-type",
		"checkout-authorization",
		"accept",
		"origin",
		"sec-fetch-site",
		"sec-fetch-mode",
		"sec-fetch-dest",
		"sec-fetch-user",
		"sec-gpc",
		"referer",
		"accept-encoding",
		"accept-language",
		"x-instana-t",
		"x-instana-s",
		"x-instana-l",
	}
)

func strIndexOf(element string, data []string) (int) {
   for k, v := range data {
       if strings.ToLower(element) == strings.ToLower(v) {
           return k
       }
   }
   return -1    //not found.
}

func (t *CheckoutTask) makeReq(method string, url_ *url.URL, _headers *[][2]string, headerOrder *[]string, _body *[]byte) (*http.Request) {
	headers := make([][2]string, 0)


	if getHeader(_headers, "sec-fetch-dest") == "" && !t.HeaderBlacklist["sec-fetch-dest"] && !strings.Contains(t.UserAgent, "Firefox") {
		secdest := "empty"
		if method == "GET" && getHeader(_headers, "Content-Type") == "" {
			secdest = "document"
		} // TODO more secdests?
		headers = t.addHeader(headers, "sec-fetch-dest", secdest)
	}
	if getHeader(_headers, "sec-fetch-site") == "" && !t.HeaderBlacklist["sec-fetch-site"] && !strings.Contains(t.UserAgent, "Firefox") {
		headers = t.addHeader(headers, "sec-fetch-site", "same-origin")
	}
	// if getHeader(_headers, "sec-gpc") == "" && !t.HeaderBlacklist["sec-gpc"] && strings.Contains(t.UserAgent, "Chrome/87") {
	// 	headers = t.addHeader(headers, "sec-gpc", "1")
	// }
	custUa := getHeader(_headers, "user-agent")
	if custUa == "" {
		headers = t.addHeader(headers, "user-agent", t.UserAgent)
	}
	if getHeader(_headers, "accept") == "" && !t.HeaderBlacklist["accept"]{
		headers = t.addHeader(headers, "Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	}
	if getHeader(_headers, "accept-encoding") == "" && !t.HeaderBlacklist["accept-encoding"] {
		headers = t.addHeader(headers, "Accept-Encoding", "gzip, deflate, br")
	}
	if getHeader(_headers, "accept-language") == "" && !t.HeaderBlacklist["accept-language"] {
		headers = t.addHeader(headers, "Accept-Language", "en-US,en;q=0.9")
	}
	if _headers != nil {
		for _,k := range *_headers {
			headers = t.addHeader(headers, k[0], k[1])
		}
	}


	var body io.ReadCloser
	// var err error
	var contentLength int64 = 0
	if _body != nil {
		if !t.HeaderBlacklist["content-length"] {
			contentLength = int64(len(*_body))
			headers = t.addHeader(headers, "content-length", strconv.Itoa(int(contentLength)))
		}
		body = ioutil.NopCloser(bytes.NewReader(*_body))
	}
	// log.Printf("%+v", headers)
	if headerOrder == nil {
		headerOrder = &t.HeaderOrder
	}
	sort.SliceStable(headers, func(i, j int) bool {
    return strIndexOf(headers[i][0], *headerOrder) < strIndexOf(headers[j][0], *headerOrder)
	})

	// rand.Shuffle(len(headers), func(i, j int) {
	// 		  headers[i], headers[j] = headers[j], headers[i] })

	host := strings.Split(url_.Host, ":")[0]

	req := &http.Request{
		Method: method,
		URL:    url_,
		Header: make(http.Header),
		RawHeader: headers,
		Host:   host,
		Body:   body,
		ContentLength: contentLength,
		Close: t.DefaultReqClose,
		Proto:  t.DefaultProto,
		// ProtoMajor: 2,
		// ProtoMinor: 0,
	}
	req = req.WithContext(t.ClientCtx)
	return req
	// if t.EnableHttpTrace {
	// 	return req.WithContext(httptrace.WithClientTrace(req.Context(), t.httpTrace))
	// } else {
	// }
}

func (t *CheckoutTask) makeGetReq(url *url.URL, headers *[][2]string) (*http.Request) {
	return t.makeReq("GET", url, headers, nil, nil)
}

func (t *CheckoutTask) GetCookieValueClient(name string, client *http.Client) string {
	if base := getCookieValue(name, t.BaseUrl, client.Jar); base != "" {
		return base
	} else {
		return getCookieValue(name, t.BaseDomainUrl, client.Jar)
	}
}

func (t *CheckoutTask) GetCookieValue(name string) string {
	return t.GetCookieValueClient(name, t.client)
}

func getCookie(name string, url *url.URL, jar http.CookieJar) *http.Cookie {
	for _, c := range jar.Cookies(url) {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func getCookieValue(name string, url *url.URL, jar http.CookieJar) string {
	for _, c := range jar.Cookies(url) {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

func dumpReq(req *http.Request) {
	// TODO do this without consuming request body
	log.Printf("%s %s", req.Method, req.URL.String())
	for _,k := range req.RawHeader {
		log.Printf("%s=%s", k[0], k[1])
		// headers = addHeader(headers, k[0], k[1])
	}
	if req.Body != nil {
			body, _ := ioutil.ReadAll(req.Body)
			req.Body.Close()
			log.Println(string(body[:]))
			req.Body = ioutil.NopCloser(bytes.NewReader(body))
	}
	log.Printf("%+v", req.Cookies())
	log.Println()
}

func dumpHeaders(header http.Header) {
	headersDbg := new(bytes.Buffer)
	header.Write(headersDbg)
	log.Println()
	log.Println(string(headersDbg.Bytes())[:])
	log.Println()
}


func captchaApiHost() string {
	apiHost := ""//os.Getenv("CAP_API_HOST")
	if apiHost == "" {
		apiHost = "https://api.rushaio.co/cap"
	}
	return apiHost
}

func apiHost() string {
	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = "https://api.rushaio.co"
	}
	return apiHost
}

func (t *CheckoutTask) getAdyenFp(retries int) (string, error) {
	if retries <= 0 {
		retries = 1e3
	}
	client := &http.Client{Timeout:30*time.Second}
	var body string
	err := t.Retry(
		func() error {
			req, _ := http.NewRequestWithContext(t.ClientCtx, "GET", apiHost() + "/ay", nil)
			req.Header = map[string][]string {
				"authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
			}
			resp, err := t.doReq(client, req)
			if err != nil {
				return err
			} else {
				body, err = readBody(resp)
				return err
			}
		},
		retry.Attempts(uint(retries)),
	)
	if err != nil {
		return "", err
	}
	return body, nil
}

func (t *CheckoutTask) GetJson(url string, dest interface{}) (error) {
	client := &http.Client{Timeout:30*time.Second}
	req, err := http.NewRequestWithContext(t.ClientCtx, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := t.doReq(client, req)
	if err != nil {
		DiscardResp(resp)
		return err
	}
	err = readRespJsonDst(resp, &dest)
	return err
}

func (t *CheckoutTask) getUserAgent() (string, error) {
	if os.Getenv("USER_AGENT") != "" {
		return os.Getenv("USER_AGENT"), nil
	}

	client := &http.Client{Timeout:30*time.Second}
	var userAgent string
	err := t.Retry(
		func() error {
			ua := apiHost() + "/ua"
			if t.Url.Host == "www.yeezysupply.com" {
				ua = apiHost() + "/uanew"
			}
			req, _ := http.NewRequestWithContext(t.ClientCtx, "GET", ua, nil)
			req.Header = map[string][]string {
				"authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
			}
			resp, err := t.doReq(client, req)
			if err != nil {
				return err
			} else {
				userAgent, err = readBody(resp)
				return err
			}
		},
		retry.Attempts(1e3),
	)
	// log.Printf("ua %s", userAgent)
	return userAgent, err
}

func (t *CheckoutTask) getSensorData(url string, abck string) ([]byte, error) {
	client := &http.Client{Timeout:30*time.Second}
	body := make(map[string]interface{})
	body["url"] = url
	if t.Url.Host == "www.yeezysupply.com" {
		if genurl, ok := t.Config["genurl"]; ok && len(genurl) > 0{
			body["url"] = genurl
		}
	}
	if abck != "" {
		body["_abck"] = abck
	} else {
		body["_abck"] = "null"
	}
	if t.UserAgent != "" {
		body["ua"] = t.UserAgent
	}

	body["stage"] = t.Config["stage"]
	if body["stage"] == "" {
		body["stage"] = "click"
	}
	if aji, err := strconv.Atoi(t.Config["aj_indx"]); err == nil {
		body["aj_indx"] = aji
	} else {
		body["aj_indx"] = 2+rand.Intn(10)
	}

	var sensorData []byte
	err := t.Retry(
		func() error {
			bodyJson, _ := json.Marshal(body)
			bodyBuf := ioutil.NopCloser(bytes.NewReader(bodyJson))
			ah := apiHost() + "/"
			if t.Url.Host == "www.yeezysupply.com" {
				ah = apiHost() + "/aknew"
			}
			req, _ := http.NewRequestWithContext(t.ClientCtx, "POST", ah, bodyBuf)
			req.Header = map[string][]string {
				"Content-Type": {"application/json",},
				"authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
				"x-version": {t.Version,},
				"wd40": {"1",},
			}
			resp, err := t.doReq(client, req)
			if err != nil {
				return err
			} else if resp.StatusCode == 200 {
				sensorData, err = readBodyBytes(resp)
				return err
			} else {
				return ErrRetrying
			}
		},
		retry.Attempts(1e3),
	)
	if err != nil {

		return []byte{}, err
	}
	// log.Println(string(sensorData))
	return sensorData, nil
}

var ErrNilResponseBody = errors.New("Nil response body")

func readBodyBytesStd(resp *stdhttp.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return []byte{}, ErrNilResponseBody
	}

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
		defer DiscardRespStd(resp)
		bb, err := ioutil.ReadAll(reader)
		if os.Getenv("DEBUG") == "1" && os.Getenv("RESPDUMP") != "0" {
			log.Printf("RESP: %s", string(bb))
		}
		return bb, err
	} else {
		return []byte{}, errors.New("Error reading body")
	}
}
func readBodyBytes(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return []byte{}, ErrNilResponseBody
	}

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
		defer DiscardResp(resp)
		bb, err := ioutil.ReadAll(reader)
		if os.Getenv("DEBUG") == "1" && os.Getenv("RESPDUMP") != "0" {
			log.Printf("RESP: %s", string(bb))
		}
		return bb, err
	} else {
		return []byte{}, errors.New("Error reading body")
	}
}

var (
	errEmptyBody = errors.New("errEmptyBody")
)

func readBody(resp *http.Response) (string, error) {
	if resp == nil {
		return "", errEmptyBody
	}
	body, err := readBodyBytes(resp)
	if err != nil {
		return "", err
	}
	return string(body[:]), nil
}

type Sku struct {
	product string
	// productVariation string
	size string
}

type ShippingLineItem struct {
	Id string `json:"id"`
}

type Shipment struct {
	ShipmentId string `json:"shipmentId"`
	ShippingLineItem ShippingLineItem `json:"shippingLineItem"`
}

type Cart struct {
	Message string `json:"message"`
	ModifiedDate string `json:"modifiedDate"`
	TotalProductCount int `json:"totalProductCount"`
	BasketId string `json:"basketId"`
	MessageList []MessageList `json:"messageList"`
	ShipmentList []Shipment `json:"shipmentList"`
}

type MessageList struct {
	Type string `json:"type"`
}

type CartAdd struct {
	DisplaySize string `json:"displaySize"`
	ProductId string `json:"productId"`
	Product_id string `json:"product_id"`
	Product_variation_sku string `json:"product_variation_sku"`
	Quantity int `json:"quantity"`
	Size string `json:"size"`
}

type Customer struct {
	Email string `json:"email"`
	ReceiveSmsUpdates bool `json:"receiveSmsUpdates"`
}
type ShippingMethod struct {
	Id string `json:"id"`
	ShipmentId string `json:"shipmentId"`
	CollectionPeriod string `json:"collectionPeriod"`
	DeliveryPeriod string `json:"deliveryPeriod"`
}

type CheckoutInit struct {
	Customer Customer `json:"customer"`
	ShippingAddress Address `json:"shippingAddress"`
	BillingAddress Address `json:"billingAddress"`
	MethodList []ShippingMethod `json:"methodList"`
	NewsletterSubscription bool `json:"newsletterSubscription"`
}

type PaymentInstrument struct {
	Holder string `json:"holder"`
	ExpMonth int `json:"expirationMonth"`
	ExpYear int `json:"expirationYear"`
	LastFour string `json:"lastFour"`
	PaymentMethodId string `json:"paymentMethodId"`
	CardType string `json:"cardType"`
}

type Payment struct {
	BasketId string `json:"basketId"`
	EncryptedInstrument string `json:"encryptedInstrument"`
	Fingerprint string `json:"fingerprint"`
	PaymentInstrument PaymentInstrument `json:"paymentInstrument"`
}

type StatusCodeRejectedError struct {
	StatusCode int
	Err error
}
func (e *StatusCodeRejectedError) Error() string { return strconv.Itoa(e.StatusCode) + ": " + e.Err.Error() }

func (t *CheckoutTask) doReqTC(req *http.Request) (*http.Response, error) {
	return t.doReq(t.client, req)
}

type SerializableHttpRequest struct {
  Method string `json:"method"`
  URL string `json:"urlstr"`
  Proto string `json:"proto"`
  RawHeader [][2]string `json:"headers"`
  ContentLength int64
  TransferEncoding []string
  Close bool
}

func SerializableRequestFromRequest(r *http.Request) *SerializableHttpRequest {
	return &SerializableHttpRequest{
		Method: r.Method,
		URL: r.URL.String(),
		Proto: r.Proto,
		RawHeader: r.RawHeader,
		ContentLength: r.ContentLength,
		Close: r.Close,
	}
}

type SerializableHttpResponse struct {
	StatusCode int
	Proto      string
	Header http.Header
	ContentLength int64
	TransferEncoding []string
	Trailer http.Header
}

func SerializableResponseFromResponse(r *http.Response) SerializableHttpResponse {
	if r == nil {
		return SerializableHttpResponse{}
	}
	return SerializableHttpResponse{
		StatusCode: r.StatusCode,
		Proto: r.Proto,
		Header: r.Header,
		ContentLength: r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Trailer: r.Trailer,
	}
}

func GetAllClientCookies(client *http.Client) *[]cookiejar.Entry {
	if jar, ok := client.Jar.(*cookiejar.Jar); ok {
		entries := jar.GetEntries()
		return &entries
	} else {
		return nil
	}
}

func GetCookieEntry(name string, client *http.Client) *cookiejar.Entry {
	if jar, ok := client.Jar.(*cookiejar.Jar); ok {
		entries := jar.GetEntries()
		for _, ent := range entries {
			if ent.Name == name {
				return &ent
			}
		}
	}
	return nil
}

func (t *CheckoutTask) doReq2(client *http.Client, req *http.Request) (*http.Response, error) {
	// TODO apply timeout ctx here

	if os.Getenv("DEBUG") != "" {
		dumpReq(req)
		// dumpCookies(client.Jar)
	}

	var resp *http.Response
	body := []byte{}
	if req.Body != nil {
		body, _ = ioutil.ReadAll(req.Body)
		// todo is there an error here? some reqs retry with a panic - something with the body jig may have (auto-recoverable) bug
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	var origckie *[]cookiejar.Entry
	if t.EnableHttpTrace || os.Getenv("LOG") == "1" {
	  origckie = GetAllClientCookies(client)
		bstr := string(body)
		if t.EnablePrivateHttpTrace {
			bstr = ""
		}
		go t.SendTelemetry(map[string]interface{} {
			"http_trace_req": struct{
				Jar *[]cookiejar.Entry `json:"jar"`
				Body string
				*SerializableHttpRequest
			}{
				origckie, bstr, SerializableRequestFromRequest(req),
			},
		})
	}
	if t.ApplyDelayInDoReq {
		if time.Now().Sub(t.LastRequestTime) < t.Delay {
			time.Sleep(t.Delay - time.Now().Sub(t.LastRequestTime))
		}
		t.LastRequestTime = time.Now()
	}
	err := t.Retry(
		func() error {

			t.LogDebug("doReq")
			// clone original request to avoid post-execution mutations from sticking (dupe cookies)
			reqC := req.Clone(t.ClientCtx)
			if req.Body != nil {
        reqC.Body = ioutil.NopCloser(bytes.NewReader(body))
			}
			rresp, err := client.Do(reqC)
			if errors.Is(err, errProtocolNegotiated) {
				reqC = req.Clone(req.Context())
				if req.Body != nil {
					reqC.Body = ioutil.NopCloser(bytes.NewReader(body))
				}
				rresp, err = client.Do(reqC)
				// return errors.New("Renegotiating Connection")
			}

			if rresp != nil && req.IgnoreBody && err != nil && errors.Is(err, http2.ErrClosedResponseBody) {
				// this is ok, we throw away body
				err = nil
			}

			resperr := err

			if t.EnableHttpTrace {
				var rr *http.Response
				var bodystr string
				if rresp != nil && rresp.Body != nil {
					if bodyraw, err := ioutil.ReadAll(rresp.Body); err == nil {
						rr = &http.Response{Body: ioutil.NopCloser(bytes.NewReader(bodyraw)), Header: rresp.Header}
						if body, err := readBodyBytes(rr); err == nil {
							bodystr = string(body)
						}
						rresp.Body = ioutil.NopCloser(bytes.NewReader(bodyraw))
					}
				}
				var header http.Header
				var statusCode int
				if rresp != nil {
					header = rresp.Header
					statusCode = rresp.StatusCode
					t.LogDebug("RESP STATUS %+v \n\nHEADER %+v \n\nBODYSTR%s\n\n", statusCode, rresp.Header, bodystr)
				}
				go t.SendTelemetry(map[string]interface{} {
					"http_trace_resp": struct {
						Url string `json:"url"`
						Body string `json:"body"`
						Header http.Header `json:"headers"`
						StatusCode int `json:"status_code"`
						Jar *[]cookiejar.Entry `json:"jar"`
						Error string `json:"errorstr"`
						SerializableHttpResponse
					}{
						req.URL.String(), bodystr, header, statusCode, GetAllClientCookies(client), fmt.Sprintf("%+v", resperr), SerializableResponseFromResponse(rresp),
					},
				})
			}

			resp = rresp
			t.LogDebug("%+v %+v", resp, resperr)
			if resperr != nil {
				if t.Debug {
					log.Printf("Request Error: %+v", errors.WithStack(err))
				}
				// t.StatusCh <- StatusRequestError

				return resperr
			}
			t.LastUrl = req.URL
			return nil
		},
		retry.Attempts(t.ReqRetries), retry.LastErrorOnly(true), retry.Delay(1*time.Second),
	)
	return resp, err
}

func (t *CheckoutTask) doReq(client *http.Client, req *http.Request) (*http.Response, error) {
	if os.Getenv("DEBUG") != "" {
		dumpReq(req)
		// dumpCookies(client.Jar)
	}

	var resp *http.Response
	body := []byte{}
	if req.Body != nil {
		body, _ = ioutil.ReadAll(req.Body)
		// todo is there an error here? some reqs retry with a panic - something with the body jig may have (auto-recoverable) bug
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	var origckie *[]cookiejar.Entry
	if t.EnableHttpTrace || os.Getenv("LOG") == "1" {
	  origckie = GetAllClientCookies(client)
		bstr := string(body)
		if t.EnablePrivateHttpTrace {
			bstr = ""
		}
		go t.SendTelemetry(map[string]interface{} {
			"http_trace_req": struct{
				Jar *[]cookiejar.Entry `json:"jar"`
				Body string
				*SerializableHttpRequest
			}{
				origckie, bstr, SerializableRequestFromRequest(req),
			},
		})
	}
	if t.ApplyDelayInDoReq {
		if time.Now().Sub(t.LastRequestTime) < t.Delay {
			time.Sleep(t.Delay - time.Now().Sub(t.LastRequestTime))
		}
		t.LastRequestTime = time.Now()
	}
	ctx, _ := context.WithTimeout(t.ClientCtx, t.dialTimeout)
	err := t.Retry(
		func() error {
			t.LogDebug("doReq")
			// clone original request to avoid post-execution mutations from sticking (dupe cookies)
			reqC := req.Clone(ctx)
			if req.Body != nil {
        reqC.Body = ioutil.NopCloser(bytes.NewReader(body))
			}
			rresp, err := client.Do(reqC)
			// log.Printf("reqC resp %+v  err %+v", rresp, err)
			if errors.Is(err, errProtocolNegotiated) {
				reqC = req.Clone(ctx)
				if req.Body != nil {
					reqC.Body = ioutil.NopCloser(bytes.NewReader(body))
				}
				rresp, err = client.Do(reqC)
				// return errors.New("Renegotiating Connection")
			}

			if rresp != nil && req.IgnoreBody && err != nil && (errors.Is(err, http2.ErrClosedResponseBody) || strings.HasSuffix(fmt.Sprintf("%s", err), "EOF")) {
				// this is ok, we throw away body
				err = nil
			} else if err == nil && rresp == nil {
				t.LogDebug("BAD RESPONSE, NIL ERR AND RESP")
				t.Metrics.Incr("req_bad_response")
				CloseH2Conns(client)
				client.CloseIdleConnections()
				return ErrRetrying
			}

			resperr := err
			if err != nil {
				t.LogDebug("UNWRAP %v unw=%v orig=%v", errors.Unwrap(err), reflect.TypeOf(errors.Unwrap(err)), reflect.TypeOf(err))
			}

			if false && t.EnableHttpTrace {
				var rr *http.Response
				var bodystr string
				if rresp != nil && rresp.Body != nil {
					if bodyraw, err := ioutil.ReadAll(rresp.Body); err == nil {
						rr = &http.Response{Body: ioutil.NopCloser(bytes.NewReader(bodyraw)), Header: rresp.Header}
						if body, err := readBodyBytes(rr); err == nil {
							bodystr = string(body)
						}
						rresp.Body = ioutil.NopCloser(bytes.NewReader(bodyraw))
					}
				}
				var header http.Header
				var statusCode int
				if rresp != nil {
					header = rresp.Header
					statusCode = rresp.StatusCode
					t.LogDebug("RESP STATUS %+v \n\nHEADER %+v \n\nBODYSTR%d\n\n", statusCode, rresp.Header, len(bodystr))
				}
				go t.SendTelemetry(map[string]interface{} {
					"http_trace_resp": struct {
						Url string `json:"url"`
						Body string `json:"body"`
						Header http.Header `json:"headers"`
						StatusCode int `json:"status_code"`
						Jar *[]cookiejar.Entry `json:"jar"`
						Error string `json:"errorstr"`
						SerializableHttpResponse
					}{
						req.URL.String(), bodystr, header, statusCode, GetAllClientCookies(client), fmt.Sprintf("%+v", resperr), SerializableResponseFromResponse(rresp),
					},
				})
			}

			resp = rresp
			t.LogDebug("%s %s %s %+v %+v", req.Method, req.URL.String(), t.UserAgent, resp, resperr)
			if resperr != nil {
				if t.Debug {
					log.Printf("Request Error: %+v", errors.WithStack(err))
				}
				if strings.Contains(resperr.Error(), "stream error") {
					DiscardResp(resp)
					CloseH2Conns(client)
					t.client.CloseIdleConnections()
				}
				// t.StatusCh <- StatusRequestError

				return resperr
			} else {
				t.LastUrl = req.URL
				return nil
			}
			// else if rresp.StatusCode >= 500 && rresp.StatusCode < 600 && rresp.StatusCode != 531 && rresp.StatusCode != 550 {
			// 	 return retry.Unrecoverable(&StatusCodeRejectedError{Err:errors.New(fmt.Sprintf("invalid status code: %d", rresp.StatusCode)),StatusCode:rresp.StatusCode,})
			// } else if rresp.StatusCode > 302 && rresp.StatusCode != 499 {
			// 	return retry.Unrecoverable(&StatusCodeRejectedError{Err:errors.New(fmt.Sprintf("invalid status code: %d", rresp.StatusCode)),StatusCode:rresp.StatusCode,})
		},
		retry.Attempts(t.ReqRetries), retry.LastErrorOnly(true), retry.Delay(1*time.Second),
	)
	return resp, err
}

type TaskConfig struct {
	SensorEndpoints map[string]string `json:"sensorEndpoints"`
	RcApiKey string `json:"rcApiKey"`
	RcAction string `json:"rcAction"`
	RcCookie string `json:"rcCookie"`
	SizeLookup map[string]string `json:"sizeLookup"`
	SlimEndpoint string `json:"slimEndpoint"`
	QueueBypassCkie map[string]string `json:"qbp"`
}

func akUrlFromPageBody(baseUrl string, pageBody string) (string, error) {
	re := regexp.MustCompile(`'_setAu', '(.+?)'`)
	match := re.FindStringSubmatch(pageBody)
	if match == nil {
		err := errors.New("Sensor endpoint not found (IP/Fingerprint Blocked)")

		return "", err
	} else {
		akUrl := baseUrl + string(match[1][1:])
		return akUrl, nil
	}
}

func (t *CheckoutTask) RemoveCookie(name string) {
	ckie := &http.Cookie{ Name: name, MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: ""}
	ckies := []*http.Cookie{ ckie }
	t.client.Jar.SetCookies(t.BaseDomainUrl, ckies)
	t.client.Jar.SetCookies(t.BaseUrl, ckies)
	ckies[0].Secure = false
	ckies[0].HttpOnly = true
	t.client.Jar.SetCookies(t.BaseDomainUrl, ckies)
	t.client.Jar.SetCookies(t.BaseUrl, ckies)
	ckies[0].HttpOnly = false
	t.client.Jar.SetCookies(t.BaseDomainUrl, ckies)
	t.client.Jar.SetCookies(t.BaseUrl, ckies)
	ckies[0].Secure = true
	ckies[0].HttpOnly = true
	t.client.Jar.SetCookies(t.BaseDomainUrl, ckies)
	t.client.Jar.SetCookies(t.BaseUrl, ckies)
}

func (t *CheckoutTask) RemoveCookieUrl(name string, url *url.URL) {
	ckie := &http.Cookie{ Name: name, MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: ""}
	ckies := []*http.Cookie{ ckie }
	t.client.Jar.SetCookies(url, ckies)
	ckies[0].Secure = false
	t.client.Jar.SetCookies(url, ckies)
}

func (t *CheckoutTask) SetCookie(cookie *http.Cookie) {
	if t.Debug {
		log.Printf("SetCookie %+v", cookie)
	}
	ckies := []*http.Cookie{ cookie }
	for _, ckie := range t.client.Jar.Cookies(t.BaseDomainUrl) {
		// log.Printf("ckck %v", ckie)
		if ckie.Name == cookie.Name {
			ckies[0].Domain = t.BaseDomainUrl.Host
			ckies[0].Secure = ckie.Secure
			ckies[0].HttpOnly = ckie.HttpOnly
			t.client.Jar.SetCookies(t.BaseDomainUrl, ckies)
			// TODO abck doesnt get set on all sites if return here
			if t.Url.Host != "www.walmart.com" {
				return
			}
		}
	}

	for _, ckie := range t.client.Jar.Cookies(t.BaseUrl) {
		if ckie.Name == cookie.Name {
			ckies[0].Domain = t.BaseUrl.Host
			t.client.Jar.SetCookies(t.BaseUrl, ckies)
			return
		}
	}

	ckies[0].Domain = t.BaseUrl.Host
	t.client.Jar.SetCookies(t.BaseUrl, ckies)
}

func (t *CheckoutTask) postSensorData(sensorData []byte, client *http.Client, ctx context.Context, ua string, akUrlStr string, url *url.URL) (string, error) {
	akUrl, err := url.Parse(akUrlStr) // TODO refactor to akUrl *url.URL
	if err != nil {
		return "", err
	}
	headerOrder := []string {
		"x-newrelic-id",
		"user-agent",
		"accept",
		"sec-fetch-site",
		"sec-fetch-mode",
		"sec-fetch-dest",
		"accept-language",
		"accept-encoding",
		"x-instana-t",
		"x-instana-s",
		"x-instana-l",
		"content-type",
		"content-length",
		"origin",
		"referer",
	}
	is := randHex(8)
	// headers := make([][2]string, 0 )
	// headers = addHeader(headers, "x-newrelic-id", "UwcDVlVUGwIHUVZXAQMHUA==")
	// headers = addHeader(headers, "content-length", strconv.FormatInt(sensorDataLen, 10))
	headers := [][2]string{
			{"x-instana-t", is},
		{"x-instana-s", is},
		{"x-instana-l", "1,correlationType=web;correlationId=" + is},
	}
	headers = addHeader(headers, "user-agent", ua)
	headers = addHeader(headers, "Content-Type", "text/plain;charset=UTF-8")
	headers = addHeader(headers, "Accept", "*/*")
	headers = addHeader(headers, "accept-language","en-US;en;q=0.5")
	headers = addHeader(headers, "origin", "https://" + akUrl.Host) // strings.Replace(strings.Replace(url.String(), url.EscapedPath(), "", 1), "?" + url.RawQuery, )
	headers = addHeader(headers, "Referer", url.String())
	if !strings.Contains(ua, "Firefox") {
		headers = addHeader(headers, "sec-fetch-mode", "cors")
	}
	req := t.makeReq("POST", akUrl, &headers, &headerOrder, &sensorData).WithContext(ctx)
	// req.Close = true
	resp, err := t.doReq(client, req)
	if err != nil {
		DiscardResp(resp)
		return "", err
	}
	body, err := readBody(resp)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		err = errors.New(fmt.Sprintf("Error bypassing akamai: %d", resp.StatusCode))
	}
	if err != nil {
		log.Printf("Error bypassing akamai: %+v", err)

		return "", err
	}
	if len(sensorData) > 0 && !strings.Contains(body, "\"success\"") {
		return "", ErrAkamaiBanned
	}
	t.LogDebug(t.GetCookieValue("_abck"))
	t.LogDebug(body)
	return body, nil
}

func (t *CheckoutTask) updateSensorData(client *http.Client, ctx context.Context, ua string, akUrlStr string, url *url.URL) (string, error) {
	if os.Getenv("DEBUG") != "" {
		log.Printf("abck %s akUrl %s url %s", getCookieValue("_abck", url,  client.Jar), akUrlStr, url.String() )
	}
	// sensorData, err := t.getSensorData(url.String(), getCookieValue("_abck", url, client.Jar))
	sensorData, err := t.getSensorData("https://" + url.Host + "/", getCookieValue("_abck", url, client.Jar))
	if err != nil {
		return "", err
	}
	t.LogDebug(string(sensorData[:]))
	// if os.Getenv("DEBUG") != "" {
	// }
	return t.postSensorData(sensorData, client, ctx, ua, akUrlStr, url)
}


/*
{
  "id": "FV6125",
  "availability_status": "IN_STOCK",
  "variation_list": [
    {
      "sku": "FV6125_530",
      "availability": 0,
      "availability_status": "NOT_AVAILABLE",
      "size": "4"
    },
*/

// func ftlEuAk() {
// 	proxy, err := url.Parse("http://urbanam2:Cd3bKRjO337W7xNm_country-Netherlands_session-IAW8dUdl@proxy.gyroproxies.com:31112")
// 	rt := newRoundTripper(dial(proxy))
// 	// rt := newRoundTripper(net.Dial)
// 	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
// 	client := &http.Client{Transport: rt, Jar: jar,Timeout:30*time.Second}
// 	url, err := url.Parse("https://www.footlocker.co.uk/en/homepage")
// 	resp, err := t.doReq(client, makeGetReq(url, nil))
// 	if err != nil {
// 		log.Fatalf("%+v", err)
// 	}
// 	log.Println(readBody(resp))

// 	for {
// 		updateSensorData(client, "https://www.footlocker.co.uk/public/2193ac875175a2192040657880dad", url)
// 		time.Sleep(time.Duration(3)*time.Second)
// 	}
// }

// func footpatrolAk() {
// 	rt := newRoundTripper(net.Dial)
// 	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
// 	client := &http.Client{Transport: rt, Jar: jar,Timeout:30*time.Second}
// 	url, err := url.Parse("https://m.footpatrol.com/product/nike-x-off-white-dunk-low-qs/351934_footpatrolcom/")
// 	resp, err := t.doReq(client, makeGetReq(url, nil))
// 	if err != nil {
// 		log.Fatalf("%+v", err)
// 	}
// 	log.Println(readBody(resp))

// 	for {
// 		updateSensorData(client, "https://m.footpatrol.com/resources/424d13344a16024a868e82ece46748", url)
// 		time.Sleep(time.Duration(3)*time.Second)
// 	}
// 	updateSensorData(client, "https://m.footpatrol.com/resources/424d13344a16024a868e82ece46748", url)
// 	updateSensorData(client, "https://m.footpatrol.com/resources/424d13344a16024a868e82ece46748", url)
// }


func (t *CheckoutTask) shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	var errUnwrapped error = nil
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		errUnwrapped = unwrapped
	}

	shouldRetry := (t.Ctx.Err() == nil && t.ClientCtx.Err() == nil &&
		retry.IsRecoverable(err) &&
		(errUnwrapped == nil || retry.IsRecoverable(errUnwrapped)))
	if t.Debug {
		debug.PrintStack()
		log.Printf("retry err: IsRecoverable=%+v %+v %+v %+v", retry.IsRecoverable(err), err, errUnwrapped, shouldRetry)
	}
	return shouldRetry
}

func (t *CheckoutTask) RetryTask(rfunc retry.RetryableFunc, opts ...retry.Option) error {
	if t.Ctx.Err() != nil {
		return t.Ctx.Err()
	}
	if t.ClientCtx.Err() != nil {
		return t.ClientCtx.Err()
	}
	opts = append([]retry.Option{retry.RetryIf(t.shouldRetry), retry.Context(t.Ctx), retry.Delay(3 * time.Second), retry.Attempts(1e6), retry.LastErrorOnly(true), retry.DelayType(retry.FixedDelay)}, opts...)
	return retry.Do(rfunc, opts...)
}

func (t *CheckoutTask) Retry(rfunc retry.RetryableFunc, opts ...retry.Option) error {
	/*
	!ErrIsTaskUnrecoverable(err) &&
	(errUnwrapped == nil || !ErrIsTaskUnrecoverable(errUnwrapped)) &&
	*/
	if t.Ctx.Err() != nil {
		return t.Ctx.Err()
	}
	if t.ClientCtx.Err() != nil {
		return t.ClientCtx.Err()
	}
	opts = append([]retry.Option{retry.RetryIf(func(err error) bool {
		shouldRetry := t.shouldRetry(err)
		return shouldRetry && (err == nil || !ErrIsTaskUnrecoverable(err))
	}), retry.Delay(3 * time.Second), retry.Context(t.ClientCtx), retry.Attempts(1e6), retry.LastErrorOnly(true), retry.DelayType(retry.FixedDelay)}, opts...)
	if t.Debug { debug.PrintStack() }
	t.LogDebug("RETRY")
	return retry.Do(rfunc, opts...)
}

func (t *CheckoutTask) SendTaskResult(status string, fields [][2]string) {
	result := map[string]interface{} {}
	result["status"] = status
	for _, kv := range fields {
		result[strings.ToLower(strings.Replace(kv[0], " ", "_", -1))] = kv[1]
	}
	result["__sendnow__"] = "1"
	t.SendTelemetry(result)
}

func (t *CheckoutTask) SendTelemetry(tel map[string]interface{}) {
	// log.Printf("SendTelemetry %+v", tel)
	now := int64(timeMillis())

	t.TelemetryContextMut.Lock()
	for k, v := range t.TelemetryContext {
		tel[k] = v
	}
	t.TelemetryContextMut.Unlock()
	tel["@timestamp"] = time.Unix(0, now * int64(time.Millisecond)).UTC().Format("2006-01-02T15:04:05.000Z")
	tel["elapsed"] = uint64(now) - t.StartTime
  tel["ts"] = float64(now / 1000.0)
	tel["ts_millis"] = now
	if t.Delay > 0 {
		tel["delay"] = int64(t.Delay / time.Millisecond)
	}

	if tel["__sendnow__"] == "1" {
		delete(tel, "__sendnow__")
		go DefaultCloudLogger.SendTelemetryNow(tel)
	} else {
		go DefaultCloudLogger.SendTelemetry(tel)
	}
}


func (t *CheckoutTask) DeclineWebhookFields(fields [][2]string) {
	for idx, field := range fields[:] {
		if field[0] == "Product" && t.Url.Scheme != "ig" {
			fields[idx][1] = fmt.Sprintf("[%s](%s)", field[1], t.Url.String())
		}
	}

	if t.StartAtcTime > 0 && t.StartCheckoutTime > 0 {
		if (t.StartCheckoutTime - t.StartAtcTime) < 1e9 {
			fields = append(fields, [2]string{"ATC Time", fmt.Sprintf("%.02fs", float64(t.StartCheckoutTime - t.StartAtcTime)/float64(1e3))})
		}
	}

	if t.StartCheckoutTime > 0 {
		if t.StartPaymentTime > 0 {
			t.StartCheckoutTime += (timeMillis() - t.StartPaymentTime)
		}
		fields = append(fields, [2]string{"Checkout Time", fmt.Sprintf("%.02fs", float64(timeMillis() - t.StartCheckoutTime)/float64(1e3))})
	}


	if t.StartPaymentTime > 0 {
		fields = append(fields, [2]string{"Payment Time", fmt.Sprintf("%.02fs", float64(timeMillis() - t.StartPaymentTime)/float64(1e3))})
	}

	fields = append(fields, [2]string{"Task Time", fmt.Sprintf("%.02fs", float64(timeMillis() - t.StartTime)/float64(1e3))})

	fields = append(fields, [2]string{"Version", os.Getenv("VERSION")})

	if os.Getenv("PUBLIC_WEBHOOK") != "0" {
		// if t.GetTelemetryContext("cj") == "1" {
		// 	go t.SendWebhook(
		// 		"https://discord.com/api/webhooks/786536792123637800/ZPBp83gVgR5p_phHHvVMnIXavVdL9Wq3LVhIuy6HIsOvM4zBPiv7BZLPiezvlVkBbreS",
		// 		"**RushAIO: Card Declined**",
		// 		fields,
		// 	)
		// } else {
			go t.SendWebhook(
				"https://discord.com/api/webhooks/715809731612246046/KMKutAYc94RxD76kwbC_XeSdb94bUG8NZEi410Odwspl4e-TuH3ITdFiHJpzL8P2S-Mt",
				"**RushAIO: Card Declined**",
				fields,
			)
		// }
	}

	if t.Proxy != nil {
		fields = append(fields, [2]string{"Proxy", "||" + t.Proxy.String() + "||"})
	}
	fields = append(fields, [2]string{"Profile", "||" + t.Profile.Name + "||"})

	go t.SendTaskResult("decline", fields)

	if t.Webhook != "" {
		go t.SendWebhook(
			t.Webhook,
			"**RushAIO: Card Declined**",
			fields,
		)
	}
}
func (t *CheckoutTask) DeclineWebhook(site string, product string, size string) {
	fields := [][2]string{
		{"Website", site},
		{"Product", product},
		{"Size", size},
	}

	go t.DeclineWebhookFields(fields)
}

var NYCTZ, _ = time.LoadLocation("America/New_York")

func (t *CheckoutTask) SendWebhook(webhookUrl string, title string, fields [][2]string) {
	// log.Printf("SendWebhook(%s)", webhookUrl)
	if os.Getenv("WEBHOOK") == "0" {
		return
	}

	if !strings.Contains(webhookUrl, "?wait=true") {
		// gross
		webhookUrl = webhookUrl + "?wait=true"
	}

	var thumburl string
	if fields[0][0] == "image" {
		thumburl = fields[0][1]
		fields = fields[1:]
	}
	bodyMap := map[string]interface{} {
	  "content": "",
	  "username": "RushAIO",
	  "avatar_url": "https://cdn.discordapp.com/icons/671183199912853504/18195857424cad2fbf03d1d39e288a8a.webp?size=512",
	  "tts": false,
	  "embeds": []map[string]interface{} {
	    map[string]interface{} {
	      "title": title,
	      "type": "rich",
	      "description": "",
	      "color": "14177041",
	      "footer": map[string]string {
	        "text": fmt.Sprintf("RushAIO • %s", time.Now().In(NYCTZ).Format("Mon Jan 2 15:04 MST")),
	        "icon_url": "https://cdn.discordapp.com/icons/671183199912853504/18195857424cad2fbf03d1d39e288a8a.webp?size=512",
	      },
	      "thumbnail": map[string]string {
	        "url": thumburl,
	      },
	      "fields": []map[string]interface{} {},
	    },
	  },
	}
	embeds := (bodyMap["embeds"]).([]map[string]interface{})
	emFields := []map[string]interface{} {}
	for _, field := range fields {
		value := field[1]
		if value == "" {
			field[1] = "(None)"
		}
		emFields = append(emFields, map[string]interface{} {
			"name": field[0],
			"value": field[1],
			"inline": true,
		})
	}
	embeds[0]["fields"] = emFields
	body, err := json.Marshal(bodyMap)
	if err != nil {
		log.Printf("%+v", err)

	}
	client := http.Client{}

	retry.Do(func() error {
		resp, err := client.Post(webhookUrl, "application/json", bytes.NewReader(body))
		if resp != nil {
			DiscardResp(resp)
			t.LogDebug("webhook resp %s", resp.Status)
		}
		if err != nil {
			log.Printf("%+v", err)

			return err
		}
		if resp.StatusCode != 200 && resp.StatusCode != 404 {
			errStr := fmt.Sprintf("Webhook Error: %s", resp.Status)
			log.Printf("%s", errStr)
			return errors.New(errStr)
		}
		return nil
	}, retry.Attempts(1e3))
}

func (t *CheckoutTask) SuccessWebhookFields(fields [][2]string) {
	for idx, field := range fields[:] {
		if field[0] == "Product" && t.Url.Scheme != "ig" {
			fields[idx][1] = fmt.Sprintf("[%s](%s)", field[1], t.Url.String())
		}
	}

	if t.StartAtcTime > 0 && t.StartCheckoutTime > 0 {
			fields = append(fields, [2]string{"ATC Time", fmt.Sprintf("%.02fs", float64(t.StartCheckoutTime - t.StartAtcTime)/float64(1e3))})
		}

		if t.StartCheckoutTime > 0 {
			if t.StartPaymentTime > 0 {
				t.StartCheckoutTime += (timeMillis() - t.StartPaymentTime)
			}
			fields = append(fields, [2]string{"Checkout Time", fmt.Sprintf("%.02fs", float64(timeMillis() - t.StartCheckoutTime)/float64(1e3))})
		}


		if t.StartPaymentTime > 0 {
			fields = append(fields, [2]string{"Payment Time", fmt.Sprintf("%.02fs", float64(timeMillis() - t.StartPaymentTime)/float64(1e3))})
		}

		fields = append(fields, [2]string{"Task Time", fmt.Sprintf("%.02fs", float64(timeMillis() - t.StartTime)/float64(1e3))})

	fields = append(fields, [2]string{"Version", os.Getenv("VERSION")})

	if os.Getenv("PUBLIC_WEBHOOK") != "0" {
		// if t.GetTelemetryContext("cj") == "1" {
		// 	go t.SendWebhook(
		// 		"https://discord.com/api/webhooks/786536792123637800/ZPBp83gVgR5p_phHHvVMnIXavVdL9Wq3LVhIuy6HIsOvM4zBPiv7BZLPiezvlVkBbreS",
		// 		"**RushAIO: Card Declined**",
		// 		fields,
		// 	)
		// } else {
			go t.SendWebhook(
				"https://discord.com/api/webhooks/715703844218208356/MN8JcbB-zan7sLZ9B-2cLhMmNNQCKvySp1Vkdlhw2yaG5o770XlVSgOHLWPiSTGmumZm",
				"**RushAIO: Successful Checkout**",
				fields,
			)
		// }
	}

	if t.Proxy != nil {
		fields = append(fields, [2]string{"Proxy", "||" + t.Proxy.String() + "||"})
	}
	fields = append(fields, [2]string{"Profile", "||" + t.Profile.Name + "||"})
	t.TelemetryContextMut.Lock()
	defer t.TelemetryContextMut.Unlock()
	if onum, ok := t.TelemetryContext["ordernum"].(string); ok {
    fields = append(fields, [2]string{"Order Number", "||"+onum+"||"})
  }
	go t.SendTaskResult("success", fields)

	if t.Webhook != "" {
		go t.SendWebhook(
			t.Webhook,
			"**RushAIO: Successful Checkout**",
			fields,
		)
	}
}

func (t *CheckoutTask) DelTelemetryContext(key string) {
	t.TelemetryContextMut.Lock()
	defer t.TelemetryContextMut.Unlock()
	delete(t.TelemetryContext, key)
}

func (t *CheckoutTask) SetTelemetryContext(key string, value interface{}) {
	t.TelemetryContextMut.Lock()
	defer t.TelemetryContextMut.Unlock()
	t.TelemetryContext[key] = value
}

func (t *CheckoutTask) GetTelemetryContext(key string) interface{} {
	t.TelemetryContextMut.Lock()
	defer t.TelemetryContextMut.Unlock()
	return t.TelemetryContext[key]
}

func (t *CheckoutTask) SuccessWebhook(site string, product string, size string) {
	fields := [][2]string{
		{"Website", site},
		{"Product", product},
		{"Size", size},
	}

	go t.SuccessWebhookFields(fields)
}

type AdyenPaymentError struct {
	Message     string `json:"message"`
	ErrorCode   string `json:"errorCode"`
	MessageList []struct {
		Type    string `json:"type"`
		Details struct {
			Code    string `json:"code"`
			Status  int    `json:"status"`
			Message string `json:"message"`
		} `json:"details"`
	} `json:"messageList"`
}

var (
	SIZES = []string {
		"4K",
		"4.5K",
		"5K",
		"5.5K",
		"6K",
		"6.5K",
		"7K",
		"7.5K",
		"8K",
		"8.5K",
		"9K",
		"9.5K",
		"10K",
		"10.5K",
		"11K",
		"11.5K",
		"12K",
		"12.5K",
		"13K",
		"13.5K",
		"1",
		"1.5",
		"2",
		"2.5",
		"3",
		"3.5",
		"4",
		"4.5",
		"5",
		"5.5",
		"6",
		"6.5",
		"7",
		"7.5",
		"8",
		"8.5",
		"9",
		"9.5",
		"10",
		"10.5",
		"11",
		"11.5",
		"12",
		"12.5",
		"13",
		"13.5",
		"14",
		"14.5",
		"15",
		"16",
		"17",
	}

	SIZE_LOOKUP = map[string]string {
		"4K": "230",
		"5K": "240",
		"5.5K": "250",
		"6K": "260",
		"6.5K": "270",
		"7K": "280",
		"7.5K": "290",
		"8K": "310",
		"8.5K": "320",
		"9K": "330",
		"9.5K": "340",
		"10K": "350",
		"10.5K": "370",
		"11K": "380",
		"11.5K": "390",
		"12K": "410",
		"12.5K": "420",
		"13K": "430",
		"13.5K": "440",
		"1": "450",
		"1.5": "470",
		"2": "480",
		"2.5": "490",
		"3": "510",
		"3.5": "520",
		"4": "530",
		"4.5": "540",
		"5": "550",
		"5.5": "560",
		"6": "570",
		"6.5": "580",
		"7": "590",
		"7.5": "600",
		"8": "610",
		"8.5": "620",
		"9": "630",
		"9.5": "640",
		"10": "650",
		"10.5": "660",
		"11": "670",
		"11.5": "680",
		"12": "690",
		"12.5": "700",
		"13": "710",
		"13.5": "720",
		"14": "730",
		"14.5": "740",
		"15": "750",
		"16": "760",
		"17": "780",
	}
)

func GenFakeProfile() (Profile) {
	phone := faker.PhoneNumber().CellPhone()
	name := faker.Name().Name()
	ns := strings.Split(faker.Name().Name(), " ")
	var firstName string
	var lastName string

	if len(ns) == 2 {
		firstName = strings.Split(name, " ")[0]
		lastName = strings.Split(name, " ")[1]
	} else {
		return GenFakeProfile()
	}

	card := Card {
			Number: "5555555555554444",
			Cvc: strconv.Itoa(rand.Intn(899)+100),
			Name: name,
			ExpMonth: rand.Intn(11) + 1,
			ExpYear: rand.Intn(4) + 2021,
			Type: "MASTER",
			PaymentMethodId: "CREDIT_CARD",
			LastFour: "4444", // MASTER, VISA, AMEX , ...
		}
	profile := Profile{
			Email: faker.Internet().UserName() + randHex(4) + "@gmail.com",
			ShippingAddress: Address {
				Address1: "232 2nd St NW",
				Address2: "",
				City: "Washington, DC",
				Zip: "20001",
				State: "DC",
				Country: "US",
				FirstName: firstName,
				LastName: lastName,
				Phone: phone,
			},

			BillingAddress: Address {
				Address1: "232 2nd St NW",
				Address2: "",
				City: "Washington, DC",
				Zip: "20001",
				State: "DC",
				Country: "US",
				FirstName: firstName,
				LastName: lastName,
				Phone: phone,
			},
			Card: card,

		}
	pj, _ := json.Marshal(profile)
	fmt.Println(string(pj[:]))
	return profile
	/*
	*/
}

func defaultTaskApiConfig() TaskConfig  {
	return TaskConfig {
		SensorEndpoints: map[string]string {
			"www.yeezysupply.com": "https://www.yeezysupply.com/static/34c6c422aceti17995e322821c576ad2a",
			"www.finishline.com": "https://www.finishline.com/public/4991abe77no205ffdded0ef7aa5d0fb",
		},
		RcAction: "yeezy_supply_wr_pageview_",
		RcApiKey: "6Lcoe9wUAAAAAOtgfc4c6rnvgptxiBZwDBX3Tqvl",
		RcCookie: "xhwUqgFqfW88H50",
		SizeLookup: SIZE_LOOKUP,
		SlimEndpoint: "__queue/yzysply",
		QueueBypassCkie: map[string]string {
			"akavpau_FL": "",
		},
	}
}

func (tc *TaskConfig) slimUrl() *url.URL {
	if tc.SlimEndpoint == "" {
		return nil
	} else {
		slimUrl, _ := url.Parse(fmt.Sprintf("%s%s", YS_BASE, tc.SlimEndpoint))
		return slimUrl
	}
}

func (t *CheckoutTask) ApiGet(endpoint string) ([]byte, error) {
	body := []byte{}
	err := t.Retry(
		func() error {
			req, _ := stdhttp.NewRequestWithContext(t.ClientCtx, "GET", apiHost() + endpoint, nil)
			req.Header = map[string][]string {
				"authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
				"x-task-id": {t.Id,},
				"x-version": {t.Version,},
			}
			resp, err := t.ApiClient.Do(req)
			if err != nil {
				t.LogDebug("APIGET ERR %+v", err)
				return err
			} else {
				body, err = readBodyBytesStd(resp)
				return err
			}
		},
		retry.Attempts(1e3),
	)
	if err != nil {
		return body, err
	}
	return body, nil
}

func (t *CheckoutTask) ApiGetRetries(endpoint string, retries uint) ([]byte, error) {
	body := []byte{}
	err := t.Retry(
		func() error {
			t.LogDebug(apiHost() + endpoint)
			req, _ := stdhttp.NewRequestWithContext(t.ClientCtx, "GET", apiHost() + endpoint, nil)
			req.Header = map[string][]string {
				"authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
				"x-task-id": {t.Id,},
				"x-version": {t.Version,},
				"wd40": {"1",},
			}
			resp, err := t.ApiClient.Do(req)
			if err != nil {
				t.LogDebug("APIGETRET ERR %+v", err)
				// if m, ok := err.(*StatusCodeRejectedError); ok {
				// 	// if m.StatusCode == 404 {
				// 		return err
				// 	// }
				// }
				return err
			} else {
				body, err = readBodyBytesStd(resp)
				return err
			}
		},
		retry.Attempts(retries),
	)
	if err != nil {
		return body, err
	}
	return body, nil
}

func (t *CheckoutTask) ApiPost(endpoint string, data []byte) ([]byte, error) {
	body := []byte{}
	req, _ := stdhttp.NewRequestWithContext(t.ClientCtx, "GET", apiHost() + endpoint, nil)
	req.Header = map[string][]string {
		"authorization": {fmt.Sprintf("JWT %s", t.GetJwt()),},
		"x-task-id": {t.Id,},
		"x-version": {t.Version,},
		"wd40": {"1",},
	}

	t.LogDebug("API REQ %+v", req)
	req = req.WithContext(context.Background())
	resp, err := t.ApiClient.Do(req)
	if err != nil {
		t.LogDebug("APIPOST ERR %+v", err)
		return body, err
	}
	if resp.StatusCode == 401 {
		return body, &StatusCodeRejectedError{Err:errors.New(fmt.Sprintf("invalid status code: %d", resp.StatusCode)),StatusCode:resp.StatusCode,}
	}
	body, err = readBodyBytesStd(resp)
	t.LogDebug("APIPOST ERR %+v", err)
	t.LogDebug("API RESP %s", string(body))
	return body, err
}

func (t *CheckoutTask) getTaskConfig() (TaskConfig, error) {
	var taskConfig TaskConfig
	body, err := t.ApiGet("/config")
	if err != nil {
		return taskConfig, err
	}
	err = json.Unmarshal(body, &taskConfig)
	return taskConfig, err
}

func (t *CheckoutTask) getClientHello() utls.ClientHelloID {
	ch := utls.HelloChrome_83
  if strings.Contains(t.Url.Host, "i.instagram.com") {// || strings.Contains(host, "prodmobloy2") {
    ch = utls.HelloIOS_Auto
  }
  if strings.Contains(t.Url.Host, "www.yeezysupply.com") && strings.Contains(t.UserAgent, "Firefox") {
    ch = utls.HelloFirefox_82
  }
  // if strings.Contains(t.Url.Host, "www.champssports.com") && t.DefaultProto == "http/1.1" {
  // 	ch = utls.HelloRandomizedNoALPN
  // }
  // if strings.Contains(t.UserAgent, "iOS") || strings.Contains(t.UserAgent, "iPhone") || strings.Contains(t.UserAgent, "iPad") {
  // 	ch = utls.HelloIOS_14
  // }
  return ch
}
func (t *CheckoutTask) newHttpClientProxy(proxy *url.URL) (*http.Client, error) {
	var rt http.RoundTripper
	if proxy != nil {
		rt = newRoundTripper(t.ClientCtx, t.dial(proxy), "tcp", t.getClientHello(), t.DebugCountBytes)
	} else {
		rt = newRoundTripper(t.ClientCtx, t.dialDnsOverHttps(), "tcp", t.getClientHello(), t.DebugCountBytes)
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: rt,
		Jar: jar,
		Timeout:t.dialTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
		    return http.ErrUseLastResponse
		},
	}
	return client, nil
}

func (t *CheckoutTask) newHttpClient() (*http.Client, error) {
	var rt http.RoundTripper
	if t.Proxy != nil {
		if t.Debug {
			log.Printf("Proxy: %v", t.Proxy)
		}
		if t.Proxy.Scheme == "http" {
			var network = "tcp"
			if strings.Contains(t.Proxy.Host, "::") {
				network = "tcp6"
			}
			rt = newRoundTripper(t.ClientCtx, t.dial(t.Proxy), network, t.getClientHello(), t.DebugCountBytes)
		} else if t.Proxy.Scheme == "socks5" {
			var auth *proxy.Auth = nil
			if t.Proxy.User.String() != "" {
				password, _ := t.Proxy.User.Password()
				auth = &proxy.Auth{
					User: t.Proxy.User.Username(),
					Password: password,
				}
			}
			log.Printf("Using proxy %s", t.Proxy.Host)
			dialer, err := proxy.SOCKS5("tcp", t.Proxy.Host, auth, proxy.Direct)
			if err != nil {
				return nil, err
			}
			rt = newRoundTripper(t.ClientCtx, (dialer.(proxy.ContextDialer)).DialContext, "tcp", t.getClientHello(), t.DebugCountBytes)
		}
	} else {
		network := "tcp"
		if false && common.Any(t.DnsCache[t.Url.Host], func(s string) bool { return strings.Contains(s, ":") }) {
			network = "tcp6"
		}
		rt = newRoundTripper(t.ClientCtx, t.dialDnsOverHttps(), network, t.getClientHello(), t.DebugCountBytes)
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: rt,
		Jar: jar,
		Timeout: t.dialTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
		    return http.ErrUseLastResponse
		},
	}
	return client, nil
}

func timeMillis() uint64 {
	return uint64(time.Now().UnixNano() / int64(time.Millisecond))
}

func timeNanos() uint64 {
	return uint64(time.Now().UnixNano())
}


type Dialer func(network, addr string) (net.Conn, error)

func (t *CheckoutTask) makeDialer(fingerprint []byte, skipCAVerification bool) Dialer {
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
				// log.Println("Pinned Key found")
				keyPinValid = true
			}
		}
		if keyPinValid == false {
			tel := map[string]interface{} {
				"event": "apisecping",
				"__sendnow__": "1",
				"issuer": connstate.PeerCertificates[0].Issuer.String(),
				"addr": c.RemoteAddr().String(),
			}
			t.SendTelemetry(tel)
			t.StatusCh <- Status{Message: " ( ͡° ͜ʖ ͡°) <-- meet Charles , he works for me "}
			return nil, errors.New("hey deadcode")
		}
		return c, nil
	}
}

func (t *CheckoutTask) resetClient(newCtx bool) {
	if t.client != nil {
		CloseH2Conns(t.client)
		t.client.CloseIdleConnections()
		if newCtx {
			t.ClientCtxCancel()
			t.ClientCtx, t.ClientCtxCancel = context.WithCancel(t.Ctx)
		}
	}
}

func (t *CheckoutTask) initClient() error {
	if t.client != nil {
		CloseH2Conns(t.client)
		t.client.CloseIdleConnections()
		t.ClientCtxCancel()
	}
	var err error
	t.ClientCtx, t.ClientCtxCancel = context.WithCancel(t.Ctx)
	t.ApiClient = &stdhttp.Client{}
	t.ApiClient.Transport = &stdhttp.Transport{
		DialTLS: t.makeDialer([]byte{0x42, 0x1d, 0x96, 0x88, 0xaf, 0x26, 0x20, 0x42, 0xe7, 0x34, 0x54, 0xcd, 0x72, 0x7, 0x23, 0xa7, 0xcc, 0x22, 0x19, 0x22, 0x4d, 0x2b, 0x83, 0x8e, 0x1b, 0xaf, 0x3a, 0xf4, 0xe1, 0x22, 0x80, 0xd0}, false),
	}
	t.ApiClient.Timeout = t.dialTimeout
	t.client, err = t.newHttpClient()
	t.LastClientInitAt = time.Now()
	return err
}

func (t *CheckoutTask) refreshTaskConfig() (*TaskConfig, error) {
	taskConfig, terr := t.getTaskConfig()
	if terr != nil {
		return &taskConfig, terr
	}
	t.ApiConfig = taskConfig
	return &t.ApiConfig, nil
}

func (t *CheckoutTask) refreshS3TaskConfig() (*TaskConfig, error) {

	taskConfig, terr := t.getTaskConfig()
	if terr != nil {
		return &taskConfig, terr
	}
	t.ApiConfig = taskConfig
	return &t.ApiConfig, nil
}


func (t *CheckoutTask) CookiesStr() string {
	var cookies []string
	for _, c := range t.client.Jar.Cookies(t.BaseUrl) {
		cookies = append(cookies, c.Name + "=" + c.Value)
	}
	return strings.Join(cookies, "; ")
}

func (t *CheckoutTask) S3Get(key string, retries uint) ([]byte, error) {
	var bb []byte
	url_, err := url.Parse("https://rushaio.s3.amazonaws.com/" + key)
	if err != nil {
		return bb, err
	}
	err = t.Retry(func() error {
		r := t.makeGetReq(url_, nil)
		r.Close = true
		resp, err := t.doReq(t.s3client, r)
		if err != nil {
		  DiscardResp(resp)
		  return err
		}
		bb, err = readBodyBytes(resp)
		return err
	}, retry.Attempts(retries))
	return bb, err
}

func (t *CheckoutTask) Run() error {
	defer func() {
		if t.client != nil {
			CloseH2Conns(t.client)
			t.client.CloseIdleConnections()
		}
		if t.Cancel != nil {
			t.Cancel()
		}
		if t.ClientCtxCancel != nil {
			t.ClientCtxCancel()
		}
	}()
	t.WakeUp = make(chan interface{})
	t.cloudCache = NewCloudCache(t.GetJwt)
	whenv := os.Getenv("WEBHOOK")
	if len(whenv) >0 {
		if _, err := url.Parse(whenv); err == nil {
			t.Webhook = os.Getenv("WEBHOOK")
		}
	}
	t.ReqRetries = 3
	if rr, err := strconv.Atoi(os.Getenv("DOREQ_RETRIES")); err == nil {
		t.ReqRetries = uint(rr)
	}
	t.OneCheckoutPerProf = t.Url.Query().Get("oneperprof") == "1"
	t.Url.Query().Del("oneperprof")
	t.dialTimeout = dialTimeout
	t.Metrics = TaskMetrics{Metrics: map[string]interface{}{}}
	t.TelemetryContextMut = &sync.Mutex{}
	t.DnsCacheMut = &sync.Mutex{}
	t.ApiConfig = defaultTaskApiConfig()
	if os.Getenv("HTTP_TRACE") == "1" {
		t.EnableHttpTrace = true
	}
	if t.Debug {
		http2.VerboseLogs = true
	}
	t.DefaultProto = "HTTP/2"
	t.telemetryClient = &http.Client{}
	t.TelemetryContext = map[string]interface{} {
		"uid": t.UserId,
		"task_id": t.Id,
		"url": t.Url.String(),
		"version": os.Getenv("VERSION"),
	}
	if dd := t.Url.Query().Get("dyndelay"); dd == "1" || os.Getenv("DYNDELAY") == "1" {
		t.SetTelemetryContext("dyndelay", "1")
	}
	t.Url.Query().Del("dyndelay")
	t.SetTelemetryContext("cache", os.Getenv("CACHE"))
	if t.Proxy != nil {
		t.TelemetryContextMut.Lock()
		t.TelemetryContext["proxy"] = t.Proxy.String()
		t.ProxyStr = t.Proxy.String()
		t.TelemetryContextMut.Unlock()
	}
	t.DnsCache = map[string][]string{}
	t.DnsCacheActiveHost = map[string]string{}
	t.HeaderOrder = DEFAULT_HEADER_PRI
	t.HeaderCaseFn = strings.ToLower
	t.ProxyTestResults = map[string]bool{}
	t.Debug = os.Getenv("DEBUG") == "1"
	t.BaseUrl, _ = url.Parse(t.Url.Scheme + "://" + t.Url.Host + "/")
	t.BaseDomainUrl, _ = url.Parse(t.Url.Scheme + "://" + strings.Replace(t.Url.Host, "www.", "", 1) + "/")
	// t.DefaultReqClose = true
	t.DefaultReqClose = false // adds Connection: close header..
	if t.UserAgent == "" {
		t.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.102 Safari/537.36"
	}
	t.HeaderBlacklist = map[string]bool {}
	defer close(t.StatusCh)

	rand.Seed(time.Now().UnixNano()) // TODO this is global, should hold prng instance with seed unique to task
	err := t.initClient()
	if err != nil {
		return err
	}
	defer t.ClientCtxCancel()
	t.StartTime = timeMillis()

	if configSize, ok := t.Config["size"]; ok {
		sizes := strings.Split(configSize, ",")
		sizes = Map(sizes, func(size string) string {
			return strings.TrimSpace(strings.ToLower(size))
		})
		sizes = Filter(sizes, func(size string) bool {
			return size != ""
		})
		t.Sizes = sizes
	}

	t.s3client = 	&http.Client{
		// Transport: rt,
		// Jar: jar,
		Timeout:10*time.Second,
	}

	if os.Getenv("JA3") == "1" {
		jurl, _ := url.Parse("https://ja3er.com/json")
		resp, err := t.doReq(t.client, t.makeGetReq(jurl, nil))
		if err != nil {
			log.Fatalf("%+v", err)
		}
		log.Println(readBody(resp))
	}
	if os.Getenv("PROXYCHAIN") == "1" {
		var berr error
		if t.Proxy, berr = NewProxyPassThru(t.Ctx, t.Proxy); berr != nil {
			panic(fmt.Sprintf("%+v", berr))
			return nil
		}
		log.Printf("%+v", t.Proxy.String())
	}

	if t.Url.Query().Get("v6") == "1" || strings.Contains(t.Url.Host, "::") || os.Getenv("SIX") == "9" || (t.Proxy != nil && strings.Contains(t.Proxy.Host, "::")) {
		t.UseIpv6 = true
		t.Url.Query().Del("v6")
		// t.DnsCache = nil
	}
	if t.Url.Query().Get("cjj") == "1" || os.Getenv("CJ") == "1" {
		t.SetTelemetryContext("cj", "1")
		t.Url.Query().Del("cj")
	}
	if err == nil {
		if t.Url.Host == "www.yeezysupply.com" {

			// if rand.Float32() < 0.1 {
			// 	t.TelemetryContext["ys_use_rand_token"] = "1"
			// }
			// if err != nil {
			// 	return err
			// }
			if t.UseIpv6 || rand.Float32() < 0.5 {
				t.UpdateDnsCache(t.Url.Host)
			}
			// if !t.UseIpv6 && len(t.DnsCache) > 0 && rand.Float32() > 0.5 {
			// 	t.SetTelemetryContext("ys_jig", "1")
			// 	t.DnsCache = map[string][]string {
			// 		"www.yeezysupply.com": []string{
			// 			"23.200.159.23",
			// 			"23.200.159.27",
			// 			"23.222.241.58",
			// 			"23.222.241.58",
			// 			"23.217.116.242",
			// 			"104.124.56.66",
			// 			"104.124.56.66",
			// 			"23.200.159.28",
			// 			"23.200.159.28",
			// 			"23.200.159.28",
			// 			"104.77.178.186",
			// 			"104.77.178.186",
			// 			"23.200.159.31",
			// 			"23.200.159.31",
			// 			"23.222.241.57",
			// 			"23.12.145.145",
			// 			"104.76.198.170",
			// 			"23.12.145.132",
			// 			"184.28.50.65",
			// 			"23.200.159.36",
			// 			"23.200.159.36",
			// 			"23.61.11.175",
			// 			"104.77.178.173",
			// 			"23.59.250.25",
			// 			"104.77.178.160",
			// 			"23.15.5.242",
			// 			"23.15.5.242",
			// 			"69.192.7.194",
			// 			"69.192.7.208",
			// 			"23.208.140.97",
			// 			"23.222.241.36",
			// 			"23.222.241.56",
			// 			"23.11.229.201",
			// 			"23.11.229.201",
			// 			"23.11.229.201",
			// 			"23.217.116.170",
			// 			"104.77.178.159",
			// 			"23.42.158.147",
			// 			"23.42.158.147",
			// 			"184.28.219.42",
			// 			"104.77.178.158",
			// 			"184.28.219.56",
			// 			"23.12.145.169",
			// 			"23.200.159.6",
			// 			"23.200.159.6",
			// 			"23.200.159.6",
			// 			"23.222.241.36",
			// 			"104.77.178.151",
			// 			"104.77.178.151",
			// 			"104.77.178.151",
			// 			"23.12.145.164",
			// 			"184.26.53.137",
			// 			"23.200.157.151",
			// 			"23.15.5.170",
			// 			"104.77.178.152",
			// 			"184.28.219.64",
			// 			"23.62.35.80",
			// 			"23.12.145.143",
			// 			"172.232.21.35",
			// 			"23.200.159.25",
			// 			"184.25.225.40",
			// 			"23.38.103.81",
			// 			"23.38.103.81",
			// 			"23.38.103.81",
			// 			"23.61.11.164",
			// 			"23.201.22.209",
			// 			"23.200.159.30",
			// 			"23.50.225.184",
			// 			"23.50.225.184",
			// 			"23.50.225.184",
			// 			"23.217.116.154",
			// 			"23.217.116.154",
			// 			"23.217.116.154",
			// 			"23.61.11.132",
			// 			"104.77.178.158",
			// 			"104.77.178.158",
			// 			"104.77.178.158",
			// 			"23.61.11.139",
			// 			"23.50.51.89",
			// 			"23.50.51.89",
			// 			"23.50.51.89",
			// 			"184.28.219.40",
			// 			"184.28.110.98",
			// 			"23.50.225.115",
			// 			"23.50.225.115",
			// 			"23.38.103.115",
			// 			"104.123.70.186",
			// 			"23.197.195.19",
			// 			"104.124.56.43",
			// 			"23.59.250.80",
			// 			"104.77.178.165",
			// 			"184.28.190.73",
			// 			"104.77.178.168",
			// 			"23.217.116.216",
			// 			"23.217.116.216",
			// 			"104.76.198.160",
			// 			"104.123.70.209",
			// 			"104.124.56.51",
			// 			"104.124.56.51",
			// 			"23.217.116.163",
			// 			"23.217.116.163",
			// 			"23.217.116.163",
			// 			"23.217.116.184",
			// 			"23.217.116.184",
			// 			"104.77.178.159",
			// 			"104.77.178.159",
			// 			"23.62.35.88",
			// 			"23.62.35.88",
			// 			"104.77.178.156",
			// 			"23.12.145.149",
			// 			"23.12.145.149",
			// 			"23.12.145.149",
			// 			"23.12.145.148",
			// 			"23.12.145.148",
			// 			"23.50.51.82",
			// 			"23.201.22.192",
			// 			"23.201.22.192",
			// 			"23.201.22.192",
			// 			"23.12.145.168",
			// 			"104.76.198.201",
			// 			"23.42.158.169",
			// 			"23.12.145.158",
			// 			"104.76.198.147",
			// 			"23.222.241.37",
			// 			"23.222.241.36",
			// 			"23.200.159.20",
			// 			"23.77.218.242",
			// 			"23.77.218.242",
			// 			"23.77.218.242",
			// 			"104.77.178.138",
			// 			"23.50.225.82",
			// 			"23.217.116.202",
			// 			"23.15.5.153",
			// 			"104.76.198.200",
			// 			"104.76.198.200",
			// 			"104.76.198.200",
			// 			"23.50.225.224",
			// 			"23.50.225.224",
			// 			"23.198.6.27",
			// 			"23.67.48.72",
			// 			"23.67.48.72",
			// 			"23.34.62.168",
			// 			"184.28.219.49",
			// 			"104.77.178.155",
			// 			"23.200.159.53",
			// 			"23.217.116.160",
			// 			"23.12.145.150",
			// 			"23.12.145.150",
			// 			"23.67.48.56",
			// 			"23.200.157.149",
			// 			"104.123.70.137",
			// 			"23.197.195.18",
			// 			"23.48.36.132",
			// 			"23.217.129.155",
			// 			"23.12.145.133",
			// 			"23.61.11.162",
			// 			"23.63.251.106",
			// 			"104.77.178.171",
			// 			"104.77.178.171",
			// 			"184.51.149.227",
			// 			"184.51.149.227",
			// 			"104.77.178.165",
			// 			"23.35.71.17",
			// 			"104.76.198.171",
			// 			"23.59.250.120",
			// 			"23.38.170.217",
			// 			"23.77.218.243",
			// 			"23.12.145.153",
			// 			"96.6.42.208",
			// 			"104.123.70.137",
			// 			"23.12.145.141",
			// 			"23.61.11.142",
			// 			"23.200.159.39",
			// 			"23.200.159.39",
			// 			"104.77.178.155",
			// 			"104.77.178.155",
			// 			"104.77.178.155",
			// 			"23.217.116.210",
			// 			"23.217.116.210",
			// 			"23.217.116.210",
			// 			"96.6.42.16",
			// 			"23.217.129.169",
			// 			"23.12.145.137",
			// 			"23.12.145.137",
			// 			"23.12.145.137",
			// 			"23.12.145.148",
			// 			"23.201.22.184",
			// 			"104.77.178.167",
			// 			"104.76.198.217",
			// 			"104.76.198.217",
			// 			"23.201.219.51",
			// 			"23.201.219.51",
			// 			"23.34.62.152",
			// 			"23.34.62.152",
			// 			"23.61.11.176",
			// 			"104.76.198.192",
			// 			"23.46.211.96",
			// 			"23.46.211.96",
			// 			"23.46.211.96",
			// 			"23.12.145.167",
			// 			"23.12.145.167",
			// 			"23.217.116.240",
			// 			"23.67.48.81",
			// 			"23.61.11.159",
			// 			"23.61.11.133",
			// 			"96.6.42.195",
			// 			"23.203.48.154",
			// 			"104.77.178.161",
			// 			"23.199.50.162",
			// 			"23.12.145.137",
			// 			"104.77.178.170",
			// 			"104.77.178.170",
			// 			"104.77.178.170",
			// 			"104.76.198.170",
			// 			"23.38.113.80",
			// 			"23.38.113.80",
			// 			"23.203.48.159",
			// 			"23.59.250.8",
			// 			"23.48.36.126",
			// 			"23.63.251.81",
			// 			"184.28.219.33",
			// 			"23.198.6.3",
			// 			"23.217.116.209",
			// 			"23.217.116.209",
			// 			"23.222.241.43",
			// 			"23.217.129.163",
			// 			"23.217.129.163",
			// 			"23.199.50.232",
			// 			"23.199.50.232",
			// 			"104.123.70.162",
			// 			"172.232.21.89",
			// 			"184.26.53.186",
			// 			"23.198.6.30",
			// 			"23.198.7.185",
			// 			"23.12.145.139",
			// 			"184.51.149.224",
			// 			"23.201.22.161",
			// 			"23.201.22.161",
			// 			"104.77.178.178",
			// 			"104.77.178.178",
			// 			"23.217.116.233",
			// 			"23.62.35.59",
			// 			"23.62.35.59",
			// 			"184.51.149.144",
			// 			"23.61.11.139",
			// 			"23.200.159.10",
			// 			"104.77.178.147",
			// 			"104.77.178.147",
			// 			"104.76.198.146",
			// 			"23.59.250.51",
			// 			"184.26.53.178",
			// 			"104.77.178.142",
			// 			"23.200.159.12",
			// 			"104.77.178.154",
			// 			"104.77.178.154",
			// 			"104.77.178.151",
			// 			"104.77.178.151",
			// 			"23.200.159.17",
			// 			"104.77.178.156",
			// 			"23.217.116.240",
			// 			"172.232.21.89",
			// 			"23.67.48.81",
			// 			"23.46.211.65",
			// 			"104.76.198.138",
			// 			"23.67.48.66",
			// 			"104.76.198.153",
			// 			"23.12.145.147",
			// 			"23.59.250.83",
			// 			"184.51.149.144",
			// 			"23.61.11.132",
			// 			"23.62.35.42",
			// 			"104.77.178.163",
			// 			"23.50.225.72",
			// 			"23.12.145.140",
			// 			"104.76.198.171",
			// 			"23.61.11.144",
			// 			"104.77.178.166",
			// 			"104.77.178.188",
			// 			"23.12.145.137",
			// 			"23.12.145.137",
			// 			"23.12.145.140",
			// 			"23.12.145.140",
			// 			"23.46.211.17",
			// 			"23.46.211.17",
			// 			"23.217.116.226",
			// 			"23.217.116.226",
			// 			"23.217.116.226",
			// 			"23.198.6.3",
			// 			"23.38.170.169",
			// 			"104.77.178.158",
			// 			"104.77.178.158",
			// 			"104.77.178.158",
			// 			"23.61.11.162",
			// 			"23.61.11.162",
			// 			"104.77.178.174",
			// 			"104.77.178.174",
			// 			"23.217.116.251",
			// 			"23.12.145.139",
			// 			"23.200.159.26",
			// 			"23.61.11.164",
			// 			"23.217.116.232",
			// 			"172.232.21.74",
			// 			"23.61.11.140",
			// 			"104.76.198.170",
			// 			"23.46.211.96",
			// 			"23.12.145.148",
			// 			"23.222.241.52",
			// 			"23.222.241.52",
			// 			"23.61.11.162",
			// 			"23.197.50.248",
			// 			"23.197.50.248",
			// 			"23.12.145.144",
			// 			"23.61.11.144",
			// 			"23.12.145.135",
			// 			"23.200.157.155",
			// 			"23.200.159.44",
			// 			"172.232.20.66",
			// 			"23.12.145.155",
			// 			"23.200.157.143",
			// 			"172.232.21.89",
			// 			"104.77.178.161",
			// 			"104.77.178.161",
			// 			"104.77.178.190",
			// 			"23.50.225.83",
			// 			"172.232.21.35",
			// 			"23.12.145.162",
			// 			"23.217.116.162",
			// 			"23.11.231.177",
			// 			"23.61.11.146",
			// 			"23.50.225.81",
			// 			"184.51.149.153",
			// 			"23.217.116.208",
			// 			"104.77.178.166",
			// 			"23.48.36.233",
			// 			"184.51.149.152",
			// 			"96.6.42.154",
			// 			"104.77.178.170",
			// 			"23.61.11.141",
			// 			"23.55.57.145",
			// 			"23.50.225.155",
			// 			"23.203.48.188",
			// 			"23.203.48.188",
			// 			"23.50.225.138",
			// 			"104.76.198.201",
			// 			"23.203.48.177",
			// 			"23.203.48.177",
			// 			"104.123.70.186",
			// 			"104.123.70.186",
			// 			"23.12.145.142",
			// 			"23.200.159.55",
			// 			"23.50.225.243",
			// 			"23.61.11.140",
			// 			"23.217.116.242",
			// 			"23.61.11.162",
			// 			"23.200.159.46",
			// 			"23.61.11.156",
			// 			"23.61.11.156",
			// 			"23.200.159.20",
			// 			"23.200.159.20",
			// 			"23.200.159.20",
			// 			"172.232.7.75",
			// 			"172.232.7.75",
			// 			"23.12.145.135",
			// 			"23.12.145.135",
			// 			"172.232.20.83",
			// 			"23.217.116.171",
			// 			"23.217.116.171",
			// 			"23.217.116.171",
			// 			"104.77.178.147",
			// 			"104.77.178.147",
			// 			"104.77.178.147",
			// 			"23.12.145.163",
			// 			"23.203.48.173",
			// 			"23.203.48.173",
			// 			"23.203.48.173",
			// 			"104.77.178.158",
			// 			"23.12.145.150",
			// 			"23.217.116.201",
			// 			"104.76.198.139",
			// 			"104.76.198.139",
			// 			"23.217.116.227",
			// 			"23.217.116.227",
			// 			"23.217.116.227",
			// 			"184.28.190.89",
			// 			"184.28.190.89",
			// 			"184.28.190.89",
			// 			"23.194.190.233",
			// 			"23.194.190.233",
			// 			"184.51.149.136",
			// 			"23.199.50.176",
			// 			"104.76.198.155",
			// 			"23.221.22.203",
			// 			"23.221.22.203",
			// 			"23.221.22.203",
			// 			"104.76.198.177",
			// 			"104.76.198.177",
			// 			"104.76.198.177",
			// 			"104.77.178.190",
			// 			"23.12.145.135",
			// 			"23.12.145.135",
			// 			"23.12.145.135",
			// 			"23.200.159.45",
			// 			"23.200.159.45",
			// 			"23.67.48.81",
			// 			"23.217.116.202",
			// 			"23.217.116.202",
			// 			"23.194.131.185",
			// 			"23.221.236.88",
			// 			"104.77.178.154",
			// 			"104.77.178.154",
			// 			"23.217.129.176",
			// 			"23.61.11.132",
			// 			"104.77.193.13",
			// 			"104.77.193.13",
			// 			"23.203.48.181",
			// 			"23.203.48.165",
			// 			"184.51.149.208",
			// 			"104.76.198.201",
			// 			"104.76.198.201",
			// 			"104.76.198.201",
			// 			"23.61.11.164",
			// 			"23.61.11.164",
			// 			"23.39.172.34",
			// 			"23.39.172.34",
			// 			"23.39.172.34",
			// 			"23.11.229.211",
			// 			"23.12.145.169",
			// 			"69.192.7.146",
			// 			"69.192.7.146",
			// 			"23.217.116.210",
			// 			"23.217.116.210",
			// 			"23.217.116.210",
			// 			"104.77.178.151",
			// 			"104.77.178.151",
			// 			"23.222.241.52",
			// 			"23.222.241.52",
			// 			"23.222.241.52",
			// 			"23.50.225.209",
			// 			"23.50.225.209",
			// 			"104.77.178.170",
			// 			"104.77.178.170",
			// 			"104.77.178.154",
			// 			"104.77.178.154",
			// 			"104.77.178.154",
			// 			"23.12.147.86",
			// 			"23.12.147.68",
			// 			"23.12.147.86",
			// 			"23.12.147.86",
			// 			"23.12.147.68",
			// 			"23.12.147.68",
			// 			"23.12.147.86",
			// 			"23.12.147.68",
			// 			"23.12.147.68",
			// 			"23.11.229.225",
			// 			"23.12.147.68",
			// 			"23.11.229.160",
			// 			"23.11.229.225",
			// 			"23.11.229.160",
			// 			"23.11.229.225",
			// 			"23.12.147.86",
			// 			"23.12.147.68",
			// 			"23.12.147.68",
			// 			"23.11.229.160",
			// 			"23.12.147.86",
			// 			"184.28.219.8",
			// 			"104.123.70.137",
			// 			"184.28.219.58",
			// 			"184.28.219.8",
			// 			"104.123.70.179",
			// 			"23.217.129.114",
			// 			"184.25.254.185",
			// 			"104.123.70.137",
			// 			"184.28.219.8",
			// 			"184.25.254.192",
			// 			"172.232.16.72",
			// 			"184.25.254.185",
			// 			"184.28.219.8",
			// 			"184.25.254.185",
			// 			"184.25.254.192",
			// 			"184.28.219.8",
			// 			"172.232.16.72",
			// 			"172.232.16.72",
			// 			"104.123.70.137",
			// 			"172.232.16.74",
			// 			"184.28.219.8",
			// 			"104.123.70.137",
			// 			"184.28.219.8",
			// 			"23.46.211.73",
			// 			"104.124.56.43",
			// 			"172.232.21.34",
			// 			"23.6.118.97",
			// 			"23.222.78.10",
			// 			"23.217.129.128",
			// 			"23.222.78.10",
			// 			"23.46.211.83",
			// 			"104.123.70.208",
			// 			"23.55.63.59",
			// 			"23.35.71.179",
			// 			"172.232.16.72",
			// 			"184.28.219.10",
			// 			"184.28.219.8",
			// 			"23.194.190.233",
			// 			"23.219.36.64",
			// 			"23.198.6.3",
			// 			"104.77.150.171",
			// 			"104.77.150.158",
			// 			"104.77.150.187",
			// 			"184.26.137.59",
			// 			"23.198.6.46",
			// 			"23.46.211.96",
			// 			"172.232.7.56",
			// 			"23.194.190.152",
			// 			"104.77.150.158",
			// 			"23.198.6.25",
			// 			"23.46.211.10",
			// 			"23.198.6.10",
			// 			"23.12.147.86",
			// 			"104.77.150.158",
			// 			"23.198.6.46",
			// 			"23.221.236.160",
			// 			"23.38.113.75",
			// 			"23.12.147.68",
			// 			"23.198.6.46",
			// 			"104.77.150.187",
			// 			"23.198.6.3",
			// 			"104.77.150.171",
			// 			"104.77.150.187",
			// 			"23.12.147.86",
			// 			"23.12.147.86",
			// 			"23.198.6.46",
			// 			"23.198.6.10",
			// 			"23.198.6.18",
			// 			"184.25.254.192",
			// 			"104.77.150.187",
			// 			"104.77.150.187",
			// 			"23.38.113.80",
			// 			"104.77.150.187",
			// 			"23.199.49.10",
			// 			"172.232.10.17",
			// 			"23.198.6.25",
			// 			"23.194.190.152",
			// 			"104.77.150.187",
			// 			"23.38.170.25",
			// 			"23.198.6.18",
			// 			"104.77.150.187",
			// 			"104.77.150.158",
			// 			"23.198.6.3",
			// 			"184.25.254.185",
			// 			"104.77.150.187",
			// 			"104.77.150.187",
			// 			"23.198.6.3",
			// 			"172.232.7.35",
			// 			"104.77.150.158",
			// 			"23.12.147.86",
			// 			"23.198.6.3",
			// 			"104.77.150.158",
			// 			"104.77.150.187",
			// 			"23.222.78.96",
			// 			"23.198.6.25",
			// 			"104.77.150.171",
			// 			"104.77.150.187",
			// 			"104.77.150.158",
			// 			"104.77.150.171",
			// 			"184.28.219.10",
			// 			"23.198.6.18",
			// 			"104.77.150.158",
			// 			"23.198.6.3",
			// 			"23.54.18.211",
			// 			"104.77.150.158",
			// 			"23.46.211.96",
			// 			"104.77.150.187",
			// 			"23.198.6.25",
			// 			"23.194.190.233",
			// 			"23.198.6.25",
			// 			"184.28.219.43",
			// 			"23.198.6.18",
			// 			"23.43.165.200",
			// 			"96.6.42.176",
			// 			"104.77.150.187",
			// 			"23.36.241.152",
			// 			"23.32.46.107",
			// 			"23.12.147.68",
			// 			"23.198.6.18",
			// 			"104.77.150.158",
			// 			"23.198.6.10",
			// 			"23.199.49.10",
			// 			"23.12.147.86",
			// 			"23.198.6.18",
			// 			"23.199.49.50",
			// 			"104.77.150.187",
			// 			"23.194.190.152",
			// 			"23.12.147.68",
			// 			"23.194.190.233",
			// 			"69.192.7.154",
			// 			"104.77.150.158",
			// 			"23.12.147.68",
			// 			"23.198.6.18",
			// 			"23.198.6.25",
			// 			"23.197.195.18",
			// 			"23.201.103.104",
			// 			"104.77.150.158",
			// 			"23.198.6.18",
			// 			"23.38.113.80",
			// 			"23.221.236.88",
			// 			"23.12.147.86",
			// 			"23.198.6.25",
			// 			"23.201.103.104",
			// 			"23.198.6.25",
			// 			"23.198.6.18",
			// 			"23.55.63.82",
			// 			"23.194.190.152",
			// 			"23.12.147.74",
			// 			"172.232.7.35",
			// 			"104.77.150.155",
			// 			"23.222.78.83",
			// 			"172.232.7.35",
			// 			"23.12.147.68",
			// 			"23.12.147.68",
			// 			"104.77.150.171",
			// 			"104.77.150.171",
			// 			"104.77.150.171",
			// 			"104.123.70.137",
			// 			"184.28.219.10",
			// 			"104.123.70.179",
			// 			"184.26.53.163",
			// 			"23.66.122.168",
			// 			"184.25.254.185",
			// 			"184.28.219.10",
			// 			"172.232.7.59",
			// 			"184.28.219.10",
			// 			"104.123.70.137",
			// 			"23.12.145.141",
			// 			"23.48.36.235",
			// 			"23.222.78.83",
			// 			"23.198.6.48",
			// 			"23.194.131.178",
			// 			"23.222.78.99",
			// 			"23.12.145.137",
			// 			"23.219.93.89",
			// 			"23.12.145.137",
			// 			"23.222.78.99",
			// 			"23.46.211.10",
			// 			"23.198.6.58",
			// 			"23.222.78.98",
			// 			"23.222.78.114",
			// 			"23.200.1.12",
			// 			"23.200.1.16",
			// 			"23.219.36.139",
			// 			"184.51.3.218",
			// 			"104.126.116.32",
			// 			"23.198.6.33",
			// 			"23.50.225.99",
			// 			"23.33.238.128",
			// 			"184.28.219.10",
			// 			"184.25.254.185",
			// 			"23.66.114.163",
			// 			"184.28.219.8",
			// 			"104.123.70.179",
			// 			"23.50.233.86",
			// 			"104.123.70.179",
			// 			"184.28.219.10",
			// 		},
			// 	}
			// }

			// t.DefaultReqClose = true
			taskConfig, terr := t.refreshTaskConfig()
			if terr != nil {
				return err
			}
			t.StatusCh <- Status{Message: "Starting"}
			err = t.YsCheckout(t.Url, t.Profile, *taskConfig)
		} else if t.Url.Host == "store.nba.com" {
			// t.UpdateDnsCache(t.Url.Host)
			err = t.NbaCheckout()
		} else if t.Url.Host == "www.dickssportinggoods.com" {
			if t.Url.Scheme == "app" {
				t.DsgLoginAuth0()
			} else {
				// t.UpdateDnsCache(t.Url.Host)
				err = t.DicksCheckout()
			}
		} else if t.Url.Host == "www.hibbett.com" {
			// t.DnsCache = map[string][]string {
			// 	"www.hibbett.com": []string{
			// 		"2606:4700:3030::6818:6973",
			// 	},
			// }
			// t.UseIpv6 = true
			t.HeaderBlacklist = map[string]bool {
				"sec-fetch-dest": true,
				"sec-fetch-site": true,
				"accept": true,
				"accept-language": true,
			}
			if t.Url.Scheme == "https" {
				// t.UpdateDnsCache(t.Url.Host)
				t.HeaderOrder = []string{
					"content-length",
					"upgrade-insecure-requests",
					"accept",
					"x-requested-with",
					"user-agent",
					"x-auth-cookie-required",
					"content-type",
					"checkout-authorization",
					"origin",
					"sec-fetch-site",
					"sec-fetch-mode",
					"sec-fetch-dest",
					"sec-fetch-user",
					"referer",
					"accept-encoding",
					"accept-language",
				}
				// t.DefaultReqClose = true
				err = t.HibCheckout(t.Url, t.Config["size"], t.Profile)
			} else if t.Url.Scheme == "app" {
				err = t.HibAppCheckout()
			}
		} else if t.Url.Host == "www.eblens.com" {
			err = t.EblensCheckout()
		} else if t.Url.Host == "www.nordstrom.com" {
			err = t.NordCheckout()
		} else if t.Url.Host == "www.finishline.com"  || t.Url.Host == "www.jdsports.com" {
			// t.DefaultReqClose = false
			if t.Url.Scheme == "https" {
				t.UpdateDnsCache(t.Url.Host)
				// t.DefaultReqClose = false
				err = t.FnlWebCheckout()
			} else if t.Url.Scheme == "app" {
				t.DefaultReqClose = true // TODO test keep alive
				err = t.FnlAppCheckout()
			} else {
				err = &UnsupportedSiteError{Site: t.Url.Scheme + "://" + t.Url.Host}
			}
		} else if t.Url.Host == "shop.ccs.com" {
			err = t.CcsCheckout()
		} else if t.Url.Scheme == "ig" {
			err = t.IgCheckout()
		} else if t.Url.Host == "www.footlocker.dk" || t.Url.Host == "www.footlocker.no" || t.Url.Host == "www.footlocker.ca" || t.Url.Host == "www.footlocker.com" || t.Url.Host == "www.champssports.com" || t.Url.Host == "www.footaction.com" || t.Url.Host == "www.eastbay.com" || t.Url.Host == "www.kidsfootlocker.com" || t.Url.Host == "www.footlocker.eu" {
			if t.UseIpv6 && rand.Float32() < 0.5 {
				// t.DefaultProto = "http/1.1"
			} else if rand.Float32() < 0.5 {
				// t.DefaultProto = "http/1.1"
			}
			t.SetTelemetryContext("default_proto", t.DefaultProto)
			if false {// t.Url.Host == "www.kidsfootlocker.com" || t.Url.Host == "www.footaction.com" {
				t.StatusCh <- Status{Message: "Site Disabled"}
				err = &UnsupportedSiteError{Site: t.Url.Host}
			} else {
				t.StatusCh <- Status{Message: "Starting"}
				// if bb, err := t.S3Get("ddai.enabled", 1); err == nil {
				// 	if pct, err := strconv.Atoi(string(bb)); err == nil {
				// 		if rand.Float32()*100 <= float32(pct) {
				// 			t.DnsCache = map[string][]string{
				// 				"www.footlocker.com": []string{
				// 					"151.101.2.132",
				// 					"151.101.66.132",
				// 					"151.101.130.132",
				// 					"151.101.194.132",
				// 				},
				// 				"www.eastbay.com": []string{
				// 					"151.101.2.132",
				// 					"151.101.66.132",
				// 					"151.101.130.132",
				// 					"151.101.194.132",
				// 				},
				// 			}
				// 	}
				// if os.Getenv("FASTLY") == "1" {
				// 	pn := rand.Intn(255)
				// 				t.DnsCache = map[string][]string{
				// 					"www.footlocker.com": []string{
				// 						fmt.Sprintf("151.101.2.%d", pn),
				// 						fmt.Sprintf("151.101.66.%d", pn),
				// 						fmt.Sprintf("151.101.130.%d", pn),
				// 						fmt.Sprintf("151.101.194.%d", pn),
				// 					},
				// 					"www.eastbay.com": []string{
				// 						"151.101.2.132",
				// 						"151.101.66.132",
				// 						"151.101.130.132",
				// 						"151.101.194.132",
				// 					},
				// 				}
				// 	// t.DefaultReqClose = true
				// }
				// if os.Getenv("FASTLYSEX") == "1" {
				// 	t.UseIpv6 = true

				// }
					// t.DnsCache = map[string][]string{
					// 	"www.footlocker.com": []string{
					// 		"151.101.2.132",
					// 		"151.101.66.132",
					// 		"151.101.130.132",
					// 		"151.101.194.132",
					// 	},
					// 	"www.eastbay.com": []string{
					// 		"[2600:1406:3f::173f:fb6a]",
					// 		// "151.101.2.132",
					// 		// "151.101.66.132",
					// 		// "151.101.130.132",
					// 		// "151.101.194.132",
					// 	},
					// 	"www.footaction.com": []string{
					// 		// "[2a04:4e42::431]",
					// 		// "[2a04:4e42:200::644]",
					// 		// "[2a04:4e42:400::644]",
					// 	},
					// }
					// t.DefaultReqClose = true
					// t.UseIpv6 = true
				if t.UseIpv6 {
								t.DnsCache = map[string][]string{
									t.Url.Host: []string{
														// "[2a04:4e42::431]",

							"2a04:4e42:200::644",
							"2a04:4e42:400::643",
							"2a04:4e42::144",
									},
								}
				} else {
					// t.DnsCache = map[string][]string{
					// 	t.Url.Host: {"20.36.233.132"},
					// }
				}

				// if _, ok := t.DnsCache[t.Url.Host]; !ok {
				// 	t.UpdateDnsCache(t.Url.Host)
				// }

				if t.Delay == 0 {
					t.Delay = time.Duration(1000 + rand.Intn(12000)) * time.Millisecond
				}
				if os.Getenv("DELAY") != "" {
					if delay, err := strconv.Atoi(os.Getenv("DELAY")); err == nil {
						t.Delay = time.Duration(delay) * time.Millisecond
					}
				}
				if t.Config["color"] == "" {
	 				t.Config["color"] = t.Url.Query().Get("color")
	 			}
				err = t.FtlCheckout()
			}
		} else if t.Url.Host == "www.walmart.com" {
			if t.Delay == 0 {
				t.Delay = time.Duration(rand.Intn(10000)) * time.Millisecond
			}
			if os.Getenv("DELAY") != "" {
				if delay, err := strconv.Atoi(os.Getenv("DELAY")); err == nil {
					t.Delay = time.Duration(delay) * time.Millisecond
				}
			}
			// t.UpdateDnsCache(t.Url.Host)
			// t.DnsCache = map[string][]string {
			// 	t.Url.Host: {"35.186.220.184"},
			// }
      t.DefaultReqClose = false
			err = t.WmtCheckout()
		} else if t.Url.Host == "www.totalwine.com" {
			err = t.TotalWineScrape()
		} else if t.Url.Host == "www.footlocker.nl" {
			err = t.FtlEuTest()
	  } else {
			t.StatusCh <- StatusUnsupportedSite
			err = &UnsupportedSiteError{Site: t.Url.Host}
		}
	}

	if t.Url.Host == "www.zalando.co.uk" {
		t.AbckGenLoop("https://www.zalando.co.uk/resources/d2cffe3967rn259e917d81f3e8bfe2cb", func(abck string) bool {
			return strings.HasSuffix(abck, "=~-1~-1~-1") && !strings.HasSuffix(abck, "==~-1~-1~-1") && !strings.Contains(abck, "||")
		})
	}

	if t.Url.Host == "m.zalando.nl" {
		t.AbckGenLoop("https://m.zalando.nl/assets/11da60daui178ca3a73d28f8979d59", func(abck string) bool {
			return strings.HasSuffix(abck, "=~-1~-1~-1") && !strings.HasSuffix(abck, "==~-1~-1~-1") && !strings.Contains(abck, "||")
		})
	}

	if t.Url.Host == "www.solebox.com" {
		t.SbxGenPxCookie()
	}

	if t.Url.Host == "www.xero.com" {
		t.AbckGenLoop("https://www.xero.com/resources/2a40842e63rn1817f7dece8e34759b09", func(abck string) bool {
			return false
		})
	}

	if t.Url.Host == "www.footpatrol.com" {
		if t.Url.Scheme == "app" {
			t.FootpatrolAppAtc()
		} else {
			t.AbckGenLoop("https://www.footpatrol.com/resources/4adc9bc2d3rn2333b4dd2f78ad1daa17", func(abck string) bool {
				return strings.HasSuffix(abck, "=~-1~-1~-1") && !strings.HasSuffix(abck, "==~-1~-1~-1")
			})
		}
	}

	if t.Url.Host == "greenspaces.id" {
		err = t.GsCheckout()
	}

	if t.Url.Host == "www.nike.com" {
		// headers := [][2]string {
		//   {"upgrade-insecure-requests", "1"},
		//   {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		//   {"sec-fetch-site", "none"},
		//   {"sec-fetch-mode", "navigate"},
		//   {"sec-fetch-dest", "document"},
		//   {"accept-encoding", "gzip, deflate, br"},
		//   {"accept-language", "en-US,en;q=0.9"},
		// }
		// u, _ := url.Parse("https://api.nike.com/product_feed/threads/v2/?filter=marketplace%28US%29&filter=language%28en%29&filter=channelName%28Nike.com%29")
		// resp, _ := t.doReq(t.client, t.makeGetReq(u, &headers))
		// b, _ := readBody(resp)
		// log.Println(b)


		t.AbckGenLoop("https://www.nike.com/static/91698b6f882ti2091b8958ff21d04d86a", func(abck string) bool {
			t.LogDebug("check %s", abck)
			return strings.HasSuffix(abck, "=~-1~-1~-1") && !strings.HasSuffix(abck, "==~-1~-1~-1") && !strings.Contains(abck, "||")
		})
	}

	// if t.Url.Host == "www.footlocker.nl" {
	// 	t.AbckGenLoop("https://www.footlocker.nl/static/0ada259f70fti1957ec6fbe7ad5599a75", func(abck string) bool {
	// 		log.Println(abck)
	// 		return strings.HasSuffix(abck, "==~-1~-1~-1") && !strings.Contains(abck, "||")
	// 	})
	// }

	if t.Url.Host == "www.size.co.uk" {
		// t.AbckGenLoop("https://m.size.co.uk/assets/9283d004ui26261004e3a7c1fcc801", func(abck string) bool {
		// 	log.Println(abck)
		// 	return false
		// 	// return strings.HasSuffix(abck, "=~-1~-1~-1") && !strings.HasSuffix(abck, "==~-1~-1~-1") && !strings.Contains(abck, "||")
		// })
		err = t.SizeAtc()
	}

	CloseH2Conns(t.client)

	// log.Printf("task done")

	if err != nil {
		errstr := fmt.Sprintf("%+v", err)
		t.LogDebug(errstr)
		if !strings.Contains(errstr, "canceled") && t.StatusTelemetryEnabled {
			go t.SendTelemetry(map[string]interface{} {
				"event": "task_end_err",
				"task_end_err": errstr,
			})
		}
		//
	}
	// if len(t.AbckJar) > 0 {
	// 	jarJson, _ := json.Marshal(t.AbckJar)
	// 	go t.SendTelemetry(map[string]interface{} {
	// 		"event": "task_end_abck_jar",
	// 		"jar": string(jarJson),
	// 	})
	// }

	return err
}

func CloseH2Conns(client *http.Client) {
	defer func() {
		 recover()
		}()
	if rt, ok := client.Transport.(*roundTripper); ok {
	 if h2trs, ok := rt.transport.(*http2.Transport); ok {
	 	 if h2trs.ConnPool != nil {
	   		h2trs.ConnPool.CloseAll()
	 		}
		}
	}
}

func urlEncodeParamsMap(params map[string]string) []byte {
	bodyKvs := [][]byte{}
	for k,v := range params {
		bodyKvs = append(bodyKvs, []byte(fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v))))
	}
	return bytes.Join(bodyKvs, []byte("&"))
}

func urlEncodeParams(params [][2]string) []byte {
	bodyKvs := [][]byte{}
	for _, kv := range params {
		bodyKvs = append(bodyKvs, []byte(fmt.Sprintf("%s=%s", url.QueryEscape(kv[0]), url.QueryEscape(kv[1]))))
	}
	return bytes.Join(bodyKvs, []byte("&"))
}

func readRespJson(resp *http.Response) (map[string]interface{}, error) {
	dest := map[string]interface{} {}
	by, err := readBodyBytes(resp)
	if err != nil {
		return dest, err
	}
	err = json.Unmarshal(by, &dest)
	return dest, err
}

func readRespJsonDst(resp *http.Response, dest interface{}) (error) {
	by, err := readBodyBytes(resp)
	if err != nil {
		return err
	}
	return json.Unmarshal(by, &dest)
}

func DiscardResp(resp *http.Response) {
	if resp != nil {
		if resp.Body != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}

func DiscardRespStd(resp *stdhttp.Response) {
	if resp != nil {
		if resp.Body != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}


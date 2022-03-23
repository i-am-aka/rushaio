package task

import (
  // "bytes"
  // "bufio"
  // "encoding/asn1"
  "encoding/pem"
  "context"
  "crypto/x509"
  "time"
  // "runtime/debug"
  "github.com/pkg/errors"
  "fmt"
  "io"
  "net"
  "rush/net/http"
  "rush/net/http/mitm"
  "net/url"
  "os"
  "strings"
  "sync"

  utls "github.com/refraction-networking/utls"
  "golang.org/x/net/http2"
)

var errProtocolNegotiated = errors.New("protocol negotiated")

type DialFunc func(context.Context, string, string) (net.Conn, error)

type roundTripper struct {
  sync.Mutex

  transport http.RoundTripper
  dialFn    DialFunc

  initConn net.Conn
  initHost string

  keyLog io.Writer

  context context.Context

  InsecureSkipVerify bool

  Network string

  HttpMitm *mitm.HttpMitm

  ClientHello utls.ClientHelloID

  DebugCountBytes func(uint8, uint)
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
  if rt.transport == nil {
    if err := rt.getTransport(req, rt.context); err != nil {
      return nil, err
    }
  }
  return rt.transport.RoundTrip(req)
}

func (rt *roundTripper) getTransport(req *http.Request, ctx context.Context) error {
  switch strings.ToLower(req.URL.Scheme) {
  case "http":
    rt.transport = &http.Transport{DialContext: rt.dialFn, HttpMitm: rt.HttpMitm, DebugCountBytes: rt.DebugCountBytes}
    return nil
  case "https":
  default:
    return fmt.Errorf("invalid URL scheme: '%v'", req.URL.Scheme)
  }
  // fmt.Printf("REQ PROTO %+v\n", req.Proto)
  var err error
  _, err = rt.dialTLSContextH2(ctx, rt.Network, getDialTLSAddr(req.URL))
  switch err {
  case errProtocolNegotiated:
  case nil:
    return ErrUnexpectedState
    // Should never happen.
    // panic("dialTLSContext returned no error when determining transport")
  default:
    return err
  }

  return nil
}

var ftlPem = []byte(`-----BEGIN CERTIFICATE-----
MIIGIDCCBQigAwIBAgIMFGjfUmwnHpQYxQqeMA0GCSqGSIb3DQEBCwUAMFAxCzAJ
BgNVBAYTAkJFMRkwFwYDVQQKExBHbG9iYWxTaWduIG52LXNhMSYwJAYDVQQDEx1H
bG9iYWxTaWduIFJTQSBPViBTU0wgQ0EgMjAxODAeFw0yMDExMTgyMzExMDJaFw0y
MTEyMjAyMzExMDJaMHMxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlh
MRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRUwEwYDVQQKEwxGYXN0bHksIEluYy4x
IDAeBgNVBAMTF2ouc25pLmdsb2JhbC5mYXN0bHkubmV0MIIBIjANBgkqhkiG9w0B
AQEFAAOCAQ8AMIIBCgKCAQEAvlhS3fQeGeCGNyXXoUjjZDH4LfsAHwb5m5YbjX+D
C94ijKvrXJzt9upZF6WPfQQue/amXM4zzItCQx1adTrbKWdm/XIoxVvFuP2vk7cz
8pts1w94g21HgpFyeaHKOoXQ8BLj+VilloV5XkWJsKHgBtRqt0d6f+nKvyBs8ZD9
9YZI/aTq5PsER8kFWXIalx2Qqrqyyn3imptV1LCqUGkh0vhd+EbEgFHXPwX/8E5z
KFd+p3YutqGZka7bbFrxF/ykvDHkrDih3xwxnF3DT2TA9f08RZsdYVgHrk2w4ovt
xgWAzdNsif34niXLiCT7LPvarVCO/5COvNyb3XFHOYEBywIDAQABo4IC1TCCAtEw
DgYDVR0PAQH/BAQDAgWgMIGOBggrBgEFBQcBAQSBgTB/MEQGCCsGAQUFBzAChjho
dHRwOi8vc2VjdXJlLmdsb2JhbHNpZ24uY29tL2NhY2VydC9nc3JzYW92c3NsY2Ey
MDE4LmNydDA3BggrBgEFBQcwAYYraHR0cDovL29jc3AuZ2xvYmFsc2lnbi5jb20v
Z3Nyc2FvdnNzbGNhMjAxODBWBgNVHSAETzBNMEEGCSsGAQQBoDIBFDA0MDIGCCsG
AQUFBwIBFiZodHRwczovL3d3dy5nbG9iYWxzaWduLmNvbS9yZXBvc2l0b3J5LzAI
BgZngQwBAgIwCQYDVR0TBAIwADA/BgNVHR8EODA2MDSgMqAwhi5odHRwOi8vY3Js
Lmdsb2JhbHNpZ24uY29tL2dzcnNhb3Zzc2xjYTIwMTguY3JsMCIGA1UdEQQbMBmC
F2ouc25pLmdsb2JhbC5mYXN0bHkubmV0MB0GA1UdJQQWMBQGCCsGAQUFBwMBBggr
BgEFBQcDAjAfBgNVHSMEGDAWgBT473/yzXhnqN5vjySNiPGHAwKz6zAdBgNVHQ4E
FgQU7LwD7lOR93efOC8AF+3DBPgxPvIwggEFBgorBgEEAdZ5AgQCBIH2BIHzAPEA
dwBc3EOS/uarRUSxXprUVuYQN/vV+kfcoXOUsl7m9scOygAAAXXdoDd9AAAEAwBI
MEYCIQCcSI+GEe8pLJadt+8fVPOn4ymfx6pCLiSKGAI738ufCAIhAMxW17YlkoV+
KWWIwWYyGMB9Rapz5IDUTVFDxvSSBlnTAHYA9lyUL9F3MCIUVBgIMJRWjuNNExkz
v98MLyALzE7xZOMAAAF13aA1AwAABAMARzBFAiB+XDNoF628GRTJml98pVKSRn/m
5eDVa6uHNZD4B0rpsAIhAOL/g6Af14uhaxFPmLdkGnEHiPMGpkvPnuWUbKoQdI/J
MA0GCSqGSIb3DQEBCwUAA4IBAQCD0SjT3Vp/k/GEHpg8o/BN+UzRI49dMzEWPgfh
YDIXU6TQvzlpsqGSPNcNFCmBNHExiDrfOS5PQdIZU/AroCCwf9wzuSNGrNuFWoq1
ed8+Cq1wVDO8uakC6Kenw/jylNNvBohIWu7ZO3T6ja1lhbLLbNWjiHAjW7n5eiDE
WkmNL01ODY70QB71uJ5W/g2FFWWS9DsEHKSPcmbYhi7pbmgAfSK2q7u3OKVpL9d0
tv+6jOOT1LRIHEKqe45oqZjS/gxI0rMWXpvRS6z5WCt2RUqeL11CrhwQztwmmp/t
8YAVpOVSOSGxe4us37yloOg9WAuWzE6YUCuvxfRVLpWuYo5J
-----END CERTIFICATE-----`)
var FtlBlock, _ = pem.Decode(ftlPem)
var FtlCert, _ = x509.ParseCertificate(FtlBlock.Bytes)


func (rt *roundTripper) dialTLSContextH2(ctx context.Context, network, addr string) (net.Conn, error) {
  if os.Getenv("DEBUG") != "" {
    fmt.Printf("dialTLSContextH2 %s %s %+v cerr=%+v\n", network, addr, ctx, ctx.Err())
    // debug.PrintStack()
  }
  // Unlike rt.transport, this is protected by a critical section
  // since past the initial manual call from getTransport, the HTTP
  // client will be the caller.
  rt.Lock()
  defer rt.Unlock()

  // If we have the connection from when we determined the HTTPS
  // transport to use, return that.

  var host string
  var err error
  if host, _, err = net.SplitHostPort(addr); err != nil {
    host = addr
  }


  if conn := rt.initConn; conn != nil {//&& strings.Contains(conn.RemoteAddr().String(), addr) {
    rt.initConn = nil
    if rt.initHost == host {
      if os.Getenv("DEBUG") != "" {
        fmt.Printf("using initConn\n")
      }
      return conn, nil
    }
  }

  rawConn, err := rt.dialFn(ctx, network, addr)
  if err != nil {
    return nil, err
  }


  // fmt.Println(host)



  // if os.Getenv("HTTP1") == "1" {
  //   ch = utls.HelloChrome_83_H1
  // }
  // fmt.Printf("%+v\n", ch)
  isFoots := (host == "www.footlocker.dk" || host == "www.footlocker.ca" || host == "www.footlocker.com" || host == "www.champssports.com" || host == "www.footaction.com" || host == "www.eastbay.com" || host == "www.kidsfootlocker.com" || host == "www.footlocker.eu")
  conn := utls.UClient(rawConn, &utls.Config{
    ServerName: host,
    InsecureSkipVerify: true, //os.Getenv("PPP") == "1",
    // ClientSessionCache: utls.NewLRUClientSessionCache(0),
    KeyLogWriter: rt.keyLog}, rt.ClientHello)


  if false { conn.SetDeadline(time.Now().Add(7*time.Second)) }
    // KeyLogWriter: rt.keyLog}, utls.HelloRandomized)

  if err = conn.Handshake(); err != nil {
    if os.Getenv("DEBUG") != "" {
      fmt.Printf("handshake err %+v\n", err)
    }
    conn.Close()
    return nil, err
  }

  // fmt.Printf("%+v\n", conn.ConnectionState())

  // commented to fix mixed use http2/http1.1 with single roundtripper
  // if rt.transport != nil {
  //   // fmt.Printf("returning transport %+v\n", err)
  //   return conn, nil
  // }

  // No http.Transport constructed yet, create one based on the results
  // of ALPN.
  // fmt.Println(fmt.Sprintf("protocol %s %s", conn.ConnectionState().NegotiatedProtocol, http2.NextProtoTLS))


  // if isFoots{
  //   fmt.Printf("\n\nCEQ=%v\n\n", strings.Contains(conn.ConnectionState().PeerCertificates[0].Issuer.String(), "GlobalSign CloudSSL CA"))
  // }
  iss := conn.ConnectionState().PeerCertificates[0].Issuer.String()
  if os.Getenv("PPP") != "1" && isFoots && !(strings.Contains(iss, "COMODO") || strings.Contains(iss, "GlobalSign") || strings.Contains(iss, "Let's Encrypt") || strings.Contains(iss, "GeoTrust")|| strings.Contains(iss, "DigiCert")){
    fmt.Printf("%s\n", iss)
    return nil, ErrForbiddeen
  }

  prot := conn.ConnectionState().NegotiatedProtocol
  // fmt.Printf("%v\n", string(b.Bytes()))
  // fmt.Printf("PROT (H2) %+v\n", prot)
  switch prot {
  case http2.NextProtoTLS:
    // The remote peer is speaking HTTP 2 + TLS.
    if rt.ClientHello == utls.HelloFirefox_82 {
      rt.transport = &http2.Transport{
        Context: rt.context,
        DialTLSContext: rt.dialTLSHTTP2,
        DisableCompression: true,
        MaxHeaderListSize: 0xffffffff,
        InitialWindowSize: 131072,
        InitialHeaderTableSize: 65536,
        // MaxHeaderListSize: 32768,
        // InitialWindowSize: 65535,
        // InitialHeaderTableSize: 4096,
        PushHandler: newPushHandler(),
        DebugCountBytes: rt.DebugCountBytes,

        // ReadIdleTimeout: 5*time.Second,
        // PingTimeout: 5*time.Second,
        // IdleConnTimeout: 10*time.Second,
        // ConnPool:
      }
    } else {
      rt.transport = &http2.Transport{
        Context: rt.context,
        DialTLSContext: rt.dialTLSHTTP2,
        DisableCompression: true,
        MaxHeaderListSize: 262144,
        InitialWindowSize: 6291456,
        InitialHeaderTableSize: 65536,
        PushHandler: newPushHandler(),
        DebugCountBytes: rt.DebugCountBytes,

        // ReadIdleTimeout: 5*time.Second,
        // PingTimeout: 5*time.Second,
        // IdleConnTimeout: 10*time.Second,
        // ConnPool:
      }
    }
        // rt.transport.InitialWindowSize = 65535
      // rt.transport.InitialHeaderTableSize = 4096
      // rt.transport.MaxHeaderListSize = 32768
  default:
    // Assume the remote peer is speaking HTTP 1.x + TLS.
    rt.transport = &http.Transport{DebugCountBytes: rt.DebugCountBytes, DialTLSContext: rt.dialTLSContextH2, DisableCompression: true, DisableKeepAlives: false, MaxIdleConns: 1, HttpMitm: rt.HttpMitm }
  }

  // Stash the connection just established for use servicing the
  // actual request (should be near-immediate).
  rt.initConn = conn
  rt.initHost = host

  return nil, errProtocolNegotiated
}

func newPushHandler() *PushHandler {
  return &PushHandler{
    done: make(chan struct{}),
  }
}

type PushHandler struct {
  promise       *http.Request
  origReqURL    *url.URL
  origReqHeader http.Header
  push          *http.Response
  pushErr       error
  done          chan struct{}
}

func (ph *PushHandler) HandlePush(r *http2.PushedRequest) {
  ph.promise = r.Promise
  ph.origReqURL = r.OriginalRequestURL
  ph.origReqHeader = r.OriginalRequestHeader
  ph.push, ph.pushErr = r.ReadResponse(r.Promise.Context())
  if ph.pushErr != nil {
    DiscardResp(ph.push)
  }
  if ph.push != nil {
    DiscardResp(ph.push)
  }
}

func (rt *roundTripper) dialTLSHTTP1(context context.Context, network, addr string) (net.Conn, error) {
  return rt.dialTLSContextH2(context, network, addr)
}

func (rt *roundTripper) dialTLSHTTP2(context context.Context, network, addr string) (net.Conn, error) {
  return rt.dialTLSContextH2(context, network, addr)
}

func getDialTLSAddr(u *url.URL) string {
  host, port, err := net.SplitHostPort(u.Host)
  if err == nil {
    return net.JoinHostPort(host, port)
  }

  return net.JoinHostPort(u.Host, u.Scheme)
}


func getKeyLog() (io.Writer, error)  {
  return os.OpenFile(os.Getenv("KEYLOG_FN"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}


func newRoundTripper(ctx context.Context, dialFn DialFunc, network string, clientHello utls.ClientHelloID, DebugCountBytes func(uint8, uint)) http.RoundTripper {
  ctx, _ = context.WithCancel(ctx)
  _keyLog, err := getKeyLog()
  keyLog := _keyLog
  if err != nil {
    keyLog = nil
  }

  return &roundTripper{
    DebugCountBytes: DebugCountBytes,
    dialFn: dialFn,
    keyLog: keyLog,
    context: ctx,
    Network: network,
    InsecureSkipVerify: true,//os.Getenv("INSECURE_SKIP_VERIFY") == "1",
    ClientHello: clientHello,
  }
}

// func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
//   timeout := time.After(rt.RoundTripTimeout)
//   respc := make(chan respAndErr, 1)
//   pctx := req.Context()
//   if pctx == nil  {
//     pctx = rt.context
//   }
//   ctx, cancel := context.WithDeadline(pctx, time.Now().Add(rt.RoundTripTimeout))
//   defer cancel()
//   req = req.WithContext(ctx)
//   go func() {
//     if rt.transport == nil {
//       if err := rt.getTransport(req, rt.context); err != nil {
//         respc <- respAndErr{resp: nil, err: err}
//         return
//       }
//     }
//     resp, err := rt.transport.RoundTrip(req)
//     respc <- respAndErr{resp: resp, err: err}
//   }()
//   select {
//   case <-timeout:
//     return nil, netTimeoutError{
//       error: fmt.Errorf("timed out after %s", rt.RoundTripTimeout),
//     }
//   case r := <-respc:
//     return r.resp, r.err
//   }
// }
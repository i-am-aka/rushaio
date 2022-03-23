package fp

import (
  utls "github.com/refraction-networking/utls"
  "golang.org/x/net/http2"
)

type Http2Fingerprint struct {
  PseudoHeaderOrder []string
  DisableCompression bool
  MaxHeaderListSize uint64
  InitialWindowSize uint64
  InitialHeaderTableSize uint64
  TransportDefaultConnFlow int
  Priority http2.PriorityParam
}

type Fingerprint struct {
  UserAgent string
  HeaderOrder []string
  TlsSpec utls.ClientHelloSpec
  Http2Spec Http2Fingerprint
}

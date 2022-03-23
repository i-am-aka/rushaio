package mitm

type HttpMitm struct {
  RoundTripCompleted func(*RoundTrip)
}

type RoundTrip struct {
  Status          string      `json:"status"`
  Method          string      `json:"method"`
  ProtocolVersion string      `json:"protocolVersion"`
  Scheme          string      `json:"scheme"`
  Host            string      `json:"host"`
  ActualPort      int         `json:"actualPort"`
  Path            string      `json:"path"`
  Query           interface{} `json:"query"`
  Tunnel          bool        `json:"tunnel"`
  KeptAlive       bool        `json:"keptAlive"`
  WebSocket       bool        `json:"webSocket"`
  RemoteAddress   string      `json:"remoteAddress"`
  ClientAddress   string      `json:"clientAddress"`
  ClientPort      int         `json:"clientPort"`
  Times           Times       `json:"times"`
  Durations       Durations   `json:"durations"`
  Speeds          Speeds      `json:"speeds"`
  TotalSize       int         `json:"totalSize"`
  Ssl             Ssl         `json:"ssl"`
  Alpn            Alpn        `json:"alpn"`
  Request         Request     `json:"request"`
  Response        Response    `json:"response"`
}
type Times struct {
  Start           string `json:"start"`
  RequestBegin    string `json:"requestBegin"`
  RequestComplete string `json:"requestComplete"`
  ResponseBegin   string `json:"responseBegin"`
  End             string `json:"end"`
}
type Durations struct {
  Total    int         `json:"total"`
  DNS      interface{} `json:"dns"`
  Connect  interface{} `json:"connect"`
  Ssl      interface{} `json:"ssl"`
  Request  int         `json:"request"`
  Response int         `json:"response"`
  Latency  int         `json:"latency"`
}
type Speeds struct {
  Overall  int `json:"overall"`
  Request  int `json:"request"`
  Response int `json:"response"`
}
type Ssl struct {
  Protocol    string `json:"protocol"`
  CipherSuite string `json:"cipherSuite"`
}
type Alpn struct {
  Protocol string `json:"protocol"`
}
type Sizes struct {
  Headers int `json:"headers"`
  Body    int `json:"body"`
}
type Headers struct {
  Name  string `json:"name"`
  Value string `json:"value"`
}
type Header struct {
  Headers []Headers `json:"headers"`
}
type Request struct {
  Sizes           Sizes       `json:"sizes"`
  MimeType        string      `json:"mimeType"`
  Charset         interface{} `json:"charset"`
  ContentEncoding interface{} `json:"contentEncoding"`
  Header          Header      `json:"header"`
}
type Body struct {
  Text    string `json:"text"`
  Charset string `json:"charset"`
  Decoded bool   `json:"decoded"`
}
type Response struct {
  Status          int    `json:"status"`
  Sizes           Sizes  `json:"sizes"`
  MimeType        string `json:"mimeType"`
  Charset         string `json:"charset"`
  ContentEncoding string `json:"contentEncoding"`
  Header          Header `json:"header"`
  Body            Body   `json:"body"`
}

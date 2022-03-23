package task

type CharlesHttpTrace []struct {
  Status          string `json:"status"`
  Method          string `json:"method"`
  ProtocolVersion string `json:"protocolVersion"`
  Scheme          string `json:"scheme"`
  Host            string `json:"host"`
  ActualPort      int    `json:"actualPort"`
  Path            string `json:"path"`
  Query           string `json:"query"`
  Tunnel          bool   `json:"tunnel"`
  KeptAlive       bool   `json:"keptAlive"`
  WebSocket       bool   `json:"webSocket"`
  RemoteAddress   string `json:"remoteAddress"`
  ClientAddress   string `json:"clientAddress"`
  ClientPort      int    `json:"clientPort"`
  Times           struct {
    Start           string `json:"start"`
    RequestBegin    string `json:"requestBegin"`
    RequestComplete string `json:"requestComplete"`
    ResponseBegin   string `json:"responseBegin"`
    End             string `json:"end"`
  } `json:"times"`
  Durations struct {
    Total    int         `json:"total"`
    DNS      interface{} `json:"dns"`
    Connect  interface{} `json:"connect"`
    Ssl      interface{} `json:"ssl"`
    Request  int         `json:"request"`
    Response int         `json:"response"`
    Latency  int         `json:"latency"`
  } `json:"durations"`
  Speeds struct {
    Overall  int `json:"overall"`
    Request  int `json:"request"`
    Response int `json:"response"`
  } `json:"speeds"`
  TotalSize int `json:"totalSize"`
  Ssl       struct {
    Protocol    string `json:"protocol"`
    CipherSuite string `json:"cipherSuite"`
  } `json:"ssl"`
  Alpn struct {
    Protocol string `json:"protocol"`
  } `json:"alpn"`
  Request struct {
    Sizes struct {
      Headers int `json:"headers"`
      Body    int `json:"body"`
    } `json:"sizes"`
    MimeType        string      `json:"mimeType"`
    Charset         interface{} `json:"charset"`
    ContentEncoding interface{} `json:"contentEncoding"`
    Header          struct {
      Headers []struct {
        Name  string `json:"name"`
        Value string `json:"value"`
      } `json:"headers"`
    } `json:"header"`
    Body struct {
      Text    string      `json:"text"`
      Charset interface{} `json:"charset"`
    } `json:"body"`
  } `json:"request"`
  Response struct {
    Status int `json:"status"`
    Sizes  struct {
      Headers int `json:"headers"`
      Body    int `json:"body"`
    } `json:"sizes"`
    MimeType        string      `json:"mimeType"`
    Charset         interface{} `json:"charset"`
    ContentEncoding string      `json:"contentEncoding"`
    Header          struct {
      Headers []struct {
        Name  string `json:"name"`
        Value string `json:"value"`
      } `json:"headers"`
    } `json:"header"`
    Body struct {
      Text    string      `json:"text"`
      Charset interface{} `json:"charset"`
      Decoded bool        `json:"decoded"`
    } `json:"body"`
  } `json:"response"`
}
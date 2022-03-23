package main

import (
  "runtime"
  "fmt"
  "github.com/pkg/errors"
  "log"
  "bytes"
  "context"
  "net/url"
  "io"
  "io/ioutil"
  "encoding/json"
  "math"
  "net/http"
  "os"
  "os/exec"
  "math/rand"
  "regexp"
  "strings"
  "strconv"
  "sync"
  "rush/akgen"
  "github.com/gorilla/handlers"
  "github.com/dgrijalva/jwt-go"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/minio/minio-go/v7"
  "github.com/minio/minio-go/v7/pkg/credentials"

  "github.com/avast/retry-go"
)

type QCkie struct {
  Ts uint
  Name string
  Value string
}

var ErrInvalidQckieString = errors.New("Invalid QCkie String")

func ParseQCkie(ckie string) (*QCkie, error) {
  bsp := strings.Split(ckie, "=")
  if len(bsp) < 2 {
    return nil, ErrInvalidQckieString
  }
  name := bsp[0]
  val := strings.Join(bsp[1:], "=")
  if !ValidQckieNames[name] {
    return nil, ErrInvalidQckieString
  }
  if len(val) < 46 {
    return nil, ErrInvalidQckieString
  }
  ts, err := strconv.Atoi(val[:10])
  if err != nil {
    return nil, err
  }
  return &QCkie{
    Ts: uint(ts),
    Name: name,
    Value: val,
  }, nil
}
type Api struct {
  CapReqMinLimit int
  ConfigMut sync.Mutex
  CapHosts []string
  CapHostClients map[string]*http.Client
  Samples *akgen.Samples
  AdyenFps []string
  QCkieStore map[string]*QCkie
  QCkieMutex sync.RWMutex
  Debug bool
  MaxClientEnforcerSemaphore chan struct{}
  Client *http.Client
  minioClient *minio.Client
  LogRedis *redis.Client
  TaskRedis *redis.Client
  Ctx context.Context
}

func (api *Api) Log(fmtstr string, args ...interface{}) {
  if api.Debug {
    log.Printf(fmtstr, args...)
  }
}

var (
  ValidQckieNames = map[string]bool {
    "akavpau_CS": true,
    "akavpau_FA": true,
    "akavpau_EB": true,
    "akavpau_FL": true,
  }
)

type TaskSession struct {
  Id string `json:"id"`
  StartTime uint64 `json:"start_time"`
  Duration uint64 `json:"duration"`
  TaskId string `json:"task_id"`
  Version string `json:"version"`
  Host string `json:"host"`
  Proxy string `json:"proxy"`
  UserAgent string `json:"user_agent"`
  RealIp string `json:"real_ip"`
  Cookies map[string]map[string]interface{} `json:"cookies"`
  Metrics map[string]interface{} `json:"metrics"`
}

func (s *TaskSession) Key() string {
  // if s.RealIp != "" {
  //  return "sesh_" + s.Host + "_" + s.RealIp
  // } else {
  return "sesh2_" + s.Host + "_" + s.Proxy
  // }
}

func (s *TaskSession) Value() string {
  tjb, _ := json.Marshal(s)
  return string(tjb)
}

func (api *Api) handle_count_trusted_session(w http.ResponseWriter, req *http.Request) {
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "400: could not read body", 400)
    return
  }
  var sesh TaskSession
  err = json.Unmarshal(body, &sesh)
  if err != nil {
    http.Error(w, "400: could not parse body", 400)
    return
  }
  if sesh.Host == "" || sesh.Proxy == "" { // && sesh.RealIp == "") {
    http.Error(w, "400: bad request", 400)
    return
  }

  v := api.TaskRedis.Do(req.Context(), "GET", "config:max_sessions_per_proxy")
  if err = v.Err(); err == nil {
    if d, ok := v.Val().(string); ok {
      if n, err := strconv.Atoi(d); err == nil {
        maxSesh = uint(n)
      }
    }
  }
  v = api.TaskRedis.Do(req.Context(), "SCARD", sesh.Key())
  if err = v.Err(); err == nil {
    fmt.Fprintf(w, `{"count": %d, "max": %d}`, v.Val().(int64), maxSesh)
  } else {
    api.Log("%+v", err)
    http.Error(w, "404", 404)
  }
}

var maxSesh uint = 16

func (api *Api) handle_push_trusted_session(w http.ResponseWriter, req *http.Request) {
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "400: could not read body", 400)
    return
  }
  var sesh TaskSession
  err = json.Unmarshal(body, &sesh)
  if err != nil {
    http.Error(w, "400: could not parse body", 400)
    return
  }
  if sesh.Host == "" || sesh.Proxy == "" { // && sesh.RealIp == "") {
    http.Error(w, "400: bad request", 400)
    return
  }
  // api.Log("push %+v %+v", sesh, sesh.Key())

  v := api.TaskRedis.Do(req.Context(), "GET", "config:max_sessions_per_proxy")
  if err = v.Err(); err == nil {
    if d, ok := v.Val().(string); ok {
      if n, err := strconv.Atoi(d); err == nil {
        maxSesh = uint(n)
      }
    }
  }

  /*
  be like a browser jar

  session -> state
  session needs *id*
  so it can be updated at will

  proxy -> set(sid)
  sid -> session


  update sesh: set sid = newsesh

  */
  v = api.TaskRedis.Do(req.Context(), "SCARD", sesh.Key())
  if err = v.Err(); err == nil {
    if v.Val().(int64) >= int64(maxSesh) {
      v = api.TaskRedis.Do(api.Ctx, "SPOP", sesh.Key())
      if err = v.Err(); err == nil {
        api.TaskRedis.Do(api.Ctx, "DEL", sesh.Key() + "_" + v.Val().(string))
      }
    }
    api.TaskRedis.Do(api.Ctx, "SET", sesh.Key() + "_" + sesh.Id, sesh.Value())
    api.TaskRedis.Do(api.Ctx, "EXPIRE", sesh.Key() + "_" + sesh.Id, 86400*60)
    api.TaskRedis.Do(api.Ctx, "SADD", sesh.Key(), sesh.Id)
    api.TaskRedis.Do(api.Ctx, "EXPIRE", sesh.Key(), 86400*60)
  } else {
    api.Log("%+v", err)
    http.Error(w, "404", 404)
  }




  v = api.TaskRedis.Do(req.Context(), "SCARD", sesh.Key())
  if err = v.Err(); err == nil {
    fmt.Fprintf(w, `{"count":%d,"max":%d}`, v.Val().(int64), maxSesh)
  } else {
    api.Log("%+v", err)
    http.Error(w, "404", 404)
  }
}

func (api *Api) handle_push_trusted_cap_cookie(w http.ResponseWriter, req *http.Request) {
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "400: could not read body", 400)
    return
  }

  wind := int(time.Now().Unix() / 120)
  k := fmt.Sprintf("dd_cap_cachel_%d", wind)
  x := 0
  maxShare := 10
  v := api.TaskRedis.Do(req.Context(), "GET", "config:max_cap_share")
  if err = v.Err(); err == nil {
    if d, ok := v.Val().(string); ok {
      if n, err := strconv.Atoi(d); err == nil {
        maxShare = n
      }
    }
  }
  for x < maxShare {
    x += 1
    api.Log("push %s", string(body))
    go api.TaskRedis.Do(context.Background(), "LPUSH", k, string(body))
  }
  api.TaskRedis.Do(context.Background(), "EXPIREAT", k, (wind+1)*120)
}

func (api *Api) handle_get_trusted_cap_cookie(w http.ResponseWriter, req *http.Request) {
  wind := int(time.Now().Unix() / 60)
  k := fmt.Sprintf("dd_cap_cachel_%d", wind)
  v := api.TaskRedis.Do(context.Background(), "LPOP", k)
  if err := v.Err(); err == nil {
  	api.Log("pop %+v", v.Val().(string))
    io.WriteString(w, v.Val().(string))
  } else {
    api.Log("pop err=%+v", err)
    http.Error(w, "404", 404)
  }
}

func (api *Api) handle_remove_trusted_session(w http.ResponseWriter, req *http.Request) {
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "400: could not read body", 400)
    return
  }
  var sesh TaskSession
  err = json.Unmarshal(body, &sesh)
  if err != nil {
    http.Error(w, "400: could not parse body", 400)
    return
  }
  if sesh.Host == "" || (sesh.Proxy == "" && sesh.RealIp == "") {
    http.Error(w, "400: bad request", 400)
    return
  }

  api.Log("remove %+v %+v %+v", sesh, sesh.Key(), sesh.Value())
  api.TaskRedis.Do(req.Context(), "DEL", sesh.Key() + "_" + sesh.Id)
  v := api.TaskRedis.Do(req.Context(), "SREM", sesh.Key(), sesh.Id)
  if err = v.Err(); err == nil {
    fmt.Fprintf(w, "1")
  } else {
    api.Log("%+v", err)
    http.Error(w, "404", 404)
  }
}

func (api *Api) handle_get_trusted_session(w http.ResponseWriter, req *http.Request) {
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "400: could not read body", 400)
    return
  }
  var sesh TaskSession
  err = json.Unmarshal(body, &sesh)
  if err != nil {
    http.Error(w, "400: could not parse body", 400)
    return
  }
  if sesh.Host == "" || (sesh.Proxy == "" && sesh.RealIp == "") {
    http.Error(w, "400: bad request", 400)
    return
  }

  // api.Log("get %+v %+v", sesh, sesh.Key())
  v := api.TaskRedis.Do(api.Ctx, "SPOP", sesh.Key())
  if err = v.Err(); err == nil {
    vv := v.Val().(string)
    v = api.TaskRedis.Do(api.Ctx, "GET", sesh.Key() + "_" + vv)
    if err = v.Err(); err == nil {
      vv = v.Val().(string)
      io.WriteString(w, vv)
    // fmt.Fprintf(w, vv)
    }
  } else {
    api.Log("%+v", err)
    http.Error(w, "404", 404)
  }
}

type ProxyTestRequest struct {
  Proxy string `json:"proxy"`
  TestUrl string `json:"test_url"`
}

func (api *Api) handle_test_prox(w http.ResponseWriter, req *http.Request) {
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "400: could not read body", 400)
    return
  }
  var preq ProxyTestRequest
  err = json.Unmarshal(body, &preq)
  if err != nil {
    http.Error(w, "400: could not parse body", 400)
    return
  }
  if pu, err := url.Parse(preq.Proxy); err == nil {
    client := &http.Client{Timeout: 5*time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(pu)}}
    if req, err := http.NewRequestWithContext(api.Ctx, "GET", preq.TestUrl, nil); err == nil {
      if resp, err := client.Do(req); err == nil {
        if resp.StatusCode == 200 {
          io.WriteString(w, "1")
          return
        }
      }
    }
  }
  io.WriteString(w, "0")
}

func (api *Api) handle_get_abck(w http.ResponseWriter, req *http.Request) {
  // host := req.URL.Query().Get("host")
  // v := api.Redis.Do(req.Context(), "LPOP", "abck_" + host)

  // if err := v.Err(); err == nil {
  //  fmt.Fprintf(w, v.Val().(string))
  // } else {
  http.Error(w, "404", 404)
}

func (api *Api) handle_get_qckie(w http.ResponseWriter, req *http.Request) {
  name := strings.Replace(req.URL.Query().Get("name"), "akavpwr", "akavpau", -1)
  if !ValidQckieNames[name] {
    http.Error(w, "400: invalid cookie name", 400)
    return
  }

  var ckieval string
  api.QCkieMutex.Lock()
  if ckie, ok := api.QCkieStore[name]; ok {
    ckieval = ckie.Value
  }
  api.QCkieMutex.Unlock()
  io.WriteString(w, ckieval)
}

func (api *Api) handle_put_qckie(w http.ResponseWriter, req *http.Request) {
  fmt.Fprintf(w, "ok")
  return

  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "400: could not read body", 400)
    return
  }

  qckie, err := ParseQCkie(string(body))
  if err != nil {
    http.Error(w, "400: invalid qckie", 400)
    return
  }

  api.QCkieMutex.Lock()
  if oldckie, ok := api.QCkieStore[qckie.Name]; ok {
    if oldckie.Ts < qckie.Ts {
      api.QCkieStore[qckie.Name] = qckie
    }
  } else {
    api.QCkieStore[qckie.Name] = qckie
  }
  api.QCkieMutex.Unlock()
}

type SensorDataRequest struct {
  Abck string `json:"_abck"`
  Ua string `json:"ua"`
  Url string `json:"url"`
  Stage string `json:"stage"` // one of: 'onload', 'fp', 's'
  AjIndx int `json:"aj_indx"`
}

type SensorDataPayload struct {
  SensorData string `json:"sensor_data"`
}

func (api *Api) handle_root(w http.ResponseWriter, req *http.Request) {
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "400: could not read body", 400)
    return
  }
  var sensorReq SensorDataRequest
  err = json.Unmarshal(body, &sensorReq)
  if err != nil {
    http.Error(w, fmt.Sprintf("400: could not decode json body: %+v", err), 400)
    return
  }

  // withMact := req.URL.Query().Get("mact") != "0"
  // if withMact {
  //   withMact = rand.Float32() < 0.5
  // }
  sd, err := api.Samples.Gen(sensorReq.Url, sensorReq.Ua, sensorReq.Abck, sensorReq.Stage, sensorReq.AjIndx)
  if err != nil {
    errmsg := fmt.Sprintf("500: error generating sensor: %+v", err)
    api.Log("%s", errmsg)
    http.Error(w, errmsg, 500)
    return
  }

  sdJsonBytes, err := json.Marshal(SensorDataPayload { SensorData: sd.Serialize() })
  if err != nil {
    errmsg := fmt.Sprintf("500: error generating sensor: %+v", err)
    api.Log("%s", errmsg)
    http.Error(w, errmsg, 500)
    return
  }
  w.Header().Add("Content-Type", "application/json")
  w.Write([]byte(sdJsonBytes))
}

func (api *Api) handle_ua(w http.ResponseWriter, req *http.Request) {
	if len(api.Samples.UserAgents) == 0 {
		w.Write([]byte("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36"))
	} else {
  	w.Write([]byte(api.Samples.UserAgents[rand.Intn(len(api.Samples.UserAgents))]))
  }
}

func (api *Api) handle_ua_desktop(w http.ResponseWriter, req *http.Request) {
  w.Write([]byte(api.Samples.DesktopUserAgents[rand.Intn(len(api.Samples.DesktopUserAgents))]))
}

func (api *Api) handle_task_config(w http.ResponseWriter, req *http.Request) {
  w.Write([]byte(`{
    "slimEndpoint": "__queue/yzysply",
    "rcAction": "yzysply_wr_pageview",
    "rcApiKey": "6Lcoe9wUAAAAAOtgfc4c6rnvgptxiBZwDBX3Tqvl",
    "rcCookie": "xhwUqgFqfW88H50",
    "sensorEndpoints": {
      "www.yeezysupply.com": "https://www.yeezysupply.com/static/d65266c44b5ti178051447d1296404842"
    }
  }`))
}

var (
  cookieDist, _ = json.Marshal(map[string]interface{} {
    "rc_cookie": "xhwUqgFqfW88H50",
    "rc_cookie_vals": []string {

      // "03AGdBq26_q7sLpUh7fWLzBvhk7wW4oryS5x9Nab7ZFT1iToeI5kN2SXOoEH1LcHwdRSG1hqLrUnYZbzmQQJZE44d-e6CauoXv9WmOzgcOa-ATSduudtE9NtQPyw5XS7tiVji1s1D5-QNgYPPWLEOT4ba7ddcwShsMkaFdyqsLuHlrLFhuBfeMh_R1BJ9EexI--zjrd0x_yAjC6oUYhuUXz_93-pGS8V4bcWCtRrf2MrWlr_9TxlH7oNscJ6ceDVSqAP3nJwCnbYdLci0YcalJy2mCeXkN3xoo7nSq_bWYCYLUoHxd4Jt9a7o9UC0gNZfqSxYdjNHlVlG_fKCFa5Nnv9cNbs9E5NTKxm-X3O3iYWDZQspCIdRsHhMtM0NOoPZxMPF0WMTbmz2pkQCb-s_q7zxs1RVQCKo2Fw",
      // "03AGdBq24lmaA_YcgFkHoZlnkL6IcC9Gu1nwoHIEw4eeOhLO7ifz6oIePOzuMn2dpcvA2XhGbmgZnXz-zm45Xcu7D8ZgZHzAHUcsEUHqO9stZ-RU_dKqfV4kcTLPqqP6-Drp6WJVZyYymrI8NBycEAxWXWYJc6vqQJMoLgEDj9sshNncmSl-_hQDH2Wwh-ZyjXqoEo9nvG-NOG4TDK2B1aD1irkGV9I7jljHy2VXlcq7B-bzzID_GRBbX1CQ6zFPE2RAI6JwBMzuOzP8v8PGb0oOzxvbPs1pQc3QbZLLUIG5VV9opuY9hwNBy7Wk4priAazkbS7aWkm85KcA8burhjF3TYMnRdLH_lQ_ZpXg4WJLy9A5Ot0TZZJIjEMSLxaiag7W--zdSNlQh8SK-SPxZsUDnx5jzQJamm5Q",
      // "03AGdBq25eYPHOB0APhyY9IwhdVuWTMMdNKqrOz1wnJKnJQbnjwioDNqnsHJTv0jacADfAEZJBFU2nMIBeW5GKjQCPFmLiR7996t1rKs-OIQK8g849uWD_j6OWJPGYepKAw2aG_qZ3tAH5aE0dQCQeDVk2CQEQZYepIempKjnsLV8dCoTcNRdIqOnrEFk9L1uBK9Hs9-05j-affX5e_owhocIGOai5Jq4waIrjVp7tpnX0DtuNJnM3WLwMCFIICM9jHfbdfKEQbGMqqA85Zv_JudWOc-PVWcxHlkzv-MsrIwH2_5Lh3_C2_v-cAQ1bC8nkEeIxvbnX-GaQ9nkjkJd123bwyxBip6GaC58VipPGVT1oPmJdt5H24oTEfCI_jZnjOhISHJcbGc5SUueYa-_gY8VtaChw16fPuQ",
      // "03AGdBq27qCt-0DotQTfoWSx3VTML1Sorz0N2nRZD6gixBG9PPKCST4U02RHDd6glcPPWRGo2o14VIrvL0Xmi_gFaiEqVFNTSSjWRgUsHl2ok9YV-3AZs33jEmG_gPA6cgAZ9swOVnV2TcM2uPzaBjhFt6GilXDgfoH9VwmpZ_sMVlPDSwKsHshVUeLOB23GdQDRAmnfbRTNOO0vqgpj3vChIttvRBYmaF0aB5AV3Py08wupgzP60rKCBO3wrx0HnSX7SFpIvAZZnE3QE3XRLr23lbrfGLHwtXI7I2j3IqQertqTLQc-8GeMXzljUJUCfECFZwF7_WJagTTHgYeWW_DV3_MMR5pKPU-QF2SXV_9YCiXHYd2fkl1ikpeEhuXqorBkeTYqBNEbD4QdIctPokObKvVuKIPNwZEA",
      // "03AGdBq26sKyH9AiDXTYds0ixhBL17XyMlnExgMuRH7EG0RmyofzEg8EBdic9KFdGDfCH0HtAJ1F4-gjusibEKUd4fsLdL14A6NjJM2mLEFktKtHbB5Df6StiJGNNk7JQpnam4ATluEdIpabv0iivSAe7tVKvulCA9qMzhrp8KvXONBrqO_Y1g4VBtyxz_QIa8Fy5A4culDyILS267zncrQ6sft_lv-8hoXC1UUoTi-NVJUz9f8dr96H5g2EdMKYBKbPlt0ratdHAQteOSW1wzBwLbFoUTubFg14f9AVR2a958f562cnGpbIkItLKACW06BIiDiK2v6iOmK24IafDn5XlEgvvoWZxQxIU0QbpzMx8_soLpK_MOR7_tRa4uBY-nG8kPETRWBEoX1iktmoTm0asP5Hxu7Jrzkg",
      // "03AGdBq27lLg7V2m7yX5go14sjlyRr0TQxYynvcB7wfQqhbXOYu-FsKPAvk69vzCOi0aPQYKpEN0vAdxOcVHHzS_VzA90O6NUc-QEMqyTUf_idEwmRdFIuoRu_IZaA5Aldq9a5Le86KGusDcYVqT9TE-VVGVLNL-BfRGlrePxe0MWrRt3tAbBxxBKC1nTBXSh1LqmKRn2wKu_qFghCzaWSu30BfMuYTy6fo7c4nr_4g-BDqpbkrtN2sglQ_YfoAaZn3Myn1sCQKtfGiq3XGWl1oYTMiWsEiWaKKI6OR5mw0EcDMQ5eUFAWQJQHBfNpNpaK5xBquVwWClP96DgYeb5UNEjBag6ysJZK41nGLUzu1iqxU_o1sFuhUDRDi64I8H5xhWW9R5nF4gdl0vw_7OyHJ7WIiPGkD_6htQ",
      // "03AGdBq26lk2rQ74oaLUOu80hyq0QxtHFmArJlipFN_YKbx-oSUtU2MhxxseEim0T5torJ0oZDUxxMaYlMn2yyDWlDDYxrdArQLc96FECUwCje0UXuXUlTH3O4pyWmX5ax0nvnAYIEZbdjZjexpIUpHfcmwnbG-j8pbeJMX2drVlqhg1eye9Xd5yecUZp2ITd09vih6OfVZlTxnOE13KbpVFjhayu_zpRVd3r2V4DRTwiX0qWM9TD7x68OXfMu7WEzalS760zrUHoEnvkHp1uYW7-R3nTgWAA8TcXzhY83OM83KN6UZ_eRDCklbj444A4jYAyCd_oU2of4enSTo0dm9XJaRnx5CQInD5BVNVu_1NBxd-Q_6llXzAV1BAVqnwSA0nn5uONBrZmizp-SKCVnM4u8haY4Wt7vHQ",
      // "03AGdBq26ahPeNIe_uAAiGAD0n43mSByJYHGlab3kkMNNEVsgYHALkCsulpJy9IFaCbtweXJaIA9Q0KK0juqPLUFjIQU_qu_aAIXtKr9vR6GOORfQCopcA1h7fbdtKmfQEypQBCpiVKhtjeLq6SxLBe__vA7h7yIhgw3Epi1d6SAw-d6tbfvsnwlMyFJzPhkEM-XtSY8ElKqDYfqEXFlIUChfqrfnUzauhizicZDBhrjlAR4b9pfefzAj2ALjcQTe4E0Fpx-K1SqSKz_yQDZebsfLt6reGHgcZ-2QixUtDtjPJNrMRub_PwORvpeT7Spbq4dKA3skePn8JqbIvY_CzKMwBvIkR_F4Ee1tsVhyoSDzRvWSs93m27AsvJcuXQ6nasQOcXqCIb6zY0b8JNlXrznNmKjJBFRwI-g",
      // "03AGdBq24cYZ8bGQy-ChQKHi6HzezehX5fIyW5ZuRoCDiB89aPHgqyfrJsNJXeWzbAEXnhC9vFwuH4-VHhPAadZI_N-K9ywD0alQ0HTjiiHdrAtK_0gy1-tGnziaCnnZVlY_oVP8bKlA9CkRgU1xCm47DMBpk82mbej2lDvfJD0yj25glb1S7I0_WGqJuSZnBIm3D4Jl8RlJr3JteKu1Zu_LnkVewcvKcefHW-5x91-eEhV_QSEKpQgROcMf9AXhct94bZRHYogpevKl3uXHX3sd3QoUFWCsvgZssIrkOXTZTzjFE1UgOPODCyTyXhqg8AHRoQVTcn3eaN6AtGkIHOqG763w5TCn0gpdueYKnvtfKBm28sURe4qB49RksTQOBg9oXQnJja9rnnjg3Ev_KUNi4HAMavu7iMaw",

      // "THE_ONE_TRUE_RECAP_COOKIE",
      // "THE_OTHER_TRUE_RECAP_COOKIE",
      // TODO used shared data store like redis?
    },
  })
)

func timeMillis() uint64 {
  return uint64(time.Now().UnixNano() / int64(time.Millisecond))
}

func (api *Api) handle_cookie_distribute(w http.ResponseWriter, req *http.Request) {
  resp, _ := http.Get("https://rushaio.s3.amazonaws.com/cookie.txt")
  if resp != nil && resp.Body != nil {
    bodyBytes, _ := ioutil.ReadAll(resp.Body)
    vals := []string{}
    body := strings.TrimSpace(string(bodyBytes))
    if len(body) > 0 {
      vals = []string {
        body,
      }
    }
    cookieDist, _ = json.Marshal(map[string]interface{} {
      "rc_cookie": "xhwUqgFqfW88H50",
      "rc_cookie_vals": vals,
    })

    w.Write(cookieDist)
  } else {
    http.Error(w, "500", 500)
  }
}

func (api *Api) handle_ay(w http.ResponseWriter, req *http.Request) {
  resp, _ := http.Get("https://rushaio.s3.amazonaws.com/ay.txt")
  if resp != nil && resp.Body != nil {
    bodyBytes, _ := ioutil.ReadAll(resp.Body)
    body := strings.TrimSpace(string(bodyBytes))
    lines := strings.Split(body, "\n")
    fp := lines[rand.Intn(len(lines))]
    w.Write([]byte(fp))
  } else {
    http.Error(w, "500", 500)
  }
}

func (api *Api) LoadAdyenFps(path string) error {
  f, err := os.Open(path)
  if err != nil { return err }
  defer f.Close()
  fpjson, err := ioutil.ReadAll(f)
  if err != nil { return err }
  fps := [][2]string{}
  err = json.Unmarshal(fpjson, &fps)
  if err != nil { return err }
  for _, fp := range fps[:] {
    api.AdyenFps = append(api.AdyenFps, fp[0])
  }
  return nil
}

func validateClaims(claims jwt.MapClaims) bool {
  return claims["dark"] != 1.0 && claims["polo"] != "1"
}

func (api *Api) AuthHandler(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if api.Debug {
      next.ServeHTTP(w, r)
      return
    }
    auth := r.Header.Get("authorization")
    if len(auth) < 64 {
      http.Error(w, "401", 401)
      return
    }
    token, err := jwt.Parse(auth[4:], func(token *jwt.Token) (interface{}, error) {
      if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
      }

      return JWT_SECRET, nil
    })

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid && validateClaims(claims) {
      uid := "none"
      for claim, value := range claims {
        if r.URL.Path == "/" && claim == "rush_id" && r.Header.Get("x-version") == "" {
          http.Error(w, "401", 401)
          return
        }
        if claim == "rush_id" {
          r.Header.Set("x-rush-uid", value.(string))
        }
        if claim == "rush_id" || claim != "exp" {
          if vstr, ok := value.(string); ok {
            uid = vstr
          } else {
            uid = claim
          }
          break
        }
      }


      key := fmt.Sprintf("reqcount:%s:%s:%s", uid, r.URL.Path, time.Now().UTC().Format("2006-01-02--15-04"))
      // TODO per-ip tracking, # IPs per uid tracking
      go api.LogRedis.Incr(context.Background(), key)

      ipkey := fmt.Sprintf("reqcountip:%s:%s:%s:%s", uid, r.Header.Get("X-Forwarded-For"), r.URL.Path, time.Now().UTC().Format("2006-01-02--15-04"))
      go api.LogRedis.Incr(context.Background(), ipkey)

      next.ServeHTTP(w, r)
    } else {
      fmt.Printf("auth_failed err=%+v claims=%+v\n", err, claims)
      http.Error(w, "401", 401)
      return
    }
  })
}


func (api *Api) EnforceMaxClientsHandler(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    api.MaxClientEnforcerSemaphore <- struct{}{}
    defer func() { <-api.MaxClientEnforcerSemaphore }()

    next.ServeHTTP(w, r)
  })
}


func (api *Api) LaunchBaseGenServ() error {
  denoAbspath, err := exec.LookPath("deno")
  if err != nil {
    return err
  }
  cmd := exec.Command(denoAbspath, "run", "--allow-read", "--allow-write", "--unstable", "--allow-net", "./serv.js")
  cmd.Env = append(cmd.Env, "AK_SOCK=./ak.sock")
  cmd.Dir = "../akgen/js"
  // cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  go func() {
    var err error
    defer func() {
      log.Printf("recover %v", recover())
      log.Printf("base gen serv died %+v", err)
    }()
    var stdout io.Reader
    stdout, err = cmd.StdoutPipe()
    if err != nil {
      return
    }
    err = cmd.Start()
    if err != nil {
      return
    }
    b := make([]byte, 64)
    stdout.Read(b)
    b = bytes.Trim(b, "\x00")
    st := strings.TrimSpace(string(b))
    ps := strings.TrimSpace(strings.SplitN(st, ":", 2)[1])

    port, err := strconv.Atoi(ps)
    log.Printf("%+v %+v", port, err)
    defer func() {
      api.Samples.PortMut.Lock()
      defer api.Samples.PortMut.Unlock()
      pp := []string{}
      for _, sport := range api.Samples.Ports {
        if sport != ps {
          pp = append(pp, sport)
        }
      }
      api.Samples.Ports = pp
      pl := len(pp)
      api.Log("ports=%v", api.Samples.Ports)
      if pl == 0 {
        log.Fatalf("All gen servers died")
      }
    }()
    if err == nil {
      api.Log("port=%v", port)
      api.Samples.PortMut.Lock()
      api.Samples.Ports = append(api.Samples.Ports, fmt.Sprintf("%d", port))
      api.Log("ports=%v", api.Samples.Ports)
      api.Samples.PortMut.Unlock()
    }
    io.Copy(ioutil.Discard, stdout)
    err = cmd.Wait()
    api.Log("exit %v", err)
  }()

  return nil
}

type CreateCapTaskRequest struct {
  ClientKey string `json:"clientKey"`
  Task      struct {
    Type       string `json:"type"`
    WebsiteURL string `json:"websiteURL"`
    WebsiteKey string `json:"websiteKey"`
  } `json:"task"`
  SoftID       int    `json:"softId"`
  LanguagePool string `json:"languagePool"`
}

type CreateCapTaskResponse struct {
  ErrorCode int `json:"errorCode"`
  ErrorDescription string `json:"errorDescription"`
  ErrorID int `json:"errorId"`
  TaskID  json.Number `json:"taskId"`
}

type CapTaskStatusRequest struct {
  ClientKey string `json:"clientKey"`
  SiteKey string `json:"siteKey"`
  TaskID json.Number `json:"taskId"`
}

type CapSolution struct {
  ErrorID  int    `json:"errorId"`
  ErrorCode int `json:"errorCode"`
  ErrorDescription string `json:"errorDescription"`
  Status   string `json:"status"`
  Solution struct {
    Text               string `json:"text"`
    GRecaptchaResponse string `json:"gRecaptchaResponse"`
  } `json:"solution"`
  Cost       string `json:"cost"`
  IP         string `json:"ip"`
  CreateTime int    `json:"createTime"`
  EndTime    int    `json:"endTime"`
  SolveCount string `json:"solveCount"`
}

func (api *Api) reloadConfig() {
  for {
    if api.Ctx.Err() != nil {
      return
    }
    resp, err := http.Get("https://rushaio.s3.amazonaws.com/xehosts.txt")
    if resp != nil && resp.Body != nil {
      bodyBytes, _ := ioutil.ReadAll(resp.Body)
      body := strings.TrimSpace(string(bodyBytes))
      // api.Log("%s", body)
      hosts := []string{}
      api.ConfigMut.Lock()
      api.CapHosts = nil
      api.ConfigMut.Unlock()

      for _, host := range strings.Split(body, "\n") {
        go func() {
          cl := &http.Client{Timeout: 5*time.Second}
          creq, _ := http.NewRequestWithContext(api.Ctx, "GET", host + "/", nil)
          if resp, err := cl.Do(creq); err == nil {
            io.Copy(ioutil.Discard, resp.Body)
            resp.Body.Close()
          }

        }()
        hosts = append(hosts, strings.TrimSpace(host))
      }

      api.ConfigMut.Lock()
      if len(hosts) > 0 {
        api.CapHosts = hosts
        // api.CapHosts = []string{"http://86.109.1.138:4006"}
        // api.Log("hosts %+v", api.CapHosts)
        for _, host := range api.CapHosts {
          if _, ok := api.CapHostClients[host]; !ok {
            api.CapHostClients[host] = &http.Client{Timeout: 30*time.Second}
          }
        }
      }
      api.ConfigMut.Unlock()
    } else {
      api.Log("%+v", err)
    }
    // api.Log("%+v %+v", api.CapHosts, api.CapHostClients)
    time.Sleep(30*time.Second)
  }
}

var ErrFailedCreateTask = errors.New("fct")

func (api *Api) handle_cap_create_task(w http.ResponseWriter, req *http.Request) {
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "400: could not read body", 400)
    return
  }
  var capreq CreateCapTaskRequest
  err = json.Unmarshal(body, &capreq)
  if err != nil {
    http.Error(w, "400: could not parse body", 400)
    return
  }
  api.Log("%+v", capreq)


  api.ConfigMut.Lock()
  rand.Seed(time.Now().UnixNano())
  	if len(api.CapHosts) == 0 {
    api.ConfigMut.Unlock()
    http.Error(w, "503 Service Unavailable", 503)
    return
  }
    // nh := len(api.CapHosts)
    idx := rand.Intn(len(api.CapHosts))
    api.Log("%+v", idx)
    host := api.CapHosts[idx]
    api.Log("%+v", host)
    client := api.CapHostClients[host]
    api.Log("%+v", client)
    api.ConfigMut.Unlock()
    api.Log("%+v", idx, host, client)



  capReqMinLimit := api.CapReqMinLimit
  userCapReqMinLimit := 25
  v := api.TaskRedis.Do(req.Context(), "GET", "config:user_cap_req_min_limit")
  if err = v.Err(); err == nil {
    if d, ok := v.Val().(string); ok {
      if n, err := strconv.Atoi(d); err == nil {
        userCapReqMinLimit = n
      }
    }
  }
  v = api.TaskRedis.Do(req.Context(), "GET", "config:global_cap_req_min_limit")
  if err = v.Err(); err == nil {
    if d, ok := v.Val().(string); ok {
      if n, err := strconv.Atoi(d); err == nil {
        capReqMinLimit = n
      }
    }
  }

  cm := int(time.Now().Unix() / 30)
  dt := int(time.Now().Unix() % 30)
  prevmin := 0


  grk := fmt.Sprintf("dd_cap_reqs_%s_%d", "global", cm-1)
  v = api.TaskRedis.Do(req.Context(), "GET", grk)
  // log.Printf("%+v %+v", v, v.Err())
  if err := v.Err(); err == nil {
    if reqmin, ok := v.Val().(string); ok {
      // log.Printf("%+v %+v", reqmin, int64(nh * api.CapReqMinLimit))
      if rr, err := strconv.Atoi(reqmin); err == nil {
        prevmin = rr
      }
    }
  }

  grk = fmt.Sprintf("dd_cap_reqs_%s_%d", "global", cm)
  v = api.TaskRedis.Do(req.Context(), "GET", grk)
  // log.Printf("%+v %+v", v, v.Err())
  if err := v.Err(); err == nil {
    if reqmin, ok := v.Val().(string); ok {
      // log.Printf("%+v %+v", reqmin, int64(nh * api.CapReqMinLimit))
      if rr, err := strconv.Atoi(reqmin); err == nil {
        api.Log("rate: %d", prevmin*((30-dt)/30) + rr)
        if (prevmin*((30-dt)/30) + rr) > capReqMinLimit {
          api.Log("Above request per minute limit")
          http.Error(w, "503 Service Unavailable", 503)
          return
        }
      }
    }
  }

  rk := fmt.Sprintf("dd_cap_reqs_%s_%d", req.Header.Get("uid"), cm-1)
  v = api.TaskRedis.Do(req.Context(), "GET", rk)
  // log.Printf("%+v %+v", v, v.Err())
  if err := v.Err(); err == nil {
    if reqmin, ok := v.Val().(string); ok {
      // log.Printf("%+v %+v", reqmin, int64(nh * api.CapReqMinLimit))
      if rr, err := strconv.Atoi(reqmin); err == nil {
        prevmin = rr
      }
    }
  }

  rk = fmt.Sprintf("dd_cap_reqs_%s_%d", req.Header.Get("uid"), cm)
  v = api.TaskRedis.Do(req.Context(), "GET", rk)
  // log.Printf("%+v %+v", v, v.Err())
  if err := v.Err(); err == nil {
    if reqmin, ok := v.Val().(string); ok {
      // log.Printf("%+v %+v", reqmin, int64(nh * api.CapReqMinLimit))
      if rr, err := strconv.Atoi(reqmin); err == nil {
        api.Log("rate: %d", prevmin*((30-dt)/30) + rr)
        if (prevmin*((30-dt)/30) + rr) > userCapReqMinLimit {
          api.Log("Above request per minute limit")
          http.Error(w, "503 Service Unavailable", 503)
          return
        }
      }
    }
  }
  api.TaskRedis.Do(context.Background(), "INCR", rk)
  api.TaskRedis.Do(context.Background(), "EXPIRE", rk, 120)
  api.TaskRedis.Do(context.Background(), "INCR", grk)
  api.TaskRedis.Do(context.Background(), "EXPIRE", grk, 120)


  // TODO enforce creation limits based on reasonable rps to each xevil instance
  // CapHostsStats

  go func() {
    // var err error
    // defer func() {
    //  api.Log("%+v", err)
    // }()

    ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
    defer cancel()

    // TODO retry till slot available
    api.Log("%s", host + "/createTask")
    api.Log("%s", string(body))
    api.Log("%s", host + "/createTask")
    creq, _ := http.NewRequestWithContext(ctx, "POST", host + "/createTask", nil)
    creq.Close = true
    // creq.Body =ioutil.NopCloser(bytes.NewReader(body))
    creq.ContentLength = int64(len(body))
    creq.Header = make(http.Header)
    creq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    // bbuf := &bytes.Buffer{}
    // creq.Write(bbuf)
    // api.Log("%s", "REQ")
    // api.Log("%s", string(bbuf.Bytes()))

    retry.Do(func() error {
      if ctx.Err() != nil {
        return nil
      }
      creq.Body = ioutil.NopCloser(bytes.NewReader(body))
      resp, err := client.Do(creq)
      if err != nil {
        return err
      }
      api.Log("%s", "readCreateTaskBody")
      var rb []byte
      if rb, err = ioutil.ReadAll(resp.Body); err != nil {
        return err
      }
      resp.Body.Close()
      api.Log("createResp: %s", string(rb))
      var createResp CreateCapTaskResponse
      if err = json.Unmarshal(rb, &createResp); err == nil {
        if createResp.TaskID == json.Number("0") {
          return ErrFailedCreateTask
        }
        rk = fmt.Sprintf("xe_cap_ids_%s_%v_%d", host, int(time.Now().Unix() / 900), createResp.TaskID)
        v = api.TaskRedis.Do(ctx, "GETSET", rk, string(createResp.TaskID))
        log.Printf("%+v %+v", v, v.Err())
        if err := v.Err(); err == nil {
          if _, ok := v.Val().(string); ok {
            api.Log("WARN: dupe")
            return ErrFailedCreateTask
          }
        }
        api.TaskRedis.Do(context.Background(), "EXPIRE", rk, 900)


        time.Sleep(10*time.Second)
        api.Log("%s", "statusReq", string(rb))
        statusReq := CapTaskStatusRequest{TaskID: createResp.TaskID}
        srb, _ := json.Marshal(statusReq)
        for {
          if ctx.Err() != nil {
            return nil
          }
          // api.Log("%s", "getTaskResult")
          creq, _ = http.NewRequestWithContext(ctx, "POST", host + "/getTaskResult", ioutil.NopCloser(bytes.NewReader(srb)))
          creq.ContentLength = int64(len(srb))
          creq.Header = make(http.Header)
          creq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
          if resp, err = client.Do(creq); err == nil {
            if rb, err = ioutil.ReadAll(resp.Body); err == nil {
              resp.Body.Close()
              var soln CapSolution
              api.Log("taskResult %s", string(rb))
              if err = json.Unmarshal(rb, &soln); err == nil {
                if soln.ErrorCode > 0 {
                  // api.TaskRedis.Do(context.Background(), "INCR", rk)
                  return ErrFailedCreateTask
                }
                if soln.Status == "ready" && !strings.Contains(soln.Solution.Text, "ERROR") && len(soln.Solution.GRecaptchaResponse) > 32 {
                  soln.EndTime = int(time.Now().Unix())
                  sbbt, _ := json.Marshal(soln)
                  k := fmt.Sprintf("dd_cap_sols_%s_%s_%d", req.Header.Get("uid"), capreq.Task.WebsiteKey, int(time.Now().Unix() / 120))
                  api.Log("pushing %s %s", k, string(sbbt))
                  api.TaskRedis.Do(context.Background(), "LPUSH", k, string(sbbt))
                  api.TaskRedis.Do(context.Background(), "EXPIRE", k, 120)
                  return nil
                }
              }
              api.Log("%+v %+v", soln, err)
            }
          }

          time.Sleep(time.Duration(5+rand.Intn(10))*time.Second)
          // time.Sleep(time.Duration(2*time.Second))
        }
      }
      return ErrFailedCreateTask
    }, retry.Attempts(8), retry.MaxDelay(16*time.Second))
  }()

  bb, _ := json.Marshal(CreateCapTaskResponse{
    TaskID: json.Number("420"),
  })
  fmt.Fprintf(w, string(bb))
}

func (api *Api) handle_cap_get_result(w http.ResponseWriter, req *http.Request) {
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "400: could not read body", 400)
    return
  }
  var statusreq CapTaskStatusRequest
  json.Unmarshal(body, &statusreq)

  api.Log("%s", string(body))
  var capsolbytes []byte
  var capsol CapSolution
  curslot := int(time.Now().Unix() / 120)
  // prevslot := int(time.Now().Unix() / 120) - 1

  if statusreq.SiteKey == "" {
    statusreq.SiteKey = "6LccSjEUAAAAANCPhaM2c-WiRxCZ5CzsjR_vd8uX"
  }

  // v := api.TaskRedis.Do(req.Context(), "LPOP", fmt.Sprintf("dd_cap_sols_%s_%d", statusreq.SiteKey, prevslot))
  // if err := v.Err(); err == nil {
  //   if capsolss, ok := v.Val().(string); ok {
  //     api.Log("capsolss %+v", capsolss)
  //     capsolbytes = []byte(capsolss)
  //     if err = json.Unmarshal(capsolbytes, &capsol); err == nil {
  //       if time.Now().Unix() - int64(capsol.EndTime) < 90 {
  //         api.Log("got capsol=%+v", capsol)
  //         if capsol.Status == "ready" {
  //           io.WriteString(w, capsolss)
  //           return
  //         }
  //       } else {
  //         api.Log("popped expired capsol=%+v", capsol)
  //       }
  //     }
  //   } else {
  //     fmt.Printf("could not get str from val")
  //   }
  // }

  api.Log("%s", fmt.Sprintf("dd_cap_sols_%s_%s_%d", req.Header.Get("uid"), statusreq.SiteKey, curslot))
  v := api.TaskRedis.Do(req.Context(), "LPOP", fmt.Sprintf("dd_cap_sols_%s_%s_%d", req.Header.Get("uid"), statusreq.SiteKey, curslot))
  if err = v.Err(); err == nil {
    if capsolss, ok := v.Val().(string); ok {
      api.Log("capsolss %+v", capsolss)
      capsolbytes = []byte(capsolss)
      if err = json.Unmarshal(capsolbytes, &capsol); err == nil {
        if time.Now().Unix() - int64(capsol.EndTime) < 90 {
          api.Log("got capsol=%+v", capsol)
          if capsol.Status == "ready" {
            io.WriteString(w, capsolss)
            return
          }
        } else {
          api.Log("popped expired capsol=%+v", capsol)
        }
      }
    } else {
      fmt.Printf("could not get str from val")
    }
  }

  http.Error(w, "No Result Yet", 404)
}

func (api *Api) handle_hcache(w http.ResponseWriter, req *http.Request) {
  key := req.URL.Query().Get("k")
  if key == "" || len(key) > 255 {
    http.Error(w, "400", 400)
    return
  }

  if req.Method == "GET" {
    o, err := api.minioClient.GetObject(req.Context(), "rushaio", "hcache/" + key, minio.GetObjectOptions{})
    if err != nil {
      ls := fmt.Sprintf("%+v", err)
      api.Log("%s", ls)
      http.Error(w, ls, 404)
      return
    }
    if _, err := io.Copy(w,o); err != nil {
      ls := fmt.Sprintf("%+v", err)
      api.Log("%s", ls)
      http.Error(w, ls, 404)
      return
    }
  } else if req.Method == "POST" {
    cl, _ := strconv.Atoi(req.Header.Get("Content-Length"))
    if cl > 0 {
      inf, err := api.minioClient.PutObject(req.Context(), "rushaio", "hcache/" + key, req.Body, int64(cl), minio.PutObjectOptions{ContentType: "text/plain"})
      api.Log("%+v %+v", inf, err)
      if err != nil {
        w.Write([]byte("Success"))
      } else {
        w.Write([]byte(fmt.Sprintf("%+v", err)))
      }
    } else {
      http.Error(w, "No Data", 400)
    }
  }
}

func (api *Api) handle_skudb(w http.ResponseWriter, req *http.Request) {
  ps := strings.Split(req.URL.Path, "/")
  pid := ps[len(ps)-1]
  objname := fmt.Sprintf("hibbett-mobileapi.prolific.io/ecommerce/products/%s", pid)
  if req.Method == "GET" {
    o, err := api.minioClient.GetObject(req.Context(), "rushaio", objname, minio.GetObjectOptions{})
    if err != nil {
      ls := fmt.Sprintf("%+v", err)
      api.Log("%s", ls)
      http.Error(w, ls, 404)
      return
    }
    if _, err := io.Copy(w,o); err != nil {
      ls := fmt.Sprintf("%+v", err)
      api.Log("%s", ls)
      http.Error(w, ls, 404)
      return
    }


    // req, err := http.NewRequest("GET", urll, nil)
    // resp, err := api.Client.Do(req)
    // if err != nil {
    //  ls := fmt.Sprintf("%+v", err)
    //  api.Log("%s", ls)
    //  http.Error(w, ls, 500)
    //  return
    // }
    // defer resp.Body.Close()
    // o, err := ioutil.ReadAll(resp.Body)
    // if err != nil {
    //  ls := fmt.Sprintf("%+v", err)
    //  api.Log("%s", ls)
    //  http.Error(w, ls, 500)
    //  return
    // }
    // if resp.StatusCode != 200 {
    //  http.Error(w, string(o), 500)
    // } else {
    //  w.Write(o)
    // }
  } else if req.Method == "POST" {
    cl, _ := strconv.Atoi(req.Header.Get("Content-Length"))
    if cl > 0 {
      inf, err := api.minioClient.PutObject(req.Context(), "rushaio", objname, req.Body, int64(cl), minio.PutObjectOptions{ContentType: "application/json"})
      api.Log("%+v %+v", inf, err)
      if err != nil {
        w.Write([]byte("Success"))
      } else {
        w.Write([]byte(fmt.Sprintf("%+v", err)))
      }
    } else {
      http.Error(w, "No Data", 400)
    }
  }
}

func GetEnv(key string, defaultVal string) string {
  if value, exists := os.LookupEnv(key); exists {
    return value
  }

  return defaultVal
}

func main() {
  rand.Seed(time.Now().UnixNano())

  port := os.Getenv("PORT")
  if port == "" {
    port = "8081"
  }
  addr := ":" + port

  sema := make(chan struct{}, math.MaxInt32)
  if os.Getenv("MAX_CLIENTS") != "" {
    if maxClients, err := strconv.Atoi(os.Getenv("MAX_CLIENTS")); err == nil {
      sema = make(chan struct{}, maxClients)
    }
  }

  // Get the AWS credentials
  endpoint := "s3.amazonaws.com"
  accessKeyID := "AKIAXKSACTYP7TUIQQN4"
  secretAccessKey := "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"
  useSSL := true
  // Initialize minio client object.
  minioClient, err := minio.New(endpoint, &minio.Options{
    Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
    Secure: useSSL,
  })
  if err != nil {
    log.Fatalln(err)
  }

  REDIS_HOST := GetEnv("REDIS_HOST", "127.0.0.1")
  REDIS_PORT := GetEnv("REDIS_PORT", "6379")
  REDIS_PW := GetEnv("REDIS_PW", "")
  REDIS_DB_STR := GetEnv("REDIS_DB", "0")
  REDIS_DB, _ := strconv.Atoi(REDIS_DB_STR)

  logRedis := redis.NewClient(&redis.Options{
    Addr:     fmt.Sprintf("%s:%s", REDIS_HOST, REDIS_PORT),
    Password: REDIS_PW, // no password set
    DB:       REDIS_DB,  // use default DB
  }).WithTimeout(5*time.Second)

  TASK_REDIS_HOST := GetEnv("TASK_REDIS_HOST", REDIS_HOST)
  TASK_REDIS_PORT := GetEnv("TASK_REDIS_PORT", REDIS_PORT)
  TASK_REDIS_PW := GetEnv("TASK_REDIS_PW", REDIS_PW)
  TASK_REDIS_DB_STR := GetEnv("TASK_REDIS_DB", REDIS_DB_STR)
  TASK_REDIS_DB, _ := strconv.Atoi(TASK_REDIS_DB_STR)

  taskRedis := redis.NewClient(&redis.Options{
    Addr:     fmt.Sprintf("%s:%s", TASK_REDIS_HOST, TASK_REDIS_PORT),
    Password: TASK_REDIS_PW, // no password set
    DB:       TASK_REDIS_DB,  // use default DB
  }).WithTimeout(5*time.Second)

  capReqMinLimit := 500
  limenv := os.Getenv("CAP_REQ_MIN_LIMIT")
  if ll, err := strconv.Atoi(limenv); err == nil {
    capReqMinLimit =ll
  }
  if s, err := akgen.ParseAkWebSensor("7a74G7m23Vrp0o5c9115221.66-1,2,-94,-100,Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:83.0) Gecko/20100101 Firefox/83.0,uaend,11059,20100101,en-US,Gecko,0,0,0,0,395334,2327765,1920,1177,1920,1200,1920,660,1920,,cpen:0,i1:0,dm:0,cwen:0,non:1,opc:0,fc:1,sc:0,wrc:1,isc:97,vib:1,bat:0,x11:0,x12:1,6006,0.692208787346,803371163882,0,loc:-1,2,-94,-101,do_en,dm_en,t_dis-1,2,-94,-105,-1,2,-94,-102,-1,2,-94,-108,-1,2,-94,-110,-1,2,-94,-117,-1,2,-94,-111,-1,2,-94,-109,-1,2,-94,-114,-1,2,-94,-103,-1,2,-94,-112,https://www.yeezysupply.com/product/F36640-1,2,-94,-115,1,32,32,0,0,0,0,347,0,1606742327764,7,17188,0,0,2864,0,0,355,0,0,1034D724EC79251935E65755E21A2BEC~-1~YAAQhUZ7aCv8DJV1AQAAkEJOGQQp+syNgnn0+TFKMxijr+GyqdAr6ShQsfnz79Jkfzzf+D1ze7bh8lVVNey4s5j3wWdgvoj2A0WwOtlfAD4sGbLi8l4x8cJLKieky8iBKrbXX7cpsqUvHHEgK0GzEMPa31l30vsX/pFW+uhLdYhho/B0tVl4oG3Wc37RKFdvo3pn/FH1+eUdQt7YCoMVvPSTT9jFAz2PK7D/Kn0xz4W/bhRDVX1imfo1XJ50ZtpOGRULYVxMPOlZ2Y/YLecW3Wk3YI17LEmbdoM/zyh355kLbsWbWDVchTrOy5IqkJY3uIS7MzhDYwuocBYKKa6oywz4yoBwjpz0DfSNYKqQBlyOIJLicyVQBiPEYCIdP/MlLeNE7Kb9CVvUp05SGCJgIWl2DQ5R1RFq7ISqayezgtS/130qfn6trmodADUsfk/X7Ix7zs1zMKdIvSeYbCzScoKPsJITXAK98rRHLUP3~-1~||1-TDkikpBRbF-1-10-1000-2||~-1,46430,412,-1869995902,26067385,PiZtE,37212,106-1,2,-94,-106,8,1-1,2,-94,-119,-1-1,2,-94,-122,0,0,0,0,1,0,0-1,2,-94,-123,-1,2,-94,-124,0.e7aebf5de65398,0.4ecf520b4e6db,0.3514f2cfc1729,0.0c1a9c0072ecb,0.0b6ccfe77d3ec,0.425450471612f8,0.556df6850d4f,0.b2965532fdbb7,0.ea0de1e712a83,0.57c892e2d6cfe8;1,1,0,1,0,0,0,0,0,0;0,3,1,4,4,9,3,1,6,17;1034D724EC79251935E65755E21A2BEC,1606742327764,TDkikpBRbF,1034D724EC79251935E65755E21A2BEC1606742327764TDkikpBRbF,1,1,0.e7aebf5de65398,1034D724EC79251935E65755E21A2BEC1606742327764TDkikpBRbF10.e7aebf5de65398,239,172,48,228,96,239,36,120,112,97,66,172,0,205,12,92,248,144,3,44,55,58,28,224,139,124,216,126,78,209,19,123,263,0,1606742328111;-1,2,-94,-126,-1,2,-94,-127,11133333331333333333-1,2,-94,-70,-1799924720;1301633122;dis;;true;true;true;480;true;24;24;true;false;unspecified-1,2,-94,-80,6454-1,2,-94,-116,6983235-1,2,-94,-118,125149-1,2,-94,-129,31fe9217394afb1878f28b2db8c842e613e3efda8ffcece7fbe91439429e1a82,1,cd15889e4b58585ec9f3c796725c752f6dd7926965daec54124da062a5aaf8e1,,,,0-1,2,-94,-121,;13;6;0"); err == nil {
    log.Printf("%+v", s)
  }
  api := Api{
    CapHostClients: map[string]*http.Client{},
    Samples: &akgen.Samples{},
    LogRedis: logRedis,
    CapReqMinLimit: capReqMinLimit,
    TaskRedis: taskRedis,
    ConfigMut: sync.Mutex{},
    AdyenFps: []string{},
    QCkieStore: map[string]*QCkie{},
    Debug: os.Getenv("DEBUG") == "1",
    MaxClientEnforcerSemaphore: sema,
    Client: &http.Client{Timeout: 10*time.Second},
    minioClient: minioClient,
    Ctx: context.Background(),
  }
  if os.Getenv("NOSENSOR") != "1" {
    err = api.Samples.LoadMactSensors("../akgen/fp/mact.gob.gz")
    if err != nil { panic(err) }
    sensorfn := os.Getenv("SENSOR_GOB")
    if sensorfn == "" { sensorfn = "../akgen/fp/sensor/sensors.gob.gz" }
    log.Println(sensorfn)
    err = api.Samples.LoadSensors(sensorfn)
    if err != nil { panic(err) }
  }
  err = api.LoadAdyenFps("./adyen_fp.json")
  if err != nil { panic(err) }

  nc := int(runtime.NumCPU()) - 1
  if nc == 0 {
    nc = 1
  }
  for i := 0; i < nc; i++ {
    err = api.LaunchBaseGenServ()
    if err != nil { panic(err) }
  }

  go api.reloadConfig()

  r := http.NewServeMux()

  r.Handle("/", api.EnforceMaxClientsHandler(api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_root))))))
  r.Handle("/aknew", api.EnforceMaxClientsHandler(api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_root))))))
  r.Handle("/ay", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_ay)))))
  r.Handle("/ua", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_ua)))))
  r.Handle("/uanew", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_ua)))))
  r.Handle("/ua_desktop", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_ua_desktop)))))
  r.Handle("/config", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_task_config)))))
  r.Handle("/ys_rc_cookie", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_cookie_distribute)))))
  r.Handle("/qckie/update", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_put_qckie)))))
  r.Handle("/qckie", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_get_qckie)))))
  r.Handle("/abck", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_get_abck)))))

  r.Handle("/sesh/count", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_count_trusted_session)))))
  r.Handle("/sesh/push", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_push_trusted_session)))))
  r.Handle("/sesh/get", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_get_trusted_session)))))
  r.Handle("/sesh/remove", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_remove_trusted_session)))))
  r.Handle("/sesh/getcap", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_get_trusted_cap_cookie)))))
  r.Handle("/sesh/pushcap", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_push_trusted_cap_cookie)))))

  r.Handle("/cap/createTask", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_cap_create_task)))))
  r.Handle("/cap/getTaskResult", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_cap_get_result)))))

  r.Handle("/test_prox", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_test_prox)))))

  r.Handle("/hcache", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(http.HandlerFunc(api.handle_hcache)))))
  rehandler := &RegexpHandler{}

  re_skudb := regexp.MustCompile(`/skudb/.*`)
  rehandler.HandleFunc(re_skudb, api.handle_skudb)
  r.Handle("/skudb/", api.AuthHandler(handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(rehandler))))

  fmt.Println("listen on", addr)
  if err := http.ListenAndServe(addr, handlers.CompressHandler(r)); err != http.ErrServerClosed {
      fmt.Printf("ListenAndServe(): %v", err)
  }
}


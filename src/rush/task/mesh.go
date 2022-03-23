package task

import (
	"strings"
  "log"
	"encoding/json"
	"rush/net/http"
	"net/url"
  "os/exec"
)

func MeshHawkAuth() string {
  cmd := exec.Command("python3", "testt.py")
  cmd.Dir = "/Users/aka/drive/ysfc/mesh"
  out, _ := cmd.CombinedOutput()
  log.Println(string(out))
  return strings.TrimSpace(string(out))
}

func (t *CheckoutTask) FootpatrolAppAtc() (*http.Response, error) {
	t.HeaderBlacklist = map[string]bool {
		"content-length": true,
		"sec-fetch-site": true,
		"sec-fetch-mode": true,
		"sec-fetch-dest": true,
		"accept-language": true,
	}
	t.HeaderCaseFn = func(s string) string { return s }
	// t.UserAgent = "footpatrolgb/2.7.0.7395 (android-app-phone; Android 9; Build/ONEPLUS A5010_43_191008)"
	t.DefaultProto = "HTTP/1.1"
  url_, err := url.Parse("https://prod.jdgroupmesh.cloud/stores/footpatrolgb/carts")
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "X-NewRelic-ID",
    "x-api-key",
    "MESH-Commerce-Channel",
    "mesh-version",
    "User-Agent",
    "X-Request-Auth",
    "X-acf-sensor-data",
    "Content-Type",
    "Accept",
    "Transfer-Encoding",
    "Host",
    "Connection",
    "Accept-Encoding",
  }
  s, _ := t.getSensor()
  headers := [][2]string {
    {"X-NewRelic-ID", "VQYDUFVWDRABV1BbDwUDUlc="},
    {"x-api-key", "6F6A58E7640E44BBA383479E19658DFD"},
    {"MESH-Commerce-Channel", "android-app-phone"},
    {"mesh-version", "cart=4"},
    {"User-Agent", "footpatrolgb/2.7.0.7395 (android-app-phone; Android 9; Build/ONEPLUS A5010_43_191008)"},
    {"X-Request-Auth", MeshHawkAuth()},
    {"X-acf-sensor-data", s},
    {"Content-Type", "application/json; charset=UTF-8"},
    {"Accept", "application/json"},
    {"Host", url_.Host},
    {"Connection", "Keep-Alive"},
    {"Accept-Encoding", "gzip"},
  }
  bodyStructure := map[string]interface{} {
  "channel": "android-app-phone",
  "contents": []interface{} {
    map[string]interface{} {
          "SKU": "371804_footpatrolcom.001717067",
          "quantity": 1,
          "type": "cartProduct",
        },
  },
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  req := t.makeReq("POST", url_, &headers, &headerOrder, &body)
  req.Proto = "HTTP/1.1"
  // req.ContentLength = 0
  // req.TransferEncoding = []string{"chunked"}
  req.ContentLength = -1
  return t.doReq(t.client, req)
}


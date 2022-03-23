package task

import (
	"fmt"
	"strconv"
	"bytes"
	"context"
	"encoding/json"
	"rush/net/http"
	"sync"
	"log"
	"time"
	"github.com/avast/retry-go"
	"os"
)

type CloudLogger struct {
	Context context.Context
	CtxData map[string]interface{}
	Cancel context.CancelFunc
	client *http.Client
	mchan chan map[string]interface{}
	backpressure uint64
	bpmut sync.Mutex
}

func (c *CloudLogger) Init(ctx context.Context) {
	c.Context, c.Cancel = context.WithCancel(ctx)
	c.client = &http.Client{Timeout: 5*time.Second}
	c.mchan = make(chan map[string]interface{}, BUF_SIZE*2)
	c.CtxData = map[string]interface{}{}
	go c.Shipper()
}

var CL_MAX_WAITING uint64 = 5e3

func (c *CloudLogger) SendTelemetry(tel map[string]interface{}) {
	defer recover()
	// log.Printf("SendTelemetry %+v", tel)
	c.bpmut.Lock()
	if c.backpressure > CL_MAX_WAITING {
		c.bpmut.Unlock()
		// todo dbg log
		return
	}
	c.backpressure += 1
	c.bpmut.Unlock()

	c.mchan <- tel

	c.bpmut.Lock()
	c.backpressure -= 1
	c.bpmut.Unlock()
}

func (c *CloudLogger) SendTelemetryNow(tel map[string]interface{}) {
	// log.Printf("SendTelemetry %+v", tel)
	defer recover()
	c.DoSendTelemetry([]map[string]interface{} { tel })
}

var BUF_SIZE = 512
func (c *CloudLogger) Shipper() {
	flushInterval := 10*time.Second
	if bufFlushIntervalStr := os.Getenv("CLOUDLOG_INTERVAL"); len(bufFlushIntervalStr) > 0 {
		if n, err := strconv.Atoi(bufFlushIntervalStr); err == nil {
			flushInterval = time.Duration(n)*time.Second
		}
	}
	log.Printf("%+v", flushInterval)
	var buf  = make([]map[string]interface{}, BUF_SIZE)
	var bufn = 0
	defer recover()
	for {
		select{
			case <-c.Context.Done():
				// TODO pop msgs off mchan before ending loop
				go c.DoSendTelemetry(buf[:bufn])
				return
			case tel := <-c.mchan:
				if bufn < BUF_SIZE {
					buf[bufn] = tel
					bufn += 1
				}
				if bufn >= BUF_SIZE {
					c.DoSendTelemetry(buf)
					buf = nil
					buf = make([]map[string]interface{}, BUF_SIZE)
					bufn = 0
				}
				break
			case <-time.After(flushInterval):
				if bufn > 0 {
					c.DoSendTelemetry(buf[:bufn])
					buf = nil
					buf = make([]map[string]interface{}, BUF_SIZE)
					bufn = 0
				}
				break
		}
	}
}

func (c *CloudLogger) DoSendTelemetry(tel []map[string]interface{}) {
	defer recover()
	teljs := [][]byte{}
	for _, tt := range tel {
		for key, value := range c.CtxData {
			tt[key] = value
		}
		body, err := json.Marshal(tt)
		if err != nil {
			continue
		}
		teljs = append(teljs, []byte("{ \"index\": {} }"))
		teljs = append(teljs, body)
	}
	if len(teljs) == 0 {
		return
	}
	teljs = append(teljs, []byte{})
	tellines := bytes.Join(teljs, []byte("\n"))
	// log.Println(string(tellines))
	retry.Do(func() error {
		req, _ := http.NewRequest("POST", fmt.Sprintf("https://zzzzz.us-east-1.aws.found.io:9243/rush-%s/_doc/_bulk", time.Now().Format("20060102")), bytes.NewReader(tellines))
		req = req.WithContext(c.Context)
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("rushbeta", "rushaio")
		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		// if t.Debug {
			if os.Getenv("LOG") == "1" {
				b,_ := readBody(resp)
				log.Println(b)
			} else {
		// } else {
				DiscardResp(resp)
			}
		// }
		return nil
	}, retry.Attempts(3), retry.Delay(250*time.Millisecond))
}

var DefaultCloudLogger = CloudLogger{}
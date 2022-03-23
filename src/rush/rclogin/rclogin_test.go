package main


// package main

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"github.com/chromedp/chromedp"
// 	"io/ioutil"
// 	"log"
// 	"os"
// 	"os/exec"
// 	"bufio"
// 	"testing"
// 	"net/http"
// 	"time"
// 	"github.com/stretchr/testify/assert"
// )

// func startServer(ctx context.Context, t *testing.T) {
// 	assert := assert.New(t)
// 	cmd := exec.CommandContext(ctx, "go", "run", ".")
// 	err := cmd.Start()
// 	assert.Nil(err)

// 	// wait for server to start
// 	client := &http.Client{}
// 	waitForServerStart := func() {
// 		timeout := time.After(3 * time.Second)
// 		tick := time.Tick(500 * time.Millisecond)
// 		for {
// 			select {
// 			case <-timeout:
// 				assert.FailNow("timed out waiting for server start")
// 			case <-tick:
// 				_, err := client.Get("http://localhost:42424/health")
// 				if err == nil {
// 					return
// 				} else {
// 					log.Printf("%+v", err)
// 				}
// 			}
// 		}
// 	}
// 	waitForServerStart()
// }

// func TestGetToken(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	assert := assert.New(t)
// 	startServer(ctx, t)

// 	body := TokenRequest{
// 		Url: "https://www.yeezysupply.com/",
// 		ApiKey: "6Lcoe9wUAAAAAOtgfc4c6rnvgptxiBZwDBX3Tqvl",
// 		Action: "yeezy_supply_wr_pageview",
// 		Proxy: "",
// 	}
// 	bodyJson, err := json.Marshal(body)
// 	assert.Nil(err)


// 	client := &http.Client{}
// 	resp, err := client.Post("http://localhost:42424/", "application/json", ioutil.NopCloser(bytes.NewReader(bodyJson)))
// 	assert.Nil(err)
// 	respBytes, err := ioutil.ReadAll(resp.Body)
// 	assert.Nil(err)

// 	assert.True(bytes.ContainsAny(respBytes, "-_"))
// 	assert.Greater(len(respBytes), 10)
// }

// func TestBrowserProxy(t *testing.T) {
// 	assert := assert.New(t)

// 	client := &http.Client{}
// 	resp, err := client.Get("https://ipecho.net/plain")
// 	assert.Nil(err)
// 	respBytes, err := ioutil.ReadAll(resp.Body)
// 	assert.Nil(err)
// 	hostPublicIp := string(respBytes[:])

// 	rc, err := NewRCContext()
// 	assert.Nil(err)
// 	defer rc.cancel()
// 	rc.SetUpstreamProxy("http://proxy:zUpxM1S@199.188.89.151:8000")
// 	ctx, cancel := chromedp.NewContext(rc.browserCtx)
// 	defer cancel()

// 	var browserPublicIp string
// 	err = chromedp.Run(ctx,
// 		chromedp.Navigate("https://ipecho.net/plain"),
// 		chromedp.Evaluate("document.body.innerText", &browserPublicIp),
// 	)
// 	assert.Nil(err)
// 	assert.Greater(len(browserPublicIp), 6)
// 	log.Printf("host %s  proxy %s", hostPublicIp, browserPublicIp)
// 	assert.NotEqual(hostPublicIp, browserPublicIp)
// }

// func TestSpoofHeadful(t *testing.T) {
// 	assert := assert.New(t)

// 	rc, err := NewRCContext()
// 	assert.Nil(err)
// 	defer rc.cancel()
// 	ctx, cancel := chromedp.NewContext(rc.browserCtx)
// 	defer cancel()

// 	var testFailures string
// 	err = chromedp.Run(ctx,
// 		SpoofHeadfulAction(),
// 		chromedp.Navigate("https://intoli.com/blog/not-possible-to-block-chrome-headless/chrome-headless-test.html"),
// 		chromedp.Evaluate(`
// 			var failed = Array.from(document.getElementsByClassName("failed"));
// 			failed.map(fail => fail.parentNode.innerText.replace("\n", ":")).join("\n")
// 		`, &testFailures),
// 	)
// 	assert.Equal(testFailures, "")
// }

// type Score struct {
// 	Score float32 `json:"score"`
// }

// func TestGetTokenProxies(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	assert := assert.New(t)

// 	startServer(ctx, t)

// 	file, err := os.Open("/Users/andyh2/drive/ysfc/proxies_3_24.txt")
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer file.Close()

//   client := &http.Client{}

//   scanner := bufio.NewScanner(file)
//   for scanner.Scan() {
//   	proxy := scanner.Text()
// 		body := TokenRequest{
// 			Url: "http://score.rushaio.co/",
// 			ApiKey: "6LcI0-MUAAAAAGdTwaIHO-VfOrpwlaqMNeBpGR1E",
// 			Action: "sc_pageview",
// 			Proxy: proxy,
// 		}
// 		bodyJson, err := json.Marshal(body)
// 		assert.Nil(err)

// 		resp, err := client.Post("http://localhost:42424/", "application/json", ioutil.NopCloser(bytes.NewReader(bodyJson)))
// 		assert.Nil(err)
// 		respBytes, err := ioutil.ReadAll(resp.Body)
// 		assert.Nil(err)

// 		token := map[string]string {
// 			"token": string(respBytes[:]),
// 		}
// 		tokenJson, err := json.Marshal(token)
// 		scoreResp, err := client.Post("http://score.rushaio.co/token", "application/json", ioutil.NopCloser(bytes.NewReader(tokenJson)))
// 		assert.Nil(err)
// 		scoreBytes, err := ioutil.ReadAll(scoreResp.Body)
// 		assert.Nil(err)
// 		var score Score
// 		err = json.Unmarshal(scoreBytes, &score)
// 		assert.Nil(err)
// 		log.Println(string(scoreBytes[:]))
// 		log.Printf("%s %f", proxy, score.Score)
//   }
// }

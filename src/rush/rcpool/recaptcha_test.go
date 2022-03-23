package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/network"
	"io/ioutil"
	"log"
	"runtime"
	"os"
	"os/exec"
	// "bufio"
	"syscall"
	"testing"
	"os/signal"
	"net/http"
	"time"
	"github.com/stretchr/testify/assert"
)

func getCtx() (context.Context, context.CancelFunc) {
  ctx, cancel := context.WithCancel(context.Background())
  c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    // defer func() {
    //   signal.Stop(c)
    //   cancel()
    // }()
    go func() {
      select {
      case <-c:
        cancel()
      case <-ctx.Done():
      }
    }()

  return ctx, cancel
}

func startServer(ctx context.Context, t *testing.T) {
	assert := assert.New(t)
	cmd := exec.Command("go", "run", ".")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := cmd.Start()
	assert.Nil(err)

	go func() {
		<-ctx.Done()

		log.Println("killing...")

		// Windows doesn't support Interrupt
		if runtime.GOOS == "windows" {
			_ = cmd.Process.Signal(os.Kill)
			return
		}

		// go func() {
		// 	time.Sleep(2 * time.Second)
		// }()
		pid := cmd.Process.Pid
		log.Printf("killed %d", pid)
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		// if err := cmd.Process.Kill(); err != nil {
		//     log.Fatal("failed to kill process: ", err)
		// } else {
		// }
	}()

	// wait for server to start
	client := &http.Client{}
	waitForServerStart := func() {
		timeout := time.After(3 * time.Second)
		tick := time.Tick(500 * time.Millisecond)
		for {
			select {
			case <-timeout:
				assert.FailNow("timed out waiting for server start")
			case <-tick:
				_, err := client.Get("http://localhost:42424/health")
				if err == nil {
					return
				} else {
					log.Printf("waiting...%+v", err)
				}
			}
		}
	}
	waitForServerStart()
}

func TestGetToken(t *testing.T) {
	ctx, cancel := getCtx()
	defer cancel()
	assert := assert.New(t)
	startServer(ctx, t)

	body := TokenRequest{
		Url: "https://www.yeezysupply.com/",
		ApiKey: "6Lcoe9wUAAAAAOtgfc4c6rnvgptxiBZwDBX3Tqvl",
		Action: "yeezy_supply_wr_pageview",
		Proxy: "",
	}
	bodyJson, err := json.Marshal(body)
	assert.Nil(err)


	client := &http.Client{}
	resp, err := client.Post("http://localhost:42424/", "application/json", ioutil.NopCloser(bytes.NewReader(bodyJson)))
	assert.Nil(err)
	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)

	assert.True(bytes.ContainsAny(respBytes, "-_"))
	assert.Greater(len(respBytes), 10)
}

func TestBrowserProxy(t *testing.T) {
	assert := assert.New(t)

	client := &http.Client{}
	resp, err := client.Get("https://ipecho.net/plain")
	assert.Nil(err)
	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)
	hostPublicIp := string(respBytes[:])

	rc, err := NewRCContext()
	assert.Nil(err)
	defer rc.cancel()

	err = rc.SetUpstreamProxy("http://proxy:8SZuFA2@199.188.92.231:8000")
	assert.Nil(err)
	ctx, cancel := chromedp.NewContext(rc.browserCtx)
	defer cancel()

	var browserFirstProxyPublicIp string
	err = chromedp.Run(ctx,
		network.ClearBrowserCache(),
		chromedp.Navigate("https://ipecho.net/plain"),
		chromedp.Evaluate("document.body.innerText", &browserFirstProxyPublicIp),
	)
	assert.Nil(err)
	log.Printf("host %s  proxy %s", hostPublicIp, browserFirstProxyPublicIp)
	assert.Greater(len(browserFirstProxyPublicIp), 6)
	assert.NotEqual(hostPublicIp, browserFirstProxyPublicIp)
	cancel()

	err = rc.SetUpstreamProxy("http://835317+153595+US:RushAIOFTW1@us-30m.homeip.io:8000")
	assert.Nil(err)
	ctx, cancel = chromedp.NewContext(rc.browserCtx)
	defer cancel()

	time.Sleep(time.Duration(1)*time.Second)
	log.Println("trying next")
	var browserSecondProxyPublicIp string
	err = chromedp.Run(ctx,
		chromedp.Navigate("https://ipecho.net/plain"),
		// chromedp.WaitVisible("#token"),
		chromedp.Evaluate("document.body.innerText", &browserSecondProxyPublicIp),
	)
	assert.Nil(err)
	log.Printf("old %s  new %s", browserFirstProxyPublicIp, browserSecondProxyPublicIp)
	assert.Greater(len(browserSecondProxyPublicIp), 6)
	assert.NotEqual(browserFirstProxyPublicIp, browserSecondProxyPublicIp)
}

func TestSpoofHeadful(t *testing.T) {
	assert := assert.New(t)

	rc, err := NewRCContext()
	assert.Nil(err)
	defer rc.cancel()
	ctx, cancel := chromedp.NewContext(rc.browserCtx)
	defer cancel()

	var testFailures string
	err = chromedp.Run(ctx,
		SpoofHeadfulAction(),
		chromedp.Navigate("https://intoli.com/blog/not-possible-to-block-chrome-headless/chrome-headless-test.html"),
		chromedp.Evaluate(`
			var failed = Array.from(document.getElementsByClassName("failed"));
			failed.map(fail => fail.parentNode.innerText.replace("\n", ":")).join("\n")
		`, &testFailures),
	)
	assert.Nil(err)
	assert.Equal(testFailures, "")

	err = chromedp.Run(ctx,
		SpoofHeadfulAction(),
		chromedp.Navigate("http://arh.antoinevastel.com/bots/areyouheadless"),
			chromedp.Evaluate(`document.getElementById('res').innerText`, &testFailures))
	assert.Nil(err)
	log.Println(testFailures)
}

// type Score struct {
// 	Score float32 `json:"score"`
// }

// func TestGetTokenProxies(t *testing.T) {
// 	assert := assert.New(t)
// 	ctx, cancel := getCtx()
// 	defer cancel()

// 	startServer(ctx, t)

// 	file, err := os.Open("/Users/andyh2/drive/ysfc/proxies_4_11.txt")
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
// 			Proxy: proxy,//"http://proxy:vyreWAv@199.188.89.135:8000",
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
// 		log.Printf("%s %.1f", proxy, score.Score)
//   }
// }

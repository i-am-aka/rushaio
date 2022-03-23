package main

import _ "rush/net/http/pprof"

import (
	"strconv"
	"math/rand"
	"database/sql"
	"runtime/pprof"
	_ "github.com/mattn/go-sqlite3"
  "rush/task"
  "runtime/debug"
  "github.com/getsentry/sentry-go"
  "net/url"
  "strings"
  "context"
  "log"
  "rush/net/http"
  "time"
  "os"
  "sync"
  "flag"
  "fmt"
  "encoding/json"
  "rush/util"
  "runtime"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	fmt.Println("**START task_single**")
	task.RaiseFdLimit()
	PPROF := os.Getenv("PPROF") == "1"
	if os.Getenv("VERSION") == "" {
  	os.Setenv("VERSION", "aka")
	}

	os.Setenv("RUSH_DB", "/Users/aka/Library/Application Support/RushAIO/rush.db")
	db, _ := getDb()

	startEpochTime := time.Now().Unix()
	if PPROF {
		 f, err := os.Create(fmt.Sprintf("prof/cpuprofile-%d", startEpochTime))
     if err != nil {
         log.Fatal("could not create CPU profile: ", err)
     }
     defer f.Close() // error handling omitted for example
     if err := pprof.StartCPUProfile(f); err != nil {
         log.Fatal("could not start CPU profile: ", err)
     }
     defer pprof.StopCPUProfile()
	}
		// Flush buffered events before the program terminates.
		defer sentry.Flush(5 * time.Second)

		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
		// log.SetOutput(ioutil.Discard)
		log.SetPrefix("RUSH@")
		productUrl := flag.String("productUrl", "", "productUrl")

		size := flag.String("size", "", "size")

		// "http://foo:bar@127.0.0.1:8084",
		proxyStr := flag.String("proxy", "", "proxy")

		count := flag.Int("count", 1, "count")

		keywords := flag.String("keywords", "", "keywords")

		profileArg := flag.String("profile", "", "json-encoded profile string or file")

		config := flag.String("config", "", "comma separated kv pairs a=1,b=2")

		flag.Parse()
		required := []string{"productUrl", "profile"}
		seen := make(map[string]bool)
	  flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	  for _, req := range required {
	      if !seen[req] {
	          fmt.Fprintf(os.Stderr, "missing required argument -%s\n", req)
	          flag.PrintDefaults()
	          os.Exit(2)
	      }
	  }

	  var inbytes uint64 = 0
	  var outbytes uint64 = 0
	  bmut := sync.Mutex{}
	  printData := os.Getenv("BANDWIDTH") == "1"
	  DebugCountBytes := func(typ uint8, bytes uint) {
	    bmut.Lock()
	    defer bmut.Unlock()
	    if typ == 0 {
	      inbytes += uint64(bytes)
	      if uint64(inbytes / 1e3) % 10 == 0 {
	      }
	    } else if typ == 1 {
	      outbytes += uint64(bytes)
	      if uint64(outbytes / 1e3) % 10 == 0 {
	      }
	    }
	    if printData {
		    log.Printf("%d %d", typ, bytes)
		    log.Printf("in: %dkb", uint64(inbytes / 1e3))
		    log.Printf("out: %dkb", uint64(outbytes / 1e3))
		  }
	  }

	  // esWriter := zapcore.AddSync(ioutil.Discard)
	  // esEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())


		// var profileJson []byte
		// if _, err := os.Stat(*profileArg); os.IsNotExist(err) {
		// } else {
		// 	fcontents, err := ioutil.ReadFile(*profileArg)
		// 	if err != nil {
		// 		log.Fatal(fmt.Errorf("could not read profile: %+v", err))
		// 	}
		// 	profileJson = fcontents
		// }
		profileJson := []byte(*profileArg)
		var profile task.Profile
		if *profileArg == "fake" {
			profile = task.GenFakeProfile()
		} else {
			err := json.Unmarshal(profileJson, &profile)
			if err != nil {
				sentry.CaptureException(err)
				log.Println(fmt.Errorf("could not read profile: %+v", err))
				return
			}
		}

		url_, err := url.Parse(*productUrl)
		if err != nil {
			sentry.CaptureException(err)
			log.Println(fmt.Errorf("could not make product URL: %+v", err))
			return
		}
		 proxies := []*url.URL{}
		if *proxyStr != "" {
			for _, prox := range strings.Split(strings.Replace(*proxyStr, "-proxy", "", -1), "\n") {
        prox = strings.TrimSpace(prox)
        if prox == "" { continue }
				proxy, err := url.Parse(prox)
				// log.Println(proxy)

				if err != nil {
					sentry.CaptureException(err)
					log.Println(fmt.Errorf("could not parse proxy URL: %+v", err))
					return
				}
				proxies = append(proxies, proxy)
			}
		}
		log.Println()

		kwArr := strings.Split(*keywords, ",")
		kwArr = util.DeleteEmpty(kwArr)

		runTask := func(wg *sync.WaitGroup, i int, proxy *url.URL) {
			defer func(){
				if os.Getenv("CHECKOUTS") == "" {
	         if x:=recover();x!=nil{
	         		 debug.PrintStack()
	             fmt.Printf("%+v\n",x)
	         }
	       }
         wg.Done()
     }()
			tconfig := map[string]string {
				"size": *size,
			}
			cliconf := strings.Split(*config, ",")
			for _, cc := range cliconf {
				sp := strings.Split(cc, "=")
				if len(sp) == 1 {
					continue
				}
				tconfig[sp[0]] = sp[1]
			}
			statusCh := make(chan task.Status)
			task.DefaultCloudLogger.Init(context.Background())
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			task := task.CheckoutTask{
				Id: fmt.Sprintf("%d", i),
				Db: db,
				Url: url_,
				Keywords: kwArr,
				Proxy: proxy,
				ProxyGroup: os.Getenv("PROXY_GROUP"),
				Profile: profile,
				Cancel: cancel,
				DebugCountBytes: DebugCountBytes,
				Config: tconfig,
				StatusCh: statusCh,
				Ctx: ctx,
				StatusTelemetryEnabled: true,
				GetJwt: func() string { return os.Getenv("RUSH_JWT") },
				UserId: fmt.Sprintf("task_single_%d", rand.Intn(100000)),
				Version: os.Getenv("VERSION"),
			}
			go func() {
			  for {
		    	status, more := <-statusCh
		      if more {
		        log.Printf("(%v) - %v", task.Id, status)
		      } else {
		        return
		      }
			  }
			}()
      if os.Getenv("WEBHOOK") != "" {
        task.Webhook = os.Getenv("WEBHOOK")
      }

			if err := task.Run(); err == nil {
				if ck, err := strconv.Atoi(os.Getenv("CHECKOUTS")); err == nil {
					if ck > 1 {
						os.Setenv("CHECKOUTS", fmt.Sprintf("%d", ck-1))
					} else {
						panic("done")
					}
				}
			}
			// if err != nil {
			// 	log.Printf("%v", err.Error())
			// }
		}
		var wg sync.WaitGroup
		for i := 0; i < *count; i++ {
			wg.Add(1)
			// log.Println(proxies[i % len(proxies)])
			var prox *url.URL = nil
			if len(proxies) > 0 {
				prox = proxies[i % len(proxies)]
			}
			go runTask(&wg, i, prox)
		}
		wg.Wait()
		if os.Getenv("CLOUDLOG_WAIT") == "1" {
			log.Println("done, waiting to let cloudlogs flush")
			time.Sleep(30 * time.Second)
		} else {
			time.Sleep(1*time.Second)
		}

		if PPROF {
			f, err := os.Create(fmt.Sprintf("prof/memprofile-%d", startEpochTime))
			if err != nil {
			    log.Fatal("could not create memory profile: ", err)
			}
			defer f.Close() // error handling omitted for example
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
			    log.Fatal("could not write memory profile: ", err)
			}
		}
}

func getDb() (*sql.DB, error) {
  dbPath := os.Getenv("RUSH_DB")
  if dbPath == "" {
    dbPath = "./rush.db"
  }
  db, err := sql.Open("sqlite3", dbPath + "?_busy_timeout=5000&_journal_mode=WAL&cached=shared")
  if err != nil {
    return nil, err
  }

  return db, nil
}
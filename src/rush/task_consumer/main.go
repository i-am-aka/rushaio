package main

import (
  // "os/signal"
  "os/exec"
  "crypto/sha256"
  "fmt"
  "io/ioutil"
  "path/filepath"
	"rush/task"
	"math/rand"
	"context"
	"github.com/getsentry/sentry-go"
	"net/url"
	"log"
  "sync"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
	"time"
  "bytes"
	"os"
	"encoding/json"
  "runtime"
  "rush/ws"
  "github.com/go-zeromq/zmq4"
  "runtime/pprof"
)

func noopuint(_ uint){}
type TaskTx struct {
  Id string `json:"id"`
  Profile task.Profile `json:"profile"`
  Delay int `json:"delay"`
  Size string `json:"size"`
  Proxy string `json:"proxy"`
  ProxyGroup string `json:"proxy_group"`
  Url string `json:"url"`
  Keywords []string `json:"keywords"`
  Config map[string]string `json:"config"`
}

type CancelTx struct {
  Id string `json:"id"`
}

type DBUpdate struct {
  TaskId string `json:"task_id"`
  Url string `json:"url"`
  State string `json:"state"`
  Status string`json:"status"`
  Level int
}

type Consumer struct {
  TaskCancels map[string]context.CancelFunc
  TasksByHost map[string]uint32
  Hits map[string]uint32
  HitMut sync.Mutex
  Db *sql.DB
  l sync.RWMutex
  Ctx context.Context
  Cancel context.CancelFunc
  DbWriteChan chan DBUpdate
  TailChan chan DBUpdate
  AuthStatus ws.AuthStatus
  AuthLock sync.RWMutex
  Webhook string
  UserId string
  Version string
  Debug bool
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

func (c *Consumer) updateDbTaskState(taskId string, state string) (error) {
  tx, err := c.Db.Begin()
  if err != nil {
    log.Printf("failed to begin transaction: %+v", err)
    return err
  }
  stmt, err := tx.Prepare("update tasks2 set state=? where id = ?")
  if err != nil {
    log.Printf("failed to prepare db state update: %+v", err)
    return err
  }
  _, err = stmt.Exec(state, taskId)
  if err != nil {
    log.Printf("failed to update state: %+v", err)
    return err
  }
  tx.Commit()
  stmt.Close()
  return nil
}

func (c *Consumer) updateDbTask(taskId string, state string, status string, level int) (error) {
  tx, err := c.Db.Begin()
  if err != nil {
    log.Printf("failed to begin transaction: %+v", err)
    return err
  }
  stmt, err := tx.Prepare("update tasks2 set state=?, status=?, status_level=? where id = ?")
  if err != nil {
    log.Printf("failed to prepare db status update: %+v", err)
    return err
  }
  _, err = stmt.Exec(state, status, level, taskId)
  if err != nil {
    log.Printf("failed to update status: %+v", err)
    return err
  }
  tx.Commit()
  stmt.Close()
  return nil
}

func (c *Consumer) dbWriter() {
  for {
    // log.Println("dbWriter")
    select {
      case <-c.Ctx.Done():
        return
      case dbUpdate := <-c.DbWriteChan:
        // log.Printf("dbUpdate %+v", dbUpdate)
        if dbUpdate.Status == "" {
          c.updateDbTaskState(dbUpdate.TaskId, dbUpdate.State)
        } else {
          c.updateDbTask(dbUpdate.TaskId, dbUpdate.State, dbUpdate.Status, dbUpdate.Level)
        }
    }
  }
}


func (c *Consumer) stopper() {
  sock := zmq4.NewSub(c.Ctx)

  err := sock.Dial("tcp://127.0.0.1:42423")
  if err != nil {
    panic(err)
    return
  }

  err = sock.SetOption(zmq4.OptionSubscribe, "stop")
  if err != nil {
    panic(err)
    return
  }

  defer sock.Close()
  defer c.Cancel()

  for {
    // log.Println("stopper")
    msg, err := sock.Recv()
    if err != nil {
      log.Printf("could not receive message: %v", err)
      time.Sleep(1)
      continue
    }
    if c.Debug {
      log.Println(msg.Clone().String())
    }
    var stopTx CancelTx
    msgJson := bytes.SplitN(msg.Bytes(), []byte(" "), 2)[1]
    err = json.Unmarshal(msgJson, &stopTx)
    if err != nil {
      log.Printf("warning: failed to parse message %v", err)
      continue
    }

    c.l.Lock()
    cancel, has := c.TaskCancels[stopTx.Id]
    if has {
      // log.Printf("stopped %s", stopTx.Id)
      cancel()
      delete(c.TaskCancels, stopTx.Id)
    }
    c.l.Unlock()

    c.DbWriteChan <- DBUpdate{TaskId: stopTx.Id, State: "Stopped"}
  }
}

func (c *Consumer) cloudsocket(licenseKey string) {
  authStatusCh := make(chan ws.AuthStatus)
  go ws.AuthPoll(licenseKey, authStatusCh, c.Ctx)
  for {
    select {
    case <-c.Ctx.Done():
      return
    case authStatus, more := <-authStatusCh:
      if !more {
        panic("Auth socket died")
        return
      }
      c.AuthLock.Lock()
      c.AuthStatus = authStatus
      c.AuthLock.Unlock()
    }
  }
}

func (c *Consumer) GetJwt() string {
  c.AuthLock.Lock()
  jwt := c.AuthStatus.Jwt
  c.AuthLock.Unlock()
  return jwt
}

func (c *Consumer) AddHit(host string, pname string) {
  c.HitMut.Lock()
  defer c.HitMut.Unlock()
  if _, ok := c.Hits[host + "_" + pname]; ok {
    c.Hits[host + "_" + pname] += 1
  } else {
    c.Hits[host + "_" + pname] = 1
  }
}

func (c *Consumer) CheckHit(host string, pname string) uint32 {
  c.HitMut.Lock()
  defer c.HitMut.Unlock()
  if h, ok := c.Hits[host + "_" + pname]; ok {
    return h
  }
  return 0
}

func (c *Consumer) starter() {
  sock := zmq4.NewSub(c.Ctx)

  err := sock.Dial("tcp://127.0.0.1:42423")
  if err != nil {
    panic(err)
    return
  }

  err = sock.SetOption(zmq4.OptionSubscribe, "start")
  if err != nil {
    log.Printf("could not subscribe: %v", err)
    panic(err)
    return
  }

  defer c.Cancel()
  defer sock.Close()

  log.Println("Waiting for task...")
  for {
    // log.Println("starter")
    msg, err := sock.Recv()
    if err != nil {
      log.Printf("could not receive message: %v", err)
      time.Sleep(1*time.Second) // todo expbackoff
      continue
    }
    if os.Getenv("DEBUG") != "" {
      log.Println(msg.Clone().String())
    }
    var taskTx TaskTx
    msgJson := bytes.SplitN(msg.Bytes(), []byte(" "), 2)[1]
    err = json.Unmarshal(msgJson, &taskTx)
    if err != nil {
      log.Printf("warning: failed to parse message %v", err)
      continue
    }

    c.AuthLock.Lock()
    if c.AuthStatus.Jwt == "" {
      errMsg := "Unlicensed"
      if c.AuthStatus.ErrorMsg != "" {
        errMsg = c.AuthStatus.ErrorMsg
      }
      c.DbWriteChan <- DBUpdate{TaskId: taskTx.Id, State: "Stopped", Status: errMsg}
      c.AuthLock.Unlock()
      continue
    }
    c.AuthLock.Unlock()

    // log.Printf("%+v", taskTx)
    url_, err := url.Parse(taskTx.Url)
    if err != nil {
      log.Printf("warning: failed to parse task url %v", err)
      continue
    }

    ctx, cancel := context.WithCancel(c.Ctx)
    c.l.Lock()
    // c.TasksByHost[url_.Host] += 1
    c.TaskCancels[taskTx.Id] = cancel
    // log.Printf("tasks: %+v", c.TasksByHost[url_.Host])
    // if c.TasksByHost[url_.Host] > 100 && url_.Host == "www.hibbett.com" {
    //   c.l.Unlock()
    //   c.DbWriteChan <- DBUpdate{TaskId: taskTx.Id, State: "Stopped", Status: "Task Limit Exceeded"}
    //   continue
    // }
    c.l.Unlock()

    var proxyUrl *url.URL = nil
    if taskTx.Proxy != "" {
      proxyUrl, err = url.Parse(taskTx.Proxy)
      if err != nil {
        log.Printf("warning: failed to parse proxy url %v", err)
        continue
      }
    }

    statusCh := make(chan task.Status)
    var inbytes uint64 = 0
    var outbytes uint64 = 0
    bmut := sync.Mutex{}
    DebugCountBytes := func(typ uint8, bytes uint) {
      return

      bmut.Lock()
      defer bmut.Unlock()
      if typ == 0 {
        inbytes += uint64(bytes)
        if uint64(inbytes / 1e3) % 10 == 0 {
          log.Printf("in: %dkb", uint64(inbytes / 1e3))
        }
      } else if typ == 1 {
        outbytes += uint64(bytes)
        if uint64(outbytes / 1e3) % 10 == 0 {
          log.Printf("out: %dkb", uint64(outbytes / 1e3))
        }
      }
    }

    ct := task.CheckoutTask{
      DebugCountBytes: DebugCountBytes,
      Id: taskTx.Id,
      AddHit: c.AddHit,
      CheckHit: c.CheckHit,
      Db: c.Db,
      Delay: time.Duration(taskTx.Delay) * time.Millisecond,
      Version: c.Version,
      Url: url_,
      Proxy: proxyUrl,
      ProxyGroup: taskTx.ProxyGroup,
      Profile: taskTx.Profile,
      Config: map[string]string {
        "size": taskTx.Size,
      },
      StatusCh: statusCh,
      LastStatus: task.Status{},
      Ctx: ctx,
      Keywords: taskTx.Keywords,
      GetJwt: c.GetJwt,
      Webhook: c.Webhook,
      UserId: c.UserId,
      StatusTelemetryEnabled: os.Getenv("TELEMETRY") == "1" || rand.Float64() < 0.005,
    }
    for k, v := range taskTx.Config {
      ct.Config[k] = v
    }
    go func() {
      defer sentry.Recover()

      running := true
      for running {
        select {
        case <-ctx.Done():
          running = false
          break
        case status, more := <-statusCh:
          if status != ct.LastStatus && status.Message != "" && ctx.Err() == nil {
            ct.LastStatus = status
            dbUpdate := DBUpdate{TaskId: taskTx.Id, State: "Rushing", Url: ct.Url.String(), Status: status.Message, Level: status.Level}
            c.DbWriteChan <- dbUpdate
            // c.TailChan <- dbUpdate
            if ct.StatusTelemetryEnabled {
              if dbUpdateBytes, err := json.Marshal(dbUpdate); err == nil {
                var dbUpdateMap map[string]interface{}
                if err = json.Unmarshal(dbUpdateBytes, &dbUpdateMap); err == nil {
                  go ct.SendTelemetry(dbUpdateMap)
                }
              }
            }
          }
          if !more {
            // log.Println("Status closed")
            running = false
            break
          }
        }
      }
      c.DbWriteChan <- DBUpdate{TaskId: taskTx.Id, State: "Stopped"}


      c.l.Lock()
      defer c.l.Unlock()
      delete(c.TaskCancels, taskTx.Id)
      cancel()
    }()
    go func() {
      // defer close(statusCh)
      defer cancel()
      defer func() {
        err := recover()

        if err != nil {
          go func() {
            defer func() {
              recover()
            }()
            if cancel != nil {
              cancel()
            }
          }()
          sentry.CurrentHub().Recover(err)
          sentry.Flush(time.Second * 2)
        }
      }()
      ct.Run()
    }()
  }
}

func (c *Consumer) tailPublisher() {
  pub := zmq4.NewPub(context.Background())
  defer pub.Close()
  // pub.SetOption()

  err := pub.Listen("tcp://127.0.0.1:42425")
  if err != nil {
    sentry.CaptureException(err)
    // panic(err)
    return
  }

  for {
    select {
      case <-c.Ctx.Done():
        return
      case dbUpdate := <-c.TailChan:
        if frame, err := json.Marshal(dbUpdate); err == nil {
          msg := zmq4.NewMsg(frame)
          err = pub.Send(msg)
          if err != nil {
            log.Printf("tailPublisher err: %+v", err)
          }
        }
    }
  }
}

func (c *Consumer) ddGen() {
  defer func() {
    if r := recover(); r != nil {
      log.Printf("ddgen exit %+v", r)
    }
  }()
  ctx, cancel := context.WithCancel(c.Ctx)
  defer cancel()

  backoff := 0
  for {
    time.Sleep(time.Duration(backoff)*time.Second)
    if backoff > 30 {
      backoff = 30
    }
    backoff += 1
    if ctx.Err() != nil {
      return
    }
    // todo usually old serv alive, how to fix

    cmd := exec.CommandContext(ctx, task.GetDenoPath(), "run", "--allow-all", "datadome/dd.serv.js")
    cmd.Env = os.Environ()
    if c.Debug {
      cmd.Stdout = os.Stdout
      cmd.Stderr = os.Stdout
    } else {
      cmd.Stdout = nil
      cmd.Stderr = nil
    }
    err := cmd.Start()
    if err != nil {
      continue
    }
    err = cmd.Wait()
    if err != nil {
      continue
    }
  }
}

func (c *Consumer) pxGen() {
  ctx, cancel := context.WithCancel(c.Ctx)
  defer cancel()

  backoff := 0
  for {
    time.Sleep(time.Duration(backoff)*time.Second)
    if backoff > 30 {
      backoff = 30
    }
    backoff += 1
    if ctx.Err() != nil {
      return
    }
    cmd := exec.CommandContext(ctx, task.GetDenoPath(), "run", "--allow-all", "pxgen/serv.bundle.js")
    cmd.Env = os.Environ()
    if c.Debug {
      cmd.Stdout = os.Stdout
      cmd.Stderr = os.Stdout
    } else {
      cmd.Stdout = nil
      cmd.Stderr = nil
    }
    err := cmd.Start()
    if err != nil {
      sentry.CaptureException(err)
      continue
    }
    err = cmd.Wait()
    if err != nil {
      sentry.CaptureException(err)
      continue
    }
  }
}

func (c *Consumer) Status() {

}

func main() {
  if os.Getenv("PPROF") == "1" {
     f, err := os.Create(fmt.Sprintf("prof/cpuprofile-%d", time.Now().Unix()))
     if err != nil {
         log.Fatal("could not create CPU profile: ", err)
     }
     // defer f.Close() // error handling omitted for example
     if err := pprof.StartCPUProfile(f); err != nil {
         log.Fatal("could not start CPU profile: ", err)
     }
     go func() {
      time.Sleep(30*time.Second)
      pprof.StopCPUProfile()
      f.Close()
     }()
     // defer pprof.StopCPUProfile()
  }
  if runtime.GOOS != "windows" {
    task.RaiseFdLimit()
  }

	rand.Seed(time.Now().UnixNano())
  // log.SetOutput(ioutil.Discard)
  err := sentry.Init(sentry.ClientOptions{
    Dsn: SENTRY_DSN,
    AttachStacktrace: true,
    // Debug: true,
    Release: os.Getenv("VERSION"),
  })

  if err != nil {
    log.Fatalf("sentry.Init: %s", err)
  }

  defer sentry.Flush(2*time.Second)

  RecoverRepanic(func() {
    db, err := getDb()
    if err != nil {
      sentry.CaptureException(err)
      return
    }
    defer db.Close()

    ctx, cancel := context.WithCancel(context.Background())
    c := Consumer{
      Hits: make(map[string]uint32),
      HitMut: sync.Mutex{},
      TaskCancels: make(map[string]context.CancelFunc),
      TasksByHost: make(map[string]uint32),
      Db: db,
      l: sync.RWMutex{},
      Ctx: ctx,
      Cancel: cancel,
      DbWriteChan: make(chan DBUpdate),
      TailChan: make(chan DBUpdate),
      Version: os.Getenv("VERSION"),
      Debug: os.Getenv("DEBUG") != "",
    }
    licenseKey := os.Getenv("RUSH_LICENSE_KEY")
    if licenseKey == "" {
      if os.Getenv("RUSH_APPDIR") != "" {
        dat, err := ioutil.ReadFile(filepath.Join(os.Getenv("RUSH_APPDIR"), "key"))
        if err == nil {
          licenseKey = string(dat)
        }
      }
    }

    c.UserId = fmt.Sprintf("%x", sha256.Sum256([]byte(licenseKey)))
    sentry.ConfigureScope(func(scope *sentry.Scope) {
      scope.SetUser(sentry.User{
        ID: c.UserId,
      })
    })

    webhookBytes, err := ioutil.ReadFile(filepath.Join(os.Getenv("RUSH_APPDIR"), "config", "webhook"))
    if err == nil {
      c.Webhook = string(webhookBytes)
    }

    task.DefaultCloudLogger.Init(c.Ctx)
    task.DefaultCloudLogger.CtxData["uid"] = c.UserId
    task.DefaultCloudLogger.CtxData["version"] = os.Getenv("VERSION")
    task.DefaultCloudLogger.CtxData["env"] = os.Environ()

    go RecoverRepanic(func() { c.cloudsocket(licenseKey) })
    go RecoverRepanic(c.starter)
    go RecoverRepanic(c.stopper)
    go RecoverRepanic(c.dbWriter)
    go c.ddGen()
    // go RecoverRepanic(c.tailPublisher) // TODO this crashes

    for {
      select {
        case <-c.Ctx.Done():
        close(c.DbWriteChan)
        sentry.Flush(2*time.Second)
        return
      }
    }
  })
}


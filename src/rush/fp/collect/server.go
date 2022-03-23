package main

import (
	"io/ioutil"
	"log"
  "rush/net/http"
  "os"
)

func enableCors(w *http.ResponseWriter) {
  (*w).Header().Set("Access-Control-Allow-Origin", "*")
}

type FPServer struct {
	indexHtml []byte
  fpJs []byte
  pJs []byte
  ayJs []byte
  fpChan chan []byte
  mobileFpChan chan []byte
  outfile *os.File
  outfile_mobile *os.File
}

func (fps *FPServer) handle_index(w http.ResponseWriter, req *http.Request) {
	w.Write(fps.indexHtml)
}

func (fps *FPServer) handle_fpjs(w http.ResponseWriter, req *http.Request) {
  enableCors(&w)
  if (*req).Method == "OPTIONS" {
    return
  }
  w.Write(fps.fpJs)
}

func (fps *FPServer) handle_pjs(w http.ResponseWriter, req *http.Request) {
  enableCors(&w)
  if (*req).Method == "OPTIONS" {
    return
  }
  w.Write(fps.pJs)
}

func (fps *FPServer) handle_ay(w http.ResponseWriter, req *http.Request) {
  enableCors(&w)
  if (*req).Method == "OPTIONS" {
    return
  }
  w.Write(fps.ayJs)
}

func (fps *FPServer) handle_add(w http.ResponseWriter, req *http.Request) {
  enableCors(&w)
  if (*req).Method == "OPTIONS" {
    return
  }
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "400: could not read body", 400)
    return
  }
  fps.fpChan <- body
}

func (fps *FPServer) handle_log_mobile(w http.ResponseWriter, req *http.Request) {
  enableCors(&w)
  if (*req).Method == "OPTIONS" {
    return
  }
  body, err := ioutil.ReadAll(req.Body)
  if err != nil {
    http.Error(w, "400: could not read body", 400)
    return
  }
  fps.mobileFpChan <- body
}

func (fps *FPServer) writer() {
  running := true
  for running {
    select {
    case fp, more := <-fps.fpChan:
      log.Printf("write %s", string(fp))
      fps.outfile.Write(append(fp, []byte("\n")...))
      if !more {
        running = false
      }
    }
  }
}

func (fps *FPServer) mobile_writer() {
  running := true
  for running {
    select {
    case fp, more := <-fps.mobileFpChan:
      log.Printf("write %s", string(fp))
      fps.outfile_mobile.Write(append(fp, []byte("\n")...))
      if !more {
        running = false
      }
    }
  }
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}
	addr := ":" + port
	log.Println("listen on", addr)
  srv := &http.Server{Addr: addr}

  indexHtml, err := ioutil.ReadFile("index.html")
  if err != nil {
    log.Fatalf("could not load file: %+v", err)
  }
  fpJs, err := ioutil.ReadFile("fp.js")
  if err != nil {
    log.Fatalf("could not load file: %+v", err)
  }
  pJs, err := ioutil.ReadFile("p.js")
  if err != nil {
    log.Fatalf("could not load file: %+v", err)
  }
  ayJs, err := ioutil.ReadFile("ay.js")
  if err != nil {
    log.Fatalf("could not load file: %+v", err)
  }

  outfn := os.Getenv("FPFILE")
  if outfn == "" {
    outfn = "fp.txt"
  }
  outfile, err := os.OpenFile(outfn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  defer outfile.Close()
  if err != nil {
    log.Fatalf("could not open outfile: %+v", err)
  }
  fpChan := make(chan []byte)
  defer close(fpChan)

  mobile_outfn := os.Getenv("FPFILE_MOBILE")
  if mobile_outfn == "" {
    mobile_outfn = "fp_mobile.txt"
  }
  mobile_outfile, err := os.OpenFile(mobile_outfn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  defer mobile_outfile.Close()
  if err != nil {
    log.Fatalf("could not open outfile: %+v", err)
  }
  mobileFpChan := make(chan []byte)
  defer close(mobileFpChan)

  fps := FPServer {
  	indexHtml: indexHtml,
    fpJs: fpJs,
    pJs: pJs,
    ayJs: ayJs,
    fpChan: fpChan,
    mobileFpChan: mobileFpChan,
    outfile: outfile,
    outfile_mobile: mobile_outfile,
  }
  http.HandleFunc("/", fps.handle_index)
  http.HandleFunc("/ay.js", fps.handle_ay)
  http.HandleFunc("/p.js", fps.handle_pjs)
  http.HandleFunc("/fp.js", fps.handle_fpjs)
  http.HandleFunc("/add", fps.handle_add)
  http.HandleFunc("/log_mobile", fps.handle_log_mobile)

  go fps.writer()
  go fps.mobile_writer()
  if err := srv.ListenAndServe(); err != http.ErrServerClosed {
      log.Printf("ListenAndServe(): %v", err)
  }
}


package main

import (
  "encoding/json"
  "encoding/gob"
  "io/ioutil"
  "os"
  "rush/akgen"
)

func main() {
  bytes, err := ioutil.ReadFile("mact.json")
  if err != nil {
    panic(err)
  }
  var mact []akgen.AkWebSensorData

  json.Unmarshal(bytes, &mact)
  gobFn, err := os.Create("mact.gob")
  if err != nil {
    panic(err)
  }
  enc := gob.NewEncoder(gobFn)
  enc.Encode(mact)
  gobFn.Close()
}

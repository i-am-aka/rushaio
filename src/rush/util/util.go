package util

import (
  "io/ioutil"
  "log"
  "os"
)

func ReadFileOrStr(fnOrStr string) []byte {
  var jb []byte
  if _, err := os.Stat(fnOrStr); !os.IsNotExist(err) {
    f, err := ioutil.ReadFile(fnOrStr)
    if err != nil {
      log.Fatalf("could not load file: %+v", err)
    }
    jb = f
  } else {
    jb = []byte(fnOrStr)
  }
  return jb
}

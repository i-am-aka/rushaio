package task

import (
	// "log"
  "encoding/json"
  "os"
  "path/filepath"
	// "strings"
	// "runtime"
	// "os/exec"
)

func GetWmtDenoDir() (string, error) {
  dir, err := os.Getwd()
  if err != nil {
    return "", err
  }
  return filepath.Join(dir, "wmt", "deno"), nil
}

type WmtCardToken struct {
  NumEnc string `json:"numEnc"`
  CvcEnc string `json:"cvcEnc"`
  Integrity string `json:"integrity"`
}

func GetWmtToken (payKey *WmtPayKey, cardNumber string, cvc string) (*WmtCardToken, error) {
  var token WmtCardToken
  // denoDir, err := GetWmtDenoDir()
  // if err != nil {
  //   return nil, err
  // }

  payKeyJson, err := json.Marshal(payKey)
  if err != nil {
    return nil, err
  }

  out, err := JsEvalFile(map[string]string{}, "wmt/pay.js", string(payKeyJson), cardNumber, cvc)
  err = json.Unmarshal([]byte(out), &token)
  if err != nil {
    return nil, err
  }

  return &token, nil
}

package task

import (
	// "log"
  "os"
  "path/filepath"
	"strings"
	"runtime"
	"os/exec"
)

func GetEprotectDenoDir() (string, error) {
  dir, err := os.Getwd()
  if err != nil {
    return "", err
  }
  return filepath.Join(dir, "eprotect", "deno"), nil
}

func GetEprotectToken(cardNumber string, cvc string) (string, error) {
  var deno string
  if runtime.GOOS == "windows" {
    deno = "external/deno.exe"
  } else {
    deno = "external/deno_osx_x64"
  }
  denoAbspath, err := exec.LookPath(deno)
  if err != nil {
    return "", err
  }
  denoDir, err := GetEprotectDenoDir()
  if err != nil {
    return "", err
  }

  cmd := exec.Command(denoAbspath, "run", "--cached-only", "--allow-read", "--allow-net", "eprotect/tokenize.js", cardNumber, cvc)
  cmd.Env = append(os.Environ(), "DENO_DIR=" + denoDir)

  // TODO proxy
  out, err := cmd.Output()
  if err != nil {
    return "", err
  }
  return strings.TrimSpace(string(out[:])), nil
}

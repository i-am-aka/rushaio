package task

import (
	"encoding/json"
	"strings"
	"runtime"
	"path/filepath"
	"os"
	"os/exec"
)

func GetZnDenoDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
	  return "", err
	}
	return filepath.Join(dir, "zeronaught", "deno"), nil
}

func GetZeronaughtHeaders(scriptFn string) ([][2]string, error) {
  headers := make([][2]string, 0)
  var deno string
  if runtime.GOOS == "windows" {
    deno = "external/deno.exe"
  } else {
    deno = "external/deno_osx_x64"
  }
  denoAbspath, err := exec.LookPath(deno)
  if err != nil {
    return headers, err
  }
  denoDir, err := GetZnDenoDir()
  if err != nil {
    return headers, err
  }

  cmd := exec.Command(denoAbspath, "run", "--cached-only", "--allow-read", "zeronaught/zn.js", scriptFn)
  cmd.Env = append(os.Environ(), "DENO_DIR=" + denoDir)
  out, err := cmd.Output()
  if err != nil {
  	return headers, err
  }
  outLines := strings.Split(string(out), "\n")
  var header [2]string
  for _, line := range outLines {
  	err := json.Unmarshal([]byte(line), &header)
  	if err == nil {
  		headers = append(headers, header)
  	}
  }
  return headers, nil
}

func GetCkZeronaughtHeaders(proxy string) ([][2]string, map[string]string, error) {
	headers := make([][2]string, 0)
	cookies := map[string]string{}

	var deno string
	if runtime.GOOS == "windows" {
	  deno = "external/deno.exe"
	} else {
	  deno = "external/deno_osx_x64"
	}
	denoAbspath, err := exec.LookPath(deno)
	if err != nil {
	  return headers, cookies, err
	}
	denoDir, err := GetZnDenoDir()
	if err != nil {
	  return headers,cookies, err
	}
	args := []string {
		"run", "--cached-only", "--allow-read", "--allow-write", "--allow-net",
	}
	cert := os.Getenv("CERT")
	if cert != "" {
		args = append(args, "--cert")
		args = append(args, cert)
	}
	args = append(args, "zeronaught/znck.js")
	cmd := exec.Command(denoAbspath, args...)
	cmd.Env = append(os.Environ(), "DENO_DIR=" + denoDir)
	cmd.Env = append(cmd.Env, "HTTPS_PROXY=" + proxy)
	out, err := cmd.Output()
	if err != nil {
		return headers, cookies, err
	}
	outLines := strings.Split(string(out), "\n")
	var outMap map[string]map[string]string
	for _, line := range outLines {
		err := json.Unmarshal([]byte(line), &outMap)
		if err == nil {
			break
		}
	}
	headerMap := outMap["headers"]
	for k, v := range headerMap {
		headers = append(headers, [2]string{k,v})
	}
	return headers, outMap["cookies"], nil
}

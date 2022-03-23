package task

import (
	// "log
	"fmt"
	"context"
	"io/ioutil"
	"net/url"
	"strings"
	"os"
	"os/exec"
	"runtime"
)

type JsRuntime struct {

}

func GetDenoPath() string {
	if runtime.GOOS == "windows" {
	  return "external/deno.exe"
	} else if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		return "external/deno_osx_arm64"
	}
	return "external/deno_osx_64"
	// }
}

func DataDomeSendPayload(context context.Context, proxy *url.URL) error {
	// var authlessprox *url.URL
	// var err error
	// if proxy != nil {
	// 	authlessprox, err = NewProxyPassThru(context, proxy.String())
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// log.Printf("%+v", authlessprox)
	// // TODO have deno halt after xhr
	// // TODO invoke deno
	return nil
}

func JsEvalFile(env map[string]string, args ...string) (string, error) {
	args = append([]string{"run"}, args...)
	cmd := exec.Command(GetDenoPath(),  args...)
	cmd.Env = nil
	for key, value := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	// cmd.Env = append(os.Environ(), fmt.Sprintf("DENO_DIR=%s", denoDir))
	ret, err := cmd.Output()
	return strings.TrimSpace(string(ret)), err
}

func JsEval(js string) (string, error) {
	ret := ""

	content := []byte(js)
	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		return ret, err
	}

	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(content); err != nil {
		return ret, err
	}
	if err := tmpfile.Close(); err != nil {
		return ret, err
	}

	return JsEvalFile(map[string]string{}, tmpfile.Name())
}


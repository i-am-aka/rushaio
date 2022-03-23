package task

import (
	"log"
	"encoding/json"
	"strings"
	"runtime"
	// "path/filepath"
	"os"
	"os/exec"
	"github.com/pkg/errors"
)

type IncapChallengeSolution struct {
	Cookie string `json:"cookie"`
	ImgEndpoint string `json:"endpoint"`
}

func SolveIncapChallenge(proxy string, jsUrl string, cookies string) (IncapChallengeSolution, error) {
	var soln IncapChallengeSolution

	var deno string
	if runtime.GOOS == "windows" {
	  deno = "external/deno.exe"
	} else {
	  deno = "external/deno_osx_x64"
	}
	denoAbspath, err := exec.LookPath(deno)
	if err != nil {
	  return soln, err
	}
	denoDir, err := GetZnDenoDir()
	if err != nil {
	  return soln, err
	}
	args := []string {
		"run", "--cached-only", "--allow-read", "--allow-write", "--allow-net",
	}
	cert := os.Getenv("CERT")
	if cert != "" {
		args = append(args, "--cert")
		args = append(args, cert)
	}
	args = append(args, "incap/env.js")
	args = append(args, jsUrl)
	args = append(args, cookies)
	cmd := exec.Command(denoAbspath, args...)
	cmd.Env = append(os.Environ(), "DENO_DIR=" + denoDir)
	cmd.Env = append(cmd.Env, "HTTPS_PROXY=" + proxy)
	out, err := cmd.CombinedOutput()
	log.Printf("%s", string(out))
	if err != nil {
		return soln, err
	}
	outLines := strings.Split(string(out), "\n")
	for _, line := range outLines {
		err := json.Unmarshal([]byte(line), &soln)
		if err == nil {
			return soln, nil
		}
	}
	return soln, errors.New("No incap solution generated")
}

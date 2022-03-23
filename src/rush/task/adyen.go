package task

import (
  "math/rand"
  "fmt"
	"encoding/json"
  "log"
  "strings"
  "path/filepath"
  "os"
  "os/exec"
  "github.com/pkg/errors"
)

var (
  errAdyenEncryptFailed = errors.New("Adyen Encryption Failed")
)

func GetAdyenDenoDir() (string, error) {
  if os.Getenv("DENO_DIR") != "" {
    return os.Getenv("DENO_DIR"), nil
  }

  dir, err := os.Getwd()
  if err != nil {
    return "", err
  }
  return filepath.Join(dir, "adyen", "deno"), nil
}

func adyenEncrypt(card Card, cartMtime string) (string, error) {
	cardJson, err := json.Marshal(card)
	if err != nil {
		return "", err
	}


  log.Printf("%s", strings.Join([]string{GetDenoPath(), "run", "--cached-only", "--allow-read", "adyen/adyen.js", string(cardJson[:]), cartMtime}, " "))
  var cmd *exec.Cmd
  if rand.Float32() < 0.5 {
    cmd = exec.Command(GetDenoPath(), "run", "--cached-only", "--allow-read", "adyen/adyen.js", string(cardJson[:]), cartMtime)
  } else {
    cmd = exec.Command(GetDenoPath(), "run", "--cached-only", "--allow-read", "adyen/adyenfast.js", string(cardJson[:]), cartMtime)
  }
  // log.Println(cmd.String())
  // denoDir, err := GetAdyenDenoDir()
  // if err != nil {
  //   return "", err
  // }
  // log.Println(denoDir)

  // cmd.Env = append(os.Environ(), fmt.Sprintf("DENO_DIR=%s", denoDir))
  // out, err := cmd.Output()
  out, err := cmd.Output()
  log.Println(string(out))
  if err != nil {
    return "", err
  }
  return strings.TrimSpace(string(out[:])), nil
}

type AdyenMultiEncryptPayload struct {
  Number string
  ExpMonth string
  ExpYear string
  Cvc string
}

func AdyenMultiEncrypt(card Card, fp string) (*AdyenMultiEncryptPayload, error) {
  args := []string {
    "run", "--cached-only", "adyen2/enc.js", card.Number[0:4] + " " + card.Number[4:8] + " " + card.Number[8:12] + " " + card.Number[12:],
       fmt.Sprintf("%02d", card.ExpMonth), fmt.Sprintf("%d", card.ExpYear), card.Cvc,
  }
  if fp != "" {
    args = append(args, fp)
  }
  cmd := exec.Command(GetDenoPath(), args...)
  // log.Printf("%s", cmd.String())
  out, err := cmd.Output()

  if err != nil {
    return nil, err
  }
  lines := strings.Split(strings.TrimSpace(string(out[:])), "\n")
  if len(lines) != 4 {
   return nil, errAdyenEncryptFailed
  }
  var payload AdyenMultiEncryptPayload
  payload.Number = lines[0]
  payload.ExpMonth = lines[1]
  payload.ExpYear = lines[2]
  payload.Cvc = lines[3]
  return &payload, nil
}

func RCipher(str string) (string,error) {
  out, err := exec.Command(GetDenoPath(), "run", "adyen/rc.js", str).Output()
  if err != nil {
    return "", err
  }
  return strings.TrimSpace(string(out[:])), nil
}

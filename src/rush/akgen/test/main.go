package main

import (
	"rush/akgen"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var (
	CWD, _ = os.Getwd()
)
func testParse() {
	sdjPath := filepath.Join(CWD, "sd.json")
	sdj, err := ioutil.ReadFile(sdjPath)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	var sdfixture akgen.AkWebSensorData
	err = json.Unmarshal([]byte(sdj), &sdfixture)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	sdstrPath := filepath.Join(CWD, "sensor.txt")
	sdstr, err := ioutil.ReadFile(sdstrPath)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	_, err = akgen.ParseAkWebSensor(string(sdstr))
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func main() {
  samples := akgen.Samples{}
  err := samples.LoadMactSensors("../fp/mact.gob")
  if err != nil { panic(err) }
  err = samples.LoadSensors("../fp/sensor/sensors.gob")
  if err != nil { panic(err) }
  sdgen, err := samples.Gen("https://www.footlocker.ca", "Mozilla/5.0 (Linux; Android 10; SAMSUNG SM-G975U) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/11.1 Chrome/75.0.3770.143 Mobile Safari/537.36", "")
  	if err != nil { panic(err) }
	sdgen, err = samples.Gen("https://www.footlocker.ca", "Mozilla/5.0 (Linux; Android 10; SAMSUNG SM-G975U) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/11.1 Chrome/75.0.3770.143 Mobile Safari/537.36", "93B797A80793DB938F66247A6A071509~-1~YAAQpzkrF1IrU2BwAQAAxNWjsAPR+af3qtPf5noUJrmTXQPuesu3zlu1te4Ahsu3/4HKcPA5pyuSM9TG25z6Rc3BwQzxo/lCDTFksn5603xKZqdxbNNEnyd4PQyRnzrkp7iMYy9bjrk1iGT/50H4Gxad/rG1Yt/pG9z0iddR4V7BCKW0+EyOBMtthoBWqRtQtBjpXZJ7IYC69xwm8I0hFFzwkhWuXUXXiByzti/1k9pINe+v8wn8EeME8BfU+eMyhnEvMn3X0mmbBRjKOvC7w75ltJiLzhgA9KwkAtdDDcGVzSnQ+ZdMtDvgCp68bRnn/m/TKAeiy5nW8f5o3e2Qw71s/uv1FiDIm0s=~-1~||1-cxfsNyNlKq-4500-100-3000-2||~-1")
	if err != nil {
		log.Fatalf("%+v", err)
	}
	log.Printf("%+v", sdgen)
}

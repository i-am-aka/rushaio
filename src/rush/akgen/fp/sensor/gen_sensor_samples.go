package main

import (
  "encoding/gob"
  "io/ioutil"
  "log"
  "os"
  "runtime/debug"
  "rush/akgen"
  "rush/common"
  "strings"
)

func main() {
  fn := os.Getenv("FPFILE")
  if fn == "" {
    fn = "sensor.txt"
  }
  bytes, err := ioutil.ReadFile(fn)
  if err != nil {
    panic(err)
  }
  sensorStrs := strings.Split(string(bytes), "\n")
  sensors := []akgen.AkWebSensorData{}
  sensorsByUa := map[string]map[string]akgen.AkWebSensorData{}
  for _, sensor := range sensorStrs {
    func() {
      defer func() {
        if r := recover(); r != nil {
          // log.Println(sensor)
          log.Println("stacktrace from panic: \n" + string(debug.Stack()))
        }
      }()
    	s, err := akgen.ParseAkWebSensor(sensor)
    	if err != nil {
        log.Printf("%+v", err)
    		return
    	}
    	if common.Any(akgen.UA_BLACKLIST, func(uab string) bool {
    		return strings.Contains(s.Ua, uab)
    	}) {
    		return
    	}
      if !common.Any(akgen.UA_MOBILE, func(uab string) bool {
        return strings.Contains(s.Ua, uab)
      }) {
        return
      }
      // for _, uaopt := range SpreadUa(s.Ua) {
      // log.Printf("%s", s.Ua)
      //   scp := *s
      //   scp.Ua = uaopt
      if _, ok := sensorsByUa[s.Ua]; ok {
        if _, ok := sensorsByUa[s.Ua][sensor]; !ok {
          sensorsByUa[s.Ua][sensor] = *s
        }
      } else {
        sensorsByUa[s.Ua] = map[string]akgen.AkWebSensorData{ sensor: *s }
      }
    	// sensors = append(sensors, *s)
    }()
  }

  tot := 0
  for ua, uasensors := range sensorsByUa {
    log.Printf("(%d) %s", len(uasensors), ua)
    tot += len(uasensors)
    if len(uasensors) > 512 {
      if len(uasensors) > 35555 {
        t := 0
        for _, sensor := range uasensors {
          if t > 35555 {
            break
          }
          t += 1
          sensors = append(sensors, sensor)
        }
      } else {
        for _, sensor := range uasensors {
          sensors = append(sensors, sensor)
        }
      }
    }
  }
  log.Printf("%d samples (%d total)", len(sensors), tot)

  gobFn, err := os.Create("sensors.gob")
  if err != nil {
    panic(err)
  }
  enc := gob.NewEncoder(gobFn)
  enc.Encode(sensors)
  gobFn.Close()
}

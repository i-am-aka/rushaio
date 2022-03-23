package akgen

import (
  "fmt"
  "log"
  "strconv"
  "io"
  // "math"
  "compress/gzip"
  "io/ioutil"
  "math/rand"
  "net"
  "os"
  "encoding/gob"
  "github.com/pkg/errors"
  "strings"
  "rush/common"
  "sync"
  // "github.com/google/go-cmp/cmp"
  "os/exec"
)

var (
  CWD, _ = os.Getwd()
  AK_SOCK = os.Getenv("AK_SOCK")
  errUnknownUserAgent = errors.New("Unknown user agent")
)

type Samples struct {
  Ports []string
  PortMut sync.Mutex
  MactSensors []AkWebSensorData
  Sensors []AkWebSensorData
  UserAgents []string
  DesktopUserAgents []string
  SensorIdxByUa map[string][]int
  SensorIdxWith map[string][]int
}

func (s *Samples) LoadMactSensors(fn string) error {
  var f io.ReadCloser
  var err error
  f, err = os.Open(fn)
  if err != nil { return err }
  ff := f
  defer ff.Close()
  if strings.HasSuffix(fn, ".gz") {
    f, err = gzip.NewReader(f)
    if err != nil { return err }
    defer f.Close()
  }
  dec := gob.NewDecoder(f)
  sensors := []AkWebSensorData{}
  err = dec.Decode(&sensors)
  if err != nil { return err }

  for idx, _ := range sensors {
    msref := &sensors[idx]
    xMax := 0
    yMax := 0
    yMin := int(1e9)
    xMin := int(1e9)
    for _, m := range msref.Mact[:] {
      if m.X > xMax {
        xMax = m.X
      }
      if m.Y > yMax {
        yMax = m.Y
      }
      if m.Y < yMin {
        yMin = m.Y
      }
      if m.X < xMin {
        xMin = m.X
      }
    }

    xdelt := 0
    ydelt := 0

    if xMax > msref.WindowInnerWidth {
      xdelt = xMax - msref.WindowInnerWidth + 1
      if xMin < xdelt {
        // log.Printf("skipping %v, need scale", msref.Mact[0])
        continue
      }
    }
    if yMax > msref.WindowInnerHeight {
      ydelt = yMax - msref.WindowInnerHeight + 1
      if yMin < ydelt {
        // log.Printf("skipping %v, need scale", msref.Mact[0])
        continue
      }
    }
    // if yMin < 0 || xMin < 0 {
    //  log.Printf("ymin %d xmin %d", yMin, xMin)
    // }
    if yMin < 0 {
      ydelt = yMin - 1
    }
    if xMin < 0 {
      xdelt = xMin - 1
    }

    for idx, _ := range msref.Mact[:]{
      mref := &msref.Mact[idx]
      mref.X -= xdelt
      mref.Y -= ydelt
      if mref.Y < 0 || mref.X < 0 {
        log.Printf("%+v %v %v", mref, xdelt, ydelt)
      }
    }

    s.MactSensors = append(s.MactSensors, *msref)
  }
  return nil
}

var UA_BLACKLIST = []string {
  "Intel Mac OS X 10.0;",
  "rv:69.0",
  "Mozilla/5.0 (Linux; Android 9; Pixel 2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36",
  // "Chrome/86",
  "VOG-L29",
  "FxiOS/30.0",
  "FxiOS",
  // "Firefox",
  // "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.92 Safari/537.36",
  // "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36",
  // "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.92 Safari/537.36",
  // "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.87 Safari/537.36",
  // "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.106 Safari/537.36",
  // "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.106 Safari/537.36",
  // "Headless",
  // "Mac OS X 10_10",
  // "Mac OS X 10_11",
  // "Mac OS X 10_12",
  // "Chrome/6",
  "Chrome/4",
  "Chrome/5",
  "Chrome/6",
  "Chrome/7",
  "Chrome/80",
  "Chrome/81",
  "Chrome/82",
  "Chrome/83",
  "Chrome/84",
  // "Chrome/85",
  "CPU OS 9_",
  "CPU OS 10_",
  "CPU OS 11_",
  "CPU OS 12_",
  "iPhone OS 9",
  "iPhone OS 10",
  "iPhone OS 11",
  "iPhone OS 12",
  "iPhone OS 12",
  "Version/13.0",
  "Version/13.1",
  "Version/13.2",

 //  "Chrome/80",
 //  "Chrome/81",
 //  "Chrome/82",
  // "iPhone",
  // "iPad",
  // "iPad; CPU OS 12",

  // "Version/14", // TODO fix iOS 14, probably clear fp difference

  "Android 4",
  "Android 5",
  "Android 6",
  // "Android 7",
  // "Android 7.1",
  "Version/11",
  "Intel Mac OS X 11_0",
  "iPod touch",

  // "Mobile/15E148",
  // "iPhone",
  // "iPad",
  // "CriOS",

  // "Chrome/74.0.3729.136",
  "MicroMessenger/7.0.1",
  // "Android 7",
  // "Mozilla/5.0 (Linux; Android 9; SM-J737A) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.117 Mobile Safari/537.36",
  // "Mozilla/5.0 (Linux; Android 10; Pixel 3 XL) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.117 Mobile Safari/537.36",
}

var UA_MOBILE = []string {
  "Android",
  "Mobile",
  "iPhone",
  "iPad",
  "Firefox",
  "CriOS",
  // "Chrome/87",
  // "Chrome/88",
  // "Chrome/83",
  // "Chrome/84",
}
var UA_WHITELIST_DESKTOP = []string {
  "Chrome/83",
  "Chrome/84",
}
// var UA_ALIAS = map[string][]string {
//  `Chrome/85.* `: []string {
//    "85.0.4183.81"
//    "85.0.4183.101"
//    "85.0.4183.109"
//    "85.0.4183.120"
//    "85.0.4183.125"
//    "85.0.4183.127",
//  },
// }

func (s *Samples) LoadSensors(fn string) error {
  var f io.ReadCloser
  var err error
  f, err = os.Open(fn)
  if err != nil { return err }
  ff := f
  defer ff.Close()
  if strings.HasSuffix(fn, ".gz") {
    f, err = gzip.NewReader(f)
    if err != nil { return err }
    defer f.Close()
  }
  dec := gob.NewDecoder(f)
  var ss []AkWebSensorData
  err = dec.Decode(&ss)
  if err != nil { return err }
  for _, sen := range ss {
    // if len(sen.SpeechSynthVoices) < 16 {
    //  continue
    // }
    // if !strings.Contains(sen.Ua, "Firefox") {
    //  continue
    // }
    s.Sensors = append(s.Sensors, sen)
  }
  s.SensorIdxByUa = make(map[string][]int)
  s.SensorIdxWith = make(map[string][]int)
  cnt := 0
  for idx, sensor := range s.Sensors[:] {
    bt := ""
    if strings.Contains(sensor.Ua, "Firefox") {
      bt = "firefox"
    } else if strings.Contains(sensor.Ua, "Chrome") {
      bt = "chrome"
    }
    if strings.Contains(sensor.Ua, "Mobile") || strings.Contains(sensor.Ua, "Android") {
      bt += "_android"
    }
    if len(sensor.Perms) > 1 {
      s.SensorIdxWith["Perms"] = append(s.SensorIdxWith["Perms"], idx)
      s.SensorIdxWith[bt + "_Perms"] = append(s.SensorIdxWith[bt + "_Perms"], idx)
      s.SensorIdxWith[sensor.Ua + "_Perms"] = append(s.SensorIdxWith[sensor.Ua + "_Perms"], idx)

    }
    if len(sensor.SpeechSynthVoices) >= 16 {
      s.SensorIdxWith["SpeechSynthVoices"] = append(s.SensorIdxWith["SpeechSynthVoices"], idx)
      s.SensorIdxWith[bt + "_SpeechSynthVoices"] = append(s.SensorIdxWith[bt + "_SpeechSynthVoices"], idx)
      s.SensorIdxWith[sensor.Ua + "_SpeechSynthVoices"] = append(s.SensorIdxWith[sensor.Ua + "_SpeechSynthVoices"], idx)

    }
    if len(sensor.FpvalFonts) > 0 {
      s.SensorIdxWith[sensor.Ua + "_fpval"] = append(s.SensorIdxWith[sensor.Ua + "_fpval"], idx)
    }
    if strings.Contains(sensor.Ua, "Firefox") {
      s.SensorIdxWith["firefox"] = append(s.SensorIdxWith["firefox"], idx)
    }
    if strings.Contains(sensor.Ua, "Android") {
      s.SensorIdxWith["android"] = append(s.SensorIdxWith["android"], idx)
    }
    for _, ua := range SpreadUa(sensor.Ua) {
      if _, exists := s.SensorIdxByUa[ua]; !exists {
        s.UserAgents = append(s.UserAgents, ua)
      }

      s.SensorIdxByUa[ua] = append(s.SensorIdxByUa[ua], idx)
      cnt += 1
    }
  }
  // for _, ua := range s.UserAgents {
  //  // log.Println(ua)
  // }
  log.Printf("%d UAs %d FPs", len(s.UserAgents), cnt)
  return nil
}


var (
  DESKTOP_SCREEN_RES_PARTIAL_SENSORS = []AkWebSensorData{
    AkWebSensorData{
      ScreenWidth: 1366,
      ScreenHeight: 768,
    },
    AkWebSensorData{
      ScreenWidth: 1440,
      ScreenHeight: 900,
    },
    AkWebSensorData{
      ScreenWidth: 1920,
      ScreenHeight: 1080,
    },
    AkWebSensorData{
      ScreenWidth: 1280,
      ScreenHeight: 1024,
    },
    AkWebSensorData{
      ScreenWidth: 1920,
      ScreenHeight: 1200,
    },
    AkWebSensorData{
      ScreenWidth: 1536,
      ScreenHeight: 864,
    },
  }
)

var Cr85Versions = []string{
  "Chrome/87.0.4280.66",
  "Chrome/86.0.4240.185",
  "Chrome/86.0.4240.114",
  "Chrome/86.0.4240.110",
  "Chrome/86.0.4240.99",
  "Chrome/86.0.4240.75",
  "Chrome/85.0.4183.127",
  "Chrome/85.0.4183.81",
}

func SpreadUa(ua string) []string {
  return []string{ua}

  parts := strings.Split(ua, " ")
  uaopts := []string{ua}
  for _, part := range parts {
    if strings.Contains(part, "Chrome/") || strings.Contains(part, "Chrome/") {
      for _, sp := range Cr85Versions {
        if strings.Compare(part, sp) > 0 {
          uaopts = append(uaopts, strings.Replace(ua, part, sp, -1))
        }
      }
    }
    //  else if strings.Contains(part, "Chrome/84.0.4147.") {
    //   for _, sp := range Cr84Versions {
    //     if sp != part {
    //       uaopts = append(uaopts, strings.Replace(ua, part, sp, -1))
    //     }
    //   }
    // }
  }
  return uaopts
}

func (s *Samples) Gen(url string, ua string, abck string, stage string, ajIndx int) (*AkWebSensorData, error) {
  uaSensorIndices, exists := s.SensorIdxByUa[ua]
  if !exists {
    fmt.Println(ua)
    return nil, errUnknownUserAgent
  }

  // TODO can optimize by only calling for challenges / speeding up deno
  sd, err := GenSensorBase(abck, s.Ports[rand.Intn(len(s.Ports))])
  if err != nil {
    return nil, err
  }
  // fmt.Printf("%+v\n", sd)

  genFull := true // !strings.HasSuffix(abck, "==~-1~-1~-1") && len(sd.Chal1Soln.T) == 0// && strings.Contains(abck, "-1~||-1||~-1")) // false //; len(sd.Chal1Soln.T) > 0
  bt := ""
  if strings.Contains(ua, "Firefox") {
    bt = "firefox"
  } else if strings.Contains(ua, "Chrome") {
    bt = "chrome"
  }
  if strings.Contains(ua, "Mobile") || strings.Contains(ua, "Android") {
    bt += "_android"
  }

  src := s.Sensors[uaSensorIndices[rand.Intn(len(uaSensorIndices))]]
  // fmt.Printf("SRC %+v\n", src)
  // fmt.Printf("expected %s got ua %s\n", ua, src.Ua)
  sd.Ua = ua
  sd.UaHash = AkAb(ua)
  sd.Psub = src.Psub
  sd.X12 = 1
  sd.InitTime = "0"
  sd.Pen = "0"
  sd.Wen = "0"
  sd.Den = "0"
  sd.Cpen = "0"
  sd.Lang = "en-US"
  sd.I1 = "0"
  sd.Dm = "0"
  sd.Non = 1
  sd.Opc = "0"
  sd.Fc = src.Fc
  sd.Sc = "0"
  sd.Isc = src.Isc
  sd.X11 = "0"

  // sd.D3 = int(sd.StartTs % 1e7)
  // sd.DumbassPipeCharcode = src.DumbassPipeCharcode
  sd.Loc = src.Loc



  sd.Vib = src.Vib
  sd.Bat = src.Bat
  sd.Wrc = src.Wrc
  // sd.D3 = src.D3
  // sd.O9 = src.O9
  sd.PeCnt = "0"
  isFirefox := strings.Contains(sd.Ua, "Firefox")
  sd.Xagg = src.Xagg
  if isFirefox && sd.Xagg != 11059 {
    fmt.Printf("bad xagg %d\n", sd.Xagg)
  }
  sd.HasWindowDeviceMotionEvent = src.HasWindowDeviceMotionEvent
  sd.HasWindowDeviceOrientationEvent = src.HasWindowDeviceOrientationEvent
  sd.HasWindowTouchEvent = src.HasWindowTouchEvent
  // sd.ScreenPixelDepth = src.ScreenPixelDepth
  sd.Plen = src.Plen

  isMobile := strings.Contains(src.Ua, "Mobile") || strings.Contains(src.Ua, "Android")
  if isMobile {
    // if rand.Float32() < 0.5 {
    //   sd.ScreenWidth = src.ScreenHeight
    //   sd.ScreenHeight = src.ScreenWidth
    //   sd.ScreenAvailWidth = src.ScreenAvailHeight
    //   sd.ScreenAvailHeight = src.ScreenAvailWidth
    //   sd.WindowInnerHeight = src.WindowInnerWidth
    //   sd.WindowInnerWidth = src.WindowInnerHeight
    //   sd.WindowOuterWidth = src.ScreenAvailHeight
    // } else {
      sd.ScreenWidth = src.ScreenWidth
      sd.ScreenHeight = src.ScreenHeight
      sd.ScreenAvailWidth = src.ScreenAvailWidth
      sd.ScreenAvailHeight = src.ScreenAvailHeight
      sd.WindowInnerHeight = src.WindowInnerHeight
      sd.WindowInnerWidth = src.WindowInnerWidth
      sd.WindowOuterWidth = src.WindowOuterWidth
    // }
  } else {
  	// screen := DESKTOP_SCREEN_RES_PARTIAL_SENSORS[rand.Intn]
    sd.ScreenWidth = src.ScreenWidth
    sd.ScreenHeight = src.ScreenHeight
    sd.ScreenAvailWidth = src.ScreenAvailWidth
    sd.ScreenAvailHeight = src.ScreenAvailHeight
    sd.WindowInnerHeight = src.WindowInnerHeight
    sd.WindowInnerWidth = src.WindowInnerWidth
    sd.WindowOuterWidth = src.WindowOuterWidth
    // if sd.WindowOuterWidth == 0 {
    //   sd.WindowOuterWidth = src.WindowInnerWidth
    // }

    // screenSd := DESKTOP_SCREEN_RES_PARTIAL_SENSORS[rand.Intn(len(DESKTOP_SCREEN_RES_PARTIAL_SENSORS))]
    // fmt.Printf("%+v\n", screenSd)
    // sd.ScreenWidth = screenSd.ScreenWidth
    // sd.ScreenHeight = screenSd.ScreenHeight
    // sd.ScreenAvailWidth = sd.ScreenWidth
    // sd.ScreenAvailHeight = sd.ScreenHeight - (src.ScreenHeight - src.ScreenAvailHeight)
    // fmt.Printf("\n\n%v = %v - %v\n\n", sd.ScreenAvailHeight, sd.ScreenHeight, src.ScreenHeight - src.ScreenAvailHeight)
    // // fmt.Printf("\n\n%v = %v - %v\n\n", src)
    // // sd.ScreenWidth = src.ScreenWidth
    // // sd.ScreenHeight = src.ScreenHeight
    // // sd.ScreenAvailWidth = src.ScreenAvailWidth
    // // sd.ScreenAvailHeight = src.ScreenAvailHeight


    // sd.WindowInnerHeight = sd.ScreenAvailHeight / 2 + rand.Intn(sd.ScreenAvailHeight / 2)
    // sd.WindowInnerWidth = sd.ScreenAvailWidth / 2 + rand.Intn(sd.ScreenAvailWidth / 2)
    // if src.WindowOuterWidth > 0 {
    // 	sd.WindowOuterWidth = sd.WindowInnerWidth + (src.WindowOuterWidth - src.WindowInnerWidth)
    // } else {
    // 	sd.WindowOuterWidth = 0 ///sd.WindowInnerWidth
    // }
    // log.Printf("%+v", sd)
  }


  sd.Perms = "8"
  genFull = stage != "onload"
  if genFull {
    bperms := s.SensorIdxWith[sd.Ua + "_Perms"]
    if len(bperms) > 0 {
      sd.Perms = s.Sensors[bperms[rand.Intn(len(bperms) - 1)]].Perms
    } else {
    	bperms = s.SensorIdxWith[bt + "_Perms"]
    	if len(bperms) > 0 {
    		sd.Perms = s.Sensors[bperms[rand.Intn(len(bperms) - 1)]].Perms
    	}
    }

    bperms = s.SensorIdxWith[sd.Ua + "_SpeechSynthVoices"]
    if len(bperms) > 0 {
      sd.SpeechSynthVoices = s.Sensors[bperms[rand.Intn(len(bperms) - 1)]].SpeechSynthVoices
    } else {
    	bperms = s.SensorIdxWith[bt + "_SpeechSynthVoices"]
    	if len(bperms) > 0 {
    		sd.Perms = s.Sensors[bperms[rand.Intn(len(bperms) - 1)]].SpeechSynthVoices
    	}
    }
  }

  // if isFirefox {
  //  // // sd.Perms = "11133333333333333333"
  //  // // sd.Perms = "8"
  //  // if genFull {
  //  //  sd.Perms = "11133333331333333333"
  //  //  sd.Perms = [rand.Intn(len(s.SensorIdxWith[bt + "_Perms"]) - 1)]].Perms
  //  // } else {
  //  //  sd.Perms = "8"
  //  // }
  // }
  sd.NavigatorDoNotTrack = src.NavigatorDoNotTrack
  sd.Loc = ""
  // if !isMobile {
  //  sd.Perms = s.Sensors[s.SensorIdxWith["Perms"][rand.Intn(len(s.SensorIdxWith["Perms"]) - 1)]].Perms
  // } else {
  //  sd.Perms = src.Perms
  // }
  // log.Printf("%+v", sd.Perms)


  if genFull {
    fps := s.SensorIdxWith[sd.Ua + "_fpval"]
    srcfp := src
    if src.FpvalFonts == "" {
      srcfp = s.Sensors[fps[rand.Intn(len(fps))]]
    }
    // fpvals := []interface{} {
    //   srcfp.FpvalCanvas1,
    //   srcfp.FpvalCanvas2,
    //   srcfp.FpvalFonts,
    //   srcfp.FpvalPluginInfo,
    //   srcfp.FpvalSessionStorage,
    //   srcfp.FpvalLocalStorage,
    //   srcfp.FpvalIndexedDb,
    //   srcfp.FpvalTzoffset,
    //   srcfp.FpvalWebtc,
    //   srcfp.ScreenColorDepth,
    //   srcfp.ScreenPixelDepth,
    //   srcfp.NavigatorCookieEnabled,
    //   srcfp.NavigatorJavaEnabled,
    //   srcfp.NavigatorDoNotTrack,
    // }
    // fpvalStrs := make([]string, len(fpvals))
    // for idx, fpval := range fpvals {
    //   fpvalStrs[idx] = fmt.Sprintf("%v", fpval)
    // }
    sd.FpvalTzoffset = srcfp.FpvalTzoffset
    // fpvalStr := strings.Join(fpvalStrs, ";")
    // sd.FpvalHash = AkAb(fpvalStr)
    sd.IsBrave = srcfp.IsBrave
    sd.FontHash = srcfp.FontHash
    sd.DevicePixelRatio = srcfp.DevicePixelRatio
    // log.Printf("speech %s", sd.SpeechSynthVoices)
    // if len(src.SpeechSynthVoices) >= 16 {
    //  sd.SpeechSynthVoices = src.SpeechSynthVoices
    // }
    // sd.WebglVendor = src.WebglVendor
    // sd.WebglRenderer = src.WebglRenderer
    // sd.WebglExtHash = src.WebglExtHash
    // sd.WebglExtLen = src.WebglExtLen
    sd.CanvasTeststr = srcfp.CanvasTeststr
    sd.CanvasHash = srcfp.CanvasHash
    sd.FpvalFonts = srcfp.FpvalFonts
    sd.FpvalSessionStorage = srcfp.FpvalSessionStorage
    sd.FpvalLocalStorage = srcfp.FpvalLocalStorage
    sd.FpvalIndexedDb = srcfp.FpvalIndexedDb
    sd.FpvalCanvas1 = srcfp.FpvalCanvas1
    sd.FpvalCanvas2 = srcfp.FpvalCanvas2
    sd.ScreenColorDepth = srcfp.ScreenColorDepth
    sd.ScreenPixelDepth = srcfp.ScreenPixelDepth
    sd.FpvalPluginInfo = srcfp.FpvalPluginInfo
    sd.FpvalWebtc = srcfp.FpvalWebtc
    sd.NavigatorJavaEnabled = srcfp.NavigatorJavaEnabled
    sd.NavigatorCookieEnabled = srcfp.NavigatorCookieEnabled
    sd.NavigatorDoNotTrack = srcfp.NavigatorDoNotTrack
    sd.NavigatorFpFas = srcfp.NavigatorFpFas
    sd.FpvalTd = 5 + rand.Intn(100)
  }
  //  else {
  //  t.FpvalTd = -999999
  // }

  if len(sd.Chal1Soln.T) ==  0 {
    if isFirefox {
      //  t.fpValstr = '-1', t.fpValCalculated = !1, t.rVal = '-1', t.rCFP = '-1', t.cache = {}, t.td = -999999
      // t.fpValstr

      // sd.NavigatorFpFas = 25543097
      if genFull {
        sd.MathAbsT = rand.Intn(2) * 200
        sd.MathAcosT = rand.Intn(2) * 200
        sd.MathAsinT = rand.Intn(2) * 200
        sd.MathAtanhT = rand.Intn(2) * 200
        sd.MathCbrtT = rand.Intn(2) * 200
        sd.MathExpT = rand.Intn(2) * 200
        sd.MathRandomT = 0
        sd.MathRoundT = rand.Intn(2) * 200
        sd.MathSqrtT = 0
        sd.IsfiniteT = 0
        sd.IsnanT = 0
        sd.ParsefloatT = 0
        sd.ParseintT = 0
        sd.JsonparseT = rand.Intn(2) * 200
      } else {
        sd.MathAbsT = -1
      }
    } else {
      sd.MathAbsT = 25 + rand.Intn(125)
      sd.MathAcosT = 25 + rand.Intn(125)
      sd.MathAsinT = 30 + rand.Intn(150)
      sd.MathAtanhT = 28 + rand.Intn(380)
      sd.MathCbrtT = 50 + rand.Intn(170)
      sd.MathExpT = 55 + rand.Intn(175)
      sd.MathRandomT = 30 + rand.Intn(270)
      sd.MathRoundT = 12 + rand.Intn(118)
      sd.MathSqrtT = 15 + rand.Intn(690)
      sd.IsfiniteT = 8 + rand.Intn(117)
      sd.IsnanT = 11 + rand.Intn(100)
      sd.ParsefloatT = 8 + rand.Intn(22)
      sd.ParseintT = 15 + rand.Intn(30)
      sd.JsonparseT = 400 + rand.Intn(800)
    }
  }
  if isFirefox {
    sd.SdEndDt = 1// 2 + rand.Intn(15)
    sd.Tst = 6// 4 + rand.Intn(10)
    sd.SdEndDt2 = 0
  } else {
    sd.SdEndDt = 3 + rand.Intn(47)
    sd.Tst = rand.Intn(25)
    sd.SdEndDt2 = 1 + rand.Intn(14)
  }

  if abck == "" {
    sd.NoCookie = "1"
  } else {
    sd.NoCookie = "0"
  }

  sd.URL = url




  sd.Kact = []AkKact{}

  if stage == "click" {
    sd.SdDt = 9000 + rand.Intn(50990)
    sd.SdDt2 = sd.SdDt + 1

    // sd.Mact = []AkMact{}
    mactSample := s.MactSensors[rand.Intn(len(s.MactSensors))]
    mt0 := 50 + rand.Intn(203)
    mtend := mt0 + 100 + rand.Intn(100)
    clickAt := mtend + 807 + rand.Intn(common.MinInt(sd.SdDt - 1562, mtend + 17484) - (mt0+250))
    mt := mt0
    sd.Mact = make([]AkMact, len(mactSample.Mact))

    // log.Printf("%+v", mactSample.Mact)
    // TODO 50/50 fwds/backwards based on sample
    // TODO random rotations of entire sample

    // mactReverse := mactSample.Mact[0:len(mactSample.Mact) - 1]
    // ReverseSlice(mactReverse)
    // mactSample.Mact = append(mactReverse, mactSample.Mact[len(mactSample.Mact) - 1])
    xMin := int(^uint(0) >> 1)
    xMax := 0
    yMin := int(^uint(0) >> 1)
    yMax := 0
    xSum := 0
    ySum := 0


    // tX = 0
    // tY = 0

    // log.Printf("%v x %v", sd.WindowInnerWidth, sd.WindowInnerHeight)
    for idx, m := range mactSample.Mact[:] {
      sd.Mact[idx] = AkMact {
        MeCnt: m.MeCnt,
        Typ: m.Typ,
        Dt: mt,
        X: int(float64(sd.WindowInnerWidth) * (float64(m.X) / float64(mactSample.WindowInnerWidth))),
        Y: int(float64(sd.WindowInnerHeight) * (float64(m.Y) / float64(mactSample.WindowInnerHeight))),
        EleIDHash: m.EleIDHash,
        MouseBtn: m.MouseBtn,
      }
      // if sd.Mact[idx].Y < 0 || sd.Mact[idx].X < 0 {
      //  log.Printf("%v %v", sd.Mact[idx], m)
      // }

      if sd.Mact[idx].X < xMin {
        xMin = sd.Mact[idx].X
      }
      if sd.Mact[idx].X > xMax {
        xMax = sd.Mact[idx].X
      }
      if sd.Mact[idx].Y < yMin {
        yMin = sd.Mact[idx].Y
      }
      if sd.Mact[idx].Y > yMax {
        yMax = sd.Mact[idx].Y
      }
      if idx != len(mactSample.Mact) - 1 {
        xSum += sd.Mact[idx].X
        ySum += sd.Mact[idx].Y
      }
      mt += rand.Intn(3)
    }

    if stage != "click" {
      sd.Mact = nil
    }
    // works good on ys, but is small adjust. try whole window
    // tX := -1*xMin + rand.Intn(xMin + xMax) // TODO margin ?
    // tY := -1*yMin + rand.Intn(yMin + yMax)

    right := sd.WindowInnerWidth - xMax
    bottom := sd.WindowInnerHeight - yMax
    // log.Printf("%+v", mactSample.Mact)
    // log.Printf("right=%v bottom=%v width=%v height=%v xMin=%v xMax=%v yMin=%v yMax=%v", right, bottom, sd.WindowInnerWidth, sd.WindowInnerHeight, xMin, xMax, yMin, yMax)
    tX := -1*xMin + rand.Intn(right + xMin + 1) // TODO margin ?
    tY := -1*yMin + rand.Intn(bottom + yMin + 1)


    // TODO translate to random position in window

    // tX := -100 + rand.Intn(200)
    // tY := -100 + rand.Intn(200)

    // if xMin + tX < 10 {
    //  tX = 10 + rand.Intn(40)
    // }
    // if yMin + tY < 10 {
    //  tY = 10 + rand.Intn(40)
    // }
    // if xMax + tX > (sd.WindowInnerWidth - 50) {
    //  tX = -50 + rand.Intn(30)
    // }
    // if yMax + tY > (sd.WindowInnerHeight - 50) {
    //  tY = -50 + rand.Intn(30)
    // }

    scale := (0.4 + rand.Float64()*0.6)
    // scale := 1.0

    // xCtr := float64(xSum) / float64(len(mactSample.Mact)-1)
    // yCtr := float64(ySum) / float64(len(mactSample.Mact)-1)

    flipX := rand.Float32() < 0.5
    flipY := rand.Float32() < 0.5
    // flipX = false
    // flipY = false

    // log.Printf("tx=%v ty=%v | ctr=(%v,%v) rot=%v", tX, tY, xCtr, yCtr, rot)
    for idx, _ := range sd.Mact[:] {
      m := &sd.Mact[idx]
      // log.Printf("%+v", m)

      if flipX {
        m.X = xMax - m.X
      }
      if flipY {
        m.Y = yMax - m.Y
      }
      // m.X = int(xCtr + math.Cos(rot) * (float64(m.X)-xCtr) - math.Sin(rot) * (float64(m.Y)-yCtr))
      // m.Y = int(yCtr + math.Sin(rot) * (float64(m.X)-xCtr) + math.Cos(rot) * (float64(m.Y)-yCtr))

      // log.Printf("%+v", m)

      // log.Printf("%d %d %d", tX, tY, scale)
      // m.X = tX + m.X
      // m.X = tX + m.Y
      m.X = tX + xMin + int(scale*float64(m.X - xMin))// int(scale * float64(m.X))
      m.Y = tY + yMin + int(scale*float64(m.Y - yMin)) // int(scale * float64(m.Y))



      // log.Printf("%+v", m)
      // log.Println()
    }

    // mactReverse := mactSample.Mact[0:len(mactSample.Mact) - 1]
    // ReverseSlice(mactReverse)
    // for idx, _ := range mactSample.Mact[:len(mactSample.Mact) - 1] {
    //  mr := &sd.Mact[idx]
    //  mr.X = int(float64(sd.WindowInnerWidth) * (float64(mactReverse[idx].X) / float64(mactSample.WindowInnerWidth)))
    //  mr.Y = int(float64(sd.WindowInnerHeight) * (float64(mactReverse[idx].Y) / float64(mactSample.WindowInnerHeight)))
    // }


    // log.Printf("%+v", sd.Mact)
    // TESTING FIRST SENSOR
    // sd.; []AkMact{}

    if len(sd.Mact) > 0 {
      clickRef := &sd.Mact[len(sd.Mact) - 1]
      clickRef.MeCnt = 700 + rand.Intn(6300)
      clickRef.EleIDHash = "-1"
      clickRef.Dt = clickAt
      // TODO randomize click loc
      sd.MeCnt = sd.Mact[len(sd.Mact) - 1].MeCnt + 1
    } else {
      sd.MeCnt = 0
    }
  }

  if len(sd.Chal1Soln.T) > 0 {
    sd.AjType = 8
    sd.AjIndx = 2
  } else if len(sd.Mact) > 0 {
    sd.AjType = 1  // todo 5 for tact if ever used
    sd.AjIndx = 3+rand.Intn(7) // TODO increase
  } else if stage == "fp" {
    sd.AjType = 9
    sd.AjIndx = 1
  } else {
    sd.AjType = 0
    sd.AjIndx = 0
  }
  if ajIndx > -1 && stage == "click" {
    sd.AjIndx = ajIndx
  }
  // } else if isFirefox {
  //  sd.AjType = 0
  //  sd.AjIndx = 0

  if !isFirefox {
    sd.Doact = []AkDoAct {
      AkDoAct {
        DoeCnt: 0,
        Dt: 18 + rand.Intn(47),
        Alpha: "-1",
        Beta: "-1",
        Gamma: "-1",
      },
    }
    sd.Dmact = []AkDmAct {
      AkDmAct {
        DmeCnt: 0,
        Dt: 18 + rand.Intn(32),
        Ax: "-1",
        Ay: "-1",
        Az: "-1",
        Agx: "-1",
        Agy: "-1",
        Agz: "-1",
        Ra: "-1",
        Rb: "-1",
        Rg: "-1",
      },
    }
  }

  sd.Ta = 0

  for _, ke := range sd.Kact[:] {
    typ, _ := strconv.Atoi(ke.Typ)
    sd.KeVel += (typ + ke.KeCnt + ke.Dt + -3)
    sd.Ta += ke.Dt
  }
  sd.KeCnt = len(sd.Kact)

  for _, me := range sd.Mact[:] {
    typ, _ := strconv.Atoi(me.Typ)
    sd.Ta += me.Dt
    sd.MeVel += me.MeCnt + typ + me.Dt + me.X + me.Y
  }

  if len(sd.Doact) > 0 {
    sd.DoeVel = sd.Doact[0].Dt
    sd.Ta += sd.DoeVel
  }
  if len(sd.Dmact) > 0{
    sd.DmeVel = sd.Dmact[0].Dt
    sd.Ta += sd.DmeVel
  }

  // sd.TeVel = 32
  // TODO tact

  sd.VelSum = (sd.KeVel) + (sd.MeVel) + sd.DoeVel + sd.DmeVel + sd.TeVel

  // sd.Chal0Soln.T = genChalTiming(len(sd.Chal0Soln.T))
  // sd.Chal1Soln.T = genChalTiming(len(sd.Chal1Soln.T))
  // if len(sd.Chal1Soln.Lg) > 0 {
  //  tt := 4000+rand.Intn(2000)
  //  sd.Chal1Soln.Lg[len(sd.Chal1Soln.Lg) - 3] = fmt.Sprintf("%d", tt)
  //  sd.Chal1Soln.Lg[len(sd.Chal1Soln.Lg) - 1] = fmt.Sprintf("%d", sd.StartTs + int64(tt) + 100 + rand.Intn(15))
  // }
  // sd.Chal2Soln.T = genChalTiming(len(sd.Chal2Soln.T))
  // if len(sd.Chal2Soln.Lg) > 0 {
  //  sd.Chal2Soln.Lg[len(sd.Chal2Soln.Lg) - 3] = fmt.Sprintf("%d", 4000+rand.Intn(2000))
  // }

  // log.Printf("%+v\n", sd)
  // if diff := cmp.Diff(src, *sd); diff != "" {
  //   log.Printf("diff:\n%s", diff)
  // }
  // log.Println()
  return sd, nil
}

func genChalTiming(len int) []string {
  t := make([]string, len)
  for idx, _ := range t[:] {
    t[idx] = strconv.Itoa(1 + rand.Intn(114))
  }
  return t
}

func GetSensorBaseString(sdDt int, abck string) (string, error) {
  denoAbspath, err := exec.LookPath("deno")
  if err != nil {
    return "", err
  }
  // fmt.Printf("%s %s %s\n", denoAbspath, string(cardJson[:]), cartMtime)
  cmd := exec.Command(denoAbspath, "run", "--allow-read", "test.js", fmt.Sprintf("%d", sdDt), abck)
  cmd.Dir = "../akgen/js"
  out, err := cmd.Output()
  if err != nil {
    return "", err
  }
  return strings.TrimSpace(string(out[:])), nil
}

func GenSensorBase(abck string, port string) (*AkWebSensorData, error) {
  sdDt := 9000 + rand.Intn(50990)

  // AK_SOCK = os.Getenv("AK_SOCK")
  // if AK_SOCK == "" {
  //  AK_SOCK = "../akgen/js/ak.sock"
  // }
  c, err := net.Dial("tcp", "127.0.0.1:" + port)
  if err != nil {
    return nil, err
  }
  defer c.Close()
  _, err = c.Write([]byte(fmt.Sprintf("%d~_~%s", sdDt, abck)))
  if err != nil {
    return nil, err
  }
  resp, err := ioutil.ReadAll(c)
  if err != nil {
    return nil, err
  }
  sdstr := string(resp)

  sd, err := ParseAkWebSensor(sdstr)
  if err != nil {
    return nil, err
  }
  // log.Printf("%+v", sd)
  // if len(sd.Chal1Soln.T) != 0 {
  //   log.Printf("setting sddt=%d %+v", sdDt, sd.Chal1Soln.T)
  //   sd.SdDt = sdDt
  // }
  // sd.SdDt2 = sd.SdDt + 1

  return sd, nil
}


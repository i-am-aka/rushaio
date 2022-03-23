package akgen

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"rush/common"
	"strconv"
	"github.com/pkg/errors"
)

var (
	errInvalidSensorString = errors.New("Invalid sensor string")
)
type AkChalSoln struct {
	Al []string
	T []string
	Il []string
	Lg []string
}

type AkKact struct {
	KeCnt           int    `json:"ke_cnt"`
	Typ             string `json:"typ"`
	// TypStr          string `json:"typ_str"`
	Dt              int    `json:"dt"`
	KeycodeClass    string `json:"keycode_class"`
	// KeycodeClassStr string `json:"keycode_class_str"`
	L               string `json:"l"`
	Specialkeys     string `json:"specialkeys"`
	InpEleIDHash    string `json:"inp_ele_id_hash"`
}

type AkMact struct {
	MeCnt     int         `json:"me_cnt"`
	Typ       string      `json:"typ"`
	// TypStr    string      `json:"typ_str"`
	Dt        int         `json:"dt"`
	X         int         `json:"x"`
	Y         int         `json:"y"`
	EleIDHash interface{} `json:"ele_id_hash"`
	MouseBtn  interface{} `json:"mouse_btn"`
}

type AkDoAct struct {
	DoeCnt int    `json:"doe_cnt"`
	Dt     int    `json:"dt"`
	Alpha  string `json:"alpha"`
	Beta   string `json:"beta"`
	Gamma  string `json:"gamma"`
}

type AkDmAct struct {
	DmeCnt int    `json:"dme_cnt"`
	Dt     int    `json:"dt"`
	Ax     string `json:"ax"`
	Ay     string `json:"ay"`
	Az     string `json:"az"`
	Agx    string `json:"agx"`
	Agy    string `json:"agy"`
	Agz    string `json:"agz"`
	Ra     string `json:"ra"`
	Rb     string `json:"rb"`
	Rg     string `json:"rg"`
}

type AkWebSensorData struct {
	Version                         string `json:"version"`
	Hash                            string `json:"hash"`
	Ua                              string `json:"ua"`
	Xagg                            int    `json:"xagg"`
	Psub                            int    `json:"psub"`
	Lang                            string `json:"lang"`
	Prod                            string `json:"prod"`
	Plen                            string `json:"plen"`
	Pen                             string `json:"pen"`
	Wen                             string `json:"wen"`
	Den                             string `json:"den"`
	Z1                              int    `json:"z1"`
	D3                              int    `json:"d3"`
	ScreenAvailHeight               int    `json:"screen_availHeight"`
	ScreenAvailWidth                int    `json:"screen_availWidth"`
	ScreenWidth                     int    `json:"screen_width"`
	ScreenHeight                    int    `json:"screen_height"`
	WindowInnerWidth                int    `json:"window_innerWidth"`
	WindowInnerHeight               int    `json:"window_innerHeight"`
	WindowOuterWidth                int    `json:"window_outerWidth"`
	Cpen                            string `json:"cpen"`
	I1                              string `json:"i1"`
	Dm                              string `json:"dm"`
	Cwen                            string `json:"cwen"`
	Non                             int    `json:"non"`
	Opc                             string `json:"opc"`
	Fc                              string `json:"fc"`
	Sc                              string `json:"sc"`
	Wrc                             int    `json:"wrc"`
	Isc                             string `json:"isc"`
	Vib                             int    `json:"vib"`
	Bat                             int    `json:"bat"`
	X11                             string `json:"x11"`
	X12                             int    `json:"x12"`
	UaHash                          int    `json:"ua_hash"`
	MathRandom                      string `json:"math_random"`
	StartTsDiv2                     string  `json:"start_ts_div2"`
	IsBrave													string
	Loc                             string `json:"loc"`
	HasWindowDeviceOrientationEvent string `json:"has_window_DeviceOrientationEvent"`
	HasWindowDeviceMotionEvent      string `json:"has_window_DeviceMotionEvent"`
	HasWindowTouchEvent             string `json:"has_window_TouchEvent"`
	Kact                            []AkKact `json:"kact"`
	Mact []AkMact `json:"mact"`
	Tact  []interface{} `json:"tact"`
	Doact []AkDoAct `json:"doact"`
	Dmact []AkDmAct `json:"dmact"`
	Pact                   []interface{} `json:"pact"`
	Vcact                  []interface{} `json:"vcact"`
	URL                    string        `json:"url"`
	KeVel                  int           `json:"ke_vel"`
	MeVel                  int           `json:"me_vel"`
	TeVel                  int           `json:"te_vel"`
	DoeVel                 int           `json:"doe_vel"`
	DmeVel                 int           `json:"dme_vel"`
	PeVel                  string        `json:"pe_vel"`
	VelSum                 int           `json:"vel_sum"`
	SdDt                   int           `json:"sd_dt"`
	InitTime               string        `json:"init_time"`
	StartTs                int64         `json:"start_ts"`
	FpvalTd                int           `json:"fpval_td"`
	D2                     int           `json:"d2"`
	KeCnt                  int           `json:"ke_cnt"`
	MeCnt                  int           `json:"me_cnt"`
	L                      int           `json:"l"`
	PeCnt                  string        `json:"pe_cnt"`
	TeCnt                  string        `json:"te_cnt"`
	SdDt2                  int           `json:"sd_dt2"`
	Ta                     int           `json:"ta"`
	NoCookie               string        `json:"no_cookie"`
	Cookie                 string        `json:"cookie"`
	CookieHash             int           `json:"cookie_hash"`
	CanvasTeststr          int           `json:"canvas_teststr"`
	CanvasHash             int           `json:"canvas_hash"`
	NavigatorFpFas         int           `json:"navigator_fp_fas"`
	StartTsJig 						 string
	DumbassPipeCharcode    string
	AnotherOne 						 string

	AjType                 int           `json:"aj_type"`
	AjIndx                 int           `json:"aj_indx"`
	MathAbsT               int           `json:"math_abs_t"`
	MathAcosT              int           `json:"math_acos_t"`
	MathAsinT              int           `json:"math_asin_t"`
	MathAtanhT             int           `json:"math_atanh_t"`
	MathCbrtT              int           `json:"math_cbrt_t"`
	MathExpT               int           `json:"math_exp_t"`
	MathRandomT            int           `json:"math_random_t"`
	MathRoundT             int           `json:"math_round_t"`
	MathSqrtT              int           `json:"math_sqrt_t"`
	IsfiniteT              int           `json:"isfinite_t"`
	IsnanT                 int           `json:"isnan_t"`
	ParsefloatT            int           `json:"parsefloat_t"`
	ParseintT              int           `json:"parseint_t"`
	JsonparseT             int           `json:"jsonparse_t"`
	Chal0Soln AkChalSoln
	Chal1Soln AkChalSoln
	Chal2Soln AkChalSoln
	Perms                  string        `json:"perms"`
	FpvalCanvas1           int           `json:"fpval_canvas1"`
	FpvalCanvas2           int           `json:"fpval_canvas2"`
	FpvalFonts             string        `json:"fpval_fonts"`
	FpvalPluginInfo        string        `json:"fpval_pluginInfo"`
	FpvalSessionStorage    string        `json:"fpval_sessionStorage"`
	FpvalLocalStorage      string        `json:"fpval_localStorage"`
	FpvalIndexedDb         string        `json:"fpval_indexedDb"`
	FpvalTzoffset          int           `json:"fpval_tzoffset"`
	FpvalWebtc             string        `json:"fpval_webtc"`
	ScreenColorDepth       int           `json:"screen_colorDepth"`
	ScreenPixelDepth       int           `json:"screen_pixelDepth"`
	NavigatorCookieEnabled string        `json:"navigator_cookieEnabled"`
	NavigatorJavaEnabled   string        `json:"navigator_javaEnabled"`
	NavigatorDoNotTrack    int           `json:"navigator_doNotTrack"`

	FontHash string
	DevicePixelRatio string
	SpeechSynthVoices string
	WebglVendor string
	WebglRenderer string
	WebglExtHash string
	WebglExtLen string

	FpvalHash              int           `json:"fpval_hash"`
	O9                     int           `json:"o9"`
	SdFirstpartHash        int           `json:"sd_firstpart_hash"`
	SdEndDt                int           `json:"sd_end_dt"`
	Tst                    int           `json:"tst"`
	SdEndDt2               int           `json:"sd_end_dt2"`
}

const (
	AKA_VERSION = "1.66" // TODO RESTORE
)


func ParseAkWebSensor(sdstr string) (*AkWebSensorData, error) {
	sdstr = strings.TrimSpace(sdstr)
	var err error
	sd := AkWebSensorData{Version: AKA_VERSION}
	re := regexp.MustCompile(`-1,2,-94,-[0-9]+,`)
	parts := re.Split(sdstr, -1)
	if len(parts) < 26 {
		return nil, errors.WithStack(errInvalidSensorString)
	}
	sd.Hash = strings.Split(parts[0], AKA_VERSION)[0]

	part1 := strings.Split(parts[1], ",")
	uaEndIdx := common.Index(part1, "uaend")
	sd.Ua = strings.Join(part1[0:uaEndIdx], ",")
	part1 = part1[uaEndIdx+1:]
	sd.Xagg, err = strconv.Atoi(part1[0])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sd.Psub, err = strconv.Atoi(part1[1])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sd.Lang = part1[2]
	sd.Prod = part1[3]
	sd.Plen = part1[4]
	sd.Pen = part1[5]
	sd.Wen = part1[6]
	sd.Den = part1[7]
	sd.Z1, err = strconv.Atoi(part1[8])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sd.D3, err = strconv.Atoi(part1[9])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sd.ScreenAvailWidth, err = strconv.Atoi(part1[10])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sd.ScreenAvailHeight, err = strconv.Atoi(part1[11])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sd.ScreenWidth, err = strconv.Atoi(part1[12])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sd.ScreenHeight, err = strconv.Atoi(part1[13])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sd.WindowInnerWidth, err = strconv.Atoi(part1[14])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sd.WindowInnerHeight, err = strconv.Atoi(part1[15])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sd.WindowOuterWidth, err = strconv.Atoi(part1[16])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	part1 = part1[1:]
	sd.Cpen = strings.Split(part1[17], ":")[1]
	sd.I1 = strings.Split(part1[18], ":")[1]
	sd.Dm = strings.Split(part1[19], ":")[1]
	sd.Cwen = strings.Split(part1[20], ":")[1]
	sd.Non, err = strconv.Atoi(strings.Split(part1[21], ":")[1])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.Opc = strings.Split(part1[22], ":")[1]
	sd.Fc = strings.Split(part1[23], ":")[1]
	sd.Sc = strings.Split(part1[24], ":")[1]
	sd.Wrc, err = strconv.Atoi(strings.Split(part1[25], ":")[1])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.Isc = strings.Split(part1[26], ":")[1]
	sd.Vib, err = strconv.Atoi(strings.Split(part1[27], ":")[1])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.Bat, err = strconv.Atoi(strings.Split(part1[28], ":")[1])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.X11 = strings.Split(part1[29], ":")[1]
	sd.X12, err = strconv.Atoi(strings.Split(part1[30], ":")[1])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.UaHash, err = strconv.Atoi(part1[31])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.MathRandom = part1[32]
	sd.StartTsDiv2 = part1[33]
	sd.IsBrave = part1[34]
	sd.Loc = strings.Split(part1[35], ":")[1]

	part2 := strings.Split(parts[2], ",")
	sd.HasWindowDeviceOrientationEvent = part2[0]
	sd.HasWindowDeviceMotionEvent = part2[1]
	sd.HasWindowTouchEvent = part2[2]


	kact := strings.Split(parts[5], ";")
	sd.Kact = make([]AkKact, len(kact) - 1 )
	for idx, part := range kact {
		kp := strings.Split(part, ",")
		if len(kp) != 7 { continue }
		ka := &sd.Kact[idx]
		ka.KeCnt, err = strconv.Atoi(kp[0])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		ka.Dt, err = strconv.Atoi(kp[2])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		ka.Typ = kp[1]
		ka.KeycodeClass = kp[3]
		ka.L = kp[4]
		ka.Specialkeys = kp[5]
		ka.InpEleIDHash = kp[6]
	}

	mact := strings.Split(parts[6], ";")
	sd.Mact = make([]AkMact, len(mact) - 1)
	for idx, part := range mact {
		mp := strings.Split(part, ",")
		if len(mp) < 5 { continue }
		ma := &sd.Mact[idx]
		meCnt, err := strconv.Atoi(mp[0])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		dt, err := strconv.Atoi(mp[2])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		x, err := strconv.Atoi(mp[3])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		y, err := strconv.Atoi(mp[4])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		if len(mp) > 5 {
			ma.EleIDHash = mp[5]
		}
		if len(mp) > 6 {
			ma.MouseBtn = mp[6]
		}
		ma.MeCnt = meCnt
		ma.Typ = mp[1]
		ma.Dt = dt
		ma.X = x
		ma.Y = y
	}



	doact := strings.Split(parts[8], ";")
	sd.Doact = make([]AkDoAct, len(doact) - 1)
	for idx, part := range doact {
		mp := strings.Split(part, ",")
		if len(mp) != 5 { continue }
		doa := &sd.Doact[idx]
		doa.DoeCnt, err = strconv.Atoi(mp[0])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		doa.Dt, err = strconv.Atoi(mp[1])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		doa.Alpha = mp[2]
		doa.Beta = mp[3]
		doa.Gamma = mp[4]
	}


	dmact := strings.Split(parts[9], ";")
	sd.Dmact = make([]AkDmAct, len(dmact) - 1)
	for idx, part := range dmact {
		dmp := strings.Split(part, ",")
		if len(dmp) != 11 { continue }
		dma := &sd.Dmact[idx]
		dma.DmeCnt, err = strconv.Atoi(dmp[0])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		dma.Dt, err = strconv.Atoi(dmp[1])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		dma.Ax = dmp[2]
		dma.Ay = dmp[3]
		dma.Az = dmp[4]
		dma.Agx = dmp[5]
		dma.Agy = dmp[6]
		dma.Agz = dmp[7]
		dma.Ra = dmp[8]
		dma.Rb = dmp[9]
		dma.Rg = dmp[10]
	}

	sd.URL = parts[12]

	p13 := strings.Split(parts[13], ",")
	sd.KeVel, err = strconv.Atoi(p13[0])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.KeVel -= 1
	sd.MeVel, err = strconv.Atoi(p13[1])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.MeVel -= 32
	sd.TeVel, err = strconv.Atoi(p13[2])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.TeVel -= 32

	sd.DoeVel, err = strconv.Atoi(p13[3])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.DmeVel, err = strconv.Atoi(p13[4])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.PeVel = p13[5]
	sd.VelSum, err = strconv.Atoi(p13[6])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.SdDt, err = strconv.Atoi(p13[7])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.InitTime = p13[8]
	sd.StartTs, err = strconv.ParseInt(p13[9], 10, 64)
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.FpvalTd, err = strconv.Atoi(p13[10])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.D2, err = strconv.Atoi(p13[11])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.KeCnt, err = strconv.Atoi(p13[12])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.MeCnt, err = strconv.Atoi(p13[13])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.L, err = strconv.Atoi(p13[14])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.PeCnt = p13[15]
	sd.TeCnt = p13[16]

	sd.SdDt2, err = strconv.Atoi(p13[17])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.Ta, err = strconv.Atoi(p13[18])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.NoCookie = p13[19]
	sd.Cookie = p13[20]
	sd.CookieHash, err = strconv.Atoi(p13[21])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.CanvasTeststr, err = strconv.Atoi(p13[22])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.CanvasHash, err = strconv.Atoi(p13[23])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.NavigatorFpFas, err = strconv.Atoi(p13[24])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	if len(p13) > 25 {
		sd.StartTsJig = p13[25]
		sd.DumbassPipeCharcode = p13[26]
		if len(p13) > 27 {
			sd.AnotherOne = p13[27]
		} else {
			sd.AnotherOne = "-1"
		}
	}

	aj := strings.Split(parts[14], ",")
	sd.AjType, err = strconv.Atoi(aj[0])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.AjIndx, err = strconv.Atoi(aj[1])
	if err != nil {
	  return nil, errors.WithStack(err)
	}

	math := strings.Split(parts[15], ",")
	if len(math) > 10 {
		sd.MathAbsT, err = strconv.Atoi(math[0])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.MathAcosT, err = strconv.Atoi(math[1])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.MathAsinT, err = strconv.Atoi(math[2])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.MathAtanhT, err = strconv.Atoi(math[3])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.MathCbrtT, err = strconv.Atoi(math[4])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.MathExpT, err = strconv.Atoi(math[5])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.MathRandomT, err = strconv.Atoi(math[6])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.MathRoundT, err = strconv.Atoi(math[7])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.MathSqrtT, err = strconv.Atoi(math[8])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.IsfiniteT, err = strconv.Atoi(math[9])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.IsnanT, err = strconv.Atoi(math[10])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.ParsefloatT, err = strconv.Atoi(math[11])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.ParseintT, err = strconv.Atoi(math[12])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.JsonparseT, err = strconv.Atoi(math[13])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
	}

	sd.Chal0Soln = ParseChalSoln(parts[17])
	sd.Chal1Soln = ParseChalSoln(parts[18])
	sd.Chal2Soln = ParseChalSoln(parts[19])

	sd.Perms = parts[20]

	fpval := strings.Split(parts[21], ";")
	if len(fpval) > 11 {
		sd.FpvalCanvas1, _ = strconv.Atoi(fpval[0])
		sd.FpvalCanvas2, _ = strconv.Atoi(fpval[1])
		sd.FpvalFonts = fpval[2]
		sd.FpvalPluginInfo = fpval[3]
		sd.FpvalSessionStorage = fpval[4]
		sd.FpvalLocalStorage = fpval[5]
		sd.FpvalIndexedDb = fpval[6]
		sd.FpvalTzoffset, err = strconv.Atoi(fpval[7])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.FpvalWebtc = fpval[8]
		sd.ScreenColorDepth, err = strconv.Atoi(fpval[9])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.ScreenPixelDepth, err = strconv.Atoi(fpval[10])
		if err != nil {
		  return nil, errors.WithStack(err)
		}
		sd.NavigatorCookieEnabled = fpval[11]
		sd.NavigatorJavaEnabled = fpval[12]
		sd.NavigatorDoNotTrack, err = strconv.Atoi(fpval[13])
		if err != nil {
			sd.NavigatorDoNotTrack = -1
		}
	}
	sd.FpvalHash, err = strconv.Atoi(parts[22])
	if err != nil {
	  return nil, errors.WithStack(err)
	}

	sd.O9, err = strconv.Atoi(parts[23])
	if err != nil {
	  return nil, errors.WithStack(err)
	}

	sd.SdFirstpartHash, err = strconv.Atoi(parts[24])
	if err != nil {
	  return nil, errors.WithStack(err)
	}

	glfp := strings.Split(parts[25], ",")
	if len(glfp) > 5 {
		sd.FontHash = glfp[0]
		sd.DevicePixelRatio = glfp[1]
		sd.SpeechSynthVoices = glfp[2]
		sd.WebglVendor = glfp[3]
		sd.WebglRenderer = glfp[4]
		sd.WebglExtHash = glfp[5]
		sd.WebglExtLen = glfp[6]
	}
	last := strings.Split(parts[26], ";")[1:]
	sd.SdEndDt, err = strconv.Atoi(last[0])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.Tst, err = strconv.Atoi(last[1])
	if err != nil {
	  return nil, errors.WithStack(err)
	}
	sd.SdEndDt2, err = strconv.Atoi(last[2])
	if err != nil {
	  return nil, errors.WithStack(err)
	}

	// log.Println("")

	return &sd, nil
}

func ParseChalSoln(solnStr string) AkChalSoln {
	p := strings.Split(solnStr, ";")
	if len(p) != 5 {
		return AkChalSoln{}
	}
	return AkChalSoln {
		Al: strings.Split(p[0], ","),
		T: strings.Split(p[1], ","),
		Il: strings.Split(p[2], ","),
		Lg: strings.Split(p[3], ","),
	}
}

func AkAb(str string) int {
	s := 0
	for i, _ := range str {
		n := int(str[i])
		if n < 128 {
			s += n
		}
	}
	return s
}

func (sd *AkWebSensorData) Serialize() string {
	var b strings.Builder
	fmt.Fprintf(&b, `%s-1,2,-94,-100,%s,uaend,`, sd.Version, sd.Ua)
	group0 := []interface{}{
		sd.Xagg,
		sd.Psub,
		sd.Lang,
		sd.Prod,
		sd.Plen,
		sd.Pen,
		sd.Wen,
		sd.Den,
		sd.StartTs / 4064256,
		sd.D3,
		sd.ScreenAvailWidth,
		sd.ScreenAvailHeight,
		sd.ScreenWidth,
		sd.ScreenHeight,
		sd.WindowInnerWidth,
		sd.WindowInnerHeight,
		sd.WindowOuterWidth,
	}
	for _, g := range group0 {
		fmt.Fprintf(&b, `%v,`, g)
	}
	sd.StartTsDiv2 = fmt.Sprintf("%.1f", float64(sd.StartTs) / float64(2))
	if strings.HasSuffix(sd.StartTsDiv2, ".0") {
		sd.StartTsDiv2 = sd.StartTsDiv2[:len(sd.StartTsDiv2) - 2]
	}
	fmt.Fprintf(&b, `,cpen:%s,i1:%s,dm:%s,cwen:%s,non:%d,opc:%s,fc:%s,sc:%s,wrc:%d,isc:%s,vib:%d,bat:%d,x11:%s,x12:%d,%d,%s,%s,%s,loc:%s`,
		sd.Cpen, sd.I1, sd.Dm, sd.Cwen, sd.Non, sd.Opc, sd.Fc, sd.Sc, sd.Wrc, sd.Isc, sd.Vib, sd.Bat, sd.X11, sd.X12, sd.UaHash, sd.MathRandom, sd.StartTsDiv2, sd.IsBrave, sd.Loc)

	fmt.Fprintf(&b, "-1,2,-94,-101,%s,%s,%s-1,2,-94,-105,-1,2,-94,-102,-1,2,-94,-108,", sd.HasWindowDeviceOrientationEvent, sd.HasWindowDeviceMotionEvent, sd.HasWindowTouchEvent)
	// fmt.Fprintf(&b, "-1,2,-94,-101,%s,%s,%s-1,2,-94,-105,-1,2,-94,-102,0,-1,0,1,1384,520,0;-1,2,-94,-108,", sd.HasWindowDeviceOrientationEvent, sd.HasWindowDeviceMotionEvent, sd.HasWindowTouchEvent)

	for _, k := range sd.Kact {
		fmt.Fprintf(&b, "%d,%s,%d,%s,%s,%s,%s;", k.KeCnt, k.Typ, k.Dt, k.KeycodeClass, k.L, k.Specialkeys, k.InpEleIDHash)
	}
	fmt.Fprintf(&b, "-1,2,-94,-110,")
	for _, m := range sd.Mact {
		fmt.Fprintf(&b, "%d,%s,%d,%d,%d", m.MeCnt, m.Typ, m.Dt, m.X, m.Y)
		if m.EleIDHash != nil {
			fmt.Fprintf(&b, ",%v", m.EleIDHash)
		}
		if m.MouseBtn != nil {
			fmt.Fprintf(&b, ",%v", m.MouseBtn)
		}
		fmt.Fprintf(&b, ";")
	}
	fmt.Fprintf(&b, "-1,2,-94,-117,-1,2,-94,-111,")
	for _, d := range sd.Doact {
		fmt.Fprintf(&b, "%d,%d,%s,%s,%s;", d.DoeCnt, d.Dt, d.Alpha, d.Beta, d.Gamma)
	}
	fmt.Fprintf(&b, "-1,2,-94,-109,")
	for _, d := range sd.Dmact {
		fmt.Fprintf(&b, "%d,%d,%s,%s,%s,%s,%s,%s,%s,%s,%s;", d.DmeCnt, d.Dt, d.Ax, d.Ay, d.Az, d.Agx, d.Agy, d.Agz, d.Ra, d.Rb, d.Rg)
	}
	fmt.Fprintf(&b, "-1,2,-94,-114,-1,2,-94,-103,-1,2,-94,-112,%s-1,2,-94,-115,", sd.URL)
	group1 := []interface{}{
		sd.KeVel + 1,
		sd.MeVel + 32,
		sd.TeVel + 32,
		sd.DoeVel,
		sd.DmeVel,
		sd.PeVel,
		sd.VelSum,
		sd.SdDt,
		sd.InitTime,
		sd.StartTs,
		sd.FpvalTd,
		(sd.StartTs / 4064256) / 23,
		sd.KeCnt,
		sd.MeCnt,
		sd.L,
		sd.PeCnt,
		sd.TeCnt,
		sd.SdDt2,
		sd.Ta,
		sd.NoCookie,
		string(bytes.Trim([]byte(sd.Cookie), "\x00")),
		sd.CookieHash,
		sd.CanvasTeststr,
		sd.CanvasHash,
		sd.NavigatorFpFas,
		sd.StartTsJig,
		sd.DumbassPipeCharcode,
	}
	// fmt.Printf("%+v\n", sd.StartTsJig)
	// fmt.Printf("%+v\n", sd.DumbassPipeCharcode)
	for _, v := range group1 {
		fmt.Fprintf(&b, "%v,", v)
	}
	fmt.Fprintf(&b, "%v-1,2,-94,-106,%v,%v-1,2,-94,-119,", sd.AnotherOne, sd.AjType, sd.AjIndx)
	if sd.MathAbsT == -1 {
		fmt.Fprintf(&b, "-1")
	} else {
		group2 := []int {
			sd.MathAbsT,
			sd.MathAcosT,
			sd.MathAsinT,
			sd.MathAtanhT,
			sd.MathCbrtT,
			sd.MathExpT,
			sd.MathRandomT,
			sd.MathRoundT,
			sd.MathSqrtT,
			sd.IsfiniteT,
			sd.IsnanT,
			sd.ParsefloatT,
			sd.ParseintT,
			sd.JsonparseT,
		}
		for _, v := range group2 {
			fmt.Fprintf(&b, "%d,", v)
		}
	}
	fmt.Fprintf(&b, "-1,2,-94,-122,0,0,0,0,1,0,0-1,2,-94,-123,")
	if len(sd.Chal0Soln.Al) > 0 {
		fmt.Fprintf(&b, "%s;%s;%s;%s;",
			strings.Join(sd.Chal0Soln.Al, ","),
			strings.Join(sd.Chal0Soln.T, ","),
			strings.Join(sd.Chal0Soln.Il, ","),
			strings.Join(sd.Chal0Soln.Lg, ","))
	}
	fmt.Fprintf(&b, "-1,2,-94,-124,")
	if len(sd.Chal1Soln.Al) > 0 {
		fmt.Fprintf(&b, "%s;%s;%s;%s;",
			strings.Join(sd.Chal1Soln.Al, ","),
			strings.Join(sd.Chal1Soln.T, ","),
			strings.Join(sd.Chal1Soln.Il, ","),
			strings.Join(sd.Chal1Soln.Lg, ","))
	}
	fmt.Fprintf(&b, "-1,2,-94,-126,")
	if len(sd.Chal2Soln.Al) > 0 {
		fmt.Fprintf(&b, "%s;%s;%s;%s;",
			strings.Join(sd.Chal2Soln.Al, ","),
			strings.Join(sd.Chal2Soln.T, ","),
			strings.Join(sd.Chal2Soln.Il, ","),
			strings.Join(sd.Chal2Soln.Lg, ","))
	}
	fmt.Fprintf(&b, "-1,2,-94,-127,%s", sd.Perms)

	sd0 := b.String()
	sd0_hash := 24 ^ AkAb(sd0)

	b = strings.Builder{}
	fmt.Fprintf(&b, "%s%s-1,2,-94,-70,", sd.Hash, sd0)
	if sd.FpvalCanvas1 != 0 {
		group3 := []interface{}{
			sd.FpvalCanvas1,
			sd.FpvalCanvas2,
			sd.FpvalFonts,
			sd.FpvalPluginInfo,
			sd.FpvalSessionStorage,
			sd.FpvalLocalStorage,
			sd.FpvalIndexedDb,
			sd.FpvalTzoffset,
			sd.FpvalWebtc,
			sd.ScreenColorDepth,
			sd.ScreenPixelDepth,
			sd.NavigatorCookieEnabled,
			sd.NavigatorJavaEnabled,
		}
		for _, v := range group3 {
			fmt.Fprintf(&b, "%v;", v)
		}
	} else {
		fmt.Fprintf(&b, "-1")
	}

	var glfpb strings.Builder
	var dnt string
	if sd.FontHash != "" {
		group4 := []interface{}{
			sd.FontHash,
			sd.DevicePixelRatio,
			sd.SpeechSynthVoices,
			sd.WebglVendor,
			sd.WebglRenderer,
			sd.WebglExtHash,
			sd.WebglExtLen,
			// cf0af17f2d8432e085203b3da9e6a6369155dd3d3662aec3741e0d7431181ce9,
			// 3,
			// ,
			// Apple Inc.,
			// Apple GPU,
			// d4a16b46adbc0f5037f39cea633c2576fb46695b25d31400161da7d124616cc3,
			// 16
		}
		for _, v := range group4[:len(group4)-1] {
			fmt.Fprintf(&glfpb, "%v,", v)
		}
		fmt.Fprintf(&glfpb, "%v", group4[len(group4) - 1])
		dnt = fmt.Sprintf("%d", sd.NavigatorDoNotTrack)
		if dnt == "-1" {
			dnt = "1"
		}
	}

	fmt.Fprintf(&b, "%s-1,2,-94,-80,%d-1,2,-94,-116,%d-1,2,-94,-118,%d-1,2,-94,-129,%s-1,2,-94,-121,;%d;%d;%d", dnt, sd.FpvalHash, sd.O9, sd0_hash, glfpb.String(), sd.SdEndDt, sd.Tst, sd.SdEndDt2)

	return b.String()
}

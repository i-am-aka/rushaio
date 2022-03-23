package task

type BmpSensorData struct {
  Version                 string        `json:"version"`
  DeviceHeight            string        `json:"device_height"`
  DeviceWidth             string        `json:"device_width"`
  BatteryReceiverFlag     string        `json:"battery_receiver_flag"`
  FirebaseAnalyticsLevel  string        `json:"firebase_analytics_level"`
  AppConfigOrientation    string        `json:"app_config_orientation"`
  Language                string        `json:"language"`
  BuildVersionRelease     string        `json:"build_version_release"`
  AcceleromerRotation     string        `json:"acceleromer_rotation"`
  BuildModel              string        `json:"build_model"`
  BuildBootloader         string        `json:"build_bootloader"`
  BuildHardware           string        `json:"build_hardware"`
  Package                 string        `json:"package"`
  AndroidID               string        `json:"android_id"`
  HasKeyboard             string        `json:"has_keyboard"`
  IsAdbEnabled            string        `json:"is_adb_enabled"`
  BuildVersionCodename    string        `json:"build_version_codename"`
  BuildVersionIncremental string        `json:"build_version_incremental"`
  BuildVersionSdkInt      string        `json:"build_version_sdk_int"`
  BuildManufacturer       string        `json:"build_manufacturer"`
  BuildProduct            string        `json:"build_product"`
  BuildTags               string        `json:"build_tags"`
  BuildType               string        `json:"build_type"`
  BuildUser               string        `json:"build_user"`
  BuildDisplay            string        `json:"build_display"`
  BuildBoard              string        `json:"build_board"`
  BuildBrand              string        `json:"build_brand"`
  BuildDevice             string        `json:"build_device"`
  BuildFingerprint        string        `json:"build_fingerprint"`
  BuildHost               string        `json:"build_host"`
  BuildID                 string        `json:"build_id"`
  DevinfoHash             string        `json:"devinfo_hash"`
  Randint                 string        `json:"randint"`
  T0WallDiv2              int64         `json:"t0_wall_div2"`
  DoEn                    string        `json:"do_en"`
  DmEn                    string        `json:"dm_en"`
  TEn                     string        `json:"t_en"`
  TouchEvents             []interface{} `json:"touch_events"`
  OrientEvents            []struct {
    Dt         float64 `json:"dt"`
    Yaw        float64 `json:"yaw"`
    Pitch      float64 `json:"pitch"`
    Roll       float64 `json:"roll"`
    UsuallyOne float64 `json:"usually_one"`
  } `json:"orient_events"`
  MotionEvents []struct {
    Dt         float64 `json:"dt"`
    Dax        float64 `json:"dax"`
    Day        float64 `json:"day"`
    Daz        float64 `json:"daz"`
    Ax         float64 `json:"ax"`
    Ay         float64 `json:"ay"`
    Az         float64 `json:"az"`
    Gax        float64 `json:"gax"`
    Gay        float64 `json:"gay"`
    Gaz        float64 `json:"gaz"`
    UsuallyOne float64 `json:"usually_one"`
  } `json:"motion_events"`
  OrientQuant struct {
    Min        string `json:"min"`
    Max        string `json:"max"`
    SampleHash string `json:"sampleHash"`
    QuantStr   string `json:"quantStr"`
  } `json:"orient_quant"`
  OrientQuant2 []struct {
    Min        string `json:"min"`
    Max        string `json:"max"`
    SampleHash string `json:"sampleHash"`
    QuantStr   string `json:"quantStr"`
  } `json:"orient_quant2"`
  MotionQuant struct {
    Min        string `json:"min"`
    Max        string `json:"max"`
    SampleHash string `json:"sampleHash"`
    QuantStr   string `json:"quantStr"`
  } `json:"motion_quant"`
  MotionQuant2 []struct {
    Min        string `json:"min"`
    Max        string `json:"max"`
    SampleHash string `json:"sampleHash"`
    QuantStr   string `json:"quantStr"`
  } `json:"motion_quant2"`
  TouchDtSum             string `json:"touch_dt_sum"`
  OrientQuantSum         int64  `json:"orient_quant_sum"`
  MotionQuantSum         int64  `json:"motion_quant_sum"`
  TouchOrientMotionSum   int64  `json:"touch_orient_motion_sum"`
  SensorDt               string `json:"sensor_dt"`
  TextChangeCount        string `json:"text_change_count"`
  VP1000                 string `json:"v.p*1000"`
  SensorUptimeDt         string `json:"sensor_uptime_dt"`
  HasOrientationListener string `json:"has_orientation_listener"`
  EvtHash                string `json:"evt_hash"`
  StartTs                int64  `json:"start_ts"`
  Perf0                  string `json:"perf_0"`
  Perf1                  string `json:"perf_1"`
  Perf2                  string `json:"perf_2"`
  Perf3                  string `json:"perf_3"`
  Perf4                  string `json:"perf_4"`
  Perf5                  string `json:"perf_5"`
  Perf6                  string `json:"perf_6"`
  Perf7                  string `json:"perf_7"`
  Perf8                  string `json:"perf_8"`
  Lifecycle              []struct {
    Ts   int64  `json:"ts"`
    Type string `json:"type"`
  } `json:"lifecycle"`
}
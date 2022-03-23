package task

type DnsOverHttpResp struct {
  Status   int  `json:"Status"`
  TC       bool `json:"TC"`
  RD       bool `json:"RD"`
  RA       bool `json:"RA"`
  AD       bool `json:"AD"`
  CD       bool `json:"CD"`
  Question []struct {
    Name string `json:"name"`
    Type int    `json:"type"`
  } `json:"Question"`
  Answer []struct {
    Name string `json:"name"`
    Type int    `json:"type"`
    TTL  int    `json:"TTL"`
    Data string `json:"data"`
  } `json:"Answer"`
}
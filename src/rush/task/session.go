package task

import (
	// "fmt"
	"strconv"
	"rush/net/http/cookiejar"
	"rush/common"
	"github.com/google/uuid"
)
// import "time"

type RushTaskSessionStats struct {
	Count uint `json:"count"`
	FarmMax uint `json:"max"`
}

var SESSION_CKIE_BLACKLIST = []string{
	"cart-guid",
	"JSESSIONID",
	"waiting_room",
	"CRT",
	"hasCRT",
	"hasPCID",
	"PCID",
}

func (t *CheckoutTask) TaskSessionSnapshot() TaskSession {
	icookies := map[string]map[string]interface{}{}
	for k, v := range t.client.Jar.GetIEntriesMapCopy() {
		ev := map[string]interface{}{}
		for kk, vv := range v {
			if e, ok := vv.(cookiejar.Entry); ok {
				if common.Any(SESSION_CKIE_BLACKLIST, func(s string) bool { return s == e.Name }) {
					continue
				}
				if n, err := strconv.Atoi(e.Name); err == nil{
					if n >= 50 {
						ev[kk] = vv
					}
				} else {
					ev[kk] = vv
				}
			}
		}
		icookies[k] = ev
	}
	idd := uuid.New().String()
	if t.LastSessionSnapshot != nil {
		idd = (*t.LastSessionSnapshot).Id
	}
	return TaskSession{
		Id: idd,
		StartTime: t.StartTime,
		Duration: timeMillis() - t.StartTime,
		TaskId: t.Id,
		Version: t.Version,
		Host: t.Url.Host,
		Proxy: t.ProxyStr,
		UserAgent: t.UserAgent,
		RealIp: t.RealIp,
		Cookies: icookies,
		Metrics: t.Metrics.Copy(),
	}
}

type TaskSession struct {
	Id string `json:"id"`
	StartTime uint64 `json:"start_time"`
	Duration uint64 `json:"duration"`
	TaskId string `json:"task_id"`
	Version string `json:"version"`
	Host string `json:"host"`
	Proxy string `json:"proxy"`
	UserAgent string `json:"user_agent"`
	RealIp string `json:"real_ip"`
	Cookies map[string]map[string]interface{} `json:"cookies"`
	Metrics map[string]interface{} `json:"metrics"`
}

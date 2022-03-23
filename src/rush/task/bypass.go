package task

import (
	// "log"
	"net/url"
	"rush/net/http"
	"strings"
	"time"
)

func (t *CheckoutTask) hasAkavpau() bool {
	for _, c := range t.client.Jar.Cookies(t.Url)[:] {
		if strings.HasPrefix(c.Name, "akavpau") {
			return true
		}
	}
	return false
}

func (t *CheckoutTask) ftlHandleQueue() bool {
	// name := "akavpwr_FL"
 //  qckiebytes, err := t.ApiGet("/qckie?name=" + name)
 //  qckie := strings.TrimSpace(string(qckiebytes))
	// // TODO REENABLE
 //  if err == nil && len(qckie) > 0 {
 //  	t.FtlQueueBypass(name, qckie)
 //  	return false
 //  } else {
 //  return true
 //  }

	for _, c := range t.client.Jar.Cookies(t.Url)[:] {
	  // fmt.Println(c.Name)
	  if strings.Contains(c.Name, "akavpwr") {
	    t.StatusCh <- StatusInQueue
			t.SetCookie(&http.Cookie{ Name: c.Name, MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "", })
			time.Sleep(10 * time.Second)
			// time.Sleep(10 * time.Second)
			return true
	  }
	}
	return false
}

func (t *CheckoutTask) GenericQueueBypass() bool {
	var passed = false
	for _, c := range t.client.Jar.Cookies(t.Url)[:] {
	  if strings.Contains(c.Name, "akavpwr") {
			t.client.Jar.SetCookies(t.Url,
				[]*http.Cookie{
					&http.Cookie{ Name: strings.Replace(c.Name, "akavpwr", "akavpau", -1), Value: c.Value, Path: "/", },
					&http.Cookie{ Name: c.Name, MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "", },
				},
			)
	  	passed = true
	  }
	}

	return passed
}

func (t *CheckoutTask) GenericQueueBypassUrl(url_ *url.URL) bool {
	var passed = false
	for _, c := range t.client.Jar.Cookies(url_)[:] {
		// log.Printf("%+v\n", c)
	  if strings.Contains(c.Name, "akavpwr") {
			t.client.Jar.SetCookies(url_,
				[]*http.Cookie{
					&http.Cookie{ Name: strings.Replace(c.Name, "akavpwr", "akavpau", -1), Value: c.Value, Path: "/", },
					&http.Cookie{ Name: c.Name, MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "", },
				},
			)
	  	passed = true
	  }
	}

	return passed
}


func (t *CheckoutTask) FtlQueueBypass(name string, passval string) {
	passckiename := strings.Replace(name, "akavpwr", "akavpau", -1)
	t.client.Jar.SetCookies(t.Url,
		[]*http.Cookie{
			&http.Cookie{ Name: passckiename, Value: passval, Path: "/", },
			&http.Cookie{ Name: name, MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "", },
		},
	)
	// for _, c := range t.client.Jar.Cookies(t.Url)[:] {
	//   if strings.Contains(c.Name, "akavpwr") {
	//   }
	// }
}

func (t *CheckoutTask) FnlQueueBypass(url_ *url.URL) {
	queueCkieValue := getCookieValue("akavpwr_FL-MobApp", url_, t.client.Jar)
	if queueCkieValue != ""{
		t.client.Jar.SetCookies(url_,
			[]*http.Cookie{
				&http.Cookie{ Name: "akavpau_FL-MobApp", Value: queueCkieValue, Path: "/", },
				&http.Cookie{ Name: "akavpwr_FL-MobApp", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "", },
			},
		)
	}

	queueCkieValue = getCookieValue("akavpwr_FL-MobApp-Product", url_, t.client.Jar)
	if queueCkieValue != ""{
		t.client.Jar.SetCookies(url_,
			[]*http.Cookie{
				&http.Cookie{ Name: "akavpau_FL-MobApp-Product", Value: queueCkieValue, Path: "/", },
				&http.Cookie{ Name: "akavpwr_FL-MobApp-Product", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "", },
			},
		)
	}
}
// t.StatusCh <- Status{Message:fmt.Sprintf("Bypassing Queue (%s)", description)}
// queueCkieValue := getCookieValue("akavpwr_FL-MobApp", url_, t.client.Jar)
// if queueCkieValue != ""{
// 	t.client.Jar.SetCookies(url_,
// 		[]*http.Cookie{
// 			&http.Cookie{ Name: "akavpau_FL-MobApp", Value: queueCkieValue, Path: "/", },
// 			&http.Cookie{ Name: "akavpwr_FL-MobApp", MaxAge: -1, Expires: time.Unix(0,0), Secure: true, Path: "/", Value: "", },
// 		},
// 	)
// }
// go t.SendTelemetry(map[string]interface{}{ "event": "queue_bypass", "cookies": t.client.Jar.Cookies(url_)})

package main

import (
	"log"
	"net/url"
	// "time"
	"net/http"
	"fmt"
	"os"
	"strings"
 	"github.com/stripe/stripe-go/v71"
 	"io/ioutil"
 	"github.com/stripe/stripe-go/v71/checkout/session"
 	"github.com/avast/retry-go"
 	"sync"
 	"github.com/gorilla/handlers"
 	// "gopkg.in/ezzarghili/recaptcha-go.v4"
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

var maxCodes = 400
var codes = 0
var codeLock sync.Mutex
func http_handler(w http.ResponseWriter, req *http.Request) {
	enableCors(&w)
	if (*req).Method == "OPTIONS" {
		return
	}

	phone := req.URL.Query().Get("phone")
	phone = strings.Replace(phone, " ", "", -1)
	phone = strings.Replace(phone, "-", "", -1)
	if phone == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	log.Println(phone)
	// if phone[0] != "+" {
	// 	phone = "+1"
	// }
	msgData := url.Values{}
	msgData.Set("To",phone)
	msgData.Set("Channel","sms")
	msgDataReader := *strings.NewReader(msgData.Encode())

	accountSid := "ACdf37b9ba1b27797212283adf9602afc0"
	authToken := ""
	urlStr := "https://verify.twilio.com/v2/Services/VA378b4c0647397b8cbeac3ed419d85323/Verifications"

	// var urlStr = https://verify.twilio.com/v2/Services/VAXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX/Verifications
  client := &http.Client{}
  treq, _ := http.NewRequest("POST", urlStr, &msgDataReader)
  treq.SetBasicAuth(accountSid, authToken)
  treq.Header.Add("Accept", "application/json")
  treq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

  codeLock.Lock()
  codes += 1
  if codes > maxCodes {
  	codeLock.Unlock()
  	http.Error(w, "error", http.StatusBadRequest)
  	return
  }
  codeLock.Unlock()
  resp, err := client.Do(treq)
  if err != nil {
  	http.Error(w, "error", http.StatusBadRequest)
  	return
  }
  bb, err := ioutil.ReadAll(resp.Body)
  log.Println(string(bb))
  fmt.Fprintf(w, string(bb))
}

func http_handler_bak(w http.ResponseWriter, req *http.Request) {
	enableCors(&w)
	if (*req).Method == "OPTIONS" {
		return
	}

	email := req.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	price := req.URL.Query().Get("price")
	if price == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	key := strings.ToLower(req.URL.Query().Get("keyy"))
	if key == "" || key != os.Getenv("RUSH_PAY_KEY") {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	phone := req.URL.Query().Get("phone")
	if phone == ""  {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	code := req.URL.Query().Get("code")
	if code == ""  {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	msgData := url.Values{}
	msgData.Set("To",phone)
	msgData.Set("Code",code)
	msgDataReader := *strings.NewReader(msgData.Encode())
	accountSid := "ACdf37b9ba1b27797212283adf9602afc0"
	authToken := ""
	urlStr := "https://verify.twilio.com/v2/Services/VA378b4c0647397b8cbeac3ed419d85323/VerificationCheck"

	// var urlStr = https://verify.twilio.com/v2/Services/VAXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX/Verifications
  client := &http.Client{}
  treq, _ := http.NewRequest("POST", urlStr, &msgDataReader)
  treq.SetBasicAuth(accountSid, authToken)
  treq.Header.Add("Accept", "application/json")
  treq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

  resp, err := client.Do(treq)
  if err != nil {
  	http.Error(w, "error", http.StatusBadRequest)
  	return
  }
  bb, err := ioutil.ReadAll(resp.Body)
  if !strings.Contains(string(bb), "approved") {
	  http.Error(w, "bad token", http.StatusBadRequest)
	  return
  }
  s := email + "," + phone
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String("setup"),
		// LineItems: []*stripe.CheckoutSessionLineItemParams {
		// 	&stripe.CheckoutSessionLineItemParams {
		// 		Price: stripe.String(price),
		// 		Quantity: stripe.Int64(1),
		// 	},
		// },
	  SuccessURL: stripe.String(fmt.Sprintf("https://join.rushaio.co/success.html?price=%s", url.QueryEscape(price))),
	  CancelURL:  stripe.String("https://join.rushaio.co/"),
	  PaymentMethodTypes: stripe.StringSlice([]string{
	    "card",
	  }),

	  SetupIntentData: &stripe.CheckoutSessionSetupIntentDataParams{
	  	Description: &s,
	  },
	  // PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
	  // 	CaptureMethod: stripe.String("manual"),
	  // },
	}

	var sessionId string
	retry.Do(func() error {
		sesh, err := session.New(params)
		sessionId = sesh.ID
		return err
	})
	fmt.Fprintf(w, sessionId)
}

func main() {
	stripeKey := os.Getenv("STRIPE_KEY")
	if stripeKey == "" {
		stripeKey = "sk_test_AtJEFSWzg8MAweKxdyYDvZSV008nXpAS8D"
	}
	stripe.Key = stripeKey

	addr := ":8093"
	// srv := &http.Server{Addr: addr}
	log.Printf("Listening on %s", addr)
	r := http.NewServeMux()

	r.Handle("/session_botflippersbtfo", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(http_handler)))
	r.Handle("/session", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(http_handler_bak)))

	http.ListenAndServe(addr, handlers.CompressHandler(r))
	// if err := http.ListenAndServe(":8000", handlers.CompressHandler(r)); err != http.ErrServerClosed {
	//     log.Fatalf("ListenAndServe(): %v", err)
	// }
}

package task

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"math/rand"
	"rush/net/http"
	"strings"
  "time"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
)

type HibAppSession struct {
	CheckoutTask
	SessionID  string `json:"sessionId"`
	ExpDate    string `json:"expDate"`
  ExpTimeUnix    int64
	CustomerID string `json:"customerId"`
	BasketId string `json:"basketId"`
  Px3 string
  Cart *HibAppCart
}

func (t *CheckoutTask) HibAppCheckout() error {
	// t.UpdateDnsCache("hibbett-mobileapi.prolific.io")
  var err error
	rootPid := HibGetRootPid(t.Url)
	if rootPid == "" {
		return ErrInvalidUrl
	}

  t.UserAgent, err = t.GetPxMobileUserAgent()
  if err != nil {
    return err
  }

  ckies, err := t.GetPxCookieApi(Px3Request{
    Proxy: "", // todo use prox
    Cookie: t.CookiesStr(),
    Url: "https://www.perimeterx.com/",
    UserAgent: "PerimeterX Android SDK/1.9.2",
    // TODO cookies. can regenning help avoid 8min limit on checkout poll?
    AppId: "px9qx3rve4",
    Vid: "",
    Uuid: "",
    Host: "",
    JsSrc: "https://client.perimeterx.net/PX9Qx3Rve4/main.min.js",
    CapJsSrc: "",
  })
  log.Printf("%+v (error=%+v)", ckies, err)
  if err != nil {
    return err
  }
  var px3 string
  for _, ckie := range ckies {
    if ckie.Name == "_px3" {
      px3 = ckie.Value
      break
    }
  }
  t.StatusCh <- StatusStartingSession
	sesh, err := t.HibAppStartSession("1", px3)
	if err != nil {
		return err
	}
  sesh.Px3 = px3
	log.Printf("%+v %+v", sesh, err)

  t.StatusCh <- StatusPrecarting
	err = sesh.CreateCart()
	if err != nil {
		return err
	}

	product, err := sesh.HibPidLookup(rootPid, "3")
	if err != nil {
		return err
	}
  skus := product.Skus
	log.Printf("%+v %+v", skus, err)
	sku := t.HibChooseSku(skus, t.Sizes, HibGetColorId(t.Url), false)
	if sku == nil {
		return retry.Unrecoverable(errors.New("Size Unavailable"))
	}
  log.Printf("%+v", sku)

  _, err = sesh.HibSaveForLater(sku.Pid)
  if err != nil {
    return err
  }

  // TODO more (confirmed instock) precart skus
	cart, err := sesh.HibAppAddToCart(rootPid, "34669671")// sku.Pid)
	if err != nil {
		return err
	}
  sesh.Cart = cart
	log.Printf("%+v %+v", cart, err)

  t.StatusCh <- StatusSubmittingInfo
	err = sesh.SetEmail()
	if err != nil {
		return err
	}

	err = sesh.SetShippingAddress()
	if err != nil {
		return err
	}

  cart, err = sesh.HibSetShipOptions()
  if err != nil {
    return err
  }
  sesh.Cart = cart

  _, err = sesh.SetPaymentMethod()
  if err != nil {
    return err
  }

  cart, err = sesh.SetBillingAddress()
  if err != nil {
    return err
  }

  precartItemId := cart.CartItems[0].ID

  _, err = sesh.HibDeleteItemFromCart(precartItemId)
  if err != nil {
    return err
  }

  err = t.Retry(func() error {
    cart, err = sesh.HibAppAddToCart(rootPid, sku.Pid)
    if cart != nil && cart.Code == 1000 {
      t.StatusCh <- StatusWaitingForRestock
      return ErrRetrying
    }
    if err != nil {
      return err
    }

    t.StartCheckoutTime = timeMillis()
    t.StatusCh <- StatusOrdering
    order, err := sesh.PlaceOrder()
    log.Printf("%+v", order)
    whFields := [][2]string{
      {"Website", t.Url.Host + " (APP)"},
      {"Product", cart.CartItems[0].Product.Name},
      {"Size", cart.CartItems[0].Sku.Size},
      {"Color", cart.CartItems[0].Sku.Color.Label},
    }
    if order != nil && order.Code == 13020 {
      go t.DeclineWebhookFields(whFields)
      return nil
    } else if order != nil && order.Code == 0 {
      go t.SuccessWebhookFields(whFields)
      return nil
    } else if err != nil {
      return err
    } else {
      return ErrRetrying
    }
    return ErrRetrying
  }, retry.Delay(1*time.Second))

	return nil
}

func (t *CheckoutTask) HibAppInitSession() (*HibAppSession, error) {
  var err error
  sesh := HibAppSession{
    *t, "", "", 0,  "", "", "", nil,
  }
  sesh.DefaultProto = "HTTP/2"
  sesh.DefaultReqClose = true
  sesh.HeaderCaseFn = strings.ToLower
  sesh.client, err = sesh.newHttpClient()
  if err != nil {
    return nil, err
  }
  sesh.UpdateDnsCache("hibbett-mobileapi.prolific.io")
  sesh.UserAgent, err = sesh.GetPxMobileUserAgent()
  if err != nil {
    return nil, err
  }
  return &sesh, nil
}

func (t *CheckoutTask) HibAppStartSession(pxAuth string, px3 string) (*HibAppSession, error) {
  url_, err := url.Parse("https://hibbett-mobileapi.prolific.io/users/guest")
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-px-original-token",
    "x-api-key",
    "content-type",
    "version",
    "platform",
    "user-agent",
    "content-length",
    "accept-encoding",
  }
  headers := [][2]string {
    {"x-px-authorization", pxAuth},
    {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    {"content-type", "application/json"},
    {"version", "4.4.0"},
    {"content-length", "0"},
    {"platform", "android"},
    {"accept-encoding", "gzip"},
  }
  sesh, err := t.HibAppInitSession()
  if err != nil {
    return nil, err
  }
  if px3 != "" {
    sesh.Px3 = px3
    headers = append(headers, [2]string{"x-px-original-token", "3:" + px3})
  }
  resp, err := sesh.doReq(sesh.client, sesh.makeReq("POST", url_, &headers, &headerOrder, nil))
  if err != nil {
  	DiscardResp(resp)
  	return nil, err
  }
  err = readRespJsonDst(resp, &sesh)
  if sesh.ExpDate != "" {
    // 2020-08-13T16:51:24-04:00"
    if exptime, err := time.Parse("2006-01-02T15:04:05-07:00", sesh.ExpDate); err == nil {
      sesh.ExpTimeUnix = exptime.Unix()
      log.Printf("SESH EXP %+v %+v", sesh.ExpTimeUnix, sesh.ExpTimeUnix - time.Now().Unix())
    }
  }
  // sesh.CreateCart()
  return sesh, err
}

func (t *HibAppSession) SetBillingAddress() (*HibAppCart, error) {
  url_, err := url.Parse(fmt.Sprintf("https://hibbett-mobileapi.prolific.io/ecommerce/cart/%s/billing_address?useAsShipping=false", t.BasketId))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-api-key",
    "version",
    "platform",
    "user-agent",
    "authorization",
    "content-type",
    "content-length",
    "accept-encoding",
  }
  headers := [][2]string {
    {"x-px-authorization", t.Px3},
    {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    {"version", "4.4.0"},
    {"platform", "android"},
    {"authorization", t.Authorization()},
    {"content-type", "application/json; charset=UTF-8"},
    {"accept-encoding", "gzip"},
  }
  bodyStructure := map[string]interface{} {
  "address1": t.Profile.BillingAddress.Address1,
  "address2": t.Profile.BillingAddress.Address2,
  "city": t.Profile.BillingAddress.City,
  "country": t.Profile.BillingAddress.Country,
  "firstName": t.Profile.BillingAddress.FirstName,
  "lastName": t.Profile.BillingAddress.LastName,
  "save": true,
  "state": t.Profile.BillingAddress.State,
  "zip": t.Profile.BillingAddress.Zip,
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  resp, err := t.doReq(t.client, t.makeReq("PUT", url_, &headers, &headerOrder, &body))
  if err != nil {
    DiscardResp(resp)
    return nil, err
  }
  var cart HibAppCart
  err = readRespJsonDst(resp, &cart)
  return &cart, err
}

func (t *HibAppSession) HibSetShipOptions() (*HibAppCart, error) {
  url_, err := url.Parse(fmt.Sprintf("https://hibbett-mobileapi.prolific.io/ecommerce/cart/%s/shipments/me/shipping_options", t.BasketId))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-api-key",
    "version",
    "platform",
    "user-agent",
    "authorization",
    "content-type",
    "content-length",
    "accept-encoding",
  }
  headers := [][2]string {
    {"x-px-authorization", t.Px3},
    {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    {"version", "4.4.0"},
    {"platform", "android"},
    {"authorization", t.Authorization()},
    {"content-type", "application/json; charset=UTF-8"},
    {"accept-encoding", "gzip"},
  }
  bodyStructure := map[string]interface{} {
  "id": "ANY_GND",
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  resp, err := t.doReq(t.client, t.makeReq("PUT", url_, &headers, &headerOrder, &body))
  if err != nil {
    DiscardResp(resp)
    return nil, err
  }
  var cart HibAppCart
  err = readRespJsonDst(resp, &cart)
  return &cart, err
}

type HibAppOrderResponse struct {
  Code int `json:"code"`
  Message string `json:"message"`
}

func (t *HibAppSession) PlaceOrder() (*HibAppOrderResponse, error) {
  url_, err := url.Parse(fmt.Sprintf("https://hibbett-mobileapi.prolific.io/ecommerce/cart/%s/place_order?customerId=%s", t.BasketId, t.CustomerID))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-api-key",
    "content-type",
    "version",
    "platform",
    "user-agent",
    "authorization",
    "content-length",
    "accept-encoding",
  }
  headers := [][2]string {
    {"x-px-authorization", t.Px3},
    {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    {"content-type", "application/json"},
    {"version", "4.4.0"},
    {"platform", "android"},
    {"authorization", t.Authorization()},
    {"accept-encoding", "gzip"},
  }
  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, nil))
  if resp != nil {
    var orderResp HibAppOrderResponse
    err = readRespJsonDst(resp, &orderResp)
    return &orderResp, err
  }
  return nil, err
}

type HibAppCartId struct {
	BasketId string `json:"basketId"`
}

func (t *HibAppSession) CreateCart() (error) {
  url_, err := url.Parse("https://hibbett-mobileapi.prolific.io/ecommerce/cart/create")
  if err != nil {
    return err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-api-key",
    "content-type",
    "version",
    "platform",
    "user-agent",
    "content-length",
    "authorization",
    "content-length",
    "accept-encoding",
  }
  headers := [][2]string {
    {"x-px-authorization", t.Px3},
    {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    {"content-type", "application/json"},
    {"version", "4.4.0"},
    {"platform", "android"},
    {"authorization", fmt.Sprintf("Bearer %s", t.SessionID)},
    {"accept-encoding", "gzip"},
    {"content-length", "0"},
  }
  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, nil))
  if err != nil {
  	DiscardResp(resp)
  	return err
  }
  var cart HibAppCartId
  if err = readRespJsonDst(resp, &cart); err == nil {
  	t.BasketId = cart.BasketId
  }
  return err
}

func (t *HibAppSession) HibPidLookup(rootPid string, pxAuth string) (*HibProductApiResponse, error) {
	var product HibProductApiResponse
  path := fmt.Sprintf("hibbett-mobileapi.prolific.io/ecommerce/products/%s", rootPid)
  apipath := fmt.Sprintf("/skudb/%s", path)
  if respBytes, err := t.ApiGetRetries(apipath, 3); err == nil && len(respBytes) > 0 {
    if err = json.Unmarshal(respBytes, &product); err == nil {
      return &product, nil
    }
  }

  url_, err := url.Parse(fmt.Sprintf("https://%s", path))
  	// url_, err := url.Parse(fmt.Sprintf("https://hibbett-mobileapi.prolific.io/ecommerce/products/%s?customerId=%s", rootPid, t.CustomerID))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-api-key",
    "content-type",
    "version",
    "platform",
    "user-agent",
    "authorization",
    "if-none-match",
    "accept-encoding",
  }
  auth := t.Px3
  if pxAuth != "" {
    auth = pxAuth
  }
  headers := [][2]string {
    {"x-px-authorization", auth},
    // {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    // {"content-type", "application/json"},
    // {"version", "4.4.0"},
    // {"platform", "android"},
    // {"authorization", t.Authorization()},
    // {"accept-encoding", "gzip"},
  }
  log.Printf("%+v", rand.Intn(1))
  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
  if resp != nil {
  	respBytes, err := readBodyBytes(resp)

  	err = json.Unmarshal(respBytes, &product)
  	if err != nil {
  		rsp := string(respBytes)

  		if strings.Contains(rsp, "\"action\":\"captcha\"") {
  			if err != nil {
  				log.Printf("%+v")
  				t.StatusCh <- Status{Message: "PX Banned (PID Lookup)"}
  				return nil ,ErrPxBanned
  			}
  			// return retry.Unrecoverable(errPxBanned)
  		} else if strings.Contains(rsp, "<title>Please go to</title>") {
  			return nil, retry.Unrecoverable(ErrProxyNotLocatedInUsa)
  		} else if strings.Contains(rsp, "Request blocked.") {
  			t.StatusCh <- StatusProxyBanned
  			return nil, retry.Unrecoverable(ErrProxyBanned)
  		}
  		return nil, errPxBanned
  	} else {
  		for _, sku := range product.Skus[:] {
        s := &sku
  			s.SizeId, _ = HibSizeId(sku.Size)
  			// product.Skus[idx] = sku
  		}
      if product.LaunchDate != "" {
        if exptime, err := time.Parse("2006-01-02T15:04:05-07:00", product.LaunchDate); err == nil {
          product.ReleaseTime = exptime.Unix()
        }
      }
      if len(product.Skus) > 0 {
        // func (t *CheckoutTask) ApiPost(endpoint string, data []byte) ([]byte, error) {
        go t.ApiPost(apipath, respBytes)
      }
  		log.Printf("%+v", product)
  		return &product, nil
  	}
  }
  return nil, err
}

type HibAppCart struct {
  Code int `json:"code"`
  Message string `json:"message"`
	ID              string      `json:"id"`
	ItemCount       int         `json:"itemCount"`
	ItemTotalCount  int         `json:"itemTotalCount"`
	SubTotal        float64         `json:"subTotal"`
	Tax             interface{} `json:"tax"`
	TotalAdjustment float64         `json:"totalAdjustment"`
	Total           interface{} `json:"total"`
	CartItems       []struct {
		ID      string `json:"id"`
		Product struct {
			ID              string      `json:"id"`
			Name            string      `json:"name"`
			Price           float64         `json:"price"`
			DiscountedPrice interface{} `json:"discountedPrice"`
			ImageResources  struct {
				One0000 []struct {
					URL   string `json:"url"`
					Usage string `json:"usage"`
				} `json:"1000-0"`
				One0001 []struct {
					URL   string `json:"url"`
					Usage string `json:"usage"`
				} `json:"1000-1"`
				One0002 []struct {
					URL   string `json:"url"`
					Usage string `json:"usage"`
				} `json:"1000-2"`
				One0003 []struct {
					URL   string `json:"url"`
					Usage string `json:"usage"`
				} `json:"1000-3"`
				One0004 []struct {
					URL   string `json:"url"`
					Usage string `json:"usage"`
				} `json:"1000-4"`
				One0005 []struct {
					URL   string `json:"url"`
					Usage string `json:"usage"`
				} `json:"1000-5"`
				One0006 []struct {
					URL   string `json:"url"`
					Usage string `json:"usage"`
				} `json:"1000-6"`
			} `json:"imageResources"`
			ImageIds    []string `json:"imageIds"`
			Description string   `json:"description"`
		} `json:"product"`
		Quantity int `json:"quantity"`
		Sku      struct {
			ID                string      `json:"id"`
			IsAvailable       bool        `json:"isAvailable"`
			AvailableQuantity interface{} `json:"availableQuantity"`
			Price             float64         `json:"price"`
			DiscountedPrice   interface{} `json:"discountedPrice"`
			ImageIds          []string    `json:"imageIds"`
			Color             struct {
				ID           string `json:"id"`
				Label        string `json:"label"`
				ImagePattern struct {
					URL   string `json:"url"`
					Usage string `json:"usage"`
				} `json:"imagePattern"`
			} `json:"color"`
			Size           string      `json:"size"`
			GiftCardAmount interface{} `json:"giftCardAmount"`
		} `json:"sku"`
		Personalizations          []interface{} `json:"personalizations"`
		OtherSkusAvailable        bool          `json:"otherSkusAvailable"`
		Adjustments               []interface{} `json:"adjustments"`
		TotalPrice                float64           `json:"totalPrice"`
		TotalDiscountedPrice      interface{}   `json:"totalDiscountedPrice"`
		ShipmentID                string        `json:"shipmentId"`
		StoreID                   string        `json:"storeId"`
		PromotionDescription      interface{}   `json:"promotionDescription"`
		IsRaffle                  bool          `json:"isRaffle"`
		LocateQty                 int           `json:"locateQty"`
		StoreQty                  int           `json:"storeQty"`
		LaunchDate                interface{}   `json:"launchDate"`
		IsSDD                     bool          `json:"isSDD"`
		IsSTS                     bool          `json:"isSTS"`
		MasterID                  string        `json:"masterId"`
		AvailableForInStorePickup bool          `json:"availableForInStorePickup"`
		IsReserveItem             bool          `json:"isReserveItem"`
		IsAvailableInStoreOnly    bool          `json:"isAvailableInStoreOnly"`
		DropshipVendorNumber      interface{}   `json:"dropshipVendorNumber"`
	} `json:"cartItems"`
	CostUntilFreeShipping int           `json:"costUntilFreeShipping"`
	Adjustments           []interface{} `json:"adjustments"`
	BillingAddress        interface{}   `json:"billingAddress"`
	Shipments             []struct {
		ID              string      `json:"id"`
		ShippingAddress interface{} `json:"shippingAddress"`
		ShippingOption  interface{} `json:"shippingOption"`
	} `json:"shipments"`
	PaymentMethods []interface{} `json:"paymentMethods"`
	FlashMessages  []struct {
		Type    string      `json:"type"`
		Message string      `json:"message"`
		Details interface{} `json:"details"`
	} `json:"flashMessages"`
	CustomerInformation struct {
		Email interface{} `json:"email"`
	} `json:"customerInformation"`
}

func (t *HibAppSession) Authorization() string {
	return "Bearer " + t.SessionID
}

func (t *HibAppSession) HibAppAddToCart(rootPid string, sku string) (*HibAppCart, error) {
  url_, err := url.Parse(fmt.Sprintf("https://hibbett-mobileapi.prolific.io/ecommerce/cart/%s/items?skuIds=%s&customerId=%s", t.BasketId, sku, t.CustomerID))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-api-key",
    "version",
    "platform",
    "user-agent",
    "authorization",
    "content-type",
    "content-length",
    "accept-encoding",
  }
  headers := [][2]string {
    {"x-px-authorization", t.Px3},
    {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    {"version", "4.4.0"},
    {"platform", "android"},
    {"authorization", t.Authorization()},
    {"content-type", "application/json; charset=UTF-8"},
    {"accept-encoding", "gzip"},
  }
  bodyStructure := map[string]interface{} {
  "cartItems": []interface{} {
    map[string]interface{} {
          "customerId": t.CustomerID,
          "personalizations": []interface{} {

          },
          "product": map[string]interface{} {
            "id": rootPid,
            "isRaffle": false,
          },
          "quantity": 1,
          "sku": map[string]interface{} {
            "id": sku,
          },
        },
  },
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  if err != nil {
    DiscardResp(resp)
    return nil, err
  }
  var cart HibAppCart
  err = readRespJsonDst(resp, &cart)
  return &cart, err
}

func (t *HibAppSession) HibSaveForLater(pid string) (*http.Response, error) {
  url_, err := url.Parse(fmt.Sprintf("https://hibbett-mobileapi.prolific.io/ecommerce/wishlist/%s/items", t.CustomerID))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-api-key",
    "version",
    "platform",
    "user-agent",
    "authorization",
    "content-type",
    "content-length",
    "accept-encoding",
  }
  headers := [][2]string {
    {"x-px-authorization", t.Px3},
    {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    {"version", "4.4.0"},
    {"platform", "android"},
    {"authorization", t.Authorization()},
    {"content-type", "application/json; charset=UTF-8"},
    {"accept-encoding", "gzip"},
  }
  bodyStructure := map[string]interface{} {
  "productId": pid,
  "quantity": 1,
  "type": "product",
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  DiscardResp(resp)
  return nil, err
}

func (t *HibAppSession) HibDeleteItemFromCart(itemId string) (*http.Response, error) {
  url_, err := url.Parse(fmt.Sprintf("https://hibbett-mobileapi.prolific.io/ecommerce/cart/%s/items/%s", t.BasketId, itemId))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-api-key",
    "content-type",
    "version",
    "platform",
    "user-agent",
    "authorization",
    "accept-encoding",
  }
  headers := [][2]string {
    {"x-px-authorization", t.Px3},
    {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    {"content-type", "application/json"},
    {"version", "4.4.0"},
    {"platform", "android"},
    {"authorization", t.Authorization()},
    {"accept-encoding", "gzip"},
  }
  resp, err := t.doReq(t.client, t.makeReq("DELETE", url_, &headers, &headerOrder, nil))
  DiscardResp(resp)
  return nil, err
}

func (t *HibAppSession) HibDeleteProductFromWishlist(rootPid string) (*http.Response, error) {
  url_, err := url.Parse(fmt.Sprintf("https://hibbett-mobileapi.prolific.io/ecommerce/wishlist/%s/product/%s", t.CustomerID, rootPid))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-api-key",
    "content-type",
    "version",
    "platform",
    "user-agent",
    "authorization",
    "accept-encoding",
  }
  headers := [][2]string {
    {"x-px-authorization", t.Px3},
    {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    {"content-type", "application/json"},
    {"version", "4.4.0"},
    {"platform", "android"},
    {"authorization", t.Authorization()},
    {"accept-encoding", "gzip"},
  }
  resp, err := t.doReq(t.client, t.makeReq("DELETE", url_, &headers, &headerOrder, nil))
  DiscardResp(resp)
  return nil, err
}

func (t *HibAppSession) SetEmail() (error) {
  url_, err := url.Parse(fmt.Sprintf("https://hibbett-mobileapi.prolific.io/ecommerce/cart/%s/customer", t.BasketId))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-api-key",
    "version",
    "platform",
    "user-agent",
    "authorization",
    "content-type",
    "content-length",
    "accept-encoding",
  }
  headers := [][2]string {
    {"x-px-authorization", t.Px3},
    {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    {"version", "4.4.0"},
    {"platform", "android"},
    {"authorization", t.Authorization()},
    {"content-type", "application/json; charset=UTF-8"},
    {"accept-encoding", "gzip"},
  }
  bodyStructure := map[string]interface{} {
	  "email": t.Profile.Email,
	}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  resp, err := t.doReq(t.client, t.makeReq("PUT", url_, &headers, &headerOrder, &body))
  // todo verify resp customerInformation
  DiscardResp(resp)
  return err
}

func (t *HibAppSession) SetShippingAddress() (error) {
  url_, err := url.Parse(fmt.Sprintf("https://hibbett-mobileapi.prolific.io/ecommerce/cart/%s/shipments/me/shipping_address?useAsBilling=true", t.BasketId))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-api-key",
    "version",
    "platform",
    "user-agent",
    "authorization",
    "content-type",
    "content-length",
    "accept-encoding",
  }
  headers := [][2]string {
    {"x-px-authorization", t.Px3},
    {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    {"version", "4.4.0"},
    {"platform", "android"},
    {"authorization", t.Authorization()},
    {"content-type", "application/json; charset=UTF-8"},
    {"accept-encoding", "gzip"},
  }
  bodyStructure := map[string]interface{} {
  "address1": t.Profile.ShippingAddress.Address1,
  "address2": t.Profile.ShippingAddress.Address2,
  "city": t.Profile.ShippingAddress.City,
  "country": t.Profile.ShippingAddress.Country,
  "firstName": t.Profile.ShippingAddress.FirstName,
  "lastName": t.Profile.ShippingAddress.LastName,
  "phone": t.Profile.ShippingAddress.NordPhone(),
  "save": true,
  "state": t.Profile.ShippingAddress.State,
  "zip": t.Profile.ShippingAddress.Zip,
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  resp, err := t.doReq(t.client, t.makeReq("PUT", url_, &headers, &headerOrder, &body))
  DiscardResp(resp)
  return err
}

func (t *HibAppSession) SetPaymentMethod() (*http.Response, error) {
  url_, err := url.Parse(fmt.Sprintf("https://hibbett-mobileapi.prolific.io/ecommerce/cart/%s/payment_methods", t.BasketId))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "x-px-authorization",
    "x-api-key",
    "version",
    "platform",
    "user-agent",
    "authorization",
    "content-type",
    "content-length",
    "accept-encoding",
  }
  headers := [][2]string {
    {"x-px-authorization", t.Px3},
    {"x-api-key", "0PutYAUfHz8ozEeqTFlF014LMJji6Rsc8bpRBGB0"},
    {"version", "4.4.0"},
    {"platform", "android"},
    {"authorization", t.Authorization()},
    {"content-type", "application/json; charset=UTF-8"},
    {"accept-encoding", "gzip"},
  }

  t.StatusCh <- Status{Message: "Checkout 2/4: Encrypting Card"}
  _, radialToken, encryptedCvv, err := t.hibGetRadialFields(t.Profile.Card)
  if err != nil {
    log.Printf("%+v", err)
    return nil, retry.Unrecoverable(errors.New("Card Rejected Pre-Authorization"))
  }

  bodyStructure := map[string]interface{} {
  "amount": t.Cart.Total,
  "encryptedCVNValue": strings.Replace(encryptedCvv[1:], "|javascript_1_3_10", "", -1),
  "paymentObject": map[string]interface{} {
    "cardType": t.Profile.Card.HibCardType(),
    "creditCardToken": radialToken,
    "expirationMonth": t.Profile.Card.ExpMonth,
    "expirationYear": t.Profile.Card.ExpYear,
    "nameOnCard": t.Profile.Card.Name,
    "number": "************" + t.Profile.Card.Number[len(t.Profile.Card.Number) - 4:],
  },
  "type": "CREDIT_CARD",
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  resp, err := t.doReq(t.client, t.makeReq("POST", url_, &headers, &headerOrder, &body))
  DiscardResp(resp)
  return nil, err
}


package task

import (
	"github.com/avast/retry-go"
	"strings"
	"github.com/pkg/errors"
	"rush/net/http"
	"github.com/streadway/simpleuuid"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	crand "crypto/rand"
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

var IgApiUrl, _ = url.Parse("https://i.instagram.com/")
var IgLauncherUrl, _ = url.Parse("https://b.i.instagram.com/api/v1/launcher/sync/")
var IgLaunchTokenUrl, _ = url.Parse("https://b.i.instagram.com/api/v1/zr/token/result/")
var IgExperimentSyncUrl, _ = url.Parse("https://b.i.instagram.com/api/v1/qe/sync/")
var IgLoginUrl, _ = url.Parse("https://i.instagram.com/api/v1/accounts/login/")
var IgGraphUrl, _ = url.Parse("https://i.instagram.com/api/v1/wwwgraphql/ig/query/")
var FbPayTokenUrl, _ = url.Parse("https://secure.facebook.com/payments/generate_token")
var IgAtcUrl, _ = url.Parse("https://i.instagram.com/api/v1/commerce/bag/add/")

type IgSession struct {
	Task *CheckoutTask
	Username string
	Password string
	Uuid string
	DeviceId string
	SessionId string
	PhoneId string
	PwPubKey string
	PwKeyId int
	Authorization string
	WwwClaim string
}

func generateUuid() string {
	now := time.Now()
	uuid, _ := simpleuuid.NewTime(now)
	return uuid.String()
}

func randomHex(n int) string {
  bytes := make([]byte, n)
  rand.Read(bytes)
  return hex.EncodeToString(bytes)
}

func generateDeviceId() string {
	return "android-" + randomHex(8)
}

func (s *IgSession) Init() {
	s.Uuid = generateUuid()
	s.DeviceId = generateDeviceId()
	s.SessionId = generateUuid()
	s.PhoneId = generateUuid()
}

func (s *IgSession) SignRequestBody(body map[string]interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return []byte{}, err
	}
	h := hmac.New(sha256.New, IgSignatureKey)
	h.Write(jsonBody)
	hmacb64 := base64.StdEncoding.EncodeToString(h.Sum(nil))
	log.Println(hmacb64)

	params := [][2]string {
		// {"signed_body", hmacb64 + "." + string(jsonBody)},
		{"signed_body", "SIGNATURE." + string(jsonBody)},
		// {"ig_sig_key_version", IgKeyData["signature_key_version"]},
	}
	return urlEncodeParams(params), nil
}

func SumStrBytes(str string) string {
	sum := 0
	for _, c := range str {
		sum += int(c)
	}
	return fmt.Sprintf("2%d", sum)
}

func (s *IgSession) GetAdId() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s.Username)))
}

func (s *IgSession) EncPw() (string, error) {
	var encpw string
	key := make([]byte, 32)
	iv := make([]byte, 12)
	rand.Read(key)
	rand.Read(iv)
	pkeybytes, err := base64.StdEncoding.DecodeString(s.PwPubKey)
	if err != nil {
		return encpw, err
	}
	bb, _ := pem.Decode(pkeybytes)
	pkey, err := x509.ParsePKIXPublicKey(bb.Bytes)
	if err != nil {
		return encpw, err
	}
	rsaEnc, err := rsa.EncryptPKCS1v15(crand.Reader, pkey.(*rsa.PublicKey), key)
	if err != nil {
		return encpw, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return encpw, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return encpw, err
	}
	ts := time.Now().Unix()
	crypted := aesgcm.Seal(nil, iv, []byte(s.Password), []byte(fmt.Sprintf("%d", ts)))
	aescrypted := crypted[:len(s.Password)]
	tag := crypted[len(s.Password):]

	cb := []byte{byte(1), byte(s.PwKeyId)}
	cb = append(cb, iv...)
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(len(rsaEnc)))
	cb = append(cb, b...)
	cb = append(cb, rsaEnc...)
	cb = append(cb, tag...)
	cb = append(cb, aescrypted...)

	cb64 := base64.StdEncoding.EncodeToString(cb)

	return fmt.Sprintf("#PWD_INSTAGRAM:4:%d:%s", ts, cb64), nil
}

func (s *IgSession) Login() error {
	pwenc, err := s.EncPw()
	if err != nil {
		return err
	}
	loginBody, err := s.SignRequestBody(map[string]interface{} {
		"jazoest": SumStrBytes(s.PhoneId),
		"country_codes": "[{\"country_code\":\"1\",\"source\":[\"default\"]}]",
		"phone_id": s.PhoneId,
		"enc_password": pwenc,
		"_csrftoken": getCookieValue("csrftoken", IgApiUrl, s.Task.client.Jar),
		"username": s.Username,
		"adid": s.GetAdId(),
		"guid": s.Uuid,
		"device_id": s.DeviceId,
		"google_tokens": "[]",
		// "password": s.Password,
		"login_attempt_count": "0",
	})
	headers := s.BaseHeaders()
	headers = append(headers, [2]string{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"})
	resp, err := s.Task.doReq(s.Task.client, s.Task.makeReq("POST", IgLoginUrl, &headers, nil, &loginBody))
	if err != nil {
		if resp != nil && resp.StatusCode == 400 {
			j, _ := readRespJson(resp)
			return errors.New(fmt.Sprintf("Login denied: visit %s in your browser to resolve", j["challenge"].(map[string]interface{})["url"]))
		}
		return err
	}
	s.Authorization = resp.Header.Get("ig-set-authorization")

	return nil
}

func (s *IgSession) Launch() error {
	syncBody, err := s.SignRequestBody(map[string]interface{} {
		"id": s.Uuid,
		"server_config_retrieval": "1",
	})
	headers := s.BaseHeaders()
	headers = append(headers, [2]string{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"})
	_, err = s.Task.doReq(s.Task.client, s.Task.makeReq("POST", IgLauncherUrl, &headers, nil, &syncBody))
	if err != nil {
		return err
	}

	headers = s.BaseHeaders()
	turl := IgLaunchTokenUrl
	turl.RawQuery = string(urlEncodeParams([][2]string {
		{"device_id", s.DeviceId},
		{"token_hash", ""},
		{"custom_device_id", s.Uuid},
		{"fetch_reason", "token_expired"},
	}))
	_, err = s.Task.doReq(s.Task.client, s.Task.makeReq("GET", turl, &headers, nil, nil))
	if err != nil {
		return err
	}

	csrf := getCookieValue("csrftoken", IgApiUrl, s.Task.client.Jar)
	expBody, err := s.SignRequestBody(map[string]interface{} {
		"_csrftoken": csrf,
		"id": s.Uuid,
		"server_config_retrieval": "1",
		"experiments": IgExperiments,
	})
	headers = s.BaseHeaders()
	headers = append(headers, [2]string{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"})
	resp, err := s.Task.doReq(s.Task.client, s.Task.makeReq("POST", IgExperimentSyncUrl, &headers, nil, &expBody))
	if err != nil {
		return err
	}
	s.PwPubKey = resp.Header.Get("ig-set-password-encryption-pub-key")
	s.PwKeyId, _ = strconv.Atoi(resp.Header.Get("ig-set-password-encryption-key-id"))

	return nil
}

type IgGraphResponse struct {
	Data map[string]interface{} `json:"data"`
	Errors []IgGraphError `json:"errors"`
}

type IgGraphError struct {
	AllowRetry bool `json:"allow_user_retry"`
	Description string `json:"description"`
	Summary string `json:"summary"`
}

func (e *IgGraphError) Error() string { return e.Description }


type IgMailingAddress struct {
	Address1 string `json:"street1"`
	Address2 string `json:"street2"`
	Id string `json:"id"`
}

type IgPaymentMethod struct {
	Id string `json:"id"`
	Last4 string `json:"last4"`
}

func (s *IgSession) ListShipMethods() ([]IgMailingAddress, error) {
	var addys = []IgMailingAddress{}
	shipMethods, err := s.GraphQuery("IgPaymentsSettingsShippingInfoContainerAppQuery", "2316988651664331", map[string]interface{} {})
	if err != nil {
		return addys, err
	}
	if addysIf, ok := shipMethods.Data["me"].(map[string]interface{})["mailing_addresses"].(map[string]interface{})["nodes"]; ok {
		bb, err := json.Marshal(addysIf)
		if err != nil {
			return addys, err
		}
		err = json.Unmarshal(bb, &addys)
		return addys, err
	}
	return addys, err
}

func (s *IgSession) ListPaymentMethods() ([]IgPaymentMethod, error) {
	var methods = []IgPaymentMethod{}
	shipMethods, err := s.GraphQuery("IgPaymentsSettingsPaymentMethodsContainerAppQuery", "2894013883953543", map[string]interface{} {"payment_type": "ig_payment_settings"})
	if err != nil {
		return methods, err
	}
	if credsIf, ok := shipMethods.Data["me"].(map[string]interface{})["pay_consumer_payment_account"].(map[string]interface{})["payment_credentials"]; ok {
		bb, err := json.Marshal(credsIf)
		if err != nil {
			return methods, err
		}
		err = json.Unmarshal(bb, &methods)
		return methods, err

		return methods, nil
	}
	return methods, err
}

func (s *IgSession) SetPaymentAccountName() (string, error) {
	sp := s.Task.Profile.ShippingAddress
	resp, err := s.GraphQuery("IgPaymentsUpdatePaymentAccountNameMutation", "1095111680613571", map[string]interface{} {
		"input": map[string]interface{} {
			"client_mutation_id":"25","payment_type":"IG_PAYMENT_SETTINGS","payer_name":fmt.Sprintf("%s %s", sp.FirstName, sp.LastName),
		},
	})
	if err != nil {
		return "", err
	}
	if id, ok := (resp.Data["payment_account_update_payer_name"].(map[string]interface{}))["payment_account"].(map[string]interface{})["id"].(string); ok {
		return id, nil
	}
	return "", err
}

func (s *IgSession) SetPaymentAccountEmail() (string, error) {
	p := s.Task.Profile
	resp, err := s.GraphQuery("IgPaymentsAddPaymentAccountEmailMutation", "2190632550954177", map[string]interface{} {
		"input": map[string]interface{} {
			"client_mutation_id":"26","logging_id":generateUuid(),"payment_type":"IG_PAYMENT_SETTINGS","user_input_email_address":p.Email,
		},
	})
	if err != nil {
		return "", err
	}
	if id, ok := (resp.Data["add_payment_account_email"].(map[string]interface{}))["payment_account_email"].(map[string]interface{})["id"].(string); ok {
		return id, nil
	}
	return "", err
}

func (s *IgSession) GraphQuery(name string, docId string, data map[string]interface{}) (IgGraphResponse, error) {
	var gresp IgGraphResponse
	IgGraphUrl.RawQuery = string(urlEncodeParams([][2]string{ { "locale", "en_US" }}))
	dbytes, err := json.Marshal(data)
	if err != nil {
		return gresp, err
	}
	body := urlEncodeParamsMap(map[string]string {
		"access_token": "undefined",
		"fb_api_caller_class": "RelayModern",
		"fb_api_req_friendly_name": name,
		"variables": string(dbytes),
		"doc_id": docId,
	})
	headers := s.BaseHeaders()
	headers = append(headers, [2]string{"content-type", "application/x-www-form-urlencoded"})
	resp, err := s.Task.doReq(s.Task.client, s.Task.makeReq("POST", IgGraphUrl, &headers, nil, &body))
	if err != nil {
		return gresp, err
	}
	if resp.Header.Get("x-ig-set-www-claim") != "" {
		s.WwwClaim = resp.Header.Get("x-ig-set-www-claim")
	}
	rbody, err := readBodyBytes(resp)
	if err != nil {
		return gresp, err
	}
	err = json.Unmarshal(rbody, &gresp)
	if len(gresp.Errors) > 0 {
		err = &gresp.Errors[0]
	}
	return gresp, err
}

func (s *IgSession) GraphQueryCustom(docId string, qps [][2]string, data map[string]interface{}) (IgGraphResponse, error) {
	gu, _ := url.Parse(IgGraphUrl.String())
	gu.RawQuery = ""
	s.Task.client.Jar.SetCookies(IgApiUrl, []*http.Cookie{
		&http.Cookie {
			Name: "igfl",
			Value: s.Username,
			Secure: true,
		},
		&http.Cookie {
			Name: "is_starred_enabled",
			Value: "yes",
			Secure: true,
		},
		&http.Cookie {
			Name: "ig_direct_region_hint",
			Value: "PRN",
			Secure: true,
		},
	})

	var gresp IgGraphResponse
	dbytes, err := json.Marshal(data)
	if err != nil {
		return gresp, err
	}
	docids := [][2]string {
		{"doc_id", docId},
	}
	docids = append(docids, qps...)
	docids = append(docids, [2]string{"query_params", string(dbytes)})
	body := urlEncodeParams(docids)
	headerOrder := []string {
	  "x-ig-connection-type",
	  "x-ig-capabilities",
	  "x-ig-app-id",
	  "user-agent",
	  "accept-language",
	  "cookie",
	  "authorization",
	  "x-mid",
	  "ig-u-ig-direct-region-hint",
	  "ig-u-shbid",
	  "ig-u-shbts",
	  "ig-u-ds-user-id",
	  "ig-u-rur",
	  "content-type",
	  "content-length",
	  "accept-encoding",
	  "x-fb-http-engine",
	}
	headers := [][2]string {
	  {"x-ig-connection-type", "WIFI"},
	  {"x-ig-capabilities", "3brTvx8="},
	  {"x-ig-app-id", "567067343352427"},
	  {"accept-language", "en-US"},
	  {"authorization", s.Authorization},
	  {"x-mid", getCookieValue("mid", IgApiUrl, s.Task.client.Jar)},
	  {"ig-u-ig-direct-region-hint", "PRN"},
	  {"ig-u-shbid", getCookieValue("shbid", IgApiUrl, s.Task.client.Jar)},
	  {"ig-u-shbts", getCookieValue("shbts", IgApiUrl, s.Task.client.Jar)},
	  {"ig-u-ds-user-id", getCookieValue("ds_user_id", IgApiUrl, s.Task.client.Jar)},
	  {"ig-u-rur", "ASH"},
	  {"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
	  {"accept-encoding", "gzip, deflate"},
	  {"x-fb-http-engine", "Liger"},
	}
	resp, err := s.Task.doReq(s.Task.client, s.Task.makeReq("POST", gu, &headers, &headerOrder, &body))
	if err != nil {
		return gresp, err
	}
	if resp.Header.Get("x-ig-set-www-claim") != "" {
		s.WwwClaim = resp.Header.Get("x-ig-set-www-claim")
	}
	rbody, err := readBodyBytes(resp)
	if err != nil {
		return gresp, err
	}
	err = json.Unmarshal(rbody, &gresp)
	if len(gresp.Errors) > 0 {
		err = &gresp.Errors[0]
	}
	return gresp, err
}

func (s *IgSession) AddProfileShippingAddress() (string,error) {
	sp := s.Task.Profile.ShippingAddress
	resp, err := s.GraphQuery("IgPaymentsAddressAddMutation", "2538095506309894", map[string]interface{} {
		"input": map[string]interface{} {
			"client_mutation_id": "21",
			"care_of": fmt.Sprintf("%s %s", sp.FirstName, sp.LastName),
			"street1": sp.Address1,
			"street2": sp.Address2,
			"city": sp.City,
			"state": sp.State,
			"postal_code": sp.Zip,
			"country": sp.Country,
			"is_default": true,
			"payment_type": "IG_PAYMENT_SETTINGS",
		},
	})
	if err != nil {
		return "", err
	}
	if id, ok := (resp.Data["add_mailing_address"].(map[string]interface{}))["mailing_address"].(map[string]interface{})["id"].(string); ok {
		return id, nil
	}
	return "", err
}

func (s *IgSession) GetPaymentAccountId() (string, error) {
	resp, err := s.GraphQuery("IgPaymentsSettingsPaymentMethodsContainerAppQuery", "2894013883953543", map[string]interface{} {
		"payment_type": "ig_payment_settings",
	})
	if id, ok := (resp.Data["me"].(map[string]interface{}))["pay_consumer_payment_account"].(map[string]interface{})["id"].(string); ok {
		return id, nil
	}
	return "", err
}

func (s *IgSession) GetPaymentToken() (string, error) {
	body := urlEncodeParamsMap(map[string]string {
		"creditCardNumber": s.Task.Profile.Card.Number,
		"csc": s.Task.Profile.Card.Cvc,
	})
	headers := [][2]string {
		{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
	}
	client, _ := s.Task.newHttpClient()
	resp, err := s.Task.doReq(client, s.Task.makeReq("POST", FbPayTokenUrl, &headers, nil, &body))
	if err != nil {
		return "", err
	}
	rbody, err := readBodyBytes(resp)
	if err != nil {
		return "", err
	}
	var bmap map[string]interface{}
	err = json.Unmarshal(rbody, &bmap)
	if err != nil {
		return "", err
	}
	if token, ok := bmap["token"].(string); ok {
		return token, nil
	} else {
		return "", err
	}
}

func (s *IgSession) AddPaymentMethod() (string, error) {
	bs := s.Task.Profile.BillingAddress
	card := s.Task.Profile.Card
	token, err := s.GetPaymentToken()
	if err != nil {
		return "", err
	}
	// payment_account_id, err := s.GetPaymentAccountId()
	// if err != nil {
	// 	return "", err
	// }
	// seshid := generateUuid()

	resp, err := s.GraphQueryCustom("3866971486710435", [][2]string{
		{"locale", "en_US"},
		{"vc_policy", "default"},
		{"signed_body", "SIGNATURE."},
		{"strip_nulls", "true"},
		{"strip_defaults", "true"},
	}, map[string]interface{} {
		"input": map[string]interface{} {
			"client_mutation_id":generateUuid(),
			"expiry_month": fmt.Sprintf("%02d", card.ExpMonth),
			"expiry_year": fmt.Sprintf("%d", card.ExpYear),
			"cardholder_name": card.Name,
			"billing_address": map[string]string {
			  "street1": bs.Address1,
			  "street2": bs.Address2,
			  "city": bs.City,
			  "state":bs.State,
			  "zip":bs.Zip,
			  "country_code":bs.Country,
			},
			"payment_type":"fbpay_hub",
			"credit_card_token":token,
		},
	})
	if err != nil {
		return "", err
	}
	if id, ok := (resp.Data["add_credit_card"].(map[string]interface{}))["credit_card"].(map[string]interface{})["id"].(string); ok {
		return id, nil
	}
	return "", err
}

type IgAmount struct {
	Amount string `json:"amount"`
	Currency string `json:"currency"`
	Offset int `json:"offset"`
	AmountWithOffset string `json:"amount_with_offset"`
}

type IgProductVariant struct {
	Name string `json:"name"`
	Id int64 `json:"product_id"`
	Merchant IgMerchant `json:"merchant"`
	Values []IgVariantValues `json:"variant_values"`
	CheckoutProperties IgCheckoutProperties `json:"checkout_properties"`
}

type IgPrice struct {
	Amount string `json:"amount"`
	Currency string `json:"currency"`
}

type IgTotalPrice struct {
	Price IgPrice `json:"price"`
	DisplayPrice string `json:"display_price"`
	Label string `json:"label"`
	PriceType string `json:"price_type"`
}

type IgCheckoutScreen struct {
	Name string `json:"__typename"`
	TotalPrice IgTotalPrice `json:"total_price"`
	ShippingOptionId string `json:"selected_delivery_option_id"`
}

type IgCheckoutInfo struct {
	Screens []IgCheckoutScreen `json:"checkout_screen_components"`
	PaymentSessionId string `json:"payment_session_id"`
}
type IgProductResponse struct {
  ProductGroup IgProductGroup `json:"product_group"`
  Status       string       `json:"status"`
}
type IgValues struct {
  Label string `json:"label"`
}
type IgVariantDimensions struct {
  ID          string   `json:"id"`
  Name        string   `json:"name"`
  Values      []IgValues `json:"values"`
  VisualStyle string   `json:"visual_style"`
}
type IgMerchant struct {
  Id            int64  `json:"pk"`
  Username      string `json:"username"`
  ProfilePicURL string `json:"profile_pic_url"`
}
type IgTextWithEntities struct {
  Text              string        `json:"text"`
  InlineStyleRanges []interface{} `json:"inline_style_ranges"`
}
type IgRichTextDescription struct {
  Blocktype        string           `json:"block_type"`
  Depth            int              `json:"depth"`
  TextWithEntities IgTextWithEntities `json:"text_with_entities"`
}
type IgCandidates struct {
  URL    string `json:"url"`
  Width  int    `json:"width"`
  Height int    `json:"height"`
}
type IgImageVersions2 struct {
  Candidates []IgCandidates `json:"candidates"`
}
type IgMainImage struct {
  ImageVersions2 IgImageVersions2 `json:"image_versions2"`
  Preview        string         `json:"preview"`
}
type IgThumbnailImage struct {
  ImageVersions2 IgImageVersions2 `json:"image_versions2"`
  Preview        string         `json:"preview"`
}
type IgCurrencyAmount struct {
  Currency         string `json:"currency"`
  Amount           string `json:"amount"`
  AmountWithOffset string `json:"amount_with_offset"`
  Offset           int    `json:"offset"`
}
type IgShippingCost struct {
  Currency         string `json:"currency"`
  Amount           string `json:"amount"`
  AmountWithOffset string `json:"amount_with_offset"`
  Offset           int    `json:"offset"`
}
type IgShippingAndReturn struct {
  ShippingCost IgShippingCost `json:"shipping_cost"`
}
type IgCheckoutProperties struct {
  Amount           IgAmount    `json:"currency_amount"`
  ReceiverID               json.Number             `json:"receiver_id"`
  InventoryQuantity        int               `json:"inventory_quantity"`
  HasFreeShipping          bool              `json:"has_free_shipping"`
  CanAddToBag              bool              `json:"can_add_to_bag"`
  CanEnableRestockReminder bool              `json:"can_enable_restock_reminder"`
  ViewerPurchaseLimit      int               `json:"viewer_purchase_limit"`
  ProductGroupHasInventory bool              `json:"product_group_has_inventory"`
  ShippingAndReturn        IgShippingAndReturn `json:"shipping_and_return"`
  IgReferrerFbid           int64             `json:"ig_referrer_fbid"`
}
type IgVariantValues struct {
  ID              string      `json:"id"`
  Value           string      `json:"value"`
  NormalizedValue interface{} `json:"normalized_value"`
  Name            string      `json:"name"`
  IsPreselected   bool        `json:"is_preselected"`
  VisualStyle     string      `json:"visual_style"`
}
type IgIncentiveInformation struct {
  Incentives []interface{} `json:"incentives"`
}
type IgProductImages struct {
  ImageVersions2 IgImageVersions2 `json:"image_versions2"`
  Preview        string         `json:"preview"`
}
type IgProductItems struct {
  Name                    string                `json:"name"`
  Price                   string                `json:"price"`
  CurrentPrice            string                `json:"current_price"`
  FullPrice               string                `json:"full_price"`
  Id               int64                 `json:"product_id"`
  Merchant                IgMerchant              `json:"merchant"`
  CompoundProductID       string                `json:"compound_product_id"`
  Description             string                `json:"description"`
  RichTextDescription     []IgRichTextDescription `json:"rich_text_description"`
  RetailerID              string                `json:"retailer_id"`
  HasViewerSaved          bool                  `json:"has_viewer_saved"`
  MainImage               IgMainImage             `json:"main_image"`
  ThumbnailImage          IgThumbnailImage        `json:"thumbnail_image"`
  ReviewStatus            string                `json:"review_status"`
  ExternalURL             string                `json:"external_url"`
  CheckoutStyle           string                `json:"checkout_style"`
  CanShareToStory         bool                  `json:"can_share_to_story"`
  CanSeeInsightsForViewer bool                  `json:"can_see_insights_for_viewer"`
  FullPriceStripped       string                `json:"full_price_stripped"`
  CurrentPriceStripped    string                `json:"current_price_stripped"`
  CheckoutProperties      IgCheckoutProperties    `json:"checkout_properties"`
  VariantValues           []IgVariantValues       `json:"variant_values"`
  IncentiveInformation    IgIncentiveInformation  `json:"incentive_information"`
  ProductImages           []IgProductImages       `json:"product_images"`
}
type IgProductGroup struct {
  VariantDimensions []IgVariantDimensions `json:"variant_dimensions"`
  ProductItems      []IgProductItems      `json:"product_items"`
}

func (s *IgSession) GetProductVariants(merchantId string, productId string) ([]IgProductItems, error) {
	variants := []IgProductItems{}
	params := urlEncodeParams([][2]string {
		{"device_width", "1080",},
		{"shopping_bag_enabled", "true"},
		{"merchant_id", merchantId},
	})
	productVariantUrl, err := url.Parse(fmt.Sprintf("https://i.instagram.com/api/v1/commerce/products/%s/group/?%s", productId, string(params)))
	if err != nil {
		return variants, err
	}

	headers := s.BaseHeaders()
	resp, err := s.Task.doReq(s.Task.client, s.Task.makeReq("GET", productVariantUrl, &headers, nil, nil))
	if err != nil {
		return variants, err
	}
	var igresp IgProductResponse
	err = readRespJsonDst(resp, &igresp)
	if err != nil {
		return variants, err
	}
	return igresp.ProductGroup.ProductItems, nil
}

type IgBagItem struct {
	Quantity int `json:"quantity"`
	Product IgProductVariant `json:"product"`
}
type IgBag struct {
	Items []IgBagItem `json:"items"`
}
type IgCart struct {
	Bags []IgBag `json:"bags"`
	MerchantBagInfos []IgMerchantBagInfo `json:"merchant_bag_infos"`
}
type IgMerchantBagInfo struct {
	Merchant IgMerchant `json:"merchant"`
	MerchantBagId uint64 `json:"merchant_bag_id"`
	GlobalBagId uint64 `json:"multi_merchant_bag_id"`
}
func (s *IgSession) Checkout(paymentMethodId string, shipAddressId string, emailId string) error {
	merchantId := s.Task.Url.Host
	productId := s.Task.Url.Path[1:]
	s.Task.StatusCh <- Status{Message: "Visiting Product"}
	variants, err := s.GetProductVariants(merchantId, productId)
	// log.Printf("%+v %+v", variants, err)
	if err != nil {
		return retry.Unrecoverable(err)
	}
	if len(variants) == 0 {
		return retry.Unrecoverable(errors.New("No Product Variants Available"))
	} else if string(variants[0].CheckoutProperties.ReceiverID) == "" {
		s.Task.StatusCh <- Status{Message: "Account Does Not Support IG Checkout"}
		return retry.Unrecoverable(errors.New("Account Cannot Access IG Checkout (Try A More Active One)"))
	}

	rand.Shuffle(len(variants), func(i, j int) {
			variants[i], variants[j] = variants[j], variants[i] })
	var product *IgProductItems
	for _, pv := range variants {
		// if pv.CheckoutProperties.InventoryQuantity > 0 && (len(s.Task.Sizes) == 0 || Any(s.Task.Sizes, func(size string) bool { return pv.VariantValues[0].Value == size })) {
			if (len(s.Task.Sizes) == 0 || Any(s.Task.Sizes, func(size string) bool { return pv.VariantValues[0].Value == size })) {
			product = &pv
			break
		}
	}

	if product == nil {
		s.Task.StatusCh <- Status{Message: "Out of Stock"}
		return errors.New("Out of Stock")
	}

	atcBody, err := s.SignRequestBody(map[string]interface{} {
		"_csrftoken": getCookieValue("csrftoken", IgApiUrl, s.Task.client.Jar),
		"_uid": getCookieValue("ds_user_id", IgApiUrl, s.Task.client.Jar),
		"_uuid": s.Uuid,
		"items": []map[string]interface{} {
			map[string]interface{} {
				"merchant_id": merchantId,
				"product_id": productId,
				"additional_quantity": 1,
			},
		},
	})
	headers := s.BaseHeaders()
	headers = append(headers, [2]string{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"})
	resp, err := s.Task.doReq(s.Task.client, s.Task.makeReq("POST", IgAtcUrl, &headers, nil, &atcBody))
	if err != nil {
		return err
	}
	if resp.Header.Get("x-ig-set-www-claim") != "" {
		s.WwwClaim = resp.Header.Get("x-ig-set-www-claim")
	}
	// log.Printf("%+v", resp.Header)

	// for _, sc := range resp.Header["Set-Cookie"] {
	// 	log.Printf("%+v", sc)
	// 	if strings.Contains(sc, "urlgen") {
	// 		log.Println(sc[7:strings.Index(sc, ";")])
	// 		s.Task.client.Jar.SetCookies(IgApiUrl, []*http.Cookie {
	// 			&http.Cookie {
	// 				Name: "urlgen",
	// 				Value: sc[8:strings.Index(sc, "\";")],
	// 				Secure: true,
	// 			},
	// 		})
	// 	}
	// }
	respb, err := readBodyBytes(resp)
	if err != nil {
		return err
	}
	var cart IgCart
	err = json.Unmarshal(respb, &cart)
	if err != nil {
		return err
	}

	if len(cart.MerchantBagInfos) == 0 {
		time.Sleep(1 * time.Second)
		return errors.New("Failed to ATC")
	}
	var merchantBagInfo IgMerchantBagInfo
	for _, mbi := range cart.MerchantBagInfos {
		if fmt.Sprintf("%d", mbi.Merchant.Id) == merchantId {
			merchantBagInfo = mbi
			break
		}
	}

	perUnitPrice := map[string]string {
		"amount": product.CheckoutProperties.Amount.Amount,
		"currency": product.CheckoutProperties.Amount.Currency,
	}
	return s.Task.Retry(func() error {
		checkoutUuid := generateUuid()
		ip := fmt.Sprintf("10.0.1.%d", 2 + rand.Intn(251))
		deviceIdHash := strings.Split(s.DeviceId, "android-")[1]
		riskFeatures, err := json.Marshal(map[string]interface{} {
		  "MobileDeviceJailBroken": false,
		  "MobileSimOperatorName": "",
		  "MobileAppVersion": "159.0.0.40.122",
		  "MobileDeviceVPN": false,
		  "MobileDeviceLocalTimezone": "Pacific Daylight Time",
		  "MobileDeviceUptime": int((3600 * 1000) + int(86400*3*1000*(rand.Float64()))),
		  "MobileDeviceModel": "OnePlus",
		  "MobileIpAddresses": []interface{} {
		        "fe80::4c8b:61ff:fe25:b4ff%dummy0",
				    "fe80::4cca:842a:a955:920a%r_rmnet_data0",
				    "fe80::9665:2dff:fe40:cb8%wlan0",
				    "10.0.1.26",
				    "fe80::37ec:434a:7363:b702%rmnet_data0",
				    "fe80::4a5b:8190:8c7d:2802%rmnet_data1",
				    "2607:fc20:2799:91f3:4a5b:8190:8c7d:2802",
		  },
		  "MobileTimezoneOffsetMs": -25200000,
		  "MobileDeviceName": "OnePlus5T",
		  "MobileDeviceProxySet": false,
		  "SourceIP": ip,
		  "MobileSimSerialNumber": nil,
		  "MobileSerialNumber": "PERMISSION_DENIED",
		  "MobileConnType": "wifi",
		  "MobileCompVersion": "1.0.0",
		  "MobilePayloadType": "full",
		  "MobileConfVersion": "1.0.0",
		  "MobileAppGuid": s.SessionId,
		  "MobileAppLastUpdateTime": int(1587740752010 - int(rand.Float64()*(86400 * 1000))),
		  "MobileGsfId": nil,
		  "MobileAppName": "com.instagram.android",
		  "MobileDeviceIsEmulator": false,
		  "MobileAndroidID": deviceIdHash,
		  "MobileOSType": "Android",
		  "MobileOSVersion": "9",
		  "MobileDayLightSavingsTimeEnabled": true,
		  "MobileCdmaNetworkId": nil,
		  "MobileAppFirstInstallTime": int(1587740752010 - int(rand.Float64()*((86400*90) * 1000))),
		  "MobileDeviceLocaleCountry": "US",
		  "MobileCdmaSystemId": nil,
		  "MobileDeviceLocaleLanguage": "en",
		  "MobileRequestTimestamp": time.Now().Unix(),
		  "MobileSubscriberId": nil,
		  "MobileRoaming": false,
		  "MobileDeviceCanSendText": true,
		  "MobileDeviceId": deviceIdHash,
		  "MobileLine1Number": nil,
		  "MobileDeviceTotalDiskSpace": 56511012864,
		  "MobileDeviceRadioType": 1,
		})
		if err != nil {
			return err
		}
		extraData, err := json.Marshal(map[string]interface{} {
		  "is_from_drops_onboarding":false,
		  "products": []map[string]interface{} {
		    map[string]interface{} {
		      "product_id":fmt.Sprintf("%d", product.Id),"quantity":1,
		      "price": product.CheckoutProperties.Amount,
		    },
		  },
		})
		if err != nil {
			return err
		}

		var checkoutInfo IgCheckoutInfo
		s.Task.StatusCh <- Status{Message: "Waiting for Release"}
		// log.Printf("receiverid %d", int64(product.CheckoutProperties.ReceiverId))
		log.Printf("prodchk %+v", product.CheckoutProperties)
		gresp, err := s.GraphQueryCustom("3542956782415694", [][2]string{
			{"locale", "en_US"},
			{"vc_policy", "default"},
			{"signed_body", "SIGNATURE."},
			{"strip_nulls", "true"},
			{"strip_defaults", "true"},
		}, map[string]interface{} {
			"input": map[string]interface{} {
				"log_fbpay_experience_exposure": true,
				"receiver_id": string(product.CheckoutProperties.ReceiverID),
				"logging_id": checkoutUuid,
				"payment_type":"IG_NMOR_SHOPPING",
				"products": []map[string]interface{} {
					{"product_id":fmt.Sprintf("%d", product.Id),"quantity":1,"per_unit_price":perUnitPrice},
				},
				"risk_features": string(riskFeatures),
				"extra_data": string(extraData),
				"client_mutation_id":generateUuid(),
				"actor_id": getCookieValue("ds_user_id", IgApiUrl, s.Task.client.Jar),
			},
			"paymentType":"IG_NMOR_SHOPPING",
		})
		if err != nil {
			return err
		}
		if checkoutInfoIf, ok := (gresp.Data["checkout_information_mutation"].(map[string]interface{}))["checkout_information"]; ok {
			bb, err := json.Marshal(checkoutInfoIf)
			if err != nil {
				return err
			}
			err = json.Unmarshal(bb, &checkoutInfo)
			if err != nil {
				return err
			}
		}

		var totalPrice IgPrice
		var shippingOptionId string
		for _, screen := range checkoutInfo.Screens {
			if screen.Name == "PaymentCheckoutScreenPriceTable" {
				totalPrice = screen.TotalPrice.Price
			} else if screen.Name == "PaymentCheckoutScreenDeliveryOptions" {
				shippingOptionId = screen.ShippingOptionId
			}
		}

		s.Task.StatusCh <- Status{Message: "Attempting Checkout"}
		gresp, err = s.GraphQuery("IgPaymentsCheckoutInformationMutation", "3121451107937139", map[string]interface{} {
			"input": map[string]interface{} {
				"client_mutation_id":"1",
				"payment_session_id": checkoutInfo.PaymentSessionId,
				"payment_type":"IG_NMOR_SHOPPING",
				"logging_id": "",
				"log_fbpay_experience_exposure": true,
				"receiver_id": string(product.CheckoutProperties.ReceiverID),
				"products": []map[string]interface{} {
					{"product_id":fmt.Sprintf("%d", product.Id),"quantity":1,"per_unit_price":perUnitPrice},
				},
				"extra_data": string(extraData),
				"selected_shipping_address_id": shipAddressId,
				"selected_delivery_option_id": shippingOptionId,
				"selected_contact_email_id": emailId,
				"contact_name": fmt.Sprintf("%s %s", s.Task.Profile.ShippingAddress.FirstName, s.Task.Profile.ShippingAddress.LastName),
				"selected_delivery_option_ids": []string{},
				"selected_payment_credentials": []map[string]interface{} {
					map[string]interface{} {
					  "credential_type":"CREDIT_CARD",
					  "credential_id": paymentMethodId,
					  "balance_amount": totalPrice,
					},
				},
				"risk_features": string(riskFeatures),
			},
			"paymentType":"IG_NMOR_SHOPPING",
		})
		if err != nil {
			return err
		}
		if checkoutInfoIf, ok := (gresp.Data["checkout_information_mutation"].(map[string]interface{}))["checkout_information"]; ok {
			bb, err := json.Marshal(checkoutInfoIf)
			if err != nil {
				return err
			}
			err = json.Unmarshal(bb, &checkoutInfo)
			if err != nil {
				return err
			}
		}


		for _, screen := range checkoutInfo.Screens {
			if screen.Name == "PaymentCheckoutScreenPriceTable" {
				totalPrice = screen.TotalPrice.Price
			} else if screen.Name == "PaymentCheckoutScreenDeliveryOptions" {
				shippingOptionId = screen.ShippingOptionId
			}
		}

		extraDataJson, err := json.Marshal(map[string]interface{} {
		  "products": []map[string]interface{} {
		  	map[string]interface{} {
		      "product_id": fmt.Sprintf("%d", product.Id),
		      "quantity": 1,
		      "price": product.CheckoutProperties.Amount,
		     },
		  },
		  "attribution_data": map[string]interface{} {
		    "channel": "ig_shopping",
		    "ig_extra_data": map[string]interface{} {
		      "merchant_igid": merchantId,
		      "prior_module": "profile",
		      "entry_point": "shop_profile",
		      "global_bag_id": fmt.Sprintf("%d", merchantBagInfo.GlobalBagId),
		      "merchant_bag_id": fmt.Sprintf("%d", merchantBagInfo.MerchantBagId),
		    },
		    "session_id": checkoutUuid,
		  },
		  "is_from_shopping_bag": false,
		  "is_from_drops_onboarding": false,
		  "ig_referrer_fbid": "17841401078560143",
		  "buyer_applied_incentives": []string{},
		})
		if err != nil {
			return err
		}

		// time.Sleep(5)
		s.Task.StatusCh <- Status{Message: "Submitting Payment"}
		gresp, err = s.GraphQuery("IgPaymentsCheckoutMakeAsyncPaymentMutation", "2189220951204137", map[string]interface{} {
		    "input": map[string]interface{} {
		      "client_mutation_id":"2",
		      "contact_name": s.Task.Profile.ShippingAddress.FirstName + " " + s.Task.Profile.ShippingAddress.LastName,
		      "coupon_code": nil,
		      "credentials": []map[string]interface{} {
		        map[string]interface{} {
		          "credential_type":"CREDIT_CARD",
		          "payment_amount": totalPrice,
		          "credential_id": paymentMethodId,
		        },
		      },
		      "email_address_id": emailId,
		      "email_opt_in": false ,
		      "extra_data_json": string(extraDataJson),
		      "logging_id": checkoutUuid,
		      "mailing_address_id": shipAddressId,
		      "payment_session_id": checkoutInfo.PaymentSessionId,
		      "payment_type":"IG_NMOR_SHOPPING",
		      "products":[]map[string]interface{} {
		        map[string]interface{} {
		          "product_id": fmt.Sprintf("%d", product.Id),
		          "quantity":1,
		          "per_unit_price": perUnitPrice,
		        },
		      },
		      "risk_features": string(riskFeatures),
		      "security_pin": nil,
		      "shipping_option_id": shippingOptionId,
		      "selected_delivery_option_ids": []string{},
		      "total_payment_amount": totalPrice,
		    },
		})

		// time.Sleep(2 * time.Second)

		var paymentStatus IgPaymentStatus
		paymentStatus.Status = "IN_PROGRESS"
		for paymentStatus.Status == "IN_PROGRESS" {
			statusResp, err := s.GraphQuery("IgPaymentsPaymentSessionStatusQuery", "2541303555921357", map[string]interface{} {"payment_session_id": checkoutInfo.PaymentSessionId, "logging_id": checkoutUuid})
			if err != nil {
				return retry.Unrecoverable(err)
			}
			if seshStatusIf, ok := statusResp.Data["consumer_payment_session_status"]; ok {
				bb, err := json.Marshal(seshStatusIf)
				if err != nil {
					return retry.Unrecoverable(err)
				}
				err = json.Unmarshal(bb, &paymentStatus)
				if err != nil {
					return retry.Unrecoverable(err)
				}
			} else {
				return retry.Unrecoverable(errors.New("Unknown checkout error during status update"))
			}
			time.Sleep(1 * time.Second)
		}
		// log.Printf("%+v", paymentStatus)

		if paymentStatus.Status == "SUCCESS" {
			s.Task.StatusCh <- Status{Message: "COOKED!!!"}
			go s.Task.SuccessWebhook("ig://" + product.Merchant.Username, product.Name, product.VariantValues[0].Value)
				return nil
		} else if paymentStatus.Status == "FAILURE" {
			s.Task.StatusCh <- Status{Message: "Payment Declined"}
			if paymentStatus.Error.Title == "Payment Declined" {
				go s.Task.DeclineWebhook("ig://" + product.Merchant.Username, product.Name, product.VariantValues[0].Value)
					return nil
			} else {
				s.Task.StatusCh <- Status{Message: "Payment Error: " + paymentStatus.Error.Title}
				return retry.Unrecoverable(errors.New(paymentStatus.Error.Title))
			}
		}
		return ErrRetrying
	}, retry.OnRetry(func(_ uint, err error) {
		if strings.Contains(fmt.Sprintf("%+v", err), "product launch") {
			// s.Task.StatusCh <- Status{Message: "Waiting for Release"}
		} else {
			s.Task.StatusCh <- Status{Message: "Checkout error. Trying again..."}
		}
	}), retry.Delay(1*time.Second))
}

type IgPaymentError struct {
	Title string `json:"error_title"`
}

type IgPaymentStatus struct {
	Status string `json:"payment_status"`
	Error IgPaymentError `json:"payments_error"`
}

func (s *IgSession) BaseHeaders() [][2]string {
	headers := [][2]string {
		{"x-ig-app-locale", "en_US"},
		{"x-ig-device-locale", "en_US"},
		{"x-ig-mapped-locale", "en_US"},
		{"x-pigeon-session-id", s.SessionId},
		{"x-pigeon-rawclienttime", fmt.Sprintf("%d.000", time.Now().Unix())},
		{"x-ig-connection-speed", "-1kbps"},
		{"x-ig-bandwidth-speed-kbps", "-1.000"},
		{"x-ig-bandwidth-totalbytes-b", "0"},
		{"x-ig-bandwidth-totaltime-ms", "0"},
		{"x-ig-app-startup-country", "US"},
		{"x-bloks-version-id", "af8a708ea80a7779936b0322bba13d118a7896b44d416dfc742db66a1532be0c"},
		{"x-bloks-is-layout-rtl", "false"},
		{"x-ig-device-id", s.Uuid},
		{"x-ig-android-id", s.DeviceId},
		{"x-ig-connection-type", "WIFI"},
		{"x-ig-capabilities", "3brTvx8="},
		{"x-ig-app-id", "567067343352427"},
		// {"user-agent", "Instagram 136.0.0.34.124 Android (28/9; 420dpi; 1080x2135; OnePlus; ONEPLUS A6010; OnePlus6T; qcom; en_US; 208061712)"},
		{"accept-language", "en-US"},
		{"accept-encoding", "gzip, deflate"},
		{"x-fb-http-engine", "Liger"},
	}


	if s.WwwClaim != "" {
		headers = append(headers, [2]string{"x-ig-www-claim", s.WwwClaim})
	} else {
		headers = append(headers, [2]string{"x-ig-www-claim", "0"})
	}
	if s.Authorization != "" {
		headers = append(headers, [2]string{"authorization", s.Authorization})
	}

	cval := getCookieValue("mid", IgApiUrl, s.Task.client.Jar)
	if cval != "" {
		headers = append(headers, [2]string{"x-mid", cval})
	}
	cval = getCookieValue("rur", IgApiUrl, s.Task.client.Jar)
	if cval != "" {
	  headers = append(headers, [2]string{"ig-u-rur", cval})
	}
	cval = getCookieValue("ds_user_id", IgApiUrl, s.Task.client.Jar)
	if cval != "" {
	  headers = append(headers, [2]string{"ig-u-ds-user-id", cval})
	}
	cval = getCookieValue("shbid", IgApiUrl, s.Task.client.Jar)
	if cval != "" {
	  headers = append(headers, [2]string{"ig-u-shbid", cval})
	}
	cval = getCookieValue("shbts", IgApiUrl, s.Task.client.Jar)
	if cval != "" {
	  headers = append(headers, [2]string{"ig-u-shbts", cval})
	}
	cval = getCookieValue("ig_direct_region_hint", IgApiUrl, s.Task.client.Jar)
	if cval != "" {
	  headers = append(headers, [2]string{"ig-u-ig-direct-region-hint", cval})
	} else {
		headers = append(headers, [2]string{"ig-u-ig-direct-region-hint", "PRN"})
	}

	return headers
}

func (t *CheckoutTask) IgCheckout() error {
	t.HeaderBlacklist = map[string]bool {
		"sec-fetch-site": true,
		"sec-fetch-mode": true,
		"sec-fetch-dest": true,
		"accept": true,
	}
	t.UserAgent = "Instagram 159.0.0.40.122 Android (28/9; 420dpi; 1080x2034; OnePlus; ONEPLUS A5010; OnePlus5T; qcom; en_US; 245196047)"
	password, _ := t.Url.User.Password()
	if t.Url.User == nil || t.Url.User.Username() == "" || password == "" {
		return errors.New("Missing username or password")
	}
	sesh := IgSession{Task: t, Username: t.Url.User.Username(), Password: password}
	sesh.Init()
	t.StatusCh <- Status{Message:"Launching"}
	err := sesh.Launch()
	if err != nil {
		return err
	}
	t.StatusCh <- Status{Message:"Logging In"}
	err = sesh.Login()
	if err != nil {
		return err
	}

	// TODO test on fresh account, do we SetPaymentAccountName here for fresh payment acc ??
	t.StatusCh <- Status{Message:"Adding Shipping Address"}
	shipMethods, err := sesh.ListShipMethods()
	if err != nil {
		return err
	}

	// // TODO check if ship method exists
	shipAddressId := ""
	for _, method := range shipMethods {
		if strings.ToLower(method.Address1) == strings.ToLower(t.Profile.ShippingAddress.Address1) && strings.ToLower(method.Address2) == strings.ToLower(t.Profile.ShippingAddress.Address2) {
			shipAddressId = method.Id
			break
		}
	}
	if shipAddressId == "" {
		shipAddressId, err = sesh.AddProfileShippingAddress()
		if err != nil {
			return err
		}
	}
	// log.Println(shipAddressId)
	// }

	t.StatusCh <- Status{Message:"Adding Payment Method"}
	_, err = sesh.SetPaymentAccountName()
	if err != nil {
		return err
	}

	payMethods, err := sesh.ListPaymentMethods()
	if err != nil {
		return err
	}
	// log.Printf("%+v", payMethods)

	payMethodId := ""
	for _, method := range payMethods {
		if method.Last4 == t.Profile.Card.Number[len(t.Profile.Card.Number) - 4:] {
			payMethodId = method.Id
			break
		}
	}
	if payMethodId == "" {
		payMethodId, err = sesh.AddPaymentMethod()
		if err != nil {
			return err
		}
	}
	// log.Println(payMethodId)

	// paymentAccId, err := sesh.SetPaymentAccountName()
	// if err != nil {
	// 	return err
	// }
	// log.Printf("payaccid %s", paymentAccId)
	emailId, err := sesh.SetPaymentAccountEmail()
	if err != nil {
		return err
	}
	// log.Printf("emailid %s", emailId)

	return t.Retry(func() error {
		return sesh.Checkout(payMethodId, shipAddressId, emailId)
	}, retry.Delay(1 * time.Second), retry.DelayType(retry.FixedDelay))
}

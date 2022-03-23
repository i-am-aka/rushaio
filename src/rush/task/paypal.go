package task

import (
  // "context"
  crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	// "log"
  "os"
	"net/url"
	"regexp"
	"rush/net/http"
	"strings"
)

var StatusPaypalCardRejected = Status{Message: "Card Rejected by PayPal"}
var ErrPaypalCheckoutFailed = errors.New("PayPal Checkout Failed")

type PaypalCheckoutSession struct {
	Task *CheckoutTask
	Client *http.Client

	// checkout state
	Headers [][2]string
	ButtonUrl *url.URL

	// site specific
  MerchantId string
	TokenizationKey string
	ExperienceProfile map[string]interface{}
	MetaFields map[string]string

	// checkout specific
  payResource *PaypalPaymentResource
  XCookies map[string]string
	RefUrl *url.URL
	Amount float64
  sessionId string
  buttonSessionID string
  uid string
  ectoken string
  csrfheader string
  // ectoken *PaypalApiResponse
}

var proxUrl, _  =  url.Parse("http://127.0.0.1:8889")
var PaypalBaseUrl, _ = url.Parse("https://www.paypal.com/")

func NewPaypalCheckoutSession(tokenizationKey string, task *CheckoutTask, refUrl *url.URL) *PaypalCheckoutSession {
  var client *http.Client
  if os.Getenv("PP_CHLS") == "1" {
	  client, _ = task.newHttpClientProxy(proxUrl)
  } else {
    client, _ = task.newHttpClient()
  }

  // ctx, cancel := context.WithCancel(task.Ctx)
  // ctx := context.Background()
	return &PaypalCheckoutSession{
		Task: task,
		TokenizationKey: tokenizationKey,
		ExperienceProfile: map[string]interface{} {
	    "brandName": "Greenspaces.id", // todo per brand
	    "noShipping": "false",
	    "addressOverride": false,
	  },
	  MetaFields: map[string]string {
  	  "platform": "web",
  	  "sdkVersion": "3.29.0",
  	  "source": "client",
  	  "integration": "custom",
  	  "integrationType": "custom",
	  	"merchantAppId": "https://" + refUrl.Host,
	  	"sessionId": uuid.New().String(),
	  },
	  RefUrl: refUrl,
	  Client: client,
    // Context: ctx,
    // Cancel: cancel,
    MerchantId: "Y9MNL28R45M2Y",
	}
}

func randHex(nbytes int) string {
  var randbytes = make([]byte, nbytes)
  crand.Read(randbytes)
  return hex.EncodeToString(randbytes)
}

func getPaypalSessionId() (string, error) {
  randhex := randHex(5)
  datePart, err := JsEvalFile(map[string]string{}, "paypal/encode.js")
  if err != nil {
    return "", err
  }
  return randhex + "_" + datePart, nil
}

var ErrPaypalCardRejected = errors.New("Card Rejected")

func (s *PaypalCheckoutSession) Checkout() (*PaypalCheckoutSuccess, error) {
	var err error
	defer func() {
		s.Task.LogDebug("%+v", err)
	}()

	// config, err := s.PaypalGetConfig()
	// if err != nil {
	// 	return nil, err
	// }
 //  s.Task.LogDebug("config %+v", config)

  if sessionId, err := getPaypalSessionId(); err == nil {
    s.sessionId = sessionId
  } else {
    return nil, err
  }
  if buttonSessionID, err := getPaypalSessionId(); err == nil {
    s.buttonSessionID = buttonSessionID
  } else {
    return nil, err
  }
  var randbytes = make([]byte, 5)
  crand.Read(randbytes)
  s.uid = hex.EncodeToString(randbytes)

  if url_, err := url.Parse(
    fmt.Sprintf("https://www.paypal.com/webapps/hermes/button?version=4.0.173&env=production&style.size=responsive&style.color=silver&style.shape=rect&style.tagline=false&sessionID=%s&buttonSessionID=%s&funding.disallowed=venmo&locale.x=en_US&logLevel=warn&uid=%s&xcomponent=1",
      s.sessionId, s.buttonSessionID, s.uid)); err == nil {
    s.ButtonUrl = url_
  } else {
    return nil, err
  }


	// err = s.PaypalSetInitialHeaders()
	// if err != nil {
	// 	return nil, err
	// }
  // err = s.PaypalInitSeshGraphql()
  // if err != nil {
  //   return nil, err
  // }
  // err = s.PaypalInitSeshNativeElig()
  // if err != nil {
  //   return nil, err
  // }
	// if payResource, err := s.PaypalCreatePaymentResource(); err == nil {
 //    s.payResource = payResource
 //  } else {
	// 	return nil, err
	// }
 //  if s.payResource == nil {
 //    return nil, ErrPaypalCardRejected
 //  }
 //  s.ectoken = strings.Split(s.payResource.PaymentResource.RedirectURL, "token=")[1]
 s.ectoken = s.TokenizationKey
	s.Task.LogDebug("%+v %+v", s.payResource, err)
	// if ecToken, err := s.GetPaypalEctoken(payResource.PaymentResource.PaymentToken); err == nil {
 //    s.ectoken = ecToken
 //  } else {
 //    return err
 //  }
  // s.Task.LogDebug("%+v %+v", s.ectoken, err)

  err = s.PaypalGetCheckoutNow()
  if err != nil {
    return nil, err
  }


  ppresp, err := s.PaypalSendProfile()
  if err != nil {
    return nil, err
  }
  s.Task.LogDebug("%+v %+v", ppresp, err)

  // TODO can we reuse csrf from button
  csrf, err := s.PaypalVisitXoon()
  if err != nil {
    return nil, err
  }
  // s.Task.StatusCh <- Status{Message: "Sending Profile to PayPal"}

  ppresp2, err := s.PaypalSendProfile2(csrf)
  s.Task.LogDebug("%+v %+v", ppresp2, err)
  if err != nil {
    return nil, err
  }

  ppsesh, err := s.PaypalCreateSession(ppresp2.Data.FundingID)
  s.Task.LogDebug("%+v %+v", ppsesh, err)
  if err != nil {
    return nil, err
  }

  if ppsesh.Data.State == "NON_PAYABLE" {
    return nil, ErrPaypalCardRejected
  }

  // s.Task.StatusCh <- Status{Message: "Authorizing with PayPal"}
  err = s.PaypalAuthorize()
  s.Task.LogDebug("%+v", err)
  if err != nil {
    return nil, err
  }

  accs, err := s.PaypalGetAccounts(ppsesh.Data.Payer.ID)
  s.Task.LogDebug("%+v %+v",accs, err)
  if err != nil {
    return nil, err
  }

	return &PaypalCheckoutSuccess{
    Nonce: accs.PaypalAccounts[0].Nonce,
    PayerId: accs.PaypalAccounts[0].Details.PayerInfo.PayerID,
  }, nil
}

type PaypalCheckoutSuccess struct {
  Nonce string
  PayerId string
}

type PaypalPaymentResource struct {
	PaymentResource struct {
		PaymentToken    string      `json:"paymentToken"`
		Intent          string      `json:"intent"`
		RedirectURL     string      `json:"redirectUrl"`
		AuthenticateURL interface{} `json:"authenticateUrl"`
	} `json:"paymentResource"`
}

type PaypalApiResponse struct {
	Ack  string `json:"ack"`
  Contingency string `json:"contingency"`
	Data struct {
    AccessToken                string `json:"accessToken"`
    FundingID                  string `json:"fundingId"`
    IsEmailLinkedToFullAccount bool   `json:"isEmailLinkedToFullAccount"`
		Type  string `json:"type"`
		Token string `json:"token"`
	} `json:"data"`
	Meta struct {
		Calc string `json:"calc"`
		Rlog string `json:"rlog"`
	} `json:"meta"`
	Server string `json:"server"`
}

type PaypalConfig struct {
	Challenges   []interface{} `json:"challenges"`
	Environment  string        `json:"environment"`
	ClientAPIURL string        `json:"clientApiUrl"`
	AssetsURL    string        `json:"assetsUrl"`
	AuthURL      string        `json:"authUrl"`
	Analytics    struct {
		URL string `json:"url"`
	} `json:"analytics"`
	ThreeDSecureEnabled bool `json:"threeDSecureEnabled"`
	PaypalEnabled       bool `json:"paypalEnabled"`
	Paypal              struct {
		DisplayName              string      `json:"displayName"`
		ClientID                 string      `json:"clientId"`
		PrivacyURL               string      `json:"privacyUrl"`
		UserAgreementURL         string      `json:"userAgreementUrl"`
		BaseURL                  string      `json:"baseUrl"`
		AssetsURL                string      `json:"assetsUrl"`
		DirectBaseURL            interface{} `json:"directBaseUrl"`
		AllowHTTP                bool        `json:"allowHttp"`
		EnvironmentNoNetwork     bool        `json:"environmentNoNetwork"`
		Environment              string      `json:"environment"`
		UnvettedMerchant         bool        `json:"unvettedMerchant"`
		BraintreeClientID        string      `json:"braintreeClientId"`
		BillingAgreementsEnabled bool        `json:"billingAgreementsEnabled"`
		MerchantAccountID        string      `json:"merchantAccountId"`
		CurrencyIsoCode          string      `json:"currencyIsoCode"`
	} `json:"paypal"`
	MerchantID  string `json:"merchantId"`
	Venmo       string `json:"venmo"`
	CreditCards struct {
		SupportedCardTypes []interface{} `json:"supportedCardTypes"`
		SupportedGateways  []interface{} `json:"supportedGateways"`
		CollectDeviceData  bool          `json:"collectDeviceData"`
	} `json:"creditCards"`
	Kount struct {
	} `json:"kount"`
	FraudProvider struct {
	} `json:"fraudProvider"`
	GraphQL struct {
		URL      string   `json:"url"`
		Date     string   `json:"date"`
		Features []string `json:"features"`
	} `json:"graphQL"`
}

func (s *PaypalCheckoutSession) PaypalGetConfig() (*PaypalConfig, error) {
	configUrl := fmt.Sprintf("https://api.braintreegateway.com/merchants/%s/client_api/v1/configuration?tokenizationKey=%s", s.MerchantId, s.TokenizationKey)
	for key, value := range s.MetaFields {
		configUrl += ("&"+ url.QueryEscape(fmt.Sprintf("_meta[%s]=%s", key, value)))
	}
  url_, err := url.Parse(configUrl)
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "User-Agent",
    "Content-Type",
    "Accept",
    "Origin",
    "Sec-Fetch-Site",
    "Sec-Fetch-Mode",
    "Sec-Fetch-Dest",
    "Referer",
    "Accept-Encoding",
    "Accept-Language",
  }
  headers := [][2]string {
    {"Connection", "keep-alive"},
    {"Content-Type", "application/json"},
    {"Accept", "*/*"},
    {"Origin", "https://" + s.RefUrl.Host},
    {"Sec-Fetch-Site", "cross-site"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", s.RefUrl.String()},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }
  resp, err := s.Task.doReq(s.Client, s.Task.makeReq("GET", url_, &headers, &headerOrder, nil))
  if err != nil {
  	DiscardResp(resp)
  	return nil, err
  }

  var rs PaypalConfig
  err = readRespJsonDst(resp, &rs)
  return &rs, err
}

func (s *PaypalCheckoutSession) PaypalSendProfile() (*PaypalResponse, error) {
  url_, err := url.Parse("https://www.paypal.com/graphql?OnboardGuestMutation")
  if err != nil {
    return nil, err
  }

  cardType := s.Task.Profile.Card.Type
  if cardType == "MASTER" {
    cardType = "MASTER_CARD"
  }
  headerOrder := []string {
    "content-length",
    "x-locale",
    "paypal-client-metadata-id",
    "user-agent",
    "paypal-client-context",
    "accept",
    "content-type",
    "x-country",
    "x-app-name",
    "origin",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"x-locale", "en_US"},
    {"paypal-client-metadata-id", s.ectoken},
    {"paypal-client-context", s.ectoken},
    {"accept", "*/*"},
    {"content-type", "application/json"},
    {"x-country", s.Task.Profile.BillingAddress.Country},
    {"x-app-name", "checkoutuinodeweb_onboarding_lite"},
    {"origin", "https://www.paypal.com"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    // {"referer", "https://www.paypal.com/checkoutweb/signup?version=4.0.173&locale.x=en_US&fundingSource=paypal&sessionID=49f51e57d4_ga3tumrrhiztg&buttonSessionID=6c2f93cda5_ga3tumrrhizti&env=production&logLevel=warn&uid=61698ec11f&token=EC-61W45039CY5790306&xcomponent=1&country.x=US&locale.x=en_US&country.x=US"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  bodyStructure := map[string]interface{} {
  "operationName": "OnboardGuestMutation",
  "variables": map[string]interface{} {
    "card": map[string]interface{} {
      "cardNumber": s.Task.Profile.Card.Number,
      "expirationDate": fmt.Sprintf("%02d/%d", s.Task.Profile.Card.ExpMonth, s.Task.Profile.Card.ExpYear),
      "securityCode": s.Task.Profile.Card.Cvc,
      "type": cardType, // TODO check mastercard, amex, discover type enum
    },
    "country": s.Task.Profile.BillingAddress.Country,
    "email": s.Task.Profile.Email,
    "firstName": s.Task.Profile.BillingAddress.FirstName,
    "lastName": s.Task.Profile.BillingAddress.LastName,
    "phone": map[string]interface{} {
      "countryCode": "1",
      "number": s.Task.Profile.BillingAddress.NordPhone(),
      "type": "MOBILE",
    },
    "supportedThreeDsExperiences": []interface{} {
      "IFRAME",
    },
    "token": s.ectoken,
    "billingAddress": map[string]interface{} {
      "line1": s.Task.Profile.BillingAddress.Address1,
      "line2": s.Task.Profile.BillingAddress.Address2,
      "city": s.Task.Profile.BillingAddress.City,
      "state": s.Task.Profile.BillingAddress.State,
      "postalCode": s.Task.Profile.BillingAddress.Zip,
      "accountQuality": map[string]interface{} {
        "autoCompleteType": "MERCHANT_PREFILLED",
        "isUserModified": true,
        "twoFactorPhoneVerificationId": "",
      },
      "country": s.Task.Profile.BillingAddress.Country,
      "familyName": s.Task.Profile.BillingAddress.LastName,
      "givenName": s.Task.Profile.BillingAddress.FirstName,
    },
    "shippingAddress": map[string]interface{} {
      "line1": s.Task.Profile.ShippingAddress.Address1,
      "line2": s.Task.Profile.ShippingAddress.Address2,
      "city": s.Task.Profile.ShippingAddress.City,
      "state": s.Task.Profile.ShippingAddress.State,
      "postalCode": s.Task.Profile.ShippingAddress.Zip,
      "accountQuality": map[string]interface{} {
        "autoCompleteType": "MERCHANT_PREFILLED",
        "isUserModified": true,
        "twoFactorPhoneVerificationId": "",
      },
      "country": s.Task.Profile.BillingAddress.Country,
      "familyName": s.Task.Profile.ShippingAddress.LastName,
      "givenName": s.Task.Profile.ShippingAddress.FirstName,
    },
  },
  "query": "mutation OnboardGuestMutation($bank: BankAccountInput, $billingAddress: AddressInput, $card: CardInput, $country: CountryCodes, $currencyConversionType: CheckoutCurrencyConversionType, $dateOfBirth: DateOfBirth, $email: String, $firstName: String!, $lastName: String!, $phone: PhoneInput, $shareAddressWithDonatee: Boolean, $shippingAddress: AddressInput, $supportedThreeDsExperiences: [ThreeDSPaymentExperience], $token: String!) {\n  onboardAccount: onboardGuest(bank: $bank, billingAddress: $billingAddress, card: $card, country: $country, currencyConversionType: $currencyConversionType, dateOfBirth: $dateOfBirth, email: $email, firstName: $firstName, lastName: $lastName, phone: $phone, shareAddressWithDonatee: $shareAddressWithDonatee, shippingAddress: $shippingAddress, token: $token) {\n    buyer {\n      auth {\n        accessToken\n        __typename\n      }\n      userId\n      __typename\n    }\n    flags {\n      is3DSecureRequired\n      __typename\n    }\n    paymentContingencies {\n      threeDomainSecure(experiences: $supportedThreeDsExperiences) {\n        status\n        redirectUrl {\n          href\n          __typename\n        }\n        method\n        parameter\n        experience\n        requestParams {\n          key\n          value\n          __typename\n        }\n        __typename\n      }\n      threeDSContingencyData {\n        name\n        causeName\n        resolution {\n          type\n          resolutionName\n          paymentCard {\n            id\n            type\n            number\n            bankIdentificationNumber\n            __typename\n          }\n          contingencyContext {\n            deviceDataCollectionUrl {\n              href\n              __typename\n            }\n            jwtSpecification {\n              jwtDuration\n              jwtIssuer\n              jwtOrgUnitId\n              type\n              __typename\n            }\n            reason\n            referenceId\n            source\n            __typename\n          }\n          __typename\n        }\n        __typename\n      }\n      __typename\n    }\n    __typename\n  }\n}\n",
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  resp, err := s.Task.doReq(s.Client, s.Task.makeReq("POST", url_, &headers, &headerOrder, &body))
  if err != nil {
    return nil, err
  }
  var ppresp PaypalResponse
  err = readRespJsonDst(resp, &ppresp)
  return &ppresp, err
}


type PaypalCheckoutSessionResponse struct {
  Ack  string `json:"ack"`
  Data struct {
    Payer struct {
      ID        string `json:"id"`
      Email     string `json:"email"`
      Phone     string `json:"phone"`
      PhoneType string `json:"phone_type"`
      UserType  string `json:"user_type"`
    } `json:"payer"`
    State    string `json:"state"`
    Shipping struct {
      ID                   string   `json:"id"`
      Name                 string   `json:"name"`
      Line1                string   `json:"line1"`
      Line2                string   `json:"line2"`
      City                 string   `json:"city"`
      State                string   `json:"state"`
      Country              string   `json:"country"`
      PostalCode           string   `json:"postal_code"`
      NormalizationStatus  string   `json:"normalization_status"`
      PreferredAddress     bool     `json:"preferred_address"`
      Default              bool     `json:"default"`
      S2SAddress           bool     `json:"s2s_address"`
      FullAddress          string   `json:"full_address"`
      FullAddressMultiline []string `json:"full_address_multiline"`
    } `json:"shipping"`
    Plan struct {
      ID      string `json:"id"`
      Sources []struct {
        Name               string `json:"name"`
        Type               string `json:"type"`
        SubType            string `json:"sub_type"`
        CountryCode        string `json:"country_code"`
        UnconfirmedBank    bool   `json:"unconfirmed_bank"`
        CurrencyCode       string `json:"currency_code"`
        Last4              string `json:"last4"`
        CardType           string `json:"card_type"`
        InstrumentID       string `json:"instrument_id"`
        IsPinlessDebit     bool   `json:"is_pinless_debit"`
        IsPreferred        bool   `json:"is_preferred"`
        IsBillingPreferred bool   `json:"isBillingPreferred"`
        Amount             struct {
          Amount                  string `json:"amount"`
          AmountFormatted         string `json:"amount_formatted"`
          AmountFormattedForex    string `json:"amount_formatted_forex"`
          AmountFormattedNumber   string `json:"amount_formatted_number"`
          AmountFormattedCurrency string `json:"amount_formatted_currency"`
          CurrencyCode            string `json:"currency_code"`
          ForexInCurrencyFormat   string `json:"forex_in_currency_format"`
        } `json:"amount"`
      } `json:"sources"`
      Backup         interface{} `json:"backup"`
      Conversion     interface{} `json:"conversion"`
      SoftDescriptor string      `json:"soft_descriptor"`
      PaypalCredit   bool        `json:"paypal_credit"`
      Default        bool        `json:"default"`
      CvvData        interface{} `json:"cvvData"`
    } `json:"plan"`
    FundingOptions struct {
      DisallowedInstruments []struct {
        Type               string `json:"type"`
        ID                 string `json:"id"`
        Name               string `json:"name"`
        LastDigits         string `json:"lastDigits"`
        IsPinlessDebit     bool   `json:"isPinlessDebit"`
        IsPreferred        bool   `json:"isPreferred"`
        IsBillingPreferred bool   `json:"isBillingPreferred"`
        ProductClass       string `json:"productClass"`
        ExpireYear         string `json:"expire_year"`
        ExpireMonth        string `json:"expire_month"`
        PartialFpan        bool   `json:"partial_fpan"`
        DeclineReason      string `json:"declineReason"`
        DisplayName        string `json:"displayName"`
      } `json:"disallowedInstruments"`
      Instruments []struct {
        Type           string `json:"type"`
        ID             string `json:"id"`
        Name           string `json:"name"`
        LastDigits     string `json:"lastDigits"`
        IsPinlessDebit bool   `json:"isPinlessDebit"`
        Amount         struct {
          Amount                  string `json:"amount"`
          AmountFormatted         string `json:"amount_formatted"`
          AmountFormattedForex    string `json:"amount_formatted_forex"`
          AmountFormattedNumber   string `json:"amount_formatted_number"`
          AmountFormattedCurrency string `json:"amount_formatted_currency"`
          CurrencyCode            string `json:"currency_code"`
          ForexInCurrencyFormat   string `json:"forex_in_currency_format"`
        } `json:"amount"`
        IsPreferred        bool   `json:"isPreferred"`
        IsBillingPreferred bool   `json:"isBillingPreferred"`
        ProductClass       string `json:"productClass"`
        ExpireYear         string `json:"expire_year"`
        ExpireMonth        string `json:"expire_month"`
        PartialFpan        bool   `json:"partial_fpan"`
        PlanID             string `json:"planId"`
        Selected           bool   `json:"selected"`
        EditDisallowedMsg  string `json:"editDisallowedMsg"`
      } `json:"instruments"`
      Plans []struct {
        PlanID    string `json:"planId"`
        PrimaryFs struct {
          Type           string `json:"type"`
          ID             string `json:"id"`
          Name           string `json:"name"`
          LastDigits     string `json:"lastDigits"`
          IsPinlessDebit bool   `json:"isPinlessDebit"`
          Amount         struct {
            Amount                  string `json:"amount"`
            AmountFormatted         string `json:"amount_formatted"`
            AmountFormattedForex    string `json:"amount_formatted_forex"`
            AmountFormattedNumber   string `json:"amount_formatted_number"`
            AmountFormattedCurrency string `json:"amount_formatted_currency"`
            CurrencyCode            string `json:"currency_code"`
            ForexInCurrencyFormat   string `json:"forex_in_currency_format"`
          } `json:"amount"`
          IsPreferred        bool   `json:"isPreferred"`
          IsBillingPreferred bool   `json:"isBillingPreferred"`
          ProductClass       string `json:"productClass"`
          ExpireYear         string `json:"expire_year"`
          ExpireMonth        string `json:"expire_month"`
          PartialFpan        bool   `json:"partial_fpan"`
        } `json:"primaryFs"`
        SecondaryFIs        bool          `json:"secondaryFIs"`
        AllIncentiveOptions []interface{} `json:"allIncentiveOptions"`
        Selected            bool          `json:"selected"`
      } `json:"plans"`
    } `json:"fundingOptions"`
    BmlOffer                      int         `json:"bmlOffer"`
    Installments                  interface{} `json:"installments"`
    SepaMandate                   interface{} `json:"sepa_mandate"`
    Resolve3Ds                    interface{} `json:"resolve_3ds"`
    UnconfirmedUpop               interface{} `json:"unconfirmed_upop"`
    IsCreditDisallowedForMerchant bool        `json:"isCreditDisallowedForMerchant"`
    ErrorData                     struct {
    } `json:"errorData"`
    CvvData                       interface{} `json:"cvvData"`
    NegBals                       interface{} `json:"negBals"`
    OrigContName                  interface{} `json:"origContName"`
    ShowRemittanceDisclosure      interface{} `json:"showRemittanceDisclosure"`
    ThreeDSResolutionData         interface{} `json:"threeDSResolutionData"`
    ThreeDSPayerAction            interface{} `json:"threeDSPayerAction"`
    PayerConsentResolutionData    interface{} `json:"payerConsentResolutionData"`
    RealTimeBalanceResolutionData interface{} `json:"realTimeBalanceResolutionData"`
  } `json:"data"`
  Meta struct {
    Calc string `json:"calc"`
    Rlog string `json:"rlog"`
  } `json:"meta"`
  Server string `json:"server"`
}

func (s *PaypalCheckoutSession) PaypalCreateSession(cardId string) (*PaypalCheckoutSessionResponse, error) {
  url_, err := url.Parse(fmt.Sprintf("https://www.paypal.com/webapps/xoonboarding/api/checkout/%s/session/create", s.ectoken))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-requested-with",
    "x-csrf-jwt",
    "user-agent",
    "content-type",
    "origin",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "*/*"},
    {"x-requested-with", "XMLHttpRequest"},
    {"x-csrf-jwt", s.csrfheader},
    {"content-type", "application/json;charset=UTF-8"},
    {"origin", "https://www.paypal.com"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    // {"referer", "https://www.paypal.com/webapps/xoonboarding?buttonSessionID=1047761972_gizdumrwhi2tk&env=production&fundingSource=paypal&locale.x=en_US&logLevel=warn&logoutFrom=PAYPAL&sessionID=ccafb960d5_gizdumrwhi2ti&token=EC-59085267G18130741&uid=57459f2110&ulClientRedirect=guest_logout&version=4.0.173&xcomponent=1&country.x=US&locale.x=en_US&country.x=US"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  bodyStructure := map[string]interface{} {
  "data": map[string]interface{} {
    "funding_instruments": []interface{} {
      map[string]interface{} {
              "payment_card": map[string]interface{} {
                "id": cardId,
              },
            },
    },
    "shipping_address": map[string]interface{} {
      "first_name": s.Task.Profile.ShippingAddress.FirstName,
      "last_name": s.Task.Profile.ShippingAddress.LastName,
      "recipient_name": s.Task.Profile.ShippingAddress.FirstName + " "  + s.Task.Profile.ShippingAddress.LastName,
      "line1": s.Task.Profile.ShippingAddress.Address1,
      "line2": s.Task.Profile.ShippingAddress.Address2,
      "city": s.Task.Profile.ShippingAddress.City,
      "state": s.Task.Profile.ShippingAddress.State,
      "postal_code": s.Task.Profile.ShippingAddress.Zip,
      "country": s.Task.Profile.ShippingAddress.Country,
    },
  },
  "meta": map[string]interface{} {
    "token": s.ectoken,
    "calc": randHex(7)[:13],
    "csci": randHex(16),
    "locale": map[string]interface{} {
      "country": s.Task.Profile.ShippingAddress.Country,
      "language": "en",
    },
    "state": "ui_checkout_guest",
    "app_name": "xoonboardingnodeweb",
  },
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  var rr PaypalCheckoutSessionResponse
  resp, err := s.Task.doReq(s.Client, s.Task.makeReq("POST", url_, &headers, &headerOrder, &body))
  bb, _ := readBodyBytes(resp)
  if err != nil {
    if resp != nil {
      json.Unmarshal(bb, &rr)
      go s.Task.SendTelemetry(map[string]interface{} {
        "event": "paypal_checkout_session_create_failed",
        "paypal_checkout_session_create_failure": string(bb),
        "paypal_checkout_session_create_failure_status": resp.StatusCode,
        "paypal_checkout_session_create_err": fmt.Sprintf("%+v", err),
      })
      if len(rr.Data.FundingOptions.DisallowedInstruments) > 0 {
        return nil, ErrPaypalCardRejected
      }
    }
    return nil, err
  }
  if resp.Header.Get("x-csrf-jwt") != "" {
    s.csrfheader = resp.Header.Get("x-csrf-jwt")
  }
  err = json.Unmarshal(bb, &rr)
  if rr.Ack != "success" {
    go s.Task.SendTelemetry(map[string]interface{} {
      "event": "paypal_checkout_session_create_failed",
      "paypal_checkout_session_create_failure": string(bb),
      "paypal_checkout_session_create_failure_status": resp.StatusCode,
    })
    return nil, ErrPaypalCardRejected
  }
  return &rr, err
}


func (s *PaypalCheckoutSession) PaypalAuthorize() (error) {
  url_, err := url.Parse(fmt.Sprintf("https://www.paypal.com/webapps/xoonboarding/api/checkout/%s/session/authorize", s.ectoken))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-requested-with",
    "x-csrf-jwt",
    "user-agent",
    "content-type",
    "origin",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "*/*"},
    {"x-requested-with", "XMLHttpRequest"},
    {"x-csrf-jwt", s.csrfheader},
    {"content-type", "application/json;charset=UTF-8"},
    {"origin", "https://www.paypal.com"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    // {"referer", "https://www.paypal.com/webapps/xoonboarding?buttonSessionID=1047761972_gizdumrwhi2tk&env=production&fundingSource=paypal&locale.x=en_US&logLevel=warn&logoutFrom=PAYPAL&sessionID=ccafb960d5_gizdumrwhi2ti&token=EC-59085267G18130741&uid=57459f2110&ulClientRedirect=guest_logout&version=4.0.173&xcomponent=1&country.x=US&locale.x=en_US&country.x=US"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  bodyStructure := map[string]interface{} {
  "meta": map[string]interface{} {
    "token": s.ectoken,
    "calc": randHex(7)[:13],
    "csci": randHex(16),
    "locale": map[string]interface{} {
      "country": s.Task.Profile.ShippingAddress.Country,
      "language": "en",
    },
    "state": "ui_checkout_guest",
    "app_name": "xoonboardingnodeweb",
  },
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  resp, err := s.Task.doReq(s.Client, s.Task.makeReq("POST", url_, &headers, &headerOrder, &body))
  DiscardResp(resp)
  if resp != nil && resp.Header.Get("x-csrf-jwt") != "" {
    s.csrfheader = resp.Header.Get("x-csrf-jwt")
  }
  return err
}

type PaypalAccountsResponse struct {
  PaypalAccounts []struct {
    Type        string `json:"type"`
    Nonce       string `json:"nonce"`
    Description string `json:"description"`
    Consumed    bool   `json:"consumed"`
    Details     struct {
      PayerInfo struct {
        Email           string `json:"email"`
        FirstName       string `json:"firstName"`
        LastName        string `json:"lastName"`
        PayerID         string `json:"payerId"`
        ShippingAddress struct {
          RecipientName string `json:"recipientName"`
          Line1         string `json:"line1"`
          City          string `json:"city"`
          State         string `json:"state"`
          PostalCode    string `json:"postalCode"`
          CountryCode   string `json:"countryCode"`
        } `json:"shippingAddress"`
        Phone       string `json:"phone"`
        CountryCode string `json:"countryCode"`
      } `json:"payerInfo"`
      CorrelationID   string      `json:"correlationId"`
      BillingAddress  interface{} `json:"billingAddress"`
      ShippingAddress struct {
        RecipientName string `json:"recipientName"`
        Line1         string `json:"line1"`
        City          string `json:"city"`
        State         string `json:"state"`
        PostalCode    string `json:"postalCode"`
        CountryCode   string `json:"countryCode"`
      } `json:"shippingAddress"`
    } `json:"details"`
  } `json:"paypalAccounts"`
}
func (s *PaypalCheckoutSession) PaypalGetAccounts(payerId string) (*PaypalAccountsResponse, error) {
  url_, err := url.Parse("https://api.braintreegateway.com/merchants/rfbkw27jcwmw2xgp/client_api/v1/payment_methods/paypal_accounts")
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Content-Length",
    "User-Agent",
    "Content-Type",
    "Accept",
    "Origin",
    "Sec-Fetch-Site",
    "Sec-Fetch-Mode",
    "Sec-Fetch-Dest",
    "Referer",
    "Accept-Encoding",
    "Accept-Language",
  }
  headers := [][2]string {
    {"Connection", "keep-alive"},
    {"Content-Type", "application/json"},
    {"Accept", "*/*"},
    {"Origin", "https://www.footlocker.com"},
    {"Sec-Fetch-Site", "cross-site"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", "https://www.footlocker.com/cart"},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }
  bodyStructure := map[string]interface{} {
  "paypalAccount": map[string]interface{} {
    "correlationId": s.ectoken,
    "options": map[string]interface{} {
      "validate": false,
    },
    "paymentToken": s.payResource.PaymentResource.PaymentToken,
    "payerId": payerId,
    "unilateral": false,
    "intent": "authorize",
  },
  "braintreeLibraryVersion": "braintree/web/3.29.0",
  "_meta": s.MetaFields,
  "tokenizationKey": s.TokenizationKey,
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  resp, err := s.Task.doReq(s.Client, s.Task.makeReq("POST", url_, &headers, &headerOrder, &body))
  if err != nil {
    return nil, err
  }
  var rr PaypalAccountsResponse
  err = readRespJsonDst(resp, &rr)
  return &rr, err
}

var ErrGuestCheckoutDisabled = errors.New("Guest checkout disabled")

func (s *PaypalCheckoutSession) PaypalGetCheckoutNow() (error) {
  url_, err := url.Parse(
    fmt.Sprintf("https://www.paypal.com/checkoutnow?buttonSessionID=%s&env=production&fundingSource=paypal&locale.x=en_US&logLevel=warn&logoutFrom=PAYPAL&sessionID=%s&token=%s&uid=%s&ulClientRedirect=guest_logout&version=4.0.173&xcomponent=1",
    s.buttonSessionID, s.sessionId, s.ectoken, s.uid))

  if err != nil {
    return err
  }
  headerOrder := []string {
    "upgrade-insecure-requests",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-dest", "document"},
    {"referer", url_.String()},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := s.Task.doReq(s.Client, s.Task.makeReq("GET", url_, &headers, &headerOrder, nil))
  if body, err := readBody(resp); err != nil {
    if strings.Contains(body, "startGuestOnboardingFlow") {
      return nil
    } else {
      return ErrGuestCheckoutDisabled
    }
  } else {
    DiscardResp(resp)
  }
  return err
}

func (s *PaypalCheckoutSession) PaypalSetInitialHeaders() (error) {
  headerOrder := []string {
    "upgrade-insecure-requests",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "cross-site"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-dest", "iframe"},
    {"referer", "https://www.footlocker.com/"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := s.Task.doReq(s.Client, s.Task.makeReq("GET", s.ButtonUrl, &headers, &headerOrder, nil))
  if err != nil {
    return err
  }
 	body, err := readBody(resp)
 	if err != nil {
 	  return err
 	}
 	re := regexp.MustCompile(`"x-cookies":.+?}}`)
 	matches := re.FindStringSubmatch(body)
 	if len(matches) == 0 {
 		return errors.New("No headers in button response")
 	}
 	jj := matches[0]
 	jj = "{" + jj[:len(jj)-1]
 	var headerMap map[string]string
 	err = json.Unmarshal([]byte(jj), &headerMap)
 	if err != nil {
 		return err
 	}
 	s.Task.LogDebug("%+v", headerMap)

 	s.Headers = [][2]string{}
 	for key, value := range headerMap {
    if strings.ToLower(key) == "x-cookies-hash" {
      continue
    } else {
      s.Headers = append(s.Headers, [2]string{key, value})
    }
    // else if strings.ToLower(key) == "x-cookies" {
    //   json.Unmarshal([]byte(value), &s.XCookies)
    // }

 	}
 	return nil
}

func (s *PaypalCheckoutSession) GetXCookiesHeader() [2]string {
  var key = "tD08unW5xWPYcc3Vtbf3fJ3V3AQpBSPfm6WSV5oz4qyqFW9g"
  mm := map[string]string {}
  mm[key] = s.XCookies[key]
  bb, _ := json.Marshal(mm)
  return [2]string{"x-cookies", string(bb)}
}

func (s *PaypalCheckoutSession) PaypalVisitXoon() (string, error) {
  url_, err := url.Parse(fmt.Sprintf("https://www.paypal.com/webapps/xoonboarding?version=4.0.173&locale.x=en_US&fundingSource=paypal&sessionID=%s&buttonSessionID=%s&env=production&logLevel=warn&uid=%s&token=%s&xcomponent=1&country.x=US&fallback=1",
    s.sessionId, s.buttonSessionID, s.uid, s.ectoken,
  ))
  if err != nil {
    return "", err
  }
  headerOrder := []string {
    "upgrade-insecure-requests",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-dest", "document"},
    // {"referer", "https://www.paypal.com/checkoutweb/signup?version=4.0.173&locale.x=en_US&fundingSource=paypal&sessionID=49f51e57d4_ga3tumrrhiztg&buttonSessionID=6c2f93cda5_ga3tumrrhizti&env=production&logLevel=warn&uid=61698ec11f&token=EC-61W45039CY5790306&xcomponent=1&country.x=US&locale.x=en_US&country.x=US"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := s.Task.doReq(s.Client, s.Task.makeReq("GET", url_, &headers, &headerOrder, nil))
  DiscardResp(resp)
  if resp != nil {
    return resp.Header.Get("x-csrf-jwt"), err
  } else {
    return "", err
  }
}



func (s *PaypalCheckoutSession) PaypalInitSeshNativeElig() (error) {
  url_, err := url.Parse("https://www.paypal.com/graphql")
  if err != nil {
    return err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-requested-with",
    "x-app-name",
    "user-agent",
    "content-type",
    "origin",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "application/json"},
    {"x-requested-with", "XMLHttpRequest"},
    {"x-app-name", "xo_buttonjs"},
    {"content-type", "application/json"},
    {"origin", "https://www.paypal.com"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    // {"referer", "https://www.paypal.com/webapps/hermes/button?version=4.0.173&env=production&style.size=responsive&style.color=silver&style.shape=rect&style.tagline=false&sessionID=9b03654528_gizdumrrhi2ti&buttonSessionID=be4c2a43cb_gizdumrshizdq&funding.disallowed=venmo&funding.remembered=paypal&locale.x=en_US&logLevel=warn&uid=5220455acf&xcomponent=1"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  ckiesummary := ""
  for _, ckie := range *GetAllClientCookies(s.Client) {
    if !ckie.HttpOnly && strings.HasSuffix(ckie.Domain, "paypal.com") {
      ckiesummary += ckie.Name + "=x;"
    }
  }
  bodyStructure := map[string]interface{} {
  "query": "\n            query NativeEligibility(\n                $shippingCallbackEnabled : Boolean,\n                $facilitatorClientID : String,\n                $buyerCountry : String,\n                $currency : String,\n                $userAgent : String,\n                $buttonSessionID : String,\n                $cookies : String,\n                $version : String,\n                $domain : String\n            ) {\n                mobileSDKEligibility(\n                    shippingCallbackEnabled: $shippingCallbackEnabled,\n                    facilitatorClientID: $facilitatorClientID,\n                    buyerCountry: $buyerCountry,\n                    currency: $currency,\n                    userAgent: $userAgent,\n                    buttonSessionID: $buttonSessionID,\n                    cookies : $cookies,\n                    sdkjsVersion : $version,\n                    domain : $domain\n                ) {\n                    paypal {\n                        eligibility\n                        ineligibilityReason\n                    }\n                    venmo {\n                        eligibility\n                        ineligibilityReason\n                    }\n                }\n            }\n        ",
  "variables": map[string]interface{} {
    "shippingCallbackEnabled": false,
    "facilitatorClientID": nil,
    "buyerCountry": s.Task.Profile.BillingAddress.Country,
    // todo use real, mobile might behave diff ?
    "userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36",
    "buttonSessionID": s.buttonSessionID,
    "cookies": ckiesummary,
    "version": "4.0.173",
    "domain": "https://www.footlocker.com",
  },
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  resp, err := s.Task.doReq(s.Client, s.Task.makeReq("POST", url_, &headers, &headerOrder, &body))
  DiscardResp(resp)
  return err
}



func (s *PaypalCheckoutSession) GetPaypalEctoken(payid string) (*PaypalApiResponse, error) {
  url_, err := url.Parse(fmt.Sprintf("https://www.paypal.com/webapps/hermes/api/payment/%s/ectoken", payid))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-requested-with",
    "x-cookies",
    "x-csrf-jwt",
    "x-requested-by",
    "user-agent",
    "content-type",
    "origin",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "application/json"},
    {"x-requested-with", "XMLHttpRequest"},
    {"x-csrf-jwt", getCookieValue("x-csrf-jwt", PaypalBaseUrl, s.Client.Jar)},
    {"x-requested-by", "smart-payment-buttons"},
    {"content-type", "application/json"},
    {"origin", "https://www.paypal.com"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", s.ButtonUrl.String()},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
    {"user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36"},
    // s.GetXCookiesHeader(),
  }
  for _, pair := range s.Headers {
  	headers = append(headers, pair)
  }
  bodyStructure := map[string]interface{} {
  "meta": map[string]interface{} {

  },
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  jbak := s.Client.Jar
  s.Client.Jar = nil
  defer func() {
    s.Client.Jar = jbak
  }()
  resp, err := s.Task.doReq(s.Client, s.Task.makeReq("POST", url_, &headers, &headerOrder, &body))
  if err != nil {
  	DiscardResp(resp)
  	return nil, err
  }

  var rs PaypalApiResponse
  err = readRespJsonDst(resp, &rs)
  return &rs, err
}

func (s *PaypalCheckoutSession) PaypalInitSeshGraphql() (error) {
  url_, err := url.Parse("https://www.paypal.com/targeting/graphql")
  if err != nil {
    return err
  }
  headerOrder := []string {
    "content-length",
    "user-agent",
    "content-type",
    "accept",
    "origin",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"content-type", "application/json"},
    {"accept", "*/*"},
    {"origin", "https://www.paypalobjects.com"},
    {"sec-fetch-site", "cross-site"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    // {"referer", "https://www.paypalobjects.com/muse/analytics/index.html?frameId=8b2e47f8-40ab-4532-9019-1a7aebedf648&propertyId=2NJ4YGHSGAPCC-1&flow=visitor-info&variant=analytics&mrid=2NJ4YGHSGAPCC&isMobileEnabled=true&isDesktopEnabled=true&shouldCheckCountry=true&mobileVariant=analytics&mobileFlow=visitor-info"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  // TODO match api ua
  deviceInfoMap := map[string]interface{} {
    "screenWidth":1920,
    "screenHeight":1200,
    "colorDepth":24,
    "rosettaLanguage":"en-US,en",
    "deviceType":"Desktop",
    "browserHeight":2666,
    "browserWidth":1905,
  }
  dibytes, err := json.Marshal(deviceInfoMap)
  if err != nil {
    return err
  }
  bodyStructure := map[string]interface{} {
    "query": fmt.Sprintf("{ visitorInfo(\n            country: \"N/A\",\n            deviceInfo: \"%s\"\n          ) }", url.QueryEscape(string(dibytes))),
  }
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return err
  }
  resp, err := s.Task.doReq(s.Client, s.Task.makeReq("POST", url_, &headers, &headerOrder, &body))
  DiscardResp(resp)
  return err
}


func (s *PaypalCheckoutSession) PaypalCreatePaymentResource() (*PaypalPaymentResource, error) {
  url_, err := url.Parse("https://api.braintreegateway.com/merchants/rfbkw27jcwmw2xgp/client_api/v1/paypal_hermes/create_payment_resource")
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "Host",
    "Connection",
    "Content-Length",
    "User-Agent",
    "Content-Type",
    "Accept",
    "Origin",
    "Sec-Fetch-Site",
    "Sec-Fetch-Mode",
    "Sec-Fetch-Dest",
    "Referer",
    "Accept-Encoding",
    "Accept-Language",
  }
  headers := [][2]string {
    {"Connection", "keep-alive"},
    {"Content-Type", "application/json"},
    {"Accept", "*/*"},
    {"Origin", "https://" + s.RefUrl.Host},
    {"Sec-Fetch-Site", "cross-site"},
    {"Sec-Fetch-Mode", "cors"},
    {"Sec-Fetch-Dest", "empty"},
    {"Referer", s.RefUrl.String()},
    {"Accept-Encoding", "gzip, deflate, br"},
    {"Accept-Language", "en-US,en;q=0.9"},
  }
  bodyStructure := map[string]interface{} {
	  "returnUrl": "x",
	  "cancelUrl": "x",
	  "offerPaypalCredit": false,
	  "experienceProfile": s.ExperienceProfile,
	  "amount": s.Amount,
	  "currencyIsoCode": "USD",
	  "intent": "authorize",
	  "line1": s.Task.Profile.BillingAddress.Address1,
    "line2": s.Task.Profile.BillingAddress.Address2,
	  "city": strings.ToUpper(s.Task.Profile.BillingAddress.City),
	  "state": s.Task.Profile.BillingAddress.City,
	  "postalCode": s.Task.Profile.BillingAddress.Zip,
	  "countryCode": s.Task.Profile.BillingAddress.Country,
	  "phone": s.Task.Profile.BillingAddress.NordPhone(),
	  "recipientName": s.Task.Profile.BillingAddress.FirstName + " " + s.Task.Profile.BillingAddress.LastName,
	  "braintreeLibraryVersion": "braintree/web/3.29.0",
	  "_meta": s.MetaFields,
	  "tokenizationKey": s.TokenizationKey,
	}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  req := s.Task.makeReq("POST", url_, &headers, &headerOrder, &body)
  resp, err := s.Task.doReq(s.Client, req)
  if err != nil {
  	DiscardResp(resp)
  	return nil, err
  }

  var rs PaypalPaymentResource
  err = readRespJsonDst(resp, &rs)
  return &rs, err
}

func (s *PaypalCheckoutSession) PaypalSendProfile2(csrf string) (*PaypalApiResponse, error) {
  url_, err := url.Parse("https://www.paypal.com/webapps/xoonboarding/api/onboard/guest")
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "content-length",
    "accept",
    "x-requested-with",
    "x-csrf-jwt",
    "user-agent",
    "content-type",
    "origin",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "*/*"},
    {"x-requested-with", "XMLHttpRequest"},
    {"content-type", "application/json;charset=UTF-8"},
    {"origin", "https://www.paypal.com"},
    {"sec-fetch-site", "same-origin"},
    {"x-csrf-jwt", csrf},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    // {"referer", "https://www.paypal.com/webapps/xoonboarding?version=4.0.173&locale.x=en_US&fundingSource=paypal&sessionID=49f51e57d4_ga3tumrrhiztg&buttonSessionID=6c2f93cda5_ga3tumrrhizti&env=production&logLevel=warn&uid=61698ec11f&token=EC-61W45039CY5790306&xcomponent=1&country.x=US&fallback=1&reason=dW5oYW5kbGVkQ29udGluZ2VuY3k%3D"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  cardType := s.Task.Profile.Card.Type
  if cardType == "MASTER" {
    cardType = "MASTER_CARD"
  }
  bodyStructure := map[string]interface{} {
  "data": map[string]interface{} {
    "user": map[string]interface{} {
      "first_name": s.Task.Profile.ShippingAddress.FirstName,
      "last_name": s.Task.Profile.ShippingAddress.LastName,
      "email": s.Task.Profile.Email,
      "countryOfResidence": s.Task.Profile.BillingAddress.Country,
      "country": s.Task.Profile.BillingAddress.Country,
      "nationality": s.Task.Profile.BillingAddress.Country,
    },
    "billing_address": map[string]interface{} {
      "line1": s.Task.Profile.BillingAddress.Address1,
      "line2": s.Task.Profile.BillingAddress.Address2,
      "city": s.Task.Profile.BillingAddress.City,
      "state": s.Task.Profile.BillingAddress.State,
      "postal_code": s.Task.Profile.BillingAddress.Zip,
      "country": s.Task.Profile.BillingAddress.Country,
    },
    "shipping_address": map[string]interface{} {
      "first_name": s.Task.Profile.ShippingAddress.FirstName,
      "last_name": s.Task.Profile.ShippingAddress.LastName,
      "line1": s.Task.Profile.ShippingAddress.Address1,
      "line2": s.Task.Profile.ShippingAddress.Address2,
      "city": s.Task.Profile.ShippingAddress.City,
      "state": s.Task.Profile.ShippingAddress.State,
      "postal_code": s.Task.Profile.ShippingAddress.Zip,
      "country": s.Task.Profile.BillingAddress.Country,
    },
    "phone": map[string]interface{} {
      "type": "Mobile",
      "number": s.Task.Profile.ShippingAddress.NordPhone(),
      "countryCode": "1",
    },
    "shipping_address_validation": false,
    "billing_address_present": true,
    "prox_flow": false,
    "testParams": map[string]interface{} {

    },
    "content_identifier": "US:en:3.0.262:undefined",
    "frs_optin_flow_applicable": false,
    "frs_optin_active_flow": false,
    "flow_eligibility_data": map[string]interface{} {
      "is_unbranded": false,
      "merchant_preferences": map[string]interface{} {
        "id": "G34JZXYXKZ4YW",
        "meta": map[string]interface{} {
          "populated": true,
        },
        "auto_return": map[string]interface{} {
          "url": "http://www.cardinalcommerce.com",
          "enabled": true,
        },
        "pdt": map[string]interface{} {
          "enabled": true,
        },
        "account_optional": map[string]interface{} {
          "enabled": true,
        },
        "merchant_blacklist": map[string]interface{} {
          "enabled": false,
        },
        "merchant_vertical_high_risk": map[string]interface{} {
          "enabled": false,
        },
        "charset": "windows-1252",
      },
      "merchant_country": s.Task.Profile.BillingAddress.Country,
    },
    "bypass_address_verification": true,
    "bypass_authorization": true,
    "card": map[string]interface{} {
      "type": cardType,
      "number": s.Task.Profile.Card.Number,
      "security_code": s.Task.Profile.Card.Cvc,
      "expiry_month": fmt.Sprintf("%02d", s.Task.Profile.Card.ExpMonth),
      "expiry_year": fmt.Sprintf("%d", s.Task.Profile.Card.ExpYear),
    },
    "skipInitiateAuth": true,
  },
  "meta": map[string]interface{} {
    "token": s.ectoken,
    "calc": randHex(7)[:13], // todo
    "csci": randHex(16),
    "locale": map[string]interface{} {
      "country": s.Task.Profile.BillingAddress.Country,
      "language": "en",
    },
    "state": "ui_checkout_guest",
    "app_name": "xoonboardingnodeweb",
  },
}
  body, err := json.Marshal(bodyStructure)
  if err != nil {
    return nil, err
  }
  resp, err := s.Task.doReq(s.Client, s.Task.makeReq("POST", url_, &headers, &headerOrder, &body))
  if resp != nil {
    if resp.Header.Get("x-csrf-jwt") != "" {
      s.csrfheader = resp.Header.Get("x-csrf-jwt")
    }
    bb, _ := readBodyBytes(resp)
    go s.Task.SendTelemetry(map[string]interface{} {
      "event": "pp_profile_response",
      "pp_profile_response": map[string]interface{} {
        "body": string(bb),
      },
    })
    var jj PaypalApiResponse
    json.Unmarshal(bb, &jj)
    if jj.Ack != "success" {
      return nil, ErrPaypalCardRejected
    }
    return &jj, err
  }
  return nil, err
}



type PaypalResponse struct {
  Errors []struct {
    Message     string   `json:"message"`
    Path        []string `json:"path"`
    Stack       string   `json:"stack"`
    Checkpoints []string `json:"checkpoints"`
    Meta        struct {
    } `json:"meta"`
    ErrorData struct {
      Num0 struct {
        Field string `json:"field"`
        Code  string `json:"code"`
      } `json:"0"`
      AccessToken string `json:"accessToken"`
    } `json:"errorData"`
    Contingency bool `json:"contingency"`
    StatusCode  int  `json:"statusCode"`
  } `json:"errors"`
  Data *PaypalResponseData `json:"data"`
}

type PaypalResponseData struct {
  OnboardAccount struct {
    Buyer struct {
      Auth struct {
        AccessToken string `json:"accessToken"`
        Typename    string `json:"__typename"`
      } `json:"auth"`
      UserID   string `json:"userId"`
      Typename string `json:"__typename"`
    } `json:"buyer"`
    Flags struct {
      Is3DSecureRequired bool   `json:"is3DSecureRequired"`
      Typename           string `json:"__typename"`
    } `json:"flags"`
    PaymentContingencies struct {
      ThreeDomainSecure      interface{} `json:"threeDomainSecure"`
      ThreeDSContingencyData interface{} `json:"threeDSContingencyData"`
      Typename               string      `json:"__typename"`
    } `json:"paymentContingencies"`
    Typename string `json:"__typename"`
  } `json:"onboardAccount"`
}
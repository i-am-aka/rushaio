package task

import (
	"encoding/json"
	"math/rand"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	// "rush/net/http"
	"log"
)
type TwProduct struct {
	Brand struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"brand"`
	Categories []struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Type          string `json:"type"`
		URL           string `json:"url"`
		StorefrontURL string `json:"storefrontUrl"`
	} `json:"categories"`
	ContainerType         string  `json:"containerType"`
	CustomerAverageRating float64 `json:"customerAverageRating"`
	CustomerReviewsCount  int     `json:"customerReviewsCount"`
	Department            string  `json:"department"`
	DirectType            string  `json:"directType"`
	ID                    string  `json:"id"`
	Images                []struct {
		ImageType          string `json:"imageType"`
		MobileOptimizedURL string `json:"mobileOptimizedUrl"`
		ThumbnailURL       string `json:"thumbnailUrl"`
		URL                string `json:"url"`
		ZoomImageURL       string `json:"zoomImageUrl"`
		AltText            string `json:"altText"`
	} `json:"images"`
	Name               string `json:"name"`
	PackageDescription string `json:"packageDescription"`
	Price              []struct {
		Price float64 `json:"price"`
		Type  string  `json:"type"`
	} `json:"price"`
	ProductURL    string `json:"productUrl"`
	Review        string `json:"review"`
	Rating        int    `json:"rating"`
	SalesStrategy struct {
		Name string `json:"name"`
	} `json:"salesStrategy"`
	ShoppingOptions []struct {
		Eligible bool   `json:"eligible"`
		Location string `json:"location"`
		Type     string `json:"type"`
		Selected bool   `json:"selected"`
	} `json:"shoppingOptions"`
	SkuID      string `json:"skuId"`
	StockLevel []struct {
		PurchaseLimit int `json:"purchaseLimit"`
		Stock         int `json:"stock"`
	} `json:"stockLevel"`
	StoreDistance    float64 `json:"storeDistance"`
	StoreID          string  `json:"storeId"`
	StoreName        string  `json:"storeName"`
	ItemTasteProfile string  `json:"itemTasteProfile"`
	ItemStyle        string  `json:"itemStyle"`
	ItemBody         string  `json:"itemBody"`
	Transactional    bool    `json:"transactional"`
	Type             string  `json:"type"`
	Volume           string  `json:"volume"`
	RatingSource     string  `json:"ratingSource,omitempty"`
}

type TwProductApiResp struct {
	SearchText   string `json:"searchText"`
	CategoryName string `json:"categoryName"`
	Pagination   struct {
		Page         int    `json:"page"`
		PageSize     int    `json:"pageSize"`
		TotalPages   int    `json:"totalPages"`
		TotalResults int    `json:"totalResults"`
		NextPageURL  string `json:"nextPageUrl"`
	} `json:"pagination"`
	Facets []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		MultiSelect bool   `json:"multiSelect"`
		Priority    int    `json:"priority"`
		Values      []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Count    int    `json:"count"`
			Selected bool   `json:"selected"`
			Priority int    `json:"priority"`
			URL      string `json:"url"`
		} `json:"values"`
	} `json:"facets"`
	Sorts []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Selected bool   `json:"selected"`
		URL      string `json:"url"`
	} `json:"sorts"`
	RedirectionURL string `json:"redirectionUrl"`
	Products       []TwProduct `json:"products"`
	AutoCorrect struct {
	} `json:"autoCorrect"`
	CategoryBreadcrumbs []struct {
		URL          string `json:"url"`
		Name         string `json:"name"`
		CategoryCode string `json:"categoryCode"`
	} `json:"categoryBreadcrumbs"`
	SeoMetaData []struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	} `json:"seoMetaData"`
	SearchAllStores bool `json:"searchAllStores"`
	IsRelaxed       bool `json:"isRelaxed"`
}

type TwStore struct {
	Pagination struct {
		CurrentPage  int `json:"currentPage"`
		PageSize     int `json:"pageSize"`
		TotalPages   int `json:"totalPages"`
		TotalResults int `json:"totalResults"`
	} `json:"pagination"`
	Metadata struct {
		Geolocation struct {
			Latitude     float64 `json:"latitude"`
			Longitude    float64 `json:"longitude"`
			State        string  `json:"state"`
			StateIsoCode string  `json:"stateIsoCode"`
		} `json:"geolocation"`
		States []struct {
			StateIsoCode string `json:"stateIsoCode"`
			State        string `json:"state"`
			Count        int    `json:"count"`
			Selected     bool   `json:"selected,omitempty"`
		} `json:"states"`
	} `json:"metadata"`
	Stores []struct {
		Address1         string `json:"address1"`
		Address2         string `json:"address2"`
		BeerTastingHours struct {
			HasHours  bool `json:"hasHours"`
			ShowHours bool `json:"showHours"`
		} `json:"beerTastingHours"`
		City                       string `json:"city"`
		DisplayMessage             bool   `json:"displayMessage"`
		DisplaySpecialInstructions bool   `json:"displaySpecialInstructions"`
		SpecialInstructionsHeader  string `json:"specialInstructionsHeader"`
		SpecialInstructionsText    string `json:"specialInstructionsText"`
		GenericHeader              struct {
		} `json:"genericHeader"`
		DisplayWeeklyAd               bool    `json:"displayWeeklyAd"`
		WeeklyAdLink                  string  `json:"weeklyAdLink"`
		Distance                      float64 `json:"distance"`
		FormattedDistance             string  `json:"formattedDistance"`
		Latitude                      float64 `json:"latitude"`
		Longitude                     float64 `json:"longitude"`
		Growler                       bool    `json:"growler"`
		Humidor                       bool    `json:"humidor"`
		Classroom                     bool    `json:"classroom"`
		MarketingStatus               string  `json:"marketingStatus"`
		MapImage                      string  `json:"mapImage"`
		MessageHeader                 string  `json:"messageHeader"`
		Message                       string  `json:"message"`
		Name                          string  `json:"name"`
		Phone                         string  `json:"phone"`
		PhoneFormatted                string  `json:"phoneFormatted"`
		RegulatoryStore               bool    `json:"regulatoryStore"`
		CustomerServicePhone          string  `json:"customerServicePhone"`
		CustomerServicePhoneFormatted string  `json:"customerServicePhoneFormatted"`
		SpiritsHours                  struct {
			HasHours  bool `json:"hasHours"`
			ShowHours bool `json:"showHours"`
		} `json:"spiritsHours"`
		SpiritsTastingHours struct {
			HasHours  bool `json:"hasHours"`
			ShowHours bool `json:"showHours"`
		} `json:"spiritsTastingHours"`
		State        string `json:"state"`
		StateShort   string `json:"stateShort"`
		StateIsoCode string `json:"stateIsoCode"`
		StoreHours   struct {
			HasHours            bool   `json:"hasHours"`
			ShowHours           bool   `json:"showHours"`
			SpecialHoursMessage string `json:"specialHoursMessage"`
			IsSpecialHours      bool   `json:"isSpecialHours"`
			Days                []struct {
				ClosedStatus           bool   `json:"closedStatus"`
				ClosingTime            string `json:"closingTime"`
				DayOfWeek              string `json:"dayOfWeek"`
				OpeningTime            string `json:"openingTime"`
				AlternateOperationTime bool   `json:"alternateOperationTime,omitempty"`
			} `json:"days"`
		} `json:"storeHours"`
		NextWeekStoreHours struct {
			HasHours  bool `json:"hasHours"`
			ShowHours bool `json:"showHours"`
			Days      []struct {
				ClosedStatus bool   `json:"closedStatus"`
				ClosingTime  string `json:"closingTime"`
				DayOfWeek    string `json:"dayOfWeek"`
				OpeningTime  string `json:"openingTime"`
			} `json:"days"`
		} `json:"nextWeekStoreHours"`
		StoreImages []struct {
			AltText   string `json:"altText"`
			ImageType string `json:"imageType"`
			Format    string `json:"format"`
			URL       string `json:"url"`
		} `json:"storeImages"`
		StoreHeaderImage struct {
		} `json:"storeHeaderImage"`
		ThreeIcons []struct {
			Display     bool   `json:"display"`
			HeaderText  string `json:"headerText"`
			Description string `json:"description"`
			IconName    string `json:"iconName"`
		} `json:"threeIcons,omitempty"`
		ContentRightImage []struct {
			HeaderText  string `json:"headerText"`
			Description string `json:"description"`
			LinkText    string `json:"linkText"`
			LinkURL     string `json:"linkUrl"`
			Image       struct {
				AltText   string `json:"altText"`
				ImageType string `json:"imageType"`
				Format    string `json:"format"`
				URL       string `json:"url"`
			} `json:"image,omitempty"`
			Display       bool   `json:"display,omitempty"`
			SubHeaderText string `json:"subHeaderText,omitempty"`

		} `json:"contentRightImage,omitempty"`
		StoreNumber      string `json:"storeNumber"`
		WineTastingHours struct {
			HasHours  bool `json:"hasHours"`
			ShowHours bool `json:"showHours"`
			Days      []struct {
				ClosedStatus bool   `json:"closedStatus"`
				ClosingTime  string `json:"closingTime"`
				DayOfWeek    string `json:"dayOfWeek"`
				OpeningTime  string `json:"openingTime"`
			} `json:"days"`
		} `json:"wineTastingHours"`
		WifiAvailable            bool   `json:"wifiAvailable"`
		Zip                      string `json:"zip"`
		SpiritsProhibited        bool   `json:"spiritsProhibited"`
		HideTotalDiscovery       bool   `json:"hideTotalDiscovery"`
		DeliveryEligible         bool   `json:"deliveryEligible"`
		DeliveryTipEligible      bool   `json:"deliveryTipEligible"`
		TimeZone                 string `json:"timeZone"`
		LoyaltyProgram           string `json:"loyaltyProgram"`
		FulfillmentDelayShipping bool   `json:"fulfillmentDelayShipping"`
		FulfillmentDelayISP      bool   `json:"fulfillmentDelayISP"`
		FutureDeliveryAllowed    bool   `json:"futureDeliveryAllowed"`
		CurbsideAvailable        bool   `json:"curbsideAvailable"`
		ServiceTypes             []struct {
			Code              string `json:"code"`
			OrderReadyByHours int    `json:"orderReadyByHours"`
		} `json:"serviceTypes"`
		SocialMedia []struct {
			Display         bool   `json:"display"`
			SocialMediaType string `json:"socialMediaType"`
			HeaderLabel     string `json:"headerLabel"`
			URL             string `json:"url"`
			URLLabel        string `json:"urlLabel,omitempty"`
		} `json:"socialMedia,omitempty"`
	} `json:"stores"`
}

func (t *CheckoutTask) TwGenPxCookie() error {
	var proxy string
	if t.Proxy != nil {
		proxy = t.Proxy.String()
	}
	// ckies, err := GetPxCookie(Px3Request{
	ckies, err := t.GetPxCookieApi(Px3Request{
		Proxy: proxy,
		Cookie: t.CookiesStr(),
		Url: t.Url.String(),
		// TODO cookies. can regenning help avoid 8min limit on checkout poll?
		AppId: "PXFF0j69T5",
		Vid: "",
		Uuid: "",
		Host: "",
		JsSrc: "https://www.totalwine.com/FF0j69T5/init.js",
		CapJsSrc: "",
	})
	if err != nil {
		return err
	}

	for _, ckie := range ckies {
		t.SetCookie(ckie)
	}
	return nil
}

func (t *CheckoutTask) TwVisitProductIndex(page int) (error) {
  url_, err := url.Parse(fmt.Sprintf("https://www.totalwine.com/wine/c/c0020?viewall=true&page=%d&pageSize=200", page))
  if err != nil {
    return err
  }
  headerOrder := []string {
    "upgrade-insecure-requests",
    "user-agent",
    "accept",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-user",
    "sec-fetch-dest",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"upgrade-insecure-requests", "1"},
    {"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
    {"sec-fetch-site", "none"},
    {"sec-fetch-mode", "navigate"},
    {"sec-fetch-user", "?1"},
    {"sec-fetch-dest", "document"},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
  DiscardResp(resp)
  return err
}



func (t *CheckoutTask) TwStockCheck(state string, storeId string, page int) (*TwProductApiResp, error) {
  url_, err := url.Parse(
  	fmt.Sprintf("https://www.totalwine.com/search/api/product/categories/v2/categories/c0020/products?page=%d&pageSize=200&state=US-%s&shoppingMethod=INSTORE_PICKUP&userShoppingMethod=INSTORE_PICKUP&allStoresCount=true&storeId=%s&ts=%d",
  	page,
  	state,
  	storeId,
  	timeMillis(),
  ))
  if err != nil {
    return nil, err
  }
  headerOrder := []string {
    "accept",
    "user-agent",
    "sec-fetch-site",
    "sec-fetch-mode",
    "sec-fetch-dest",
    "referer",
    "accept-encoding",
    "accept-language",
    "cookie",
  }
  headers := [][2]string {
    {"accept", "application/json, text/plain, */*"},
    {"sec-fetch-site", "same-origin"},
    {"sec-fetch-mode", "cors"},
    {"sec-fetch-dest", "empty"},
    {"referer", fmt.Sprintf("https://www.totalwine.com/wine/c/c0020?viewall=true&page=%d&pageSize=200", page)},
    {"accept-encoding", "gzip, deflate, br"},
    {"accept-language", "en-US,en;q=0.9"},
  }
  resp, err := t.doReq(t.client, t.makeReq("GET", url_, &headers, &headerOrder, nil))
  if err != nil {
  	return nil, err
  }

  var presp TwProductApiResp
  err = readRespJsonDst(resp, &presp)
  return &presp, err
}

func (t *CheckoutTask) TotalWineScrape() error {
	t.DefaultReqClose = true
	t.UserAgent = DD_UAS[rand.Intn(len(DD_UAS))]

	// t.TwGenPxCookie()


	t.FastlyJig()


	// b, _ := ioutil.ReadAll(req.Request.Body)
	// t.LogDebug("%s", string(b))
	states := []string{
		"AL",
		"AK",
		"AZ",
		"AR",
		"CA",
		"CO",
		"CT",
		"DE",
		"FL",
		"GA",
		"HI",
		"ID",
		"IL",
		"IN",
		"IA",
		"KS",
		"KY",
		"LA",
		"ME",
		"MD",
		"MA",
		"MI",
		"MN",
		"MS",
		"MO",
		"MT",
		"NE",
		"NV",
		"NH",
		"NJ",
		"NM",
		"NY",
		"NC",
		"ND",
		"OH",
		"OK",
		"OR",
		"PA",
		"RI",
		"SC",
		"SD",
		"TN",
		"TX",
		"UT",
		"VT",
		"VA",
		"WA",
		"WV",
		"WI",
		"WY",
	}
	storeIds := []string{"1101"}
	stateByStoreId := map[string]string{}
	for _, state := range states {
		log.Println(state)
		req, err := t.MakeChlsRequest(TW_STORES_CHLSJ, nil)
		t.LogDebug("%+v %+v", req, err)
		req.Request.URL.Path = strings.Replace(req.Request.URL.Path, "CA", state, -1)
		req.Request.Close = true
		resp, err := t.doReq(t.client, req.Request)
		if err != nil {
			return err
		}
		var store TwStore
		if err := readRespJsonDst(resp, &store); err == nil {
			for _, s := range store.Stores {
				log.Println(s.StoreNumber)
				storeIds = append(storeIds, s.StoreNumber)
				stateByStoreId[s.StoreNumber] = state
			}
		}
	}
	log.Println()
	log.Printf("%v", storeIds)
	for i := 0; i < len(storeIds); i++ {
		state := stateByStoreId[storeIds[i]]
		var page int = 1
		var maxPage int = 100
		log.Printf("storeId %s", storeIds[i])
		twProducts := []TwProduct{}
		for page < maxPage {
			log.Printf("page %d", page)
			// t.TwVisitProductIndex(page)
			// TODO only gen every so often
			if (page - 1) % 4 == 0 && !t.UseIpv6 {
				t.TwGenPxCookie()
			}
			stock, err := t.TwStockCheck(state, storeIds[i], page)
			if stock != nil {
				if stock.Pagination.TotalPages > 0 {
					maxPage = stock.Pagination.TotalPages
				}
				twProducts = append(twProducts, stock.Products...)
				log.Printf("npage=%d n=%+v %+v", stock.Pagination.TotalPages,  len(stock.Products), err)
				page += 1
			} else {
				t.FastlyJig()
			}
			// time.Sleep(100*time.Milli)
		}
		twpb, _ := json.Marshal(twProducts)
		fn := fmt.Sprintf("tw-%s-%d.json", storeIds[i], timeMillis())
		log.Printf("wrote %s", fn)
		ioutil.WriteFile(fn, twpb, 0644)
	}
	// log.Printf("%s", string(twpb))

	return nil
}
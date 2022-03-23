package task

type WmtProductItem struct {
	IsDigitalVariant            bool   `json:"isDigitalVariant"`
	IsAudioBookVariant          bool   `json:"isAudioBookVariant"`
	ShowBuyNowButton            bool   `json:"showBuyNowButton"`
	ShowFreeTrialButton         bool   `json:"showFreeTrialButton"`
	IsKeySpecFeatureEnabled     bool   `json:"isKeySpecFeatureEnabled"`
	IsReduceOtherSellersEnabled bool   `json:"isReduceOtherSellersEnabled"`
	BrandName                   string `json:"brandName"`
	OtherInfoLabel              string `json:"otherInfoLabel"`
	OtherInfoValue              string `json:"otherInfoValue"`
	AnalyticsType               string `json:"analyticsType"`
	ProductID                   string `json:"productId"`
	UsItemID                    string `json:"usItemId"`
	Fetched                     bool   `json:"fetched"`
	Images                      []struct {
		ID       string `json:"id"`
		URL      string `json:"url"`
		Zoomable bool   `json:"zoomable"`
	} `json:"images"`
	Variants []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"variants"`
	ProductName                  string      `json:"productName"`
	BvShellProductName           string      `json:"bvShellProductName"`
	CategoryPathID               string      `json:"categoryPathId"`
	PersonalizationData          interface{} `json:"personalizationData"`
	InflexibleKitGroupComponents struct {
		GroupComponents []struct {
			Quantity int64    `json:"quantity"`
			OfferID  string `json:"offerId"`
		} `json:"groupComponents"`
	} `json:"inflexibleKitGroupComponents"`
	GroupType string `json:"groupType"`
	Path      []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"path"`
	CategoryPath                 string        `json:"categoryPath"`
	RhPath                       string        `json:"rhPath"`
	PrimaryShelfID               string        `json:"primaryShelfId"`
	ShortDescription             string        `json:"shortDescription"`
	DetailedDescription          string        `json:"detailedDescription"`
	ProductHighlightedAttributes []interface{} `json:"productHighlightedAttributes"`
	SalesRank                    struct {
	} `json:"salesRank"`
	ClassID          string        `json:"classId"`
	IronbankCategory string        `json:"ironbankCategory"`
	OfferID          string        `json:"offerId"`
	OfferType        string        `json:"offerType"`
	Offers           []interface{} `json:"offers"`
	OfferCount       int64           `json:"offerCount"`
	IsOnline         bool          `json:"isOnline"`
	SellerOfferID    string        `json:"sellerOfferId"`
	ShippingOptions  []struct {
		FulfillmentPrice struct {
			Price              float64 `json:"price"`
			CurrencyUnit       string  `json:"currencyUnit"`
			CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
		} `json:"fulfillmentPrice"`
		FulfillmentDateRange struct {
			ExactDeliveryDate int64 `json:"exactDeliveryDate"`
		} `json:"fulfillmentDateRange"`
		FulfillmentPriceType string `json:"fulfillmentPriceType"`
		ShipMethod           string `json:"shipMethod"`
	} `json:"shippingOptions"`
	HighlightedShippingOption struct {
		Index int64    `json:"index"`
		Title string `json:"title"`
	} `json:"highlightedShippingOption"`
	ShippingOptionIndex        int64 `json:"shippingOptionIndex"`
	FreeShippingThresholdPrice struct {
	} `json:"freeShippingThresholdPrice"`
	Shippable                  bool `json:"shippable"`
	UpsellFulfillmentOption    struct {
		Upsellable    bool `json:"upsellable"`
		ShippingPrice struct {
			Price              float64 `json:"price"`
			CurrencyUnit       string  `json:"currencyUnit"`
			CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
		} `json:"shippingPrice"`
		ShipMethod                           string `json:"shipMethod"`
		ArrivalDate                          int64  `json:"arrivalDate"`
		DisplayArrivalDate                   string `json:"displayArrivalDate"`
		SkyLineAdEnabled                     bool   `json:"skyLineAdEnabled"`
		DisplayUpsellForAllCategoriesEnabled bool   `json:"displayUpsellForAllCategoriesEnabled"`
		HighlightedDeliveryDate              bool   `json:"highlightedDeliveryDate"`
		Title                                string `json:"title"`
	} `json:"upsellFulfillmentOption"`
	ShippingTitleToDisplay         string `json:"shippingTitleToDisplay"`
	IsEDeliveryItem                bool   `json:"isEDeliveryItem"`
	IsBelowShippingThreshold bool    `json:"isBelowShippingThreshold"`
	TwoDayShippingEligible   bool    `json:"twoDayShippingEligible"`
	NextDayEligible          bool    `json:"nextDayEligible"`
	HasFreightShipping       bool    `json:"hasFreightShipping"`
	ShippingPrice            float64 `json:"shippingPrice"`
	SurchargeType            string  `json:"surchargeType"`
	ShipAsIs  bool   `json:"shipAsIs"`
	ClassType string `json:"classType"`
	PriceFlagsList                  []interface{} `json:"priceFlagsList"`
	Pickupable                      bool          `json:"pickupable"`
	PickupOptions                   []interface{} `json:"pickupOptions"`
	PickupDiscountEligible          bool          `json:"pickupDiscountEligible"`
	PickupTodayEligible             bool          `json:"pickupTodayEligible"`
	PickupMethod                    string        `json:"pickupMethod"`
	AvailabilityStatus              string        `json:"availabilityStatus"`
	GlobalProductAvailability       string        `json:"globalProductAvailability"`
	UrgentQuantity                  interface{}   `json:"urgentQuantity"`
	MinQuantity                     int64           `json:"minQuantity"`
	MaxQuantity                     int64           `json:"maxQuantity"`
	StoreIds                        []interface{} `json:"storeIds"`
	SellerID                        string        `json:"sellerId"`
	CatalogSellerID                 string        `json:"catalogSellerId"`
	SellerDisplayName               string        `json:"sellerDisplayName"`
	ShowSoldBy                      bool          `json:"showSoldBy"`
	IsEGiftCard                     bool          `json:"isEGiftCard"`
	IsPhysicalGiftCard              bool          `json:"isPhysicalGiftCard"`
	IsTiresItem                     bool          `json:"isTiresItem"`
	IsFsaEligible                   bool          `json:"isFsaEligible"`
	FlexibleSpendingAccountEligible interface{}   `json:"flexibleSpendingAccountEligible"`
	RemoveATC                       bool          `json:"removeATC"`
	ReviewsCount                    int64           `json:"reviewsCount"`
	AverageRating                   float64       `json:"averageRating"`
	PremiumDeliveryOptions          []interface{} `json:"premiumDeliveryOptions"`
	BrandURL                        string        `json:"brandUrl"`
	ProductTypeID                   string        `json:"productTypeId"`
	Manufacturer                    string        `json:"manufacturer"`
	CanonicalURL                    string        `json:"canonicalUrl"`
	SkuID                           string        `json:"skuId"`
	BlitzDayStartTime               interface{}   `json:"blitzDayStartTime"`
	BlitzItem                       bool          `json:"blitzItem"`
	BlitzStoreMsg                   interface{}   `json:"blitzStoreMsg"`
	BlitzDayStartMsg                interface{}   `json:"blitzDayStartMsg"`
	ProductType                     string        `json:"productType"`
	ProductDietaryAttributes        []interface{} `json:"productDietaryAttributes"`
	SpecificationHighlights         []interface{} `json:"specificationHighlights"`
	HomeServiceType                 string        `json:"homeServiceType"`
	IsWirelessItem                  bool          `json:"isWirelessItem"`
	IsWirelessPrepaid               bool          `json:"isWirelessPrepaid"`
	Subscription                    struct {
	} `json:"subscription"`
	SubscriptionIntervalFrequency string `json:"subscriptionIntervalFrequency"`
	BundleGroupID                 string `json:"bundleGroupId"`
}
type WalmartMobileProduct struct {
	Errors []struct {
		ErrorType        string `json:"errorType,omitempty"`
		Message          string `json:"message,omitempty"`
		Description      string `json:"description,omitempty"`
		Info             string `json:"info,omitempty"`
		ErrorIdentifiers []struct {
			ErrorType   string `json:"errorType"`
			Message     string `json:"message"`
			Description string `json:"description"`
			Info        string `json:"info"`
		} `json:"errorIdentifiers,omitempty"`
		Context struct {
			Access  bool   `json:"access"`
			Device  string `json:"device"`
			Browser string `json:"browser"`
			TopTxID string `json:"topTxId"`
		} `json:"context,omitempty"`
		Timing struct {
			Total struct {
				Elapsed     int64 `json:"elapsed"`
				Active      int64 `json:"active"`
				Threads     int64 `json:"threads"`
				Futures     int64 `json:"futures"`
				Delay       int64 `json:"delay"`
				Orchestrate int64 `json:"orchestrate"`
			} `json:"total"`
		} `json:"timing,omitempty"`
	} `json:"errors"`
	Data struct {
		ProductByProductID struct {
			OfferCount             int64    `json:"offerCount"`
			TransactableOfferCount int64    `json:"transactableOfferCount"`
			BundleType             string `json:"bundleType"`
			SellerList             []struct {
				CatalogSellerID   string `json:"catalogSellerId"`
				SellerName        string `json:"sellerName"`
				SellerDisplayName string `json:"sellerDisplayName"`
				SellerID          string `json:"sellerId"`
				SellerType        string `json:"sellerType"`
			} `json:"sellerList"`
			RegularItem       bool     `json:"regularItem"`
			CollectionID      string   `json:"collectionId"`
			ProductType       string   `json:"productType"`
			Offers            []string `json:"offers"`
			ProductID         string   `json:"productId"`
			ProductAttributes struct {
				ManufacturerName string `json:"manufacturerName"`
				NumberOfReviews  int64    `json:"numberOfReviews"`
				IronbankCategory string `json:"ironbankCategory"`
				ProductName      string `json:"productName"`
				ProductCategory  struct {
					Path []struct {
						Name string `json:"name"`
						URL  string `json:"url"`
					} `json:"path"`
					CategoryPathID string `json:"categoryPathId"`
					CategoryPath   string `json:"categoryPath"`
				} `json:"productCategory"`
				ClassID               string `json:"classId"`
				WalmartEGiftCard      bool   `json:"walmartEGiftCard"`
				ProductClassification struct {
				} `json:"productClassification"`
				ManufacturerProductID string `json:"manufacturerProductId"`
				VerticalInfo          struct {
					VerticalID    string `json:"verticalId"`
					VerticalTheme string `json:"verticalTheme"`
					WalledGarden  bool   `json:"walledGarden"`
				} `json:"verticalInfo"`
				AverageRating                float64 `json:"averageRating"`
				ProductHighlightedAttributes []struct {
					Value string `json:"value"`
					Key   string `json:"key"`
				} `json:"productHighlightedAttributes"`
				KarfIsAlcohol   string `json:"karfIsAlcohol"`
				RhPath          string `json:"rhPath"`
				Brand           string `json:"brand"`
				LegalAttributes struct {
				} `json:"legalAttributes"`
				CanonicalURL        string `json:"canonicalUrl"`
				ShortDescription    string `json:"shortDescription"`
				DetailedDescription string `json:"detailedDescription"`
				MovieAttributes     struct {
					NumberOfCustRatings string `json:"numberOfCustRatings"`
				} `json:"movieAttributes"`
				WalmartItemNumber  string `json:"walmartItemNumber"`
				WalmartGiftCard    bool   `json:"walmartGiftCard"`
				PrimaryShelfID     string `json:"primaryShelfId"`
				SpecialityGiftCard bool   `json:"specialityGiftCard"`
				ClassType          string `json:"classType"`
			} `json:"productAttributes"`
			Upc              string `json:"upc"`
			PrimaryProductID string `json:"primaryProductId"`
			ProductSegment   string `json:"productSegment"`
			ProductTypeID    string `json:"productTypeId"`
			BtvAvailable     bool   `json:"btvAvailable"`
			UsItemID         string `json:"usItemId"`
			Consumable       bool   `json:"consumable"`
			OfferList        []struct {
				ProductAvailability struct {
					AvailabilityStatus string `json:"availabilityStatus"`
				} `json:"productAvailability"`
				OfferInfo struct {
					OfferType       string `json:"offerType"`
					QuantityOptions struct {
						OrderLimit    int64 `json:"orderLimit"`
						OrderMinLimit int64 `json:"orderMinLimit"`
					} `json:"quantityOptions"`
					PreorderInfo struct {
					} `json:"preorderInfo"`
					OfferID               string `json:"offerId"`
					OfferMarketAttributes struct {
					} `json:"offerMarketAttributes"`
					ShowSoldBy bool `json:"showSoldBy"`
				} `json:"offerInfo"`
				PricesInfo struct {
					Prices struct {
						Current struct {
							Price              float64 `json:"price"`
							PriceType          string  `json:"priceType"`
							CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
							CurrencyUnit       string  `json:"currencyUnit"`
						} `json:"current"`
						Unit struct {
							UnitOfMeasure      string  `json:"unitOfMeasure"`
							Price              float64 `json:"price"`
							PriceType          string  `json:"priceType"`
							CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
							CurrencyUnit       string  `json:"currencyUnit"`
						} `json:"unit"`
					} `json:"prices"`
					PriceDisplayCodes struct {
						Rollback                     bool   `json:"rollback"`
						HidePriceForSOI              bool   `json:"hidePriceForSOI"`
						ReducedPrice                 bool   `json:"reducedPrice"`
						UnitPriceDisplayCondition    string `json:"unitPriceDisplayCondition"`
						EligibleForAssociateDiscount bool   `json:"eligibleForAssociateDiscount"`
						Clearance                    bool   `json:"clearance"`
						Strikethrough                bool   `json:"strikethrough"`
					} `json:"priceDisplayCodes"`
				} `json:"pricesInfo"`
				GiftingOptions struct {
					GiftOverboxEligible   bool   `json:"giftOverboxEligible"`
					PriceSuppressionInd   bool   `json:"priceSuppressionInd"`
					GiftingEligibilityInd bool   `json:"giftingEligibilityInd"`
					ReasonCode            string `json:"reasonCode"`
					GiftMessageEligible   bool   `json:"giftMessageEligible"`
					GiftReceiptEligible   bool   `json:"giftReceiptEligible"`
				} `json:"giftingOptions"`
				SellerID                string `json:"sellerId"`
				TwoDayShippingEligible  bool   `json:"twoDayShippingEligible"`
				NextDayEligible         bool   `json:"nextDayEligible"`
				ID                      string `json:"id"`
				GeoItemClassification   string `json:"geoItemClassification"`
				UpsellFulfillmentOption struct {
					DisplayUpsellForAllCategoriesEnabled bool `json:"displayUpsellForAllCategoriesEnabled"`
					ShippingPrice                        struct {
						Price              float64 `json:"price"`
						CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
						CurrencyUnit       string  `json:"currencyUnit"`
					} `json:"shippingPrice"`
					HighlightedDeliveryDate bool   `json:"highlightedDeliveryDate"`
					SkyLineAdEnabled        bool   `json:"skyLineAdEnabled"`
					ShipMethod              string `json:"shipMethod"`
					DisplayArrivalDate      string `json:"displayArrivalDate"`
					TitlePrice              struct {
						Price              float64 `json:"price"`
						CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
						CurrencyUnit       string  `json:"currencyUnit"`
					} `json:"titlePrice"`
					Title       string `json:"title"`
					ArrivalDate string `json:"arrivalDate"`
					Upsellable  bool   `json:"upsellable"`
				} `json:"upsellFulfillmentOption"`
				OfferFlags struct {
					S2SOption            bool `json:"s2SOption"`
					PickupTodayAvailable bool `json:"pickupTodayAvailable"`
					S2HOption            bool `json:"s2HOption"`
				} `json:"offerFlags"`
				Consumable  bool `json:"consumable"`
				Fulfillment struct {
					HasShippingRestrictions   bool `json:"hasShippingRestrictions"`
					HighlightedShippingOption struct {
						Index      int64 `json:"index"`
						TitlePrice struct {
							Price              float64 `json:"price"`
							CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
							CurrencyUnit       string  `json:"currencyUnit"`
						} `json:"titlePrice"`
						Title string `json:"title"`
					} `json:"highlightedShippingOption"`
					UpsellShippingOptionIndex int64  `json:"upsellShippingOptionIndex"`
					Shippable                 bool `json:"shippable"`
					NextDayShippingOption     struct {
						Index      int64 `json:"index"`
						TitlePrice struct {
							Price              float64 `json:"price"`
							CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
							CurrencyUnit       string  `json:"currencyUnit"`
						} `json:"titlePrice"`
						Title string `json:"title"`
					} `json:"nextDayShippingOption"`
					FreeShippingThresholdPrice struct {
						Price              float64 `json:"price"`
						CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
						CurrencyUnit       string  `json:"currencyUnit"`
					} `json:"freeShippingThresholdPrice"`
					Pickupable        bool `json:"pickupable"`
					LocationSurcharge bool `json:"locationSurcharge"`
					PickupOptions     []struct {
						Distance        int64    `json:"distance"`
						PickupMethod    string `json:"pickupMethod"`
						Availability    string `json:"availability"`
						ProductLocation struct {
							Section int64    `json:"section"`
							Aisle   int64    `json:"aisle"`
							Zone    string `json:"zone"`
						} `json:"productLocation"`
						StoreID             int64    `json:"storeId"`
						InStoreStockStatus  string `json:"inStoreStockStatus"`
						InStorePackagePrice struct {
							Price              float64 `json:"price"`
							CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
							CurrencyUnit       string  `json:"currencyUnit"`
						} `json:"inStorePackagePrice"`
						StoreAddress            string `json:"storeAddress"`
						StorePhone              string `json:"storePhone"`
						TireInstallationService struct {
							AccOnlineScheduling bool   `json:"accOnlineScheduling"`
							Phone               string `json:"phone"`
						} `json:"tireInstallationService,omitempty"`
						StoreTimeZone    string `json:"storeTimeZone"`
						StoreName        string `json:"storeName"`
						StoreCity        string `json:"storeCity"`
						FulfillmentPrice struct {
							Price              float64 `json:"price"`
							CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
							CurrencyUnit       string  `json:"currencyUnit"`
						} `json:"fulfillmentPrice"`
						StorePostalCode          string `json:"storePostalCode"`
						StoreStateOrProvinceCode string `json:"storeStateOrProvinceCode"`
					} `json:"pickupOptions"`
					ShippingOptions []struct {
						FulfillmentDateRange struct {
							ExactDeliveryDate string `json:"exactDeliveryDate"`
						} `json:"fulfillmentDateRange"`
						FulfillmentPriceType string `json:"fulfillmentPriceType"`
						ShipMethod           string `json:"shipMethod"`
						FulfillmentPrice     struct {
							Price              float64 `json:"price"`
							CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
							CurrencyUnit       string  `json:"currencyUnit"`
						} `json:"fulfillmentPrice"`
					} `json:"shippingOptions"`
				} `json:"fulfillment"`
				PersonalizationData struct {
				} `json:"personalizationData"`
				ShipAsIs bool   `json:"shipAsIs"`
				Status   string `json:"status"`
			} `json:"offerList"`
			ImageList []struct {
				Rank               int64    `json:"rank"`
				Type               string `json:"type"`
				ImageAssetSizeUrls struct {
					Default string `json:"default"`
				} `json:"imageAssetSizeUrls"`
				AssetID string `json:"assetId"`
			} `json:"imageList"`
			Status string `json:"status"`
		} `json:"productByProductId"`
	} `json:"data"`
}

type WmtProduct struct {
	Item struct {
		Ads struct {
			Config struct {
				LazyHomepageExpose1         string `json:"lazy-homepage-expose1"`
				LazySearchExpose1           string `json:"lazy-search-expose1"`
				LazyBrowseExpose1           string `json:"lazy-browse-expose1"`
				LazyCategoryExpose1         string `json:"lazy-category-expose1"`
				NoCategoryMarquee2          bool   `json:"no-category-marquee2"`
				NoDealsSkyline1             bool   `json:"no-deals-skyline1"`
				NoHomepageTwocolumnhp       bool   `json:"no-homepage-twocolumnhp"`
				LazyItemExpose1             string `json:"lazy-item-expose1"`
				LazyItemMarquee2            string `json:"lazy-item-marquee2"`
				LazyItemRightrail2          string `json:"lazy-item-rightrail2"`
				DisplayAds                  string `json:"displayAds"`
				AdblockImgSource            string `json:"adblockImgSource"`
				IsTwoDayDeliveryTextEnabled bool   `json:"isTwoDayDeliveryTextEnabled"`
				Ads2S                       bool   `json:"ads2s"`
				Bypassproxy                 bool   `json:"bypassproxy"`
				AdblockDetectionEnabled     bool   `json:"adblockDetectionEnabled"`
				RemoteAddress               string `json:"remoteAddress"`
				Cloud                       string `json:"cloud"`
			} `json:"config"`
			Display struct {
			} `json:"display"`
			NetworkCode int64 `json:"networkCode"`
			Wpa         struct {
			} `json:"wpa"`
			Version   string `json:"version"`
			Crawler   bool   `json:"crawler"`
			Bot       bool   `json:"bot"`
			IvtScore  string `json:"ivtScore"`
			IvtTorbot string `json:"ivtTorbot"`
			Mapping   struct {
			} `json:"mapping"`
		} `json:"ads"`
		UUID                   interface{} `json:"uuid"`
		Host                   string      `json:"host"`
		IsMobile               bool        `json:"isMobile"`
		IsPacLoaded            bool        `json:"isPacLoaded"`
		PacRedirectURL         string      `json:"pacRedirectUrl"`
		IsBot                  bool        `json:"isBot"`
		IsEsiEnabled           bool        `json:"isEsiEnabled"`
		IsInitialStateDeferred bool        `json:"isInitialStateDeferred"`
		IsServiceWorkerEnabled bool        `json:"isServiceWorkerEnabled"`
		IsShellRequest         bool        `json:"isShellRequest"`
		ProductID              string      `json:"productId"`
		FetchConfig            struct {
			TerraUsePOST      bool   `json:"terraUsePOST"`
			TerraFetchURL     string `json:"terraFetchUrl"`
			TerraAjaxFetchURL string `json:"terraAjaxFetchUrl"`
			UseLocation       bool   `json:"useLocation"`
			TerraConsumerID   string `json:"terraConsumerId"`
		} `json:"fetchConfig"`
		Ccm struct {
		} `json:"ccm"`
		DeviceType string      `json:"deviceType"`
		Referer    interface{} `json:"referer"`
		VerticalID string      `json:"verticalId"`
		SiteMode   int64         `json:"siteMode"`
		Product    struct {
			Ads struct {
				Config struct {
					LazyHomepageExpose1         string `json:"lazy-homepage-expose1"`
					LazySearchExpose1           string `json:"lazy-search-expose1"`
					LazyBrowseExpose1           string `json:"lazy-browse-expose1"`
					LazyCategoryExpose1         string `json:"lazy-category-expose1"`
					NoCategoryMarquee2          bool   `json:"no-category-marquee2"`
					NoDealsSkyline1             bool   `json:"no-deals-skyline1"`
					NoHomepageTwocolumnhp       bool   `json:"no-homepage-twocolumnhp"`
					LazyItemExpose1             string `json:"lazy-item-expose1"`
					LazyItemMarquee2            string `json:"lazy-item-marquee2"`
					LazyItemRightrail2          string `json:"lazy-item-rightrail2"`
					DisplayAds                  string `json:"displayAds"`
					AdblockImgSource            string `json:"adblockImgSource"`
					IsTwoDayDeliveryTextEnabled bool   `json:"isTwoDayDeliveryTextEnabled"`
					Ads2S                       bool   `json:"ads2s"`
					Bypassproxy                 bool   `json:"bypassproxy"`
					AdblockDetectionEnabled     bool   `json:"adblockDetectionEnabled"`
					RemoteAddress               string `json:"remoteAddress"`
					Cloud                       string `json:"cloud"`
				} `json:"config"`
				Display struct {
				} `json:"display"`
				NetworkCode int64 `json:"networkCode"`
				Wpa         struct {
				} `json:"wpa"`
				Version   string `json:"version"`
				Crawler   bool   `json:"crawler"`
				Bot       bool   `json:"bot"`
				IvtScore  string `json:"ivtScore"`
				IvtTorbot string `json:"ivtTorbot"`
				Mapping   struct {
				} `json:"mapping"`
			} `json:"ads"`
			AddOnServices struct {
			} `json:"addOnServices"`
			CarePlans struct {
			} `json:"carePlans"`
			HomeServices struct {
			} `json:"homeServices"`
			CellCoverageFinder struct {
				Status       string `json:"status"`
				ZipCode      string `json:"zipCode"`
				Availability string `json:"availability"`
				Heading      string `json:"heading"`
			} `json:"cellCoverageFinder"`
			CollectionID    string `json:"collectionId"`
			CompleteTheLook struct {
			} `json:"completeTheLook"`
			FashionCategoryNavBar struct {
			} `json:"fashionCategoryNavBar"`
			IdmlMap struct {
				OneGL4BHI7OEX9 struct {
					Modules struct {
					} `json:"modules"`
				} `json:"1GL4BHI7OEX9"`
			} `json:"idmlMap"`
			IdmlModal struct {
				Show bool `json:"show"`
			} `json:"idmlModal"`
			ItemBadge struct {
			} `json:"itemBadge"`
			ItemPOVS struct {
			} `json:"itemPOVS"`
			LongLeadTime bool `json:"longLeadTime"`
			MidasContext struct {
				PageType                    string  `json:"pageType"`
				SubType                     string  `json:"subType"`
				ProductID                   string  `json:"productId"`
				ItemID                      string  `json:"itemId"`
				Price                       float64 `json:"price"`
				Preorder                    bool    `json:"preorder"`
				Online                      bool    `json:"online"`
				FreeShipping                bool    `json:"freeShipping"`
				InStore                     bool    `json:"inStore"`
				Query                       string  `json:"query"`
				Brand                       string  `json:"brand"`
				CategoryPathID              string  `json:"categoryPathId"`
				CategoryPathName            string  `json:"categoryPathName"`
				Manufacturer                string  `json:"manufacturer"`
				VerticalID                  string  `json:"verticalId"`
				VerticalTheme               string  `json:"verticalTheme"`
				Smode                       int64     `json:"smode"`
				IsTwoDayDeliveryTextEnabled bool    `json:"isTwoDayDeliveryTextEnabled"`
				Zip                         string  `json:"zip"`
			} `json:"midasContext"`
			OosView            bool   `json:"oosView"`
			PreferredStoreID   string `json:"preferredStoreId"`
			PremiumBrandBanner struct {
			} `json:"premiumBrandBanner"`
			PromotionalMessage struct {
			} `json:"promotionalMessage"`
			QuestionAnswers struct {
				QuestionDetails []struct {
					QuestionID        string `json:"questionId"`
					QuestionSummary   string `json:"questionSummary"`
					UserNickname      string `json:"userNickname"`
					PositiveVoteCount int64    `json:"positiveVoteCount"`
					NegativeVoteCount int64    `json:"negativeVoteCount"`
					TotalAnswersCount int64    `json:"totalAnswersCount"`
					SubmissionDate    string `json:"submissionDate"`
				} `json:"questionDetails"`
				TotalResults int64 `json:"totalResults"`
				Pagination   struct {
					Total int64 `json:"total"`
					Pages []struct {
						Num    int64    `json:"num"`
						Gap    bool   `json:"gap"`
						Active bool   `json:"active"`
						URL    string `json:"url"`
					} `json:"pages"`
					CurrentSpan string `json:"currentSpan"`
				} `json:"pagination"`
			} `json:"questionAnswers"`
			Reviews struct {
				ActiveAspect     interface{} `json:"activeAspect"`
				FrequentMentions struct {
				} `json:"frequentMentions"`
				AverageOverallRating        float64 `json:"averageOverallRating"`
				RoundedAverageOverallRating float64 `json:"roundedAverageOverallRating"`
				OverallRatingRange          int64     `json:"overallRatingRange"`
				TotalReviewCount            int64     `json:"totalReviewCount"`
				RecommendedPercentage       int64     `json:"recommendedPercentage"`
				RatingValueOneCount         int64     `json:"ratingValueOneCount"`
				RatingValueTwoCount         int64     `json:"ratingValueTwoCount"`
				RatingValueThreeCount       int64     `json:"ratingValueThreeCount"`
				RatingValueFourCount        int64     `json:"ratingValueFourCount"`
				RatingValueFiveCount        int64     `json:"ratingValueFiveCount"`
				PercentageOneCount          int64     `json:"percentageOneCount"`
				PercentageTwoCount          int64     `json:"percentageTwoCount"`
				PercentageThreeCount        int64     `json:"percentageThreeCount"`
				PercentageFourCount         int64     `json:"percentageFourCount"`
				PercentageFiveCount         int64     `json:"percentageFiveCount"`
				ActivePage                  int64     `json:"activePage"`
				ActiveSort                  string  `json:"activeSort"`
				Pagination                  struct {
					Total int64 `json:"total"`
					Pages []struct {
						Num    int64    `json:"num"`
						Gap    bool   `json:"gap"`
						Active bool   `json:"active"`
						URL    string `json:"url"`
					} `json:"pages"`
					CurrentSpan string `json:"currentSpan"`
				} `json:"pagination"`
				CustomerReviews []struct {
					ReviewID             string `json:"reviewId"`
					AuthorID             string `json:"authorId"`
					Recommended          bool   `json:"recommended"`
					ShowRecommended      bool   `json:"showRecommended"`
					NegativeFeedback     int64    `json:"negativeFeedback"`
					PositiveFeedback     int64    `json:"positiveFeedback"`
					Rating               int64    `json:"rating"`
					ReviewTitle          string `json:"reviewTitle"`
					ReviewText           string `json:"reviewText"`
					ReviewSubmissionTime string `json:"reviewSubmissionTime"`
					UserNickname         string `json:"userNickname"`
					UserAttributes       struct {
					} `json:"userAttributes"`
					Photos         []interface{} `json:"photos"`
					Videos         []interface{} `json:"videos"`
					ExternalSource string        `json:"externalSource"`
				} `json:"customerReviews"`
				TopPositiveReview struct {
					ReviewID             string `json:"reviewId"`
					AuthorID             string `json:"authorId"`
					Recommended          bool   `json:"recommended"`
					ShowRecommended      bool   `json:"showRecommended"`
					NegativeFeedback     int64    `json:"negativeFeedback"`
					PositiveFeedback     int64    `json:"positiveFeedback"`
					Rating               int64    `json:"rating"`
					ReviewTitle          string `json:"reviewTitle"`
					ReviewText           string `json:"reviewText"`
					ReviewSubmissionTime string `json:"reviewSubmissionTime"`
					UserNickname         string `json:"userNickname"`
					UserAttributes       struct {
					} `json:"userAttributes"`
					Photos         []interface{} `json:"photos"`
					Videos         []interface{} `json:"videos"`
					ExternalSource string        `json:"externalSource"`
				} `json:"topPositiveReview"`
				LookupID string `json:"lookupId"`
				Aspects  []struct {
					ID           int64    `json:"id"`
					Score        int64    `json:"score"`
					Name         string `json:"name"`
					SnippetCount int64    `json:"snippetCount"`
				} `json:"aspects"`
				TotalImageReviewsCount int64 `json:"totalImageReviewsCount"`
				RecommendedCount       int64 `json:"recommendedCount"`
			} `json:"reviews"`
			BuyBox struct {
				PrimaryProductID string `json:"primaryProductId"`
				PrimaryUsItemID  string `json:"primaryUsItemId"`
				Prices           []struct {
					Price              float64 `json:"price"`
					CurrencyUnitSymbol string  `json:"currencyUnitSymbol"`
				} `json:"prices"`
				States []struct {
					Action   string `json:"action"`
					Product  int64    `json:"product"`
					Criteria []struct {
						Selected bool `json:"selected"`
						Values   []struct {
							Next            int64    `json:"next"`
							Selected        bool   `json:"selected"`
							Availability    string `json:"availability"`
							Price           int64    `json:"price"`
							UnitValue       int64    `json:"unitValue"`
							TwoDayShipping  bool   `json:"twoDayShipping"`
							NextDayEligible bool   `json:"nextDayEligible"`
							Product         int64    `json:"product"`
						} `json:"values"`
					} `json:"criteria"`
				} `json:"states"`
				Criteria []struct {
					Name         string `json:"name"`
					ResourceType string `json:"resourceType"`
					Values       []struct {
						Title     string `json:"title"`
						VariantID string `json:"variantId"`
					} `json:"values"`
				} `json:"criteria"`
				Images   []interface{} `json:"images"`
				Products []WmtProductItem `json:"products"`
				VerticalID               string `json:"verticalId"`
				VerticalTheme            string `json:"verticalTheme"`
				WalledGarden             bool   `json:"walledGarden"`
				AthenaCategoryPathID     string `json:"athenaCategoryPathId"`
				AthenaPrimaryShelfID     string `json:"athenaPrimaryShelfId"`
				AthenaOfferType          string `json:"athenaOfferType"`
				AthenaAvailabilityStatus string `json:"athenaAvailabilityStatus"`
				LastSelectedItem         string `json:"lastSelectedItem"`
				Select                   string `json:"select"`
			} `json:"buyBox"`
			AppState struct {
				Action   string `json:"action"`
				Product  int64    `json:"product"`
				Criteria []struct {
					Selected bool `json:"selected"`
					Values   []struct {
						Next            int64    `json:"next"`
						Selected        bool   `json:"selected"`
						Availability    string `json:"availability"`
						Price           int64    `json:"price"`
						UnitValue       int64    `json:"unitValue"`
						TwoDayShipping  bool   `json:"twoDayShipping"`
						NextDayEligible bool   `json:"nextDayEligible"`
						Product         int64    `json:"product"`
					} `json:"values"`
				} `json:"criteria"`
			} `json:"appState"`
			AddToCart struct {
				ActionStatus   string `json:"actionStatus"`
				AddToCartError struct {
				} `json:"addToCartError"`
			} `json:"addToCart"`
			ImagesState           int64           `json:"imagesState"`
			SelectedAddOnServices []interface{} `json:"selectedAddOnServices"`
			CheckoutComments      struct {
				ProductComments struct {
				} `json:"productComments"`
				Loading           bool `json:"loading"`
				FeedbackSubmitted struct {
				} `json:"feedbackSubmitted"`
				ProductID          string `json:"productId"`
				TotalCommentsCount int64    `json:"totalCommentsCount"`
				CxoComments        []struct {
					CommentID             string `json:"commentId"`
					UserNickname          string `json:"userNickname"`
					CommentText           string `json:"commentText"`
					CommentSubmissionTime string `json:"commentSubmissionTime"`
				} `json:"cxoComments"`
			} `json:"checkoutComments"`
		} `json:"product"`
		ShowTrustModal     bool `json:"showTrustModal"`
		FulfillmentOptions struct {
			Type                 string `json:"type"`
			ShowContainer        bool   `json:"showContainer"`
			Status               string `json:"status"`
			DidInvalidate        bool   `json:"didInvalidate"`
			LocationErrorMessage string `json:"locationErrorMessage"`
			Loaded               bool   `json:"loaded"`
		} `json:"fulfillmentOptions"`
		Feedback struct {
			ShowFeedbackContainer bool `json:"showFeedbackContainer"`
		} `json:"feedback"`
		AddToRegistry struct {
		} `json:"addToRegistry"`
		AddToList struct {
		} `json:"addToList"`
		WpaMap struct {
			Loading     bool `json:"loading"`
			MidasConfig struct {
			} `json:"midasConfig"`
			MidasContext struct {
			} `json:"midasContext"`
			Result struct {
				AdModules []struct {
					Title    string `json:"title"`
					Products []struct {
						ID struct {
							UsItemID   string `json:"usItemId"`
							ProductID  string `json:"productId"`
							OfferID    string `json:"offerId"`
							UsSellerID string `json:"usSellerId"`
						} `json:"id"`
						Price struct {
							FromPrice          interface{} `json:"fromPrice"`
							MinPrice           float64     `json:"minPrice"`
							MaxPrice           float64     `json:"maxPrice"`
							CurrentPrice       float64     `json:"currentPrice"`
							ComparisonPrice    interface{} `json:"comparisonPrice"`
							ComparisonMinPrice float64     `json:"comparisonMinPrice"`
							ComparisonMaxPrice float64     `json:"comparisonMaxPrice"`
							IsStrikeThrough    bool        `json:"isStrikeThrough"`
							SubmapType         interface{} `json:"submapType"`
						} `json:"price"`
						Ratings struct {
							Rating       string `json:"rating"`
							TotalRatings string `json:"totalRatings"`
						} `json:"ratings"`
						Flags struct {
							IsClearance    bool `json:"isClearance"`
							IsRollback     bool `json:"isRollback"`
							IsSpecialBuy   bool `json:"isSpecialBuy"`
							IsReducedPrice bool `json:"isReducedPrice"`
						} `json:"flags"`
						ImageSrc                 string `json:"imageSrc"`
						ProductType              string `json:"productType"`
						BundleType               string `json:"bundleType"`
						CanAddToCart             bool   `json:"canAddToCart"`
						ProductURL               string `json:"productUrl"`
						Brand                    string `json:"brand"`
						IsTwoDayShippingEligible bool   `json:"isTwoDayShippingEligible"`
						IsNextDayEligible        bool   `json:"isNextDayEligible"`
						GeoItemClassification    string `json:"geoItemClassification"`
						ProductName              string `json:"productName"`
						Quantity                 int64    `json:"quantity"`
						AvailabilityStatus       string `json:"availabilityStatus"`
						MidasData                struct {
							ImpBeacon        string      `json:"impBeacon"`
							ClickBeacon      string      `json:"clickBeacon"`
							CampaignID       string      `json:"campaignId"`
							AdGroupID        string      `json:"adGroupId"`
							SellerID         interface{} `json:"sellerId"`
							PangaeaSellerID  interface{} `json:"pangaeaSellerId"`
							IsSellerCampaign bool        `json:"isSellerCampaign"`
							SellerName       string      `json:"sellerName"`
							RelRank          string      `json:"relRank"`
							Details          string      `json:"details"`
							AdType           string      `json:"adType"`
							WpaQs            string      `json:"wpaQs"`
						} `json:"midasData"`
					} `json:"products"`
					MidasModuleData struct {
						AdModule    string `json:"adModule"`
						PageBeacon  string `json:"pageBeacon"`
						PageBeacons struct {
						} `json:"pageBeacons"`
						PageID            string      `json:"pageId"`
						PageType          string      `json:"pageType"`
						BucketID          string      `json:"bucketId"`
						Details           string      `json:"details"`
						ModulePosition    interface{} `json:"modulePosition"`
						UUID              string      `json:"uuid"`
						FeaturedImage     interface{} `json:"featuredImage"`
						FeaturedImageName interface{} `json:"featuredImageName"`
						FeaturedURL       interface{} `json:"featuredUrl"`
						FeaturedHeadline  interface{} `json:"featuredHeadline"`
						LogoClickTrackURL interface{} `json:"logoClickTrackUrl"`
						TraceID           interface{} `json:"traceId"`
						UnpublishedItems  interface{} `json:"unpublishedItems"`
						AdExpID           interface{} `json:"adExpId"`
						ModuleInfo        interface{} `json:"moduleInfo"`
					} `json:"midasModuleData"`
				} `json:"adModules"`
				Details string `json:"details"`
			} `json:"result"`
		} `json:"wpaMap"`
		BtvMap struct {
			BtvPlacementBeaconAction struct {
			} `json:"btvPlacementBeaconAction"`
			Terra struct {
				VariantCategoriesMap struct {
				} `json:"variantCategoriesMap"`
				Products struct {
				} `json:"products"`
				Offers struct {
				} `json:"offers"`
				Images struct {
				} `json:"images"`
				Sellers struct {
				} `json:"sellers"`
				BuyTogetherValue struct {
				} `json:"buyTogetherValue"`
			} `json:"terra"`
			Selection interface{} `json:"selection"`
			UI        struct {
				Hidden   bool `json:"hidden"`
				IsChoice bool `json:"isChoice"`
				Anchor   struct {
					SelectedVariants []interface{} `json:"selectedVariants"`
				} `json:"anchor"`
				Accessories []interface{} `json:"accessories"`
				Offers      struct {
					Badges                 []interface{} `json:"badges"`
					AgeRestriction         interface{}   `json:"ageRestriction"`
					LegalRatingType        interface{}   `json:"legalRatingType"`
					LegalRatingValue       interface{}   `json:"legalRatingValue"`
					TotalPrice             int64           `json:"totalPrice"`
					TotalSavings           int64           `json:"totalSavings"`
					SelectedOffers         []interface{} `json:"selectedOffers"`
					HasSavings             bool          `json:"hasSavings"`
					ShippingPassEligible   bool          `json:"shippingPassEligible"`
					TwoDayShippingEligible bool          `json:"twoDayShippingEligible"`
				} `json:"offers"`
				Seller struct {
					CatalogSellerID   int64    `json:"catalogSellerId"`
					SellerDisplayName string `json:"sellerDisplayName"`
					SellerID          int64    `json:"sellerId"`
					SellerType        string `json:"sellerType"`
				} `json:"seller"`
			} `json:"ui"`
		} `json:"btvMap"`
		PostQuestion struct {
		} `json:"postQuestion"`
		LastAction []string `json:"lastAction"`
		NextDay    struct {
			Status           string `json:"status"`
			CutoffTime       int64    `json:"cutoffTime"`
			Eligible         bool   `json:"eligible"`
			TempUnavailable  bool   `json:"tempUnavailable"`
			TargetDate       int64    `json:"targetDate"`
			PageID           string `json:"pageId"`
			CrtID            string `json:"crtId"`
			FulfillmentLabel string `json:"fulfillmentLabel"`
		} `json:"nextDay"`
		EasyReorder struct {
			QuantitiesInCart struct {
			} `json:"quantitiesInCart"`
		} `json:"easyReorder"`
		Location struct {
			Location struct {
				PostalCode   string `json:"postalCode"`
				City         string `json:"city"`
				State        string `json:"state"`
				Country      string `json:"country"`
				Type         string `json:"type"`
				IsZipLocated bool   `json:"isZipLocated"`
				NextDay      struct {
					CutoffTime       int64    `json:"cutoffTime"`
					Eligible         bool   `json:"eligible"`
					TempUnavailable  bool   `json:"tempUnavailable"`
					TargetDate       int64    `json:"targetDate"`
					Status           string `json:"status"`
					PageID           string `json:"pageId"`
					CrtID            string `json:"crtId"`
					FulfillmentLabel string `json:"fulfillmentLabel"`
				} `json:"nextDay"`
				Stores []struct {
					StoreID string        `json:"storeId"`
					Types   []interface{} `json:"types"`
				} `json:"stores"`
			} `json:"location"`
		} `json:"location"`
		IsAjaxCall        bool `json:"isAjaxCall"`
		AccessModeEnabled bool `json:"accessModeEnabled"`
		ProductQuimbyData struct {
			FetchBTV bool `json:"fetchBTV"`
		} `json:"productQuimbyData"`
		CartData struct {
		} `json:"cartData"`
		ProductComparison struct {
			UI struct {
				TooltipMessage interface{} `json:"tooltipMessage"`
			} `json:"ui"`
			Products struct {
				List []interface{} `json:"list"`
				Map  struct {
				} `json:"map"`
			} `json:"products"`
		} `json:"productComparison"`
		RelmData   []struct {
			ID    string `json:"id"`
			Image string `json:"image"`
			Name  string `json:"name"`
			URL   string `json:"url"`
		} `json:"relmData"`
		Reviews struct {
			CustomerPhotos struct {
				ActiveReviewID   interface{} `json:"activeReviewId"`
				ActivePhotoIndex int64         `json:"activePhotoIndex"`
				CustomerReviews  struct {
					Ids      []interface{} `json:"ids"`
					Entities struct {
					} `json:"entities"`
				} `json:"customerReviews"`
				CurrentPageCursor    int64           `json:"currentPageCursor"`
				ModalTemplate        string        `json:"modalTemplate"`
				ModalVisible         bool          `json:"modalVisible"`
				ModalInReview        bool          `json:"modalInReview"`
				TopReviewsWithImages []interface{} `json:"topReviewsWithImages"`
				TotalImageCount      int64           `json:"totalImageCount"`
				UsItemID             string        `json:"usItemId"`
				ProductID            string        `json:"productId"`
				PhotoReviewPopup     struct {
					Count                  int64  `json:"count"`
					ViewFullGalleryClicked bool `json:"viewFullGalleryClicked"`
				} `json:"photoReviewPopup"`
				TotalImageReviewsPaginationCount int64 `json:"totalImageReviewsPaginationCount"`
			} `json:"customerPhotos"`
		} `json:"reviews"`
		ProductBuyBoxAppState struct {
		} `json:"productBuyBoxAppState"`
		ProductBuyBoxLoader struct {
			IsLoading bool `json:"isLoading"`
		} `json:"productBuyBoxLoader"`
		Heart struct {
			Hearts struct {
			} `json:"hearts"`
		} `json:"heart"`
	} `json:"item"`
}



package task

// import "log"
import "time"

type FtlProductPdpResponse struct {
  Skus []FtlSpecificProduct
  Brand      string `json:"brand"`
  Categories []struct {
    Code string `json:"code"`
    Name string `json:"name"`
  } `json:"categories"`
  Description  string    `json:"description"`
  DropShip     bool      `json:"dropShip"`
  FreeShipping bool      `json:"freeShipping"`
  GiftCosts    []float64 `json:"giftCosts"`
  Images       []struct {
    Code       string `json:"code"`
    Variations []struct {
      AltText string `json:"altText"`
      Format  string `json:"format"`
      URL     string `json:"url"`
    } `json:"variations"`
  } `json:"images"`
  IsNewProduct  bool   `json:"isNewProduct"`
  IsSaleProduct bool   `json:"isSaleProduct"`
  ModelNumber   string `json:"modelNumber"`
  Name          string `json:"name"`
  SellableUnits []FtlSpecificProduct `json:"sellableUnits"`
  SizeChartGridMap []struct {
    Label string   `json:"label"`
    Sizes []string `json:"sizes"`
  } `json:"sizeChartGridMap"`
  SizeChartImage    string `json:"sizeChartImage"`
  SizeChartTipTx    string `json:"sizeChartTipTx"`
  VariantAttributes []FtlSelected `json:"variantAttributes"`
}
func (s *FtlProductPdpResponse) ReleaseTime() time.Time {
  if len(s.VariantAttributes) > 0 {
    return s.VariantAttributes[0].ReleaseTime()
  } else {
    return time.Time{}
  }
}
type FtlCart struct {
  Time time.Time `json:"time"`
  AppliedCoupons             []interface{}     `json:"appliedCoupons"`
  AppliedOrderPromotions     []interface{}     `json:"appliedOrderPromotions"`
  AppliedProductPromotions   []interface{}     `json:"appliedProductPromotions"`
  AppliedVouchers            []interface{}     `json:"appliedVouchers"`
  CartMerged                 bool              `json:"cartMerged"`
  Code                       string            `json:"code"`
  DeliveryOrderGroups        []interface{}     `json:"deliveryOrderGroups"`
  Entries                    []FtlEntries         `json:"entries"`
  GfPaymentInfo              []interface{}     `json:"gfPaymentInfo"`
  GiftBoxAdded               bool              `json:"giftBoxAdded"`
  GiftOrder                  bool              `json:"giftOrder"`
  GUID                       string            `json:"guid"`
  IsCartContainGiftCard      bool              `json:"isCartContainGiftCard"`
  IsCarthasOnlyEMailGiftCard bool              `json:"isCarthasOnlyEMailGiftCard"`
  OrderDiscounts             FtlOrderDiscounts    `json:"orderDiscounts"`
  OutOfStockProducts         []interface{}     `json:"outOfStockProducts"`
  PickupOrderGroups          []interface{}     `json:"pickupOrderGroups"`
  PotentialOrderPromotions   []interface{}     `json:"potentialOrderPromotions"`
  PotentialProductPromotions []interface{}     `json:"potentialProductPromotions"`
  ProductDiscounts           FtlCartProductDiscounts  `json:"productDiscounts"`
  PaymentAddress             FtlPaymentAddress     `json:"paymentAddress"`
  SubTotal                   FtlSubTotal          `json:"subTotal"`
  TotalDiscounts             FtlTotalDiscounts    `json:"totalDiscounts"`
  TotalItems                 int               `json:"totalItems"`
  TotalPrice                 FtlTotalPrice        `json:"totalPrice"`
  TotalPriceWithTax          FtlTotalPriceWithTax `json:"totalPriceWithTax"`
  TotalTax                   FtlTotalTax          `json:"totalTax"`
  TotalUnitCount             int               `json:"totalUnitCount"`
  Type                      string            `json:"type"`
}
type FtlCountry struct {
  Isocode string `json:"isocode"`
  Name    string `json:"name"`
}
type FtlRegion struct {
  CountryIso   string `json:"countryIso"`
  Isocode      string `json:"isocode"`
  IsocodeShort string `json:"isocodeShort"`
  Name         string `json:"name"`
}
type FtlPaymentAddress struct {
  BillingAddress       bool    `json:"billingAddress"`
  Country              FtlCountry `json:"country"`
  DefaultAddress       bool    `json:"defaultAddress"`
  Email                string  `json:"email"`
  FirstName            string  `json:"firstName"`
  FormattedAddress     string  `json:"formattedAddress"`
  ID                   string  `json:"id"`
  LastName             string  `json:"lastName"`
  Line1                string  `json:"line1"`
  Phone                string  `json:"phone"`
  PostalCode           string  `json:"postalCode"`
  Region               FtlRegion  `json:"region"`
  SetAsBilling         bool    `json:"setAsBilling"`
  ShippingAddress      bool    `json:"shippingAddress"`
  Town                 string  `json:"town"`
  VisibleInAddressBook bool    `json:"visibleInAddressBook"`
}
type FtlImages struct {
  AltText string `json:"altText"`
  Format  string `json:"format"`
  URL     string `json:"url"`
}
type FtlPriceData struct {
  CurrencyIso            string `json:"currencyIso"`
  FormattedOriginalPrice string `json:"formattedOriginalPrice"`
  FormattedValue         string `json:"formattedValue"`
  OriginalPrice          float64    `json:"originalPrice"`
  PriceType             string `json:"priceType"`
  Value                  float64    `json:"value"`
}
type FtlStock struct {
  StockLevel       int    `json:"stockLevel"`
  StockLevelStatus string `json:"stockLevelStatus"`
}
type FtlVariantOptions struct {
  AgeBucket                 []interface{} `json:"ageBucket"`
  BackOrderable             bool          `json:"backOrderable"`
  Code                      string        `json:"code"`
  Images                    []interface{} `json:"images"`
  MapEnable                 bool          `json:"mapEnable"`
  MobileBarCode             string        `json:"mobileBarCode"`
  Name                      string        `json:"name"`
  PotentialPromotions       []interface{} `json:"potentialPromotions"`
  PreOrder                  bool          `json:"preOrder"`
  PriceData                 FtlPriceData     `json:"priceData"`
  ShippingRestrictionExists bool          `json:"shippingRestrictionExists"`
  Size                      string        `json:"size"`
  SizeAvailableInStores     bool          `json:"sizeAvailableInStores"`
  Stock                     FtlStock         `json:"stock"`
  VariantOptionQualifiers   []interface{} `json:"variantOptionQualifiers"`
  VariantOptions            []interface{} `json:"variantOptions"`
}
type FtlSelected struct {
  SkuLaunchDate             string           `json:"skuLaunchDate"`
  AgeBucket                 []interface{}    `json:"ageBucket"`
  BackOrderable             bool             `json:"backOrderable"`
  Code                      string           `json:"code"`
  DisplayCountDownTimer     bool             `json:"displayCountDownTimer"`
  Images                    []FtlImages         `json:"images"`
  LaunchProduct             bool             `json:"launchProduct"`
  MapEnable                 bool             `json:"mapEnable"`
  PotentialPromotions       []interface{}    `json:"potentialPromotions"`
  PreOrder                  bool             `json:"preOrder"`
  Name                      string        `json:"name"`
  PriceData                 FtlPriceData        `json:"priceData"`
  RecaptchaOn               bool             `json:"recaptchaOn"`
  ShippingRestrictionExists bool             `json:"shippingRestrictionExists"`
  SizeAvailableInStores     bool             `json:"sizeAvailableInStores"`
  Sku                       string           `json:"sku"`
  Size string `json:"size"`
  Stock                     FtlStock            `json:"stock"`
  Style                     string           `json:"style"`
  VariantOptionQualifiers   []interface{}    `json:"variantOptionQualifiers"`
  VariantOptions            []FtlVariantOptions `json:"variantOptions"`
}
func (s *FtlSelected) ReleaseTime() time.Time {
  // Oct 30 2020 14:00:00 GMT+0000
  // 01/02 03:04:05PM '06 -0700

  if t, err := time.Parse("Jan 02 2006 15:04:05 MST", s.SkuLaunchDate); err == nil {
    return t
  } else {
    // log.Printf("\n\n%+v\n\n", err)
  }
  return time.Time{}
}
type FtlBaseOptions struct {
  Options     []FtlSelected `json:"options"`
  Selected    FtlSelected      `json:"selected,omitempty"`
  VariantType string        `json:"variantType"`
}
type FtlCartPrice struct {
  CurrencyIso    string `json:"currencyIso"`
  FormattedValue string `json:"formattedValue"`
  PriceType     string `json:"priceType"`
  Value          float64    `json:"value"`
}
type FtlCartProduct struct {
  BaseOptions           []FtlBaseOptions `json:"baseOptions"`
  BaseProduct           string        `json:"baseProduct"`
  Categories            []interface{} `json:"categories"`
  Classifications       []interface{} `json:"classifications"`
  Code                  string        `json:"code"`
  DisplayCountDownTimer bool          `json:"displayCountDownTimer"`
  FreeShipping          bool          `json:"freeShipping"`
  FutureStocks          []interface{} `json:"futureStocks"`
  GiftCosts             []interface{} `json:"giftCosts"`
  Images                []interface{} `json:"images"`
  LaunchProduct         bool          `json:"launchProduct"`
  PotentialPromotions   []interface{} `json:"potentialPromotions"`
  Price                 FtlCartPrice         `json:"price"`
  ProductReferences     []interface{} `json:"productReferences"`
  Reviews               []interface{} `json:"reviews"`
  SizeChartGridMap      []interface{} `json:"sizeChartGridMap"`
  SkuExclusions         bool          `json:"skuExclusions"`
  Stock                 FtlStock         `json:"stock"`
  StyleVariantCode      []interface{} `json:"styleVariantCode"`
  VariantMatrix         []interface{} `json:"variantMatrix"`
  VariantOptions        []interface{} `json:"variantOptions"`
  VolumePrices          []interface{} `json:"volumePrices"`
}
type FtlTotalPrice struct {
  CurrencyIso    string `json:"currencyIso"`
  FormattedValue string `json:"formattedValue"`
  Pricetype     string `json:"priceType"`
  Value          float64    `json:"value"`
}
type FtlEntries struct {
  EntryNumber             int           `json:"entryNumber"`
  Product                 FtlCartProduct       `json:"product"`
  ProductPriceVariation   bool          `json:"productPriceVariation"`
  Quantity                int           `json:"quantity"`
  ShippingRestricted      bool          `json:"shippingRestricted"`
  TotalPrice              FtlTotalPrice    `json:"totalPrice"`
}
type FtlOrderDiscounts struct {
  CurrencyIso    string `json:"currencyIso"`
  FormattedValue string `json:"formattedValue"`
  Pricetype string `json:"priceType"`
  value          float64    `json:"value"`
}
type FtlCartProductDiscounts struct {
  CurrencyIso    string `json:"currencyIso"`
  FormattedValue string `json:"formattedValue"`
  Pricetype string `json:"priceType"`
  value          float64    `json:"value"`
}
type FtlSubTotal struct {
  CurrencyIso    string `json:"currencyIso"`
  FormattedValue string `json:"formattedValue"`
  Pricetype string `json:"priceType"`
  value          float64    `json:"value"`
}
type FtlTotalDiscounts struct {
  CurrencyIso    string `json:"currencyIso"`
  FormattedValue string `json:"formattedValue"`
  Pricetype string `json:"priceType"`
  value          float64    `json:"value"`
}
type FtlTotalPriceWithTax struct {
  CurrencyIso    string `json:"currencyIso"`
  FormattedValue string `json:"formattedValue"`
  Pricetype string `json:"priceType"`
  value          float64    `json:"value"`
}
type FtlTotalTax struct {
  CurrencyIso    string `json:"currencyIso"`
  FormattedValue string `json:"formattedValue"`
  Pricetype string `json:"priceType"`
  value          float64    `json:"value"`
}

type FtlRootProduct struct {
  Code                      string     `json:"code"`
  DisplayCountDownTimer     bool       `json:"displayCountDownTimer"`
  FreeShipping              bool       `json:"freeShipping"`
  FreeShippingMessage       string     `json:"freeShippingMessage"`
  IsSelected                bool       `json:"isSelected,omitempty"`
  LaunchProduct             bool       `json:"launchProduct"`
  MapEnable                 bool       `json:"mapEnable"`
  Price                     FtlPrice      `json:"price"`
  RecaptchaOn               bool       `json:"recaptchaOn"`
  Riskified                 bool       `json:"riskified"`
  ShipToAndFromStore        bool       `json:"shipToAndFromStore"`
  ShippingRestrictionExists bool       `json:"shippingRestrictionExists"`
  Sku                       string     `json:"sku"`
  SkuExclusions             bool       `json:"skuExclusions"`
  StockLevelStatus          string     `json:"stockLevelStatus"`
  WebOnlyLaunch             bool       `json:"webOnlyLaunch"`
  Products                  []FtlSpecificProduct `json:"products"`
  Style                     string     `json:"style"`
}

type FtlPrice struct {
  CurrencyIso            string `json:"currencyIso"`
  FormattedOriginalPrice string `json:"formattedOriginalPrice"`
  FormattedValue         string `json:"formattedValue"`
  OriginalPrice          float64    `json:"originalPrice"`
  Value                  float64    `json:"value"`
}

type FtlAttribute struct {
  ID    string `json:"id"`
  Type  string `json:"type"`
  Value string `json:"value"`
}
type FtlSpecificProduct struct {
  Name string
  Attributes            []FtlAttribute `json:"attributes"`
  BarCode               string       `json:"barCode"`
  Code                  string       `json:"code"`
  IsBackOrderable       bool         `json:"isBackOrderable"`
  IsPreOrder            bool         `json:"isPreOrder"`
  IsRecaptchaOn         bool         `json:"isRecaptchaOn"`
  Price                 FtlPrice        `json:"price"`
  SingleStoreInventory  bool         `json:"singleStoreInventory"`
  SizeAvailableInStores bool         `json:"sizeAvailableInStores"`
  StockLevelStatus      string       `json:"stockLevelStatus"`
  Style                 FtlAttribute        `json:"style"`
  Size                  FtlAttribute         `json:"size"`
}
func (p *FtlSpecificProduct) GetAttrValue(type_ string) string {
  var val string
  for _, attr := range p.Attributes {
    if attr.Type == type_ {
      val = attr.Value
      break
    }
  }
  return val
}
type FtlProductDetails struct {
	ByPath map[string][]FtlRootProduct `json:"data"`
}

type FtlProductPageDetails struct {
  Code                           string   `json:"code"`
  CstSkuLaunchDate               string   `json:"cstSkuLaunchDate"`
  DefinedTimeForCountDown        string   `json:"definedTimeForCountDown"`
  DisplayCountDownTimer          bool     `json:"displayCountDownTimer"`
  EligiblePaymentTypesForProduct []string `json:"eligiblePaymentTypesForProduct"`
  FitVariant                     string   `json:"fitVariant"`
  FreeShipping                   bool     `json:"freeShipping"`
  FreeShippingMessage            string   `json:"freeShippingMessage"`
  IsSelected                     bool     `json:"isSelected"`
  LaunchProduct                  bool     `json:"launchProduct"`
  MapEnable                      bool     `json:"mapEnable"`
  PdpActivationDate              string   `json:"pdpActivationDate"`
  Price                          struct {
    CurrencyIso            string `json:"currencyIso"`
    FormattedOriginalPrice string `json:"formattedOriginalPrice"`
    FormattedValue         string `json:"formattedValue"`
  } `json:"price"`
  RecaptchaOn               bool   `json:"recaptchaOn"`
  Riskified                 bool   `json:"riskified"`
  ShipToAndFromStore        bool   `json:"shipToAndFromStore"`
  ShippingRestrictionExists bool   `json:"shippingRestrictionExists"`
  Sku                       string `json:"sku"`
  SkuExclusions             bool   `json:"skuExclusions"`
  SkuLaunchDate             string `json:"skuLaunchDate"`
  StockLevelStatus          string `json:"stockLevelStatus"`
  WebOnlyLaunch             bool   `json:"webOnlyLaunch"`
  WebOnlyMsg                string `json:"webOnlyMsg"`
  Width                     string `json:"width"`
  Products                  []struct {
    Attributes []struct {
      ID    string `json:"id"`
      Type  string `json:"type"`
      Value string `json:"value"`
    } `json:"attributes"`
    BarCode         string `json:"barCode,omitempty"`
    Code            string `json:"code"`
    IsBackOrderable bool   `json:"isBackOrderable"`
    IsPreOrder      bool   `json:"isPreOrder"`
    IsRecaptchaOn   bool   `json:"isRecaptchaOn"`
    Price           struct {
      CurrencyIso            string `json:"currencyIso"`
      FormattedOriginalPrice string `json:"formattedOriginalPrice"`
      FormattedValue         string `json:"formattedValue"`
    } `json:"price"`
    SingleStoreInventory  bool   `json:"singleStoreInventory"`
    SizeAvailableInStores bool   `json:"sizeAvailableInStores"`
    StockLevelStatus      string `json:"stockLevelStatus"`
    Style                 struct {
      ID    string `json:"id"`
      Type  string `json:"type"`
      Value string `json:"value"`
    } `json:"style"`
    Size struct {
      ID    string `json:"id"`
      Type  string `json:"type"`
      Value string `json:"value"`
    } `json:"size"`
  } `json:"products"`
  Style string `json:"style"`
}

type FtlProductPageSelectedProduct struct {
  Price struct {
    CurrencyIso            string `json:"currencyIso"`
    FormattedOriginalPrice string `json:"formattedOriginalPrice"`
    FormattedValue         string `json:"formattedValue"`
    Value                  float64 `json:"value"`
  } `json:"price"`
  Style           string `json:"style"`
  Width           string `json:"width"`
  StyleSku        string `json:"styleSku"`
  StyleCode       string `json:"styleCode"`
  SizeCode string `json:"sizeCode"`
  MapEnable       bool   `json:"mapEnable"`
  WebOnlyMsg      string `json:"webOnlyMsg"`
  FreeShipping    bool   `json:"freeShipping"`
  WebOnlyLaunch   bool   `json:"webOnlyLaunch"`
  SkuExclusions   bool   `json:"skuExclusions"`
  IsLaunchProduct bool   `json:"isLaunchProduct"`
  IsInStock       bool   `json:"isInStock"`
  Fit             string `json:"fit"`
  Launch          struct {
    RecaptchaOn             bool   `json:"recaptchaOn"`
    CstSkuLaunchDate        string `json:"cstSkuLaunchDate"`
    DisplayCountDownTimer   bool   `json:"displayCountDownTimer"`
    SkuLaunchDate           int64  `json:"skuLaunchDate"`
    DefinedTimeForCountDown int64  `json:"definedTimeForCountDown"`
    LaunchDate              int    `json:"launchDate"`
    LaunchMonth             int    `json:"launchMonth"`
    LaunchMinutes           string `json:"launchMinutes"`
    LaunchHours             int    `json:"launchHours"`
    LaunchMeridiem          string `json:"launchMeridiem"`
  } `json:"launch"`
}

type FtlPdpSize struct {
  Name       string `json:"name"`
  Code       string `json:"code,omitempty"`
  IsDisabled bool   `json:"isDisabled"`
}
type FtlProductPageProductSummary struct {
  Name string `json:"name"`
}
type FtlProductPageState struct {
  Details struct {
    // Data map[string][]FtlProductPageDetails `json:"data"`
    Product map[string]FtlProductPageProductSummary `json:"product"`
    Sizes map[string][]FtlPdpSize `json:"sizes"`
    Styles map[string][]FtlPdpSize `json:"styles"`
    Selected map[string]FtlProductPageSelectedProduct `json:"selected"`
  } `json:"details"`

  HasAtcButton bool
}

func (s *FtlProductPageState) Name() string {
  for _, prod := range s.Details.Product {
    return prod.Name
  }
  return ""
}

func (s *FtlProductPageState) Style() string {
  for _, prod := range s.Details.Selected {
    return prod.Style
  }
  return ""
}

func (s *FtlProductPageState) Sizes() *[]FtlPdpSize {
  for _, prod := range s.Details.Sizes {
    if len(prod) > 0 {
      return &prod
    }
  }
  for _, sel := range s.Details.Selected {
    if len(sel.SizeCode) > 0 {
      return &[]FtlPdpSize{
        FtlPdpSize{Code:sel.SizeCode,Name:sel.Style},
      }
    }
  }
  return nil
}

func (s *FtlProductPageState) Product() *FtlProductPageSelectedProduct {
  for _, prod := range s.Details.Selected {
    return &prod
  }
  return nil
}

func (s *FtlProductPageState) ReleaseTime() time.Time {
  if p := s.Product(); p != nil {
    return time.Unix(int64(p.Launch.SkuLaunchDate / 1e3), 0)
  }
  return time.Time{}
}

type FtlProductResponse struct {
  // Product *FtlSelected
  ReleaseTime time.Time
  Skus []FtlSpecificProduct
	AkUrl string
  DetailsDEPRECATED FtlProductDetails `json:"details"`

  Config struct {
    Remote struct {
      PaymentAdyen            bool   `json:"paymentAdyen"`
    } `json:"remote"`
  } `json:"config"`
}

type FtlSession struct {
	Success bool `json:"success"`
	Data FtlSessionData `json:"data"`
  Time time.Time
}

type FtlSessionData struct {
	CsrfToken string `json:"csrfToken"`
}

type FtlOrderSummary struct {
  Order struct {
    Code string `json:"code"`
  } `json:"order"`
}
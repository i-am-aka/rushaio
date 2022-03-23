package task

var StatusSizeUnavailable = Status{Message: "Size Unavailable"}
var StatusVisitProduct = Status{Message: "Visiting Product"}
var StatusCooked = Status{Message: "COOKED!!!!", Level: 9999}
var StatusRequestError = Status{Message: "Request Error"}
var StatusProxyBanned = Status{Message: "Proxy Banned"}
var StatusOutOfStock = Status{Message: "Out of Stock"}
var StatusInQueue = Status{Message: "In Queue"}
var StatusCardDeclined = Status{Message: "Card Declined"}
var StatusPxBanned = Status{Message: "PX Banned"}
var StatusPxGen = Status{Message: "Generating PX Cookie"}
var StatusAkaBanned = Status{Message: "Akam*i Banned, Restarting Task"}
var StatusInWaitingRoom = Status{Message: "Waiting Room"}
var StatusRefreshingRecaptcha = Status{Message: "Refreshing Recaptcha"}
var StatusOosOnOrder = Status{Message: "OOS On Order: Trying Another Size"}
var StatusGeneratingCookie = Status{Message: "Generating Cookie"}
var StatusVisitHomepage = Status{Message: "Visiting Homepage"}
var StatusShippingAddyRejected = Status{Message: "Shipping Address Rejected"}
var StatusUnsupportedSite = Status{Message: "Unsupported Site"}
var StatusFnlProductNeedsSpecificity = Status{Message: "URL must include product ID, styleId, and colorId"}
var StatusWaitingForRelease = Status{Message: "Waiting For Release"}
var StatusWaitingForRestock = Status{Message: "Waiting For Restock"}
var StatusStartingSession = Status{Message: "Starting Session"}
var StatucAtc = Status{Message: "ATC"}
var StatusPrecarting = Status{Message: "Pre-carting"}
var StatusItemRemovedFromCart = Status{Message: "Item removed from cart"}
var StatusSubmittingInfo = Status{Message: "Sending Profile"}
var StatusOrdering = Status{Message: "Ordering"}
var StatusAtcFailed = Status{Message: "ATC Failed"}
var StatusCartJacked = Status{Message: "Cart Jacked"}
var StatusPassedWaitingRoom = Status{Message: "Passed Waiting Room"}
var StatusAccessDenied = Status{Message: "Access Denied (Retrying)"}
var StatusProfileAlreadyCheckedOutItem = Status{Message: "Profile Already Checked Out Item"}
var StatusCartExpired = Status{Message: "Cart Expired, Restarting Task"}
var StatusCartLost = Status{Message: "Cart Lost (Likely OOS)"}
var StatusWaitForCart = Status{Message: "Waiting for Cart"}
var StatusTestingProxy = Status{Message: "Testing Proxy"}
var StatusSessionBanRestart = Status{Message: "Session Banned, Restarting Task"}
var StatusPaypalCheckoutFailed = Status{Message: "PayPal Checkout Failed"}
var StatusStartingPpCheckout = Status{Message: "Starting PayPal Checkout"}
var StatusProxyBannedRestart = Status{Message: "Proxy Banned At ATC - Backing Off and Restarting Task"}
var StatusCarted = Status{Message: "Carted"}

var StatusDdEstablishingTrust = Status{Message: "Establishing Trust"}
var StatusCapSolved = Status{Message: "Captcha Solved"}
var StatusCapError = Status{Message: "Captcha Error"}

var StatusWaitStock = Status{Message: "Waiting for Stock"}
var StatusSolvingCap = Status{Message: "Waiting for Captcha"}

var StatusCaptchaSolveFailedRetrying = Status{Message: "Waiting for Captcha: Retrying"}
var StatusShipOptionUnavailable = Status{Message: "Shipping option unavailable"}
var StatusLoggingIn = Status{Message: "Signing in"}
var StatusCreatingAccount = Status{Message: "Creating Account"}
var StatusRotatingProxy = Status{Message: "Rotating Proxy"}
var StatusFarming = Status{Message: "Waiting for Farm"}
var StatusProxyConnectionFailed = Status{Message: "Proxy Connection Failed"}
var StatusFarmUnreachable = Status{Message: "Farm Unreachable, Make Sure It Is Running"}
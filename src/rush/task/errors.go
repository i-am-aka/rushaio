package task

import (
	"github.com/pkg/errors"
)

type errTaskUnrecoverable struct {
	error
}

func TaskErrUnrecoverable(err error) error {
	return errTaskUnrecoverable{err}
}

func ErrIsTaskUnrecoverable(err error) bool {
	if err == nil {
		return false
	}
	_, match := errors.Unwrap(err).(errTaskUnrecoverable)
	return match
}

type UnsupportedSiteError struct {
	Site string
	Err error
}
func (e *UnsupportedSiteError) Error() string { return "Unsupported site: " + e.Site }
func (e *UnsupportedSiteError) Unwrap() error { return e.Err }

type ErrSessionBanned struct {
	Stage string
	Err error
}
func (e *ErrSessionBanned) Error() string { return "Session Banned (" + e.Stage + ")" }
func (e *ErrSessionBanned) Unwrap() error { return TaskErrUnrecoverable(e.Err) }

var ErrRateLimited = errors.New("Rate limited")
var errInQueue = errors.New("In Queue")
var errForbidden = errors.New("403 Forbidden")
var ErrPxBanned = errors.New("PX Banned")
var ErrAkamaiBanned = errors.New("Ak*mai Banned")
var ErrUnexpectedState = errors.New("Unexpected State")
var ErrProductUnavailable = errors.New("Product Unavailable")
var ErrOutOfStock = errors.New("Out of Stock")
var ErrProxyBanned = errors.New("Proxy Banned")
var ErrInvalidProductResponse = errors.New("Invalid Product Response")
var ErrRetrying = errors.New("Retrying")
var ErrFailedToGenCookie = errors.New("Failed to Gen Cookie")
var ErrUnknownCheckoutError = errors.New("Unknown Checkout Error")
var ErrOosOnOrder = errors.New("OOS at Order")
var ErrAtcFailed = errors.New("ATC Failed")
var ErrNoSizesAvailable = errors.New("Desired Sizes OOS")
var ErrProxyNotLocatedInUsa =  errors.New("Error: Proxy Not Located In USA")
var ErrNotImplemented = errors.New("ErrNotImplemented")
var ErrInvalidUrl = errors.New("Invalid URL")
var ErrCartJacked = errors.New("Cart Jacked")
var ErrSessionExpired = errors.New("Session expired")
var ErrCookieValidationDisabled = errors.New("Cookie validation disabled")
var ErrCardRejected = errors.New("card rejected")
var ErrWaitingForPermaQueue = errors.New("waiting for perma queue")
var ErrShippingAddressRejected = errors.New("Shipping address rejected")
var ErrDdBanned = errors.New("Datadome banned")
var ErrTimedOut = errors.New("timed out")
var ErrFailedToSolveCaptcha= errors.New("Failed to solved captcha")
var ErrAlreadyFarmed = errors.New("Already farmed")
var ErrProxyConnectionFailed = errors.New("Proxy Connection Failed")
var ErrRotatedProxySessionRestartRequired = errors.New("Rotated Proxy")
var ErrShipOptionUnavailable = errors.New("Shipping option unavailable")
var ErrUnknownResponse = errors.New("Unknown response")
var ErrForbiddeen = errors.New("forbidden")
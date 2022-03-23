package task

import (
	"os"
	"fmt"
	"strings"
)

func (t *CheckoutTask) GetRecapToken(host string, sitekey string, action string) (string, error) {
	t.StatusCh <- StatusRefreshingRecaptcha
	origApiHost := os.Getenv("API_HOST")
	os.Setenv("API_HOST", "https://api.rushaio.co")
	defer os.Setenv("API_HOST", origApiHost)
	resp, err := t.ApiGetRetries(fmt.Sprintf("/rcgen?host=%s&sitekey=%s&action=%s&staff=1", host, sitekey, action), 40)
	return strings.TrimSpace(string(resp)), err
}
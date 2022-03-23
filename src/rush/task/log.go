package task

import (
	"log"
	"os"
)

func (t *CheckoutTask) LogDebug(fmtstr string, args ...interface{}) {
	if t.Debug || os.Getenv("LOG") == "1" {
		log.Printf(fmtstr, args...)
	}
}
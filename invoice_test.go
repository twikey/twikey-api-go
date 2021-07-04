package twikey

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestInvoiceFeed(t *testing.T) {
	c := TwikeyClient{
		BaseURL: getEnv("TWIKEY_URL", "https://api.twikey.com"),
		ApiKey:  os.Getenv("TWIKEY_API_KEY"),
		//Debug: log.Default(),
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}

	t.Run("InvoiceFeed", func(t *testing.T) {
		err := c.InvoiceFeed(func(invoice Invoice) {
			t.Log("Invoice", invoice.Number, invoice.State)
		})
		if err != nil {
			return
		}
	})
}

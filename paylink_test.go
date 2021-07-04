package twikey

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestPaylinkFeed(t *testing.T) {
	c := TwikeyClient{
		BaseURL: getEnv("TWIKEY_URL", "https://api.twikey.com"),
		ApiKey:  os.Getenv("TWIKEY_API_KEY"),
		//Debug: log.Default(),
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}

	t.Run("PaylinkFeed", func(t *testing.T) {
		err := c.PaylinkFeed(func(paylink Paylink) {
			t.Log("Paylink", paylink.Amount, paylink.Msg, paylink.State)
		})
		if err != nil {
			return
		}
	})
}

package twikey

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestPaylinkFeed(t *testing.T) {
	if os.Getenv("TWIKEY_API_KEY") == "" {
		t.Skip("No TWIKEY_API_KEY available")
	}

	c := Client{
		BaseURL: getEnv("TWIKEY_URL", "https://api.beta.twikey.com"),
		APIKey:  os.Getenv("TWIKEY_API_KEY"),
		//Debug: log.Default(),
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}

	paylink, err := c.PaylinkNew(&PaylinkRequest{
		Title:          "Test Message",
		Remittance:     "12345679810",
		Amount:         10.0,
		CustomerNumber: "123",
		Email:          "no-repy@twikey.com",
		RedirectUrl:    "",
		Method:         "ideal",
	})
	if err != nil {
		t.Fatal(err)
	}

	if paylink == nil || paylink.Url == "" {
		t.Error("No valid link retrieved")
	}

	t.Run("PaylinkFeed", func(t *testing.T) {
		err := c.PaylinkFeed(func(paylink *Paylink) {
			t.Log("Paylink", paylink.Amount, paylink.Msg, paylink.State)
		})
		if err != nil {
			return
		}
	})
}

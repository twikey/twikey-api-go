package twikey

import (
	"context"
	"os"
	"testing"
)

func TestPaylinkFeed(t *testing.T) {
	if os.Getenv("TWIKEY_API_KEY") == "" {
		t.Skip("No TWIKEY_API_KEY available")
	}

	c := newTestClient()
	paylink, err := c.PaylinkNew(context.Background(), &PaylinkRequest{
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
		err := c.PaylinkFeed(context.Background(), func(paylink *Paylink) {
			t.Logf("Paylink update #%d %.2f Euro with new state=%s", paylink.Id, paylink.Amount, paylink.State)
		})
		if err != nil {
			t.Error(err)
		}
	})
}

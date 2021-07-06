package twikey

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestDocumentFeed(t *testing.T) {
	if os.Getenv("TWIKEY_API_KEY") == "" {
		t.Skip("No TWIKEY_API_KEY available")
	}

	c := TwikeyClient{
		BaseURL: getEnv("TWIKEY_URL", "https://api.beta.twikey.com"),
		ApiKey:  os.Getenv("TWIKEY_API_KEY"),
		//Debug:   log.Default(),
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}

	invite, err := c.DocumentInvite(InviteRequest{
		ct:             getEnv("CT", "1"),
		customerNumber: "123",
		amount:         "123.10",
		email:          "john@doe.com",
		firstname:      "John",
		lastname:       "Doe",
		l:              "en",
		address:        "Abbey road",
		city:           "Liverpool",
		zip:            "1526",
		country:        "BE",
		mobile:         "",
		companyName:    "",
		coc:            "",
		iban:           "",
		bic:            "",
		mandateNumber:  "",
		contractNumber: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(invite.Url)

	t.Run("DocumentFeed", func(t *testing.T) {
		err := c.DocumentFeed(func(new Mndt) {
			t.Log("new", new.MndtId)
		}, func(update Mndt, reason AmdmntRsn) {
			t.Log("update", update.MndtId, reason.Rsn)
		}, func(mandate string, reason CxlRsn) {
			t.Log("cancelled", mandate, reason.Rsn)
		})
		if err != nil {
			return
		}
	})

}

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

	c := Client{
		BaseURL: getEnv("TWIKEY_URL", "https://api.beta.twikey.com"),
		APIKey:  os.Getenv("TWIKEY_API_KEY"),
		//Debug:   log.Default(),
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}

	invite, err := c.DocumentInvite(InviteRequest{
		Template:       getEnv("CT", "1"),
		CustomerNumber: "123",
		Amount:         "123.10",
		Email:          "john@doe.com",
		Firstname:      "John",
		Lastname:       "Doe",
		Language:       "en",
		Address:        "Abbey road",
		City:           "Liverpool",
		Zip:            "1526",
		Country:        "BE",
		Mobile:         "",
		CompanyName:    "",
		Coc:            "",
		Iban:           "",
		Bic:            "",
		MandateNumber:  "",
		ContractNumber: "",
	})
	if err != nil {
		t.Fatal(err)
	}

	if invite == nil || invite.Url == "" {
		t.Error("No valid invite retrieved")
	}

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

func TestDocumentDetail(t *testing.T) {
	if os.Getenv("TWIKEY_API_KEY") == "" {
		t.Skip("No TWIKEY_API_KEY available")
	}

	if os.Getenv("MNDTNUMBER") == "" {
		t.Skip("No MNDTNUMBER available")
	}

	c := Client{
		BaseURL: getEnv("TWIKEY_URL", "https://api.beta.twikey.com"),
		APIKey:  os.Getenv("TWIKEY_API_KEY"),
		//Debug:   log.Default(),
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}

	mndt, err := c.DocumentDetail(os.Getenv("MNDTNUMBER"))
	if err != nil {
		t.Fatal(err)
	}

	if mndt == nil {
		t.Error("No valid mandate retrieved")
	} else {
		t.Log("new", mndt.MndtId)
	}

}

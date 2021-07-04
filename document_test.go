package twikey

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestDocumentFeed(t *testing.T) {
	c := TwikeyClient{
		BaseURL: getEnv("TWIKEY_URL", "https://api.twikey.com"),
		ApiKey:  os.Getenv("TWIKEY_API_KEY"),
		//Debug: log.Default(),
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
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

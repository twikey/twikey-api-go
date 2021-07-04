package twikey

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestTransactions(t *testing.T) {
	c := TwikeyClient{
		BaseURL: getEnv("TWIKEY_URL", "https://api.twikey.com"),
		ApiKey:  os.Getenv("TWIKEY_API_KEY"),
		//Debug: log.Default(),
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}

	t.Run("New Transaction without valid mandate", func(t *testing.T) {
		tx, err := c.TransactionNew(TransactionRequest{
			DocumentReference: "ABC",
			Msg:               "My Transaction",
			Ref:               "My Reference",
			Amount:            10.90,
		})
		if err != nil {
			if err.Error() != "No contract was found" && err.Error() != "Not authorised" {
				t.Fatal(err)
			}
		} else {
			t.Fatal(tx)
		}
	})

	t.Run("TransactionFeed", func(t *testing.T) {
		err := c.TransactionFeed(func(transaction Transaction) {
			t.Log("Transaction", transaction.Amount, transaction.BookedError, transaction.Final)
		})
		if err != nil {
			return
		}
	})
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

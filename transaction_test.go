package twikey

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestTransactions(t *testing.T) {
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

	t.Run("New Transaction without valid mandate", func(t *testing.T) {
		tx, err := c.TransactionNew(&TransactionRequest{
			DocumentReference: "ABC",
			Msg:               "No valid mandate",
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

	t.Run("New reservation with valid mandate ", func(t *testing.T) {
		if os.Getenv("MNDTNUMBER") == "" {
			t.Skip("No MNDTNUMBER available")
		}

		tx, err := c.ReservationNew(&TransactionRequest{
			DocumentReference: os.Getenv("MNDTNUMBER"),
			Msg:               "My Transaction",
			Ref:               "My Reference",
			Amount:            10.90,
			Force:             true, // allow second reservation
		})
		if err != nil {
			if err.Error() != "No contract was found" && err.Error() != "Not authorised" {
				t.Fatal(err)
			}
		} else if tx == nil {
			t.Fatal("No reservation was done")
		}
	})

	t.Run("TransactionFeed", func(t *testing.T) {
		err := c.TransactionFeed(func(transaction *Transaction) {
			t.Log("Transaction", transaction.Amount, transaction.BookedError, transaction.Final)
		}, "meta")
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

package twikey

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestInvoiceFeed(t *testing.T) {
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

	t.Run("InvoiceFeed", func(t *testing.T) {
		err := c.InvoiceFeed(func(invoice *Invoice) {
			t.Log("Invoice", invoice.Number, invoice.State)
		})
		if err != nil {
			return
		}
	})
}

func TestInvoiceAdd(t *testing.T) {
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

	t.Run("Invoice", func(t *testing.T) {
		invoice, err := c.InvoiceAdd(context.Background(), &Invoice{
			Number:     "123",
			Title:      "TestInvoice 123",
			Date:       "2021-01-01",
			Duedate:    "2021-03-01",
			Remittance: "123",
			Amount:     10.00,
			Customer: &Customer{
				CustomerNumber: "123",
				Email:          "support@twikey.com",
				Address:        "Derbystraat 43",
				City:           "Gent",
				Zip:            "9051",
				Country:        "BE",
				Language:       "nl",
			},
		})
		if err != nil {
			t.Error(err)
		} else {
			t.Log("New invoice", invoice.Id)
		}
	})
}

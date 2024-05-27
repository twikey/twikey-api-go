package twikey

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestInvoiceFeed(t *testing.T) {
	if os.Getenv("TWIKEY_API_KEY") == "" {
		t.Skip("No TWIKEY_API_KEY available")
	}

	c := newTestClient()
	t.Run("InvoiceFeed", func(t *testing.T) {
		err := c.InvoiceFeed(context.Background(), func(invoice *Invoice) {
			newState := ""
			if invoice.State == "PAID" {
				lastPayment := (*invoice.LastPayment)[0]
				via := ""
				if lastPayment["method"] != nil {
					via = fmt.Sprintf(" via %s", lastPayment["method"])
				}
				date := ""
				if lastPayment["date"] != nil {
					via = fmt.Sprintf(" on %s", lastPayment["date"])
				}
				newState = "PAID" + via + date
			} else {
				newState = "now has state " + invoice.State
			}

			t.Logf("Invoice update with number %s %.2f euro %s", invoice.Number, invoice.Amount, newState)
		}, FeedInclude("lastpayment", "meta", "customer"))
		if err != nil {
			t.Error(err)
		}
	})
}

func TestInvoiceAddAndUpdate(t *testing.T) {
	if os.Getenv("TWIKEY_API_KEY") == "" {
		t.Skip("No TWIKEY_API_KEY available")
	}

	c := newTestClient()
	t.Run("InvoiceAddAndUpdate", func(t *testing.T) {
		invoice, err := c.InvoiceAdd(context.Background(), &NewInvoiceRequest{
			Invoice: &Invoice{
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
			},
			Origin: "Go-Test",
		})
		if err != nil {
			t.Error(err)
		} else {
			t.Log("New invoice", invoice.Id)
		}

		ctx := context.Background()
		cnote, err := c.InvoiceAdd(ctx, &NewInvoiceRequest{
			Invoice: &Invoice{
				Number:         "124",
				RelatedInvoice: "123",
				Title:          "TestCreditNote 123",
				Date:           "2021-01-02",
				Duedate:        "2021-03-01",
				Remittance:     "123",
				Amount:         -1.00,
				Customer: &Customer{
					CustomerNumber: "123",
				},
				Extra: map[string]string{
					"attr1": "test",
				},
			},
			Origin: "Go-Test",
		})
		if err != nil {
			t.Error(err)
		} else {
			t.Log("New CreditNote", cnote.Id)
		}

		if err := c.InvoiceUpdate(ctx, &UpdateInvoiceRequest{
			ID:    invoice.Id,
			Title: "Some updated title",
		}); err != nil {
			if err != nil {
				t.Error(err)
			} else {
				t.Log("Updated invoice", invoice.Id)
			}
		}
	})
}

func TestInvoiceUpdateWithInvalidRequest(t *testing.T) {
	if os.Getenv("TWIKEY_API_KEY") == "" {
		t.Skip("No TWIKEY_API_KEY available")
	}

	c := newTestClient()
	t.Run("InvoiceUpdateWithInvalidRequest", func(t *testing.T) {
		invoice, err := c.InvoiceAdd(context.Background(), &NewInvoiceRequest{
			Invoice: &Invoice{
				Number:     "123123",
				Title:      "TestInvoice 123123",
				Date:       "2021-01-01",
				Duedate:    "2021-03-01",
				Remittance: "123123",
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
			},
			Origin: "Go-Test",
		})
		if err != nil {
			t.Error(err)
		} else {
			t.Log("New invoice", invoice.Id)
		}

		ctx := context.Background()
		if err := c.InvoiceUpdate(ctx, &UpdateInvoiceRequest{
			Title: "Some updated title",
		}); err == nil {
			t.Error("Update invoice call did not return an error even though we send no ID")
		}
	})
}

package twikey

import (
	"context"
	"os"
	"testing"
)

func TestTransactions(t *testing.T) {
	if os.Getenv("TWIKEY_API_KEY") == "" {
		t.Skip("No TWIKEY_API_KEY available")
	}

	c := newTestClient()
	t.Run("New Transaction without valid mandate", func(t *testing.T) {
		tx, err := c.TransactionNew(context.Background(), &TransactionRequest{
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

		tx, err := c.ReservationNew(context.Background(), &ReservationRequest{
			DocumentReference: os.Getenv("MNDTNUMBER"),
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
		err := c.TransactionFeed(context.Background(), func(transaction *Transaction) {
			state := transaction.State
			final := transaction.Final
			ref := transaction.Ref
			if ref == "" {
				ref = transaction.Message
			}
			_state := state
			_final := ""
			if state == "PAID" {
				_state = "is now paid"
			} else if state == "ERROR" {
				_state = "failed due to '" + transaction.BookedError + "'"
			}
			// final means Twikey has gone through all dunning steps, but customer still did not pay
			if final {
				_final = "with no more dunning steps"
			}
			t.Log("Transaction update", transaction.Amount, "euro with", ref, _state, _final)
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

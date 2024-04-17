package twikey

import (
	"context"
	"os"
	"testing"
)

func TestRefundFeed(t *testing.T) {
	if os.Getenv("TWIKEY_API_KEY") == "" {
		t.Skip("No TWIKEY_API_KEY available")
	}

	c := newTestClient()
	t.Run("RefundFeed", func(t *testing.T) {
		err := c.RefundFeed(context.Background(), func(refund *Refund) {
			t.Logf("Refund update #%s %.2f Euro with new state=%s", refund.Id, refund.Amount, refund.State)
		})
		if err != nil {
			t.Error(err)
		}
	})
}

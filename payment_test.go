package twikey

import (
	"encoding/json"
	"testing"
)

func TestPaymentAmountUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    float64
		wantErr bool
	}{
		{name: "number", payload: `{"amount":12.34}`, want: 12.34},
		{name: "integer", payload: `{"amount":12}`, want: 12},
		{name: "string", payload: `{"amount":"12.34"}`, want: 12.34},
		{name: "missing", payload: `{}`, want: 0},
		{name: "null", payload: `{"amount":null}`, want: 0},
		{name: "invalid", payload: `{"amount":"abc"}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payment
			err := json.Unmarshal([]byte(tt.payload), &p)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %s", tt.payload)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Amount != tt.want {
				t.Errorf("amount = %v, want %v", p.Amount, tt.want)
			}
		})
	}
}

func TestPaymentUnmarshalKeepsOtherFields(t *testing.T) {
	payload := `{
		"eventId": "evt-1",
		"eventType": "payment_failure",
		"occurredAt": "2026-08-23T10:54:00Z",
		"amount": "42.00",
		"currency": "EUR",
		"origin": {"object": "invoice", "number": "2026-001"},
		"error": {"code": "AC04", "description": "Closed account"}
	}`

	var p Payment
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Amount != 42.00 {
		t.Errorf("amount = %v, want 42.00", p.Amount)
	}
	if p.EventID != "evt-1" || p.EventType != "payment_failure" || p.Currency != "EUR" {
		t.Errorf("unexpected scalar fields: %+v", p)
	}
	if p.OccurredAt.IsZero() {
		t.Error("occurredAt was not parsed")
	}
	if p.Origin.Object != "invoice" || p.Origin.Number != "2026-001" {
		t.Errorf("unexpected origin: %+v", p.Origin)
	}
	if p.Error == nil || p.Error.Code != "AC04" {
		t.Errorf("unexpected error object: %+v", p.Error)
	}
}

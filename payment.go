package twikey

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type PaymentFeed struct {
	Payments []Payment `json:"Payments"`
}

type Payment struct {
	EventID    string         `json:"eventId"`
	EventType  string         `json:"eventType"`
	OccurredAt time.Time      `json:"occurredAt"`
	Amount     float64        `json:"amount"`
	Currency   string         `json:"currency"`
	Origin     Origin         `json:"origin"`
	Gateway    Gateway        `json:"gateway"`
	Details    PaymentDetails `json:"details"`
	Error      *PaymentError  `json:"error,omitempty"`
}

// UnmarshalJSON accepts the amount as either a json number or a quoted string,
// as the api returns both depending on the origin of the payment.
func (p *Payment) UnmarshalJSON(data []byte) error {
	type alias Payment
	aux := struct {
		Amount json.RawMessage `json:"amount"`
		*alias
	}{alias: (*alias)(p)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	raw := strings.TrimSpace(string(aux.Amount))
	if raw == "" || raw == "null" {
		return nil
	}

	amount, err := strconv.ParseFloat(strings.Trim(raw, `"`), 64)
	if err != nil {
		return fmt.Errorf("invalid amount %s: %w", raw, err)
	}
	p.Amount = amount
	return nil
}

type Origin struct {
	Object string `json:"object"`
	ID     string `json:"id"`
	Number string `json:"number"`
	Ref    string `json:"ref"`
}

type Gateway struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`
	Type string  `json:"type"`
	IBAN *string `json:"iban"`
}

type PaymentDetails struct {
	// Common
	Source string `json:"source"`

	// Direct debit
	PaymentID      *int    `json:"paymentId,omitempty"`
	TransactionE2E *string `json:"transactionE2e,omitempty"`
	MandateID      *string `json:"mndtId,omitempty"`

	// Payment link
	LinkID     *int    `json:"linkId,omitempty"`
	LinkMethod *string `json:"linkMethod,omitempty"`

	// Refund / credit transfer
	CustomerIBAN *string `json:"customerIban,omitempty"`
	RefundE2E    *string `json:"refundE2e,omitempty"`
}

type PaymentError struct {
	Code         string `json:"code"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	ExternalCode string `json:"externalCode"`
	Action       string `json:"action"`
	ActionStep   int    `json:"actionStep"`
}

package twikey

import "time"

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

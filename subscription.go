package twikey

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Recurrence string

const (
	RecurrenceWeekly     Recurrence = "1w"
	RecurrenceMonthly    Recurrence = "1m"
	RecurrenceBiMonthly  Recurrence = "2m"
	RecurrenceQuarterly  Recurrence = "3m"
	RecurrenceTrimestral Recurrence = "4m"
	RecurrenceSemiAnnual Recurrence = "6m"
	RecurrenceAnnual     Recurrence = "12m"
)

type SubscriptionState string

const (
	SubscriptionStateActive    SubscriptionState = "active"
	SubscriptionStateSuspended SubscriptionState = "suspended"
	SubscriptionStateCancelled SubscriptionState = "cancelled"
	SubscriptionStateClosed    SubscriptionState = "closed"
)

type Subscription struct {
	Id         int               `json:"id"`
	State      SubscriptionState `json:"state"`
	Amount     float64           `json:"amount"`
	Message    string            `json:"message"`
	Ref        string            `json:"ref"`
	Plan       int               `json:"plan"`
	Runs       int               `json:"runs"`
	StopAfter  int               `json:"stopAfter"`
	Start      string            `json:"start"`
	Next       string            `json:"next"`
	Recurrence Recurrence        `json:"recurrence"`
	MndtId     string            `json:"mndtId"`
}

type SubscriptionAddRequest struct {
	// Unique key usable only once per request every 24hrs.
	IdempotencyKey string
	// Mandate reference for which to add the subscription to.
	MndtId string
	// The message the subscriber will see.
	Message string
	// Name of the base plan.
	Plan string
	// Reference of the subscription (important for updates), it is converted to uppercase and can't contain any spaces.
	Ref string
	// Amount of the transaction.
	Amount float64
	// Number of time the subscription should be executed. Set to a value lower than 1 for an unbounded subscription.
	StopAfter int
	// The frequency of the subscription, by default it will be monthly.
	Recurrence Recurrence
	// Start of subscription eg. 2022-11-01 (only future dates are allowed).
	StartDate string
}

// asUrlParams returns the form URL encoded parameters for the incoming request.
func (r *SubscriptionAddRequest) asUrlParams() string {
	params := url.Values{}
	params.Add("mndtId", r.MndtId)
	params.Add("message", r.Message)
	params.Add("amount", fmt.Sprintf("%.2f", r.Amount))
	params.Add("start", r.StartDate)
	if r.Plan != "" {
		params.Add("plan", r.Plan)
	}
	if r.Ref != "" {
		params.Add("ref", r.Ref)
	}
	if r.StopAfter > 0 {
		params.Add("stopAfter", strconv.Itoa(r.StopAfter))
	}
	if r.Recurrence != "" {
		params.Add("recurrence", string(r.Recurrence))
	}
	return params.Encode()
}

// SubscriptionAdd will add a subscription to an existing agreement. This means than when the subscription is run a
// new transaction will automatically be created using the defined schedule.
func (c *Client) SubscriptionAdd(ctx context.Context, payload *SubscriptionAddRequest) (*Subscription, error) {
	input := strings.NewReader(payload.asUrlParams())
	endpoint := c.BaseURL + "/creditor/subscription"
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, input)
	if payload.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", payload.IdempotencyKey)
	}

	var output Subscription
	if err := c.sendRequest(req, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

type UpdateSubscriptionRequest struct {
	// The mandate reference of the subscription, creates a new subscription when different from current id.
	MndtId string
	// Message to the subscriber.
	Message string
	// Amount of the transaction.
	Amount float64
	// Start date of the subscription (yyyy-mm-dd). This is also the first execution date. Only a future date is accepted.
	Start string
	// Name of the base plan, When passing a plan the values of message, amount and recurrence are ignored if passed during the request.
	Plan string
	// The frequency of the subscription, by default it will be monthly.
	Recurrence Recurrence
	// Number of times to execute. Previous executions (runs) are not taken into account for the new subscription. It starts from 0 runs.
	StopAfter int
}

func (r *UpdateSubscriptionRequest) asUrlParams() string {
	params := url.Values{}
	params.Add("mndtId", r.MndtId)
	params.Add("message", r.Message)
	params.Add("amount", fmt.Sprintf("%.2f", r.Amount))
	params.Add("start", r.Start)
	if r.Plan != "" {
		params.Add("plan", r.Plan)
	}
	if r.Recurrence != "" {
		params.Add("recurrence", string(r.Recurrence))
	}
	if r.StopAfter > 0 {
		params.Add("stopAfter", strconv.Itoa(r.StopAfter))
	}
	return params.Encode()
}

// SubscriptionUpdate will update a subscription. This endpoint allows the update by using the previously passed reference
// for a specific agreement. The update subscription and patch subscription are similar requests, the difference be that
// with the [Client.SubscriptionUpdate] you can replace a subscription (cancel current and start new) the [Client.SubscriptionPatch] can't replace a subscription.
func (c *Client) SubscriptionUpdate(ctx context.Context, mandate string, ref string, payload *UpdateSubscriptionRequest) (*Subscription, error) {
	if mandate == "" || ref == "" {
		return nil, errors.New("mandate reference and subscription reference are required")
	}

	input := strings.NewReader(payload.asUrlParams())
	endpoint := fmt.Sprintf("%s/creditor/subscription/%s/%s", c.BaseURL, mandate, ref)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, input)
	var output Subscription
	if err := c.sendRequest(req, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

type PatchSubscriptionRequest struct {
	// Move the subscription to a different mandate.
	MndtId string
	// message to the subscriber.
	Message string
	// Amount of the transaction that will be created based on the subscription recurrence.
	Amount float64
}

func (r *PatchSubscriptionRequest) asUrlParams() string {
	params := url.Values{}
	if r.MndtId != "" {
		params.Add("mndtId", r.MndtId)
	}
	if r.Message != "" {
		params.Add("message", r.Message)
	}
	if r.Amount > 0 {
		params.Add("amount", fmt.Sprintf("%.2f", r.Amount))
	}
	return params.Encode()
}

// SubscriptionPatch will update the subscription without replacing it. It allows you to update specific fields or move the subscription to a different mandate.
func (c *Client) SubscriptionPatch(ctx context.Context, mandate string, ref string, payload *PatchSubscriptionRequest) (*Subscription, error) {
	input := payload.asUrlParams()
	endpoint := fmt.Sprintf("%s/creditor/subscription/%s/%s?%s", c.BaseURL, mandate, ref, input)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, nil)
	var output Subscription
	if err := c.sendRequest(req, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// SubscriptionCancel subscription can be cancelled by using its ref for a specific agreement.
func (c *Client) SubscriptionCancel(ctx context.Context, mandate string, ref string) error {
	endpoint := fmt.Sprintf("%s/creditor/subscription/%s/%s", c.BaseURL, mandate, ref)
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	return c.sendRequest(req, nil) // no content response
}

// SubscriptionDetail retrieves a single subscription for a specific agreement.
func (c *Client) SubscriptionDetail(ctx context.Context, mandate string, ref string) (*Subscription, error) {
	endpoint := fmt.Sprintf("%s/creditor/subscription/%s/%s", c.BaseURL, mandate, ref)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	var output Subscription
	if err := c.sendRequest(req, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

type SubscriptionListRequest struct {
	// Mandate reference
	MndtId string
	// CustomerNumber specifies the reference of a customer.
	CustomerNumber string
	// State of the subscription (active, suspended, cancelled, closed)
	State SubscriptionState
	// Page of the results (if more than 1 is available)
	Page int
}

func (r *SubscriptionListRequest) asUrlParams() string {
	params := url.Values{}
	if r.MndtId != "" {
		params.Add("mndtId", r.MndtId)
	}
	if r.CustomerNumber != "" {
		params.Add("customerNumber", r.CustomerNumber)
	}
	if r.State != "" {
		params.Add("state", string(r.State))
	}
	if r.Page > 0 {
		params.Add("page", strconv.Itoa(r.Page))
	}
	return params.Encode()
}

// NextPage will increment the current page number of the subscription list request.
func (r *SubscriptionListRequest) NextPage() *SubscriptionListRequest {
	r.Page++
	return r
}

type SubscriptionListResponse struct {
	Subscriptions []Subscription `json:"Subscriptions"`
	Links         struct {
		Previous string `json:"previous"`
		Self     string `json:"self"`
		Next     string `json:"next"`
	} `json:"_links"`
}

// HasNext will return true if another page of results is available.
func (r *SubscriptionListResponse) HasNext() bool {
	return r.Links.Next != ""
}

// SubscriptionList retrieves all subscriptions matching the query.
func (c *Client) SubscriptionList(ctx context.Context, payload *SubscriptionListRequest) (*SubscriptionListResponse, error) {
	input := payload.asUrlParams()
	endpoint := c.BaseURL + "/creditor/subscription/query"
	if input != "" {
		endpoint += "?" + input
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	var output SubscriptionListResponse
	if err := c.sendRequest(req, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// SubscriptionSuspend will suspend the referenced subscription.
func (c *Client) SubscriptionSuspend(ctx context.Context, mandate string, ref string) error {
	return c.subscriptionAction(ctx, mandate, ref, "suspend")
}

// SubscriptionResume will resume the currently suspended subscription.
func (c *Client) SubscriptionResume(ctx context.Context, mandate string, ref string) error {
	return c.subscriptionAction(ctx, mandate, ref, "resume")
}

// subscriptionAction will perform the given action on the referenced subscription.
func (c *Client) subscriptionAction(ctx context.Context, mandate string, ref string, action string) error {
	endpoint := fmt.Sprintf("%s/creditor/subscription/%s/%s/%s", c.BaseURL, mandate, ref, action)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	return c.sendRequest(req, nil)
}

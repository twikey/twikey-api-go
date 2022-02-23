package twikey

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TransactionRequest is the payload to be send to Twikey when a new transaction should be send
type TransactionRequest struct {
	DocumentReference             string
	TransactionDate               string
	RequestedCollection           string
	Msg                           string
	Ref                           string
	Amount                        float64
	Place                         string
	ReferenceIsEndToEndIdentifier bool
	Reservation                   string // via reservation request
	Force                         bool
}

// Transaction is the response from Twikey when updates are received
type Transaction struct {
	Id                  int64   `json:"id,omitempty"`
	DocumentId          int64   `json:"contractId,omitempty"`
	DocumentReference   string  `json:"mndtId"`
	Amount              float64 `json:"amount"`
	Message             string  `json:"msg"`
	Ref                 string  `json:"ref"`
	Place               string  `json:"place"`
	Final               bool    `json:"final"`
	State               string  `json:"state"`
	BookedDate          string  `json:"bkdate"`
	BookedError         string  `json:"bkerror"`
	BookedAmount        float64 `json:"bkamount"`
	RequestedCollection string  `json:"reqcolldt"`
}

// Reservation is the response from Twikey when updates are received
type Reservation struct {
	Id             string    `json:"id"`
	MndtId         string    `json:"mndtId"`
	ReservedAmount float64   `json:"reservedAmount"`
	Expires        time.Time `json:"expires"`
}

// TransactionList is a struct to contain the response coming from Twikey, should be considered internal
type TransactionList struct {
	Entries []Transaction
}

// TransactionNew sends a new transaction to Twikey
func (c *Client) TransactionNew(transaction *TransactionRequest) (*Transaction, error) {

	params := url.Values{}
	params.Add("mndtId", transaction.DocumentReference)
	params.Add("date", transaction.TransactionDate)
	params.Add("reqcolldt", transaction.RequestedCollection)
	params.Add("amount", fmt.Sprintf("%.2f", transaction.Amount))
	params.Add("message", transaction.Msg)
	params.Add("ref", transaction.Ref)
	params.Add("place", transaction.Place)
	if transaction.Force {
		params.Add("force", "true")
	}
	if transaction.ReferenceIsEndToEndIdentifier {
		params.Add("refase2e", "true")
	}

	c.debug("New transaction", params)

	req, _ := http.NewRequest("POST", c.BaseURL+"/creditor/transaction", strings.NewReader(params.Encode()))
	if transaction.Reservation != "" {
		req.Header.Add("X-RESERVATION", transaction.Reservation)
	}
	var transactionList TransactionList
	err := c.sendRequest(req, transactionList)
	if err != nil {
		return nil, err
	}
	return &transactionList.Entries[0], nil
}

// ReservationNew sends a new reservation to Twikey
func (c *Client) ReservationNew(transaction *TransactionRequest) (*Reservation, error) {

	params := url.Values{}
	params.Add("mndtId", transaction.DocumentReference)
	params.Add("date", transaction.TransactionDate)
	params.Add("reqcolldt", transaction.RequestedCollection)
	params.Add("amount", fmt.Sprintf("%.2f", transaction.Amount))
	params.Add("message", transaction.Msg)
	params.Add("ref", transaction.Ref)
	params.Add("place", transaction.Place)
	params.Add("reservation", "true")
	if transaction.Force {
		params.Add("force", "true")
	}

	c.debug("New reservation", params)
	req, _ := http.NewRequest("POST", c.BaseURL+"/creditor/reservation", strings.NewReader(params.Encode()))
	reservation := &Reservation{}
	err := c.sendRequest(req, reservation)
	return reservation, err

}

// TransactionFeed retrieves all transaction updates since the last call with a callback since there may be many
func (c *Client) TransactionFeed(callback func(transaction *Transaction), sideloads ...string) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	_url := c.BaseURL + "/creditor/transaction"
	for i, sideload := range sideloads {
		if i == 0 {
			_url = _url + "?include=" + sideload
		} else {
			_url = _url + "&include=" + sideload
		}
	}

	for {
		req, _ := http.NewRequest("GET", _url, nil)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", c.apiToken)
		req.Header.Set("User-Agent", c.UserAgent)

		res, _ := c.HTTPClient.Do(req)

		if res.StatusCode == 200 {
			_ = res.Body.Close()
			payload, _ := ioutil.ReadAll(res.Body)
			var paymentResponse TransactionList
			err := json.Unmarshal(payload, &paymentResponse)
			if err == nil {
				c.debug(fmt.Sprintf("Fetched %d transactions", len(paymentResponse.Entries)))
				if len(paymentResponse.Entries) != 0 {
					for _, transaction := range paymentResponse.Entries {
						callback(&transaction)
					}
				}
			}
			if len(paymentResponse.Entries) == 0 {
				return nil
			}
		} else {
			c.error("Invalid response from Twikey: ", res.StatusCode)
			return NewTwikeyErrorFromResponse(res)
		}
	}
}

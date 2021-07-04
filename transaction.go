package twikey

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type TransactionRequest struct {
	DocumentReference             string
	TransactionDate               string
	RequestedCollection           string
	Msg                           string
	Ref                           string
	Amount                        float64
	Place                         string
	ReferenceIsEndToEndIdentifier bool
}

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

type TransactionList struct {
	Entries []Transaction
}

func (c *TwikeyClient) TransactionNew(transaction TransactionRequest) (*Transaction, error) {

	params := url.Values{}
	params.Add("mndtId", transaction.DocumentReference)
	params.Add("date", transaction.TransactionDate)
	params.Add("reqcolldt", transaction.RequestedCollection)
	params.Add("amount", fmt.Sprintf("%.2f", transaction.Amount))
	params.Add("message", transaction.Msg)
	params.Add("ref", transaction.Ref)
	params.Add("place", transaction.Place)
	if transaction.ReferenceIsEndToEndIdentifier {
		params.Add("refase2e", "true")
	}

	c.debug("New transaction", params)

	req, _ := http.NewRequest("POST", c.BaseURL+"/creditor/transaction", strings.NewReader(params.Encode()))
	var transactionList TransactionList
	err := c.sendRequest(req, transactionList)
	if err != nil {
		return nil, err
	}
	return &transactionList.Entries[0], nil
}

func (c *TwikeyClient) TransactionFeed(callback func(transaction Transaction)) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	for {
		req, _ := http.NewRequest("GET", c.BaseURL+"/creditor/transaction", nil)
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
						callback(transaction)
					}
				}
			}
			if len(paymentResponse.Entries) == 0 {
				return nil
			}
		} else {
			c.error("Invalid response from Twikey: ", res.StatusCode)
			return errors.New(res.Status)
		}
	}
}

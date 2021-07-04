package twikey

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type Invoice struct {
	Id         string          `json:"id,omitempty"`
	Number     string          `json:"number"`
	Title      string          `json:"title"`
	Remittence string          `json:"remittence"`
	Ct         int             `json:"ct,omitempty"`
	State      string          `json:"state,omitempty"`
	Amount     float64         `json:"amount"`
	Date       string          `json:"date"`
	Duedate    string          `json:"duedate"`
	Ref        string          `json:"ref,omitempty"`
	Customer   Customer        `json:"customer"`
	Pdf        []byte          `json:"pdf,omitempty"`
	Meta       invoiceFeedMeta `json:"meta,omitempty"`
}

type Customer struct {
	Email     string `json:"email,omitempty"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Address   string `json:"address"`
	City      string `json:"city"`
	Zip       string `json:"zip"`
	Country   string `json:"country"`
	L         string `json:"l"`
	Mobile    string `json:"mobile,omitempty"`
}

type InvoiceFeed struct {
	Invoices []Invoice
}

func (inv *Invoice) IsPaid() bool {
	return inv.State == "PAID"
}

func (inv *Invoice) HasMeta() bool {
	return inv.Meta != invoiceFeedMeta{}
}

// invoices go from pending to booked or expired when payment failed
func (inv *Invoice) IsFailed() bool {
	return inv.State == "BOOKED" || inv.State == "EXPIRED"
}

type invoiceFeedMeta struct {
	LastError string `json:"lastError,omitempty"`
}

func (c *TwikeyClient) InvoiceFromUbl(ctx context.Context, ublBytes []byte, ref string) (*Invoice, error) {

	if err := c.refreshTokenIfRequired(); err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", c.BaseURL+"/creditor/invoice/ubl", bytes.NewReader(ublBytes))
	req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Authorization", c.apiToken) //Already there
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("X-Ref", ref)

	res, _ := c.HTTPClient.Do(req)
	if res.StatusCode == 200 {
		payload, _ := ioutil.ReadAll(res.Body)
		c.debug("TwikeyInvoice: ", string(payload))
		if res.Header["X-Warning"] != nil {
			c.error("Warning for", ref, res.Header["X-Warning"])
		}
		var invoice Invoice
		err := json.Unmarshal(payload, &invoice)
		if err != nil {
			return nil, err
		}
		return &invoice, nil
	}

	errcode := res.Header["Apierror"][0]
	errLoad, _ := ioutil.ReadAll(res.Body)
	c.error("ERROR sending ubl invoice to Twikey: ", string(errLoad))
	return nil, errors.New(errcode)
}

//Get invoice Feed twikey
func (c *TwikeyClient) InvoiceFeed(callback func(invoice Invoice)) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	req, _ := http.NewRequest("GET", c.BaseURL+"/creditor/invoice?include=meta", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", c.apiToken) //Already there
	req.Header.Set("User-Agent", c.UserAgent)

	var feeds InvoiceFeed
	var moreInvoices = true

	for moreInvoices {
		if err := c.sendRequest(req, &feeds); err != nil {
			return err
		}

		for _, invoice := range feeds.Invoices {
			callback(invoice)
		}

		moreInvoices = len(feeds.Invoices) >= 100
	}
	return nil
}

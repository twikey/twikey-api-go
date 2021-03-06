package twikey

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type InvoiceAction int64

const (
	InvoiceAction_EMAIL    InvoiceAction = iota // Send invitation by email
	InvoiceAction_SMS                           // Send invitation by sms
	InvoiceAction_REMINDER                      // Send a reminder by email
	InvoiceAction_LETTER                        // Send the invoice via postal letter
	InvoiceAction_REOFFER                       // Reoffer (or try to collect the invoice via a recurring mechanism)
)

// Invoice is the base object for sending and receiving invoices to Twikey
type Invoice struct {
	Id                 string           `json:"id,omitempty"`
	Number             string           `json:"number"`
	Title              string           `json:"title"`
	Remittance         string           `json:"remittance"`
	Ct                 int              `json:"ct,omitempty"`
	Manual             bool             `json:"manual,omitempty"`
	State              string           `json:"state,omitempty"`
	Amount             float64          `json:"amount"`
	Date               string           `json:"date"`
	Duedate            string           `json:"duedate"`
	Ref                string           `json:"ref,omitempty"`
	CustomerByDocument string           `json:"customerByDocument,omitempty"`
	Customer           *Customer        `json:"customer,omitempty"`
	Pdf                []byte           `json:"pdf,omitempty"`
	Meta               *InvoiceFeedMeta `json:"meta,omitempty"`
	LastPayment        *Lastpayment     `json:"lastpayment,omitempty"`
}

type Lastpayment []map[string]interface{}

type NewInvoiceRequest struct {
	Origin    string
	Reference string
	Purpose   string
	Manual    bool // Don't automatically collect
	Template  string
	Contract  string
	Invoice   *Invoice // either UBL
	UblBytes  []byte   // or an invoice item
}

// Customer is a json wrapper for usage inside the Invoice object
type Customer struct {
	CustomerNumber string `json:"customerNumber,omitempty"`
	Email          string `json:"email,omitempty"`
	CompanyName    string `json:"companyName"`
	Coc            string `json:"coc"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Address        string `json:"address"`
	City           string `json:"city"`
	Zip            string `json:"zip"`
	Country        string `json:"country"`
	Language       string `json:"l"`
	Mobile         string `json:"mobile,omitempty"`
}

func (c *Customer) asUrlParams() string {
	params := url.Values{}
	addIfExists(params, "email", c.Email)
	addIfExists(params, "companyName", c.CompanyName)
	addIfExists(params, "coc", c.Coc)
	addIfExists(params, "firstname", c.FirstName)
	addIfExists(params, "lastname", c.LastName)
	addIfExists(params, "address", c.Address)
	addIfExists(params, "city", c.City)
	addIfExists(params, "zip", c.Zip)
	addIfExists(params, "country", c.Country)
	addIfExists(params, "l", c.Language)
	addIfExists(params, "mobile", c.Mobile)
	return params.Encode()
}

// InvoiceFeed is a struct to contain the response coming from Twikey, should be considered internal
type InvoiceFeed struct {
	Invoices []Invoice
}

// IsPaid convenience method
func (inv *Invoice) IsPaid() bool {
	return inv.State == "PAID"
}

// IsPending convenience method
func (inv *Invoice) IsPending() bool {
	return inv.State == "PENDING"
}

// IsFailed allows to distinguish invoices since they go from pending to booked or expired when payment failed
func (inv *Invoice) IsFailed() bool {
	return inv.State == "BOOKED" || inv.State == "EXPIRED"
}

// HasMeta convenience method to indicate that there is extra info available on the invoice
func (inv *Invoice) HasMeta() bool {
	return inv.Meta != nil
}

type InvoiceFeedMeta struct {
	LastError string `json:"lastError,omitempty"`
}

// InvoiceFromUbl sends an invoice to Twikey in UBL format
func (c *Client) InvoiceAdd(ctx context.Context, invoiceRequest *NewInvoiceRequest) (*Invoice, error) {

	if err := c.refreshTokenIfRequired(); err != nil {
		return nil, err
	}

	var req *http.Request
	if invoiceRequest.Invoice != nil {
		invoiceBytes, err := json.Marshal(invoiceRequest)
		if err != nil {
			return nil, err
		}
		req, _ = http.NewRequest(http.MethodPost, c.BaseURL+"/creditor/invoice", bytes.NewReader(invoiceBytes))
		req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", c.apiToken) //Already there
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.UserAgent)
		// req.Header.Set("X-Ref", invoiceRequest.Reference)  ref already in json
		if invoiceRequest.Origin != "" {
			req.Header.Set("X-PARTNER", invoiceRequest.Origin)
		}
		if invoiceRequest.Purpose != "" {
			req.Header.Set("X-Purpose", invoiceRequest.Purpose)
		}
		if invoiceRequest.Manual {
			req.Header.Set("X-MANUAL", "true")
		}
	} else if len(invoiceRequest.UblBytes) != 0 {
		invoiceUrl := c.BaseURL + "/creditor/invoice/ubl"
		req, _ = http.NewRequest(http.MethodPost, invoiceUrl, bytes.NewReader(invoiceRequest.UblBytes))
		req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/xml")
		req.Header.Set("Authorization", c.apiToken) //Already there
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.UserAgent)
		if invoiceRequest.Template != "" {
			req.Header.Set("X-Template", invoiceRequest.Template)
		}
		if invoiceRequest.Contract != "" {
			req.Header.Set("X-CONTRACT", invoiceRequest.Contract)
		}
		if invoiceRequest.Origin != "" {
			req.Header.Set("X-Partner", invoiceRequest.Origin)
		}
		if invoiceRequest.Purpose != "" {
			req.Header.Set("X-Purpose", invoiceRequest.Purpose)
		}
		if invoiceRequest.Manual {
			req.Header.Set("X-Manual", "true")
		}
		if invoiceRequest.Reference != "" {
			req.Header.Set("X-Ref", invoiceRequest.Reference)
		}
	} else {
		return nil, &TwikeyError{
			Status:  0,
			Code:    "invalid_request",
			Message: "Either ubl or invoice struct is required",
		}
	}

	res, _ := c.HTTPClient.Do(req)
	if res.StatusCode == 200 {
		payload, _ := ioutil.ReadAll(res.Body)
		c.Debug.Println("TwikeyInvoice: ", string(payload))
		if res.Header["X-Warning"] != nil {
			c.Debug.Println("Warning for new invoice with ref=", invoiceRequest.Reference, res.Header["X-Warning"])
		}
		var invoice Invoice
		err := json.Unmarshal(payload, &invoice)
		if err != nil {
			return nil, err
		}
		return &invoice, nil
	}

	errLoad, _ := ioutil.ReadAll(res.Body)
	c.Debug.Println("Error sending invoice to Twikey: ", string(errLoad))
	return nil, NewTwikeyErrorFromResponse(res)
}

//InvoiceFeed Get invoice Feed twikey
func (c *Client) InvoiceFeed(ctx context.Context, callback func(invoice *Invoice), sideloads ...string) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	_url := c.BaseURL + "/creditor/invoice"
	for i, sideload := range sideloads {
		if i == 0 {
			_url = _url + "?include=" + sideload
		} else {
			_url = _url + "&include=" + sideload
		}
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, _url, nil)
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
			callback(&invoice)
		}

		moreInvoices = len(feeds.Invoices) >= 100
	}
	return nil
}

// InvoiceDetail allows a snapshot of a particular invoice, note that this is rate limited
func (c *Client) InvoiceDetail(ctx context.Context, invoiceIdOrNumber string, sideloads ...string) (*Invoice, error) {

	if err := c.refreshTokenIfRequired(); err != nil {
		return nil, err
	}

	_url := c.BaseURL + "/creditor/invoice/" + invoiceIdOrNumber
	for i, sideload := range sideloads {
		if i == 0 {
			_url = _url + "?include=" + sideload
		} else {
			_url = _url + "&include=" + sideload
		}
	}

	req, _ := http.NewRequest(http.MethodGet, _url, nil)
	req.WithContext(ctx)
	req.Header.Add("Accept-Language", "en")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Add("Authorization", c.apiToken)

	res, _ := c.HTTPClient.Do(req)
	if res.StatusCode == 200 {
		payload, _ := ioutil.ReadAll(res.Body)

		var invoice Invoice
		err := json.Unmarshal(payload, &invoice)
		if err != nil {
			return nil, err
		}

		return &invoice, nil
	}
	return nil, NewTwikeyErrorFromResponse(res)
}

// InvoiceAction allows certain actions to be done on an existing invoice
func (c *Client) InvoiceAction(ctx context.Context, invoiceIdOrNumber string, action InvoiceAction) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	_url := c.BaseURL + "/creditor/invoice/" + invoiceIdOrNumber + "/action"
	params := url.Values{}

	switch action {
	case InvoiceAction_EMAIL:
		params.Add("type", "email")
	case InvoiceAction_SMS:
		params.Add("type", "sms")
	case InvoiceAction_LETTER:
		params.Add("type", "letter")
	case InvoiceAction_REMINDER:
		params.Add("type", "reminder")
	case InvoiceAction_REOFFER:
		params.Add("type", "reoffer")
	default:
		return errors.New("Invalid action")
	}

	req, _ := http.NewRequest(http.MethodPost, _url, strings.NewReader(params.Encode()))
	req.WithContext(ctx)
	req.Header.Add("Accept-Language", "en")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Add("Authorization", c.apiToken)

	res, _ := c.HTTPClient.Do(req)
	if res.StatusCode == 204 {
		return nil
	}
	return NewTwikeyErrorFromResponse(res)
}

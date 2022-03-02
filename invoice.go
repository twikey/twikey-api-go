package twikey

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
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

// HasMeta convenience method to indicate that there is extra info available on the invoice
func (inv *Invoice) HasMeta() bool {
	return inv.Meta != nil
}

// IsFailed allows to distinguish invoices since they go from pending to booked or expired when payment failed
func (inv *Invoice) IsFailed() bool {
	return inv.State == "BOOKED" || inv.State == "EXPIRED"
}

type InvoiceFeedMeta struct {
	LastError string `json:"lastError,omitempty"`
}

// InvoiceFromUbl sends an invoice to Twikey in UBL format
func (c *Client) InvoiceAdd(ctx context.Context, invoice *Invoice) (*Invoice, error) {

	if err := c.refreshTokenIfRequired(); err != nil {
		return nil, err
	}

	invoiceBytes, err := json.Marshal(invoice)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", c.BaseURL+"/creditor/invoice", bytes.NewReader(invoiceBytes))
	req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiToken) //Already there
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	res, _ := c.HTTPClient.Do(req)
	if res.StatusCode == 200 {
		payload, _ := ioutil.ReadAll(res.Body)
		c.debug("TwikeyInvoice: ", string(payload))
		if res.Header["X-Warning"] != nil {
			c.error("Warning for", invoice.Number, res.Header["X-Warning"])
		}
		var invoice Invoice
		err := json.Unmarshal(payload, &invoice)
		if err != nil {
			return nil, err
		}
		return &invoice, nil
	}

	errLoad, _ := ioutil.ReadAll(res.Body)
	c.error("ERROR sending ubl invoice to Twikey: ", string(errLoad))
	return nil, NewTwikeyErrorFromResponse(res)
}

// InvoiceFromUbl sends an invoice to Twikey in UBL format
func (c *Client) InvoiceFromUbl(ctx context.Context, ublBytes []byte, ref string, noAutoCollection bool) (*Invoice, error) {

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
	if noAutoCollection {
		req.Header.Set("X-MANUAL", "true")
	}

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

	errLoad, _ := ioutil.ReadAll(res.Body)
	c.error("ERROR sending ubl invoice to Twikey: ", string(errLoad))
	return nil, NewTwikeyErrorFromResponse(res)
}

//InvoiceFeed Get invoice Feed twikey
func (c *Client) InvoiceFeed(callback func(invoice *Invoice), sideloads ...string) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	url := c.BaseURL + "/creditor/invoice"
	for i, sideload := range sideloads {
		if i == 0 {
			url = url + "?include=" + sideload
		} else {
			url = url + "&include=" + sideload
		}
	}

	req, _ := http.NewRequest("GET", url, nil)
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

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

type PaylinkRequest struct {
	CustomerNumber string  //	The customer number (strongly advised)
	Email          string  //	Email of the debtor	(Required to send invite)
	Lastname       string  //	lastname
	Firstname      string  //	firstname
	Language       string  //	Language (en/fr/nl/de/pt/es/it)
	Mobile         string  //	mobile number
	Ct             int64   //	contract template
	Title          string  //	Message to the debto
	Remittance     string  //	Payment message, if empty then title will be used
	Amount         float64 //	Amount to be billed
	RedirectUrl    string  //	Optional redirect after pay url (must use http(s)://)
	Place          string  //	Optional place
	Expiry         string  //	Optional expiration date
	SendInvite     string  //	Send out invite email or sms directly (email, sms)
	Address        string  //	Address (street + number)
	City           string  //	City of debtor
	Zip            string  //	Zipcode of debtor
	Country        string  //	ISO format (2 letters)
	Txref          string  //	References from existing transactions
	Method         string  //	Circumvents the payment selection with PSP (bancontact/ideal/maestro/mastercard/visa/inghomepay/kbc/belfius)
	Invoice        string  //	create payment link for specific invoice number
}

type Paylink struct {
	Id     int64   `json:"id,omitempty"`
	Amount float64 `json:"amount,omitempty"`
	Msg    string  `json:"msg,omitempty"`
	Ref    string  `json:"ref,omitempty"`
	State  string  `json:"state,omitempty"`
}

func (c *TwikeyClient) PaylinkNew(paylinkRequest PaylinkRequest) (*Paylink, error) {

	params := url.Values{}
	params.Add("ct", fmt.Sprintf("%d", paylinkRequest.Ct))
	params.Add("title", paylinkRequest.Title)
	params.Add("remittance", paylinkRequest.Remittance)
	params.Add("amount", fmt.Sprintf("%.2f", paylinkRequest.Amount))
	params.Add("redirectUrl", paylinkRequest.RedirectUrl)
	params.Add("place", paylinkRequest.Place)
	params.Add("expiry", paylinkRequest.Expiry)
	params.Add("sendInvite", paylinkRequest.SendInvite)
	params.Add("txref", paylinkRequest.Txref)
	params.Add("method", paylinkRequest.Method)
	params.Add("invoice", paylinkRequest.Invoice)

	params.Add("customerNumber", paylinkRequest.CustomerNumber)
	params.Add("email", paylinkRequest.Email)
	params.Add("lastname", paylinkRequest.Lastname)
	params.Add("firstname", paylinkRequest.Firstname)
	params.Add("l", paylinkRequest.Language)
	params.Add("mobile", paylinkRequest.Mobile)
	params.Add("address", paylinkRequest.Address)
	params.Add("city", paylinkRequest.City)
	params.Add("zip", paylinkRequest.Zip)
	params.Add("country", paylinkRequest.Country)

	req, _ := http.NewRequest("POST", c.BaseURL+"/creditor/payment/link", strings.NewReader(params.Encode()))
	var paylink Paylink
	err := c.sendRequest(req, paylink)
	if err != nil {
		return nil, err
	}
	return &paylink, nil
}

func (c *TwikeyClient) PaylinkFeed(callback func(paylink Paylink)) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	req, _ := http.NewRequest("GET", c.BaseURL+"/creditor/payment/link/feed", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", c.apiToken)
	req.Header.Set("User-Agent", c.UserAgent)

	res, _ := c.HTTPClient.Do(req)

	if res.StatusCode == 200 {
		payload, _ := ioutil.ReadAll(res.Body)
		var paylinks []Paylink
		err := json.Unmarshal(payload, &paylinks)
		if err == nil {
			for len(paylinks) != 0 {
				for _, paylink := range paylinks {
					callback(paylink)
				}
			}
		}
		_ = res.Body.Close()
		return nil
	} else {
		c.error("Invalid response from Twikey: ", res.StatusCode)
		return errors.New(res.Status)
	}
}

package twikey

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// InviteRequest contains all possible parameters that can be send to invite a customer
// to sign a document
type InviteRequest struct {
	Template       string // mandatory
	CustomerNumber string
	Email          string
	Mobile         string
	Language       string
	Lastname       string
	Firstname      string
	MandateNumber  string
	ContractNumber string
	CompanyName    string
	Coc            string
	Address        string
	City           string
	Zip            string
	Country        string
	SignDate       string
	Amount         string
	Iban           string
	Bic            string
	Campaign       string
	Method         string
	extra          map[string]string
}

func (request *InviteRequest) asUrlParams() string {
	params := url.Values{}
	params.Add("ct", request.Template)
	params.Add("customerNumber", request.CustomerNumber)
	params.Add("email", request.Email)
	params.Add("mobile", request.Mobile)
	params.Add("l", request.Language)
	params.Add("lastname", request.Lastname)
	params.Add("firstname", request.Firstname)
	params.Add("mandateNumber", request.MandateNumber)
	params.Add("contractNumber", request.ContractNumber)
	params.Add("companyName", request.CompanyName)
	params.Add("coc", request.Coc)
	params.Add("address", request.Address)
	params.Add("city", request.City)
	params.Add("zip", request.Zip)
	params.Add("country", request.Country)
	params.Add("overrideFromDate", request.SignDate)
	params.Add("amount", request.Amount)
	params.Add("iban", request.Iban)
	params.Add("bic", request.Bic)
	params.Add("campaign", request.Campaign)
	params.Add("method", request.Method)

	if request.extra != nil {
		for k, v := range request.extra {
			params.Add(k, v)
		}
	}
	return params.Encode()
}

// Invite is the response containing the documentNumber, key and the url to point the customer too.
type Invite struct {
	MndtId string // documentNumber
	Url    string // where the customer can sign the document
	Key    string // specific invite key
}

// UpdateRequest contains all possible parameters that can be send to update a document
type UpdateRequest struct {
	MandateNumber  string // Document or MandateNumber
	State          string // active or passive (activated or suspend mandate)
	Mobile         string // Owner's mobile number
	Iban           string // Debtor's IBAN
	Bic            string // Debtor's BIC code
	Email          string // email address of debtor
	Firstname      string // Firstname of the debtor
	Lastname       string // Lastname of the debtor
	CompanyName    string // Company name on the mandate
	Vatno          string // The enterprise number (can only be changed if companyName is changed)
	CustomerNumber string // The customer number (can be added, updated or used to move a mandate)
	Language       string // language on the mandate (ISO 2 letters)
	Address        string // Address (street + number)
	City           string // City of debtor
	Zip            string // Zipcode of debtor
	Country        string // Country of debtor
}

// CtctDtls contains all contact details for a specific document
type CtctDtls struct {
	EmailAdr string
	MobNb    string
	Othr     string
}

// PstlAdr contains address data for a specific document
type PstlAdr struct {
	AdrLine string
	PstCd   string
	TwnNm   string
	Ctry    string
}

// Prty contains party details for a specific document
type Prty struct {
	Nm       string
	PstlAdr  PstlAdr
	Id       string
	CtctDtls CtctDtls
}

// KeyValue key value pairs of extra data in a document
type KeyValue struct {
	Key   string
	Value interface{}
}

type FinInstnId struct {
	BICFI string
}

type DbtrAgt struct {
	FinInstnId FinInstnId
}

type Mndt struct {
	MndtId      string
	Dbtr        Prty
	DbtrAcct    string
	DbtrAgt     DbtrAgt
	RfrdDoc     string
	SplmtryData []KeyValue
}

type MndtDetail struct {
	Mndt Mndt
}

// AmdmntRsn contains the reason why something was updated
type AmdmntRsn struct {
	Rsn string
}

// CxlRsn contains the reason why something was cancelled
type CxlRsn struct {
	Rsn string
}

// MandateUpdate contains all info regarding a new/update or cancelled document
type MandateUpdate struct {
	Mndt        *Mndt
	AmdmntRsn   *AmdmntRsn `json:",omitempty"`
	CxlRsn      *CxlRsn    `json:",omitempty"`
	OrgnlMndtId string
	EvtTime     string
}

// MandateUpdates is a struct to contain the response coming from Twikey, should be considered internal
type MandateUpdates struct {
	Messages []MandateUpdate
}

// DocumentInvite allows to invite a customer to sign a specific document
func (c *Client) DocumentInvite(request *InviteRequest) (*Invite, error) {

	if err := c.refreshTokenIfRequired(); err != nil {
		return nil, err
	}

	if request.Template == "" {
		return nil, errors.New("A template is required")
	}

	params := request.asUrlParams()
	c.debug("New document", params)
	req, _ := http.NewRequest("POST", c.BaseURL+"/creditor/invite", strings.NewReader(params))
	var invite Invite
	if err := c.sendRequest(req, &invite); err != nil {
		return nil, err
	}
	return &invite, nil
}

// DocumentInvite allows to invite a customer to sign a specific document
func (c *Client) DocumentSign(request *InviteRequest) (*Invite, error) {

	if err := c.refreshTokenIfRequired(); err != nil {
		return nil, err
	}

	if request.Template == "" {
		return nil, errors.New("A template is required")
	}

	params := request.asUrlParams()
	c.debug("New document", params)
	req, _ := http.NewRequest("POST", c.BaseURL+"/creditor/sign", strings.NewReader(params))
	var invite Invite
	if err := c.sendRequest(req, &invite); err != nil {
		return nil, err
	}
	return &invite, nil
}

// DocumentUpdate allows to update a previously added document
func (c *Client) DocumentUpdate(request *UpdateRequest) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	params := url.Values{}
	if request.MandateNumber == "" {
		return NewTwikeyError("A mndtId is required")
	}

	params.Add("mndtId", request.MandateNumber)
	params.Add("state", request.State)
	params.Add("mobile", request.Mobile)
	params.Add("iban", request.Iban)
	params.Add("bic", request.Bic)
	params.Add("email", request.Email)
	params.Add("firstname", request.Firstname)
	params.Add("lastname", request.Lastname)
	params.Add("companyName", request.CompanyName)
	params.Add("vatno", request.Vatno)
	params.Add("customerNumber", request.CustomerNumber)
	params.Add("l", request.Language)
	params.Add("address", request.Address)
	params.Add("city", request.City)
	params.Add("zip", request.Zip)
	params.Add("country", request.Country)

	c.debug("Update document", request.MandateNumber, params)

	req, _ := http.NewRequest("POST", c.BaseURL+"/creditor/mandate/update", strings.NewReader(params.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", c.apiToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	err := c.sendRequest(req, nil)
	if err != nil {
		return err
	}
	return nil
}

// DocumentCancel allows to cancel (or delete if unsigned) a previously added document
func (c *Client) DocumentCancel(mandate string, reason string) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	params := url.Values{}
	params.Add("mndtId", mandate)
	params.Add("rsn", reason)

	c.debug("Cancelled document", mandate, reason)

	req, _ := http.NewRequest("DELETE", c.BaseURL+"/creditor/mandate?"+params.Encode(), nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", c.apiToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	err := c.sendRequest(req, nil)
	return err
}

// DocumentFeed retrieves all documents since the last call with callbacks since there may be many
func (c *Client) DocumentFeed(
	newDocument func(mandate *Mndt),
	updateDocument func(mandate *Mndt, reason *AmdmntRsn),
	cancelledDocument func(mandate string, reason *CxlRsn)) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	for {
		req, _ := http.NewRequest("GET", c.BaseURL+"/creditor/mandate", nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", c.apiToken)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.UserAgent)

		res, _ := c.HTTPClient.Do(req)
		if res.StatusCode == 200 {
			payload, _ := ioutil.ReadAll(res.Body)
			var updates MandateUpdates
			err := json.Unmarshal(payload, &updates)
			if err != nil {
				return err
			}

			defer res.Body.Close()
			c.debug(fmt.Sprintf("Fetched %d documents", len(updates.Messages)))
			for _, update := range updates.Messages {
				if update.CxlRsn != nil {
					cancelledDocument(update.OrgnlMndtId, update.CxlRsn)
				} else if update.AmdmntRsn != nil {
					updateDocument(update.Mndt, update.AmdmntRsn)
				} else {
					newDocument(update.Mndt)
				}
			}

			if len(updates.Messages) == 0 {
				return nil
			}
		} else {
			return NewTwikeyErrorFromResponse(res)
		}
	}
}

// DownloadPdf allows the download of a specific (signed) pdf
func (c *Client) DownloadPdf(mndtId string, downloadFile string) error {
	params := url.Values{}
	params.Add("mndtId", mndtId)

	req, _ := http.NewRequest("GET", c.BaseURL+"/creditor/mandate/pdf?"+params.Encode(), nil)
	req.Header.Add("Accept-Language", "en")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Add("Authorization", c.apiToken)

	absPath, _ := filepath.Abs(downloadFile)
	res, _ := c.HTTPClient.Do(req)
	if res.StatusCode == 200 {
		payload, _ := ioutil.ReadAll(res.Body)

		f, _ := os.Create(downloadFile)
		defer f.Close()
		_, err := f.Write(payload)
		if err != nil {
			fmt.Println("Unable to download file:", absPath, err)
		} else {
			fmt.Println("Saving to file:", absPath)
		}
		return err
	}
	fmt.Println("Unable to download file:", absPath)
	return NewTwikeyErrorFromResponse(res)
}

// DocumentDetail allows a snapshot of a particular mandate, note that this is rate limited
func (c *Client) DocumentDetail(mndtId string) (*Mndt, error) {

	if err := c.refreshTokenIfRequired(); err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("mndtId", mndtId)

	req, _ := http.NewRequest("GET", c.BaseURL+"/creditor/mandate/detail?"+params.Encode(), nil)
	req.Header.Add("Accept-Language", "en")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Add("Authorization", c.apiToken)

	res, _ := c.HTTPClient.Do(req)
	if res.StatusCode == 200 {
		payload, _ := ioutil.ReadAll(res.Body)

		var mndt MndtDetail
		err := json.Unmarshal(payload, &mndt)
		if err != nil {
			return nil, err
		}

		return &mndt.Mndt, nil
	}
	return nil, NewTwikeyErrorFromResponse(res)
}

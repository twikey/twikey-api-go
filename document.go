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

type InviteRequest struct {
	ct               string // mandatory
	customerNumber   string
	email            string
	mobile           string
	l                string
	lastname         string
	firstname        string
	mandateNumber    string
	contractNumber   string
	companyName      string
	coc              string
	address          string
	city             string
	zip              string
	country          string
	overrideFromDate string
	amount           string
	iban             string
	bic              string
	campaign         string
}

type Invite struct {
	Url string
	Key string
}

type UpdateRequest struct {
	mndtId         string // mandateNumber
	state          string //active or passive (activated or suspend mandate)
	mobile         string //	Owner's mobile number
	iban           string //	Debtor's IBAN
	bic            string //	Debtor's BIC code
	email          string //	email address of debtor
	firstname      string //	Firstname of the debtor
	lastname       string //	Lastname of the debtor
	companyName    string //	Company name on the mandate
	vatno          string //	The enterprise number (can only be changed if companyName is changed)
	customerNumber string //	The customer number (can be added, updated or used to move a mandate)
	l              string //	language on the mandate (ISO 2 letters)
	address        string //	Address (street + number)
	city           string //	City of debtor
	zip            string //	Zipcode of debtor
	country        string //	Country of debtor
}

type CtctDtls struct {
	EmailAdr string
	MobNb    string
	Othr     string
}

type PstlAdr struct {
	AdrLine string
	PstCd   string
	TwnNm   string
	Ctry    string
}

type Prty struct {
	Nm       string
	PstlAdr  PstlAdr
	Id       string
	CtctDtls CtctDtls
}

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
type AmdmntRsn struct {
	Rsn string
}
type CxlRsn struct {
	Rsn string
}

type MandateUpdate struct {
	Mndt        *Mndt
	AmdmntRsn   *AmdmntRsn `json:",omitempty"`
	CxlRsn      *CxlRsn    `json:",omitempty"`
	OrgnlMndtId string
	EvtTime     string
}

type MandateUpdates struct {
	Messages []MandateUpdate
}

func (c *TwikeyClient) DocumentInvite(request InviteRequest) (*Invite, error) {

	if err := c.refreshTokenIfRequired(); err != nil {
		return nil, err
	}

	if request.ct == "" {
		return nil, errors.New("A template is required")
	}
	params := url.Values{}
	params.Add("ct", request.ct)
	params.Add("customerNumber", request.customerNumber)
	params.Add("email", request.email)
	params.Add("mobile", request.mobile)
	params.Add("l", request.l)
	params.Add("lastname", request.lastname)
	params.Add("firstname", request.firstname)
	params.Add("mandateNumber", request.mandateNumber)
	params.Add("contractNumber", request.contractNumber)
	params.Add("companyName", request.companyName)
	params.Add("coc", request.coc)
	params.Add("address", request.address)
	params.Add("city", request.city)
	params.Add("zip", request.zip)
	params.Add("country", request.country)
	params.Add("overrideFromDate", request.overrideFromDate)
	params.Add("amount", request.amount)
	params.Add("iban", request.iban)
	params.Add("bic", request.bic)
	params.Add("campaign", request.campaign)

	c.debug("New document", params.Encode())

	req, _ := http.NewRequest("POST", c.BaseURL+"/creditor/invite", strings.NewReader(params.Encode()))
	var invite Invite
	if err := c.sendRequest(req, &invite); err != nil {
		return nil, err
	}
	return &invite, nil
}

func (c *TwikeyClient) CocumentUpdate(request UpdateRequest) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	params := url.Values{}
	if request.mndtId == "" {
		return errors.New("A mndtId is required")
	}

	params.Add("mndtId", request.mndtId)
	params.Add("state", request.state)
	params.Add("mobile", request.mobile)
	params.Add("iban", request.iban)
	params.Add("bic", request.bic)
	params.Add("email", request.email)
	params.Add("firstname", request.firstname)
	params.Add("lastname", request.lastname)
	params.Add("companyName", request.companyName)
	params.Add("vatno", request.vatno)
	params.Add("customerNumber", request.customerNumber)
	params.Add("l", request.l)
	params.Add("address", request.address)
	params.Add("city", request.city)
	params.Add("zip", request.zip)
	params.Add("country", request.country)

	c.debug("Update document", request.mndtId, params)

	req, _ := http.NewRequest("POST", c.BaseURL+"/creditor/mandate/update", strings.NewReader(params.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", c.apiToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	_, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *TwikeyClient) DocumentCancel(mandate string, reason string) error {

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

	res, _ := c.HTTPClient.Do(req)
	if res.StatusCode > 300 {
		err := res.Header["Apierror"][0]
		c.error("Invalid response from Twikey:", err, res.StatusCode)
		return errors.New(err)
	}
	return nil
}

func (c *TwikeyClient) DocumentFeed(
	newDocument func(mandate Mndt),
	updateDocument func(mandate Mndt, reason AmdmntRsn),
	cancelledDocument func(mandate string, reason CxlRsn)) error {

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

			res.Body.Close()
			c.debug(fmt.Sprintf("Fetched %d documents", len(updates.Messages)))
			for _, update := range updates.Messages {
				if update.CxlRsn != nil {
					cancelledDocument(update.OrgnlMndtId, *update.CxlRsn)
				} else if update.AmdmntRsn != nil {
					updateDocument(*update.Mndt, *update.AmdmntRsn)
				} else {
					newDocument(*update.Mndt)
				}
			}

			if len(updates.Messages) == 0 {
				return nil
			}
		} else {
			return errors.New(res.Status)
		}
	}
}

func (c *TwikeyClient) DownloadPdf(mndtId string, downloadFile string) error {
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
	} else {
		fmt.Println("Unable to download file:", absPath)
	}
	return errors.New(res.Status)
}

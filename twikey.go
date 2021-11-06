// Package twikey provides the bindings for Twikey REST APIs.
package twikey

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	baseURLV1       = "https://api.twikey.com"
	twikeyBaseAgent = "twikey-api/go-v0.1.1"
)

// Client is the base class, please use a dedicated UserAgent so we can notify the emergency contact
// if weird behaviour is perceived.
type Client struct {
	BaseURL    string
	APIKey     string
	PrivateKey string
	Salt       string
	UserAgent  string

	HTTPClient *http.Client

	Debug *log.Logger

	apiToken  string
	lastLogin time.Time
}

// NewClient is a convenience method to hit the ground running with the Twikey Rest API
func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL:   baseURLV1,
		APIKey:    apiKey,
		Salt:      "own",
		UserAgent: twikeyBaseAgent,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"` // translated according to Accept-Language
	Extra   string `json:"extra"`
}

func (c *Client) debug(v ...interface{}) {
	if c.Debug != nil {
		c.Debug.Println(v...)
	}
}

func (c *Client) error(v ...interface{}) {
	if c.Debug != nil {
		c.Debug.Fatal(v...)
	}
}

func (c *Client) sendRequest(req *http.Request, v interface{}) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.apiToken)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	//body,_ := ioutil.ReadAll(res.Body)
	//print(body)

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}
		return fmt.Errorf("Unknown error, status code: %d", res.StatusCode)
	}

	if v == nil {
		return nil
	}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return err
	}

	return nil
}

// VerifyWebhook allows the verification of incoming webhooks.
func (c *Client) VerifyWebhook(signatureHeader string, payload string) error {
	hash := hmac.New(sha256.New, []byte(c.APIKey))
	if _, err := hash.Write([]byte(payload)); err != nil {
		c.error("Cannot compute the HMAC for request: ", err)
		return err
	}

	expectedHash := strings.ToUpper(hex.EncodeToString(hash.Sum(nil)))
	if signatureHeader == expectedHash {
		return nil
	}
	return errors.New("Invalid value")
}

func addIfExists(params url.Values, paramKey string, value string) {
	if value != "" {
		params.Add(paramKey, value)
	}
}

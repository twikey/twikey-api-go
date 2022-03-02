// Package twikey provides the bindings for Twikey REST APIs.
package twikey

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
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

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is the base class, please use a dedicated UserAgent so we can notify the emergency contact
// if weird behaviour is perceived.
type Client struct {
	BaseURL    string
	APIKey     string
	PrivateKey string
	Salt       string
	UserAgent  string
	HTTPClient HTTPClient
	Debug      *log.Logger
	apiToken   string
	lastLogin  time.Time
}

// NewClient is a convenience method to hit the ground running with the Twikey Rest API
func NewClient(apiKey string) *Client {
	logger := log.Default()
	logger.SetOutput(ioutil.Discard)
	return &Client{
		BaseURL:   baseURLV1,
		APIKey:    apiKey,
		Salt:      "own",
		UserAgent: twikeyBaseAgent,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
		Debug: logger,
	}
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"` // translated according to Accept-Language
	Extra   string `json:"extra"`
}

// Ping Try the current credentials
func (c *Client) Ping() error {
	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}
	return nil
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
		c.Debug.Println("Error while connecting", err)
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return NewTwikeyError(errRes.Message)
		}
		return NewTwikeyErrorFromResponse(res)
	}

	if v == nil {
		return nil
	}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return NewTwikeyError(err.Error())
	}

	return nil
}

// VerifyWebhook allows the verification of incoming webhooks.
func (c *Client) VerifyWebhook(signatureHeader string, payload string) error {
	hash := hmac.New(sha256.New, []byte(c.APIKey))
	if _, err := hash.Write([]byte(payload)); err != nil {
		c.Debug.Println("Error cannot compute the HMAC for request: ", err)
		return err
	}

	expectedHash := strings.ToUpper(hex.EncodeToString(hash.Sum(nil)))
	if signatureHeader == expectedHash {
		return nil
	}
	return NewTwikeyError("Invalid value")
}

func addIfExists(params url.Values, paramKey string, value string) {
	if value != "" {
		params.Add(paramKey, value)
	}
}

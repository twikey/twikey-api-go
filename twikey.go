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

type TimeProvider interface {
	Now() time.Time
}

type DefaultTimeProvider struct{}

func (tp DefaultTimeProvider) Now() time.Time {
	return time.Now()
}

// Client is the base class, please use a dedicated UserAgent so we can notify the emergency contact
// if weird behaviour is perceived.
type Client struct {
	BaseURL      string
	APIKey       string
	PrivateKey   string
	Salt         string
	UserAgent    string
	HTTPClient   HTTPClient
	Debug        *log.Logger
	apiToken     string
	lastLogin    time.Time
	timeProvider TimeProvider
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
		Debug:        logger,
		timeProvider: DefaultTimeProvider{},
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

	c.Debug.Println("Calling", req.Method, req.URL)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Debug.Println("Error while connecting", err)
		return err
	}

	payload, _ := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()

	c.Debug.Println("Response for", req.Method, req.URL, "was", string(payload))

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		if res.Header.Get("Apierror") == "err_no_login" {
			c.Debug.Println("Error while using apitoken, renewing")
			c.lastLogin = time.Time{} // force re-authenticate
		}
		var errRes errorResponse
		if err = json.Unmarshal(payload, &errRes); err == nil {
			return NewTwikeyError(errRes.Code, errRes.Message, errRes.Extra)
		}
		return NewTwikeyErrorFromResponse(res)
	}

	if v == nil {
		return nil
	}

	if err = json.Unmarshal(payload, v); err != nil {
		return NewTwikeyError("system_error", err.Error(), "")
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
	return NewTwikeyError("invalid_params", "Invalid value", "")
}

func addIfExists(params url.Values, paramKey string, value string) {
	if value != "" {
		params.Add(paramKey, value)
	}
}

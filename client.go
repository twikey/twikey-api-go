package twikey

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	BaseURLV1      = "https://api.twikey.com"
	TWIKEY_VERSION = "twikey-api/go"
)

type TwikeyClient struct {
	BaseURL    string
	ApiKey     string
	PrivateKey string
	Salt       string
	UserAgent  string

	HTTPClient *http.Client

	Debug *log.Logger

	apiToken  string
	lastLogin time.Time
}

func NewClient(apiKey string) *TwikeyClient {
	return &TwikeyClient{
		BaseURL: BaseURLV1,
		ApiKey:  apiKey,
		Salt:    "own",
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

func (c *TwikeyClient) debug(v ...interface{}) {
	if c.Debug != nil {
		c.Debug.Println(v...)
	}
}

func (c *TwikeyClient) error(v ...interface{}) {
	if c.Debug != nil {
		c.Debug.Fatal(v...)
	}
}

func (c *TwikeyClient) sendRequest(req *http.Request, v interface{}) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.apiToken)

	c.debug(req.Body)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}
		return fmt.Errorf("Unknown error, status code: %d", res.StatusCode)
	}

	if err = json.NewDecoder(res.Body).Decode(v); err != nil {
		return err
	}

	return nil
}

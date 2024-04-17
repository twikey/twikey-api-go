package twikey

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Refund is the response receiving from Twikey upon a request
type Refund struct {
	Id     string    `json:"id"`
	Seq    int64     `json:"seq"`
	Iban   string    `json:"iban"`
	Bic    string    `json:"bic"`
	Amount float64   `json:"amount"`
	Msg    string    `json:"msg"`
	Place  string    `json:"place"`
	Ref    string    `json:"ref"`
	Date   string    `json:"date"`
	State  string    `json:"state"`
	Bkdate time.Time `json:"bkdate"`
}

type RefundList struct {
	Entries []Refund
}

// RefundFeed retrieves the feed of updated refunds since last call
func (c *Client) RefundFeed(ctx context.Context, callback func(refund *Refund), options ...FeedOption) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	feedOptions := parseFeedOptions(options)
	_url := c.BaseURL + "/creditor/transfer"
	for i, sideload := range feedOptions.includes {
		if i == 0 {
			_url = _url + "?include=" + sideload
		} else {
			_url = _url + "&include=" + sideload
		}
	}

	for {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, _url, nil)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", c.apiToken)
		req.Header.Set("User-Agent", c.UserAgent)
		if feedOptions.start != -1 {
			req.Header.Set("X-RESUME-AFTER", fmt.Sprintf("%d", feedOptions.start))
			feedOptions.start = -1
		}

		res, err := c.HTTPClient.Do(req)
		if err != nil {
			return err
		}
		if res.StatusCode == 200 {
			payload, _ := io.ReadAll(res.Body)
			_ = res.Body.Close()
			var refunds RefundList
			err := json.Unmarshal(payload, &refunds)
			if err == nil {
				c.Debug.Debugf("Fetched %d refunds", len(refunds.Entries))
				for _, refund := range refunds.Entries {
					callback(&refund)
				}
			} else {
				return err
			}
			if len(refunds.Entries) == 0 {
				return nil
			}
		} else {
			c.Debug.Debugf("Error response from Twikey: %d", res.StatusCode)
			return NewTwikeyErrorFromResponse(res)
		}
	}
}

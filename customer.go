package twikey

import (
	"errors"
	"net/http"
)

func (c *Client) CustomerUpdate(request *Customer) error {

	if err := c.refreshTokenIfRequired(); err != nil {
		return err
	}

	if request.CustomerNumber == "" {
		return errors.New("A customerNumber is required")
	}

	params := request.asUrlParams()
	c.Debug.Println("Update customer", params)

	req, _ := http.NewRequest("PATCH", c.BaseURL+"/customer/"+request.CustomerNumber+"?"+params, nil)
	if err := c.sendRequest(req, nil); err != nil {
		return err
	}
	return nil
}

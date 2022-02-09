package twikey

import "net/http"

type TwikeyError struct {
	code    int
	message string
}

func (err *TwikeyError) Error() string {
	return err.message
}

func (err *TwikeyError) IsUserError() bool {
	return err.code == 400
}

func NewTwikeyError(msg string) *TwikeyError {
	return &TwikeyError{
		code:    400,
		message: msg,
	}
}

func NewTwikeyErrorFromResponse(res *http.Response) *TwikeyError {
	if res.StatusCode == 400 {
		errcode := res.Header["Apierror"][0]
		return &TwikeyError{
			code:    res.StatusCode,
			message: errcode,
		}
	}
	return &TwikeyError{
		code:    res.StatusCode,
		message: res.Status,
	}
}

var SystemError error = &TwikeyError{
	code: 500,
}

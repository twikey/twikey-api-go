package twikey

import "net/http"

type TwikeyError struct {
	status  int
	code    string
	message string
	extra   string
}

func (err *TwikeyError) Error() string {
	return err.message
}

func (err *TwikeyError) IsUserError() bool {
	return err.status == 400
}

func NewTwikeyError(code string, msg string, extra string) *TwikeyError {
	return &TwikeyError{
		status:  400,
		code:    code,
		message: msg,
		extra:   extra,
	}
}

func NewTwikeyErrorFromResponse(res *http.Response) *TwikeyError {
	if res.StatusCode == 400 {
		errcode := res.Header["Apierror"][0]
		return &TwikeyError{
			status:  res.StatusCode,
			code:    errcode,
			message: errcode,
		}
	}
	return &TwikeyError{
		status:  res.StatusCode,
		code:    "system_error",
		message: res.Status,
	}
}

var SystemError error = &TwikeyError{
	status: 500,
	code:   "system_error",
}

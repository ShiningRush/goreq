package goreq

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func NewHttpCodeErr(expectedCodes []int, resp *http.Response) *HttpCodeErr {
	httpCodeErr := &HttpCodeErr{
		ExpectedStatusCodes: expectedCodes,
		StatusCode:          resp.StatusCode,
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		httpCodeErr.ReadBodyErr = err
		return httpCodeErr
	}
	httpCodeErr.Body = body
	return httpCodeErr
}

type HttpCodeErr struct {
	StatusCode          int
	ExpectedStatusCodes []int

	ReadBodyErr error
	Body        []byte
}

func (e *HttpCodeErr) Error() string {
	if e.ReadBodyErr != nil {
		return fmt.Sprintf("http code[%d] is not expected(%v) and read body failed: %s",
			e.StatusCode,
			e.ExpectedStatusCodes,
			e.ReadBodyErr)
	}

	return fmt.Sprintf("http code[%d] is not expected(%v), body: %s", e.StatusCode, e.ExpectedStatusCodes, e.Body)
}

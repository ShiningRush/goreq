package goreq

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type HttpCodeErr struct {
	Resp *http.Response
}

func (e *HttpCodeErr) Error() string {
	body, err := ioutil.ReadAll(e.Resp.Body)
	if err != nil {
		return fmt.Sprintf("http code is not ok, code:%d, body: %s", e.Resp.StatusCode, err)
	}

	return fmt.Sprintf("http code is not ok, code:%d, body: %s", e.Resp.StatusCode, body)
}

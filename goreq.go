package goreq

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	DefaultClient    = http.DefaultClient
	DefaultTransport = http.DefaultTransport
)

type Agent struct {
	Url            string
	Method         string
	ReqBodyReader  io.Reader
	ReqPreHandlers []ReqPreHandler
	RespHandler    RespHandler
	RespWrapper    Wrapper
	Client         *http.Client
}

type RespHandler func(resp *http.Response, respWrapper Wrapper) error
type ReqPreHandler func(req *http.Request) (*http.Request, error)
type AgentOp func(*Agent) error

func (a *Agent) Do() error {
	req, err := http.NewRequest(a.Method, a.Url, a.ReqBodyReader)
	if err != nil {
		return fmt.Errorf("new request failed: %w", err)
	}

	for _, h := range a.ReqPreHandlers {
		newReq, err := h(req)
		if err != nil {
			return err
		}
		if newReq != nil {
			req = newReq
		}
	}

	if a.Client == nil {
		a.Client = DefaultClient
	}
	if a.Client.Transport == nil {
		a.Client.Transport = DefaultTransport
	}

	resp, err := a.Client.Do(req)
	if err != nil {
		return fmt.Errorf("request do failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &HttpCodeErr{
			Resp: resp,
		}
	}

	if a.RespHandler != nil {
		return a.RespHandler(resp, a.RespWrapper)
	}
	return nil
}

type Wrapper interface {
	SetData(ret interface{})
	Validate() error
}

func JsonResp(ret interface{}) RespHandler {
	return func(resp *http.Response, respWrapper Wrapper) error {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read http body failed: %w", err)
		}

		if respWrapper != nil {
			respWrapper.SetData(ret)
			ret = respWrapper
		}
		if err := json.Unmarshal(body, respWrapper); err != nil {
			return fmt.Errorf("json unmarshal failed: %w, body: %s", err, string(body))
		}
		if respWrapper != nil {
			return respWrapper.Validate()
		}
		return nil
	}
}

func Get(url string, ops ...AgentOp) *Agent {
	return &Agent{
		Url:    url,
		Method: http.MethodGet,
	}
}

func Post(url string) *Agent {
	return &Agent{
		Url:    url,
		Method: http.MethodPost,
	}
}

func Put(url string) *Agent {
	return &Agent{
		Url:    url,
		Method: http.MethodPut,
	}
}

func Patch(url string) *Agent {
	return &Agent{
		Url:    url,
		Method: http.MethodPatch,
	}
}

func Delete(url string) *Agent {
	return &Agent{
		Url:    url,
		Method: http.MethodDelete,
	}
}

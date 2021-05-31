package goreq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ReqPreHandler interface {
	PreHandleRequest(req *http.Request) (*http.Request, error)
}

type ReqPreHandlerFunc func(req *http.Request) (*http.Request, error)

func (f ReqPreHandlerFunc) PreHandleRequest(req *http.Request) (*http.Request, error) {
	return f(req)
}

func JsonReq(reqBody interface{}) AgentOp {
	return AgentOpFunc(func(agent *Agent) error {
		agent.reqPreHandlers = append(agent.reqPreHandlers, ReqPreHandlerFunc(func(req *http.Request) (*http.Request, error) {
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			bs, err := json.Marshal(reqBody)
			if err != nil {
				return nil, fmt.Errorf("json marshal failed: %w", err)
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(bs))
			return nil, nil
		}))
		return nil
	})
}

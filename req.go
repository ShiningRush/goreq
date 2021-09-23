package goreq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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
			buffer := bytes.Buffer{}
			if err := json.NewEncoder(&buffer).Encode(reqBody); err != nil {
				return nil, fmt.Errorf("json marshal failed: %w", err)
			}
			req.Body = ioutil.NopCloser(&buffer)
			return nil, nil
		}))
		return nil
	})
}

func FormReq(values url.Values) AgentOp {
	return AgentOpFunc(func(agent *Agent) error {
		agent.reqPreHandlers = append(agent.reqPreHandlers, ReqPreHandlerFunc(func(req *http.Request) (*http.Request, error) {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
			switch agent.method {
			case http.MethodGet:
				if len(req.URL.RawQuery) == 0 {
					req.URL.RawQuery = values.Encode()
					return req, nil
				}

				req.URL.RawQuery = fmt.Sprintf("%s&%s", req.URL.RawQuery, values.Encode())
			default:
				req.Body = ioutil.NopCloser(strings.NewReader(values.Encode()))
			}
			return req, nil
		}))
		return nil
	})
}

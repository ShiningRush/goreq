package goreq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func TextReq(reqBody string) AgentOp {
	return AgentOpFunc(func(agent *Agent) error {
		agent.reqPreHandlers = append(agent.reqPreHandlers, ReqPreHandlerFunc(func(req *http.Request) (*http.Request, error) {
			req.Header.Add("Content-Type", "text/plain; charset=utf-8")
			r := strings.NewReader(reqBody)
			req.ContentLength = int64(r.Len())
			req.Body = io.NopCloser(r)
			return nil, nil
		}))
		return nil
	})
}

func JsonReq(reqBody interface{}) AgentOp {
	return AgentOpFunc(func(agent *Agent) error {
		agent.reqPreHandlers = append(agent.reqPreHandlers, ReqPreHandlerFunc(func(req *http.Request) (*http.Request, error) {
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			buffer := bytes.Buffer{}
			if err := json.NewEncoder(&buffer).Encode(reqBody); err != nil {
				return nil, fmt.Errorf("json marshal failed: %w", err)
			}
			req.ContentLength = int64(buffer.Len())
			req.Body = io.NopCloser(&buffer)
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
				r := strings.NewReader(values.Encode())
				req.ContentLength = int64(r.Len())
				req.Body = io.NopCloser(strings.NewReader(values.Encode()))
			}
			return req, nil
		}))
		return nil
	})
}

package goreq

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	"net/http"
	"reflect"
	"time"
)

var (
	DefaultClient    = http.DefaultClient
	DefaultTransport = http.DefaultTransport
)

type Agent struct {
	url    string
	method string
	ctx    context.Context

	reqPreHandlers      []ReqPreHandler
	respHandler         RespHandler
	respWrapper         Wrapper
	client              *http.Client
	expectedStatusCodes []int
	retryOpt            *RetryOpt

	existedOps []AgentOp
}

type RetryOpt struct {
	// the max delay of interval
	MaxDelay time.Duration
	// RetryAppError indicate if retry when "RespWrapper validate failed"
	RetryAppError bool
	Attempts      int
}

type AgentOp interface {
	InitialAgent(*Agent) error
}

type AgentOpFunc func(agent *Agent) error

func (f AgentOpFunc) InitialAgent(agent *Agent) error {
	return f(agent)
}

func (a *Agent) Do() error {
	for _, op := range a.existedOps {
		if err := op.InitialAgent(a); err != nil {
			return err
		}
	}
	if len(a.expectedStatusCodes) == 0 {
		a.expectedStatusCodes = append(a.expectedStatusCodes, http.StatusOK)
	}

	req, err := http.NewRequest(a.method, a.url, nil)
	if err != nil {
		return fmt.Errorf("new request failed: %w", err)
	}
	if a.ctx != nil {
		req = req.WithContext(a.ctx)
	}

	for _, h := range a.reqPreHandlers {
		newReq, err := h.PreHandleRequest(req)
		if err != nil {
			return err
		}
		if newReq != nil {
			req = newReq
		}
	}

	if a.client == nil {
		a.client = DefaultClient
	}
	if a.client.Transport == nil {
		a.client.Transport = DefaultTransport
	}

	if a.retryOpt == nil {
		return a.doHttp(req)
	}

	return a.retryDoHttp(req)
}

func (a *Agent) retryDoHttp(req *http.Request) error {
	attempts := 6
	if a.retryOpt.Attempts != 0 {
		attempts = a.retryOpt.Attempts
	}

	maxDelay := time.Duration(0)
	if a.retryOpt.MaxDelay != 0 {
		maxDelay = a.retryOpt.MaxDelay
	}
	return retry.Do(func() error { return a.doHttp(req) },
		retry.Attempts(uint(attempts)),
		retry.MaxDelay(maxDelay),
		retry.Context(req.Context()))
}

func (a *Agent) doHttp(req *http.Request) error {
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("request do failed: %w", err)
	}
	defer resp.Body.Close()

	if !a.isInExpectedStatusCodes(resp.StatusCode) {
		return NewHttpCodeErr(a.expectedStatusCodes, resp)
	}

	if a.respHandler != nil {
		return a.respHandler.HandleResponse(resp, a.respWrapper)
	}
	return nil
}

func (a *Agent) isInExpectedStatusCodes(code int) (find bool) {
	for _, ac := range a.expectedStatusCodes {
		if ac == code {
			find = true
			return
		}
	}
	return
}

func (a *Agent) Ops(ops ...AgentOp) *Agent {
	for _, op := range ops {
		a.existedOps = append(a.existedOps, op)
	}
	return a
}

type Wrapper interface {
	SetData(ret interface{})
	Validate() error
}

func Retry(opt *RetryOpt) AgentOpFunc {
	return func(agent *Agent) error {
		agent.retryOpt = opt
		return nil
	}
}

func ExpectedStatusCodes(codes []int) AgentOpFunc {
	return func(agent *Agent) error {
		agent.expectedStatusCodes = codes
		return nil
	}
}

func Context(ctx context.Context) AgentOpFunc {
	return func(agent *Agent) error {
		agent.ctx = ctx
		return nil
	}
}

func SetHeader(header http.Header) AgentOpFunc {
	return func(agent *Agent) error {
		agent.reqPreHandlers = append(agent.reqPreHandlers, ReqPreHandlerFunc(func(req *http.Request) (*http.Request, error) {
			req.Header = header
			return nil, nil
		}))
		return nil
	}
}
func RespWrapper(wrapper Wrapper) AgentOpFunc {
	return func(agent *Agent) error {
		if reflect.TypeOf(wrapper).Kind() != reflect.Ptr {
			return fmt.Errorf("response wrapper should be ptr")
		}
		agent.respWrapper = wrapper
		return nil
	}
}

// CustomRespHandler specify a custom RespHandler
func CustomRespHandler(handler RespHandler) AgentOpFunc {
	return func(agent *Agent) error {
		agent.respHandler = handler
		return nil
	}
}

// Get start a request with GET
func Get(url string, ops ...AgentOp) *Agent {
	return &Agent{
		url:        url,
		method:     http.MethodGet,
		existedOps: ops,
	}
}

// Post start a request with POST
func Post(url string, ops ...AgentOp) *Agent {
	return &Agent{
		url:        url,
		method:     http.MethodPost,
		existedOps: ops,
	}
}

// Put start a request with PUT
func Put(url string, ops ...AgentOp) *Agent {
	return &Agent{
		url:        url,
		method:     http.MethodPut,
		existedOps: ops,
	}
}

// Patch start a request with PATCH
func Patch(url string, ops ...AgentOp) *Agent {
	return &Agent{
		url:        url,
		method:     http.MethodPatch,
		existedOps: ops,
	}
}

// Delete start a request with DELETE
func Delete(url string, ops ...AgentOp) *Agent {
	return &Agent{
		url:        url,
		method:     http.MethodDelete,
		existedOps: ops,
	}
}

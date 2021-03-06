package goreq

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
)

// RespHandler you can implement some special cases
// TIPS: Usually JsonResp, RawResp and HybridResp handle most situations
type RespHandler interface {
	HandleResponse(resp *http.Response, respWrapper Wrapper) error
}

// RawResp use http.Response and []byes to accept response
func RawResp(resp *http.Response, bs *[]byte) *RawRespHandler {
	return &RawRespHandler{
		resp: resp,
		bs:   bs,
	}
}

// RawRespHandler is a wrapper to implement AgentOp and RespHandler
type RawRespHandler struct {
	resp *http.Response
	bs   *[]byte
}

func (h *RawRespHandler) HandleResponse(resp *http.Response, respWrapper Wrapper) error {
	if h.resp != nil {
		*h.resp = *resp
	}
	if h.bs != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read http body failed: %w", err)
		}
		*h.bs = body
	}
	return nil
}
func (h *RawRespHandler) InitialAgent(a *Agent) error {
	a.respHandler = h
	return nil
}

// HybridResp can handle hybrid response such as Json and Raw
// you can use RespHandlerPredicate to indicate when use which resp handler with Predicate
func HybridResp(predicate ...RespHandlerPredicate) *HybridHandler {
	return &HybridHandler{predicates: predicate}
}

type RespHandlerPredicate struct {
	Predicate   func(response *http.Response) bool
	RespHandler RespHandler
}

// RawRespHandler is a wrapper to implement AgentOp and RespHandler
type HybridHandler struct {
	predicates []RespHandlerPredicate
}

func (h *HybridHandler) HandleResponse(resp *http.Response, respWrapper Wrapper) error {
	for i, p := range h.predicates {
		if p.Predicate(resp) {
			if err := p.RespHandler.HandleResponse(resp, respWrapper); err != nil {
				return fmt.Errorf("hybrid resp handle failed at %d, err: %s", i, err)
			}
		}
	}

	return nil
}
func (h *HybridHandler) InitialAgent(a *Agent) error {
	a.respHandler = h
	return nil
}

// JsonResp use to handler json response, ret must be a ptr
func JsonResp(ret interface{}) *JsonRespHandler {
	return &JsonRespHandler{ret: ret}
}

// JsonRespHandler is a wrapper to implement AgentOp and RespHandler
type JsonRespHandler struct {
	ret interface{}
}

func (h *JsonRespHandler) HandleResponse(resp *http.Response, respWrapper Wrapper) error {
	if respWrapper != nil {
		respWrapper.SetData(h.ret)
		h.ret = respWrapper
	}

	// json.Decoder is very well, but it can not get invalid content when unmarshal failed
	// so we need to read all body, so can return it when unmarshal failed
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body failed: %w", err)
	}
	if err := json.Unmarshal(body, &h.ret); err != nil {
		return fmt.Errorf("unmarshal body failed: %w, body: %s", err, body)
	}

	if respWrapper != nil {
		return respWrapper.Validate()
	}
	return nil
}
func (h *JsonRespHandler) InitialAgent(a *Agent) error {
	if reflect.TypeOf(h.ret).Kind() != reflect.Ptr {
		return fmt.Errorf("result payload should be ptr")
	}
	a.respHandler = h
	return nil
}

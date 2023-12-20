package goreq

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawResp(t *testing.T) {
	var resp http.Response
	var bodyBytes []byte
	err := Get("http://httpbin.org/get", RawResp(&resp, &bodyBytes)).Do()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int(resp.ContentLength), len(bodyBytes))

	err = Post("http://httpbin.org/post", RawResp(&resp, &bodyBytes)).Do()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int(resp.ContentLength), len(bodyBytes))

	err = Delete("http://httpbin.org/delete", RawResp(&resp, &bodyBytes)).Do()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int(resp.ContentLength), len(bodyBytes))

	err = Put("http://httpbin.org/put", RawResp(&resp, &bodyBytes)).Do()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int(resp.ContentLength), len(bodyBytes))

	err = Patch("http://httpbin.org/patch", RawResp(&resp, &bodyBytes)).Do()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int(resp.ContentLength), len(bodyBytes))
}

func TestAllowStatusCodes(t *testing.T) {
	err := Get("http://httpbin.org/get", ExpectedStatusCodes([]int{http.StatusFound})).Do()
	assert.NotNil(t, err)
}

type HttpBinResp struct {
	Headers map[string]string `json:"headers"`
	Data    string            `json:"data"`
	Json    interface{}       `json:"json"`
	Form    interface{}       `json:"form"`
}

type JsonRequest struct {
	String string
	Int    int
}

func TestJsonReqResp(t *testing.T) {
	req := JsonRequest{
		String: "hello",
		Int:    666,
	}
	respBody := JsonRequest{}
	resp := HttpBinResp{
		Json: &respBody,
	}
	err := Post("http://httpbin.org/post",
		JsonReq(req),
		JsonResp(&resp)).Do()
	assert.NoError(t, err)
	assert.Equal(t, req, respBody)
}

func TestFormReq(t *testing.T) {
	req := url.Values{
		"key1": []string{"v1"},
		"key2": []string{"vv1", "vv2"},
	}
	resp := HttpBinResp{}
	err := Post("http://httpbin.org/post",
		FormReq(req),
		JsonResp(&resp)).Do()
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"key1": "v1",
		"key2": []interface{}{
			"vv1",
			"vv2",
		},
	}, resp.Form)
}

type CountResultWrapper struct {
	Headers map[string]string `json:"headers"`
	Data    string            `json:"data"`
	Json    interface{}       `json:"json"`

	doValidationCallback  func() int
	returnOkAfterRequests int
}

func (w *CountResultWrapper) SetData(ret interface{}) {
	w.Json = ret
}

func (w *CountResultWrapper) Validate() error {
	count := w.doValidationCallback()
	if count == w.returnOkAfterRequests {
		return nil
	}

	return fmt.Errorf("%d request is expected failed", count)
}

func TestRetryAndValidation(t *testing.T) {
	req := JsonRequest{
		String: "hello",
		Int:    666,
	}
	respData := JsonRequest{}

	callCount := 0
	cb := func() int {
		callCount++
		return callCount
	}

	err := Post("http://httpbin.org/post",
		JsonReq(req),
		JsonResp(&respData),
		RespWrapper(&CountResultWrapper{returnOkAfterRequests: 2, doValidationCallback: cb}),
		Retry(&RetryOpt{
			Attempts: 2,
		}),
	).Do()
	assert.NoError(t, err)
	assert.Equal(t, req, respData)

	callCount = 0
	err = Post("http://httpbin.org/post",
		JsonReq(req),
		JsonResp(&respData),
		RespWrapper(&CountResultWrapper{returnOkAfterRequests: 2, doValidationCallback: cb}),
		Retry(&RetryOpt{
			Attempts: 1,
		}),
	).Do()
	assert.NotNil(t, err)
}

package goreq

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestWrapper struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (w *TestWrapper) SetData(ret interface{}) {
	w.Data = ret
}
func (w *TestWrapper) Validate() error {
	if w.Code != 0 {
		return fmt.Errorf("server code[%d] is incorrect", w.Code)
	}

	return nil
}

func TestJsonRespHandler_HandleResponse(t *testing.T) {
	type TestRet struct {
		IntA int    `json:"intA"`
		StrB string `json:"strB"`
	}

	tests := []struct {
		caseDesc    string
		giveRet     *TestRet
		giveResp    *http.Response
		giveWrapper Wrapper
		wantErr     error
		wantRet     *TestRet
	}{
		{
			caseDesc: "normal",
			giveResp: &http.Response{
				Body: io.NopCloser(strings.NewReader(`{"code":0,"data":{"intA":5,"strB":"ok"}}`)),
			},
			giveWrapper: &TestWrapper{},
			giveRet:     &TestRet{},
			wantRet: &TestRet{
				IntA: 5,
				StrB: "ok",
			},
		},
		{
			caseDesc: "wrapper validate failed",
			giveResp: &http.Response{
				Body: io.NopCloser(strings.NewReader(`{"code":1,"data":{"intA":5,"strB":"ok"}}`)),
			},
			giveWrapper: &TestWrapper{},
			giveRet:     &TestRet{},
			wantErr:     fmt.Errorf("server code[1] is incorrect"),
			wantRet: &TestRet{
				IntA: 5,
				StrB: "ok",
			},
		},
		{
			caseDesc: "wrong json formal",
			giveResp: &http.Response{
				Body: io.NopCloser(strings.NewReader(`{"code":"1","data":{"intA":5,"strB":"ok"}}`)),
			},
			giveWrapper: &TestWrapper{},
			giveRet:     &TestRet{},
			wantErr: fmt.Errorf(`unmarshal body failed: %w, body: %s`,
				&json.UnmarshalTypeError{
					Value:  "string",
					Type:   reflect.TypeOf(1),
					Offset: 0,
					Struct: "TestWrapper",
					Field:  "code",
				},
				`{"code":"1","data":{"intA":5,"strB":"ok"}}`),
			wantRet: &TestRet{
				IntA: 5,
				StrB: "ok",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.caseDesc, func(t *testing.T) {
			h := JsonResp(tc.giveRet)
			err := h.HandleResponse(tc.giveResp, tc.giveWrapper)
			// is here a better way to compare it? when json unmarshal failed it will get
			// a wrapped json.UnmarshalJsonError
			assert.Equal(t, getErrStr(err), getErrStr(err))
			assert.Equal(t, tc.wantRet, tc.giveRet)
		})
	}
}

func getErrStr(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

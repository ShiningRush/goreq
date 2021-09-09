package goreq

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"testing/iotest"
)

func TestHttpCodeErr(t *testing.T) {
	tests := []struct {
		giveHttpCodeErr *HttpCodeErr
		wantErr         string
	}{
		{
			giveHttpCodeErr: NewHttpCodeErr([]int{http.StatusOK, http.StatusAccepted}, &http.Response{
				StatusCode: http.StatusFound,
				Body:       ioutil.NopCloser(strings.NewReader("test")),
			}),
			wantErr: "http code[302] is not expected([200 202]), body: test",
		},
		{
			giveHttpCodeErr: NewHttpCodeErr([]int{http.StatusOK, http.StatusAccepted}, &http.Response{
				StatusCode: http.StatusFound,
				Body:       ioutil.NopCloser(iotest.ErrReader(errors.New("mock error"))),
			}),
			wantErr: "http code[302] is not expected([200 202]) and read body failed: mock error",
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.wantErr, tc.giveHttpCodeErr.Error())
	}
}

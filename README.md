# goreq

[中文](./README_cn.md)

---

I have been using [gorequest](https://github.com/parnurzeal/gorequest) before, but the following pain points were found during use:
- When using this library to call between services, it is impossible to insert some common processing logic. For example, the server usually return fields like `code`, and the client needs to check whether its value is validated
- Maybe cause concurrency problems (share one agent)
- Use short connection by default
- Does not support hybrid responses, for example, a certain platform may return `json` or `text`, depending on the value of `Conten-Type`
- Not supported retry

After looking around on github, I didn't find a library that I was satisfied with, so `goreq` was born.

## Features
The supported features are as follows:
- http raw request
- serialization request
- deserialize response
- handling public wrapper
- retry 
- Validating response

### http raw request
The most basic usage of `goreq` is as follows:
```go
    // using raw bytes and http.response to handle
    var resp http.Response
    var bodyBytes []byte
    err := Get("https://httpbin.org/get", RawResp(&resp, &bodyBytes)).Do()
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    assert.Equal(t, int(resp.ContentLength), len(bodyBytes))
```

In the above demo, we directly use the byte array and `http.Response` to accept the response. This approach is suitable for scenarios where the response structure is not fixed, such as web pages, files, etc. may be returned.

### Serialized request body
goreq currently supports the serialization of request bodies in the following formats:
- json

#### JsonRequest
If you need to serialize a structure into request body with json, you can use `JsonReq`, please refer to the following code:
```go
  req := JsonRequest{
      String: "hello",
      Int: 666,
  }
  err := Post("https://httpbin.org/post",
    JsonReq(req)).Do()
  assert.NoError(t, err)
```
### Deserialize response content
goreq supports deserialization of responses:
- json
- hybrid

#### JsonResponse
If the content of the response is json format, we can use goreq deserialize the content as below:
```go
  req := JsonRequest{
    String: "hello",
    Int: 666,
  }
  respBody := JsonRequest{}
  resp := HttpBinResp{
    Json: &respBody,
  }
  err := Post("https://httpbin.org/post",
  JsonReq(req),
  JsonResp(&resp)).Do()
  assert.NoError(t, err)
  assert.Equal(t, req, respBody)
```

#### Heterogeneous response
If the response is hybrid, such as a text/html when an error occurs, and json when it is normal, we can use `HybridResp` to handle different situations:
```go
err := Post("https://httpbin.org/post",
        JsonReq(req),
        HybridResp(RespHandlerPredicate{
            Predicate: func(response *http.Response) bool {
                return response.StatusCode == http.StatusOK
            },
            // if status code is 200, using JsonResp
            RespHandler: JsonResp(&ret),
        }, RespHandlerPredicate{
Predicate: func(response *http.Response) bool {
                return response.StatusCode != http.StatusOK
            },
            // if status ocde is not 200, using RawResp
            RespHandler: RawResp(nil, &bt),
        }),
    ).Do()
```

#### Handling public wrapper
Most of the time the server will have a common wrapper when returning the results, such as:
```
{
  "code": 0,
  "msg": "",
  "data": null,
  "req_id": "xxxx"
}
```
When invoking the service, in fact, Client only cares about the content in `data` most of the time, and only needs to care other information when api failed. In order to separate concerns, we can use `RespWrapper` of goreq.
In addition to defining the public wrapper, another important responsibility of `RespWrapper` is to identify whether the current request fails. Refer to the following code:
```go
type CountResultWrapper struct {
  Headers map[string]string `json:"headers"`
  Data string `json:"data"`
  Json interface{} `json:"json"`
  
  doValidationCallback func() int
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
    Int: 666,
  }
  respData := JsonRequest{}
  
  callCount := 0
  cb := func() int {
    callCount++
    return callCount
  }

  err := Post("https://httpbin.org/post",
    JsonReq(req),
    JsonResp(&respData),
    RespWrapper(&CountResultWrapper{returnOkAfterRequests: 2, doValidationCallback: cb}),
    Retry(&RetryOpt{
    Attempts: 2,
  }),
  ).Do()
  // wrapper will ok when retry 2 times
  assert.NoError(t, err)
  assert.Equal(t, req, respData)
  
  callCount = 0
  err = Post("https://httpbin.org/post",
    JsonReq(req),
    JsonResp(&respData),
    RespWrapper(&CountResultWrapper{returnOkAfterRequests: 2, doValidationCallback: cb}),
    Retry(&RetryOpt{
        Attempts: 1,
    }),
  ).Do()
  assert.NotNil(t, err)
}
```

The above code shows most of the usage of goreq, including the retry and validated result as mentioned below.
`CountResultWrapper` will return nil when it reaches a certain number of requests, other situations will fail. You can see the above example. When the configuration is retried once (equivalent to not retrying), an error will be generated. It will succeed when you configure two retries.
Here are several important points:
- When the response is returned, the `Valiadte` of `RespWrapper` will first verify whether there is a failure
- If the request fails, you can specify goreq to retry

#### Retry
For the retry example, refer to the above codes.

#### Validating response
Here are two ways to validate response:
- Validating in the wrapper: you can implement `RespWrapper` to do it
- Validating http status code: you can set `ExpectedStatusCodes` to do it.If the http status code is not in this range, an error will be generated.
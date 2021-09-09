# goreq

---

之前一直使用的 [gorequest](https://github.com/parnurzeal/gorequest), 但是使用过程中发现如下几个痛点：
- 使用该库进行服务间调用时没法插入一些公共处理的逻辑, 比如一般服务都会返回类似 `code` 这样的字段, 调用方需要去检测它的值是否正确
- 不注意会引起并发问题( 共用一个 agent )
- 默认使用短连接
- 不支持异构响应, 比如某个平台可能会返回 `json` 也可能会返回 `text`, 这取决于 `Conten-Type` 的值
- 不支持重试

在 github 上看了一圈都都没找到自己满意的库，于是 `goreq` 诞生了。

## 特性
支持的特性如下：
- http raw request
- 序列化请求
- 反序列化响应
- 处理公共 wrapper
- 重试
- 返回结果验证

### http raw request
`goreq` 最基础的用法如下：
```go
    // using raw bytes and http.response to handle
    var resp http.Response
    var bodyBytes []byte
    err := Get("https://httpbin.org/get", RawResp(&resp, &bodyBytes)).Do()
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    assert.Equal(t, int(resp.ContentLength), len(bodyBytes))
```

上面的 demo 中我们直接使用字节数组和 `http.Response` 来承载了响应，这种做法适用于响应结构不固定的场景，比如可能返回网页，文件等。

### 序列化请求体
goreq 目前支持以下几种格式的请求体序列化：
- json

#### JsonRequest
如果你需要将一个结构体序列化为 Json 并添加到请求的 Body 中，可以使用 `JsonReq`，请参考以下代码：
```go
	req := JsonRequest{
		String: "hello",
		Int:    666,
	}
	err := Post("https://httpbin.org/post",
		JsonReq(req)).Do()
	assert.NoError(t, err)
```
### 反序列化响应内容
goreq 支持反序列化响应，目前支持:
- json
- 异构响应

#### JsonResponse
如果响应的内容是 Json 格式的，我们可以使用以下语句来自动反序列化结果：
```go
	req := JsonRequest{
		String: "hello",
		Int:    666,
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

#### 异构响应
如果响应是混合式的，比如错误时返回是一个text/html，正常时是json，我们可以使用 `HybridResp` 来处理不同情况的处理方式。
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
	        Predicate:   func(response *http.Response) bool {
                return response.StatusCode != http.StatusOK
            },
            // if status ocde is not 200, using RawResp
            RespHandler: RawResp(nil, &bt),
        }), 
    ).Do()
```

#### 处理公共 wrapper
大部分时候业务在返回结果时都会有公共的结构，比如类似：
```
{
  "code": 0,
  "msg": "",
  "data": null,
  "req_id": "xxxx"
}
```
在调用服务时，其实 Client 多数时候只关心 `data` 中的内容，只有在失败时才需要关注其他信息，为了分离关注点，我们可以给 goreq 设置 `RespWrapper`。
`RespWrapper` 除了定义出公共结构外，还有一个很重要的职责就是识别当前请求是否失败，参考以下代码：
```go
type CountResultWrapper struct {
	Headers map[string]string `json:"headers"`
	Data    string            `json:"data"`
	Json    interface{}       `json:"json"`

	doValidationCallback func() int
	returnOkAfterRequests int
}

func (w *CountResultWrapper) SetData(ret interface{})  {
	w.Json = ret
}

func (w *CountResultWrapper) Validate() error  {
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

	err := Post("https://httpbin.org/post",
		JsonReq(req),
		JsonResp(&respData),
		RespWrapper(&CountResultWrapper{returnOkAfterRequests: 2, doValidationCallback: cb}),
		Retry(&RetryOpt{
			Attempts:      2,
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
			Attempts:      1,
		}),
	).Do()
	assert.NotNil(t, err)
}
```

以上代码展示了 goreq 绝大多数功能，包括接下来提到的重试和返回结果认证。
`CountResultWrapper` 会验证成功，当达到一定请求数后，其他情况都会失败，可以看到上面的例子，在配置重试一次时（相当于不重试），是会产生错误的，而重试两次则会成功。
这里体现几个重要的点：
- 当响应返回后，会先由 `RespWrapper` 的 `Valiadte` 来验证是否存在失败的情况
- 如果请求失败，那么可以指定 goreq 进行重试

#### 重试
重试的例子参考上面

#### 验证响应
验证响应可以分为两个方向：
- 在 wrapper 中验证：这个由 `RespWrapper` 自行实现
- HttpStatusCode 验证：这个可以通过 `ExpectedStatusCodes` 来设置期望的 Code，如果 http status code 不在此范围内则会报错。
/**
 * @author zhagnxiaoping
 * @date  2024/6/15 11:57
 */
package leadutil

import (
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"strconv"
	"sync"
	"time"
)

// Description:

type FastHTTPClient struct {
	client *fasthttp.Client
}

func NewFastHTTPClient() *FastHTTPClient {
	c := &FastHTTPClient{
		client: &fasthttp.Client{
			MaxConnsPerHost:     1024,
			MaxConnWaitTimeout:  time.Second * 15,
			ReadTimeout:         time.Second * 15,
			WriteTimeout:        time.Second * 15,
			MaxIdleConnDuration: time.Second * 15,
		},
	}

	return c
}

// ReleaseResponse 回收resp资源。
// resp,err := client.Get()
// defer leadutil.ReleaseResponse(resp)
func ReleaseResponse(resp *fasthttp.Response) {
	if resp != nil {
		fasthttp.ReleaseResponse(resp)
	}
}

func (client *FastHTTPClient) Get(url string, opts ...Option) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseRequest(req)
	}()

	// DO GET request
	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI(url)

	err := client.Do(req, resp, opts...)
	return resp, err
}
func (client *FastHTTPClient) Do(req *fasthttp.Request, resp *fasthttp.Response, opts ...Option) error {
	opt := acquireOpt()
	opt.actionName = string(req.URI().Path())

	for _, o := range opts {
		o(opt)
	}

	defer func() {
		releaseOpt(opt)
	}()
	req.Header.Set("Connection", "keep-alive")

	for k, v := range opt.header {
		req.Header.Set(k, v)
	}

	now := time.Now()
	err := client.client.Do(req, resp)
	elapsed := GetElapsedMS(now)

	resp.Header.Set("elapsed", strconv.Itoa(int(elapsed)))

	if err != nil {
		RecordFailure("http", opt.actionName, elapsed, fmt.Sprint(err))
	}

	statusCode := resp.StatusCode()

	if opt.statusCodeHandler != nil {
		opt.statusCodeHandler(statusCode, req, resp)
	}

	if opt.autoRecordLocustMsg {
		if statusCode >= 400 {
			err = errors.New(fmt.Sprintf("status code:%d,body:%s", statusCode, resp.Body()))
			RecordFailure("http", opt.actionName, elapsed, err.Error())
		} else {
			RecordSuccess("http", opt.actionName, elapsed, int64(resp.Header.ContentLength()))
		}
	}

	return err
}

func (client *FastHTTPClient) Post(url string, data []byte, opts ...Option) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(url)

	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetBody(data)
	err := client.Do(req, resp)

	return resp, err
}

type option struct {
	header              map[string]string
	actionName          string
	autoRecordLocustMsg bool
	statusCodeHandler   func(code int, req *fasthttp.Request, resp *fasthttp.Response)
}

func (opt *option) Reset() {
	for k := range opt.header {
		delete(opt.header, k)
	}
}

var optPool sync.Pool

func acquireOpt() *option {
	v := optPool.Get()
	if v == nil {
		return defaultOption()
	}
	return v.(*option)
}

func releaseOpt(opt *option) {
	if opt != nil {
		opt.Reset()
		optPool.Put(opt)
	}
}

func defaultOption() *option {
	opt := &option{
		header:              map[string]string{},
		autoRecordLocustMsg: true,
	}

	return opt
}

// Option 可选项接口
type Option func(*option)

// Header 设置header
func Header(kvPairs ...string) Option {
	return func(o *option) {
		l := len(kvPairs)
		for i := 0; i < l; i += 2 {
			o.header[kvPairs[i]] = kvPairs[i+1]
		}
	}
}

// ActionName 设置当前操作名，用于上报请求数据时标记使用，默认是path
func ActionName(name string) Option {
	return func(o *option) {
		o.actionName = name
	}
}

// StatusCodeHandler 设置状态码处理函数
func StatusCodeHandler(fn func(code int, req *fasthttp.Request, resp *fasthttp.Response)) Option {
	return func(o *option) {
		o.statusCodeHandler = fn
	}
}

// DisableRecordLocustMsg 关闭自动记录locust信息
func DisableRecordLocustMsg() Option {
	return func(o *option) {
		o.autoRecordLocustMsg = false
	}
}

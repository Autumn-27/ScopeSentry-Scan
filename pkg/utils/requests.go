// utils-------------------------------------
// @file      : requests.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/19 20:48
// -------------------------------------------

package utils

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/projectdiscovery/httpx/runner"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/valyala/fasthttp"
	"net"
	"syscall"
	"time"
)

type request struct {
}

var Requests *request

var HttpClient *fasthttp.Client
var Wappalyzer *wappalyzer.Wappalyze

func InitializeRequests() {
	gologger.DefaultLogger.SetMaxLevel(levels.LevelWarning) // increase the verbosity (optional)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	HttpClient = &fasthttp.Client{
		ReadTimeout:                   time.Second * 10,
		WriteTimeout:                  time.Second * 10,
		MaxIdleConnDuration:           time.Second * 10,
		NoDefaultUserAgentHeader:      true,
		DisableHeaderNamesNormalizing: true,
		DisablePathNormalizing:        true,
		TLSConfig:                     tlsConfig,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}
	Requests = &request{}
	paralyze, err := wappalyzer.New()
	if err != nil {
		fmt.Printf("init wappalyzer error: %v", err)
	}
	Wappalyzer = paralyze
}

func (r *request) HttpGet(uri string) (types.HttpResponse, error) {
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	// 最后需要归还req、resp到池中
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI(uri)

	if err := HttpClient.Do(req, resp); err != nil {
		return types.HttpResponse{}, err
	}
	tmp := types.HttpResponse{}
	tmp.Url = uri
	// 定义最大响应体大小 (100KB)
	const maxBodySize = 100 * 1024

	// 截断 Body
	body := resp.Body()
	if len(body) > maxBodySize {
		body = body[:maxBodySize]
	}
	tmp.Body = string(body)
	tmp.StatusCode = resp.StatusCode()
	if location := resp.Header.Peek("location"); len(location) > 0 {
		tmp.Redirect = string(location)
	} else {
		rd := resp.Header.Peek("Location")
		if len(rd) > 0 {
			tmp.Redirect = string(rd)
		} else {
			tmp.Redirect = ""
		}
	}
	tmp.ContentLength = resp.Header.ContentLength()
	if tmp.ContentLength < 0 {
		tmp.ContentLength = len(resp.Body())
	}
	return tmp, nil
}
func (r *request) HttpGetByte(uri string) ([]byte, error) {
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	// 最后需要归还req、resp到池中
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI(uri)
	if err := HttpClient.Do(req, resp); err != nil {
		return make([]byte, 0), err
	}
	tmp := resp.Body()
	return tmp, nil
}

func (r *request) HttpPost(uri string, requestBody []byte, ct string) (error, *fasthttp.Response) {
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(uri)
	if ct == "json" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.SetBody(requestBody)

	if err := HttpClient.Do(req, resp); err != nil {
		return err, nil
	}
	return nil, resp
}

func (r *request) HttpGetNoRes(uri string) error {
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	// 设置请求 URI
	req.SetRequestURI(uri)

	req.Header.SetMethod(fasthttp.MethodGet)

	// 发送请求
	if err := HttpClient.Do(req, resp); err != nil {
		return err
	}
	// 直接丢弃响应体（或者关闭它）
	resp.Reset()
	return nil
}

func (r *request) HttpPostNoRes(uri string, requestBody []byte, ct string) error {
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(uri)
	if ct == "json" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.SetBody(requestBody)

	if err := HttpClient.Do(req, resp); err != nil {
		return err
	}
	resp.Reset()
	return nil
}

var dialer = &net.Dialer{
	Timeout: 3 * time.Second,
}

func (r *request) TcpRecv(ip string, port uint16) ([]byte, error) {
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	response := make([]byte, 4096)
	err = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		return []byte{}, err
	}
	length, err := conn.Read(response)
	if err != nil {
		var netErr net.Error
		if (errors.As(err, &netErr) && netErr.Timeout()) ||
			errors.Is(err, syscall.ECONNREFUSED) { // timeout error or connection refused
			return []byte{}, err
		}
		return response[:length], nil
	}
	return response[:length], nil
}

func (r *request) Httpx(targets []string, resultCallback func(r types.AssetHttp), cdncheck string, screenshot bool, screenshotTimeout int, tLSProbe bool) {
	// 设置超时上下文
	timeout := 30 * time.Minute // 设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	options := runner.Options{
		RandomAgent:               true,
		Methods:                   "GET",
		JSONOutput:                false,
		TLSProbe:                  tLSProbe,
		Threads:                   30,
		RateLimit:                 100,
		InputTargetHost:           targets,
		Favicon:                   true,
		ExtractTitle:              true,
		TechDetect:                true,
		OutputWebSocket:           true,
		OutputServerHeader:        true,
		OutputIP:                  true,
		OutputCName:               true,
		ResponseHeadersInStdout:   true,
		ResponseInStdout:          true,
		Base64ResponseInStdout:    false,
		Jarm:                      true,
		OutputCDN:                 cdncheck,
		Location:                  false,
		HostMaxErrors:             30,
		StoreResponse:             false,
		StoreChain:                false,
		MaxResponseBodySizeToRead: 100000,
		Screenshot:                screenshot,
		ScreenshotTimeout:         time.Duration(screenshotTimeout) * time.Second,
		Timeout:                   10,
		Wappalyzer:                Wappalyzer,
		DisableStdout:             true,
		//InputFile: "./targetDomains.txt", // path to file containing the target domains list
		OnResult: func(r runner.Result) {
			// 检查上下文是否已关闭
			select {
			case <-ctx.Done():
				// 如果上下文关闭，直接返回
				logger.SlogWarnLocal(fmt.Sprintf("Context closed before processing result for %s", r.Input))
				return
			default:
				// 继续处理结果
				if r.Err != nil {
					logger.SlogDebugLocal(fmt.Sprintf("HttpxScan error %s: %s", r.Input, r.Err))
				} else {
					ah := Tools.HttpxResultToAssetHttp(r)
					//fmt.Printf("%s %s %d\n", r.Input, r.Host, r.StatusCode)
					resultCallback(ah)
				}
			}
		},
	}
	if err := options.ValidateOptions(); err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("httpx options Validate error: %v", err))
	}

	httpxRunner, err := runner.New(&options)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("%v get error: %v", targets, err))
	}
	defer httpxRunner.Close()

	// 创建一个通道，用于通知任务完成
	done := make(chan struct{})

	// 启动一个 goroutine 执行 httpxRunner，并在执行完后通知主程序
	go func() {
		httpxRunner.RunEnumeration() // 执行 httpx 扫描
		close(done)                  // 扫描完成后关闭通道，通知主程序
	}()

	select {
	case <-ctx.Done(): // 超时后返回
		logger.SlogWarnLocal(fmt.Sprintf("HttpxScan for %s timed out after %v", targets, timeout))
		return
	case <-done: // 扫描完成后继续执行
		logger.SlogDebugLocal(fmt.Sprintf("HttpxScan for %s completed successfully", targets))
	}
}

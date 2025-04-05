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
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
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
		ReadBufferSize:                4 * 1024 * 1024,
		MaxResponseBodySize:           10 * 1024 * 1024,
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

	tmp.Body = string(resp.Body())
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

func (r *request) HttpGetWithCustomHeader(uri string, customHeaders []string) (types.HttpResponse, error) {
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	// 最后需要归还req、resp到池中
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI(uri)

	// 设置自定义 headers
	for _, header := range customHeaders {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(parts[0], strings.TrimSpace(parts[1]))
		}
	}

	if err := HttpClient.Do(req, resp); err != nil {
		return types.HttpResponse{}, err
	}
	tmp := types.HttpResponse{}
	tmp.Url = uri
	// 定义最大响应体大小 (100KB)
	const maxBodySize = 4 * 1024 * 1024

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

func (r *request) HttpGetByteWithCustomHeader(uri string, customHeaders []string) ([]byte, error) {
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	// 最后需要归还req、resp到池中
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI(uri)

	// 设置自定义 headers
	for _, header := range customHeaders {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(parts[0], strings.TrimSpace(parts[1]))
		}
	}

	if err := HttpClient.Do(req, resp); err != nil {
		return make([]byte, 0), err
	}
	tmp := resp.Body()
	return tmp, nil
}

func (r *request) HttpPostWithCustomHeader(uri string, requestBody []byte, ct string, customHeaders []string) (error, *fasthttp.Response) {
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(uri)

	// 设置自定义 headers
	for _, header := range customHeaders {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(parts[0], strings.TrimSpace(parts[1]))
		}
	}

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

func (r *request) HttpGetNoResWithCustomHeader(uri string, customHeaders []string) error {
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	// 设置请求 URI
	req.SetRequestURI(uri)
	req.Header.SetMethod(fasthttp.MethodGet)

	// 设置自定义 headers
	for _, header := range customHeaders {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(parts[0], strings.TrimSpace(parts[1]))
		}
	}

	// 发送请求
	if err := HttpClient.Do(req, resp); err != nil {
		return err
	}
	// 直接丢弃响应体（或者关闭它）
	resp.Reset()
	return nil
}

func (r *request) HttpPostNoResWithCustomHeader(uri string, requestBody []byte, ct string, customHeaders []string) error {
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(uri)

	// 设置自定义 headers
	for _, header := range customHeaders {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(parts[0], strings.TrimSpace(parts[1]))
		}
	}

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

func (r *request) Httpx(targets []string, resultCallback func(r types.AssetHttp), cdncheck string, screenshot bool, screenshotTimeout int, tLSProbe bool, followRedirects bool, ctx context.Context, executionTimeout int, bypassHeader bool) {
	// 设置超时上下文
	timeout := time.Duration(executionTimeout) * time.Minute // 设置超时时间
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	rootDomain := ""
	customHeaders := []string{}
	if bypassHeader {
		if len(targets) != 0 {
			domain, err := Tools.GetRootDomain(targets[0])
			if err != nil {
				rootDomain = targets[0]
			} else {
				rootDomain = domain
			}
		}
		customHeaders = []string{
			"X-Forwarded-For-Original:127.0.0.1",
			"X-Forwarded-For:127.0.0.1",
			"X-Real-IP:127.0.0.1",
			"X-Forwarded-Host:127.0.0.1",
			"X-Forwarded-Proto:127.0.0.1",
			"Forwarded:127.0.0.1",
			"Via:127.0.0.1",
			"Client-IP:127.0.0.1",
			"True-Client-IP:127.0.0.1",
			"X-Originating-IP:127.0.0.1",
			"X-Client-IP:127.0.0.1",
			"Referer:" + rootDomain,
		}
	}

	options := runner.Options{
		CustomHeaders:             customHeaders,
		FollowRedirects:           followRedirects,
		MaxRedirects:              5,
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
		MaxResponseBodySizeToRead: math.MaxInt32,
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

// HttpGetWithRetry 发起一个带重试机制和代理支持的 HTTP GET 请求。
// 参数：
//   - requestURL: 请求地址
//   - timeout: 每次请求的超时时间
//   - maxRetries: 最大重试次数
//   - retryInterval: 每次重试之间的等待间隔
//   - headers: 请求头（如 User-Agent）
//   - proxyURL: 代理地址，若为空则不使用代理
//
// 返回：
//   - *http.Response: 成功返回响应体指针
//   - error: 若请求多次失败则返回最后的错误
func (r *request) HttpGetWithRetry(requestURL string, timeout time.Duration, maxRetries int, retryInterval time.Duration, headers map[string]string, proxyURL string) (*http.Response, error) {
	// 配置 Transport，控制连接和响应头的超时时间
	var transport *http.Transport
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			logger.SlogWarnLocal(fmt.Sprintf("Invalid proxy URL: %v", err))
			return nil, err
		}
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
			DialContext: (&net.Dialer{
				Timeout:   timeout, // 连接超时
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   timeout, // TLS 握手超时
			ResponseHeaderTimeout: timeout, // 等响应头最大时间
		}
	} else {
		transport = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   timeout, // 连接超时
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   timeout, // TLS 握手超时
			ResponseHeaderTimeout: timeout, // 等响应头最大时间
		}
	}

	// 创建 HTTP 客户端（不设置整体超时）
	client := &http.Client{
		Transport: transport,
		Timeout:   0, // 不限制总耗时，避免 body 读取被中断
	}

	var resp *http.Response
	var err error

	// 重试机制
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, reqErr := http.NewRequest("GET", requestURL, nil)
		if reqErr != nil {
			return nil, reqErr
		}

		// 设置请求头
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		// 发起请求
		resp, err = client.Do(req)
		if err == nil {
			return resp, nil
		}

		// 请求失败，记录日志并重试
		logger.SlogWarnLocal(fmt.Sprintf("[%v] GET attempt %d failed: %v", requestURL, attempt, err))
		if attempt < maxRetries {
			logger.SlogWarnLocal(fmt.Sprintf("Retrying GET %v after %v...", requestURL, retryInterval))
			time.Sleep(retryInterval)
		}
	}

	return nil, err
}

// HttpPostWithRetry 发起一个带重试机制和代理支持的 HTTP POST 请求。
// 参数：
//   - requestURL: 请求地址
//   - body: 请求体（可以是 JSON、表单等）
//   - timeout: 每次请求的超时时间
//   - maxRetries: 最大重试次数
//   - retryInterval: 每次重试之间的等待间隔
//   - headers: 请求头（通常需设置 Content-Type）
//   - proxyURL: 代理地址，若为空则不使用代理
//
// 返回：
//   - *http.Response: 成功返回响应体指针
//   - error: 若请求多次失败则返回最后的错误
func (r *request) HttpPostWithRetry(requestURL string, body io.Reader, timeout time.Duration, maxRetries int, retryInterval time.Duration, headers map[string]string, proxyURL string) (*http.Response, error) {
	// 配置 Transport，控制连接和响应头的超时时间
	var transport *http.Transport
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			logger.SlogWarnLocal(fmt.Sprintf("Invalid proxy URL: %v", err))
			return nil, err
		}
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
			DialContext: (&net.Dialer{
				Timeout:   timeout, // TCP连接超时
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: timeout,
		}
	} else {
		transport = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: timeout,
		}
	}

	// 创建 HTTP 客户端（不设置总请求超时）
	client := &http.Client{
		Transport: transport,
		Timeout:   0,
	}

	var resp *http.Response
	var err error

	// 重试机制
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 注意：body 不能重复读取，所以如果需要重试，应在外部传入支持重复读取的 reader（如 bytes.Buffer）
		req, reqErr := http.NewRequest("POST", requestURL, body)
		if reqErr != nil {
			return nil, reqErr
		}

		// 设置请求头
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		// 发起请求
		resp, err = client.Do(req)
		if err == nil {
			return resp, nil
		}

		// 请求失败，记录日志并重试
		logger.SlogWarnLocal(fmt.Sprintf("[%v] POST attempt %d failed: %v", requestURL, attempt, err))
		if attempt < maxRetries {
			logger.SlogWarnLocal(fmt.Sprintf("Retrying POST %v after %v...", requestURL, retryInterval))
			time.Sleep(retryInterval)
		}
	}

	return nil, err
}

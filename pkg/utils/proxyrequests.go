// utils-------------------------------------
// @file      : proxyrequests.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/2/20 23:21
// -------------------------------------------

package utils

import (
	"crypto/tls"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/valyala/fasthttp"
	"net"
	"strings"
	"time"
)

type ProxyRequest struct {
	ProxyAddr string
	Client    *fasthttp.Client
}

var ProxyRequestsPool map[string]*ProxyRequest

type proxyRequest struct {
}

var ProxyRequests *proxyRequest

// InitializeProxyRequestsPool 初始化代理请求的管理
func InitializeProxyRequestsPool() {
	ProxyRequestsPool = make(map[string]*ProxyRequest)
	ProxyRequests = &proxyRequest{}
}

// CreateClientWithProxy 创建一个带代理的HTTP客户端
func CreateClientWithProxy(proxyAddr string) (*fasthttp.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // 不验证证书
	}

	// 通过代理设置 HTTP 客户端
	client := &fasthttp.Client{
		ReadTimeout:                   time.Second * 10,
		WriteTimeout:                  time.Second * 10,
		MaxIdleConnDuration:           time.Second * 10,
		NoDefaultUserAgentHeader:      true,
		DisableHeaderNamesNormalizing: true,
		DisablePathNormalizing:        true,
		TLSConfig:                     tlsConfig,
		Dial: func(addr string) (net.Conn, error) {
			// 创建代理的连接
			return net.Dial("tcp", proxyAddr)
		},
	}

	return client, nil
}

// GetProxyClient 根据代理地址获取客户端，若没有则新建
func GetProxyClient(proxyAddr string) (*fasthttp.Client, error) {
	// 如果该代理客户端已经创建过，直接返回
	if proxyRequest, exists := ProxyRequestsPool[proxyAddr]; exists {
		return proxyRequest.Client, nil
	}

	// 创建新客户端
	client, err := CreateClientWithProxy(proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("创建带代理的客户端失败: %v", err)
	}

	// 存储该代理和对应的客户端
	ProxyRequestsPool[proxyAddr] = &ProxyRequest{
		ProxyAddr: proxyAddr,
		Client:    client,
	}

	return client, nil
}
func (r *proxyRequest) HttpGetProxy(uri string, proxyAddr string) (types.HttpResponse, error) {
	client, err := GetProxyClient(proxyAddr)
	if err != nil {
		return types.HttpResponse{}, err
	}
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	// 最后需要归还req、resp到池中
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI(uri)

	if err := client.Do(req, resp); err != nil {
		return types.HttpResponse{}, err
	}
	tmp := types.HttpResponse{}
	tmp.Url = uri
	// 定义最大响应体大小 (10MB)
	const maxBodySize = 10 * 1024 * 1024

	// 截断 Body
	//body := resp.Body()
	//if len(body) > maxBodySize {
	//	body = body[:maxBodySize]
	//}
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
func (r *proxyRequest) HttpGetByteProxy(uri string, proxyAddr string) ([]byte, error) {
	client, err := GetProxyClient(proxyAddr)
	if err != nil {
		return []byte{}, err
	}
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	// 最后需要归还req、resp到池中
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI(uri)
	if err := client.Do(req, resp); err != nil {
		return make([]byte, 0), err
	}
	tmp := resp.Body()
	return tmp, nil
}

func (r *proxyRequest) HttpPostProxy(uri string, requestBody []byte, ct string, proxyAddr string) (error, *fasthttp.Response) {
	client, err := GetProxyClient(proxyAddr)
	if err != nil {
		return err, nil
	}
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

	if err := client.Do(req, resp); err != nil {
		return err, nil
	}
	return nil, resp
}

func (r *proxyRequest) HttpGetWithCustomHeaderProxy(uri string, customHeaders []string, proxyAddr string) (types.HttpResponse, error) {
	client, err := GetProxyClient(proxyAddr)
	if err != nil {
		return types.HttpResponse{}, err
	}
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

	if err := client.Do(req, resp); err != nil {
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

func (r *proxyRequest) HttpGetByteWithCustomHeaderProxy(uri string, customHeaders []string, proxyAddr string) ([]byte, error) {
	client, err := GetProxyClient(proxyAddr)
	if err != nil {
		return make([]byte, 0), err
	}
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

	if err := client.Do(req, resp); err != nil {
		return make([]byte, 0), err
	}
	tmp := resp.Body()
	return tmp, nil
}

func (r *proxyRequest) HttpPostWithCustomHeaderProxy(uri string, requestBody []byte, ct string, customHeaders []string, proxyAddr string) (error, *fasthttp.Response) {
	client, err := GetProxyClient(proxyAddr)
	if err != nil {
		return err, nil
	}
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

	if err := client.Do(req, resp); err != nil {
		return err, nil
	}
	return nil, resp
}

func (r *proxyRequest) HttpGetNoResProxy(uri string, proxyAddr string) error {
	client, err := GetProxyClient(proxyAddr)
	if err != nil {
		return err
	}
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	// 设置请求 URI
	req.SetRequestURI(uri)

	req.Header.SetMethod(fasthttp.MethodGet)

	// 发送请求
	if err := client.Do(req, resp); err != nil {
		return err
	}
	// 直接丢弃响应体（或者关闭它）
	resp.Reset()
	return nil
}

func (r *proxyRequest) HttpPostNoResProxy(uri string, requestBody []byte, ct string, proxyAddr string) error {
	client, err := GetProxyClient(proxyAddr)
	if err != nil {
		return err
	}
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

	if err := client.Do(req, resp); err != nil {
		return err
	}
	resp.Reset()
	return nil
}

func (r *proxyRequest) HttpGetNoResWithCustomHeaderProxy(uri string, customHeaders []string, proxyAddr string) error {
	client, err := GetProxyClient(proxyAddr)
	if err != nil {
		return err
	}
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
	if err := client.Do(req, resp); err != nil {
		return err
	}
	// 直接丢弃响应体（或者关闭它）
	resp.Reset()
	return nil
}

func (r *proxyRequest) HttpPostNoResWithCustomHeaderProxy(uri string, requestBody []byte, ct string, customHeaders []string, proxyAddr string) error {
	client, err := GetProxyClient(proxyAddr)
	if err != nil {
		return err
	}
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

	if err := client.Do(req, resp); err != nil {
		return err
	}
	resp.Reset()
	return nil
}

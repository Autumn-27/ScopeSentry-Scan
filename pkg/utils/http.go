// utils-------------------------------------
// @file      : http.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/27 22:06
// -------------------------------------------

package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// HttpClientConfig 提供 http.Client 配置选项
type HttpClientConfig struct {
	Timeout             time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
	TLSClientConfig     *tls.Config // 可自定义 TLS 配置（包括跳过验证、自定义 CA 等）
	ProxyURL            string
	FollowRedirect      bool
}

// GetHttpClient 返回一个配置好的 http.Client
func GetHttpClient(cfg HttpClientConfig) *http.Client {
	var proxyFunc func(*http.Request) (*url.URL, error)

	if cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err == nil {
			proxyFunc = http.ProxyURL(proxyURL)
		} else {
			proxyFunc = nil
		}
	}

	transport := &http.Transport{
		Proxy:               proxyFunc,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		TLSClientConfig:     cfg.TLSClientConfig,
	}

	client := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}

	// ✅ 如果不跟随跳转，设置 CheckRedirect
	if !cfg.FollowRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

type Nethttp struct {
	Client *http.Client
}

var GlobalNetHttp *Nethttp

func InitializeNetHttp() {
	GlobalNetHttp = &Nethttp{}
	GlobalNetHttp.Client = GetHttpClient(HttpClientConfig{
		Timeout:             5 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     60 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
	})
}

var (
	netHttpMap = make(map[string]*Nethttp)
	hmu        sync.Mutex
)

func getConfigKey(cfg HttpClientConfig) string {
	tlsSkipVerify := false
	if cfg.TLSClientConfig != nil {
		tlsSkipVerify = cfg.TLSClientConfig.InsecureSkipVerify
	}

	return fmt.Sprintf("p=%st=%dm=%dm=%di=%dt=%tf=%t",
		cfg.ProxyURL,
		int64(cfg.Timeout/time.Millisecond),
		cfg.MaxIdleConns,
		cfg.MaxIdleConnsPerHost,
		int64(cfg.IdleConnTimeout/time.Millisecond),
		tlsSkipVerify,
		cfg.FollowRedirect,
	)
}

// GetNetHttpByConfig 通过完整配置获取对应 Nethttp 实例，缓存复用
func GetNetHttpByConfig(cfg HttpClientConfig) *Nethttp {
	key := getConfigKey(cfg)

	hmu.Lock()
	defer hmu.Unlock()

	if instance, ok := netHttpMap[key]; ok {
		return instance
	}

	client := GetHttpClient(cfg)
	instance := &Nethttp{Client: client}

	netHttpMap[key] = instance
	return instance
}

func (n *Nethttp) HttpGetNoResWithCustomHeader(url string, customHeaders map[string]string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	for k, v := range customHeaders {
		req.Header.Set(k, v)
	}

	resp, err := n.Client.Do(req)
	if err != nil {
		//logger.SlogErrorLocal(fmt.Sprintf("HttpGetNoResWithCustomHeader func error: %v", err))
		return err
	}
	defer resp.Body.Close()

	// 丢弃响应体数据以复用连接
	io.Copy(io.Discard, resp.Body)
	return nil
}

func (n *Nethttp) HttpPostNoResWithCustomHeader(url string, body []byte, contentType string, customHeaders map[string]string) error {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	if contentType == "json" {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	for k, v := range customHeaders {
		req.Header.Set(k, v)
	}

	resp, err := n.Client.Do(req)
	if err != nil {
		//logger.SlogErrorLocal(fmt.Sprintf("HttpPostNoResWithCustomHeader func error: %v", err))
		return err
	}
	defer resp.Body.Close()

	// 丢弃响应体数据以复用连接
	io.Copy(io.Discard, resp.Body)
	return nil
}

func (n *Nethttp) HttpGetWithCustomHeader(uri string, customHeaders map[string]string) (types.HttpResponse, error) {
	// 创建请求
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return types.HttpResponse{}, err
	}

	// 添加自定义 Header
	for k, v := range customHeaders {
		req.Header.Set(k, v)
	}
	// 发起请求
	resp, err := n.Client.Do(req)
	if err != nil {
		return types.HttpResponse{}, err
	}
	defer resp.Body.Close()

	tmp := types.HttpResponse{}
	tmp.Url = uri

	// 限制最大响应体大小（10MB）
	const maxBodySize = 10 * 1024 * 1024 // 10MB
	limitedReader := io.LimitReader(resp.Body, maxBodySize)

	// 读取响应体
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("HttpGetWithCustomHeader func error: %v", err))
		return types.HttpResponse{}, err
	}
	tmp.Body = string(bodyBytes)
	tmp.StatusCode = resp.StatusCode

	// 获取重定向地址
	tmp.Redirect = resp.Header.Get("Location")

	// 设置 Content-Length
	if resp.ContentLength >= 0 {
		tmp.ContentLength = int(resp.ContentLength)
	} else {
		tmp.ContentLength = len(bodyBytes)
	}
	return tmp, nil
}

func (n *Nethttp) HttpPostWithCustomHeader(uri string, requestBody []byte, ct string, customHeaders map[string]string) (error, HttpResponse) {
	// 创建 POST 请求
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(requestBody))
	if err != nil {
		return err, HttpResponse{}
	}

	// 设置默认 Content-Type（仅当传入 ct == "json"）
	if ct == "json" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 设置自定义 Header
	for k, v := range customHeaders {
		req.Header.Set(k, v)
	}

	resp, err := n.Client.Do(req)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("HttpPostWithCustomHeader func error: %v", err))
		return err, HttpResponse{}
	}
	defer resp.Body.Close()

	// 限制最大响应体大小（10MB）
	const maxBodySize = 10 * 1024 * 1024 // 10MB
	limitedReader := io.LimitReader(resp.Body, maxBodySize)

	// 读取响应体
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return err, HttpResponse{}
	}

	// 提取响应头
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	res := HttpResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       bodyBytes,
	}
	return nil, res
}

// utils-------------------------------------
// @file      : clientpool.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/7/17 23:06
// -------------------------------------------

package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type ProxyPool struct {
	mu    sync.Mutex
	cache map[string]*url.URL
}

func NewProxyPool() *ProxyPool {
	return &ProxyPool{cache: make(map[string]*url.URL)}
}

func (p *ProxyPool) GetProxyURL(proxyStr string) (*url.URL, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if proxy, ok := p.cache[proxyStr]; ok {
		return proxy, nil
	}

	parsed, err := url.Parse(proxyStr)
	if err != nil {
		return nil, err
	}

	p.cache[proxyStr] = parsed
	return parsed, nil
}

var (
	clientMap = make(map[string]*http.Client)
	proxyPool = NewProxyPool()
	mu        sync.Mutex
)

func GetClient(proxyStr string, timeout time.Duration) (*http.Client, error) {
	mu.Lock()
	defer mu.Unlock()

	key := fmt.Sprintf("%s|%d", proxyStr, int(timeout.Seconds()))

	if client, ok := clientMap[key]; ok {
		return client, nil
	}

	var proxyFunc func(*http.Request) (*url.URL, error)

	if proxyStr == "" {
		proxyFunc = nil
	} else {
		proxyURL, err := proxyPool.GetProxyURL(proxyStr)
		if err != nil {
			return nil, err
		}
		proxyFunc = http.ProxyURL(proxyURL)
	}

	transport := &http.Transport{
		Proxy:               proxyFunc,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	clientMap[key] = client
	return client, nil
}

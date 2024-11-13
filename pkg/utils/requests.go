// utils-------------------------------------
// @file      : requests.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/19 20:48
// -------------------------------------------

package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/projectdiscovery/httpx/runner"
	"github.com/valyala/fasthttp"
	"net"
	"syscall"
	"time"
)

type request struct {
}

var Requests *request

var HttpClient *fasthttp.Client

func InitializeRequests() {
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

func (r *request) HttpPost(uri string, requestBody []byte, ct string) error {
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
	return nil
}

var dialer = &net.Dialer{
	Timeout: 2 * time.Second,
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

func (r *request) Httpx(Host string, resultCallback func(r types.AssetHttp), cdncheck string, screenshot bool) {
	gologger.DefaultLogger.SetMaxLevel(levels.LevelFatal) // increase the verbosity (optional)

	options := runner.Options{
		Methods:                   "GET",
		JSONOutput:                true,
		TLSProbe:                  true,
		Threads:                   30,
		RateLimit:                 100,
		InputTargetHost:           []string{Host},
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
		HostMaxErrors:             -1,
		StoreResponse:             false,
		StoreChain:                false,
		MaxResponseBodySizeToRead: 100000,
		Screenshot:                screenshot,
		ScreenshotTimeout:         5,
		//InputFile: "./targetDomains.txt", // path to file containing the target domains list
		OnResult: func(r runner.Result) {
			// handle error
			if r.Err != nil {
				logger.SlogDebugLocal(fmt.Sprintf("HttpxScan error %s: %s", r.Input, r.Err))
			} else {
				ah := Tools.HttpxResultToAssetHttp(r)
				//fmt.Printf("%s %s %d\n", r.Input, r.Host, r.StatusCode)
				resultCallback(ah)
			}
		},
	}
	if err := options.ValidateOptions(); err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("httpx options Validate error: %v", err))
	}

	httpxRunner, err := runner.New(&options)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("%v get error: %v", Host, err))
	}
	defer httpxRunner.Close()

	httpxRunner.RunEnumeration()
}

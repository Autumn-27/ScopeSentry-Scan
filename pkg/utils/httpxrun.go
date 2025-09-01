// utils-------------------------------------
// @file      : httpxrun.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/8/19 22:43
// -------------------------------------------

package utils

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/projectdiscovery/httpx/runner"
	"math"
	"time"
)

var HttpxRun *runner.Runner

func InitHttpx(cdncheck string, screenshot bool, screenshotTimeout int, tLSProbe bool, followRedirects bool, bypassHeader bool, threads int) {
	customHeaders := []string{}
	if bypassHeader {
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
		Threads:                   threads,
		RateLimit:                 100,
		Favicon:                   true,
		ExtractTitle:              true,
		TechDetect:                true,
		OutputWebSocket:           true,
		OutputServerHeader:        true,
		OutputIP:                  true,
		OutputCName:               true,
		DisableStdin:              true,
		ResponseHeadersInStdout:   true,
		ResponseInStdout:          true,
		Base64ResponseInStdout:    false,
		Jarm:                      true,
		OutputCDN:                 cdncheck,
		Location:                  false,
		HostMaxErrors:             10,
		StoreResponse:             false,
		StoreChain:                false,
		MaxResponseBodySizeToRead: math.MaxInt32,
		Screenshot:                screenshot,
		ScreenshotTimeout:         time.Duration(screenshotTimeout) * time.Second,
		Timeout:                   10,
		Wappalyzer:                Wappalyzer,
		DisableStdout:             true,
	}
	var err error
	HttpxRun, err = runner.New(&options)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("httpx get error: %v", err))
	}
	//defer httpxRunner.Close()
	//fmt.Println(options)
}

func RunAnalyze(target string, resultCallback func(r types.AssetHttp)) {
	resuFunc := func(r runner.Result) {
		if r.Host == "" {
			return
		}
		ah := Tools.HttpxResultToAssetHttp(r)

		resultCallback(ah)
	}
	HttpxRun.RunAnalyze(target, HttpxRun.HTTPX(), resuFunc)
}

func HttpxClose() {
	HttpxRun.Close()
}

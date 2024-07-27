// Package httpx_mode -----------------------------
// @file      : httpx-scan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/7 21:29
// -------------------------------------------
package httpxMode

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"github.com/cloudflare/cfssl/log"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/projectdiscovery/httpx/runner"
)

func HttpxScan(Host []string, resultCallback func(r types.AssetHttp)) {
	defer system.RecoverPanic("HttpxScan")
	gologger.DefaultLogger.SetMaxLevel(levels.LevelFatal) // increase the verbosity (optional)

	options := runner.Options{
		Methods:                   "GET",
		JSONOutput:                true,
		TLSProbe:                  true,
		InputTargetHost:           Host,
		Favicon:                   true,
		ExtractTitle:              true,
		TechDetect:                true,
		OutputWebSocket:           true,
		OutputServerHeader:        true,
		OutputIP:                  true,
		OutputCName:               true,
		ResponseHeadersInStdout:   true,
		ResponseInStdout:          true,
		Base64ResponseInStdout:    true,
		Jarm:                      true,
		OutputCDN:                 true,
		Location:                  false,
		HostMaxErrors:             -1,
		MaxResponseBodySizeToRead: 100000,
		//InputFile: "./targetDomains.txt", // path to file containing the target domains list
		OnResult: func(r runner.Result) {
			// handle error
			if r.Err != nil {
				system.SlogDebugLocal(fmt.Sprintf("HttpxScan error %s: %s", r.Input, r.Err))
			} else {
				ah := httpxResultToAssetHttp(r)
				//fmt.Printf("%s %s %d\n", r.Input, r.Host, r.StatusCode)
				resultCallback(ah)
			}
		},
	}

	httpxRunner, err := runner.New(&options)
	if err != nil {
		log.Fatal(err)
	}
	defer httpxRunner.Close()

	httpxRunner.RunEnumeration()
}

func httpxResultToAssetHttp(r runner.Result) types.AssetHttp {
	var ah = types.AssetHttp{
		Timestamp:    system.GetTimeNow(),
		TLSData:      r.TLSData, // You may need to set an appropriate default value based on the actual type.
		Hashes:       r.Hashes,
		CDNName:      r.CDNName,
		Port:         r.Port,
		URL:          r.URL,
		Title:        r.Title,
		Type:         r.Scheme,
		Error:        r.Error,
		ResponseBody: r.ResponseBody,
		Host:         r.Host,
		FavIconMMH3:  r.FavIconMMH3,
		FaviconPath:  r.FaviconPath,
		RawHeaders:   r.RawHeaders,
		Jarm:         r.Jarm,
		Technologies: r.Technologies, // You may need to set an appropriate default value based on the actual type.
		StatusCode:   r.StatusCode,   // You may need to set an appropriate default value.
		Webcheck:     false,
		IconContent:  r.IconContent,
		WebServer:    r.WebServer,
	}
	return ah

}

func HttpSurvival(target string) (int, int, string, error) {
	defer system.RecoverPanic("HttpSurvival")
	gologger.DefaultLogger.SetMaxLevel(levels.LevelFatal)
	var StatusCode int
	var ContentLength int
	var RespBody string
	var err error
	httpxResultsHandler := func(code int, length int, respBody string, e error) {
		if e != nil {
			err = e
		}
		StatusCode = code
		ContentLength = length
		RespBody = respBody
	}
	options := runner.Options{
		Methods:                   "GET",
		JSONOutput:                false,
		TLSProbe:                  false,
		InputTargetHost:           []string{target},
		Favicon:                   false,
		ExtractTitle:              false,
		TechDetect:                false,
		OutputWebSocket:           false,
		OutputServerHeader:        false,
		OutputIP:                  false,
		OutputCName:               false,
		ResponseHeadersInStdout:   false,
		ResponseInStdout:          false,
		Base64ResponseInStdout:    false,
		Jarm:                      false,
		OutputCDN:                 false,
		Location:                  false,
		Timeout:                   3,
		Hashes:                    "",
		HostMaxErrors:             -1,
		MaxResponseBodySizeToRead: 100000,
		OnResult: func(r runner.Result) {
			if r.Err != nil {
				httpxResultsHandler(0, 0, "", r.Err)
			} else {
				httpxResultsHandler(r.StatusCode, r.ContentLength, r.ResponseBody, nil)
			}
		},
	}

	httpxRunner, err := runner.New(&options)
	if err != nil {
		log.Fatal(err)
	}
	defer httpxRunner.Close()

	httpxRunner.RunEnumeration()
	return StatusCode, ContentLength, RespBody, err
}

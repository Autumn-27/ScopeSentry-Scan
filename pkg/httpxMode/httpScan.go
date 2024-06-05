// Package httpx_mode -----------------------------
// @file      : httpx-scan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/7 21:29
// -------------------------------------------
package httpxMode

import (
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
		OutputIP:                  true,
		OutputCName:               false,
		ResponseHeadersInStdout:   true,
		ResponseInStdout:          true,
		Base64ResponseInStdout:    true,
		Jarm:                      true,
		OutputCDN:                 true,
		Location:                  true,
		HostMaxErrors:             -1,
		MaxResponseBodySizeToRead: 100000,
		//InputFile: "./targetDomains.txt", // path to file containing the target domains list
		OnResult: func(r runner.Result) {
			// handle error
			if r.Err != nil {
				//system.SlogErrorLocal(fmt.Sprintf("HttpxScan error %s: %s", r.Input, r.Err))
			} else {
				ah := httpxResultToAssetHttp(r)
				//fmt.Printf("%s %s %d\n", r.Input, r.Host, r.StatusCode)
				resultCallback(ah)
			}
		},
	}

	if err := options.ValidateOptions(); err != nil {
		log.Fatal(err)
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
	}
	return ah

}

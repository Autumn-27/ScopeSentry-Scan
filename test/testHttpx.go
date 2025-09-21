// Package main -----------------------------
// @file      : testHttpx.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/11 21:22
// -------------------------------------------
package main

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
)

func main() {
	logger.NewLogger()
	utils.InitializeRequests()
	//gologger.DefaultLogger.SetMaxLevel(levels.LevelWarning)
	httpxResultsHandler := func(r types.AssetHttp) {
		fmt.Printf("%v %v\n", r.URL, r.Screenshot)
	}
	utils.Requests.Httpx([]string{"https://www.baidu.com/"}, httpxResultsHandler, "true", true, 10, true, true, context.Background(), 10, false)
	//StatusCode, ContentLength, err := httpxMode.HttpSurvival("https://b31dadwaaidu.com")
	//fmt.Println(StatusCode, ContentLength, err)
	//options := runner.Options{
	//	Methods:                   "GET",
	//	JSONOutput:                false,
	//	TLSProbe:                  false,
	//	InputTargetHost:           []string{"https://baidu.com"},
	//	Favicon:                   false,
	//	ExtractTitle:              false,
	//	TechDetect:                false,
	//	OutputWebSocket:           false,
	//	OutputServerHeader:        false,
	//	OutputIP:                  false,
	//	OutputCName:               false,
	//	ResponseHeadersInStdout:   false,
	//	ResponseInStdout:          false,
	//	Base64ResponseInStdout:    false,
	//	Jarm:                      false,
	//	OutputCDN:                 false,
	//	Location:                  false,
	//	Hashes:                    "",
	//	HostMaxErrors:             -1,
	//	MaxResponseBodySizeToRead: 100000,
	//	OnResult: func(r runner.Result) {
	//		fmt.Printf("%v %v", r.StatusCode, r.ContentLength)
	//	},
	//}
	//
	//if err := options.ValidateOptions(); err != nil {
	//}
	//
	//httpxRunner, err := runner.New(&options)
	//if err != nil {
	//}
	//defer httpxRunner.Close()
	//
	//httpxRunner.RunEnumeration()
}

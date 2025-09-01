// main-------------------------------------
// @file      : testm.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/8/19 21:54
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/projectdiscovery/httpx/runner"
	"log"
	"net/url"
	"strings"
)

func main() {
	options := runner.Options{
		Methods: "GET",
		OnResult: func(r runner.Result) {
			// handle error
			if r.Err != nil {
				fmt.Printf("[Err] %s: %s\n", r.Input, r.Err)
				return
			}
			fmt.Printf("%s %s %d\n", r.Input, r.Host, r.StatusCode)
		},
	}

	httpxRunner, err := runner.New(&options)
	if err != nil {
		log.Fatal(err)
	}
	defer httpxRunner.Close()
	httpxRunner.RunAnalyze("docs.babel.baidu.com", httpxRunner.HTTPX(), func(result runner.Result) {
		// 解析 URL
		parsedUrl, err := url.Parse(result.URL)
		if err != nil {
		}

		// 提取域名
		domain := parsedUrl.Host

		// 去掉端口号（如果有）
		domain = strings.Split(domain, ":")[0]

	})
}

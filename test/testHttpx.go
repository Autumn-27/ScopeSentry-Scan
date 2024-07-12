// Package main -----------------------------
// @file      : testHttpx.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/11 21:22
// -------------------------------------------
package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/httpxMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
)

func main() {
	domainList := []string{"https://baidu.com"}
	httpxResultsHandler := func(r types.AssetHttp) {
		fmt.Println(r)
	}
	httpxMode.HttpxScan(domainList, httpxResultsHandler)
}

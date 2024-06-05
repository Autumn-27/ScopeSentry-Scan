// main-------------------------------------
// @file      : testHttpDoGet.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/23 20:12
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
)

func main() {
	util.InitHttpClient()
	resp, _ := util.HttpGet("https://www.hackerone.com/")
	fmt.Println(resp)
}

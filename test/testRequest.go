// main-------------------------------------
// @file      : testRequest.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/5/28 21:32
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"strings"
)

func main() {
	util.InitHttpClient()
	_, err := util.HttpGet("http://test:666/")
	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "timed out") {
			fmt.Println("dddd")
		}
	}

}

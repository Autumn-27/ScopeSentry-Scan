// main-------------------------------------
// @file      : testWayback.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/11 21:30
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlscan/wayback/source"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
)

func main() {
	//util.InitHttpClient()
	//waybackMode.Scan([]string{"test.top"})
	//res, err := util.HttpGetByte("http://test:666/")
	//if err != nil {
	//}
	//fmt.Println(res)
	config.Initialize()
	utils.InitializeTools()
	utils.InitializeDnsTools()
	utils.InitializeRequests()
	utils.InitializeResults()
	result := make(chan source.Result, 100)
	go func() {
		for res := range result {
			fmt.Println(res.URL)
		}
	}()
	source.CommoncrawlRun("baidu.com", result)
}

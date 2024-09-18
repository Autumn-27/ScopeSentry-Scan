// main-------------------------------------
// @file      : testDns.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/17 16:48
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/projectdiscovery/dnsx/libs/dnsx"
	"time"
)

func main() {
	start := time.Now()
	// Create DNS Resolver with default options
	dnsClient, err := dnsx.New(dnsx.DefaultOptions)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	//DNS A question and returns corresponding IPs
	//result, err := dnsClient.Lookup("att-globys.idp.blogin.att.com")
	//if err != nil {
	//	fmt.Printf("err: %v\n", err)
	//	return
	//}
	//for idx, msg := range result {
	//	fmt.Printf("%d: %s\n", idx+1, msg)
	//}

	// Query
	rawResp, err := dnsClient.QueryOne("dwassentrydwasdweqewqe.dwas")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	//fmt.Printf("rawResp: %v\n", rawResp)

	jsonStr, err := rawResp.JSON()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fmt.Println(jsonStr)
	// 记录程序结束时间
	end := time.Now()

	// 计算时间差
	duration := end.Sub(start)

	// 打印程序运行时间
	fmt.Printf("程序运行时间: %v\n", duration)
	return
}

// main-------------------------------------
// @file      : testDns.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/17 16:48
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"time"
)

func main() {
	start := time.Now()
	// Create DNS Resolver with default options
	//dnsClient, err := dnsx.New(dnsx.DefaultOptions)
	//if err != nil {
	//	fmt.Printf("err: %v\n", err)
	//	return
	//}

	//DNS A question and returns corresponding IPs
	//result, err := dnsClient.Lookup("att-globys.idp.blogin.att.com")
	//if err != nil {
	//	fmt.Printf("err: %v\n", err)
	//	return
	//}
	//for idx, msg := range result {
	//	fmt.Printf("%d: %s\n", idx+1, msg)
	//}
	utils.InitializeDnsTools()
	// Query
	// 定义一个数组存储域名
	domains := []string{
		"dwassentrydwasdweqewqe.dwas",
		"example.com",
		"anotherdomain.test",
	}

	// 循环遍历域名数组
	for _, domain := range domains {
		resultDns := utils.DNS.QueryOne(domain)
		tmp := utils.DNS.DNSdataToSubdomainResult(resultDns)
		fmt.Printf("Domain: %s, Response: %s\n", domain, tmp)
	}
	// 记录程序结束时间
	end := time.Now()

	// 计算时间差
	duration := end.Sub(start)

	// 打印程序运行时间
	fmt.Printf("程序运行时间: %v\n", duration)
	return
}

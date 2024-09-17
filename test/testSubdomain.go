// main-------------------------------------
// @file      : testSubdomain.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/5/30 0:01
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/subdomainMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"time"
)

func main() {
	system.SetUp()
	//subdomainMode.SubDomainRunner("test")
	start := time.Now()
	result := subdomainMode.Verify2([]string{"oauth.idp.blogin.att.com"})
	fmt.Printf("%v", result)
	// 记录程序结束时间
	end := time.Now()

	// 计算时间差
	duration := end.Sub(start)

	// 打印程序运行时间
	fmt.Printf("程序运行时间: %v\n", duration)
}

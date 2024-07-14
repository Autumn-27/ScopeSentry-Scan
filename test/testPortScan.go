// Package main -----------------------------
// @file      : testPortScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/13 18:55
// -------------------------------------------
package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/portScanMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
)

func main() {
	//portScanMode.PortScan("test", "80")
	system.Test()
	system.CheckRustscan()
	fmt.Println(system.GetTimeNow())
	portScanMode.PortScan2("39.125.123.24", "1-65535", "port")
}

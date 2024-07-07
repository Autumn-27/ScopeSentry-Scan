// Package main -----------------------------
// @file      : testPortScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/13 18:55
// -------------------------------------------
package main

import (
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/portScanMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
)

func main() {
	//portScanMode.PortScan("test", "80")
	system.Test()
	system.CheckRustscan()
	portScanMode.PortScan2("39.105.160.88", "1-65535", "port")
}

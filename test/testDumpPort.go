// main-------------------------------------
// @file      : testDumpPort.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/6/5 18:55
// -------------------------------------------

package main

import (
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
)

func main() {
	system.Test()
	scanResult.GetPortByHost("test.top")
}

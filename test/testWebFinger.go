// main-------------------------------------
// @file      : testWebFinger.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/14 12:09
// -------------------------------------------

package main

import (
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
)

func main() {
	system.SetUp()
	tmp := types.AssetHttp{}
	scanResult.WebFingerScan(tmp)
}

// main-------------------------------------
// @file      : testVulnScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/12 21:39
// -------------------------------------------

package main

import (
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/vulnMode"
)

func main() {
	system.SetUp()
	vulnMode.Scan([]string{"http://test:666/"}, []string{"*"})
}

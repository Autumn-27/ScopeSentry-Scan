// main-------------------------------------
// @file      : testSubdomain.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/5/30 0:01
// -------------------------------------------

package main

import (
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/subdomainMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
)

func main() {
	system.SetUp()
	subdomainMode.SubDomainRunner("test")
}

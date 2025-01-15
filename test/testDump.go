// main-------------------------------------
// @file      : testDump.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/6/1 20:15
// -------------------------------------------

package main

import "github.com/Autumn-27/ScopeSentry-Scan/internal/results"

//import (
//	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
//)
//
//func main() {
//	system.SetUp()
//}

func main() {
	results.InitializeDuplicate()
	results.Duplicate.URLParams("http://ads.com?a=dwasd")
	results.Duplicate.URLParams("http://ads.com/?a=dwasdw")
	results.Duplicate.URLParams("http://ads.com/?a=dwasdw&d=sdwa")
}

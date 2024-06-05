// Package main -----------------------------
// @file      : testRandom.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/5/30 16:56
// -------------------------------------------
package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
)

func main() {
	for i := 0; i < 2; i++ {
		subdomain := util.GenerateRandomString(6)
		fmt.Println(subdomain)
	}
}

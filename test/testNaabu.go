// Package main -----------------------------
// @file      : testNaabu.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/13 20:14
// -------------------------------------------
package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/portScanMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
)

func main() {

	var PortAlives []types.PortAlive
	CallBack := func(alive []types.PortAlive) {
		PortAlives = alive
	}
	err := portScanMode.NaabuScan([]string{"127.0.0.1"}, "80,443,1555,666", CallBack)
	if err != nil {
		fmt.Printf(err.Error())
	}

	for _, value := range PortAlives {
		fmt.Printf(value.Host)
	}
}

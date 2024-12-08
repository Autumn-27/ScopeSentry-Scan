// main-------------------------------------
// @file      : testParseFlag.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/12/8 22:26
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
)

func main() {
	args, err := utils.Tools.ParseArgs("", "a", "dnslog")
	if err != nil {
		return
	}
	fmt.Println(args)
}

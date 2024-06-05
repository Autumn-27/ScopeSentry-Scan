// main-------------------------------------
// @file      : testdiff.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/6/3 19:28
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/dirScanMode/utils"
)

func main() {
	sm := utils.NewSequenceMatcher("dfhiaulkdbgvywighdvwuqiadhiuwa", "ed3qwdawefcv3QR3WTN5REYN45EA4W")
	ration := sm.Ratio()
	fmt.Println(ration)
}

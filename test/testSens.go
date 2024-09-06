// main-------------------------------------
// @file      : testSens.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/7/28 18:17
// -------------------------------------------

package main

import (
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
)

func main() {
	system.Test()
	system.UpdateSensitive()
	scanResult.UrlResult([]types.UrlResult{types.UrlResult{
		Input:  "https://promo.indrive.com/10ridestogetprize_ru/random",
		Output: "https://promo.indrive.com/10ridestogetprize_ru/random",
	}}, "xxxxx", true)
}

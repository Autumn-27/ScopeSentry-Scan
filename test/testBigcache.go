// main-------------------------------------
// @file      : testBigcache.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/18 21:43
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/bigcache"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
)

func main() {
	err := bigcache.Initialize()
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("bigcache Initialize error: %v", err))
		return
	}
	bigcache.BigCache.Set("dddd", []byte{})
	get, err := bigcache.BigCache.Get("dddd")
	if err != nil {
		fmt.Printf("%v error", err)
		return
	}
	fmt.Println(get)
}

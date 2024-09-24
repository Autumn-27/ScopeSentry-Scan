// main-------------------------------------
// @file      : testUtils.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/24 20:15
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
)

func main() {
	utils.InitializeRequests()
	getByte, err := utils.Requests.HttpGetByte("http://feedback.wooga.com.herokudns.com")
	if err != nil {
		return
	}
	s := string(getByte)
	fmt.Println(s)
}

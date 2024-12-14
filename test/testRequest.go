// main-------------------------------------
// @file      : testRequest.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/5/28 21:32
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"strings"
)

func main() {
	utils.InitializeRequests()
	_, err := utils.Requests.HttpGetByteWithCustomHeader("", []string{"True-Client-Ip: 127.0.0.1", "Via: 127.0.0.1"})
	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "timed out") {
			fmt.Println("dddd")
		}
	}
}

// subdomainTakeoverMode-------------------------------------
// @file      : subdomainTakeover.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/10 22:24
// -------------------------------------------

package subdomainTakeoverMode

import (
	"crypto/tls"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"io/ioutil"
	"net/http"
	"strings"
)

func Scan(input string, tg []string, taskId string) {
	defer system.RecoverPanic("subdomainTakeoverMode")
	var SubTakerRes []types.SubTakeResult
	for _, t := range tg {
		for _, finger := range system.SubdomainTakerFingers {
			for _, c := range finger.Cname {
				if strings.Contains(t, c) {
					body := sendhttp(t)
					for _, resp := range finger.Response {
						if strings.Contains(body, resp) {
							resultTmp := types.SubTakeResult{}
							resultTmp.Input = input
							resultTmp.Value = t
							resultTmp.Cname = c
							resultTmp.Response = resp
							SubTakerRes = append(SubTakerRes, resultTmp)
						}
					}

				}
			}
		}
	}
	if len(SubTakerRes) != 0 {
		scanResult.SubTakerResult(SubTakerRes, taskId)
	}
}

func sendhttp(tg string) string {
	// 发送HTTP请求
	resp, err := http.Get("http://" + tg)
	if err != nil {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		resp, err = client.Get("https://" + tg)
		if err != nil {
			fmt.Println("发送HTTP和HTTPS请求失败:", err)
			return ""
		}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应内容失败:", err)
		return ""
	}
	return string(body)
}

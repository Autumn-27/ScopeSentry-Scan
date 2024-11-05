// runner-------------------------------------
// @file      : pagemonitoring.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/11/5 20:51
// -------------------------------------------

package runner

import (
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"strings"
)

func PageMonitoringRunner(targets []string) {
	for _, target := range targets {
		var pageMonitResult types.PageMonit
		err := json.Unmarshal([]byte(target), &pageMonitResult)
		if err != nil {
			logger.SlogWarnLocal(fmt.Sprintf("PageMonitoringRunner json.Unmarshal error: %v", err))
			continue
		}
	}
}

func Handler(pageMonitResult types.PageMonit) {
	response, err := utils.Requests.HttpGet(pageMonitResult.Url)
	if err != nil {
		return
	}
	flag := utils.Tools.IsSuffixURL(pageMonitResult.Url, ".js")
	if flag {
		if strings.Contains(response.Body, "<!DOCTYPE html>") {
			response.StatusCode = 0
			response.Body = ""
		}
	}
	// 如果状态码相同，比较响应体hash是否相同，不同则计算两个响应体的相似度，记录相似度的值
	// 并且判断 响应体hash是一个还是两个，如果是一个，那直接加入hash 存入body
	// 如果是两个，那么删除第一个hash和对应的body，增加第二个hash和对应的body， 这里考虑一下存储body是使用body的hash还是url的hash

	// 状态码不相同，记录状态码，设置相似度为0

}

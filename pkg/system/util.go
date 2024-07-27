// Package system -----------------------------
// @file      : util.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/5/10 17:02
// -------------------------------------------
package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func RecoverPanic(tg string) {
	if r := recover(); r != nil {
		SlogError(fmt.Sprintf("%s recover panic", tg))
		SlogErrorLocal(fmt.Sprint(r))
	}
}

var rateLimiter = time.Tick(2 * time.Second)

func SendNotification(msg string) {
	<-rateLimiter
	for _, api := range NotificationApi {
		uri := strings.Replace(api.Url, "*msg*", msg, -1)
		if api.Method == "GET" {
			_, err := util.HttpGet(uri)
			if err != nil {
				SlogError(fmt.Sprintf("SendNotification HTTP Get Error: %s", uri))
			}
		} else {
			data := strings.Replace(api.Data, "*msg*", msg, -1)
			err := util.HttpPost(uri, []byte(data), api.ContentType)
			if err != nil {
				SlogError(fmt.Sprintf("SendNotification HTTP Post Error: %s", uri))
			}
		}
	}
}

func HTTPPostGetData(url string, jsonData interface{}) (map[string]interface{}, error) {
	// 将数据转换为 JSON 格式
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	// 发送 POST 请求
	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(response.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func IsMatchingFilter(fs []*regexp.Regexp, d []byte) bool {
	for _, r := range fs {
		if r.Match(d) {
			return true
		}
	}
	return false
}

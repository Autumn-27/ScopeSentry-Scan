// utils-------------------------------------
// @file      : fuzz.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/8/7 22:04
// -------------------------------------------

package utils

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"
)

func GetKeyValue(key string, id string, tp string, dnslog string) string {
	return key + "=" + url.QueryEscape("http://"+id+"."+key+"."+tp+"."+dnslog)
}

func GetJsonValue(key string, id string, dnslog string) string {
	return "http://" + id + "." + key + ".3." + dnslog
}

var cfg = HttpClientConfig{
	Timeout:             3 * time.Second,
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 50,
	IdleConnTimeout:     60 * time.Second,
	TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
	FollowRedirect:      false,
}

var client = GetNetHttpByConfig(cfg)

func FuzzGet(uri string, id string, header map[string]string, wg *sync.WaitGroup, testParam []string, dnslog string) {
	defer wg.Done()
	var parameterVal string
	for _, key := range testParam {
		if key == "" {
			continue
		}
		parameterVal += "&" + GetKeyValue(key, id, "1", dnslog)
		if len(parameterVal) > 1000 {
			parameterValTmp := parameterVal[1:]
			go func() {
				FuzzHttpGetNoResWithCustomHeader(uri+"?"+parameterValTmp, header)
			}()
			parameterVal = ""
		}
	}
	if parameterVal != "" {
		parameterValTmp := parameterVal[1:]
		go func() {
			FuzzHttpGetNoResWithCustomHeader(uri+"?"+parameterValTmp, header)
		}()
	}
}

func FuzzPostJson(uri string, id string, header map[string]string, wg *sync.WaitGroup, testParam []string, dnslog string) {
	defer wg.Done()
	params := make(map[string]string)
	for _, key := range testParam {
		if key == "" {
			continue
		}
		params[key] = GetJsonValue(key, id, dnslog)
	}
	jsonData, err := json.Marshal(params)
	if err != nil {
		// 处理 JSON 序列化错误
		fmt.Println("JSON 序列化失败:", err)
		return
	}
	go func() {
		FuzzHttpPostNoResWithCustomHeader(uri, jsonData, "json", header)
	}()
}

func FuzzPost(uri string, id string, header map[string]string, wg *sync.WaitGroup, testParam []string, dnslog string) {
	defer wg.Done()
	var parameterVal string
	for _, key := range testParam {
		if key == "" {
			continue
		}
		parameterVal += "&" + GetKeyValue(key, id, "2", dnslog)
		if len(parameterVal) > 1024*1020 {
			parameterValTmp := parameterVal[1:]
			go func() {
				FuzzHttpPostNoResWithCustomHeader(uri, []byte(parameterValTmp), "url", header)
			}()
			parameterVal = ""
		}
	}
	if parameterVal != "" {
		parameterValTmp := parameterVal[1:]
		go func() {
			FuzzHttpPostNoResWithCustomHeader(uri, []byte(parameterValTmp), "url", header)
		}()
	}
}

func FuzzHttpGetNoResWithCustomHeader(url string, header map[string]string) {
	err := client.HttpGetNoResWithCustomHeader(url, header)
	if err != nil {
	}
}

func FuzzHttpPostNoResWithCustomHeader(url string, body []byte, contentType string, customHeaders map[string]string) {
	err := client.HttpPostNoResWithCustomHeader(url, body, contentType, customHeaders)
	if err != nil {
	}
}

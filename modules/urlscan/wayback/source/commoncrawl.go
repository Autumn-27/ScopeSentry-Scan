// source-------------------------------------
// @file      : commoncrawl.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/14 20:20
// -------------------------------------------

package source

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	indexURL     = "https://index.commoncrawl.org/collinfo.json"
	maxYearsBack = 5
)

var year = time.Now().Year()

func CommoncrawlRun(rootUrl string, result chan Result, ctx context.Context) int {
	bodyBytes, err := utils.Requests.HttpGetByte(indexURL)
	if err != nil {
		return 0
	}

	var indexes []IndexResponse
	if err := json.Unmarshal(bodyBytes, &indexes); err != nil {
		logger.SlogWarnLocal(fmt.Sprintf("commoncrawl jsondecode error: %v data: %v", err, string(bodyBytes)))
		return 0
	}

	years := make([]string, 0)
	for i := 0; i < maxYearsBack; i++ {
		years = append(years, strconv.Itoa(year-i))
	}

	searchIndexes := make(map[string]string)
	for _, year := range years {
		for _, index := range indexes {
			if strings.Contains(index.ID, year) {
				if _, ok := searchIndexes[year]; !ok {
					searchIndexes[year] = index.APIURL
					break
				}
			}
		}
	}
	quantity := 0
	// 遍历符合年份的API地址，获取URL
	for _, apiURL := range searchIndexes {
		// 在每次调用 getURLs 时传递上下文
		select {
		case <-ctx.Done():
			// 如果上下文被取消，则提前退出
			return quantity
		default:
			further, n := getURLs(apiURL, rootUrl, result, ctx)
			if !further {
				break
			}
			quantity += n
		}
	}
	return quantity
}

func getURLs(searchURL, rootURL string, result chan Result, ctx context.Context) (bool, int) {
	currentSearchURL := fmt.Sprintf("%s?url=*.%s&output=text&fl=url", searchURL, rootURL)
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置超时时间
	}
	res, err := client.Get(currentSearchURL)
	if err != nil {
		return false, 0
	}
	if res.StatusCode != 200 {
		return false, 0
	}
	defer res.Body.Close()
	sc := bufio.NewScanner(res.Body)
	buffseSize := 4 * 1024
	buf := make([]byte, buffseSize)
	sc.Buffer(buf, buffseSize)
	lineCount := 0
	for sc.Scan() {
		select {
		case result <- Result{
			URL:    sc.Text(),
			Source: "commoncrawl",
		}:
			lineCount++
		case <-ctx.Done():
			// 如果上下文被取消，则提前退出
			return false, lineCount
		}
	}
	return true, lineCount
}

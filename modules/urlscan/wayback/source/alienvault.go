// source-------------------------------------
// @file      : alienvault.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/14 21:12
// -------------------------------------------

package source

import (
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"golang.org/x/net/context"
)

func AlienvaultRun(rootUrl string, result chan Result, ctx context.Context) int {
	page := 1
	lineCount := 0
	for {
		select {
		case <-ctx.Done():
			return lineCount
		default:
			apiURL := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/url_list?page=%d&limit=100", rootUrl, page)
			bodyBytes, err := utils.Requests.HttpGetByte(apiURL)
			if err != nil {
				return lineCount
			}

			var response AlienvaultResponse
			responseData := append([]byte(nil), bodyBytes...)
			// Get the response body and decode
			if err := json.Unmarshal(responseData, &response); err != nil {
				logger.SlogWarnLocal(fmt.Sprintf("Alienvault jsondecode error: %v", err))
				return lineCount
			}

			for _, record := range response.URLList {
				lineCount++
				result <- Result{URL: record.URL, Source: "alienvault"}
			}

			if !response.HasNext {
				return lineCount
			}
			page++
		}
	}
}

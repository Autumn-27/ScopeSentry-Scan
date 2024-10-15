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
)

func AlienvaultRun(rootUrl string, result chan Result) int {
	page := 1
	lineCount := 0
	for {
		apiURL := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/url_list?page=%d&limit=100", rootUrl, page)
		bodyBytes, err := utils.Requests.HttpGetByte(apiURL)
		if err != nil {
			return 0
		}

		var response AlienvaultResponse
		// Get the response body and decode
		if err := json.Unmarshal(bodyBytes, &response); err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("Alienvault jsondecode error: %v", err))
			return 0
		}

		for _, record := range response.URLList {
			lineCount++
			result <- Result{URL: record.URL, Source: "alienvault"}
		}

		if !response.HasNext {
			break
		}
		page++
	}
	return lineCount
}

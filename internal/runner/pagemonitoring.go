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
)

func PageMonitoringRunner(targets []string) {
	for _, target := range targets {
		var pageMonitResult types.PageMonitResult
		err := json.Unmarshal([]byte(target), &pageMonitResult)
		if err != nil {
			logger.SlogWarnLocal(fmt.Sprintf("PageMonitoringRunner json.Unmarshal error: %v", err))
			continue
		}

	}
}

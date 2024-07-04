// runner-------------------------------------
// @file      : pageMonitor.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/23 20:08
// -------------------------------------------

package runner

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
)

func PageMRun(target string) {
	system.SlogInfoLocal(fmt.Sprintf("Page monitoring start: %s", target))
	tmp, err := util.HttpGet(target)
	if err != nil {
		system.SlogDebugLocal(fmt.Sprintf("%v PageMRun HttpGet error: %s", target, err))
	}
	t := []types.TmpPageMonitResult{}
	t = append(t,
		types.TmpPageMonitResult{
			Url:     tmp.Url,
			Content: tmp.Body,
		})
	scanResult.PageMonitoringInitResult(t, "")
	system.SlogInfoLocal(fmt.Sprintf("Page monitoring end: %s", target))
}

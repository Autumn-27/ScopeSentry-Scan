// dirScanMode-------------------------------------
// @file      : dirScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/11 18:39
// -------------------------------------------

package dirScanMode

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/dirScanMode/core"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/dirScanMode/runner"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"path/filepath"
	"strconv"
)

func Scan(urls []string, taskId string) {
	defer system.RecoverPanic("DirScan")
	//if len(system.DirDict) == 0 {
	//	system.UpdateDirDicConfig()
	//}
	NotificationMsg := "DirScan Result:\n"
	resultHandle := func(response types.HttpResponse) {
		if response.Redirect != "" {
			NotificationMsg += fmt.Sprintf("%v - %v -%v\n", response.Url, response.StatusCode, response.Redirect)
			//system.SlogInfo(fmt.Sprintf("%v - %v -%v", response.Url, response.StatusCode, response.Redirect))
		} else {
			NotificationMsg += fmt.Sprintf("%v - %v\n", response.Url, response.StatusCode)
			//system.SlogInfo(fmt.Sprintf("%v - %v", response.Url, response.StatusCode))
		}
		scanResult.DirResult([]types.DirResult{{
			Url:    response.Url,
			Status: response.StatusCode,
			Msg:    response.Redirect,
			Length: response.ContentLength,
			TaskId: taskId,
		}})
	}
	if len(urls) != 0 {
		dirDicConfigPath := filepath.Join(system.ConfigDir, "dirdict")
		controller := runner.Controller{Targets: urls, Dictionary: dirDicConfigPath}
		DirscanThread, err := strconv.Atoi(system.AppConfig.System.DirscanThread)
		if err != nil {
			DirscanThread = 10
		}
		op := core.Options{
			Extensions:    []string{"php", "aspx", "jsp", "html", "js"},
			Thread:        DirscanThread,
			MatchCallback: resultHandle,
		}
		controller.Run(op)
		if system.NotificationConfig.DirScanNotification && len(NotificationMsg) > 20 {
			go system.SendNotification(NotificationMsg)
		}
	}
}

// task-------------------------------------
// @file      : runner.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/9 21:47
// -------------------------------------------

package runner

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handler"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/modules"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"runtime"
	"sync"
	"time"
)

func Run(op options.TaskOptions) error {
	// 清空临时文件
	defer CleanTmp()
	var wg sync.WaitGroup
	var start time.Time
	var end time.Time
	start = time.Now()
	handler.TaskHandle.StartTask()
	handler.TaskHandle.ProgressStart("scan", op.Target, op.ID, 1)
	op.ModuleRunWg = &wg
	op.TargetParser = append(op.TargetParser, "7bbaec6487f51a9aafeff4720c7643f0")
	process := modules.CreateScanProcess(&op)
	ch := make(chan interface{})
	process.SetInput(ch)
	go func() {
		err := process.ModuleRun()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}()
	ch <- op.Target
	close(ch)
	time.Sleep(10 * time.Second)
	wg.Wait()
	end = time.Now()
	duration := end.Sub(start)
	select {
	case <-contextmanager.GlobalContextManagers.GetContext(op.ID).Done():
		// 增加完成计数
		handler.TaskHandle.EndTask()
		return fmt.Errorf("task Cancel")
	default:
		// 记录模块完成日志
		handler.TaskHandle.ProgressEnd("scan", op.Target, op.ID, 1, duration)
		// 记录完成时间以及完成目标
		handler.TaskHandle.TaskEnd(op.Target, op.ID)
		// 增加完成计数
		handler.TaskHandle.EndTask()
		return nil
	}
}

func CleanTmp() {
	osType := runtime.GOOS
	if osType == "windows" {
		// Windows 系统处理
		//handleWindowsTemp()
	} else if osType == "linux" {
		// Linux 系统处理
		utils.Tools.HandleLinuxTemp()
	}
}

//func handleWindowsTemp() {
//	// 获取当前用户名
//	currentUser, err := user.Current()
//	if err != nil {
//		return
//	}
//	tempDir := fmt.Sprintf(`C:\Users\%s\AppData\Local\Temp`, currentUser.Username)
//
//	// 定义 PowerShell 命令
//	psCmd := fmt.Sprintf(`
//Get-ChildItem -Path "%s" -Directory |
//Where-Object {
//    ($_.Name -match '^\d{9}$') -or
//    ($_.Name -like 'ScopeSentry*') -or
//    ($_.Name -like '.org.chromium.Chromium*') -or
//    ($_.Name -like '*_badger') -or
//	($_.Name -like 'nuclei*') -or
//	($_.Name -like 'rod*')
//} |
//Remove-Item -Recurse -Force
//`, tempDir)
//
//	// 执行 PowerShell 命令
//	cmd := exec.Command("powershell", "-Command", psCmd)
//	_, err = cmd.CombinedOutput()
//	if err != nil {
//		fmt.Printf("执行 PowerShell 命令时出错: %v\n", err)
//		logger.SlogWarn("清空临时文件C:\\Users\\{username}\\AppData\\Local\\Temp\\[^\\d{9}$、ScopeSentry*、.org.chromium.Chromium*、*_badger]失败，请手动清空，防止占用磁盘过大，")
//		return
//	}
//}

// task-------------------------------------
// @file      : handler.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/11/1 21:22
// -------------------------------------------

package task

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pebbledb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/runner"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

func DeletePebbleTarget(PebbleStore *pebbledb.PebbleDB, targetKey []byte) {
	err := PebbleStore.Delete(targetKey)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("PebbleStore Delete error: %v", err))
	}
}

func RunPebbleTarget(runnerOption options.TaskOptions) {
	var wg sync.WaitGroup
	prefix := fmt.Sprintf("%s:", runnerOption.ID)
	targets, err := pebbledb.PebbleStore.GetKeysWithPrefix(prefix)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("pebbledb get task targets error: %v", err))
		err = pebbledb.PebbleStore.Delete([]byte(fmt.Sprintf("task:%s", runnerOption.ID)))
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("PebbleStore delete error: %v", runnerOption.ID))
		}
		return
	}
	for idTarget, _ := range targets {
		wg.Add(1)
		// 创建 runnerOption 的副本
		optionCopy := runnerOption
		target := strings.SplitN(idTarget, ":", 2)
		optionCopy.Target = target[1]
		// 使用局部变量创建闭包
		taskFunc := func(op options.TaskOptions) func() {
			return func() {
				defer wg.Done()
				select {
				case <-contextmanager.GlobalContextManagers.GetContext(op.ID).Done():
					// 任务取消直接返回
					return
				default:
					err := runner.Run(op)
					if err != nil {
						// 说明该任务取消了，直接返回不进行删除目标
						return
					} else {
						// 目标运行完毕删除目标
						DeletePebbleTarget(pebbledb.PebbleStore, []byte(op.ID+":"+op.Target))
					}
				}
			}
		}(optionCopy)

		// 提交任务
		err := pool.PoolManage.SubmitTask("task", taskFunc)
		if err != nil {
			logger.SlogError(fmt.Sprintf("task pool error: %v", err))
			// 如果提交任务失败，手动减少计数
			wg.Done()
		}
	}
	time.Sleep(3 * time.Second)
	wg.Wait()
}

func InitTaskOption(runnerOption options.TaskOptions) {
	// 初始化httpx
	parameter, _ := utils.Tools.GetParameter(runnerOption.Parameters, "AssetMapping", "3a0d994a12305cb15a5cb7104d819623")
	initHttpx(parameter)
}

func initHttpx(parameter string) {
	cdncheck := "false"
	screenshot := false
	tlsprobe := false
	FollowRedirects := true
	bypassHeader := false
	screenshotTimeout := 10
	threads := 30
	if parameter != "" {
		args, err := utils.Tools.ParseArgs(parameter, "cdncheck", "screenshot", "st", "tlsprobe", "fr", "et", "bh", "t")
		if err != nil {
		} else {
			for key, value := range args {
				if value != "" {
					switch key {
					case "cdncheck":
						cdncheck = value
					case "screenshot":
						if value == "true" {
							screenshot = true
						}
					case "tlsprobe":
						if value == "true" {
							tlsprobe = true
						}
					case "st":
						screenshotTimeout, _ = strconv.Atoi(value)
					case "fr":
						if value == "false" {
							FollowRedirects = false
						}
					case "bh":
						if value == "true" {
							bypassHeader = true
						}
					case "t":
						threads, _ = strconv.Atoi(value)
					default:
						continue
					}
				}
			}
		}
	}
	utils.InitHttpx(cdncheck, screenshot, screenshotTimeout, tlsprobe, FollowRedirects, bypassHeader, threads)
}

func OptionClose() {
	utils.HttpxClose()
}

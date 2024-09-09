// task-------------------------------------
// @file      : task.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 21:08
// -------------------------------------------

package task

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pebbledb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"os"
)

func GetTask() {
	prefix := "task:"
	keys, err := pebbledb.PebbleStore.GetKeysWithPrefix(prefix)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("pebbledb get task error: %v", err))
		os.Exit(0)
	}
	// 打印所有以 "task:" 开头的键值对
	for _, value := range keys {
		logger.SlogInfoLocal(fmt.Sprintf("get PebbleStore task: %v", string(value)))
		var runnerOption Options
		err = utils.JSONToStruct(value, &runnerOption)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("task JSONToStruct error: %v - %v", string(value), err))
			err = pebbledb.PebbleStore.Delete([]byte(fmt.Sprintf("task:%s", value)))
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("PebbleStore delete error: %v", value))
			}
			continue
		}
		prefix = fmt.Sprintf("%s:", runnerOption.ID)
		targets, err := pebbledb.PebbleStore.GetKeysWithPrefix(prefix)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("pebbledb get task targets error: %v", err))
			err = pebbledb.PebbleStore.Delete([]byte(fmt.Sprintf("task:%s", value)))
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("PebbleStore delete error: %v", value))
			}
			continue
		}
		for target, _ := range targets {
			// 创建 runnerOption 的副本
			optionCopy := runnerOption
			optionCopy.Target = target

			// 使用局部变量创建闭包
			taskFunc := func(op Options) func() {
				return func() {
					Run(op)
				}
			}(optionCopy)

			// 提交任务
			err := pool.PoolManage.SubmitTask("task", taskFunc)
			if err != nil {
			}
		}
	}
}

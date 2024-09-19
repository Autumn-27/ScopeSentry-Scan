// task-------------------------------------
// @file      : task.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 21:08
// -------------------------------------------

package task

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pebbledb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/runner"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"os"
	"strings"
	"sync"
)

func GetTask() {
	prefix := "task:"
	keys, err := pebbledb.PebbleStore.GetKeysWithPrefix(prefix)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("pebbledb get task error: %v", err))
		os.Exit(0)
	}
	// 打印所有以 "task:" 开头的键值对
	for key, value := range keys {
		var wg sync.WaitGroup
		logger.SlogInfoLocal(fmt.Sprintf("get PebbleStore task: %v", string(value)))
		var runnerOption options.TaskOptions
		err = utils.Tools.JSONToStruct(value, &runnerOption)
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
		for idTarget, _ := range targets {
			wg.Add(1)
			// 创建 runnerOption 的副本
			optionCopy := runnerOption
			target := strings.Split(idTarget, ":")
			optionCopy.Target = target[1]
			// 使用局部变量创建闭包
			taskFunc := func(op options.TaskOptions) func() {
				return func() {
					defer func(PebbleStore *pebbledb.PebbleDB, targetKey []byte) {
						fmt.Printf("gggggg")
						err := PebbleStore.Delete(targetKey)
						if err != nil {
							logger.SlogErrorLocal(fmt.Sprintf("PebbleStore Delete error: %v", err))
						}
					}(pebbledb.PebbleStore, []byte(op.ID+":"+op.Target))
					defer wg.Done()
					runner.Run(op)
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
		wg.Wait()
		err = pebbledb.PebbleStore.Delete([]byte("task:" + key))
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("PebbleStore Delete %v error: %v", key, err))
		}
		// 记得判断是否需要增加一个等待 所有目标执行完毕再任务结束
		fmt.Printf("任务结束: %v\n", runnerOption.ID)
	}
}

func GetPebbledbTaskTarget(id string) {

}

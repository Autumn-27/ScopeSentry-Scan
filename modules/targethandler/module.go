// subdomain-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:35
// -------------------------------------------

package targethandler

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handle"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"sync"
)

type Runner struct {
	Option     *options.TaskOptions
	NextModule interfaces.ModuleRunner
	Input      chan interface{}
	Name       string
}

func NewRunner(op *options.TaskOptions, nextModule interfaces.ModuleRunner) *Runner {
	return &Runner{
		Option:     op,
		NextModule: nextModule,
	}
}

func (r *Runner) SetInput(ch chan interface{}) {
	r.Input = ch
}

func (r *Runner) GetName() string {
	return "TargetParser"
}

func (r *Runner) ModuleRun() error {
	handle.TaskHandle.ProgressStart("TargetParser", r.Option.Target, r.Option.ID, len(r.Option.TargetParser))
	var plgWg sync.WaitGroup
	// 创建一个共享的 result 通道
	resultChan := make(chan interface{})

	// 结果处理 goroutine，异步读取插件的结果
	go func() {
		for result := range resultChan {
			// 处理每个插件的结果
			logger.SlogInfoLocal(fmt.Sprintf("Plugin result: %v", result))
		}
	}()

	for {
		select {
		case data, ok := <-r.Input:
			if !ok {
				// 通道已关闭，结束处理
				close(resultChan)
				handle.TaskHandle.ProgressEnd("TargetParser", r.Option.Target, r.Option.ID, len(r.Option.TargetParser))
				return nil
			}
			// 处理输入数据
			for _, pluginName := range r.Option.TargetParser {
				logger.SlogInfoLocal(fmt.Sprintf("%v plugin start execute: %v", pluginName, data))
				plg, flag := plugins.GlobalPluginManager.GetPlugin(r.GetName(), pluginName)
				if flag {
					plgWg.Add(1)
					plg.SetResult(resultChan)
					pluginFunc := func(data interface{}) func() {
						return func() {
							defer plgWg.Done()
							err := plg.Execute(data)
							if err != nil {
							}
						}
					}(data)
					err := pool.PoolManage.SubmitTask("targetHandler", pluginFunc)
					if err != nil {
						plgWg.Done()
						logger.SlogError(fmt.Sprintf("task pool error: %v", err))
					}
					plgWg.Wait()
				} else {
					logger.SlogError(fmt.Sprintf("plugin %v not found", pluginName))
				}
				logger.SlogInfoLocal(fmt.Sprintf("%v plugin end execute: %v", pluginName, data))
			}
		}
	}
}

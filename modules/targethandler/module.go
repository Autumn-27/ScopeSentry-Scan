// subdomain-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:35
// -------------------------------------------

package targethandler

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handler"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"sync"
	"time"
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
	return "TargetHandler"
}

func (r *Runner) ModuleRun() error {
	var allPluginWg sync.WaitGroup
	var resultWg sync.WaitGroup
	var nextModuleRun sync.WaitGroup
	// 创建一个共享的 result 通道
	resultChan := make(chan interface{}, 200)
	go func() {
		nextModuleRun.Add(1)
		defer nextModuleRun.Done()
		err := r.NextModule.ModuleRun()
		if err != nil {
			logger.SlogError(fmt.Sprintf("Next module run error: %v", err))
		}
	}()
	// 结果处理 goroutine，异步读取插件的结果
	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for {
			select {
			case result, ok := <-resultChan:
				if !ok {
					time.Sleep(10 * time.Second)
					r.NextModule.CloseInput()
					return
				}
				// 处理每个插件的结果
				// 对目标的输出进行去重，防止多个插件返回相同的结果
				target, ok := result.(string)
				if !ok {
					r.NextModule.GetInput() <- result
				} else {
					if r.Option.IsRestart {
						// 如果是重启的不进行去重
						r.NextModule.GetInput() <- result
					} else {
						key := "duplicates:" + r.Option.ID + ":target:" + target
						flag := results.Duplicate.DuplicateLocalCache(key)
						if flag {
							// 本地缓存中不存在，则没有重复，发到下个模块
							logger.SlogInfoLocal(fmt.Sprintf("%v module target %v result: %v", r.GetName(), r.Option.Target, result))
							r.NextModule.GetInput() <- result
						}
					}
				}
			}
		}
	}()
	var firstData bool
	firstData = false
	var start time.Time
	var end time.Time
	doneCalled := false
	for {
		select {
		case <-contextmanager.GlobalContextManagers.GetContext(r.Option.ID).Done():
			allPluginWg.Wait()
			if !doneCalled {
				close(resultChan)
				resultWg.Wait()
				r.Option.ModuleRunWg.Done()
				doneCalled = true // 标记已调用 Done
			}
			nextModuleRun.Wait()
			return nil
		case data, ok := <-r.Input:
			if !ok {
				time.Sleep(3 * time.Second)
				// 等待所有插件运行完毕
				allPluginWg.Wait()
				if firstData {
					end = time.Now()
					duration := end.Sub(start)
					handler.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.TargetHandler), duration)
				}
				if !doneCalled {
					close(resultChan)
					resultWg.Wait()
					r.Option.ModuleRunWg.Done()
					doneCalled = true // 标记已调用 Done
				}
				logger.SlogInfoLocal(fmt.Sprintf("module %v target %v close resultChan", r.GetName(), r.Option.Target))
				nextModuleRun.Wait()
				return nil
			}
			_, ok = data.(string)
			if !ok {
				r.NextModule.GetInput() <- data
				continue
			}
			if !firstData {
				start = time.Now()
				handler.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.TargetHandler))
				firstData = true
			}
			allPluginWg.Add(1)
			go func(data interface{}) {
				defer allPluginWg.Done()
				// 处理输入数据
				for _, pluginId := range r.Option.TargetHandler {
					var plgWg sync.WaitGroup
					plg, flag := plugins.GlobalPluginManager.GetPlugin(r.GetName(), pluginId)
					if flag {
						logger.SlogInfoLocal(fmt.Sprintf("%v plugin start execute: %v", plg.GetName(), data))
						plgWg.Add(1)
						args, argsFlag := utils.Tools.GetParameter(r.Option.Parameters, r.GetName(), plg.GetPluginId())
						if argsFlag {
							plg.SetParameter(args)
						} else {
							plg.SetParameter("")
						}
						plg.SetResult(resultChan)
						plg.SetTaskId(r.Option.ID)
						plg.SetTaskName(r.Option.TaskName)
						pluginFunc := func(data interface{}) func() {
							return func() {
								defer plgWg.Done()
								select {
								case <-contextmanager.GlobalContextManagers.GetContext(r.Option.ID).Done():
									return
								default:
									_, err := plg.Execute(data)
									if err != nil {
									}
								}
							}
						}(data)
						err := pool.PoolManage.SubmitTask(r.GetName(), pluginFunc)
						if err != nil {
							plgWg.Done()
							logger.SlogError(fmt.Sprintf("task pool error: %v", err))
						}
						plgWg.Wait()
						logger.SlogInfoLocal(fmt.Sprintf("%v plugin end execute: %v", plg.GetName(), data))
					} else {
						logger.SlogError(fmt.Sprintf("plugin %v not found", pluginId))
					}
				}
			}(data)

		}
	}
}

func (r *Runner) GetInput() chan interface{} {
	return r.Input
}

func (r *Runner) CloseInput() {
	close(r.Input)
}

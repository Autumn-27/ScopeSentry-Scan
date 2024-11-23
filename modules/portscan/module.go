// portscan-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 21:06
// -------------------------------------------

package portscan

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handler"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"sync"
	"time"
)

type Runner struct {
	Option     *options.TaskOptions
	NextModule interfaces.ModuleRunner
	Input      chan interface{}
}

func NewRunner(op *options.TaskOptions, nextModule interfaces.ModuleRunner) *Runner {
	return &Runner{
		Option:     op,
		NextModule: nextModule,
	}
}

func (r *Runner) ModuleRun() error {
	var allPluginWg sync.WaitGroup
	var resultWg sync.WaitGroup
	// 创建一个共享的 result 通道
	resultChan := make(chan interface{}, 100)
	go func() {
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
					// 如果 resultChan 关闭了，退出循环
					// 此模块运行完毕，关闭下个模块的输入
					r.NextModule.CloseInput()
					return
				}
				if portaliveResult, ok := result.(types.PortAlive); ok {
					port := portaliveResult.Port
					if port == "" {
						port = "null"
					}
					flag := results.Duplicate.PortIntask(r.Option.ID, portaliveResult.Host, port, r.Option.IsRestart)
					if flag {
						// 本地缓存中不存在
						r.NextModule.GetInput() <- result
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
		//
		select {
		case <-contextmanager.GlobalContextManagers.GetContext(r.Option.ID).Done():
			allPluginWg.Wait()
			if !doneCalled {
				close(resultChan)
				resultWg.Wait()
				r.Option.ModuleRunWg.Done()
				doneCalled = true // 标记已调用 Done
			}
			return nil
		case data, ok := <-r.Input:
			if !ok {
				time.Sleep(3 * time.Second)
				allPluginWg.Wait()
				// 通道已关闭，结束处理
				if firstData {
					end = time.Now()
					duration := end.Sub(start)
					handler.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.PortScan), duration)
				}
				if !doneCalled {
					close(resultChan)
					resultWg.Wait()
					r.Option.ModuleRunWg.Done()
					doneCalled = true // 标记已调用 Done
				}
				return nil
			}
			_, ok = data.(types.DomainSkip)
			if !ok {
				r.NextModule.GetInput() <- data
				continue
			}
			if !firstData {
				start = time.Now()
				handler.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.PortScan))
				firstData = true
			}
			allPluginWg.Add(1)
			go func(data interface{}) {
				defer allPluginWg.Done()
				// 无论有没有选择端口扫描 将原始数据发送到结果处
				domainSkip, _ := data.(types.DomainSkip)
				result := types.PortAlive{
					Host: domainSkip.Domain,
					IP:   "",
					Port: "",
				}
				resultChan <- result
				//发送来的数据 只能是types.DomainSkip
				if len(r.Option.PortScan) != 0 {
					// 调用插件
					for _, pluginId := range r.Option.PortScan {
						//var plgWg sync.WaitGroup
						var plgWg sync.WaitGroup
						plg, flag := plugins.GlobalPluginManager.GetPlugin(r.GetName(), pluginId)
						if flag {
							logger.SlogDebugLocal(fmt.Sprintf("%v plugin start execute: %v", plg.GetName(), data))
							plgWg.Add(1)
							args, argsFlag := utils.Tools.GetParameter(r.Option.Parameters, r.GetName(), plg.GetPluginId())
							if argsFlag {
								plg.SetParameter(args)
							} else {
								plg.SetParameter("")
							}
							newParameter := plg.GetParameter() + " -port " + r.Option.PortRange
							plg.SetParameter(newParameter)

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
							logger.SlogDebugLocal(fmt.Sprintf("%v plugin end execute: %v", plg.GetName(), data))
						} else {
							logger.SlogError(fmt.Sprintf("plugin %v not found", pluginId))
						}
					}
				}
			}(data)

		}
	}
}

func (r *Runner) SetInput(ch chan interface{}) {
	r.Input = ch
}

func (r *Runner) GetName() string {
	return "PortScan"
}

func (r *Runner) GetInput() chan interface{} {
	return r.Input
}

func (r *Runner) CloseInput() {
	close(r.Input)
}

// subdomainsecurity-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 21:06
// -------------------------------------------

package subdomainsecurity

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

// ModuleRun 子域名安全检测，如：子域名接管
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
			// 结果只有两种可能，一种是types.SubTakeResult 子域名接管结果，一种是types.DomainResolve域名解析结果
			select {
			case result, ok := <-resultChan:
				if !ok {
					// 如果 resultChan 关闭了，退出循环
					// 此模块运行完毕，关闭下个模块的输入
					r.NextModule.CloseInput()
					return
				}
				if subdomainTakeoverResult, ok := result.(types.SubTakeResult); ok {
					// 子域名接管检测结果，无需发送到下个模块
					subdomainTakeoverResult.TaskName = r.Option.TaskName
					go results.Handler.SubdomainTakeover(&subdomainTakeoverResult)
					logger.SlogInfoLocal(fmt.Sprintf("Find subdomain takeover: %v - %v", subdomainTakeoverResult.Input, subdomainTakeoverResult.Value))
				} else {
					// DomainResolve直接发送到下个模块
					r.NextModule.GetInput() <- result
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
		// 输入为DNS信息
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
				logger.SlogDebugLocal(fmt.Sprintf("%v关闭: input开始关闭", r.GetName()))
				allPluginWg.Wait()
				// 通道已关闭，结束处理
				if firstData {
					end = time.Now()
					duration := end.Sub(start)
					handler.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.SubdomainSecurity), duration)
				}
				if !doneCalled {
					close(resultChan)
					resultWg.Wait()
					r.Option.ModuleRunWg.Done()
					doneCalled = true // 标记已调用 Done
				}
				return nil
			}
			_, ok = data.(types.SubdomainResult)
			if !ok {
				r.NextModule.GetInput() <- data
				continue
			}
			if !firstData {
				start = time.Now()
				handler.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.SubdomainSecurity))
				firstData = true
			}
			allPluginWg.Add(1)
			go func(data interface{}) {
				defer allPluginWg.Done()
				// 如果开启了子域名安全检查扫描
				if len(r.Option.SubdomainSecurity) != 0 {
					skipPluginFlag := true
					// 调用插件
					for _, pluginId := range r.Option.SubdomainSecurity {
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
							// 插件没有找到跳过此插件
							logger.SlogError(fmt.Sprintf("plugin %v not found", pluginId))
							// 在多个插件都没有找到的情况下只发送一次
							if skipPluginFlag {
								resultChan <- data
								skipPluginFlag = false
							}
						}
					}
				} else {
					// 没有开启子域名安全检查扫描，将数据转换一下发送到结果
					subdomain, ok := data.(types.SubdomainResult)
					if ok {
						tmp := types.DomainResolve{
							Domain: subdomain.Host,
							IP:     subdomain.IP,
						}
						resultChan <- tmp
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
	return "SubdomainSecurity"
}

func (r *Runner) GetInput() chan interface{} {
	return r.Input
}

func (r *Runner) CloseInput() {
	close(r.Input)
}

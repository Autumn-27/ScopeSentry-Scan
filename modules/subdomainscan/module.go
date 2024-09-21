// subdomain-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:35
// -------------------------------------------

package subdomainscan

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handle"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"sync"
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
					return
				}

				if subdomainResult, ok := result.(types.SubdomainResult); ok {
					subdomainResult.TaskId = r.Option.ID
					flag := results.Duplicate.SubdomainInTask(&subdomainResult)
					if flag {
						if r.Option.IgnoreOldSubdomains {
							// 从mongodb中查询是否存在子域名进行去重
							flag = results.Duplicate.SubdomainInMongoDb(&subdomainResult)
							if flag {
								// 没有在mongodb中查询到该子域名，存入数据库中并且开始扫描
								go results.Handler.Subdomain(&subdomainResult)
								// nextInput <- result
							}
						} else {
							// 存入数据库中，并且开始扫描
							go results.Handler.Subdomain(&subdomainResult)
							// nextInput <- result
						}
					}
				}
			}
		}
	}()

	var firstData bool
	firstData = false
	for {
		select {
		case data, ok := <-r.Input:
			if !ok {
				fmt.Println("subdomainScan关闭: input开始关闭")
				allPluginWg.Wait()
				// 通道已关闭，结束处理
				if firstData {
					handle.TaskHandle.ProgressEnd("SubdomainScan", r.Option.Target, r.Option.ID, len(r.Option.TargetParser))
				}
				close(resultChan)
				fmt.Println("subdomainScan关闭: 插件运行完毕")
				resultWg.Wait()
				fmt.Println("subdomainScan关闭: 结果处理完毕")
				return nil
			}
			if !firstData {
				handle.TaskHandle.ProgressStart("SubdomainScan", r.Option.Target, r.Option.ID, len(r.Option.TargetParser))
				firstData = true
			}
			allPluginWg.Add(1)
			go func(data interface{}) {
				defer allPluginWg.Done()
				// 调用插件
				for _, pluginName := range r.Option.SubdomainScan {
					//var plgWg sync.WaitGroup
					var plgWg sync.WaitGroup
					logger.SlogInfoLocal(fmt.Sprintf("%v plugin start execute: %v", pluginName, data))
					plg, flag := plugins.GlobalPluginManager.GetPlugin(r.GetName(), pluginName)
					if flag {
						plgWg.Add(1)
						args, argsFlag := utils.Tools.GetParameter(r.Option.Parameters, r.GetName(), plg.GetName())
						if argsFlag {
							plg.SetParameter(args)
						} else {
							plg.SetParameter("")
						}
						plg.SetResult(resultChan)
						pluginFunc := func(data interface{}) func() {
							return func() {
								defer plgWg.Done()
								err := plg.Execute(data)
								if err != nil {
								}
							}
						}(data)
						err := pool.PoolManage.SubmitTask(r.GetName(), pluginFunc)
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
			}(data)

		}
	}
}

func (r *Runner) SetInput(ch chan interface{}) {
	r.Input = ch
}

func (r *Runner) GetName() string {
	return "SubdomainScan"
}

func (r *Runner) GetInput() chan interface{} {
	return r.Input
}

func (r *Runner) CloseInput() {
	close(r.Input)
}

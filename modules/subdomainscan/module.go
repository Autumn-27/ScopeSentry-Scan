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
	"net"
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
					// 此模块运行完毕，关闭下个模块的输入
					r.NextModule.CloseInput()
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
								// 将子域名发送到下个模块
								//r.NextModule.GetInput() <- subdomainResult.Host
							}
						} else {
							// 存入数据库中，并且开始扫描
							go results.Handler.Subdomain(&subdomainResult)
							// 将子域名发送到下个模块
							//r.NextModule.GetInput() <- subdomainResult.Host
						}
					}
				} else {
					// 如果发来的不是types.SubdomainResult，说明是上个模块的输出直接过来的，没有开启此模块的扫描，直接发送到下个模块
					//r.NextModule.GetInput() <- result
				}
			}
		}
	}()

	var firstData bool
	firstData = false
	for {
		// 输入有两种可能，一种域名、一种ip
		select {
		case data, ok := <-r.Input:
			if !ok {
				logger.SlogDebugLocal(fmt.Sprintf("%v关闭: input开始关闭", r.GetName()))
				allPluginWg.Wait()
				// 通道已关闭，结束处理
				if firstData {
					handle.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.SubdomainScan))
				}
				close(resultChan)
				logger.SlogDebugLocal(fmt.Sprintf("%v关闭: 插件运行完毕", r.GetName()))
				resultWg.Wait()
				logger.SlogDebugLocal(fmt.Sprintf("%v关闭: 结果处理完毕", r.GetName()))
				return nil
			}
			if !firstData {
				handle.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.SubdomainScan))
				firstData = true
			}
			allPluginWg.Add(1)
			go func(data interface{}) {
				defer allPluginWg.Done()
				target, _ := data.(string)
				// 判断是否为ip，如果是ip则直接发送到下个模块
				if net.ParseIP(target) != nil {
					// 如果是IP 地址
					resultChan <- data
				} else {
					// 如果开启了子域名扫描
					if len(r.Option.SubdomainScan) != 0 {
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
								if r.Option.SubdomainFilename != "" {
									// 如果设置有子域名字典，设置parameter参数供插件调用，ksubdomain必须有域名字典
									newParameter := plg.GetParameter() + " -subfile " + r.Option.SubdomainFilename
									plg.SetParameter(newParameter)
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
					} else {
						// 没有开启子域名扫描，直接将输入发送到下个模块
						resultChan <- data
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
	return "SubdomainScan"
}

func (r *Runner) GetInput() chan interface{} {
	return r.Input
}

func (r *Runner) CloseInput() {
	close(r.Input)
}

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
					r.Option.ModuleRunWg.Done()
					return
				}
				if subdomainResult, ok := result.(types.SubdomainResult); ok {
					subdomainResult.TaskName = r.Option.TaskName
					flag := results.Duplicate.SubdomainInTask(r.Option.ID, subdomainResult.Host)
					if flag {
						if r.Option.IgnoreOldSubdomains {
							// 从mongodb中查询是否存在子域名进行去重
							flag = results.Duplicate.SubdomainInMongoDb(&subdomainResult)
							if flag {
								// 没有在mongodb中查询到该子域名，存入数据库中并且开始扫描
								go results.Handler.Subdomain(&subdomainResult)
								// 将子域名解析结果发送到下个模块
								r.NextModule.GetInput() <- subdomainResult
							}
						} else {
							// 存入数据库中，并且开始扫描
							go results.Handler.Subdomain(&subdomainResult)
							// 将子域名解析结果发送到下个模块
							r.NextModule.GetInput() <- subdomainResult
						}
					} else {
						// 跳过当前任务中已扫描的子域名
						continue
					}
				} else {
					// 如果发来的不是types.SubdomainResult，说明是上个模块的输出直接过来的，或者是没有开启此模块的扫描，直接发送到下个模块
					target, _ := result.(string)
					// 判断该目标是否在当前任务此节点或者其他节点已经扫描过了
					flag := results.Duplicate.SubdomainInTask(r.Option.ID, target)
					if flag {
						if net.ParseIP(target) != nil {
							tmp := types.SubdomainResult{
								Host: target,
								IP:   []string{target},
							}
							r.NextModule.GetInput() <- tmp
						} else {
							resultDns := utils.DNS.QueryOne(result.(string))
							tmp := utils.DNS.DNSdataToSubdomainResult(resultDns)
							if len(tmp.IP) != 0 || len(tmp.Value) != 0 {
								tmp.TaskName = r.Option.TaskName
								go results.Handler.Subdomain(&tmp)
								r.NextModule.GetInput() <- tmp
							}
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
	for {
		// 输入有两种可能，一种域名，一种ip
		select {
		case data, ok := <-r.Input:
			if !ok {
				logger.SlogDebugLocal(fmt.Sprintf("%v关闭: input开始关闭", r.GetName()))
				allPluginWg.Wait()
				// 通道已关闭，结束处理
				if firstData {
					end = time.Now()
					duration := end.Sub(start)
					handle.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.SubdomainScan), duration)
				}
				close(resultChan)
				logger.SlogDebugLocal(fmt.Sprintf("%v关闭: 插件运行完毕", r.GetName()))
				resultWg.Wait()
				logger.SlogDebugLocal(fmt.Sprintf("%v关闭: 结果处理完毕", r.GetName()))
				return nil
			}
			if !firstData {
				start = time.Now()
				handle.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.SubdomainScan))
				firstData = true
			}
			allPluginWg.Add(1)
			go func(data interface{}) {
				defer allPluginWg.Done()
				// 将原始数据发送给下一个模块，防止漏掉原始目标的测绘
				resultChan <- data
				t, _ := data.(string)
				if net.ParseIP(t) != nil {
					return
				}
				// 如果开启了子域名扫描
				if len(r.Option.SubdomainScan) != 0 {
					// 跳过插件
					skipPluginFlag := true
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
							//if r.Option.SubdomainFilename != "" {
							//	// 如果设置有子域名字典，设置parameter参数供插件调用，ksubdomain必须有域名字典
							//	newParameter := plg.GetParameter() + " -subfile " + r.Option.SubdomainFilename
							//	plg.SetParameter(newParameter)
							//}
							plg.SetResult(resultChan)
							plg.SetTaskId(r.Option.ID)
							pluginFunc := func(data interface{}) func() {
								return func() {
									defer plgWg.Done()
									_, err := plg.Execute(data)
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
							// 插件没有找到跳过此插件
							logger.SlogError(fmt.Sprintf("plugin %v not found, Skip this plugin", pluginName))
							// 在多个插件都没有找到的情况下只发送一次
							if skipPluginFlag {
								resultChan <- data
								skipPluginFlag = false
							}
						}
						logger.SlogInfoLocal(fmt.Sprintf("%v plugin end execute: %v", pluginName, data))
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

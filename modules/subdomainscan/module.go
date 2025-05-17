// subdomain-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:35
// -------------------------------------------

package subdomainscan

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
	var nextModuleRun sync.WaitGroup
	// 创建一个共享的 result 通道
	resultChan := make(chan interface{}, 1000)
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
					// 如果 resultChan 关闭了，退出循环
					// 此模块运行完毕，关闭下个模块的输入
					r.NextModule.CloseInput()
					return
				}
				if subdomainResult, ok := result.(types.SubdomainResult); ok {
					subdomainResult.TaskName = r.Option.TaskName
					flag := results.Duplicate.SubdomainInTask(r.Option.ID, subdomainResult.Host, r.Option.IsRestart)
					if flag {
						if r.Option.Duplicates == "subdomain" && !r.Option.IsRestart {
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
					//fmt.Printf("get result begin:%v\n", result)
					// 如果发来的不是types.SubdomainResult，说明是上个模块的输出直接过来的，或者是没有开启此模块的扫描，直接发送到下个模块
					target, ok := result.(string)
					if !ok {
						r.NextModule.GetInput() <- result
					} else {
						// 判断该目标是否在当前任务此节点或者其他节点已经扫描过了
						flag := results.Duplicate.SubdomainInTask(r.Option.ID, target, r.Option.IsRestart)
						if flag {
							if net.ParseIP(target) != nil {
								tmp := types.SubdomainResult{
									Host: target,
									IP:   []string{target},
								}
								r.NextModule.GetInput() <- tmp
							} else {
								resultDns := utils.DNS.QueryOne(target)
								resultDns.Host = target
								tmp := utils.DNS.DNSdataToSubdomainResult(resultDns)
								// 无论是否有解析ip都发送到后边
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
	doneCalled := false
	for {
		// 输入有两种可能，一种域名，一种ip
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
				logger.SlogDebugLocal(fmt.Sprintf("%v关闭: input开始关闭", r.GetName()))
				allPluginWg.Wait()
				if !doneCalled {
					close(resultChan)
					resultWg.Wait()
					r.Option.ModuleRunWg.Done()
					doneCalled = true // 标记已调用 Done
				}
				// 通道已关闭，结束处理
				if firstData {
					end = time.Now()
					duration := end.Sub(start)
					handler.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.SubdomainScan), duration)
				}
				logger.SlogInfoLocal(fmt.Sprintf("module %v target %v close resultChan", r.GetName(), r.Option.Target))
				nextModuleRun.Wait()
				return nil
			}
			//_, ok = data.(string)
			//if !ok {
			//	r.NextModule.GetInput() <- data
			//	continue
			//}
			if !firstData {
				start = time.Now()
				handler.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.SubdomainScan))
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
					for _, pluginId := range r.Option.SubdomainScan {
						//var plgWg sync.WaitGroup
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
							//if r.Option.SubdomainFilename != "" {
							//	// 如果设置有子域名字典，设置parameter参数供插件调用，ksubdomain必须有域名字典
							//	newParameter := plg.GetParameter() + " -subfile " + r.Option.SubdomainFilename
							//	plg.SetParameter(newParameter)
							//}
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
							// 插件没有找到跳过此插件
							logger.SlogError(fmt.Sprintf("plugin %v not found, Skip this plugin", pluginId))
							// 在多个插件都没有找到的情况下只发送一次
							if skipPluginFlag {
								resultChan <- data
								skipPluginFlag = false
							}
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
	return "SubdomainScan"
}

func (r *Runner) GetInput() chan interface{} {
	return r.Input
}

func (r *Runner) CloseInput() {
	close(r.Input)
}

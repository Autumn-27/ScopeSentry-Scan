// assetmapping-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 21:06
// -------------------------------------------

package assetmapping

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handler"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
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
	var nextModuleRun sync.WaitGroup
	// 创建一个共享的 result 通道
	resultChan := make(chan interface{}, 500)
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
			case <-contextmanager.GlobalContextManagers.GetContext(r.Option.ID).Done():
				r.NextModule.CloseInput()
				return
			case result, ok := <-resultChan:
				if !ok {
					// 如果 resultChan 关闭了，退出循环
					// 此模块运行完毕，关闭下个模块的输入
					r.NextModule.CloseInput()
					return
				}
				// 如果没有选择资产测绘的话，这里会收到assetOther类型为http的资产，不再进行判断进行测绘
				r.NextModule.GetInput() <- result
				//if assetResult, ok := result.(types.AssetOther); ok {
				//	if assetResult.Type == "http" {
				//		// 这里可能是上个模块直接发送过来的
				//		httpxResultsHandler := func(ra types.AssetHttp) {
				//			r.NextModule.GetInput() <- ra
				//		}
				//		var url string
				//		if assetResult.Port != "" {
				//			url = assetResult.Host + ":" + assetResult.Port + assetResult.UrlPath
				//		} else {
				//			url = assetResult.Host + assetResult.UrlPath
				//		}
				//		utils.Requests.Httpx([]string{url}, httpxResultsHandler, "false", false, 10, false, true, contextmanager.GlobalContextManagers.GetContext(r.Option.ID), 10, false)
				//	} else {
				//		// 如果是other类型的资产，直接发送到下个模块
				//		r.NextModule.GetInput() <- result
				//	}
				//} else {
				//	// 如果不是types.AssetOther，就是types.AssetHttp，直接发送到下个模块
				//	r.NextModule.GetInput() <- result
				//}
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
			nextModuleRun.Wait()
			return nil
		case data, ok := <-r.Input:
			if !ok {
				time.Sleep(3 * time.Second)
				allPluginWg.Wait()
				// 通道已关闭，结束处理
				if firstData {
					end = time.Now()
					duration := end.Sub(start)
					handler.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.AssetMapping), duration)
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
			//assets, ok := data.([]interface{})
			//if !ok {
			//	r.NextModule.GetInput() <- data
			//	continue
			//}
			switch data.(type) {
			case []interface{}:
			case types.Company:
			case types.ICP:
			case types.RootDomain:
				r.NextModule.GetInput() <- data
			default:
				r.NextModule.GetInput() <- data
				continue
			}
			if !firstData {
				start = time.Now()
				handler.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.AssetMapping))
				firstData = true
			}
			allPluginWg.Add(1)
			go func(assets interface{}) {
				defer allPluginWg.Done()
				//发送来的数据 只能是types.Asset
				if len(r.Option.AssetMapping) != 0 {
					// 调用插件
					for _, pluginId := range r.Option.AssetMapping {
						//var plgWg sync.WaitGroup
						var plgWg sync.WaitGroup
						plg, flag := plugins.GlobalPluginManager.GetPlugin(r.GetName(), pluginId)
						if flag {
							logger.SlogDebugLocal(fmt.Sprintf("%v plugin start execute", plg.GetName()))
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
							// 这里和其他模块不同 传递的是数组
							pluginFunc := func(assets interface{}) func() {
								return func() {
									defer plgWg.Done()
									select {
									case <-contextmanager.GlobalContextManagers.GetContext(r.Option.ID).Done():
										return
									default:
										_, err := plg.Execute(assets)
										if err != nil {
										}
									}
								}
							}(assets)
							err := pool.PoolManage.SubmitTask(r.GetName(), pluginFunc)
							if err != nil {
								plgWg.Done()
								logger.SlogError(fmt.Sprintf("task pool error: %v", err))
							}
							plgWg.Wait()
							logger.SlogDebugLocal(fmt.Sprintf("%v plugin end execute", plg.GetName()))
						} else {
							logger.SlogError(fmt.Sprintf("plugin %v not found", pluginId))
						}
					}
				} else {
					// 如果没有开启资产测绘，将types.Asset 发送到结果处，在结果处进行转换
					switch d := assets.(type) {
					case []interface{}:
						for _, asset := range d {
							resultChan <- asset
						}
					default:
						resultChan <- d
					}

				}
			}(data)

		}
	}
	return nil
}

func (r *Runner) SetInput(ch chan interface{}) {
	r.Input = ch
}

func (r *Runner) GetName() string {
	return "AssetMapping"
}

func (r *Runner) GetInput() chan interface{} {
	return r.Input
}

func (r *Runner) CloseInput() {
	close(r.Input)
}

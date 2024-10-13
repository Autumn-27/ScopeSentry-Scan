// webcrawler-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 21:05
// -------------------------------------------

package webcrawler

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handle"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
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
					// 此模块运行完毕，关闭下个模块的输入
					r.NextModule.CloseInput()
					return
				}
				fmt.Println(result)
			}
		}
	}()

	var firstData bool
	firstData = false
	for {

		select {
		case data, ok := <-r.Input:
			if !ok {
				allPluginWg.Wait()
				// 通道已关闭，结束处理
				if firstData {
					handle.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.WebCrawler))
				}
				close(resultChan)
				resultWg.Wait()
				r.Option.ModuleRunWg.Done()
				return nil
			}
			// 该模块接收的数据为[]string、types.UrlResult、types.AssetOther 、 types.AssetHttp
			// 该模块只处理[]string 其余全部发送到下个模块
			if _, ok := data.([]string); !ok {
				r.NextModule.GetInput() <- data
				continue
			}
			if !firstData {
				handle.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.WebCrawler))
				firstData = true
			}

			allPluginWg.Add(1)
			go func(data interface{}) {
				defer allPluginWg.Done()

				if len(r.Option.WebCrawler) != 0 {
					// 调用插件
					for _, pluginName := range r.Option.WebCrawler {
						//var plgWg sync.WaitGroup
						var plgWg sync.WaitGroup
						logger.SlogDebugLocal(fmt.Sprintf("%v plugin start execute: %v", pluginName, data))
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
							logger.SlogError(fmt.Sprintf("plugin %v not found", pluginName))
						}
						logger.SlogDebugLocal(fmt.Sprintf("%v plugin end execute: %v", pluginName, data))
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
	return "WebCrawler"
}

func (r *Runner) GetInput() chan interface{} {
	return r.Input
}

func (r *Runner) CloseInput() {
	close(r.Input)
}

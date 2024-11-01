// urlscan-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 21:05
// -------------------------------------------

package urlscan

import (
	"fmt"
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
				// 这里的输入为types.UrlResult，将types.UrlResult处理一下存入数据库并发送到下个模块
				// 原始的types.AssetOther 、 types.AssetHttp 在读取input的时候已经发送到下个模块了
				// 该结果已经在插件中进行去重
				if urlResult, ok := result.(types.UrlResult); ok {
					urlResult.TaskName = r.Option.TaskName
					hash := utils.Tools.GenerateHash()
					urlResult.ResultId = hash
					go results.Handler.URL(&urlResult)
					r.NextModule.GetInput() <- urlResult
				}
			}
		}
	}()

	var firstData bool
	firstData = false
	var start time.Time
	var end time.Time
	for {
		//
		select {
		case data, ok := <-r.Input:
			if !ok {
				allPluginWg.Wait()
				// 通道已关闭，结束处理
				if firstData {
					end = time.Now()
					duration := end.Sub(start)
					handler.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.URLScan), duration)
				}
				close(resultChan)
				resultWg.Wait()
				r.Option.ModuleRunWg.Done()
				return nil
			}
			// 将原始数据发送到下个模块，这里的输入为 types.AssetOther 、 types.AssetHttp
			r.NextModule.GetInput() <- data
			// 如果是AssetOther，不运行该模块，只运行http资产
			if _, ok := data.(types.AssetOther); ok {
				continue
			}
			if !firstData {
				start = time.Now()
				handler.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.URLScan))
				firstData = true
			}

			allPluginWg.Add(1)
			go func(data interface{}) {
				defer allPluginWg.Done()

				if len(r.Option.URLScan) != 0 {
					var urlList []string
					// 调用插件
					for _, pluginName := range r.Option.URLScan {
						//var plgWg sync.WaitGroup
						var plgWg sync.WaitGroup
						logger.SlogDebugLocal(fmt.Sprintf("%v plugin start execute", pluginName))
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
							plg.SetTaskName(r.Option.TaskName)
							pluginFunc := func(data interface{}) func() {
								return func() {
									defer plgWg.Done()
									urlS, err := plg.Execute(data)
									if err == nil {
										urls, ok := urlS.([]string)
										if ok {
											if len(urls) > 0 {
												urlList = append(urlList, urls...)
											}
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
						} else {
							logger.SlogError(fmt.Sprintf("plugin %v not found", pluginName))
						}
						logger.SlogDebugLocal(fmt.Sprintf("%v plugin end execute", pluginName))
					}
					if len(urlList) > 0 {
						// 如果urlList不为空，则发送到爬虫模块，将这些url作为输入进行爬虫
						r.NextModule.GetInput() <- urlList
					} else {
						// 如果为空，则将http资产的url作为数组传递到爬虫模块进行爬虫
						if httpData, ok := data.(types.AssetHttp); ok {
							r.NextModule.GetInput() <- []string{httpData.URL}
						}
					}
				} else {
					// 如果没有开启url扫描，则将爬虫的目标发到下个模块
					if httpData, ok := data.(types.AssetHttp); ok {
						// 如果没有开启 把http转一个urlresult发往下个模块 用于检测首页的敏感信息泄露
						r.NextModule.GetInput() <- types.UrlResult{
							Input:      httpData.URL,
							Output:     httpData.URL,
							OutputType: "httpx",
							ResultId:   utils.Tools.GenerateHash(),
							Body:       httpData.ResponseBody,
						}
						r.NextModule.GetInput() <- []string{httpData.URL}
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
	return "URLScan"
}

func (r *Runner) GetInput() chan interface{} {
	return r.Input
}

func (r *Runner) CloseInput() {
	close(r.Input)
}

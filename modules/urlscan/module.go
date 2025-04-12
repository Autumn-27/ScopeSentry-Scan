// urlscan-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 21:05
// -------------------------------------------

package urlscan

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handler"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"path/filepath"
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
	resultChan := make(chan interface{}, 2000)
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
				// 这里的输入为types.UrlResult，将types.UrlResult处理一下存入数据库并发送到下个模块
				// 原始的types.AssetOther 、 types.AssetHttp 在读取input的时候已经发送到下个模块了
				// 该结果已经在插件中进行去重
				if urlResult, ok := result.(types.UrlResult); ok {
					urlResult.TaskName = r.Option.TaskName
					hash := utils.Tools.GenerateHash()
					urlResult.ResultId = hash
					if !urlResult.IsFile {
						// app文件不存入数据库url result
						go results.Handler.URL(&urlResult)
					}
					r.NextModule.GetInput() <- urlResult
				} else {
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
					handler.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.URLScan), duration)
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
			// 将原始数据发送到下个模块，这里的输入为 types.AssetOther 、 types.AssetHttp
			r.NextModule.GetInput() <- data

			if !firstData {
				start = time.Now()
				handler.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.URLScan))
				firstData = true
			}

			allPluginWg.Add(1)
			go func(data interface{}) {
				defer allPluginWg.Done()
				// 如果是AssetOther，不运行该模块，只运行http资产
				httpData, ok := data.(types.AssetHttp)
				if !ok {
					return
				}
				// 对http资产在当前任务进行去重判断

				filename := utils.Tools.CalculateMD5(httpData.URL)
				//if !r.Option.IsStart {
				//	flag := results.Duplicate.DuplicateUrlFileKey(filename, r.Option.ID)
				//	if !flag {
				//		// 重复 已经扫过了
				//		return
				//	}
				//}
				// 将原始url写入文件中
				urlFilePath := filepath.Join(global.TmpDir, filename)
				err := utils.Tools.WriteContentFileAppend(urlFilePath, httpData.URL+"\n")
				if err != nil {
				}

				if len(r.Option.URLScan) != 0 {
					// 调用插件
					for _, pluginId := range r.Option.URLScan {
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
							pluginFunc := func(data interface{}) func() {
								return func() {
									defer plgWg.Done()
									select {
									case <-contextmanager.GlobalContextManagers.GetContext(r.Option.ID).Done():
										return
									default:
										_, err := plg.Execute(data)
										if err == nil {
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
							logger.SlogDebugLocal(fmt.Sprintf("%v plugin end execute", plg.GetName()))
						} else {
							logger.SlogError(fmt.Sprintf("plugin %v not found", pluginId))
						}
					}
					// 调用url去重工具 对url数据进行去重

					// 发送urlfile
					urlFile := types.UrlFile{
						Filepath: urlFilePath,
					}
					r.NextModule.GetInput() <- urlFile

				} else {
					// 如果没有开启 把http转一个urlresult发往下个模块 用于检测首页的敏感信息泄露
					r.NextModule.GetInput() <- types.UrlResult{
						Input:      httpData.URL,
						Output:     httpData.URL,
						OutputType: "httpx",
						ResultId:   utils.Tools.GenerateHash(),
						Body:       httpData.ResponseBody,
						Status:     httpData.StatusCode,
					}
					// 发送urlfile
					urlFile := types.UrlFile{
						Filepath: urlFilePath,
					}
					r.NextModule.GetInput() <- urlFile
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

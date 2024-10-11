// assethandle-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 21:06
// -------------------------------------------

package assethandle

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
	"go.mongodb.org/mongo-driver/bson"
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
				r.NextModule.GetInput() <- result
				if assetResult, ok := result.(types.AssetOther); ok {
					assetResult.TaskId = r.Option.ID
					flag, id, bsonData := results.Duplicate.AssetInMongodb(assetResult.Host, assetResult.Port)
					if flag {
						// 数据库中存在该资产，对该资产信息进行diff
						var oldAsset types.AssetOther
						data, _ := bson.Marshal(bsonData)
						_ = bson.Unmarshal(data, &oldAsset)
						changeData := utils.Results.CompareAssetOther(oldAsset, assetResult)
						if changeData.Timestamp != "" {
							// 说明资产存在变化，将结果发送到changelog中
							changeData.AssetId = id
							go results.Handler.AssetChangeLog(&changeData)
							// 对资产进行更新,设置最新的扫描时间
							assetResult.LastScanTime = assetResult.Timestamp
							assetResult.Timestamp = oldAsset.Timestamp
							assetResult.Project = oldAsset.Project
							assetResult.RootDomain = oldAsset.RootDomain
							go results.Handler.AssetUpdate(id, assetResult)
						}
						// 资产没有变化，不进行操作
					} else {
						// 数据库中不存在该资产，直接插入。
						assetResult.LastScanTime = assetResult.Timestamp
						go results.Handler.AssetOtherInsert(&assetResult)
					}
				} else {
					assetHttpResult, okh := result.(types.AssetHttp)
					if okh {
						assetHttpResult.TaskId = r.Option.ID
						flag, id, bsonData := results.Duplicate.AssetInMongodb(assetHttpResult.Host, assetHttpResult.Port)
						if flag {
							var oldAssetHttp types.AssetHttp
							data, _ := bson.Marshal(bsonData)
							_ = bson.Unmarshal(data, &oldAssetHttp)
							changeData := utils.Results.CompareAssetHttp(oldAssetHttp, assetHttpResult)
							if changeData.Timestamp != "" {
								// 说明资产存在变化，将结果发送到changelog中
								changeData.AssetId = id
								go results.Handler.AssetChangeLog(&changeData)
								// 对资产进行更新,设置最新的扫描时间
								assetHttpResult.LastScanTime = assetHttpResult.Timestamp
								assetHttpResult.Timestamp = oldAssetHttp.Timestamp
								assetHttpResult.Project = oldAssetHttp.Project
								assetHttpResult.RootDomain = oldAssetHttp.RootDomain
								go results.Handler.AssetUpdate(id, assetHttpResult)
							}
							// 资产没有变化，不进行操作
						} else {
							// 数据库中不存在该资产，直接插入。
							go results.Handler.AssetHttpInsert(&assetHttpResult)
						}
					}
				}

			}
		}
	}()

	var firstData bool
	firstData = false
	for {
		//
		select {
		case data, ok := <-r.Input:
			if !ok {
				allPluginWg.Wait()
				// 通道已关闭，结束处理
				if firstData {
					handle.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.AssetHandle))
				}
				close(resultChan)
				resultWg.Wait()
				r.Option.ModuleRunWg.Done()
				return nil
			}
			if !firstData {
				handle.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.AssetHandle))
				firstData = true
			}
			allPluginWg.Add(1)
			go func(data interface{}) {
				defer allPluginWg.Done()
				var ty string
				var assetOther types.AssetOther
				var assetHttp types.AssetHttp
				switch a := data.(type) {
				case types.AssetOther:
					ty = "other"
					assetOther = a
				case types.AssetHttp:
					ty = "htttp"
					assetHttp = a
				}
				if len(r.Option.AssetHandle) != 0 {
					// 调用插件
					for _, pluginName := range r.Option.AssetHandle {
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
							var pluginFunc func()
							if ty == "other" {
								pluginFunc = func(data interface{}) func() {
									return func() {
										defer plgWg.Done()
										_, err := plg.Execute(data)
										if err != nil {
										}
									}
								}(&assetOther)
							} else {
								pluginFunc = func(data interface{}) func() {
									return func() {
										defer plgWg.Done()
										_, err := plg.Execute(data)
										if err != nil {
										}
									}
								}(&assetHttp)
							}

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
				// 如果没有开启此模块，或者开启此模块并且插件运行结束，将data发送到结果处理处
				if ty == "other" {
					resultChan <- assetOther
				} else {
					resultChan <- assetHttp
				}
			}(data)
		}
	}
}

func (r *Runner) SetInput(ch chan interface{}) {
	r.Input = ch
}

func (r *Runner) GetName() string {
	return "AssetHandle"
}

func (r *Runner) GetInput() chan interface{} {
	return r.Input
}

func (r *Runner) CloseInput() {
	close(r.Input)
}

// assethandle-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 21:06
// -------------------------------------------

package assethandle

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
	"go.mongodb.org/mongo-driver/bson"
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
		var assetOtherArray []types.AssetOther
		var assetHttpArray []types.AssetHttp
		for {
			select {
			case result, ok := <-resultChan:
				if !ok {
					// 如果 resultChan 关闭了，退出循环
					// 此模块运行完毕，关闭下个模块的输入
					if len(assetOtherArray) > 0 {
						r.NextModule.GetInput() <- assetOtherArray
					}

					if len(assetHttpArray) > 0 {
						r.NextModule.GetInput() <- assetHttpArray
					}

					r.NextModule.CloseInput()
					return
				}
				r.NextModule.GetInput() <- result
				switch dataTmp := result.(type) {
				case types.AssetOther:
					if dataTmp.Type == "http" {
						continue
					}
					// 过滤unknown
					if dataTmp.Service == "unknown" {
						if len(dataTmp.Raw) == 0 {
							continue
						}
					}
					dataTmp.TaskName = []string{r.Option.TaskName}
					flag, id, bsonData := results.Duplicate.AssetInMongodb(dataTmp.Host, dataTmp.Port)
					if flag {
						// 数据库中存在该资产，对该资产信息进行diff
						var oldAsset types.AssetOther
						data, _ := bson.Marshal(bsonData)
						_ = bson.Unmarshal(data, &oldAsset)
						changeData := utils.Results.CompareAssetOther(oldAsset, dataTmp)
						if changeData.Timestamp != "" {
							// 说明资产存在变化，将结果发送到changelog中
							changeData.AssetId = id
							go results.Handler.AssetChangeLog(&changeData)
							// 对资产进行更新,设置最新的扫描时间
						}
						dataTmp.LastScanTime = dataTmp.Time
						dataTmp.Time = oldAsset.Time
						dataTmp.Project = oldAsset.Project
						dataTmp.RootDomain = oldAsset.RootDomain
						dataTmp.TaskName = append(dataTmp.TaskName, oldAsset.TaskName...)
						dataTmp.TaskName = utils.Tools.RemoveStringDuplicates(dataTmp.TaskName)
						dataTmp.Tags = append(dataTmp.Tags, oldAsset.Tags...)
						dataTmp.Tags = utils.Tools.RemoveStringDuplicates(dataTmp.Tags)
						go results.Handler.AssetUpdate(id, dataTmp)
						// 资产没有变化，不进行操作
					} else {
						// 数据库中不存在该资产，直接插入。
						dataTmp.LastScanTime = dataTmp.Time
						go results.Handler.AssetOtherInsert(&dataTmp)
					}
					assetOtherArray = append(assetOtherArray, dataTmp)
					if len(assetOtherArray) > 10 {
						r.NextModule.GetInput() <- assetOtherArray
						assetOtherArray = nil
					}
				case types.AssetHttp:
					dataTmp.TaskName = []string{r.Option.TaskName}
					flag, id, bsonData := results.Duplicate.AssetInMongodb(dataTmp.Host, dataTmp.Port)
					if flag {
						var oldAssetHttp types.AssetHttp
						data, _ := bson.Marshal(bsonData)
						_ = bson.Unmarshal(data, &oldAssetHttp)
						changeData := utils.Results.CompareAssetHttp(oldAssetHttp, dataTmp)
						if changeData.Timestamp != "" {
							// 说明资产存在变化，将结果发送到changelog中
							changeData.AssetId = id
							go results.Handler.AssetChangeLog(&changeData)
						}
						// 对资产进行更新,设置最新的扫描时间
						dataTmp.LastScanTime = dataTmp.Time
						dataTmp.Time = oldAssetHttp.Time
						dataTmp.Project = oldAssetHttp.Project
						dataTmp.RootDomain = oldAssetHttp.RootDomain
						dataTmp.TaskName = append(dataTmp.TaskName, oldAssetHttp.TaskName...)
						dataTmp.TaskName = utils.Tools.RemoveStringDuplicates(dataTmp.TaskName)
						dataTmp.Tags = append(dataTmp.Tags, oldAssetHttp.Tags...)
						dataTmp.Tags = utils.Tools.RemoveStringDuplicates(dataTmp.Tags)
						go results.Handler.AssetUpdate(id, dataTmp)
						// 资产没有变化，不进行操作
					} else {
						// 数据库中不存在该资产，直接插入。
						go results.Handler.AssetHttpInsert(&dataTmp)
					}
					assetHttpArray = append(assetHttpArray, dataTmp)
					if len(assetHttpArray) > 10 {
						r.NextModule.GetInput() <- assetHttpArray
						assetHttpArray = nil
					}
				case types.RootDomain:
					dataTmp.TaskName = r.Option.TaskName
					dataTmp.Time = utils.Tools.GetTimeNow()
					go results.Handler.RootDomain(&dataTmp)
				case types.APP:
					dataTmp.TaskName = r.Option.TaskName
					dataTmp.Time = utils.Tools.GetTimeNow()
					go results.Handler.APP(&dataTmp)
				case types.MP:
					dataTmp.TaskName = r.Option.TaskName
					dataTmp.Time = utils.Tools.GetTimeNow()
					go results.Handler.MP(&dataTmp)
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
					handler.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.AssetHandle), duration)
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
			if !firstData {
				start = time.Now()
				handler.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.AssetHandle))
				firstData = true
			}
			allPluginWg.Add(1)
			go func(data interface{}) {
				defer allPluginWg.Done()
				var ty string
				var assetOther types.AssetOther
				var assetHttp types.AssetHttp
				var rootDomain types.RootDomain
				var app types.APP
				var mp types.MP
				switch a := data.(type) {
				case types.AssetOther:
					ty = "other"
					assetOther = a
				case types.AssetHttp:
					ty = "htttp"
					assetHttp = a
				case types.RootDomain:
					ty = "rootDomain"
					rootDomain = a
				case types.APP:
					ty = "app"
					app = a
				case types.MP:
					ty = "mp"
					mp = a
				default:
					r.NextModule.GetInput() <- data
					return
				}
				if len(r.Option.AssetHandle) != 0 {
					// 调用插件
					for _, pluginId := range r.Option.AssetHandle {
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
							var pluginFunc func()
							switch ty {
							case "other":
								pluginFunc = func(data interface{}) func() {
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
								}(&assetOther)
							case "htttp":
								pluginFunc = func(data interface{}) func() {
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
								}(&assetHttp)
							case "rootDomain":
								pluginFunc = func(data interface{}) func() {
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
								}(&rootDomain)
							case "app":
								pluginFunc = func(data interface{}) func() {
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
								}(&app)
							case "mp":
								pluginFunc = func(data interface{}) func() {
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
								}(&mp)
							}
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
				}
				// 如果没有开启此模块，或者开启此模块并且插件运行结束，将data发送到结果处理处

				switch ty {
				case "other":
					resultChan <- assetOther
				case "htttp":
					resultChan <- assetHttp
				case "rootDomain":
					resultChan <- rootDomain
				case "app":
					resultChan <- app
				case "mp":
					resultChan <- mp
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

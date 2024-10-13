// portfingerprint-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/26 21:09
// -------------------------------------------

package portfingerprint

import (
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handle"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"strconv"
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
				r.NextModule.GetInput() <- result
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
					handle.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.PortFingerprint), duration)
				}
				close(resultChan)
				resultWg.Wait()
				r.Option.ModuleRunWg.Done()
				return nil
			}
			if !firstData {
				start = time.Now()
				handle.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.PortFingerprint))
				firstData = true
			}
			allPluginWg.Add(1)
			go func(data interface{}) {
				defer allPluginWg.Done()
				//发送来的数据 只能是types.PortAlive
				portAlive, _ := data.(types.PortAlive)
				var asset types.AssetOther
				asset = types.AssetOther{
					Host:    portAlive.Host,
					IP:      portAlive.IP,
					Port:    portAlive.Port,
					Service: "",
				}
				// 这里如果端口为空，说明是直接发过来并没有进行端口扫描，直接发送到下个模块
				if asset.Port == "" {
					// 如果端口为空，则只测试http服务
					asset.Type = "http"
					resultChan <- asset
				} else {
					if len(r.Option.PortFingerprint) != 0 {
						// 调用插件
						for _, pluginName := range r.Option.PortFingerprint {
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
								}(&asset)
								err := pool.PoolManage.SubmitTask(r.GetName(), pluginFunc)
								if err != nil {
									plgWg.Done()
									logger.SlogError(fmt.Sprintf("task pool error: %v", err))
								}
								plgWg.Wait()
								if asset.Service != "" {
									// 如果已经识别到端口的服务，则退出循环不执行之后的插件
									break
								}
							} else {
								logger.SlogError(fmt.Sprintf("plugin %v not found", pluginName))
							}
							logger.SlogDebugLocal(fmt.Sprintf("%v plugin end execute: %v", pluginName, data))
						}
						// 如果没有检测到端口服务，则获取原始响应
						if asset.Service == "" {
							asset.Type = "other"
							asset.Service = "unknown"
							portUint64, err := strconv.ParseUint(asset.Port, 10, 16)
							if err != nil {
								fmt.Println("转换错误:", err)
								logger.SlogError(fmt.Sprintf("端口转换错误: %v", err))
							} else {
								rev, err := utils.Requests.TcpRecv(asset.IP, uint16(portUint64))
								if err == nil {
									rawResponse := string(rev)
									encodedResponse, err := json.Marshal(rawResponse)
									if err != nil {
										// 处理编码错误
										logger.SlogError(fmt.Sprintf("JSON 编码错误:", err))
									} else {
										asset.Raw = json.RawMessage(fmt.Sprintf("{\"data\":%s}", encodedResponse))
									}
								}
							}
							resultChan <- asset
						} else {
							// 识别成功
							resultChan <- asset
						}
					} else {
						// 如果没有开启端口指纹识别扫描，则只进行http测绘
						asset.Type = "http"
						resultChan <- asset
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
	return "PortFingerprint"
}

func (r *Runner) GetInput() chan interface{} {
	return r.Input
}

func (r *Runner) CloseInput() {
	close(r.Input)
}

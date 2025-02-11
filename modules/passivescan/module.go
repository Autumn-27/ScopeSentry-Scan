// passivescan-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/2/8 19:30
// -------------------------------------------

package passivescan

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handler"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"sync"
	"time"
)

type Runner struct {
	Option     *options.TaskOptions
	NextModule interfaces.ModuleRunner
	Input      chan interface{}
}

func NewRunner(op *options.TaskOptions, nextModule interfaces.ModuleRunner) *Runner { // 同样改为值类型
	return &Runner{
		Option:     op,
		NextModule: nextModule,
	}
}

func (r *Runner) ModuleRun() error {
	defer PassiveScanWgMap[r.Option.ID].Done()

	if len(r.Option.PassiveScan) != 0 {
		var start time.Time
		var end time.Time
		var wg sync.WaitGroup
		handler.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.PassiveScan))
		start = time.Now()
		// 调用插件
		for _, pluginId := range r.Option.PassiveScan {
			plg, flag := plugins.GlobalPluginManager.GetPlugin(r.GetName(), pluginId)
			if flag {
				logger.SlogDebugLocal(fmt.Sprintf("%v plugin start execute", plg.GetName()))
				go func() {
					wg.Add(1)
					defer wg.Done()
					_, err := plg.Execute("")
					if err != nil {

					}
				}()
			} else {
				logger.SlogError(fmt.Sprintf("plugin %v not found", pluginId))
			}
		}
		closePlgFunc := func() {
			for _, pluginId := range r.Option.PassiveScan {
				plg, flag := plugins.GlobalPluginManager.GetPlugin(r.GetName(), pluginId)
				if flag {
					logger.SlogInfo(fmt.Sprintf("task %v close %v module %v plugin", r.Option.ID, r.GetName(), plg.GetName()))
					go func() {
						plg.SetCustom("close task")
					}()
				}
			}
			end = time.Now()
			duration := end.Sub(start)
			handler.TaskHandle.ProgressEnd(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.PassiveScan), duration)
		}

		for {
			select {
			case <-contextmanager.GlobalContextManagers.GetContext(r.Option.ID).Done():
				closePlgFunc()
				wg.Wait()
				return nil
			case data, ok := <-r.Input:
				if !ok {
					time.Sleep(3 * time.Second)
					closePlgFunc()
					wg.Wait()
					return nil
				}

				// 获取到输入数据 通过SetCustom发送到插件中，这里为了兼容之前的，使用SetCustom传递数据，不增加新的函数
				for _, pluginId := range r.Option.PassiveScan {
					plg, flag := plugins.GlobalPluginManager.GetPlugin(r.GetName(), pluginId)
					if flag {
						logger.SlogInfo(fmt.Sprintf("task %v close %v module %v plugin", r.Option.ID, r.GetName(), plg.GetName()))
						go func() {
							plg.SetCustom(data)
						}()
					}
				}

			}
		}

	}
	return nil
}

func (r *Runner) SetInput(ch chan interface{}) {
	r.Input = ch
}

func (r *Runner) GetName() string {
	return "DirScan"
}

func (r *Runner) GetInput() chan interface{} {
	return r.Input
}

func (r *Runner) CloseInput() {
	close(r.Input)
}

var TaskPassiveScanGlobal = make(map[string]chan interface{})
var mu sync.Mutex // 声明一个互斥锁
var PassiveScanWgMap = make(map[string]*sync.WaitGroup)

func SetPassiveScanChan(op *options.TaskOptions) {
	mu.Lock()         // 获取锁，确保只有一个 goroutine 能修改 map
	defer mu.Unlock() // 函数结束时释放锁

	// 检查 id 是否已经存在
	if _, exists := TaskPassiveScanGlobal[op.ID]; !exists {
		// 如果不存在，则创建一个新的 chan 并添加到字典
		vulnerabilityInputChan := make(chan interface{}, 100)
		TaskPassiveScanGlobal[op.ID] = vulnerabilityInputChan
		passivescanModule := NewRunner(op, nil)
		passivescanModule.SetInput(vulnerabilityInputChan)
		if PassiveScanWgMap[op.ID] == nil {
			PassiveScanWgMap[op.ID] = &sync.WaitGroup{}
		}
		go func() {
			PassiveScanWgMap[op.ID].Add(1)
			err := passivescanModule.ModuleRun()
			if err != nil {

			}
		}()
	}
}

func PassiveScanChanDone(id string) {
	if _, exists := TaskPassiveScanGlobal[id]; exists {
		close(TaskPassiveScanGlobal[id])
	}
}

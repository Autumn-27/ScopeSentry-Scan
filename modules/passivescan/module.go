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
	if len(r.Option.PassiveScan) != 0 {
		handler.TaskHandle.ProgressStart(r.GetName(), r.Option.Target, r.Option.ID, len(r.Option.DirScan))
		// 调用插件
		for _, pluginId := range r.Option.PassiveScan {
			plg, flag := plugins.GlobalPluginManager.GetPlugin(r.GetName(), pluginId)
			if flag {
				logger.SlogDebugLocal(fmt.Sprintf("%v plugin start execute", plg.GetName()))
				go func() {
					_, err := plg.Execute("")
					if err != nil {

					}
				}()
			} else {
				logger.SlogError(fmt.Sprintf("plugin %v not found", pluginId))
			}
		}

		for {
			select {
			case <-contextmanager.GlobalContextManagers.GetContext(r.Option.ID).Done():
				for _, pluginId := range r.Option.PassiveScan {
					plg, flag := plugins.GlobalPluginManager.GetPlugin(r.GetName(), pluginId)
					if flag {
						logger.SlogInfo(fmt.Sprintf("task %v close %v module %v plugin", r.Option.ID, r.GetName(), plg.GetName()))
						go func() {
							plg.SetCustom("close task")
						}()
					}
				}
				return nil
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

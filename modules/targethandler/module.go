// subdomain-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:35
// -------------------------------------------

package targethandler

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handle"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
)

type Runner struct {
	Option     *options.TaskOptions
	NextModule interfaces.ModuleRunner
	Input      chan interface{}
	Name       string
}

func NewRunner(op *options.TaskOptions, nextModule interfaces.ModuleRunner) *Runner {
	return &Runner{
		Option:     op,
		NextModule: nextModule,
	}
}

func (r *Runner) SetInput(ch chan interface{}) {
	r.Input = ch
}

func (r *Runner) GetName() string {
	return "TargetParser"
}

func (r *Runner) ModuleRun() error {
	handle.TaskHandle.ProgressStart("TargetParser", r.Option.Target, r.Option.ID, len(r.Option.TargetParser))
	for {
		select {
		case data, ok := <-r.Input:
			if !ok {
				// 通道已关闭，结束处理
				handle.TaskHandle.ProgressEnd("TargetParser", r.Option.Target, r.Option.ID, len(r.Option.TargetParser))
				return nil
			}
			fmt.Printf("tareget parser: %v ", data)
			// 处理输入数据
			for _, pluginName := range r.Option.TargetParser {
				fmt.Println(pluginName)
				plugins.GlobalPluginManager.GetPlugin(r.GetName(), pluginName)
			}
		}
	}
}

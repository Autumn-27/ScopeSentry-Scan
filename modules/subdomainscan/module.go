// subdomain-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:35
// -------------------------------------------

package subdomainscan

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handle"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
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
	handle.TaskHandle.ProgressStart("SubdomainScan", r.Option.Target, r.Option.ID, len(r.Option.TargetParser))
	//var plgWg sync.WaitGroup
	var firstData bool
	firstData = false
	for {
		select {
		case data, ok := <-r.Input:
			if !ok {
				// 通道已关闭，结束处理
				handle.TaskHandle.ProgressEnd("SubdomainScan", r.Option.Target, r.Option.ID, len(r.Option.TargetParser))
				return nil
			}
			if !firstData {

			}
			// 处理输入数据
			fmt.Printf("SubdomainScan input: %v \v", data)
			//for _, pluginName := range r.Option.SubdomainScan {
			//	logger.SlogInfoLocal(fmt.Sprintf("%v plugin start execute: %v", pluginName, data))
			//	plg, flag := plugins.GlobalPluginManager.GetPlugin(r.GetName(), pluginName)
			//	if flag {
			//
			//	} else {
			//		logger.SlogError(fmt.Sprintf("plugin %v not found", pluginName))
			//	}
			//	logger.SlogInfoLocal(fmt.Sprintf("%v plugin end execute: %v", pluginName, data))
			//}
		}
	}
}

func (r *Runner) SetInput(ch chan interface{}) {
	r.Input = ch
}

func (r *Runner) GetName() string {
	return "SubdomainScan"
}

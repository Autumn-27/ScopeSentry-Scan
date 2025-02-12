// myplugins-------------------------------------
// @file      : myplugins.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/17 19:49
// -------------------------------------------

package customplugin

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
)

type Plugin struct {
	Name          string
	Module        string
	Parameter     string
	PluginId      string
	Result        chan interface{}
	Custom        interface{}
	TaskId        string
	InstallFunc   func() error
	CheckFunc     func() error
	UnInstallFunc func() error
	ExecuteFunc   func(input interface{}, op options.PluginOption) (interface{}, error)
	GetNameFunc   func() string
	SetCustomFunc func(interface{})
	TaskName      string
}

func NewPlugin(module string, plgId string, installFunc func() error, checkFunc func() error, executeFunc func(input interface{}, op options.PluginOption) (interface{}, error), unInstallFunc func() error, getNameFunc func() string, setCustomFunc func(interface{})) *Plugin {
	return &Plugin{
		Module:        module,
		PluginId:      plgId,
		InstallFunc:   installFunc,
		CheckFunc:     checkFunc,
		ExecuteFunc:   executeFunc,
		UnInstallFunc: unInstallFunc,
		GetNameFunc:   getNameFunc,
		SetCustomFunc: setCustomFunc,
	}
}

func (p *Plugin) SetTaskName(name string) {
	p.TaskName = name
}
func (p *Plugin) GetTaskName() string {
	return p.TaskName
}

func (p *Plugin) SetTaskId(id string) {
	p.TaskId = id
}

func (p *Plugin) GetTaskId() string {
	return p.TaskId
}
func (p *Plugin) SetCustom(cu interface{}) {
	p.SetCustomFunc(cu)
}

func (p *Plugin) GetCustom() interface{} {
	return p.Custom
}
func (p *Plugin) SetPluginId(id string) {
	p.PluginId = id
}

func (p *Plugin) GetPluginId() string {
	return p.PluginId
}

func (p *Plugin) SetResult(ch chan interface{}) {
	p.Result = ch
}

func (p *Plugin) SetName(name string) {
	p.Name = name
}

func (p *Plugin) GetName() string {
	return p.GetNameFunc()
}

func (p *Plugin) SetModule(module string) {
	p.Module = module
}

func (p *Plugin) GetModule() string {
	return p.Module
}

func (p *Plugin) Install() error {
	return p.InstallFunc()
}
func (p *Plugin) UnInstall() error {
	return p.UnInstallFunc()
}
func (p *Plugin) Check() error {
	return p.CheckFunc()
}

func (p *Plugin) SetParameter(args string) {
	p.Parameter = args
}

func (p *Plugin) GetParameter() string {
	return p.Parameter
}

func (p *Plugin) Log(msg string, tp ...string) {
	var logTp string
	if len(tp) > 0 {
		logTp = tp[0] // 使用传入的参数
	} else {
		logTp = "i"
	}
	logger.PluginsLog(fmt.Sprintf("[Plugins %v] %v", p.GetName(), msg), logTp, p.GetModule(), p.GetPluginId())
}

func (p *Plugin) Execute(input interface{}) (interface{}, error) {
	resultFunc := func(res interface{}) {
		p.Result <- res
	}
	op := options.PluginOption{
		Name:       p.GetName(),
		Module:     p.GetModule(),
		Parameter:  p.GetParameter(),
		PluginId:   p.GetPluginId(),
		ResultFunc: resultFunc,
		Custom:     p.Custom,
		TaskId:     p.TaskId,
		TaskName:   p.GetTaskName(),
		Log:        p.Log,
		Ctx:        contextmanager.GlobalContextManagers.GetContext(p.GetTaskId()),
	}
	if p.ExecuteFunc == nil {
		p.Log("error exec is nil", "e")
		return nil, nil
	}
	return p.ExecuteFunc(input, op)
}

func (p *Plugin) Clone() interfaces.Plugin {
	return &Plugin{
		Name:          p.Name,
		Module:        p.Module,
		PluginId:      p.PluginId,
		Custom:        p.Custom,
		TaskId:        p.TaskId,
		InstallFunc:   p.InstallFunc,
		CheckFunc:     p.CheckFunc,
		ExecuteFunc:   p.ExecuteFunc,
		UnInstallFunc: p.UnInstallFunc,
		GetNameFunc:   p.GetNameFunc,
		SetCustomFunc: p.SetCustomFunc,
	}
}

// pagemonitoring-------------------------------------
// @file      : pagemonitoring.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/27 17:11
// -------------------------------------------

package pagemonitoring

import (
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"strings"
)

type Plugin struct {
	Name      string
	Module    string
	Parameter string
	PluginId  string
	Result    chan interface{}
	Custom    interface{}
	TaskId    string
	TaskName  string
}

func NewPlugin() *Plugin {
	return &Plugin{
		Name:     "PageMonitoring",
		Module:   "URLSecurity",
		PluginId: "e52b8b16d49912ca564c22319c495403",
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
func (p *Plugin) UnInstall() error {
	return nil
}
func (p *Plugin) GetTaskId() string {
	return p.TaskId
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
func (p *Plugin) SetCustom(cu interface{}) {
	p.Custom = cu
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
	return p.Name
}

func (p *Plugin) SetModule(module string) {
	p.Module = module
}

func (p *Plugin) GetModule() string {
	return p.Module
}

func (p *Plugin) Install() error {
	return nil
}

func (p *Plugin) Check() error {
	return nil
}

func (p *Plugin) SetParameter(args string) {
	p.Parameter = args
}

func (p *Plugin) GetParameter() string {
	return p.Parameter
}

func (p *Plugin) Execute(input interface{}) (interface{}, error) {
	data, ok := input.(types.UrlResult)
	if !ok {
		return nil, errors.New("input is not types.UrlResult")
	}
	flag := utils.Tools.IsSuffixURL(data.Output, ".js")
	if flag {
		if strings.Contains(data.Body, "<!DOCTYPE html>") {
			data.Body = ""
			data.Status = 0
			data.Length = 0
		}
	}
	urlMd5 := utils.Tools.CalculateMD5(data.Output)
	bodyHash := ""
	if data.Status != 0 {
		bodyHash = utils.Tools.CalculateMD5(data.Body)
	}
	pageMonit := types.PageMonit{
		Url:        data.Output,
		StatusCode: []int{data.Status},
		Length:     []int{data.Length},
		Hash:       []string{bodyHash},
		Md5:        urlMd5,
		TaskName:   p.TaskName,
		Time:       utils.Tools.GetTimeNow(),
		State:      1,
	}
	go results.Handler.PageMonitoringInsert(&pageMonit)
	pageMonitBody := types.PageMonitBody{
		Md5:     urlMd5,
		Content: []string{data.Body},
	}
	go results.Handler.PageMonitoringInsertBody(&pageMonitBody)
	return nil, nil
}

func (p *Plugin) Clone() interfaces.Plugin {
	return &Plugin{
		Name:     p.Name,
		Module:   p.Module,
		PluginId: p.PluginId,
		Custom:   p.Custom,
		TaskId:   p.TaskId,
	}
}

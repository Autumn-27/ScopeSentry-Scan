// sentrydir-------------------------------------
// @file      : sentrydir.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/16 19:59
// -------------------------------------------

package sentrydir

import (
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/dirscan/sentrydir/dircore"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/dirscan/sentrydir/dirrunner"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"path/filepath"
	"strconv"
	"time"
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
		Name:     "SentryDir",
		Module:   "DirScan",
		PluginId: "920546788addc6d29ea63e4a314a1b85",
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
	data, ok := input.(types.AssetHttp)
	if !ok {
		return nil, errors.New("input is not types.AssetHttp")
	}
	start := time.Now()
	p.Log(fmt.Sprintf("scan terget begin: %v", data.URL))

	// 获取上下文
	ctx := contextmanager.GlobalContextManagers.GetContext(p.GetTaskId())

	resultHandle := func(response types.HttpResponse) {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			// 上下文已取消，直接返回
			return
		default:
			// 上下文未取消，继续执行后续逻辑
			var result types.DirResult
			result.Url = response.Url
			result.Length = response.ContentLength
			result.Status = response.StatusCode
			result.Msg = response.Redirect
			p.Result <- result
		}
	}
	parameter := p.GetParameter()
	dictFile := ""
	Thread := 10
	args, err := utils.Tools.ParseArgs(parameter, "d", "t")
	if err != nil {
	} else {
		for key, value := range args {
			if value != "" {
				switch key {
				case "d":
					dictFile = value
				case "t":
					Thread, _ = strconv.Atoi(value)
				}
			}
		}
	}
	if dictFile == "" {
		p.Log(fmt.Sprintf("not found dir dict, parameter :%v", parameter), "w")
		return nil, nil
	}

	dirDicConfigPath := filepath.Join(global.DictPath, dictFile)
	controller := dirrunner.Controller{Targets: []string{data.URL}, Dictionary: dirDicConfigPath}
	op := dircore.Options{
		Extensions:    []string{"php", "aspx", "jsp", "html", "js"},
		Thread:        Thread,
		MatchCallback: resultHandle,
		Ct:            ctx,
	}
	controller.Run(op)
	end := time.Now()
	duration := end.Sub(start)
	p.Log(fmt.Sprintf("scan terget end: %v time: %v", data.URL, duration))
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

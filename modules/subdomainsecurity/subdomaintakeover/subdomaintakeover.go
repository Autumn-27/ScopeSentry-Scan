// subdomaintakeover-------------------------------------
// @file      : subdomaintakeover.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/22 20:05
// -------------------------------------------

package subdomaintakeover

import (
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
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
		Name:     "SubdomainTakeover",
		Module:   "SubdomainSecurity",
		PluginId: "c0c71c101271f38b8be1767f3626d291",
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
	subdomain, ok := input.(types.SubdomainResult)
	if !ok {
		logger.SlogError(fmt.Sprintf("%v error: %v input is not a SubdomainResult\n", p.Name, input))
		return nil, errors.New("input is not a SubdomainResult")
	}
	if subdomain.Type == "CNAME" {
		// 如果是CNAME类型的子域名，开始检查子域名接管
		for _, t := range subdomain.Value {
			for _, finger := range global.SubdomainTakerFingers {
				for _, c := range finger.Cname {
					if strings.Contains(t, c) {
						bodyByte, err := utils.Requests.HttpGetByte("https://" + t)
						if err != nil {
							bodyByte, _ = utils.Requests.HttpGetByte("http://" + t)
						}
						body := string(bodyByte)
						if len(body) != 0 {
							for _, resp := range finger.Response {
								if strings.Contains(body, resp) {
									resultTmp := types.SubTakeResult{}
									resultTmp.Input = subdomain.Host
									resultTmp.Value = t
									resultTmp.Cname = c
									resultTmp.Response = resp
									p.Result <- resultTmp
								}
							}
						}
					}
				}
			}
		}
	}
	// 无论是不是CNAME解析，都需要将host发送到
	result := types.DomainResolve{
		Domain: subdomain.Host,
		IP:     subdomain.IP,
	}
	p.Result <- result
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

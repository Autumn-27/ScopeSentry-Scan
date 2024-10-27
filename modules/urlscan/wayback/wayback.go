// wayback-------------------------------------
// @file      : wayback.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/13 23:43
// -------------------------------------------

package wayback

import (
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlscan/wayback/source"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"strings"
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
		Name:     "wayback",
		Module:   "URLScan",
		PluginId: "ef244b3462744dad3040f9dcf3194eb1",
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
	data, ok := input.(types.AssetHttp)
	if !ok {
		logger.SlogError(fmt.Sprintf("%v error: %v input is not AssetHttp\n", p.Name, input))
		return nil, errors.New("input is not AssetHttp")
	}
	waybackResults := make(chan source.Result, 100)
	go func() {
		for result := range waybackResults {
			isMatch := utils.Tools.IsMatchingFilter(global.DisallowedURLFilters, []byte(result.URL))
			if isMatch {
				continue
			}
			// 去重
			flag := results.Duplicate.URL(result.URL, p.TaskId)
			if flag {
				// 没有重复
				var r types.UrlResult
				r.Input = data.URL
				r.Source = result.Source
				r.Output = result.URL
				r.OutputType = ""
				response, err := utils.Requests.HttpGet(result.URL)
				if err != nil {
					r.Status = 0
					r.Length = 0
					r.Body = ""
				} else {
					r.Status = response.StatusCode
					r.Length = len(response.Body)
					r.Body = response.Body
				}
				r.Time = utils.Tools.GetTimeNow()
				p.Result <- r
			}

		}
	}()
	start := time.Now()
	resultNumber := 0
	urlWithoutHTTP := strings.TrimPrefix(data.URL, "http://")
	urlWithoutHTTPS := strings.TrimPrefix(urlWithoutHTTP, "https://")

	// Waybackarchive
	number := source.WaybackarchiveRun(urlWithoutHTTPS, waybackResults)
	p.Log(fmt.Sprintf("Waybackarchive targert %v obtain the number of URLs: %v", urlWithoutHTTPS, number))
	resultNumber += number
	// Alienvault
	number = source.AlienvaultRun(data.Host, waybackResults)
	p.Log(fmt.Sprintf("Alienvault targert %v obtain the number of URLs: %v", urlWithoutHTTPS, number))
	resultNumber += number
	// Commoncrawl
	number = source.CommoncrawlRun(data.Host, waybackResults)
	resultNumber += number
	p.Log(fmt.Sprintf("Commoncrawl targert %v obtain the number of URLs: %v", urlWithoutHTTPS, number))
	end := time.Now()
	duration := end.Sub(start)
	p.Log(fmt.Sprintf("target %v all waybvack number %v running time:%v", urlWithoutHTTPS, resultNumber, duration))
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

// webfingerprint-------------------------------------
// @file      : webfingerprint.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/28 16:24
// -------------------------------------------

package webfingerprint

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"strings"
	"sync"
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
		Name:     "WebFingerprint",
		Module:   "AssetHandle",
		PluginId: "80718cc3fcb4827d942e6300184707e2",
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
func (p *Plugin) UnInstall() error {
	return nil
}

func (p *Plugin) Check() error {
	return nil
}

func (p *Plugin) SetParameter(args string) {
	p.Parameter = args
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

func (p *Plugin) GetParameter() string {
	return p.Parameter
}

func (p *Plugin) Execute(input interface{}) (interface{}, error) {
	httpResult, ok := input.(*types.AssetHttp)
	if !ok {
		// 说明不是http的资产，直接返回
		return nil, nil
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	maxWorkers := 10
	semaphore := make(chan struct{}, maxWorkers)

	for _, finger := range global.WebFingers {
		select {
		case <-contextmanager.GlobalContextManagers.GetContext(p.GetTaskId()).Done():
			break
		default:
			semaphore <- struct{}{} // 占用一个槽，限制并发数量
			wg.Add(1)
			go func(finger types.WebFinger) {
				defer func() {
					<-semaphore // 释放一个槽，允许新的goroutine开始
					wg.Done()
				}()
				tmpExp := []bool{}
				for _, exp := range finger.Express {
					key := ""
					value := ""
					if exp != "||" && exp != "&&" {
						r := strings.SplitN(exp, "=", 2)
						if len(r) != 2 {
							continue
						}
						key = r[0]
						value = strings.Trim(r[1], `"`)
					} else {
						key = exp
					}
					switch key {
					case "title", "title!":
						if strings.Contains(httpResult.Title, value) {
							if key == "title" {
								tmpExp = append(tmpExp, true)
							} else { // key == "title!"
								tmpExp = append(tmpExp, false)
							}
						} else {
							if key == "title!" {
								tmpExp = append(tmpExp, true)
							} else { // key == "title!"
								tmpExp = append(tmpExp, false)
							}
						}
					case "body", "body!":
						if strings.Contains(httpResult.ResponseBody, value) {
							if key == "body" {
								tmpExp = append(tmpExp, true)
							} else { // key == "title!"
								tmpExp = append(tmpExp, false)
							}
						} else {
							if key == "body!" {
								tmpExp = append(tmpExp, true)
							} else { // key == "title!"
								tmpExp = append(tmpExp, false)
							}
						}
					case "header", "header!":
						if strings.Contains(httpResult.RawHeaders, value) {
							if key == "header" {
								tmpExp = append(tmpExp, true)
							} else { // key == "title!"
								tmpExp = append(tmpExp, false)
							}
						} else {
							if key == "header!" {
								tmpExp = append(tmpExp, true)
							} else { // key == "title!"
								tmpExp = append(tmpExp, false)
							}
						}
					case "banner", "banner!":
						if strings.Contains(httpResult.RawHeaders, value) {
							if key == "banner" {
								tmpExp = append(tmpExp, true)
							} else { // key == "title!"
								tmpExp = append(tmpExp, false)
							}
						} else {
							if key == "banner!" {
								tmpExp = append(tmpExp, true)
							} else { // key == "title!"
								tmpExp = append(tmpExp, false)
							}
						}
					case "server", "server!":
						if strings.Contains(strings.ToLower(httpResult.WebServer), strings.ToLower(value)) {
							if key == "server" {
								tmpExp = append(tmpExp, true)
							} else { // key == "title!"
								tmpExp = append(tmpExp, false)
							}
						} else {
							if key == "server!" {
								tmpExp = append(tmpExp, true)
							} else { // key == "title!"
								tmpExp = append(tmpExp, false)
							}
						}
					case "||":
						secondLast, last, slice := popLastTwoBool(tmpExp)
						r := last || secondLast
						slice = append(slice, r)
						tmpExp = slice
					case "&&":
						secondLast, last, slice := popLastTwoBool(tmpExp)
						r := last && secondLast
						slice = append(slice, r)
						tmpExp = slice
					default:
						tmpExp = append(tmpExp, false)
					}
				}

				if len(tmpExp) != 1 {
					return
				}

				flag := tmpExp[0]
				if flag {
					mu.Lock()
					alreadyExists := false
					for _, tech := range httpResult.Technologies {
						if strings.ToLower(tech) == strings.ToLower(finger.Name) {
							alreadyExists = true
							break
						}
					}
					if !alreadyExists {
						httpResult.Technologies = append(httpResult.Technologies, finger.Name)
					}
					mu.Unlock()
				}
			}(finger)
		}

	}
	wg.Wait()
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

func popLastTwoBool(slice []bool) (bool, bool, []bool) {
	if len(slice) < 2 {
		return false, false, slice // 如果切片长度小于2，直接返回原切片
	}

	// 获取最后两个元素
	lastIndex := len(slice) - 1
	last := slice[lastIndex]
	secondLast := slice[lastIndex-1]

	// 使用切片操作去除最后两个元素
	slice = slice[:lastIndex-1]

	return secondLast, last, slice
}

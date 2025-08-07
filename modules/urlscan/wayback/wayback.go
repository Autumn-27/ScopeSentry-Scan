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
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlscan/wayback/source"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"io"
	"net/url"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
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
func (p *Plugin) UnInstall() error {
	return nil
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
	flag := utils.Tools.CommandExists("uro")
	if !flag {
		cmd := exec.Command("pip", "install", "uro")

		// 执行命令并获取输出
		output, err := cmd.CombinedOutput()

		// 如果有错误，打印错误信息
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("install uro error: %v\n", err))
		}
		fmt.Println(string(output))
	}
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
		//logger.SlogError(fmt.Sprintf("%v error: %v input is not AssetHttp\n", p.Name, input))
		return nil, errors.New("input is not AssetHttp")
	}
	p.Log(fmt.Sprintf("target %v running", data.URL))
	waybackResults := make(chan source.Result, 100)
	var wg sync.WaitGroup
	filename := utils.Tools.HashXX64String(data.URL)
	urlFilePath := filepath.Join(global.TmpDir, filename)
	proxy := ""
	parameter := p.GetParameter()
	if parameter != "" {
		args, err := utils.Tools.ParseArgs(parameter, "proxy")
		if err != nil {
		} else {
			for key, value := range args {
				if value != "" {
					switch key {
					case "proxy":
						proxy = value
					default:
						continue
					}
				}

			}
		}
	}
	client, err := utils.GetClient(proxy, 3*time.Second)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("wayback utils.GetClient error: %v\n", err))
		return nil, err
	}
	params := make(map[string]map[string]struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		var mu sync.Mutex
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
				parsedURL, err := url.Parse(result.URL)
				paramMap := url.Values{}
				urlPath := ""
				if err != nil {
					urlPath = result.URL
				} else {
					urlPath = parsedURL.Path
					if !strings.Contains(parsedURL.RawQuery, "=") {
					} else {
						paramMap = parsedURL.Query()
					}
				}
				r.Ext = path.Ext(urlPath)
				r.Input = data.URL
				r.Source = result.Source
				r.Output = result.URL
				r.OutputType = ""
				//var response types.HttpResponse
				//if proxy != "" {
				//	response, err = utils.ProxyRequests.HttpGetProxy(result.URL, proxy)
				//} else {
				//	response, err = utils.Requests.HttpGet(result.URL)
				//}
				response, err := client.Get(result.URL)
				if err != nil {
					r.Status = 0
					r.Length = 0
					r.Body = ""
				} else {
					r.Status = response.StatusCode
					bodyBytes, err := io.ReadAll(response.Body)
					if err != nil {
						r.Length = 0
						r.Body = ""
					} else {
						r.Body = string(bodyBytes)
						r.Length = len(r.Body)
					}
				}
				r.Time = utils.Tools.GetTimeNow()
				rootDomain, err := utils.Tools.GetRootDomain(r.Output)
				if err != nil {
					logger.SlogInfoLocal(fmt.Sprintf("%v GetRootDomain error: %v", r.Output, err))
					rootDomain = ""
				}
				r.RootDomain = rootDomain
				r.OutputType = "wayback"
				p.Result <- r
				err = utils.Tools.WriteContentFileAppend(urlFilePath, result.URL+"\n")
				if err != nil {
				}
				mu.Lock()
				if _, ok := params[rootDomain]; !ok {
					params[rootDomain] = make(map[string]struct{})
				}
				for param := range paramMap {
					params[rootDomain][param] = struct{}{}
				}
				mu.Unlock()
			}
		}
	}()
	start := time.Now()
	resultNumber := 0
	urlWithoutHTTP := strings.TrimPrefix(data.URL, "http://")
	urlWithoutHTTPS := strings.TrimPrefix(urlWithoutHTTP, "https://")
	ctx := contextmanager.GlobalContextManagers.GetContext(p.GetTaskId())
	// Waybackarchive
	number := source.WaybackarchiveRun(urlWithoutHTTPS, waybackResults, ctx)
	p.Log(fmt.Sprintf("Waybackarchive targert %v obtain the number of URLs: %v", urlWithoutHTTPS, number))
	resultNumber += number
	// Alienvault
	number = source.AlienvaultRun(data.Host, waybackResults, ctx)
	p.Log(fmt.Sprintf("Alienvault targert %v obtain the number of URLs: %v", urlWithoutHTTPS, number))
	resultNumber += number
	// Commoncrawl
	number = source.CommoncrawlRun(data.Host, waybackResults, ctx)
	resultNumber += number
	p.Log(fmt.Sprintf("Commoncrawl targert %v obtain the number of URLs: %v", urlWithoutHTTPS, number))
	end := time.Now()
	duration := end.Sub(start)
	p.Log(fmt.Sprintf("target %v all waybvack number %v running time:%v", urlWithoutHTTPS, resultNumber, duration))
	close(waybackResults)
	wg.Wait()
	for domain, paramSet := range params {
		var paramSlice []interface{}
		for param := range paramSet {
			paramSlice = append(paramSlice, param)
		}
		go results.Handler.AddParam(domain, paramSlice)
	}
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

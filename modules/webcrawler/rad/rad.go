// rad-------------------------------------
// @file      : rad.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/13 15:55
// -------------------------------------------

package rad

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Plugin struct {
	Name        string
	Module      string
	Parameter   string
	PluginId    string
	Result      chan interface{}
	Custom      interface{}
	RadFileName string
	OsType      string
	RadDir      string
	TaskId      string
	TaskName    string
}

type Request struct {
	Method  string `json:"Method"`
	URL     string `json:"URL"`
	B64Body string `json:"b64_body,omitempty"`
}

func NewPlugin() *Plugin {
	osType := runtime.GOOS
	var path string
	var dir string
	switch osType {
	case "windows":
		path = "rad.exe"
		dir = "win"
	case "linux":
		path = "rad"
		dir = "linux"
	default:
		dir = "darwin"
		path = "rad"
	}
	return &Plugin{
		Name:        "rad",
		Module:      "WebCrawler",
		PluginId:    "4b292861d3228af0e4da8e7ef979497c",
		RadFileName: path,
		RadDir:      dir,
		OsType:      osType,
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
func (p Plugin) Log(msg string, tp ...string) {
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
	radPath := filepath.Join(global.ExtDir, "rad")
	if err := os.MkdirAll(radPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create rad folder:", err))
		return err
	}
	targetPath := filepath.Join(radPath, "target")
	if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create targetPath folder:", err))
		return err
	}
	resultPath := filepath.Join(radPath, "result")
	if err := os.MkdirAll(resultPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create resultPath folder:", err))
		return err
	}
	RadExecPath := filepath.Join(radPath, p.RadFileName)
	if _, err := os.Stat(RadExecPath); os.IsNotExist(err) {
		_, err := utils.Tools.HttpGetDownloadFile(fmt.Sprintf("%v/%v/%v", "https://raw.githubusercontent.com/Autumn-27/ScopeSentry-Scan/main/tools", p.RadDir, p.RadFileName), RadExecPath)
		if err != nil {
			_, err = utils.Tools.HttpGetDownloadFile(fmt.Sprintf("%v/%v/%v", "https://gitee.com/constL/ScopeSentry-Scan/raw/main/tools", p.RadDir, p.RadFileName), RadExecPath)
			if err != nil {
				return err
			}
		}
		if p.OsType == "linux" {
			err = os.Chmod(RadExecPath, 0755)
			if err != nil {
				logger.SlogError(fmt.Sprintf("Chmod rad Tool Fail: %s", err))
				return err
			}
		}
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
	data, ok := input.(types.UrlFile)
	if !ok {
		logger.SlogError(fmt.Sprintf("%v error: %v input is not []string\n", p.Name, input))
		return nil, errors.New("input is not []string")
	}
	if data.Filepath == "" {
		p.Log(fmt.Sprintf("urlfile is null", "w"))
		return nil, nil
	}
	start := time.Now()
	var resultNumber int
	var targetFileName string
	timeRandom := utils.Tools.GetTimeNow()
	strRandom := utils.Tools.GenerateRandomString(8)
	targetFileName = utils.Tools.CalculateMD5(timeRandom + strRandom)
	resultPath := filepath.Join(filepath.Join(global.ExtDir, "rad"), "result", targetFileName)
	radConfigPath := filepath.Join(filepath.Join(global.ExtDir, "rad"), "rad_config.yml")
	defer utils.Tools.DeleteFile(resultPath)
	executionTimeout := 60
	parameter := p.GetParameter()
	proxy := ""
	if parameter != "" {
		args, err := utils.Tools.ParseArgs(parameter, "et", "proxy")
		if err != nil {
		} else {
			for key, value := range args {
				if value != "" {
					switch key {
					case "et":
						executionTimeout, _ = strconv.Atoi(value)
					case "proxy":
						proxy = value
					default:
						continue
					}
				}
			}
		}
	}
	args := []string{"--url-file", data.Filepath, "--json", resultPath, "--config", radConfigPath}
	if proxy != "" {
		args = append(args, "-http-proxy")
		args = append(args, proxy)
	}
	ctx := contextmanager.GlobalContextManagers.GetContext(p.GetTaskId())
	err := utils.Tools.ExecuteCommandWithTimeout(filepath.Join(filepath.Join(global.ExtDir, "rad"), p.RadFileName), args, time.Duration(executionTimeout)*time.Minute, ctx)
	if err != nil {
		logger.SlogError(fmt.Sprintf("%v ExecuteCommandWithTimeout error: %v", p.GetName(), err))
	}
	resultChan := make(chan string, 100)

	go func() {
		err = utils.Tools.ReadFileLineReader(resultPath, resultChan, ctx)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("%v", err))
		}
	}()
	params := make(map[string]map[string]struct{})
	var mu sync.Mutex
	for result := range resultChan {
		result = strings.TrimSpace(result)
		if result == "[" || result == "]" {
			continue
		}
		result = strings.TrimRight(result, ",")
		var req Request
		err := json.Unmarshal([]byte(result), &req)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("解析 JSON 错误: %s", err))
			continue
		}
		parsedURL, err := url.Parse(req.URL)
		paramMap := url.Values{}
		body := ""
		if err == nil {
			if !strings.Contains(parsedURL.RawQuery, "=") {
			} else {
				paramMap = parsedURL.Query()
			}
		}
		if req.B64Body != "" {
			decodedBytes, err := base64.StdEncoding.DecodeString(req.B64Body)
			if err != nil {
				fmt.Println(err)
			}
			body = string(decodedBytes)
		}
		key := ""
		if req.Method == "GET" {
			key = results.Duplicate.URLParams(req.URL)
		} else {
			postKey := results.Duplicate.URLParams(req.URL)
			if body != "" {
				if strings.HasPrefix(body, "{") {
					var jsonData map[string]interface{}
					decoder := json.NewDecoder(strings.NewReader(body))
					if err := decoder.Decode(&jsonData); err != nil {
					} else {
						// 遍历JSON中的key
						for bdkey := range jsonData {
							postKey += bdkey + ", "
							paramMap.Add(bdkey, "")
						}
					}

				} else {
					bodyKeyV := strings.Split(body, "&")
					for _, part := range bodyKeyV {
						bodyKey := strings.Split(part, "=")
						if len(bodyKey) > 1 {
							postKey += bodyKey[0]
							paramMap.Add(bodyKey[0], bodyKey[1])
						}
					}
				}

			}
			key = results.Duplicate.URLParams(postKey)
		}
		taskId := p.GetTaskId()
		dFlag := results.Duplicate.Crawler(key, taskId)
		if !dFlag {
			continue
		}
		resultNumber += 1
		crawlerResult := types.CrawlerResult{
			Url:    req.URL,
			Method: req.Method,
			Body:   body,
		}
		rootDomain, err := utils.Tools.GetRootDomain(crawlerResult.Url)
		if err == nil {
			crawlerResult.RootDomain = rootDomain
			mu.Lock()
			if _, ok := params[rootDomain]; !ok {
				params[rootDomain] = make(map[string]struct{})
			}
			for param := range paramMap {
				params[rootDomain][param] = struct{}{}
			}
			mu.Unlock()
		}
		p.Result <- crawlerResult
	}
	end := time.Now()
	duration := end.Sub(start)
	osType := runtime.GOOS
	if osType == "windows" {
		// Windows 系统处理
		//handleWindowsTemp()
	} else if osType == "linux" {
		// Linux 系统处理
		utils.Tools.HandleLinuxTemp()
	}
	for domain, paramSet := range params {
		var paramSlice []interface{}
		for param := range paramSet {
			paramSlice = append(paramSlice, param)
		}
		go results.Handler.AddParam(domain, paramSlice)
	}
	p.Log(fmt.Sprintf("target file %v get result %v time %v", data.Filepath, resultNumber, duration))
	return nil, nil
}

func (p *Plugin) Clone() interfaces.Plugin {
	return &Plugin{
		Name:        p.Name,
		Module:      p.Module,
		PluginId:    p.PluginId,
		Custom:      p.Custom,
		RadFileName: p.RadFileName,
		RadDir:      p.RadDir,
		OsType:      p.OsType,
		TaskId:      p.TaskId,
	}
}

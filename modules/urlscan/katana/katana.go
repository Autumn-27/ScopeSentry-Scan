// katana-------------------------------------
// @file      : katana.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/11 21:33
// -------------------------------------------

package katana

import (
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
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Plugin struct {
	Name           string
	Module         string
	Parameter      string
	PluginId       string
	Result         chan interface{}
	Custom         interface{}
	KatanaFileName string
	OsType         string
	KatanaDir      string
	TaskId         string
	TaskName       string
}

func NewPlugin() *Plugin {
	osType := runtime.GOOS
	var path string
	var dir string
	switch osType {
	case "windows":
		path = "katana.exe"
		dir = "win"
	case "linux":
		path = "katana"
		dir = "linux"
	default:
		dir = "darwin"
		path = "katana"
	}
	return &Plugin{
		Name:           "katana",
		Module:         "URLScan",
		PluginId:       "9669d0dcc52a5ca6dbbe580ffc99c364",
		KatanaFileName: path,
		KatanaDir:      dir,
		OsType:         osType,
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
	katanaPath := filepath.Join(global.ExtDir, "katana")
	if err := os.MkdirAll(katanaPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create katana folder:", err))
		return err
	}
	targetPath := filepath.Join(katanaPath, "target")
	if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create targetPath folder:", err))
		return err
	}
	resultPath := filepath.Join(katanaPath, "result")
	if err := os.RemoveAll(resultPath); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to clear resultPath folder: %v", err))
		return err
	}
	if err := os.MkdirAll(resultPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create resultPath folder:", err))
		return err
	}
	KatanaExecPath := filepath.Join(katanaPath, p.KatanaFileName)
	if _, err := os.Stat(KatanaExecPath); os.IsNotExist(err) {
		_, err := utils.Tools.HttpGetDownloadFile(fmt.Sprintf("%v/%v/%v", "https://raw.githubusercontent.com/Autumn-27/ScopeSentry-Scan/main/tools", p.KatanaDir, p.KatanaFileName), KatanaExecPath)
		if err != nil {
			_, err = utils.Tools.HttpGetDownloadFile(fmt.Sprintf("%v/%v/%v", "https://gitee.com/constL/ScopeSentry-Scan/raw/main/tools", p.KatanaDir, p.KatanaFileName), KatanaExecPath)
			if err != nil {
				return err
			}
		}
		if p.OsType == "linux" {
			err = os.Chmod(KatanaExecPath, 0755)
			if err != nil {
				logger.SlogError(fmt.Sprintf("Chmod katana Tool Fail: %s", err))
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
	data, ok := input.(types.AssetHttp)
	if !ok {
		//logger.SlogError(fmt.Sprintf("%v error: %v input is not AssetHttp\n", p.Name, input))
		return nil, errors.New("input is not AssetHttp")
	}
	p.Log(fmt.Sprintf("target %v running", data.URL))
	parameter := p.GetParameter()
	threads := "5"
	timeout := "5"
	maxDepth := "5"
	executionTimeout := 20
	proxy := ""
	if parameter != "" {
		args, err := utils.Tools.ParseArgs(parameter, "t", "timeout", "depth", "et", "proxy")
		if err != nil {
		} else {
			for key, value := range args {
				if value != "" {
					switch key {
					case "t":
						threads = value
					case "timeout":
						timeout = value
					case "depth":
						maxDepth = value
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
	start := time.Now()
	cmd := filepath.Join(filepath.Join(global.ExtDir, "katana"), p.KatanaFileName)
	resultFile := filepath.Join(filepath.Join(filepath.Join(global.ExtDir, "katana"), "result"), utils.Tools.GenerateRandomString(16))
	defer utils.Tools.DeleteFile(resultFile)
	args := []string{
		"-u", data.URL,
		"-depth", maxDepth,
		"-mrs", "20971520", "-hl", "-xhr",
		"-fs", "rdn", "-js-crawl", "-jsonl",
		"-ef", "png,apng,bmp,gif,ico,cur,jpg,jpeg,jfif,pjp,pjpeg,svg,tif,tiff,webp,xbm,3gp,aac,flac,mpg,mpeg,mp3,mp4,m4a,m4v,m4p,oga,ogg,ogv,mov,wav,webm,eot,woff,woff2,ttf,otf",
		"-kf", "all", "-timeout", timeout,
		"-c", threads,
		"-p", "10",
		"-o", resultFile,
	}
	if proxy != "" {
		args = append(args, "-proxy")
		args = append(args, proxy)
	}
	logger.SlogDebugLocal(fmt.Sprintf("katana target:%v result:%v", data.URL, resultFile))
	ctx := contextmanager.GlobalContextManagers.GetContext(p.GetTaskId())
	err := utils.Tools.ExecuteCommandWithTimeout(cmd, args, time.Duration(executionTimeout)*time.Minute, ctx)
	if err != nil {
		logger.SlogError(fmt.Sprintf("%v ExecuteCommandWithTimeout error: %v", p.GetName(), err))
	}
	resultChan := make(chan string, 100)
	go func() {
		err = utils.Tools.ReadFileLineReader(resultFile, resultChan, ctx)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("ReadFileLineReader %v", err))
		}
	}()
	filename := utils.Tools.CalculateMD5(data.URL)
	urlFilePath := filepath.Join(global.TmpDir, filename)
	urlNumber := 0
	params := make(map[string]map[string]struct{})
	for result := range resultChan {
		var katanaResult types.KatanaResult
		err := json.Unmarshal([]byte(result), &katanaResult)
		if err != nil {
			logger.SlogWarnLocal(fmt.Sprintf("[%v]JSON parse error:%v", result, err))
			continue
		}
		if katanaResult.Request == nil {
			logger.SlogWarnLocal(fmt.Sprintf("katanaResult.Request is nil"))
			continue
		}
		p.ParseResult(&katanaResult, p.GetTaskId(), &urlNumber, data.URL, urlFilePath, params)
		//parsedURL, err := url.Parse(katanaResult.Request.URL)
		//paramMap := url.Values{}
		//urlPath := ""
		//if err != nil {
		//	urlPath = katanaResult.Request.URL
		//} else {
		//	urlPath = parsedURL.Path
		//	if !strings.Contains(parsedURL.RawQuery, "=") {
		//	} else {
		//		paramMap = parsedURL.Query()
		//	}
		//}
		//if katanaResult.Request.Method == "POST" {
		//	key := ""
		//	postKey := results.Duplicate.URLParams(katanaResult.Request.URL)
		//	if katanaResult.Request.Body != "" {
		//		bodyKeyV := strings.Split(katanaResult.Request.Body, "&")
		//		for _, part := range bodyKeyV {
		//			bodyKey := strings.Split(part, "=")
		//			if len(bodyKey) > 1 {
		//				postKey += bodyKey[0]
		//				paramMap.Add(bodyKey[0], bodyKey[1])
		//			}
		//		}
		//	}
		//	key = results.Duplicate.URLParams(postKey)
		//	taskId := p.GetTaskId()
		//	dFlag := results.Duplicate.Crawler(key, taskId)
		//	if !dFlag {
		//		continue
		//	}
		//	crawlerResult := types.CrawlerResult{
		//		Url:    katanaResult.Request.URL,
		//		Method: katanaResult.Request.Method,
		//		Body:   katanaResult.Request.Body,
		//		Tags:   []string{"katana"},
		//	}
		//	p.Result <- crawlerResult
		//}
		//rootDomain, err := utils.Tools.GetRootDomain(katanaResult.Request.URL)
		//if err != nil {
		//	logger.SlogInfoLocal(fmt.Sprintf("%v GetRootDomain error: %v", katanaResult.Request.URL, err))
		//	rootDomain = parsedURL.Hostname()
		//}
		//// 去重
		//flag := results.Duplicate.URL(katanaResult.Request.URL, p.TaskId)
		//if flag {
		//	urlNumber += 1
		//	var r types.UrlResult
		//	r.Ext = path.Ext(urlPath)
		//	r.Ext = path.Ext(parsedURL.Path)
		//	r.Input = data.URL
		//	r.Source = katanaResult.Request.Source
		//	r.Output = katanaResult.Request.URL
		//	r.OutputType = katanaResult.Request.Attribute
		//	r.Status = katanaResult.Response.StatusCode
		//	r.Length = len(katanaResult.Response.Body)
		//	r.Body = katanaResult.Response.Body
		//	r.Time = utils.Tools.GetTimeNow()
		//	r.RootDomain = rootDomain
		//	err = utils.Tools.WriteContentFileAppend(urlFilePath, katanaResult.Request.URL+"\n")
		//	if err != nil {
		//	}
		//	p.Result <- r
		//}
		//mu.Lock()
		//if _, ok := params[rootDomain]; !ok {
		//	params[rootDomain] = make(map[string]struct{})
		//}
		//for param := range paramMap {
		//	params[rootDomain][param] = struct{}{}
		//}
		//mu.Unlock()
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
	p.Log(fmt.Sprintf("target %v all url number %v running time:%v", data.URL, urlNumber, duration))
	return nil, nil
}

func (p *Plugin) ParseResult(katanaResult *types.KatanaResult, taskId string, urlNumber *int, input string, urlFilePath string, params map[string]map[string]struct{}) {
	var mu sync.Mutex
	if katanaResult.Response == nil {
		katanaResult.Response = &types.Response{}
	}
	parsedURL, err := url.Parse(katanaResult.Request.URL)
	paramMap := url.Values{}
	urlPath := ""
	if err != nil {
		urlPath = katanaResult.Request.URL
	} else {
		urlPath = parsedURL.Path
		if !strings.Contains(parsedURL.RawQuery, "=") {
		} else {
			paramMap = parsedURL.Query()
		}
	}
	if katanaResult.Request.Method == "POST" {
		key := ""
		postKey := results.Duplicate.URLParams(katanaResult.Request.URL)
		if katanaResult.Request.Body != "" {
			bodyKeyV := strings.Split(katanaResult.Request.Body, "&")
			for _, part := range bodyKeyV {
				bodyKey := strings.Split(part, "=")
				if len(bodyKey) > 1 {
					postKey += bodyKey[0]
					paramMap.Add(bodyKey[0], bodyKey[1])
				}
			}
		}
		key = results.Duplicate.URLParams(postKey)
		dFlag := results.Duplicate.Crawler(key, taskId)
		if dFlag {
			*urlNumber++
			crawlerResult := types.CrawlerResult{
				Url:      katanaResult.Request.URL,
				Method:   katanaResult.Request.Method,
				Body:     katanaResult.Request.Body,
				Tags:     []string{"katana"},
				ResultId: utils.Tools.GenerateHash(),
				Source:   katanaResult.Request.Source,
			}
			p.Result <- crawlerResult
		}
	}
	rootDomain, err := utils.Tools.GetRootDomain(katanaResult.Request.URL)
	if err != nil {
		logger.SlogInfoLocal(fmt.Sprintf("%v GetRootDomain error: %v", katanaResult.Request.URL, err))
		rootDomain = ""
	}
	// 去重
	flag := results.Duplicate.URL(katanaResult.Request.URL, p.TaskId)
	if flag {
		*urlNumber++
		var r types.UrlResult
		r.Ext = path.Ext(urlPath)
		r.Ext = path.Ext(parsedURL.Path)
		r.Input = input
		r.Source = katanaResult.Request.Source
		r.Output = katanaResult.Request.URL
		r.OutputType = katanaResult.Request.Attribute
		r.Status = katanaResult.Response.StatusCode
		r.Length = len(katanaResult.Response.Body)
		r.Body = katanaResult.Response.Body
		r.Time = utils.Tools.GetTimeNow()
		r.RootDomain = rootDomain
		err = utils.Tools.WriteContentFileAppend(urlFilePath, katanaResult.Request.URL+"\n")
		if err != nil {
		}
		p.Result <- r
	}
	mu.Lock()
	if _, ok := params[rootDomain]; !ok {
		params[rootDomain] = make(map[string]struct{})
	}
	for param := range paramMap {
		params[rootDomain][param] = struct{}{}
	}
	mu.Unlock()
	for _, req := range katanaResult.Response.XhrRequests {
		tmpResult := types.KatanaResult{
			Request:  &req,
			Response: &types.Response{},
		}
		custHeader := []string{}
		ct := ""
		for _, head := range req.Headers {
			custHeader = append(custHeader, fmt.Sprintf("%v:%v", head, req.Headers[head]))
			if strings.ToLower(head) == "content-type" {
				if strings.Contains(req.Headers[head], "json") {
					ct = "json"
				}
			}
		}

		if req.Method == "POST" {
			err, h := utils.Requests.HttpPostWithCustomHeader(req.URL, []byte(req.Body), ct, custHeader)
			if err == nil {
				tmpResult.Response.Body = string(h.Body)
				tmpResult.Response.StatusCode = h.StatusCode
				tmpResult.Response.ContentLength = int64(len(tmpResult.Response.Body))
			}
		} else {
			h, err := utils.Requests.HttpGetWithCustomHeader(req.URL, custHeader)
			if err == nil {
				tmpResult.Response.Body = h.Body
				tmpResult.Response.StatusCode = h.StatusCode
				tmpResult.Response.ContentLength = int64(len(tmpResult.Response.Body))
			}
		}
		tmpResult.Request.Attribute = "xhr_requests"
		tmpResult.Request.Source = katanaResult.Request.URL
		p.ParseResult(&tmpResult, taskId, urlNumber, input, urlFilePath, params)
	}
}

//func (p *Plugin) Execute(input interface{}) (interface{}, error) {
//	data, ok := input.(types.AssetHttp)
//	if !ok {
//		logger.SlogError(fmt.Sprintf("%v error: %v input is not a string\n", p.Name, input))
//		return nil, errors.New("input is not a string")
//	}
//
//	parameter := p.GetParameter()
//	threads := 10
//	timeout := 3
//	maxDepth := 5
//	if parameter != "" {
//		args, err := utils.Tools.ParseArgs(parameter, "t", "timeout", "depth")
//		if err != nil {
//		} else {
//			for key, value := range args {
//				switch key {
//				case "t":
//					threads, _ = strconv.Atoi(value)
//				case "timeout":
//					timeout, _ = strconv.Atoi(value)
//				case "depth":
//					maxDepth, _ = strconv.Atoi(value)
//				default:
//					continue
//				}
//			}
//		}
//	}
//	var urllist []string
//	var mu sync.Mutex
//	options := &katanaTypes.Options{
//		MaxDepth:          maxDepth,    // Maximum depth to crawl
//		FieldScope:        "rdn",       // Crawling Scope Field
//		BodyReadSize:      math.MaxInt, // Maximum response size to read
//		ScrapeJSResponses: true,
//		ExtensionFilter:   []string{"png", "apng", "bmp", "gif", "ico", "cur", "jpg", "jpeg", "jfif", "pjp", "pjpeg", "svg", "tif", "tiff", "webp", "xbm", "3gp", "aac", "flac", "mpg", "mpeg", "mp3", "mp4", "m4a", "m4v", "m4p", "oga", "ogg", "ogv", "mov", "wav", "webm", "eot", "woff", "woff2", "ttf", "otf", "css"},
//		KnownFiles:        "robotstxt,sitemapxml",
//		Timeout:           timeout,       // Timeout is the time to wait for request in seconds
//		Concurrency:       threads,       // Concurrency is the number of concurrent crawling goroutines
//		Parallelism:       10,            // Parallelism is the number of urls processing goroutines
//		Delay:             0,             // Delay is the delay between each crawl requests in seconds
//		RateLimit:         150,           // Maximum requests to send per second
//		Strategy:          "depth-first", // Visit strategy (depth-first, breadth-first)
//		OnResult: func(result output.Result) { // Callback function to execute for result
//			var r types.UrlResult
//			r.Input = data.URL
//			r.Source = result.Request.Source
//			r.Output = result.Request.URL
//			r.OutputType = result.Request.Attribute
//			r.Status = result.Response.StatusCode
//			r.Length = len(result.Response.Body)
//			r.Body = result.Response.Body
//			mu.Lock()
//			urllist = append(urllist, result.Request.URL)
//			mu.Unlock()
//			p.Result <- r
//		},
//	}
//	crawlerOptions, err := katanaTypes.NewCrawlerOptions(options)
//	if err != nil {
//		logger.SlogErrorLocal(fmt.Sprintf("katana error %v", err.Error()))
//	}
//	defer crawlerOptions.Close()
//	crawler, err := standard.New(crawlerOptions)
//	if err != nil {
//		logger.SlogErrorLocal(fmt.Sprintf("katana standard.New error %v", err.Error()))
//	}
//	defer crawler.Close()
//	err = crawler.Crawl(data.URL)
//	if err != nil {
//		logger.SlogErrorLocal(fmt.Sprintf("katana crawler.Crawl error %v: %v", input, err.Error()))
//	}
//	return urllist, nil
//}

func (p *Plugin) Clone() interfaces.Plugin {
	return &Plugin{
		Name:           p.Name,
		Module:         p.Module,
		PluginId:       p.PluginId,
		Custom:         p.Custom,
		KatanaFileName: p.KatanaFileName,
		KatanaDir:      p.KatanaDir,
		OsType:         p.OsType,
		TaskId:         p.TaskId,
	}
}

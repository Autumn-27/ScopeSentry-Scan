// trufflehog-------------------------------------
// @file      : trufflehog.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/3/29 16:06
// -------------------------------------------

package trufflehog

import (
	"context"
	"fmt"
	ssconfig "github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"github.com/dlclark/regexp2"
	"github.com/trufflesecurity/trufflehog/v3/pkg/config"
	"github.com/trufflesecurity/trufflehog/v3/pkg/detectors"
	"github.com/trufflesecurity/trufflehog/v3/pkg/engine/defaults"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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
		Name:     "trufflehog",
		Module:   "URLSecurity",
		PluginId: "1aa212b9578dc3fb1409ee8de8ed005e",
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

var AllScanners map[string]detectors.Detector

func (p *Plugin) Install() error {
	id := ssconfig.GetDictId("trufflehog", "config")
	var Detectors []detectors.Detector
	Detectors = defaults.DefaultDetectors()
	filePath := filepath.Join(global.DictPath, id)
	if filePath != "" {
		input, err := os.ReadFile(filePath)
		if err != nil {
			logger.SlogWarn(fmt.Sprintf("[%v] trufflehog get custom file error %v", p.Name, err))
		} else {
			customDetect, err := config.NewYAML(input)
			if err != nil {
				logger.SlogWarn(fmt.Sprintf("[%v] trufflehog custom file new yaml error %v", p.Name, err))
			} else {
				logger.SlogInfoLocal(fmt.Sprintf("[%v] init custom scanner number %v", p.Name, len(customDetect.Detectors)))
				Detectors = append(Detectors, customDetect.Detectors...)
			}
		}
	}
	getAllScanners(Detectors)
	logger.SlogInfoLocal(fmt.Sprintf("[%v] init scanner number %v", p.Name, len(AllScanners)))
	return nil
}

func (p *Plugin) Check() error {
	return nil
}
func (p *Plugin) UnInstall() error {
	return nil
}
func (p *Plugin) SetParameter(args string) {
	p.Parameter = args
}

func (p *Plugin) GetParameter() string {
	return p.Parameter
}

func getAllScanners(Detectors []detectors.Detector) {
	AllScanners = make(map[string]detectors.Detector)
	flag := 0
	for _, s := range Detectors {
		secretType := reflect.Indirect(reflect.ValueOf(s)).Type().PkgPath()
		path := strings.Split(secretType, "/")[len(strings.Split(secretType, "/"))-1]
		if strings.Contains(path, "custom_detectors") {
			path = fmt.Sprintf("custom_detectors_%v", flag)
			flag += 1
		}
		AllScanners[path] = s
	}
}

type matchCollector struct {
	mu       sync.Mutex
	matchMap map[string][]string // rule.Name -> []match
}

func newMatchCollector() *matchCollector {
	return &matchCollector{
		matchMap: make(map[string][]string),
	}
}

// Add 方法使用锁保护
func (mc *matchCollector) Add(ruleName, match string) {
	mc.mu.Lock()         // 上锁
	defer mc.mu.Unlock() // 解锁

	existList := mc.matchMap[ruleName]
	for _, m := range existList {
		if m == match {
			return
		}
	}
	mc.matchMap[ruleName] = append(mc.matchMap[ruleName], match)
}

func (mc *matchCollector) GetAll() map[string][]string {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	return mc.matchMap
}

var scannerLocks = sync.Map{}

func getScannerLock(name string) *sync.Mutex {
	// 尝试从 map 中获取已有锁
	if lock, ok := scannerLocks.Load(name); ok {
		return lock.(*sync.Mutex)
	}

	// 否则创建一个新的锁，并原子保存（只存一次）
	newLock := &sync.Mutex{}
	actual, _ := scannerLocks.LoadOrStore(name, newLock)
	return actual.(*sync.Mutex)
}

func (p *Plugin) Execute(input interface{}) (interface{}, error) {
	data, ok := input.(types.UrlResult)
	if !ok {
		tmp, ok := input.(types.CrawlerResult)
		if !ok {
			return nil, nil
		}
		data = types.UrlResult{
			ResultId: tmp.ResultId,
			Body:     tmp.ResBody,
		}
		if tmp.Method == "GET" {
			data.Output = tmp.Url
		} else {
			data.Output = fmt.Sprintf("POST|%v|%v", tmp.Url, tmp.Body)
		}
	}
	if data.Body == "" {
		return nil, nil
	}
	// 检查body是否在当前任务已经检测过
	respMd5 := utils.Tools.CalculateMD5(data.Body)
	duplicateFlag := results.Duplicate.SensitiveBody(respMd5, p.TaskId, "truffle")
	ctx := contextmanager.GlobalContextManagers.GetContext(p.GetTaskId())
	exclude := []string{}
	verify := false
	//start := time.Now()
	thread := 5
	if duplicateFlag {
		pdfCheck := false
		parameter := p.GetParameter()
		if parameter != "" {
			args, err := utils.Tools.ParseArgs(parameter, "pdf", "exclude", "verify", "thread")
			if err != nil {
			} else {
				for key, value := range args {
					if value != "" {
						switch key {
						case "pdf":
							if value == "true" {
								pdfCheck = true
							}
						case "exclude":
							exclude = strings.Split(value, ",")
						case "verify":
							if value == "true" {
								verify = true
							}
						case "thread":
							thread, _ = strconv.Atoi(value)
						default:
							continue
						}
					}

				}
			}
		}
		if pdfCheck {
			if strings.ToLower(data.Ext) == ".pdf" {
				tmpFilePath := filepath.Join(global.TmpDir, utils.Tools.GenerateRandomString(6)+".pdf")
				err := utils.Tools.WriteContentFile(tmpFilePath, data.Body)
				if err == nil {
					content := utils.Tools.GetPdfContent(tmpFilePath)
					if content != "" {
						data.Body = content
					}
				}
			}
		}
		if len(exclude) != 0 {
			for _, ex := range exclude {
				delete(AllScanners, ex)
			}
			logger.SlogInfoLocal(fmt.Sprintf("[%v] scanner number %v", p.Name, len(AllScanners)))
		}
		chunkSize := 5120
		overlapSize := 100
		chunkChan := GenerateChunks(data.Body, chunkSize, overlapSize, thread)
		collector := newMatchCollector()
		if strings.Contains(data.Output, "app/dist/app-main.bundle.js") {
			fmt.Println("ddd")

		}
		for chunk := range chunkChan {
			select {
			case <-ctx.Done():
				return nil, nil
			default:
			}
			str := strings.ToLower(chunk)
			for name, scanner := range AllScanners {
				foundKeyword := false
				for _, kw := range scanner.Keywords() {
					if strings.Contains(str, kw) {
						foundKeyword = true
					}
				}
				if !foundKeyword {
					continue
				}
				result, err := func() ([]detectors.Result, error) {
					scannerLock := getScannerLock(name)
					scannerLock.Lock()
					defer scannerLock.Unlock()

					// 只保护 wasm 调用这段
					return scanner.FromData(context.Background(), verify, []byte(chunk))
				}()
				if err != nil {
					logger.SlogWarnLocal(fmt.Sprintf("[trufflehog] %v scanner.FromData error %v", scanner.Description(), err))
					continue
				}
				resName := name
				for _, res := range result {
					if res.DetectorName != "" {
						if strings.Contains(resName, "custom_detectors") {
							resName = res.DetectorName
						}
					}
					logger.SlogInfoLocal(fmt.Sprintf("[%v] %v %v %v", p.Name, data.Output, name, string(res.Raw)))
					if verify {
						if res.Verified {
							collector.Add(resName, string(res.Raw))
						}
					} else {
						collector.Add(resName, string(res.Raw))
					}
				}
			}
		}
		if len(collector.matchMap) > 0 {
			for ruleName, matchList := range collector.GetAll() {
				tmpResult := types.SensitiveResult{
					Url:      data.Output,
					UrlId:    data.ResultId,
					SID:      ruleName,
					Match:    matchList,
					Time:     utils.Tools.GetTimeNow(),
					Color:    "red",
					Md5:      respMd5,
					TaskName: p.TaskName,
					Status:   1,
					Tags:     []string{p.Name},
				}
				go results.Handler.Sensitive(&tmpResult)
			}
			results.Handler.SensitiveBody(data.Body, respMd5)
		}
	}
	//end := time.Now()
	//duration := end.Sub(start)
	//logger.SlogDebugLocal(fmt.Sprintf("[Plugins %v] target %v run time: %v", p.Name, data.Output, duration))
	return nil, nil
}

func GenerateChunks(text string, chunkSize, overlapSize int, thread int) <-chan string {
	ch := make(chan string, thread)

	go func() {
		defer close(ch)

		textLen := len(text)
		for start := 0; start < textLen; start += chunkSize {
			end := start + chunkSize
			if end > textLen {
				end = textLen
			}

			chunkEnd := end
			if end+overlapSize < textLen {
				chunkEnd = end + overlapSize
			}
			chunk := text[start:chunkEnd]
			ch <- chunk
		}
	}()

	return ch
}

func processInChunks(scanner detectors.Detector, text string, chunkSize int, overlapSize int, ctx context.Context, verify bool) ([]detectors.Result, error) {
	var result []detectors.Result
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	// 确保在函数结束时调用 cancel 以释放资源
	defer cancel()
	for start := 0; start < len(text); start += chunkSize {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}

		chunkEnd := end
		if end+overlapSize < len(text) {
			chunkEnd = end + overlapSize
		}
		foundKeyword := false
		str := strings.ToLower(text[start:chunkEnd])
		for _, kw := range scanner.Keywords() {
			if strings.Contains(str, kw) {
				foundKeyword = true
			}
		}
		if !foundKeyword {
			continue
		}
		res, err := scanner.FromData(timeoutCtx, verify, []byte(text[start:chunkEnd]))
		if err != nil {
			logger.SlogWarnLocal(fmt.Sprintf("[trufflehog] %v scanner.FromData error %v", scanner.Description(), err))
			continue
		}
		if len(res) != 0 {
			result = append(result, res...)
		}
	}
	return result, nil
}

func findMatchesInChunk(regex *regexp2.Regexp, text string) ([]string, error) {
	var matches []string
	m, _ := regex.FindStringMatch(text)
	for m != nil {
		matches = append(matches, m.String())
		m, _ = regex.FindNextMatch(m)
	}
	return matches, nil
}

func uniqueStrings(input []string) []string {
	// 创建一个映射来记录出现过的字符串
	seen := make(map[string]bool)
	var result []string

	// 遍历输入的切片
	for _, str := range input {
		// 如果该字符串还未出现，则添加到结果切片和映射中
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
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

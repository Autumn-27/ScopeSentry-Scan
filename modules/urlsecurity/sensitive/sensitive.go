// sensitive-------------------------------------
// @file      : sensitive.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/15 22:26
// -------------------------------------------

package sensitive

import (
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"github.com/dlclark/regexp2"
	"path/filepath"
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
		Name:     "sensitive",
		Module:   "URLSecurity",
		PluginId: "2949994c04a4e124b9c98383489510f0",
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
func (p *Plugin) UnInstall() error {
	return nil
}
func (p *Plugin) SetParameter(args string) {
	p.Parameter = args
}

func (p *Plugin) GetParameter() string {
	return p.Parameter
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
	if len(global.SensitiveRules) == 0 {
		return nil, errors.New("SensitiveRules is null")
	}
	//var start time.Time
	//var end time.Time
	//start = time.Now()
	// 检查body是否在当前任务已经检测过
	respMd5 := utils.Tools.CalculateMD5(data.Body)
	duplicateFlag := results.Duplicate.SensitiveBody(respMd5, p.TaskId, "sens")
	ctx := contextmanager.GlobalContextManagers.GetContext(p.GetTaskId())
	if duplicateFlag {
		pdfCheck := false
		parameter := p.GetParameter()
		if parameter != "" {
			args, err := utils.Tools.ParseArgs(parameter, "pdf")
			if err != nil {
			} else {
				for key, value := range args {
					if value != "" {
						switch key {
						case "pdf":
							if value == "true" {
								pdfCheck = true
							}
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
		chunkSize := 5120
		overlapSize := 100
		chunkChan := GenerateChunks(data.Body, chunkSize, overlapSize)
		collector := newMatchCollector()
		matchColorMap := make(map[string]string)
		for chunk := range chunkChan {
			select {
			case <-ctx.Done():
				return nil, nil
			default:
			}
			for _, rule := range global.SensitiveRules {
				if !rule.State || rule.RuleCompile == nil {
					continue
				}
				matches, err := findMatchesInChunk(rule.RuleCompile, chunk)
				if err != nil {
					p.Log(fmt.Sprintf("Error matching rule %s: %v", rule.ID, err), "e")
					continue
				}
				for _, match := range matches {
					collector.Add(rule.Name, match)
				}
				if len(matches) != 0 {
					matchColorMap[rule.Name] = rule.Color
				}
			}
		}
		if len(collector.matchMap) > 0 {
			for ruleName, matchList := range collector.GetAll() {
				color, exists := matchColorMap[ruleName]
				if !exists {
					color = ""
				}
				tmpResult := types.SensitiveResult{
					Url:      data.Output,
					UrlId:    data.ResultId,
					SID:      ruleName,
					Match:    matchList,
					Time:     utils.Tools.GetTimeNow(),
					Color:    color,
					Md5:      respMd5,
					TaskName: p.TaskName,
					Status:   1,
				}
				go results.Handler.Sensitive(&tmpResult)
			}
			results.Handler.SensitiveBody(data.Body, respMd5)
		}
	}
	//end = time.Now()
	//duration := end.Sub(start)
	//p.Log(fmt.Sprintf("target %v run time: %v", data.Output, duration))
	return nil, nil
}

func GenerateChunks(text string, chunkSize, overlapSize int) <-chan string {
	ch := make(chan string)

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

func processInChunks(regex *regexp2.Regexp, text string, chunkSize int, overlapSize int) ([]string, error) {
	var result []string
	for start := 0; start < len(text); start += chunkSize {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}

		chunkEnd := end
		if end+overlapSize < len(text) {
			chunkEnd = end + overlapSize
		}

		matches, err := findMatchesInChunk(regex, text[start:chunkEnd])
		if err != nil {
			return []string{}, err
		}

		if len(matches) > 0 {
			result = append(result, matches...)
		}
	}
	if len(result) != 0 {
		result = uniqueStrings(result)
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

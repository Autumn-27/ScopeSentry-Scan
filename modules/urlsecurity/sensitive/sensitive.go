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
	"github.com/Autumn-27/ScopeSentry-Scan/internal/configupdater"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"github.com/dlclark/regexp2"
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
	if data.Status != 200 || data.Body == "" {
		return nil, nil
	}
	if len(global.SensitiveRules) == 0 {
		configupdater.UpdateSensitive()
	}
	var start time.Time
	var end time.Time
	start = time.Now()
	// 检查body是否在当前任务已经检测过
	respMd5 := utils.Tools.CalculateMD5(data.Body)
	duplicateFlag := results.Duplicate.SensitiveBody(respMd5, p.TaskId)
	if duplicateFlag {
		chunkSize := 5120
		overlapSize := 100
		findFlag := false
		for _, rule := range global.SensitiveRules {
			//start := time.Now()
			if rule.State {
				r, err := regexp2.Compile(rule.Regular, 0)
				if err != nil {
					p.Log(fmt.Sprintf("Error compiling sensitive regex pattern: %s - %s - %v", err, rule.ID, rule.Regular), "e")
					continue
				}
				result, err := processInChunks(r, data.Body, chunkSize, overlapSize)
				if err != nil {
					p.Log(fmt.Sprintf("\"Error processing chunks: %s\", err"), "e")
				}
				if len(result) != 0 {
					var tmpResult types.SensitiveResult
					tmpResult = types.SensitiveResult{
						Url:      data.Output,
						SID:      rule.Name,
						Match:    result,
						Time:     utils.Tools.GetTimeNow(),
						Color:    rule.Color,
						Md5:      respMd5,
						TaskName: p.TaskName,
					}
					go results.Handler.Sensitive(&tmpResult)
					findFlag = true
				}
			}
		}
		if findFlag {
			results.Handler.SensitiveBody(&data.Body, respMd5)
		}
	}
	end = time.Now()
	duration := end.Sub(start)
	p.Log(fmt.Sprintf("target %v run time: %v", data.Output, duration))
	return nil, nil
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

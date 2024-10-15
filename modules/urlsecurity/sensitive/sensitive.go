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
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/dlclark/regexp2"
)

type Plugin struct {
	Name      string
	Module    string
	Parameter string
	PluginId  string
	Result    chan interface{}
	Custom    interface{}
	TaskId    string
}

func NewPlugin() *Plugin {
	return &Plugin{
		Name:     "sensitive",
		Module:   "URLSecurity",
		PluginId: "2949994c04a4e124b9c98383489510f0",
	}
}

func (p *Plugin) SetTaskId(id string) {
	p.TaskId = id
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
		logger.SlogError(fmt.Sprintf("%v error: %v input is not types.UrlResult\n", p.Name, input))
		return nil, errors.New("input is not types.UrlResult")
	}
	if data.Status != 200 || data.Body == "" {
		return nil, nil
	}
	if len(global.SensitiveRules) == 0 {
		configupdater.UpdateSensitive()
	}
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
			resultChan, errorChan := processInChunks(r, data.Body, chunkSize, overlapSize)
			for matches := range resultChan {
				if len(matches) != 0 {
					var tmpResult types.SensitiveResult
					if findFlag {
						tmpResult = types.SensitiveResult{Url: url, SID: rule.Name, Match: matches, Body: "", Time: system.GetTimeNow(), Color: rule.Color, Md5: fmt.Sprintf("md5==%v", resMd5)}
					} else {
						tmpResult = types.SensitiveResult{Url: url, SID: rule.Name, Match: matches, Body: resp, Time: system.GetTimeNow(), Color: rule.Color, Md5: resMd5}
					}
					findFlag = true
				}
			}
			if err := <-errorChan; err != nil {
				p.Log(fmt.Sprintf("\"Error processing chunks: %s\", err"), "e")
			}
		}
	}
	return nil, nil
}

func processInChunks(regex *regexp2.Regexp, text string, chunkSize int, overlapSize int) (chan []string, chan error) {
	resultChan := make(chan []string, 10)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)
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
				errorChan <- err
				return
			}

			if len(matches) > 0 {
				resultChan <- matches
			}
		}
	}()

	return resultChan, errorChan
}

func findMatchesInChunk(regex *regexp2.Regexp, text string) ([]string, error) {
	var matches []string
	m, _ := regex.FindStringMatch(text)
	for m != nil {
		matches = append(matches, m.String())
		m, _ = regex.FindNextMatch(m)
	}
	mc := uniqueStrings(matches)
	return mc, nil
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
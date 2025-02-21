// subfinder-------------------------------------
// @file      : subfinder.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:35
// -------------------------------------------

package subfinder

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"github.com/projectdiscovery/subfinder/v2/pkg/resolve"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
	"log"
	"path/filepath"
	"strconv"
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
		Name:     "subfinder",
		Module:   "SubdomainScan",
		PluginId: "d60ba73c70aac430a0a54e796e7e19b8",
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
	logger.PluginsLog(fmt.Sprintf("[Plugins %v]%v", p.GetName(), msg), logTp, p.GetModule(), p.GetPluginId())
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
	target, ok := input.(string)
	if !ok {
		//logger.SlogError(fmt.Sprintf("%v error: %v input is not a string\n", p.Name, input))
		return nil, errors.New("input is not a string")
	}
	parameter := p.GetParameter()
	threads := 10
	timeout := 30
	maxEnumerationTime := 10
	if parameter != "" {
		args, err := utils.Tools.ParseArgs(parameter, "t", "timeout", "max-time")
		if err != nil {
		} else {
			for key, value := range args {
				if value != "" {
					switch key {
					case "t":
						threads, _ = strconv.Atoi(value)
					case "timeout":
						timeout, _ = strconv.Atoi(value)
					case "max-time":
						maxEnumerationTime, _ = strconv.Atoi(value)
					default:
						continue
					}
				}
			}
		}
	}
	var resultWg sync.WaitGroup
	subdomainResult := make(chan string, 100)
	readResult := func() {
		defer resultWg.Done()
		for h := range subdomainResult {
			resultDns := utils.DNS.QueryOne(h)
			resultDns.Host = h
			tmp := utils.DNS.DNSdataToSubdomainResult(resultDns)
			p.Result <- tmp
		}

	}
	go func() {
		for i := 0; i < 100; i++ {
			resultWg.Add(1)
			go readResult()
		}
	}()
	subResult := []string{}

	rawCount := 1
	// 将原始域名增加到子域名列表
	subfinderOpts := &runner.Options{
		Threads:            threads,            // Thread controls the number of threads to use for active enumerations
		Timeout:            timeout,            // Timeout is the seconds to wait for sources to respond
		MaxEnumerationTime: maxEnumerationTime, // MaxEnumerationTime is the maximum amount of time in mins to wait for enumeration
		ProviderConfig:     filepath.Join(global.ConfigDir, "subfinderConfig.yaml"),
		// and other system related options
		ResultCallback: func(s *resolve.HostEntry) {
			rawCount += 1
			logger.SlogInfoLocal(fmt.Sprintf("subfinder target %v found subdomain: %v", target, s.Host))
			subdomainResult <- s.Host
			subResult = append(subResult, s.Host)
			//go func() {
			//	p.Result <- tmp
			//}()
		},
		Domain: []string{target},
		Output: &bytes.Buffer{},
	}

	// disable timestamps in logs / configure logger
	log.SetFlags(0)
	subfinder, err := runner.NewRunner(subfinderOpts)
	if err != nil {
		log.Fatalf("failed to create subfinder runner: %v", err)
	}
	err = subfinder.RunEnumerationWithCtx(contextmanager.GlobalContextManagers.GetContext(p.GetTaskId()))
	if err != nil {
		logger.SlogError(fmt.Sprintf("%v error: %v", p.GetName(), err))
		return nil, err
	}
	close(subdomainResult)
	resultWg.Wait()
	//subdomainVerificationResult := make(chan string, 100)
	//go utils.DNS.KsubdomainVerify(rawSubdomain, subdomainVerificationResult, 1*time.Hour, contextmanager.GlobalContextManagers.GetContext(p.GetTaskId()))
	//
	//// 读取结果
	//for result := range subdomainVerificationResult {
	//	subdomainResult := utils.DNS.KsubdomainResultToStruct(result)
	//	if subdomainResult.Host != "" {
	//		verificationCount += 1
	//		p.Result <- subdomainResult
	//	} else {
	//		logger.SlogDebugLocal(result)
	//	}
	//}
	p.Log(fmt.Sprintf("%v plugin result: %v original quantity: %v", p.GetName(), target, rawCount))
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

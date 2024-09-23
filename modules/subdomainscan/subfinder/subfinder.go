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
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"github.com/projectdiscovery/subfinder/v2/pkg/resolve"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
	"log"
	"path/filepath"
	"strconv"
	"time"
)

type Plugin struct {
	Name      string
	Module    string
	Parameter string
	Result    chan interface{}
}

func NewPlugin() *Plugin {
	return &Plugin{
		Name:   "subfinder",
		Module: "SubdomainScan",
	}
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

func (p *Plugin) Execute(input interface{}) error {
	target, ok := input.(string)
	if !ok {
		logger.SlogError(fmt.Sprintf("%v error: %v input is not a string\n", p.Name, input))
		return errors.New("input is not a string")
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

	rawCount := 1
	verificationCount := 0
	rawSubdomain := []string{}
	// 将原始域名增加到子域名列表
	rawSubdomain = append(rawSubdomain, target)
	subfinderOpts := &runner.Options{
		Threads:            threads,            // Thread controls the number of threads to use for active enumerations
		Timeout:            timeout,            // Timeout is the seconds to wait for sources to respond
		MaxEnumerationTime: maxEnumerationTime, // MaxEnumerationTime is the maximum amount of time in mins to wait for enumeration
		ProviderConfig:     filepath.Join(global.ConfigDir, "subfinderConfig.yaml"),
		// and other system related options
		ResultCallback: func(s *resolve.HostEntry) {
			rawCount += 1
			rawSubdomain = append(rawSubdomain, s.Host)
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
	err = subfinder.RunEnumeration()
	if err != nil {
		logger.SlogError(fmt.Sprintf("%v error: %v", p.GetName(), err))
		return err
	}
	subdomainVerificationResult := make(chan string, 100)
	go utils.DNS.KsubdomainVerify(rawSubdomain, subdomainVerificationResult, 1*time.Hour)

	// 读取结果
	for result := range subdomainVerificationResult {
		subdomainResult := utils.DNS.KsubdomainResultToStruct(result)
		if subdomainResult.Host != "" {
			verificationCount += 1
			p.Result <- subdomainResult
		} else {
			logger.SlogDebugLocal(result)
		}
	}
	logger.SlogInfoLocal(fmt.Sprintf("%v plugin result: %v original quantity: %v verification quantity: %v", p.GetName(), target, rawCount, verificationCount))
	return nil
}

func (p *Plugin) Clone() interfaces.Plugin {
	return &Plugin{
		Name:   p.Name,
		Module: p.Module,
	}
}

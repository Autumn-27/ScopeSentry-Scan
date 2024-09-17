// subfinder-------------------------------------
// @file      : subfinder.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:35
// -------------------------------------------

package subfinder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
	"io"
	"log"
	"path/filepath"
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

func (p *Plugin) Install() bool {
	return true
}

func (p *Plugin) Check() bool {
	return true
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
	subfinderOpts := &runner.Options{
		Threads:            10, // Thread controls the number of threads to use for active enumerations
		Timeout:            30, // Timeout is the seconds to wait for sources to respond
		MaxEnumerationTime: 10, // MaxEnumerationTime is the maximum amount of time in mins to wait for enumeration
		ProviderConfig:     filepath.Join(system.ConfigDir, "subfinderConfig.yaml"),
		// and other system related options
	}

	// disable timestamps in logs / configure logger
	log.SetFlags(0)

	subfinder, err := runner.NewRunner(subfinderOpts)
	if err != nil {
		system.SlogError(fmt.Sprintf("failed to create subfinder runner: %v", err))
	}
	output := &bytes.Buffer{}
	// To run subdomain enumeration on a single domain
	if err = subfinder.EnumerateSingleDomainWithCtx(context.Background(), target, []io.Writer{output}); err != nil {
		system.SlogError(fmt.Sprintf("failed to enumerate single domain:%v", err))
	}
	return nil
}

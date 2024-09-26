// fingerprintx-------------------------------------
// @file      : fingerprintx.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/26 21:20
// -------------------------------------------

package fingerprintx

import (
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
)

type Plugin struct {
	Name      string
	Module    string
	Parameter string
	Result    chan interface{}
}

func NewPlugin() *Plugin {
	return &Plugin{
		Name:   "fingerprintx",
		Module: "PortFingerprint",
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

func (p *Plugin) Execute(input interface{}) (interface{}, error) {
	//portAlive, ok := input.(types.PortAlive)
	//if !ok {
	//	logger.SlogError(fmt.Sprintf("%v error: %v input is not a string\n", p.Name, input))
	//	return nil, errors.New("input is not a string")
	//}
	//fxConfig := scan.Config{
	//	DefaultTimeout: time.Duration(3) * time.Second,
	//	FastMode:       false,
	//	Verbose:        false,
	//	UDP:            false,
	//}
	return nil, nil
}

func (p *Plugin) Clone() interfaces.Plugin {
	return &Plugin{
		Name:   p.Name,
		Module: p.Module,
	}
}

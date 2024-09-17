// subfinder-------------------------------------
// @file      : subfinder.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:35
// -------------------------------------------

package subfinder

type Plugin struct {
	Name   string
	Module string
	Result chan interface{}
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

func (p *Plugin) Execute(input interface{}) error {
	//target, ok := input.(string)
	//if !ok {
	//	logger.SlogError(fmt.Sprintf("%v error: %v input is not a string\n", p.Name, input))
	//	return errors.New("input is not a string")
	//}

	return nil
}

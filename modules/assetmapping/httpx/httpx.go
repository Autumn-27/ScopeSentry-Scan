// httpx-------------------------------------
// @file      : httpx.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/28 15:12
// -------------------------------------------

package httpx

import (
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
)

type Plugin struct {
	Name      string
	Module    string
	Parameter string
	Id        string
	Result    chan interface{}
}

func NewPlugin() *Plugin {
	return &Plugin{
		Name:   "httpx",
		Module: "AssetMapping",
	}
}

func (p *Plugin) SetId(id string) {
	p.Id = id
}

func (p *Plugin) GetId() string {
	return p.Id
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
	asset, ok := input.(types.AssetOther)
	if !ok {
		logger.SlogError(fmt.Sprintf("%v error: %v input is not a string\n", p.Name, input))
		return nil, errors.New("input is not a string")
	}
	if asset.Type != "http" {
		p.Result <- asset
	} else {
		httpxResultsHandler := func(r types.AssetHttp) {
			p.Result <- r
		}
		var url string
		if asset.Port != "" {
			url = asset.Host + ":" + asset.Port
		} else {
			url = asset.Host
		}
		utils.Requests.Httpx(url, httpxResultsHandler)
	}
	return nil, nil
}

func (p *Plugin) Clone() interfaces.Plugin {
	return &Plugin{
		Name:   p.Name,
		Module: p.Module,
		Id:     p.Id,
	}
}

// fingerprintx-------------------------------------
// @file      : fingerprintx.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/26 21:20
// -------------------------------------------

package fingerprintx

import (
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"github.com/praetorian-inc/fingerprintx/pkg/plugins"
	"github.com/praetorian-inc/fingerprintx/pkg/scan"
	"net/netip"
	"strconv"
	"strings"
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
		Name:     "fingerprintx",
		Module:   "PortFingerprint",
		PluginId: "648a6f49eed57b1737ac702e02985b00",
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

func (p *Plugin) Log(msg string, tp ...string) {
	var logTp string
	if len(tp) > 0 {
		logTp = tp[0] // 使用传入的参数
	} else {
		logTp = "i"
	}
	logger.PluginsLog(fmt.Sprintf("[Plugins %v] %v", p.GetName(), msg), logTp, p.GetModule(), p.GetPluginId())
}

func (p *Plugin) GetParameter() string {
	return p.Parameter
}

func (p *Plugin) Execute(input interface{}) (interface{}, error) {
	asset, ok := input.(*types.AssetOther)
	if !ok {
		logger.SlogError(fmt.Sprintf("%v error: %v input is not types.AssetOther\n", p.Name, input))
		return nil, errors.New("input is not types.AssetOther")
	}
	if asset.Service != "" {
		// 如果service不为空，说明有其他插件检出，直接返回
		return nil, nil
	}
	fxConfig := scan.Config{
		DefaultTimeout: time.Duration(3) * time.Second,
		FastMode:       false,
		Verbose:        false,
		UDP:            false,
	}
	ip, _ := netip.ParseAddr(asset.IP)
	portUint64, err := strconv.ParseUint(asset.Port, 10, 16)
	if err != nil {
		fmt.Println("转换错误:", err)
		logger.SlogError(fmt.Sprintf("%v 端口转换错误: %v ", p.GetName(), err))
		return nil, err
	}
	target := plugins.Target{
		Address: netip.AddrPortFrom(ip, uint16(portUint64)),
		Host:    asset.Host,
	}
	fingerResults, err := scan.ScanTargets([]plugins.Target{target}, fxConfig)
	for _, fingerResult := range fingerResults {
		if strings.Contains(fingerResult.Protocol, "http") {
			asset.Type = "http"
		} else {
			asset.Type = "other"
		}
		asset.Service = fingerResult.Protocol
		asset.TLS = fingerResult.TLS
		asset.Transport = fingerResult.Transport
		asset.Version = fingerResult.Version
		asset.Raw = fingerResult.Raw
		asset.Time = utils.Tools.GetTimeNow()
		asset.LastScanTime = asset.Time
		return nil, nil
	}
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

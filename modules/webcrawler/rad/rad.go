// rad-------------------------------------
// @file      : rad.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/13 15:55
// -------------------------------------------

package rad

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"os"
	"path/filepath"
	"runtime"
)

type Plugin struct {
	Name        string
	Module      string
	Parameter   string
	PluginId    string
	Result      chan interface{}
	Custom      interface{}
	RadFileName string
	OsType      string
	RadDir      string
	TaskId      string
}

func NewPlugin() *Plugin {
	osType := runtime.GOOS
	var path string
	var dir string
	switch osType {
	case "windows":
		path = "rad.exe"
		dir = "win"
	case "linux":
		path = "rad"
		dir = "linux"
	default:
		dir = "darwin"
		path = "rad"
	}
	return &Plugin{
		Name:        "rad",
		Module:      "URLScan",
		PluginId:    "9669d0dcc52a5ca6dbbe580ffc99c364",
		RadFileName: path,
		RadDir:      dir,
		OsType:      osType,
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
	radPath := filepath.Join(global.ExtDir, "rad")
	if err := os.MkdirAll(radPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create rad folder:", err))
		return err
	}
	targetPath := filepath.Join(radPath, "target")
	if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create targetPath folder:", err))
		return err
	}
	resultPath := filepath.Join(radPath, "result")
	if err := os.MkdirAll(resultPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create resultPath folder:", err))
		return err
	}
	RadExecPath := filepath.Join(radPath, p.RadFileName)
	if _, err := os.Stat(RadExecPath); os.IsNotExist(err) {
		_, err := utils.Tools.HttpGetDownloadFile(fmt.Sprintf("%v/%v/%v", "https://raw.githubusercontent.com/Autumn-27/ScopeSentry-Scan/refs/heads/1.5-restructure/tools", p.RadDir, p.RadFileName), RadExecPath)
		if err != nil {
			_, err = utils.Tools.HttpGetDownloadFile(fmt.Sprintf("%v/%v/%v", "https://gitee.com/constL/ScopeSentry-Scan/raw/main/tools", p.RadDir, p.RadFileName), RadExecPath)
			if err != nil {
				return err
			}
		}
		if p.OsType == "linux" {
			err = os.Chmod(RadExecPath, 0755)
			if err != nil {
				logger.SlogError(fmt.Sprintf("Chmod rad Tool Fail: %s", err))
				return err
			}
		}
	}
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

	return nil, nil
}

func (p *Plugin) Clone() interfaces.Plugin {
	return &Plugin{
		Name:        p.Name,
		Module:      p.Module,
		PluginId:    p.PluginId,
		Custom:      p.Custom,
		RadFileName: p.RadFileName,
		RadDir:      p.RadDir,
		OsType:      p.OsType,
		TaskId:      p.TaskId,
	}
}

// ksubdomain-------------------------------------
// @file      : ksubdomain.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/19 23:04
// -------------------------------------------

package ksubdomain

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"os"
	"path/filepath"
	"runtime"
)

type Plugin struct {
	Name      string
	Module    string
	Parameter string
	Result    chan interface{}
}

func NewPlugin() *Plugin {
	return &Plugin{
		Name:   "ksubdomain",
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
	ksubdomainPath := filepath.Join(config.ExtDir, "ksubdomain")
	if err := os.MkdirAll(ksubdomainPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create ksubdomain folder:", err))
		return false
	}
	targetPath := filepath.Join(ksubdomainPath, "target")
	if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create targetPath folder:", err))
		return false
	}
	resultPath := filepath.Join(ksubdomainPath, "result")
	if err := os.MkdirAll(resultPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create resultPath folder:", err))
		return false
	}
	osType := runtime.GOOS
	// 判断操作系统类型
	var path string
	var dir string
	switch osType {
	case "windows":
		path = "ksubdomain.exe"
		dir = "win"
	case "linux":
		path = "ksubdomain"
		dir = "linux"
	default:
		dir = "darwin"
		path = "ksubdomain"
	}
	KsubdomainPath := filepath.Join(config.ExtDir, "ksubdomain")
	KsubdomainExecPath := filepath.Join(KsubdomainPath, path)
	if _, err := os.Stat(KsubdomainExecPath); os.IsNotExist(err) {
		_, err := utils.Tools.HttpGetDownloadFile(fmt.Sprintf("%v/%v/%v", "https://raw.githubusercontent.com/Autumn-27/ScopeSentry-Scan/main/tools", dir, path), KsubdomainExecPath)
		if err != nil {
			_, err = utils.Tools.HttpGetDownloadFile(fmt.Sprintf("%v/%v/%v", "https://gitee.com/constL/ScopeSentry-Scan/raw/main/tools", dir, path), KsubdomainExecPath)
			if err != nil {
				return false
			}
		}
		if osType == "linux" {
			err = os.Chmod(KsubdomainExecPath, 0755)
			if err != nil {
				logger.SlogError(fmt.Sprintf("Chmod ksubdomain Tool Fail: %s", err))
				return false
			}
		}
	}
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
	return nil
}

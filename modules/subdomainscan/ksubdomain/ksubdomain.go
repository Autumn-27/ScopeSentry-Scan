// ksubdomain-------------------------------------
// @file      : ksubdomain.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/19 23:04
// -------------------------------------------

package ksubdomain

import (
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"os"
	"path/filepath"
	"runtime"
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

func (p *Plugin) Install() error {
	ksubdomainPath := filepath.Join(global.ExtDir, "ksubdomain")
	if err := os.MkdirAll(ksubdomainPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create ksubdomain folder:", err))
		return err
	}
	targetPath := filepath.Join(ksubdomainPath, "target")
	if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create targetPath folder:", err))
		return err
	}
	resultPath := filepath.Join(ksubdomainPath, "result")
	if err := os.MkdirAll(resultPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create resultPath folder:", err))
		return err
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
	KsubdomainPath := filepath.Join(global.ExtDir, "ksubdomain")
	KsubdomainExecPath := filepath.Join(KsubdomainPath, path)
	if _, err := os.Stat(KsubdomainExecPath); os.IsNotExist(err) {
		_, err := utils.Tools.HttpGetDownloadFile(fmt.Sprintf("%v/%v/%v", "https://raw.githubusercontent.com/Autumn-27/ScopeSentry-Scan/main/tools", dir, path), KsubdomainExecPath)
		if err != nil {
			_, err = utils.Tools.HttpGetDownloadFile(fmt.Sprintf("%v/%v/%v", "https://gitee.com/constL/ScopeSentry-Scan/raw/main/tools", dir, path), KsubdomainExecPath)
			if err != nil {
				return err
			}
		}
		if osType == "linux" {
			err = os.Chmod(KsubdomainExecPath, 0755)
			if err != nil {
				logger.SlogError(fmt.Sprintf("Chmod ksubdomain Tool Fail: %s", err))
				return err
			}
		}
	}
	return nil
}

func (p *Plugin) Check() error {
	rawSubdomain := []string{"scope-sentry.top"}
	subdomainVerificationResult := make(chan string, 1)
	verificationCount := 0
	go utils.DNS.KsubdomainVerify(rawSubdomain, subdomainVerificationResult, 5*time.Minute)
	for result := range subdomainVerificationResult {
		subdomainResult := utils.DNS.KsubdomainResultToStruct(result)
		if subdomainResult.Host != "" {
			verificationCount += 1
		} else {
			logger.SlogErrorLocal(result)
		}
	}
	if verificationCount == 0 {
		return fmt.Errorf("ksubdomain run error")
	} else {
		return nil
	}
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
	// 判断是否为泛解析，记录泛解析的ip，然后爆破之后
	return nil
}

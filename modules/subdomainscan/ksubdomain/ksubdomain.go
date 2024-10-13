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
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"os"
	"path/filepath"
	"runtime"
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
}

func NewPlugin() *Plugin {
	return &Plugin{
		Name:     "ksubdomain",
		Module:   "SubdomainScan",
		PluginId: "e8f55f5e0e9f4af1ca40eb19048b8c82",
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
	KsubdomainExecPath := filepath.Join(ksubdomainPath, path)
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

func (p *Plugin) Execute(input interface{}) (interface{}, error) {
	target, ok := input.(string)
	if !ok {
		logger.SlogError(fmt.Sprintf("%v error: %v input is not a string\n", p.Name, input))
		return nil, errors.New("input is not a string")
	}
	wildcardSubdomainResults := wildcardDNSRecords(target)
	wildcardDNSRecordsLen := len(wildcardSubdomainResults)
	parameter := p.GetParameter()
	var subfile string
	executionTimeout := 60
	if parameter != "" {
		args, err := utils.Tools.ParseArgs(parameter, "subfile", "et")
		if err != nil {
		} else {
			for key, value := range args {
				switch key {
				case "subfile":
					subfile = value
				case "et":
					executionTimeout, _ = strconv.Atoi(value)
				}
			}
		}
	} else {
		logger.SlogError(fmt.Sprintf("ksubdomain 运行失败: 没有提供子域名字典，请查看任务配置"))
		return nil, nil
	}
	subfile = filepath.Join(global.DictPath, "subdomain", subfile)
	subDictChan := make(chan string, 10)
	go func() {
		err := utils.Tools.ReadFileLineByLine(subfile, subDictChan)
		if err != nil {
			logger.SlogInfoLocal(fmt.Sprintf("%v", err))
		}
	}()
	rawSubdomain := []string{}
	dotIndex := strings.Index(target, "*.")
	// 读取子域名字典
	for result := range subDictChan {
		if dotIndex != -1 {
			tmpDomain := strings.Replace(target, "*", result, -1)
			rawSubdomain = append(rawSubdomain, tmpDomain)
		} else {
			rawSubdomain = append(rawSubdomain, result+"."+target)
		}
	}
	// 将原始域名增加到子域名列表
	rawSubdomain = append(rawSubdomain, target)
	// 拼接完子域名之后开始运行验证子域名
	subdomainVerificationResult := make(chan string, 100)
	go utils.DNS.KsubdomainVerify(rawSubdomain, subdomainVerificationResult, time.Duration(executionTimeout)*time.Minute)
	verificationCount := 0
	// 读取结果
	for result := range subdomainVerificationResult {
		subdomainResult := utils.DNS.KsubdomainResultToStruct(result)
		if subdomainResult.Host != "" {
			// wildcardDNSRecords记录的是生成的随机域名的解析结果，如果大于2，认为是存在泛解析，将此解析ip跳过
			if wildcardDNSRecordsLen >= 2 {
				// 如果subdomainResult.IP中的某个IP存在于泛解析记录中，跳过
				if isIPInWildcard(subdomainResult.IP, wildcardSubdomainResults) {
					// 发现存在泛解析记录中的IP，跳过该结果
					logger.SlogDebugLocal(fmt.Sprintf("%v 发现泛解析域名, 子域名: %v  IP: %v", target, subdomainResult.Host, subdomainResult.IP))
					continue
				}
			}
			verificationCount += 1
			p.Result <- subdomainResult
		} else {
			logger.SlogErrorLocal(result)
		}
	}
	logger.SlogInfoLocal(fmt.Sprintf("%v plugin result: %v original quantity: %v verification quantity: %v", p.GetName(), target, len(rawSubdomain), verificationCount))

	return nil, nil
}

func wildcardDNSRecords(domain string) []string {
	targets := []string{}
	for i := 0; i < 3; i++ {
		dotIndex := strings.Index(domain, "*.")
		subdomain := util.GenerateRandomString(6) + "." + domain
		if dotIndex != -1 {
			subdomain = strings.Replace(domain, "*", util.GenerateRandomString(6), -1)
		}
		targets = append(targets, subdomain)
	}
	subdomainVerificationResult := make(chan string, 1)
	go utils.DNS.KsubdomainVerify(targets, subdomainVerificationResult, 1*time.Hour)
	var results []string
	// 读取结果
	for result := range subdomainVerificationResult {
		subdomainResult := utils.DNS.KsubdomainResultToStruct(result)
		if subdomainResult.Host != "" {
			logger.SlogInfoLocal(fmt.Sprintf("%v 发现泛解析IP：%v", domain, subdomainResult.IP))
			results = append(results, subdomainResult.IP...)
		}
	}
	return results
}

// 判断是否有IP在泛解析记录中
func isIPInWildcard(ipList []string, wildcardDNSRecords []string) bool {
	for _, ip := range ipList {
		for _, wildcardIP := range wildcardDNSRecords {
			// 如果某个ip在wildcardDNSRecords中，立即返回true
			if ip == wildcardIP {
				return true
			}
		}
	}
	return false
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

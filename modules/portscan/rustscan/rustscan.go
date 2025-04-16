// rustscan-------------------------------------
// @file      : rustscan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/25 20:26
// -------------------------------------------

package rustscan

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Plugin struct {
	Name         string
	Module       string
	Parameter    string
	PluginId     string
	Result       chan interface{}
	RustFileName string
	RustDir      string
	OsType       string
	Custom       interface{}
	TaskId       string
	TaskName     string
}

func NewPlugin() *Plugin {
	osType := runtime.GOOS
	// 判断操作系统类型
	var path string
	var dir string
	switch osType {
	case "windows":
		path = "rustscan.exe"
		dir = "win"
	case "linux":
		path = "rustscan"
		dir = "linux"
	default:
		dir = "darwin"
		path = "rustscan"
	}
	return &Plugin{
		Name:         "RustScan",
		Module:       "PortScan",
		RustFileName: path,
		RustDir:      dir,
		OsType:       osType,
		PluginId:     "66b4ddeb983387df2b7ee7726653874d",
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
func (p *Plugin) Log(msg string, tp ...string) {
	var logTp string
	if len(tp) > 0 {
		logTp = tp[0] // 使用传入的参数
	} else {
		logTp = "i"
	}
	logger.PluginsLog(fmt.Sprintf("[Plugins %v] %v", p.GetName(), msg), logTp, p.GetModule(), p.GetPluginId())
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
func (p *Plugin) UnInstall() error {
	return nil
}
func (p *Plugin) Install() error {
	rustscanPath := filepath.Join(global.ExtDir, "rustscan")
	if err := os.MkdirAll(rustscanPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create rustscan folder:", err))
		return err
	}
	targetPath := filepath.Join(rustscanPath, "target")
	if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create targetPath folder:", err))
		return err
	}
	resultPath := filepath.Join(rustscanPath, "result")
	if err := os.MkdirAll(resultPath, os.ModePerm); err != nil {
		logger.SlogError(fmt.Sprintf("Failed to create resultPath folder:", err))
		return err
	}

	RustscanPath := filepath.Join(global.ExtDir, "rustscan")
	RustscanExecPath := filepath.Join(RustscanPath, p.RustFileName)
	if _, err := os.Stat(RustscanExecPath); os.IsNotExist(err) {
		_, err := utils.Tools.HttpGetDownloadFile(fmt.Sprintf("%v/%v/%v", "https://raw.githubusercontent.com/Autumn-27/ScopeSentry-Scan/main/tools", p.RustDir, p.RustFileName), RustscanExecPath)
		if err != nil {
			_, err = utils.Tools.HttpGetDownloadFile(fmt.Sprintf("%v/%v/%v", "https://gitee.com/constL/ScopeSentry-Scan/raw/main/tools", p.RustDir, p.RustFileName), RustscanExecPath)
			if err != nil {
				return err
			}
		}
		if p.OsType == "linux" {
			err = os.Chmod(RustscanExecPath, 0755)
			if err != nil {
				logger.SlogError(fmt.Sprintf("Chmod rustscan Tool Fail: %s", err))
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

var ipv6Regex = regexp.MustCompile(`^\[([0-9a-fA-F:]+)\]:(\d+)$`)

func (p *Plugin) Execute(input interface{}) (interface{}, error) {
	domainSkip, ok := input.(types.DomainSkip)
	if !ok {
		//logger.SlogError(fmt.Sprintf("%v error: %v input is not types.DomainSkip\n", p.Name, input))
		return nil, errors.New("input is not types.DomainSkip")
	}
	parameter := p.GetParameter()
	PortBatchSize := "600"
	PortTimeout := "3000"
	// 如果没有找到端口 默认扫描top1000
	executionTimeout := 60
	PortRange := ""
	maxPort := 200
	if parameter != "" {
		args, err := utils.Tools.ParseArgs(parameter, "b", "t", "port", "et", "maxport")
		if err != nil {
		} else {
			for key, value := range args {
				if value != "" {
					switch key {
					case "b":
						PortBatchSize = value
					case "t":
						PortTimeout = value
					case "port":
						if domainSkip.Skip {
							PortRange = "80,443"
						} else {
							PortRange = value
						}
					case "et":
						executionTimeout, _ = strconv.Atoi(value)
					case "maxport":
						maxPort, _ = strconv.Atoi(value)
					default:
						continue
					}
				}
			}
		}
	}
	if PortRange == "" {
		p.Log(fmt.Sprintf("PortRange is nul, parameter:%v", parameter), "e")
		return nil, nil
	}
	start := time.Now()
	args := []string{"-b", PortBatchSize, "-t", PortTimeout, "-a", domainSkip.Domain, "-r", PortRange, "--accessible", "--scripts", "None"}
	rustScanExecPath := filepath.Join(filepath.Join(global.ExtDir, "rustscan"), p.RustFileName)
	// 假设你已经有获取 TaskID 的逻辑
	taskContext := contextmanager.GlobalContextManagers.GetContext(p.GetTaskId())

	// 为命令设置一个超时时间
	timeout := time.Duration(executionTimeout) * time.Minute // 例如，设置为30分钟超时
	ctx, cancel := context.WithTimeout(taskContext, timeout)
	defer cancel()
	logger.SlogInfoLocal(fmt.Sprintf("[Plugin %v]begin scan %v", p.GetName(), domainSkip.Domain))
	cmd := exec.CommandContext(ctx, rustScanExecPath, args...)
	stdout, err := cmd.StdoutPipe()
	defer stdout.Close()
	if err != nil {
		logger.SlogError(fmt.Sprintf("RustScan StdoutPipe error： %v", err))
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		logger.SlogWarnLocal(fmt.Sprintf("RustScan cmd.Start error： %v", err))
		return nil, err
	}
	var wg sync.WaitGroup
	portFlag := 0
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		r := scanner.Text()
		logger.SlogDebugLocal(fmt.Sprintf("%v rustscan result: %v", domainSkip.Domain, r))
		if strings.Contains(r, "File limit higher than batch size") {
			continue
		}
		if strings.Contains(r, "Looks like I didn't find any open ports") {
			continue
		}
		if strings.Contains(r, "*I used") {
			continue
		}
		if strings.Contains(r, "Alternatively, increase") {
			continue
		}
		if strings.Contains(r, "Open") {
			portFlag += 1
			if !domainSkip.CIDR {
				if portFlag > maxPort {
					p.Log(fmt.Sprintf("target %v open port number > max port: %v", domainSkip.Domain, portFlag), "w")
					cancel()
					return nil, nil
				}
			}
			// 端口开放
			openIpPort := strings.SplitN(r, " ", 2)
			// 检查是否是IPv6地址
			if match := ipv6Regex.FindStringSubmatch(openIpPort[1]); match != nil {
				var result types.PortAlive
				if domainSkip.CIDR {
					result = types.PortAlive{
						Host: match[1],
						IP:   match[1], // IPv6地址
						Port: match[2], // 端口
					}
				} else {
					result = types.PortAlive{
						Host: domainSkip.Domain,
						IP:   match[1], // IPv6地址
						Port: match[2], // 端口
					}
				}
				logger.SlogInfoLocal(fmt.Sprintf("%v %v Port alive: %v", domainSkip.Domain, match[1], match[2]))
				p.Result <- result
				continue
			} else {
				// 处理IPv4地址的情况
				openPort := strings.SplitN(openIpPort[1], ":", 2)
				var result types.PortAlive
				if domainSkip.CIDR {
					result = types.PortAlive{
						Host: openPort[0],
						IP:   openPort[0], // IPv4地址
						Port: openPort[1], // 端口
					}
				} else {
					result = types.PortAlive{
						Host: domainSkip.Domain,
						IP:   openPort[0], // IPv4地址
						Port: openPort[1], // 端口
					}
				}
				logger.SlogInfoLocal(fmt.Sprintf("%v %v Port alive: %v", domainSkip.Domain, openPort[0], openPort[1]))
				p.Result <- result
				continue
			}
		}
		if strings.Contains(r, "->") {
			p.Log(fmt.Sprintf("%v Port alive: %v", domainSkip.Domain, r))
			continue
		}
		logger.SlogDebugLocal(fmt.Sprintf("%v PortScan error: %v", domainSkip.Domain, r))
	}
	if err := scanner.Err(); err != nil {
		logger.SlogWarnLocal(fmt.Sprintf("%v RustScan scanner.Err error： %v", domainSkip.Domain, err))
		wg.Wait()
		return nil, nil
	}
	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		logger.SlogDebugLocal(fmt.Sprintf("%v RustScan cmd.Wait error： %v", domainSkip.Domain, err))
	}
	wg.Wait()
	end := time.Now()
	duration := end.Sub(start)
	p.Log(fmt.Sprintf("target %v running time:%v", domainSkip.Domain, duration))
	return nil, nil
}

func (p *Plugin) Clone() interfaces.Plugin {
	return &Plugin{
		Name:         p.Name,
		Module:       p.Module,
		RustDir:      p.RustDir,
		RustFileName: p.RustFileName,
		PluginId:     p.PluginId,
		Custom:       p.Custom,
		TaskId:       p.TaskId,
	}
}

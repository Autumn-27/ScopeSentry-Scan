// targetparser-------------------------------------
// @file      : targetparser.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:53
// -------------------------------------------

package targetparser

import (
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"golang.org/x/net/idna"
	"net"
	"net/url"
	"regexp"
	"strings"
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
		Name:     "TargetParser",
		Module:   "TargetHandler",
		PluginId: "7bbaec6487f51a9aafeff4720c7643f0",
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
func (p *Plugin) UnInstall() error {
	return nil
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

// 判断字符串是否是有效的域名
func isValidDomain(domain string) bool {
	// 域名正则表达式（简化版）
	domainRegex := `^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`
	re := regexp.MustCompile(domainRegex)
	return re.MatchString(domain)
}

// 转换中文域名为 ASCII 兼容格式
func toASCII(domain string) (string, error) {
	return idna.ToASCII(domain)
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

var ipv6Regex = regexp.MustCompile(`^\[([0-9a-fA-F:]+)\]:(\d+)$`)

// Execute
//
//  1. IP 地址
//     输入: "192.168.1.1"
//     输出: "192.168.1.1"
//
//  2. 带协议的 URL（没有端口号）
//     输入: "http://example.com"
//     输出: "example.com"
//
//  3. 带协议的 URL（带端口号）
//     输入: "http://example.com:8080"
//     输出: "example.com"
//     输出: "example.com:8080"
//
//  4. 带通配符的域名
//     输入: "*.example.com"
//     输出: "*.example.com"
//     输入: "wda.*.example.com"
//     输出: "wda.*.example.com"
//
//  5. 不带协议的有效域名
//     输入: "example.com"
//     输出: "example.com"
//
//  6. 中文域名
//     输入: "例子.com"
//     输出: "xn--fsq.com"
//
//  7. 无效输入, 直接返回
//     输入: "例子公司"
//     输出: "Invalid input: 例子公司"
func (p *Plugin) Execute(input interface{}) (interface{}, error) {
	target, ok := input.(string)
	if !ok {
		logger.SlogError(fmt.Sprintf("%v error: %v input is not a string\n", p.Name, input))
		return nil, errors.New("input is not a string")
	}
	// 检查是否是 IP 地址
	if net.ParseIP(target) != nil {
		// 如果是纯 IP 地址
		p.Result <- target
		return nil, nil
	}
	// 输入CIDR:开头的 不进行分布式 直接将数据发送到rustscan进行扫描
	if strings.HasPrefix(target, "CIDR:") {
		tg := strings.Replace(target, "CIDR:", "", 1)
		tmpDominSkip := types.DomainSkip{Domain: tg, Skip: false, CIDR: true}
		p.Result <- tmpDominSkip
		return nil, nil
	}
	if strings.HasPrefix(target, "CMP:") {
		name := strings.Replace(target, "CMP:", "", 1)
		p.Result <- types.Company{Name: name}
		return nil, nil
	}
	if strings.HasPrefix(target, "ICP:") {
		name := strings.Replace(target, "ICP:", "", 1)
		p.Result <- types.Company{Name: name}
		return nil, nil
	}

	if strings.HasPrefix(target, "APP-ID:") {
		id := strings.Replace(target, "APP-ID:", "", 1)
		p.Result <- types.APP{
			BundleID: id,
		}
		return nil, nil
	}
	if strings.HasPrefix(target, "APP:") {
		name := strings.Replace(target, "APP:", "", 1)
		p.Result <- types.APP{
			Name: name,
		}
		return nil, nil
	}
	rootDomain, err := utils.Tools.GetRootDomain(target)
	if err == nil {
		if net.ParseIP(rootDomain) == nil {
			flag := results.Duplicate.RootDomain(rootDomain, p.TaskId)
			if flag {
				p.Result <- types.RootDomain{Domain: rootDomain}
			}
		}
	}

	// 尝试解析 URL
	parsedURL, err := url.Parse(target)
	if err == nil && parsedURL.Host != "" {
		host := parsedURL.Host
		rawHost := parsedURL.Host
		// 检查是否有端口号
		if strings.Contains(host, ":") {
			// 分割主机名和端口号
			hostParts := strings.Split(host, ":")
			ipOrDomain := hostParts[0]
			//port := hostParts[1]
			rawHost = hostParts[0]
			p.Result <- ipOrDomain
			//// 判断主机部分是否为 IP 地址
			//if net.ParseIP(ipOrDomain) != nil {
			//	// 处理 IP 地址 + 端口号
			//	p.Result <- ipOrDomain
			//	//p.Result <- ipOrDomain + ":" + port
			//} else {
			//	// 处理域名 + 端口号
			//	p.Result <- ipOrDomain
			//	//p.Result <- host
			//}
			hostPort := types.PortAlive{
				Host: ipOrDomain,
				IP:   "",
				Port: hostParts[1],
			}
			p.Result <- hostPort
		} else {
			// 检查主机部分是否是 IP 地址
			//if net.ParseIP(host) != nil {
			//	// 处理纯 IP 地址
			//	p.Result <- host
			//} else {
			//	// 处理域名
			//	p.Result <- host
			//}
			p.Result <- host
		}
		path := parsedURL.Path
		// 检查路径是否全部由 `/` 组成
		if strings.Trim(path, "/") == "" {
		} else {
			tmp := types.AssetOther{
				Host:    rawHost,
				Port:    parsedURL.Port(),
				UrlPath: path,
				Type:    "http",
			}
			// 转换为 []interface{}
			var interfaceList []interface{}
			interfaceList = append(interfaceList, tmp)
			p.Result <- interfaceList
		}
		return nil, nil
	}

	// 处理 `*.domain.com` 或其他不包含协议的域名
	if strings.HasPrefix(target, "*.") || strings.Contains(target, ".*.") {
		p.Result <- target
		return nil, nil
	}
	// 处理ipv6
	if match := ipv6Regex.FindStringSubmatch(target); match != nil {
		result := types.PortAlive{
			Host: match[1],
			IP:   match[1], // IPv6地址
			Port: match[2], // 端口
		}
		p.Result <- match[1]
		p.Result <- result
		return nil, nil
	}

	// 处理 127.0.0.1:666
	if strings.Contains(target, ":") {
		data := strings.Split(target, ":")
		if len(data) == 2 {
			result := types.PortAlive{
				Host: data[0],
				IP:   "",
				Port: data[1], // 端口
			}
			p.Result <- data[0]
			p.Result <- result
			return nil, nil
		}
	}

	// 处理不包含协议的域名
	if isValidDomain(target) {
		asciiDomain, err := toASCII(target)
		if err == nil {
			p.Result <- asciiDomain
		} else {
			p.Result <- target
		}
	} else if net.ParseIP(target) != nil {
		p.Result <- target
	} else {
		// 处理无效输入
		logger.SlogInfoLocal(fmt.Sprintf("%v error Invalid input:%v ", p.Name, input))
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

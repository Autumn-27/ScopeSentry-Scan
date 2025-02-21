// modules-------------------------------------
// @file      : manage.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 21:51
// -------------------------------------------

package modules

import (
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/assethandle"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/assetmapping"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/dirscan"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/portfingerprint"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/portscan"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/portscanpreparation"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/subdomainscan"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/subdomainsecurity"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/targethandler"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlscan"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlsecurity"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/vulnerabilityscan"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/webcrawler"
)

func CreateScanProcess(op *options.TaskOptions) interfaces.ModuleRunner {
	// 初始化 InputChan
	op.InputChan = make(map[string]chan interface{})

	// 漏洞扫描模块
	op.ModuleRunWg.Add(1)
	vulnerabilityModule := vulnerabilityscan.NewRunner(op, nil)
	vulnerabilityInputChan := make(chan interface{}, 5000)
	vulnerabilityModule.SetInput(vulnerabilityInputChan)
	op.InputChan["Vulnerability"] = vulnerabilityInputChan

	// 目录扫描模块
	op.ModuleRunWg.Add(1)
	dirScanModule := dirscan.NewRunner(op, vulnerabilityModule)
	dirScanInputChan := make(chan interface{}, 5000)
	dirScanModule.SetInput(dirScanInputChan)
	op.InputChan["DirScan"] = dirScanInputChan

	// url安全模块
	op.ModuleRunWg.Add(1)
	urlSecurityModule := urlsecurity.NewRunner(op, dirScanModule)
	urlSecurityInputChan := make(chan interface{}, 5000)
	urlSecurityModule.SetInput(urlSecurityInputChan)
	op.InputChan["UrlSecurity"] = urlSecurityInputChan

	// 爬虫模块
	op.ModuleRunWg.Add(1)
	webCrawlerModule := webcrawler.NewRunner(op, urlSecurityModule)
	WebCrawlerInputChan := make(chan interface{}, 5000)
	webCrawlerModule.SetInput(WebCrawlerInputChan)
	op.InputChan["WebCrawler"] = WebCrawlerInputChan

	// url扫描模块
	op.ModuleRunWg.Add(1)
	urlScanModule := urlscan.NewRunner(op, webCrawlerModule)
	urlScanInputChan := make(chan interface{}, 5000)
	urlScanModule.SetInput(urlScanInputChan)
	op.InputChan["UrlScan"] = urlScanInputChan

	// 资产处理模块
	op.ModuleRunWg.Add(1)
	assetHandleModule := assethandle.NewRunner(op, urlScanModule)
	assetHandleInputChan := make(chan interface{}, 5000)
	assetHandleModule.SetInput(assetHandleInputChan)
	op.InputChan["AssetHandle"] = assetHandleInputChan

	// 资产测绘模块
	op.ModuleRunWg.Add(1)
	assetMappingModule := assetmapping.NewRunner(op, assetHandleModule)
	assetMappingInputChan := make(chan interface{}, 5000)
	assetMappingModule.SetInput(assetMappingInputChan)
	op.InputChan["AssetMapping"] = assetMappingInputChan

	// 端口指纹识别模块
	op.ModuleRunWg.Add(1)
	portFingerprintModule := portfingerprint.NewRunner(op, assetMappingModule)
	portFingerprintInputChan := make(chan interface{}, 5000)
	portFingerprintModule.SetInput(portFingerprintInputChan)
	op.InputChan["PortFingerprint"] = portFingerprintInputChan

	// 端口扫描模块
	op.ModuleRunWg.Add(1)
	portScanModule := portscan.NewRunner(op, portFingerprintModule)
	portScanInputChan := make(chan interface{}, 5000)
	portScanModule.SetInput(portScanInputChan)
	op.InputChan["PortScan"] = portScanInputChan

	// 端口扫描预处理模块
	op.ModuleRunWg.Add(1)
	portScanPreparationModule := portscanpreparation.NewRunner(op, portScanModule)
	portScanPreparationInputChan := make(chan interface{}, 5000)
	portScanPreparationModule.SetInput(portScanPreparationInputChan)
	op.InputChan["PortScanPreparation"] = portScanPreparationInputChan

	// 子域名安全模块
	op.ModuleRunWg.Add(1)
	subdomainSecurityModule := subdomainsecurity.NewRunner(op, portScanPreparationModule)
	subdomainSecurityInputChan := make(chan interface{}, 5000)
	subdomainSecurityModule.SetInput(subdomainSecurityInputChan)
	op.InputChan["SubdomainSecurity"] = subdomainSecurityInputChan

	// 子域名扫描模块
	op.ModuleRunWg.Add(1)
	subdomainScanModule := subdomainscan.NewRunner(op, subdomainSecurityModule)
	subdomainScanInputChan := make(chan interface{}, 5000)
	subdomainScanModule.SetInput(subdomainScanInputChan)
	op.InputChan["SubdomainScan"] = subdomainScanInputChan

	// 目标处理模块
	op.ModuleRunWg.Add(1)
	targetHandlerModule := targethandler.NewRunner(op, subdomainScanModule)
	return targetHandlerModule
}

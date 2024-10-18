// plugins-------------------------------------
// @file      : plugins.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:15
// -------------------------------------------

package plugins

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/assethandle/webfingerprint"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/assetmapping/httpx"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/dirscan/sentrydir"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/portfingerprint/fingerprintx"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/portscan/rustscan"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/portscanpreparation/skipcdn"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/subdomainscan/ksubdomain"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/subdomainscan/subfinder"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/subdomainsecurity/subdomaintakeover"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/targethandler/targetparser"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlscan/katana"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlscan/wayback"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlsecurity/sensitive"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/webcrawler/rad"
	"sync"
)

type PluginManager struct {
	plugins map[string]map[string]interfaces.Plugin // 存储插件，按模块和名称索引
	mu      sync.RWMutex
}

var GlobalPluginManager *PluginManager

// NewPluginManager 创建一个新的 PluginManager 实例
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]map[string]interfaces.Plugin),
	}
}

func (pm *PluginManager) RegisterPlugin(module string, name string, plugin interfaces.Plugin) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.plugins[module]; !exists {
		pm.plugins[module] = make(map[string]interfaces.Plugin)
	}
	pm.plugins[module][name] = plugin
}

func (pm *PluginManager) GetPlugin(module, name string) (interfaces.Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if modPlugins, ok := pm.plugins[module]; ok {
		plugin, ok := modPlugins[name]
		if ok {
			return plugin.Clone(), ok // 返回新实例
		} else {
			return nil, false
		}
	}
	return nil, false
}

// InitializePlugins 初始化插件
func (pm *PluginManager) InitializePlugins() error {
	// TargetParser
	targetparserPlugin := targetparser.NewPlugin()
	pm.RegisterPlugin("TargetHandler", targetparserPlugin.Name, targetparserPlugin)
	// SubdomainScan模块
	// subfinder
	subfinderPlugin := subfinder.NewPlugin()
	pm.RegisterPlugin(subfinderPlugin.Module, subfinderPlugin.Name, subfinderPlugin)
	// kusbdomain
	ksubdomainPlugin := ksubdomain.NewPlugin()
	pm.RegisterPlugin(ksubdomainPlugin.Module, ksubdomainPlugin.Name, ksubdomainPlugin)

	// SubdomainSecurity模块
	subdomainTakeoverPlugin := subdomaintakeover.NewPlugin()
	pm.RegisterPlugin(subdomainTakeoverPlugin.Module, subdomainTakeoverPlugin.Name, subdomainTakeoverPlugin)

	// 端口扫描预处理
	skipcdnPlugin := skipcdn.NewPlugin()
	pm.RegisterPlugin(skipcdnPlugin.Module, skipcdnPlugin.Name, skipcdnPlugin)

	// 端口扫描rustscan
	rustscanPlugin := rustscan.NewPlugin()
	pm.RegisterPlugin(rustscanPlugin.Module, rustscanPlugin.Name, rustscanPlugin)

	// 端口指纹识别
	fingerprintxPlugin := fingerprintx.NewPlugin()
	pm.RegisterPlugin(fingerprintxPlugin.Module, fingerprintxPlugin.Name, fingerprintxPlugin)

	// httpx
	httpxPlugin := httpx.NewPlugin()
	pm.RegisterPlugin(httpxPlugin.Module, httpxPlugin.Name, httpxPlugin)

	// WebFingerprint
	webFingerprintPlugin := webfingerprint.NewPlugin()
	pm.RegisterPlugin(webFingerprintPlugin.Module, webFingerprintPlugin.Name, webFingerprintPlugin)

	// katana
	katanaPlugin := katana.NewPlugin()
	pm.RegisterPlugin(katanaPlugin.Module, katanaPlugin.Name, katanaPlugin)

	// wayback
	waybackPlugin := wayback.NewPlugin()
	pm.RegisterPlugin(waybackPlugin.Module, waybackPlugin.Name, waybackPlugin)

	// rad
	radPlugin := rad.NewPlugin()
	pm.RegisterPlugin(radPlugin.Module, radPlugin.Name, radPlugin)
	// sensitive
	sensitiveModule := sensitive.NewPlugin()
	pm.RegisterPlugin(sensitiveModule.Module, sensitiveModule.Name, sensitiveModule)

	// SentryDir
	dirPlugin := sentrydir.NewPlugin()
	pm.RegisterPlugin(dirPlugin.Module, dirPlugin.Name, dirPlugin)
	// 执行插件的安装和check
	for module, plugins := range pm.plugins {
		for name, plugin := range plugins {
			// 调用每个插件的 Install 函数
			if err := plugin.Install(); err != nil {
				return fmt.Errorf("failed to install plugin %s from module %s: %v", name, module, err)
			}

			// 调用每个插件的 Check 函数
			if err := plugin.Check(); err != nil {
				return fmt.Errorf("failed to check plugin %s from module %s: %v", name, module, err)
			}
		}
	}
	return nil
}

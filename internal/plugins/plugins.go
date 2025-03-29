// plugins-------------------------------------
// @file      : plugins.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:15
// -------------------------------------------

package plugins

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
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
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlsecurity/pagemonitoring"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlsecurity/sensitive"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlsecurity/trufflehog"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/vulnerabilityscan/nuclei"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/webcrawler/rad"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/cloudflare/cfssl/log"
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
	pm.RegisterPlugin("TargetHandler", targetparserPlugin.PluginId, targetparserPlugin)
	// SubdomainScan模块
	// subfinder
	subfinderPlugin := subfinder.NewPlugin()
	pm.RegisterPlugin(subfinderPlugin.Module, subfinderPlugin.PluginId, subfinderPlugin)
	// kusbdomain
	ksubdomainPlugin := ksubdomain.NewPlugin()
	pm.RegisterPlugin(ksubdomainPlugin.Module, ksubdomainPlugin.PluginId, ksubdomainPlugin)

	// SubdomainSecurity模块
	subdomainTakeoverPlugin := subdomaintakeover.NewPlugin()
	pm.RegisterPlugin(subdomainTakeoverPlugin.Module, subdomainTakeoverPlugin.PluginId, subdomainTakeoverPlugin)

	// 端口扫描预处理
	skipcdnPlugin := skipcdn.NewPlugin()
	pm.RegisterPlugin(skipcdnPlugin.Module, skipcdnPlugin.PluginId, skipcdnPlugin)

	// 端口扫描rustscan
	rustscanPlugin := rustscan.NewPlugin()
	pm.RegisterPlugin(rustscanPlugin.Module, rustscanPlugin.PluginId, rustscanPlugin)

	// 端口指纹识别
	fingerprintxPlugin := fingerprintx.NewPlugin()
	pm.RegisterPlugin(fingerprintxPlugin.Module, fingerprintxPlugin.PluginId, fingerprintxPlugin)

	// httpx
	httpxPlugin := httpx.NewPlugin()
	pm.RegisterPlugin(httpxPlugin.Module, httpxPlugin.PluginId, httpxPlugin)

	// WebFingerprint
	webFingerprintPlugin := webfingerprint.NewPlugin()
	pm.RegisterPlugin(webFingerprintPlugin.Module, webFingerprintPlugin.PluginId, webFingerprintPlugin)

	// katana
	katanaPlugin := katana.NewPlugin()
	pm.RegisterPlugin(katanaPlugin.Module, katanaPlugin.PluginId, katanaPlugin)

	// wayback
	waybackPlugin := wayback.NewPlugin()
	pm.RegisterPlugin(waybackPlugin.Module, waybackPlugin.PluginId, waybackPlugin)

	// rad
	radPlugin := rad.NewPlugin()
	pm.RegisterPlugin(radPlugin.Module, radPlugin.PluginId, radPlugin)
	// sensitive
	sensitivePlugin := sensitive.NewPlugin()
	pm.RegisterPlugin(sensitivePlugin.Module, sensitivePlugin.PluginId, sensitivePlugin)

	// trufflehog
	trufflehogPlugin := trufflehog.NewPlugin()
	pm.RegisterPlugin(trufflehogPlugin.Module, trufflehogPlugin.PluginId, trufflehogPlugin)

	// pagemonitoring
	pagemonitoringPlugin := pagemonitoring.NewPlugin()
	pm.RegisterPlugin(pagemonitoringPlugin.Module, pagemonitoringPlugin.PluginId, pagemonitoringPlugin)

	// SentryDir
	dirPlugin := sentrydir.NewPlugin()
	pm.RegisterPlugin(dirPlugin.Module, dirPlugin.PluginId, dirPlugin)

	// nuclei
	nucleiPlugin := nuclei.NewPlugin()
	pm.RegisterPlugin(nucleiPlugin.Module, nucleiPlugin.PluginId, nucleiPlugin)
	customPlugins, err := GetCustomPlugin()
	if err != nil {
		log.Error(fmt.Sprintf("load custom plugin error: %v", err))
	}
	if len(customPlugins) != 0 {
		for _, plg := range customPlugins {
			pm.RegisterPlugin(plg.GetModule(), plg.GetPluginId(), plg)
		}
	}
	nodePlgInfokey := fmt.Sprintf("NodePlg:%v", global.AppConfig.NodeName)
	// 执行插件的安装和check
	// 0 代表未安装 1代表安装失败 2代表安装成功，未检查 3代表安装成功，检查失败 4代表安装检查都成功
	for module, plugins := range pm.plugins {
		for name, plugin := range plugins {
			plgInfo := map[string]interface{}{
				plugin.GetPluginId() + "_install": 0,
				plugin.GetPluginId() + "_check":   0,
			}
			// 调用每个插件的 Install 函数
			if err := plugin.Install(); err != nil {
				plgInfoErr := redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
				if plgInfoErr != nil {
					logger.SlogErrorLocal(fmt.Sprintf("send plginfo error 1: %s", plgInfoErr))
				}
				logger.SlogErrorLocal(fmt.Sprintf("failed to install plugin %s from module %s: %v", name, module, err))
				continue
			}
			plgInfo[plugin.GetPluginId()+"_install"] = 1
			// 调用每个插件的 Check 函数
			if err := plugin.Check(); err != nil {
				plgInfoErr := redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
				if plgInfoErr != nil {
					logger.SlogErrorLocal(fmt.Sprintf("send plginfo error 3: %s", plgInfoErr))
				}
				logger.SlogErrorLocal(fmt.Sprintf("failed to check plugin %s from module %s: %v", name, module, err))
				continue
			}
			plgInfo[plugin.GetPluginId()+"_check"] = 1
			plgInfoErr := redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
			if plgInfoErr != nil {
				logger.SlogErrorLocal(fmt.Sprintf("send plginfo error 4: %s", plgInfoErr))
			}
		}
	}
	return nil
}

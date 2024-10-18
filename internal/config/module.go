// config-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/7 19:48
// -------------------------------------------

package config

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
)

type SubdomainScanConfig struct {
	GoroutineCount int `yaml:"goroutineCount"` // 协程数量
}

type SubdomainSecurityConfig struct {
	GoroutineCount int `yaml:"goroutineCount"` // 协程数量
}

type AssetMappConfig struct {
	GoroutineCount int `yaml:"goroutineCount"` // 协程数量
}

type PortScanPreparationConfig struct {
	GoroutineCount int `yaml:"goroutineCount"` // 协程数量
}
type PortScanConfig struct {
	GoroutineCount int `yaml:"goroutineCount"` // 协程数量
}

type PortFingerprintConfig struct {
	GoroutineCount int `yaml:"goroutineCount"` // 协程数量
}

type AssetHandleConfig struct {
	GoroutineCount int `yaml:"goroutineCount"` // 协程数量
}

type URLScanConfig struct {
	GoroutineCount int `yaml:"goroutineCount"` // 协程数量
}

type URLSecurityConfig struct {
	GoroutineCount int `yaml:"goroutineCount"` // 协程数量
}

type WebCrawlerConfig struct {
	GoroutineCount int `yaml:"goroutineCount"` // 协程数量
}

type DirScanConfig struct {
	GoroutineCount int `yaml:"goroutineCount"` // 协程数量
}

type VulnerabilityScanConfig struct {
	GoroutineCount int `yaml:"goroutineCount"` // 协程数量
}

type ModulesConfigStruct struct {
	MaxGoroutineCount   int                       `yaml:"maxGoroutineCount"`
	SubdomainScan       SubdomainScanConfig       `yaml:"subdomainScan"`
	SubdomainSecurity   SubdomainSecurityConfig   `yaml:"subdomainSecurity"`
	AssetMapping        AssetMappConfig           `yaml:"assetMapping"`
	AssetHandle         AssetHandleConfig         `yaml:"assetHandle"`
	PortScanPreparation PortScanPreparationConfig `yaml:"portScanPreparation"`
	PortScan            PortScanConfig            `yaml:"portScan"`
	PortFingerprint     PortFingerprintConfig     `yaml:"portFingerprint"`
	URLScan             URLScanConfig             `yaml:"URLScan"`
	URLSecurity         URLSecurityConfig         `yaml:"URLSecurity"`
	WebCrawler          WebCrawlerConfig          `yaml:"webCrawler"`
	DirScan             DirScanConfig             `yaml:"dirScan"`
	VulnerabilityScan   VulnerabilityScanConfig   `yaml:"vulnerabilityScan"`
}

var ModulesConfigPath string
var ModulesConfig *ModulesConfigStruct

func ModulesInitialize() error {
	if err := utils.Tools.ReadYAMLFile(ModulesConfigPath, &ModulesConfig); err != nil {
		return err
	}
	return nil
}

func (cfg *ModulesConfigStruct) GetGoroutineCount(moduleName string) int {
	switch moduleName {
	case "task":
		return cfg.MaxGoroutineCount // 这个模块使用全局最大协程数
	case "TargetHandler":
		return cfg.MaxGoroutineCount // 这个模块使用全局最大协程数
	case "SubdomainScan":
		return cfg.SubdomainScan.GoroutineCount
	case "SubdomainSecurity":
		return cfg.SubdomainSecurity.GoroutineCount
	case "AssetMapping":
		return cfg.AssetMapping.GoroutineCount
	case "PortScanPreparation":
		return cfg.PortScanPreparation.GoroutineCount
	case "PortScan":
		return cfg.PortScan.GoroutineCount
	case "PortFingerprint":
		return cfg.PortFingerprint.GoroutineCount
	case "AssetHandle":
		return cfg.AssetHandle.GoroutineCount
	case "URLScan":
		return cfg.URLScan.GoroutineCount
	case "URLSecurity":
		return cfg.URLSecurity.GoroutineCount
	case "WebCrawler":
		return cfg.WebCrawler.GoroutineCount
	case "DirScan":
		return cfg.DirScan.GoroutineCount
	case "VulnerabilityScan":
		return cfg.VulnerabilityScan.GoroutineCount
	default:
		logger.SlogErrorLocal(fmt.Sprintf("Module %v thread limit not found", moduleName))
		return 0 // 默认值，表示没有找到匹配的模块
	}
}

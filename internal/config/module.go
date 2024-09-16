// config-------------------------------------
// @file      : module.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/7 19:48
// -------------------------------------------

package config

import (
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
	URLScan             URLScanConfig             `yaml:"URLScan"`
	URLSecurity         URLSecurityConfig         `yaml:"URLSecurity"`
	WebCrawler          WebCrawlerConfig          `yaml:"webCrawler"`
	VulnerabilityScan   VulnerabilityScanConfig   `yaml:"vulnerabilityScan"`
}

var ModulesConfigPath string
var ModulesConfig *ModulesConfigStruct

func ModulesInitialize() error {
	if err := utils.ReadYAMLFile(ModulesConfigPath, &ModulesConfig); err != nil {
		return err
	}
	return nil
}

func (cfg *ModulesConfigStruct) GetGoroutineCount(moduleName string) int {
	switch moduleName {
	case "task":
		return cfg.MaxGoroutineCount // 这个模块使用全局最大协程数
	case "targetHandler":
		return cfg.MaxGoroutineCount // 这个模块使用全局最大协程数
	case "subdomainScan":
		return cfg.SubdomainScan.GoroutineCount
	case "subdomainSecurity":
		return cfg.SubdomainSecurity.GoroutineCount
	case "assetMapping":
		return cfg.AssetMapping.GoroutineCount
	case "portScanPreparation":
		return cfg.PortScanPreparation.GoroutineCount
	case "portScan":
		return cfg.PortScan.GoroutineCount
	case "assetHandle":
		return cfg.AssetHandle.GoroutineCount
	case "URLScan":
		return cfg.URLScan.GoroutineCount
	case "URLSecurity":
		return cfg.URLSecurity.GoroutineCount
	case "webCrawler":
		return cfg.WebCrawler.GoroutineCount
	case "vulnerabilityScan":
		return cfg.VulnerabilityScan.GoroutineCount
	default:
		return 0 // 默认值，表示没有找到匹配的模块
	}
}

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
	MaxGoroutineCount int                     `yaml:"maxGoroutineCount"`
	SubdomainScan     SubdomainScanConfig     `yaml:"subdomainScan"`
	SubdomainSecurity SubdomainSecurityConfig `yaml:"subdomainSecurity"`
	AssetMapping      AssetMappConfig         `yaml:"assetMapping"`
	AssetHandle       AssetHandleConfig       `yaml:"assetHandle"`
	PortScan          PortScanConfig          `yaml:"portScan"`
	URLScan           URLScanConfig           `yaml:"URLScan"`
	URLSecurity       URLSecurityConfig       `yaml:"URLSecurity"`
	WebCrawler        WebCrawlerConfig        `yaml:"webCrawler"`
	VulnerabilityScan VulnerabilityScanConfig `yaml:"vulnerabilityScan"`
}

var ModulesConfigPath string
var ModulesConfig *ModulesConfigStruct

func ModulesInitialize() error {
	if err := utils.ReadYAMLFile(ModulesConfigPath, &ModulesConfig); err != nil {
		return err
	}
	return nil
}

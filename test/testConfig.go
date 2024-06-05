// Package main -----------------------------
// @file      : testConfig.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/11 20:15
// -------------------------------------------
package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Config struct {
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	ID      string `yaml:"id"`
	Enabled bool   `yaml:"enabled"`
}

func main() {
	scopeSentryConfigPath := filepath.Join("C:\\Users\\ThreatBook\\Desktop\\", "a.ymal")
	yamlFile, err := os.ReadFile(scopeSentryConfigPath)
	if err != nil {

	}

	// 创建一个Config对象，用于存储解析后的配置
	var config Config

	// 使用yaml库解析YAML内容到Config对象
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {

	}

	// 打印解析后的配置
	fmt.Printf("Config: %+v\n", config)

	// 遍历规则并打印每个规则的信息
	for _, rule := range config.Rules {
		fmt.Printf("Rule ID: %s, Enabled: %t\n", rule.ID, rule.Enabled)
	}
}

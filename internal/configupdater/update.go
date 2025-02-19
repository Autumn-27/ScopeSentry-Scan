// configupdater-------------------------------------
// @file      : update.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/11/2 15:38
// -------------------------------------------

package configupdater

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

func UpdateNode(content string) {
	parts := strings.SplitN(content, "[*]", 3)
	name := parts[0]
	state := parts[1]
	// 修改状态
	if state == "True" {
		global.AppConfig.State = 1
	} else {
		global.AppConfig.State = 2
	}
	// 修改名称
	if name != global.AppConfig.NodeName {
		// 旧的名称
		key := "node:" + global.AppConfig.NodeName
		// 新的名称
		global.AppConfig.NodeName = name
		err := redis.RedisClient.Del(context.Background(), key)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("delete redis node error: %v", err))
			return
		}
	}
	err := utils.Tools.WriteYAMLFile(global.ConfigPath, global.AppConfig)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("WriteYAMLFile global.AppConfig error: %v", err))
		return
	}
	UpdateModuleConfig(parts[2])

}

func UpdateSystemConfig(content string) {
	parts := strings.SplitN(content, "[*]", 2)
	timezone := parts[0]
	if timezone != global.AppConfig.TimeZoneName {
		global.AppConfig.TimeZoneName = timezone
	}
	err := utils.Tools.WriteYAMLFile(global.ConfigPath, global.AppConfig)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("WriteYAMLFile global.AppConfig error: %v", err))
		return
	}
	UpdateModuleConfig(parts[1])
}

func UpdateModuleConfig(content string) {
	modulesConfig := config.ModulesConfigStruct{}
	err := yaml.Unmarshal([]byte(content), &modulesConfig)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("modulesConfig parse error: %v", err))
		return
	}
	if config.ModulesConfig.MaxGoroutineCount != modulesConfig.MaxGoroutineCount {
		config.ModulesConfig.MaxGoroutineCount = modulesConfig.MaxGoroutineCount
		err = pool.PoolManage.SetGoroutineCount("task", modulesConfig.MaxGoroutineCount)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("MaxGoroutineCount SetGoroutineCount error: %v", err))
		}
	}
	moduleGoroutineCounts := map[string]int{
		"SubdomainScan":       modulesConfig.SubdomainScan.GoroutineCount,
		"SubdomainSecurity":   modulesConfig.SubdomainSecurity.GoroutineCount,
		"AssetMapping":        modulesConfig.AssetMapping.GoroutineCount,
		"AssetHandle":         modulesConfig.AssetHandle.GoroutineCount,
		"PortScanPreparation": modulesConfig.PortScanPreparation.GoroutineCount,
		"PortScan":            modulesConfig.PortScan.GoroutineCount,
		"PortFingerprint":     modulesConfig.PortFingerprint.GoroutineCount,
		"URLScan":             modulesConfig.URLScan.GoroutineCount,
		"URLSecurity":         modulesConfig.URLSecurity.GoroutineCount,
		"WebCrawler":          modulesConfig.WebCrawler.GoroutineCount,
		"DirScan":             modulesConfig.DirScan.GoroutineCount,
		"VulnerabilityScan":   modulesConfig.VulnerabilityScan.GoroutineCount,
	}
	// 修改模块的线程数量
	for moduleName, newCount := range moduleGoroutineCounts {
		switch moduleName {
		case "SubdomainScan":
			if config.ModulesConfig.SubdomainScan.GoroutineCount != newCount {
				config.ModulesConfig.SubdomainScan.GoroutineCount = newCount
				if err := pool.PoolManage.SetGoroutineCount(moduleName, newCount); err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("%s SetGoroutineCount error: %v", moduleName, err))
				}
			}
		case "SubdomainSecurity":
			if config.ModulesConfig.SubdomainSecurity.GoroutineCount != newCount {
				config.ModulesConfig.SubdomainSecurity.GoroutineCount = newCount
				if err := pool.PoolManage.SetGoroutineCount(moduleName, newCount); err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("%s SetGoroutineCount error: %v", moduleName, err))
				}
			}
		case "AssetMapping":
			if config.ModulesConfig.AssetMapping.GoroutineCount != newCount {
				config.ModulesConfig.AssetMapping.GoroutineCount = newCount
				if err := pool.PoolManage.SetGoroutineCount(moduleName, newCount); err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("%s SetGoroutineCount error: %v", moduleName, err))
				}
			}
		case "AssetHandle":
			if config.ModulesConfig.AssetHandle.GoroutineCount != newCount {
				config.ModulesConfig.AssetHandle.GoroutineCount = newCount
				if err := pool.PoolManage.SetGoroutineCount(moduleName, newCount); err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("%s SetGoroutineCount error: %v", moduleName, err))
				}
			}
		case "PortScanPreparation":
			if config.ModulesConfig.PortScanPreparation.GoroutineCount != newCount {
				config.ModulesConfig.PortScanPreparation.GoroutineCount = newCount
				if err := pool.PoolManage.SetGoroutineCount(moduleName, newCount); err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("%s SetGoroutineCount error: %v", moduleName, err))
				}
			}
		case "PortScan":
			if config.ModulesConfig.PortScan.GoroutineCount != newCount {
				config.ModulesConfig.PortScan.GoroutineCount = newCount
				if err := pool.PoolManage.SetGoroutineCount(moduleName, newCount); err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("%s SetGoroutineCount error: %v", moduleName, err))
				}
			}
		case "PortFingerprint":
			if config.ModulesConfig.PortFingerprint.GoroutineCount != newCount {
				config.ModulesConfig.PortFingerprint.GoroutineCount = newCount
				if err := pool.PoolManage.SetGoroutineCount(moduleName, newCount); err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("%s SetGoroutineCount error: %v", moduleName, err))
				}
			}
		case "URLScan":
			if config.ModulesConfig.URLScan.GoroutineCount != newCount {
				config.ModulesConfig.URLScan.GoroutineCount = newCount
				if err := pool.PoolManage.SetGoroutineCount(moduleName, newCount); err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("%s SetGoroutineCount error: %v", moduleName, err))
				}
			}
		case "URLSecurity":
			if config.ModulesConfig.URLSecurity.GoroutineCount != newCount {
				config.ModulesConfig.URLSecurity.GoroutineCount = newCount
				if err := pool.PoolManage.SetGoroutineCount(moduleName, newCount); err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("%s SetGoroutineCount error: %v", moduleName, err))
				}
			}
		case "WebCrawler":
			if config.ModulesConfig.WebCrawler.GoroutineCount != newCount {
				config.ModulesConfig.WebCrawler.GoroutineCount = newCount
				if err := pool.PoolManage.SetGoroutineCount(moduleName, newCount); err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("%s SetGoroutineCount error: %v", moduleName, err))
				}
			}
		case "DirScan":
			if config.ModulesConfig.DirScan.GoroutineCount != newCount {
				config.ModulesConfig.DirScan.GoroutineCount = newCount
				if err := pool.PoolManage.SetGoroutineCount(moduleName, newCount); err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("%s SetGoroutineCount error: %v", moduleName, err))
				}
			}
		case "VulnerabilityScan":
			if config.ModulesConfig.VulnerabilityScan.GoroutineCount != newCount {
				config.ModulesConfig.VulnerabilityScan.GoroutineCount = newCount
				if err := pool.PoolManage.SetGoroutineCount(moduleName, newCount); err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("%s SetGoroutineCount error: %v", moduleName, err))
				}
			}
		}
	}
	err = utils.Tools.WriteYAMLFile(config.ModulesConfigPath, config.ModulesConfig)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("ModulesConfig  writing file error: %v", err))
		return
	}

	modulesConfigRedis, err := utils.Tools.MarshalYAMLToString(config.ModulesConfig)
	if err != nil {

	}

	nodeInfo := map[string]interface{}{
		"state":         global.AppConfig.State, //1运行中 2暂停 3未连接
		"modulesConfig": modulesConfigRedis,
	}
	key := "node:" + global.AppConfig.NodeName
	err = redis.RedisClient.HMSet(context.Background(), key, nodeInfo)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("Error setting initial values: %s", err))
		return
	}
}

func Updatedictionary(content string) {
	parts := strings.SplitN(content, ":", 2)
	t := parts[0]
	id := parts[1]
	if t == "delete" {
		filePath := filepath.Join(global.DictPath, id)
		utils.Tools.DeleteFile(filePath)
	} else {
		UpdateDictionary(id)
	}
}

func UpdatePoc(content string) {
	parts := strings.SplitN(content, ":", 2)
	t := parts[0]
	if t == "delete" {
		for _, id := range strings.Split(parts[1], ",") {
			filePath := filepath.Join(global.PocDir, string(id)+".yaml")
			utils.Tools.DeleteFile(filePath)
		}
	} else {
		var ids []string
		for _, id := range strings.Split(parts[1], ",") {
			ids = append(ids, id)
		}
		LoadPoc(ids)
	}
}

func SystemUpdate(content string) {
	filePath := filepath.Join("/apps", "UPDATE")
	err := utils.Tools.WriteContentFile(filePath, content)
	if err != nil {
		logger.SlogError(fmt.Sprintf("Error write update file: %s", err))
		return
	}
	Exit()
}

func Exit() {
	logger.SlogInfo(fmt.Sprintf("system exit"))
	os.Exit(0)
}

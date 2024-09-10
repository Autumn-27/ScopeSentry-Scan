// node-------------------------------------
// @file      : node.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/7 18:51
// -------------------------------------------

package node

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handle"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"github.com/shirou/gopsutil/v3/mem"
	"os"
	"time"
)

func Register() {
	nodeName := config.AppConfig.NodeName
	if nodeName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			fmt.Println("Error:", err)
			hostname = utils.GenerateRandomString(6)
		}
		nodeName = hostname + "-" + utils.GenerateRandomString(6)
	}
	key := "node:" + nodeName
	if nodeName != config.AppConfig.NodeName {
		config.AppConfig.NodeName = nodeName
		err := utils.WriteYAMLFile(config.ConfigPath, config.AppConfig)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("Register node error: %v", err))
			return
		}
	}
	firstRegister := true
	ticker := time.Tick(20 * time.Second)
	for {
		if firstRegister {
			memInfo, _ := mem.VirtualMemory()
			nodeInfo := map[string]interface{}{
				"updateTime":    config.GetTimeNow(),
				"running":       0,
				"finished":      0,
				"cpuNum":        0,
				"TotleMem":      float64(memInfo.Total) / 1024 / 1024,
				"memNum":        0,
				"state":         1, //1运行中 2暂停 3未连接
				"version":       config.VERSION,
				"modulesConfig": utils.MarshalYAMLToString(config.ModulesConfig),
			}
			err := redis.RedisClient.HMSet(context.Background(), key, nodeInfo)
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("Error setting initial values: %s", err))
				return
			}
			logger.SlogInfo(fmt.Sprintf("Register Success:%v - version %v", nodeName, system.VERSION))
			firstRegister = false
		} else {
			key = "node:" + config.AppConfig.NodeName
			cpuNum, memNum := utils.GetSystemUsage()
			run, fin := handle.TaskHandle.GetRunFin()
			nodeInfo := map[string]interface{}{
				"updateTime": config.GetTimeNow(),
				"cpuNum":     cpuNum,
				"memNum":     memNum,
				"maxTaskNum": config.ModulesConfig.MaxGoroutineCount,
				"running":    run,
				"finished":   fin,
				"state":      1,
				"version":    config.VERSION,
			}
			redisPingErr := redis.RedisClient.Ping(context.Background())
			if redisPingErr != nil {
				redis.Initialize()
			}
			err := redis.RedisClient.HMSet(context.Background(), key, nodeInfo)
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("Error setting initial values: %s", err))
				continue
			}
		}
		<-ticker
	}
}

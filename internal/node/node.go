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
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handler"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"github.com/shirou/gopsutil/v3/mem"
	"time"
)

func Register() {
	nodeName := global.AppConfig.NodeName
	key := "node:" + nodeName
	firstRegister := true
	ticker := time.Tick(20 * time.Second)
	for {
		if firstRegister {
			memInfo, _ := mem.VirtualMemory()
			modulesConfig, err := utils.Tools.MarshalYAMLToString(config.ModulesConfig)
			if err != nil {

			}
			nodeInfo := map[string]interface{}{
				"updateTime":    utils.Tools.GetTimeNow(),
				"running":       0,
				"finished":      0,
				"cpuNum":        0,
				"TotleMem":      float64(memInfo.Total) / 1024 / 1024,
				"memNum":        0,
				"state":         1, //1运行中 2暂停 3未连接
				"version":       global.VERSION,
				"modulesConfig": modulesConfig,
			}
			err = redis.RedisClient.HMSet(context.Background(), key, nodeInfo)
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("Error setting initial values: %s", err))
				return
			}
			logger.SlogInfo(fmt.Sprintf("Register Success:%v - version %v", nodeName, global.VERSION))
			firstRegister = false
		} else {
			key = "node:" + global.AppConfig.NodeName
			cpuNum, memNum := utils.Tools.GetSystemUsage()
			run, fin := handler.TaskHandle.GetRunFin()
			nodeInfo := map[string]interface{}{
				"updateTime": utils.Tools.GetTimeNow(),
				"cpuNum":     cpuNum,
				"memNum":     memNum,
				"maxTaskNum": config.ModulesConfig.MaxGoroutineCount,
				"running":    run,
				"finished":   fin,
				"state":      global.AppConfig.State,
				"version":    global.VERSION,
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

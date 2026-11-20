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
	"math"
	"os"
	"strconv"
	"strings"
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
				"cmdFail":       utils.Tools.EventFailureCount(),
				"cmdFail10m":    utils.Tools.EventFailureRecent(10 * time.Minute),
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
			// 心跳里的 state 不能直接写配置值：配置未设置时为 0，
			// 会把「在线」覆盖成 0，服务端据此判定节点离线（界面 0/2 在线）。
			// 1=运行中 2=暂停 3=未连接，0 视为未配置 → 默认在线。
			hbState := global.AppConfig.State
			if hbState == 0 {
				hbState = 1
			}
			nodeInfo := map[string]interface{}{
				"updateTime":   utils.Tools.GetTimeNow(),
				"cpuNum":       cpuNum,
				"memNum":       memNum,
				"containerMem": ContainerMemoryPercent(),
				"maxTaskNum":   ObserveMemoryPressure(),
				"running":      run,
				"finished":     fin,
				"state":        hbState,
				"version":      global.VERSION,
				// 外部工具失败累计（ksubdomain/rad/katana 等非零退出），界面可见
				"cmdFail":    utils.Tools.EventFailureCount(),
				"cmdFail10m": utils.Tools.EventFailureRecent(10 * time.Minute),
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

// ContainerMemoryPercent 读取容器自身的 cgroup 内存用量百分比。
// 心跳里的 memNum 是**宿主机**内存占用，容器的 limit 和它无关：
// 实测宿主机 22% 时容器已 97%（2026-11-20 节点被压死却看不出来）。
// 拿不到（非容器环境/cgroup 缺字段）时返回 -1，调用方当「未知」处理。
func ContainerMemoryPercent() float64 {
	// cgroup v2
	if cur, err := os.ReadFile("/sys/fs/cgroup/memory.current"); err == nil {
		if lim, lerr := os.ReadFile("/sys/fs/cgroup/memory.max"); lerr == nil {
			usage := strings.TrimSpace(string(cur))
			limit := strings.TrimSpace(string(lim))
			if limit != "" && limit != "max" {
				u, e1 := strconv.ParseFloat(usage, 64)
				l, e2 := strconv.ParseFloat(limit, 64)
				if e1 == nil && e2 == nil && l > 0 {
					return math.Round(u/l*1000) / 10
				}
			}
		}
	}
	// cgroup v1
	if cur, err := os.ReadFile("/sys/fs/cgroup/memory/memory.usage_in_bytes"); err == nil {
		if lim, lerr := os.ReadFile("/sys/fs/cgroup/memory/memory.limit_in_bytes"); lerr == nil {
			u, e1 := strconv.ParseFloat(strings.TrimSpace(string(cur)), 64)
			l, e2 := strconv.ParseFloat(strings.TrimSpace(string(lim)), 64)
			if e1 == nil && e2 == nil && l > 0 && l < 1e18 {
				return math.Round(u/l*1000) / 10
			}
		}
	}
	return -1
}

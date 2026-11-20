// pool-------------------------------------
// @file      : pool.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/7 21:17
// -------------------------------------------

package pool

import (
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/panjf2000/ants/v2"
	"sort"
	"sync"
	"time"
)

type Manager struct {
	pools map[string]*ants.Pool
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

var PoolManage *Manager

func Initialize() {
	PoolManage = &Manager{
		pools: make(map[string]*ants.Pool),
		locks: make(map[string]*sync.Mutex),
	}
}

func (pm *Manager) InitializeModulesPools(cfg *config.ModulesConfigStruct) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Initialize pools for each module
	var err error
	modules := []string{
		"task", "TargetHandler", "SubdomainScan", "SubdomainSecurity",
		"AssetMapping", "AssetHandle", "PortScan", "PortScanPreparation", "PortFingerprint", "URLScan",
		"URLSecurity", "WebCrawler", "DirScan", "VulnerabilityScan",
	}

	for _, moduleName := range modules {
		count := 0
		if cfg != nil {
			count = cfg.GetGoroutineCount(moduleName)
		}
		// ants.NewPool 对 size<=0 会 panic（不是返回 error），而配置被清空时
		// GetGoroutineCount 正好返回 0 —— 2026-11-20 三台节点就是这么连环崩溃的。
		// 这里先兜底，保证「配置丢了」也只是一次告警，不是集群停摆。
		if count <= 0 {
			logger.SlogWarnLocal(fmt.Sprintf("pool %s goroutine count is %d (config missing?), fallback to 2", moduleName, count))
			count = 2
		}
		pm.pools[moduleName], err = ants.NewPool(count)
		if err != nil {
			// 单个模块的协程池建不起来不该让整个节点退出：
			// 用保守默认值兜底，并在日志里留痕（2026-11-20）。
			logger.SlogWarnLocal(fmt.Sprintf("create pool for %s failed, fallback to 10: %v", moduleName, err))
			pm.pools[moduleName], _ = ants.NewPool(10)
		}
		pm.locks[moduleName] = &sync.Mutex{}
	}
}

func (pm *Manager) SetGoroutineCount(moduleName string, count int) error {
	logger.SlogInfoLocal(fmt.Sprintf("%v SetGoroutineCount: %v", moduleName, count))
	pm.mu.Lock()
	lock, exists := pm.locks[moduleName]
	if !exists {
		pm.mu.Unlock()
		return errors.New("module not found")
	}
	pm.mu.Unlock()

	lock.Lock()
	defer lock.Unlock()

	pool, exists := pm.pools[moduleName]
	if !exists {
		return errors.New("module not found")
	}

	// Adjust the pool size
	pool.Tune(count)

	return nil
}

func (pm *Manager) SubmitTask(moduleName string, task func()) error {
	pm.mu.Lock()
	lock, exists := pm.locks[moduleName]
	if !exists {
		pm.mu.Unlock()
		return errors.New("module not found")
	}
	pm.mu.Unlock()

	lock.Lock()
	defer lock.Unlock()

	pool, exists := pm.pools[moduleName]
	if !exists {
		return errors.New("module not found")
	}

	return pool.Submit(task)
}

func (pm *Manager) GetModulePoolSize(module string) (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, exists := pm.pools[module]
	if !exists {
		return 0, errors.New("module not found")
	}

	// 获取池的大小
	return pool.Running() + pool.Free(), nil
}

func (pm *Manager) CustomSubmitTask(name string, size int, task func()) error {
	var err error
	pm.mu.Lock()
	lock, exists := pm.locks[name]
	if !exists {
		if size <= 0 {
			size = 10
		}
		pm.pools[name], err = ants.NewPool(size)
		if err != nil {
			logger.SlogWarnLocal(fmt.Sprintf("create pool for %s failed, fallback to 10: %v", name, err))
			pm.pools[name], _ = ants.NewPool(10)
		}
		pm.locks[name] = &sync.Mutex{}
	}
	pm.mu.Unlock()

	lock.Lock()
	defer lock.Unlock()

	pool, exists := pm.pools[name]
	if !exists {
		return errors.New("module not found")
	}

	return pool.Submit(task)
}

func (pm *Manager) PrintRunningGoroutines(sortedModuleNames []string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if global.AppConfig.Debug {
		var result string
		result = "Module Goroutines:"
		for _, moduleName := range sortedModuleNames {
			running := pm.pools[moduleName].Running()
			result += fmt.Sprintf("[%s]- run: %d, ", moduleName, running)
		}
		// 去掉最后一个多余的逗号和空格
		if len(result) > 0 {
			result = result[:len(result)-2]
			logger.SlogDebugLocal(result)
		}
	}
}

func (pm *Manager) GetModuleRunningGoroutines(module string) int {
	running := pm.pools[module].Running()
	return running
}

func StartMonitoring() {
	ticker := time.NewTicker(10 * time.Second) // 每隔10秒打印一次
	defer ticker.Stop()
	var moduleNames []string
	for moduleName := range PoolManage.pools {
		moduleNames = append(moduleNames, moduleName)
	}
	sort.Strings(moduleNames)

	for {
		select {
		case <-ticker.C:
			PoolManage.PrintRunningGoroutines(moduleNames)
		}
	}
}

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
	"log"
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
		pm.pools[moduleName], err = ants.NewPool(cfg.GetGoroutineCount(moduleName))
		if err != nil {
			log.Fatalf("Failed to create pool for %s: %v", moduleName, err)
		}
		pm.locks[moduleName] = &sync.Mutex{}
	}
}

func (pm *Manager) SetGoroutineCount(moduleName string, count int) error {
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

func (pm *Manager) PrintRunningGoroutines(sortedModuleNames []string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if global.AppConfig.Debug {
		var result string
		result = "Module Goroutines:"
		for _, moduleName := range sortedModuleNames {
			running := pm.pools[moduleName].Running()
			result += fmt.Sprintf("%s: %d, ", moduleName, running)
		}
		// 去掉最后一个多余的逗号和空格
		if len(result) > 0 {
			result = result[:len(result)-2]
			logger.SlogDebugLocal(result)
		}
	}
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

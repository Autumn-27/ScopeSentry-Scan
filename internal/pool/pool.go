// pool-------------------------------------
// @file      : pool.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/7 21:17
// -------------------------------------------

package pool

import (
	"errors"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/panjf2000/ants/v2"
	"log"
	"sync"
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
		"task", "targetHandler", "subdomainScan", "subdomainSecurity",
		"assetMapping", "portScan", "assetResultHandl", "URLScan",
		"URLSecurity", "webCrawler", "vulnerabilityScan",
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

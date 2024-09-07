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
}

var PoolManage *Manager

func Initialize() {
	PoolManage = &Manager{
		pools: make(map[string]*ants.Pool),
	}
}

func (pm *Manager) InitializeModulesPools(cfg *config.ModulesConfigStruct) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Initialize pools for each module
	var err error
	pm.pools["subdomainScan"], err = ants.NewPool(cfg.SubdomainScan.GoroutineCount)
	if err != nil {
		log.Fatalf("Failed to create pool for subdomainScan: %v", err)
	}
	pm.pools["subdomainResultHandl"], err = ants.NewPool(cfg.SubdomainResultHandl.GoroutineCount)
	if err != nil {
		log.Fatalf("Failed to create pool for subdomainResultHandl: %v", err)
	}
	pm.pools["assetMapping"], err = ants.NewPool(cfg.AssetMapping.GoroutineCount)
	if err != nil {
		log.Fatalf("Failed to create pool for assetMapping: %v", err)
	}
	pm.pools["portScan"], err = ants.NewPool(cfg.PortScan.GoroutineCount)
	if err != nil {
		log.Fatalf("Failed to create pool for portScan: %v", err)
	}
	pm.pools["assetResultHandl"], err = ants.NewPool(cfg.AssetResultHandl.GoroutineCount)
	if err != nil {
		log.Fatalf("Failed to create pool for assetResultHandl: %v", err)
	}
	pm.pools["URLScan"], err = ants.NewPool(cfg.URLScan.GoroutineCount)
	if err != nil {
		log.Fatalf("Failed to create pool for URLScan: %v", err)
	}
	pm.pools["URLScanResultHandl"], err = ants.NewPool(cfg.URLScanResultHandl.GoroutineCount)
	if err != nil {
		log.Fatalf("Failed to create pool for URLScanResultHandl: %v", err)
	}
	pm.pools["webCrawler"], err = ants.NewPool(cfg.WebCrawler.GoroutineCount)
	if err != nil {
		log.Fatalf("Failed to create pool for webCrawler: %v", err)
	}
	pm.pools["vulnerabilityScan"], err = ants.NewPool(cfg.VulnerabilityScan.GoroutineCount)
	if err != nil {
		log.Fatalf("Failed to create pool for vulnerabilityScan: %v", err)
	}
}

func (pm *Manager) SetGoroutineCount(moduleName string, count int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

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
	defer pm.mu.Unlock()

	pool, exists := pm.pools[moduleName]
	if !exists {
		return errors.New("module not found")
	}

	return pool.Submit(task)
}

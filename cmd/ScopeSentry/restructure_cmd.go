// main-------------------------------------
// @file      : restructure_cmd.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 22:00
// -------------------------------------------

package main

import (
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/configupdater"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pebbledb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/task"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"log"
	"path/filepath"
	"time"
)

func main() {
	// 初始化系统信息
	config.Initialize()
	config.VERSION = "1.5"
	var err error
	// 初始化mongodb连接
	mongodb.Initialize()
	// 初始化redis连接
	redis.Initialize()
	// 初始化日志模块
	err = logger.NewLogger()
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	// 初始化任务计数器
	task.InitHandle()
	// 更新配置、加载字典
	configupdater.Initialize()
	// 初始化模块配置
	err = config.ModulesInitialize()
	if err != nil {
		log.Fatalf("Failed to init ModulesConfig: %v", err)
		return
	}
	// 初始化协程池
	pool.Initialize()
	// 初始化个模块的协程池
	pool.PoolManage.InitializeModulesPools(config.ModulesConfig)

	// 初始化本地持久化缓存
	pebbledbSetting := pebbledb.Settings{
		DBPath:       filepath.Join(config.AbsolutePath, "data", "pebbledb"),
		CacheSize:    64 << 20,
		MaxOpenFiles: 500,
	}
	pebbledbOption := pebbledb.GetPebbleOptions(&pebbledbSetting)
	pedb, err := pebbledb.NewPebbleDB(pebbledbOption, pebbledbSetting.DBPath)
	if err != nil {
		return
	}
	pebbledb.PebbleStore = pedb
	defer func(PebbleStore *pebbledb.PebbleDB) {
		_ = PebbleStore.Close()
	}(pebbledb.PebbleStore)

	taskE := task.Options{
		ID:                    "12321321321",
		TaskName:              "test",
		SubdomainScan:         []string{"subfinder"},
		SubdomainResultHandle: []string{"takeover"},
		AssetMapping:          []string{"httpx"},
		AssetResultHandle:     []string{""},
		PortScan:              []string{"rustscan"},
		URLScan:               []string{"test"},
		URLScanResultHandle:   []string{"test"},
		WebCrawler:            []string{"test"},
		VulnerabilityScan:     []string{"nuclei"},
	}
	jsonStr, err := utils.StructToJSON(taskE)
	if err != nil {
		return
	}
	pebbledb.PebbleStore.Put([]byte("task:12321321321"), []byte(jsonStr))
	pebbledb.PebbleStore.Put([]byte("12321321321:baidu.com"), []byte("1"))
	pebbledb.PebbleStore.Put([]byte("12321321321:google.com"), []byte("1"))
	pebbledb.PebbleStore.Put([]byte("12321321321:tes1t.com"), []byte("1"))
	pebbledb.PebbleStore.Put([]byte("12321321321:tes2t.com"), []byte("1"))
	pebbledb.PebbleStore.Put([]byte("12321321321:tes3t.com"), []byte("1"))
	pebbledb.PebbleStore.Put([]byte("12321321321:tes4t.com"), []byte("1"))
	pebbledb.PebbleStore.Put([]byte("12321321321:tes5t.com"), []byte("1"))
	pebbledb.PebbleStore.Put([]byte("12321321321:tes6t.com"), []byte("1"))
	pebbledb.PebbleStore.Put([]byte("12321321321:tes7t.com"), []byte("1"))

	task.GetTask()
	time.Sleep(10 * time.Second)

}

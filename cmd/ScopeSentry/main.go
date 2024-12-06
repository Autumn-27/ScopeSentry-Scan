// Package ScopeSentry -----------------------------
// @file      : main.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/6 17:24
// -------------------------------------------
package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/bigcache"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/configupdater"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handler"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/node"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/notification"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pebbledb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/task"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func main() {
	Banner()
	// 初始化系统信息
	config.Initialize()
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
	handler.InitHandle()
	// 更新配置、加载字典
	configupdater.Initialize()
	// 初始化模块配置
	err = config.ModulesInitialize()
	if err != nil {
		log.Fatalf("Failed to init ModulesConfig: %v", err)
		return
	}
	// 初始化上下文管理器
	contextmanager.NewContextManager()
	// 初始化tools
	utils.InitializeTools()
	utils.InitializeDnsTools()
	utils.InitializeRequests()
	utils.InitializeResults()
	// 初始化通知模块
	notification.InitializeNotification()
	// 初始化协程池
	pool.Initialize()
	// 初始化个模块的协程池
	pool.PoolManage.InitializeModulesPools(config.ModulesConfig)
	go pool.StartMonitoring()
	// 初始化内存缓存
	err = bigcache.Initialize()
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("bigcache Initialize error: %v", err))
		return
	}
	// 初始化本地持久化缓存
	pebbledbSetting := pebbledb.Settings{
		DBPath:       filepath.Join(global.AbsolutePath, "data", "pebbledb"),
		CacheSize:    64 << 20,
		MaxOpenFiles: 500,
	}
	pebbledbOption := pebbledb.GetPebbleOptions(&pebbledbSetting)
	if !global.AppConfig.Debug {
		pebbledbOption.Logger = nil
	}
	pedb, err := pebbledb.NewPebbleDB(pebbledbOption, pebbledbSetting.DBPath)
	if err != nil {
		return
	}
	pebbledb.PebbleStore = pedb
	defer func(PebbleStore *pebbledb.PebbleDB) {
		_ = PebbleStore.Close()
	}(pebbledb.PebbleStore)

	// 初始化结果处理队列，(正常的时候将该初始化放入任务开始时，任务执行完毕关闭结果队列)
	results.InitializeResultQueue()
	defer results.Close()

	// 初始化全局插件管理器
	plugins.GlobalPluginManager = plugins.NewPluginManager()
	err = plugins.GlobalPluginManager.InitializePlugins()
	if err != nil {
		log.Fatalf("Failed to init plugins: %v", err)
		return
	}
	// 性能监控
	go pprof()
	//go printMemStats(5 * time.Second)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done() // 减少计数器，表示任务完成
		for {
			task.GetTask()
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			node.Register()
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			configupdater.RefreshConfig()
		}
	}()
	time.Sleep(10 * time.Second)
	wg.Wait()
}

func Banner() {
	banner := "   _____                         _____            _              \n  / ____|                       / ____|          | |             \n | (___   ___ ___  _ __   ___  | (___   ___ _ __ | |_ _ __ _   _ \n  \\___ \\ / __/ _ \\| '_ \\ / _ \\  \\___ \\ / _ \\ '_ \\| __| '__| | | |\n  ____) | (_| (_) | |_) |  __/  ____) |  __/ | | | |_| |  | |_| |\n |_____/ \\___\\___/| .__/ \\___| |_____/ \\___|_| |_|\\__|_|   \\__, |\n                  | |                                       __/ |\n                  |_|                                      |___/ "
	fmt.Println(banner)
}

func pprof() {
	if global.AppConfig.Debug {
		go func() {
			_ = http.ListenAndServe("0.0.0.0:6060", nil)
		}()
		//go DebugMem()
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigs
			fmt.Println("收到终止信号:", sig)
		}()
	}
}

func printMemStats(interval time.Duration) {
	if global.AppConfig.Debug {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			fmt.Printf("\n==== Memory Stats ====\n")
			fmt.Printf("Allocated:\t%.2f MB\n", float64(memStats.Alloc)/1024/1024)
			fmt.Printf("Total Alloc:\t%.2f MB\n", float64(memStats.TotalAlloc)/1024/1024)
			fmt.Printf("Sys:\t\t%.2f MB\n", float64(memStats.Sys)/1024/1024)
			fmt.Printf("Heap Alloc:\t%.2f MB\n", float64(memStats.HeapAlloc)/1024/1024)
			fmt.Printf("Heap Sys:\t%.2f MB\n", float64(memStats.HeapSys)/1024/1024)
			fmt.Printf("Heap Idle:\t%.2f MB\n", float64(memStats.HeapIdle)/1024/1024)
			fmt.Printf("Heap Inuse:\t%.2f MB\n", float64(memStats.HeapInuse)/1024/1024)
			fmt.Printf("Heap Released:\t%.2f MB\n", float64(memStats.HeapReleased)/1024/1024)
			fmt.Printf("Stack Sys:\t%.2f MB\n", float64(memStats.StackSys)/1024/1024)
			fmt.Printf("======================\n")
		}
	}
}

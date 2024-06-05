// Package ScopeSentry -----------------------------
// @file      : main.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/6 17:24
// -------------------------------------------
package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/crawlerMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/node"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/task"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"github.com/shirou/gopsutil/v3/mem"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func printStackTrace() {
	// 获取堆栈信息
	buf := make([]byte, 1<<16)
	stackSize := runtime.Stack(buf, true)
	fmt.Printf("堆栈信息:\n%s\n", buf[:stackSize])
}

func main() {
	defer system.RecoverPanic("main")
	Banner()
	rand.Seed(time.Now().UnixNano())
	flag := system.SetUp()
	if !flag {
		myLog := system.CustomLog{
			Status: "Error",
			Msg:    fmt.Sprintf("SetUp Config Error"),
		}
		system.PrintLog(myLog)
		os.Exit(1)
	}
	if system.AppConfig.System.Debug {
		go func() {
			_ = http.ListenAndServe("0.0.0.0:6060", nil)
		}()
		//go DebugMem()
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigs
			fmt.Println("收到终止信号:", sig)
			printStackTrace()
		}()
	}
	util.InitHttpClient()
	var wg sync.WaitGroup

	// 初始化爬虫
	go crawlerMode.CrawlerThread(system.CrawlerThreadUpdateFlag)
	system.CrawlerThreadUpdateFlag <- true

	// node 注册、存活更新
	go func() {
		defer wg.Done() // 减少计数器，表示任务完成
		node.Register()
	}()

	// 配置更新、暂停扫描
	wg.Add(1) // 增加计数器，表示有一个任务需要等待
	// node 注册、存活更新
	go func() {
		defer wg.Done() // 减少计数器，表示任务完成
		system.RefreshConfig()
	}()

	wg.Add(1) // 增加计数器，表示有一个任务需要等待
	// node 注册、存活更新
	go func() {
		defer wg.Done() // 减少计数器，表示任务完成
		task.GetTask()
	}()
	go workerUpdateSystem(system.UpdateSystemFlag)
	time.Sleep(time.Second * 5)
	wg.Wait()
}

func DebugMem() {
	var m runtime.MemStats
	ticker := time.Tick(2 * time.Second)
	for {
		<-ticker
		memInfo, err := mem.VirtualMemory()
		system.SlogDebugLocal(fmt.Sprintf("Total Memory: %.2f MiB Used Memory: %.2f MiB", float64(memInfo.Total)/1024/1024, float64(memInfo.Used)/1024/1024))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		runtime.ReadMemStats(&m)
		system.SlogDebugLocal(fmt.Sprintf("Alloc = %v MiB\tTotalAlloc = %v MiB\tSys = %v MiB\tNumGC = %v", bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC))
	}
}

func bToMb(b uint64) any {
	return b / 1024 / 1024
}

func Banner() {
	banner := "   _____                         _____            _              \n  / ____|                       / ____|          | |             \n | (___   ___ ___  _ __   ___  | (___   ___ _ __ | |_ _ __ _   _ \n  \\___ \\ / __/ _ \\| '_ \\ / _ \\  \\___ \\ / _ \\ '_ \\| __| '__| | | |\n  ____) | (_| (_) | |_) |  __/  ____) |  __/ | | | |_| |  | |_| |\n |_____/ \\___\\___/| .__/ \\___| |_____/ \\___|_| |_|\\__|_|   \\__, |\n                  | |                                       __/ |\n                  |_|                                      |___/ "
	fmt.Println(banner)
}

func workerUpdateSystem(done chan bool) {
	//time.Sleep(2 * time.Second)
	//<-done
	//overseer.Run(overseer.Config{
	//	TerminateTimeout: 60 * time.Second,
	//	Fetcher: &fetcher.HTTP{
	//		URL:      fmt.Sprintf("%v/get/scopesentry/client?system=%s&arch=%s", system.UpdateUrl, runtime.GOOS, runtime.GOARCH),
	//		Interval: 1 * time.Second,
	//	},
	//})
}

//func preUpgrade(tempBinaryPath string) error {
//	fmt.Printf("download binary path: %s\n", tempBinaryPath)
//	return nil
//}

//func UpdateSysmeMain(state overseer.State) {
//	flag := system.SetUp()
//	if !flag {
//		myLog := system.CustomLog{
//			Status: "Error",
//			Msg:    fmt.Sprintf("SetUp Config Error"),
//		}
//		system.PrintLog(myLog)
//		os.Exit(1)
//	}
//	util.InitHttpClient()
//	var wg sync.WaitGroup
//	wg.Add(1) // 增加计数器，表示有一个任务需要等待
//	// node 注册、存活更新
//	go func() {
//		defer wg.Done() // 减少计数器，表示任务完成
//		node.Register()
//	}()
//
//	// 配置更新、暂停扫描
//	wg.Add(1) // 增加计数器，表示有一个任务需要等待
//	// node 注册、存活更新
//	go func() {
//		defer wg.Done() // 减少计数器，表示任务完成
//		system.RefreshConfig()
//	}()
//
//	wg.Add(1) // 增加计数器，表示有一个任务需要等待
//	// node 注册、存活更新
//	go func() {
//		defer wg.Done() // 减少计数器，表示任务完成
//		task.GetTask()
//	}()
//	time.Sleep(time.Second * 5)
//	wg.Wait()
//}

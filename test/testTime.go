// Package main -----------------------------
// @file      : testTime.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/2/24 21:08
// -------------------------------------------
package main

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"time"

	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

func main() {
	//fmt.Println(util.GetTimeNow())
	avgLoad, err := load.Avg()
	if err != nil {
		fmt.Println("Failed to get load average:", err)
		return
	}
	fmt.Printf("Load Average: %.2f, %.2f, %.2f\n", avgLoad.Load1, avgLoad.Load5, avgLoad.Load15)

	// 获取CPU使用率
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		fmt.Println("Failed to get CPU usage:", err)
		return
	}
	if len(percent) > 0 {
		fmt.Printf("CPU Usage: %.2f%%\n", percent[0])
	}

	// 获取内存使用率
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Failed to get memory usage:", err)
		return
	}
	fmt.Printf("Memory Usage: %.2f%%\n", memInfo.UsedPercent)

	// 获取主机信息
	hostInfo, err := host.Info()
	if err != nil {
		fmt.Println("Failed to get host info:", err)
		return
	}
	fmt.Println("OS:", hostInfo.OS)
	fmt.Println("Platform:", hostInfo.Platform)
	fmt.Println("Hostname:", hostInfo.Hostname)
	fmt.Println("Uptime:", hostInfo.Uptime)
}

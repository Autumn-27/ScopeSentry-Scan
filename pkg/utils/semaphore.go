// utils-------------------------------------
// @file      : semaphore.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/4/12 21:33
// -------------------------------------------

package utils

import (
	"golang.org/x/sync/semaphore"
	"sync"
)

var SemaphoreDict = make(map[string]*semaphore.Weighted)
var Mutex sync.Mutex

func GetSemaphore(tp string, limit int64) *semaphore.Weighted {
	Mutex.Lock()
	defer Mutex.Unlock()

	// 如果字典中已有该 tp 的信号量，直接返回
	if sem, exists := SemaphoreDict[tp]; exists {
		return sem
	}

	// 否则创建一个新的信号量
	sem := semaphore.NewWeighted(limit)

	// 将新创建的信号量加入字典
	SemaphoreDict[tp] = sem
	return sem
}

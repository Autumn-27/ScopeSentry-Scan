// task-------------------------------------
// @file      : handle.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/7 19:17
// -------------------------------------------

package task

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"sync"
)

var AppRFMutex sync.Mutex

type Handle struct {
	RunningNum int
	FinNum     int
	mu         sync.Mutex
}

var TaskHandle *Handle

func InitHandle() {
	TaskHandle = &Handle{
		RunningNum: 0,
		FinNum:     0,
	}
}

func (h *Handle) StartTask() {
	h.mu.Lock()         // 锁定互斥锁
	defer h.mu.Unlock() // 确保在函数结束时解锁
	h.RunningNum = h.RunningNum + 1
	logger.SlogDebugLocal(fmt.Sprintf("Running start value: %d", h.RunningNum))
}

func (h *Handle) EndTask() {
	h.mu.Lock()         // 锁定互斥锁
	defer h.mu.Unlock() // 确保在函数结束时解锁
	h.RunningNum = h.RunningNum - 1
	logger.SlogDebugLocal(fmt.Sprintf("Running start value: %d", h.RunningNum))
	h.FinNum = h.FinNum + 1
	logger.SlogDebugLocal(fmt.Sprintf("Running end value: %d", h.FinNum))
}

func (h *Handle) GetRunFin() (int, int) {
	h.mu.Lock()         // 锁定互斥锁
	defer h.mu.Unlock() // 确保在函数结束时解锁
	return h.RunningNum, h.FinNum
}

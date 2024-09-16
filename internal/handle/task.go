// task-------------------------------------
// @file      : handle.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/7 19:17
// -------------------------------------------

package handle

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"sync"
)

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

func (h *Handle) ProgressStart(typ string, target string, taskId string, flag int) {
	if flag == 0 {
		return
	}
	logger.SlogInfoLocal(fmt.Sprintf("%v module start scanning the target: %v", typ, target))
	key := "TaskInfo:progress:" + taskId + ":" + target
	ty := typ + "_start"
	ProgressInfo := map[string]interface{}{
		ty: config.GetTimeNow(),
	}
	err := redis.RedisClient.HMSet(context.Background(), key, ProgressInfo)
	if err != nil {
		logger.SlogError(fmt.Sprintf("ProgressStart redis error: %s", err))
		return
	}
}

func (h *Handle) ProgressEnd(typ string, target string, taskId string, flag int) {
	if flag == 0 {
		return
	}
	logger.SlogInfoLocal(fmt.Sprintf("%v module end scanning the target: %v", typ, target))
	key := "TaskInfo:progress:" + taskId + ":" + target
	ty := typ + "_end"
	ProgressInfo := map[string]interface{}{
		ty: config.GetTimeNow(),
	}
	err := redis.RedisClient.HMSet(context.Background(), key, ProgressInfo)
	if err != nil {
		logger.SlogError(fmt.Sprintf("ProgressEnd redis error: %s", err))
		return
	}
}

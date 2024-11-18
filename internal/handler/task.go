// task-------------------------------------
// @file      : handle.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/7 19:17
// -------------------------------------------

package handler

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pebbledb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"strings"
	"sync"
	"time"
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
		ty: utils.Tools.GetTimeNow(),
	}
	err := redis.RedisClient.HMSet(context.Background(), key, ProgressInfo)
	if err != nil {
		logger.SlogError(fmt.Sprintf("ProgressStart redis error: %s", err))
		return
	}
}

func (h *Handle) ProgressEnd(typ string, target string, taskId string, flag int, time time.Duration) {
	if flag == 0 {
		return
	}
	logger.SlogInfoLocal(fmt.Sprintf("%v module end scanning the target: %v running time: %v", typ, target, time))
	key := "TaskInfo:progress:" + taskId + ":" + target
	ty := typ + "_end"
	ProgressInfo := map[string]interface{}{
		ty: utils.Tools.GetTimeNow(),
	}
	err := redis.RedisClient.HMSet(context.Background(), key, ProgressInfo)
	if err != nil {
		logger.SlogError(fmt.Sprintf("ProgressEnd redis error: %s", err))
		return
	}
}

func (h *Handle) TaskEnd(target string, taskId string) {
	key := "TaskInfo:time:" + taskId
	err := redis.RedisClient.Set(context.Background(), key, utils.Tools.GetTimeNow())
	if err != nil {
		logger.SlogError(fmt.Sprintf("TaskEnds push redis error: %s", err))
		return
	}

	key = "TaskInfo:tmp:" + taskId
	_, err = redis.RedisClient.AddToList(context.Background(), key, target)
	if err != nil {
		logger.SlogError(fmt.Sprintf("TaskEnds push redis error: %s", err))
		return
	}
}

func (h *Handle) StopTask(id string) {
	logger.SlogInfo(fmt.Sprintf("stop task: %v", id))
	pebbledb.PebbleStore.Delete([]byte("task:" + id))
	TaskNodeName := "NodeTask:" + global.AppConfig.NodeName
	exists, err := redis.RedisClient.Exists(context.Background(), TaskNodeName)
	if err != nil {
		logger.SlogError(fmt.Sprintf("StopTask GetTask info error: %v", err))
		contextmanager.GlobalContextManagers.CancelContext(id)
		return
	}
	if exists {
		// 获取列表中的所有元素
		listLength, err := redis.RedisClient.LLen(context.Background(), TaskNodeName)
		if err != nil {
			logger.SlogError(fmt.Sprintf("Error getting list length: %v", err))
			contextmanager.GlobalContextManagers.CancelContext(id)
			return
		}
		if listLength == 0 {
			logger.SlogInfo("StopTask list is empty.")
			contextmanager.GlobalContextManagers.CancelContext(id)
			return
		}
		// 使用 LRANGE 获取列表中的所有值
		values, err := redis.RedisClient.LRange(context.Background(), TaskNodeName, 0, listLength-1)
		if err != nil {
			logger.SlogError(fmt.Sprintf("Error fetching list values: %v", err))
			contextmanager.GlobalContextManagers.CancelContext(id)
			return
		}
		// 遍历列表，检查是否包含目标字符串
		for _, value := range values {
			if strings.Contains(value, id) {
				// 删除包含目标字符串的元素
				err := redis.RedisClient.LRem(context.Background(), TaskNodeName, 0, value)
				if err != nil {
					logger.SlogError(fmt.Sprintf(" StopTaskError removing value: %v", err))
					contextmanager.GlobalContextManagers.CancelContext(id)
					return
				} else {
				}
			}
		}
	}
	contextmanager.GlobalContextManagers.CancelContext(id)
}

func (h *Handle) DeleteTask(content string) {
	for _, id := range strings.Split(content, ",") {
		h.StopTask(id)
		prefix := fmt.Sprintf("%s:", id)
		targets, err := pebbledb.PebbleStore.GetKeysWithPrefix(prefix)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("DeleteTask get targets error: %v", err))
			continue
		}
		for idTarget, _ := range targets {
			err := pebbledb.PebbleStore.Delete([]byte(idTarget))
			logger.SlogInfoLocal(fmt.Sprintf("DeleteTask target: %v", idTarget))
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("PebbleStore DeleteTask %v error: %v", idTarget, err))
			}
		}
	}
}

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
	logger.SlogInfo(fmt.Sprintf("Running start value: %d", h.RunningNum))
}

func (h *Handle) EndTask() {
	h.mu.Lock()         // 锁定互斥锁
	defer h.mu.Unlock() // 确保在函数结束时解锁
	h.RunningNum = h.RunningNum - 1
	h.FinNum = h.FinNum + 1
	logger.SlogInfo(fmt.Sprintf("Running start value: %d Running end value: %d", h.RunningNum, h.FinNum))
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
	logger.SlogInfo(fmt.Sprintf("%v module start scanning the target: %v", typ, target))
	key := "TaskInfo:progress:" + taskId + ":" + target
	ty := typ + "_start"
	ProgressInfo := map[string]interface{}{
		ty: utils.Tools.GetTimeNow(),
	}
	if typ == "scan" {
		ProgressInfo["node"] = global.AppConfig.NodeName
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
	logger.SlogInfo(fmt.Sprintf("%v module end scanning the target: %v running time: %v", typ, target, time))
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
	_ = h.PopTaskId(id)
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

func (h *Handle) PopTaskId(id string) error {
	TaskNodeName := "NodeTask:" + global.AppConfig.NodeName
	exists, err := redis.RedisClient.Exists(context.Background(), TaskNodeName)
	if err != nil {
		logger.SlogError(fmt.Sprintf("PopTaskId GetTask info error: %v", err))
		return err
	}
	if exists {
		// 获取列表中的所有元素
		listLength, err := redis.RedisClient.LLen(context.Background(), TaskNodeName)
		if err != nil {
			logger.SlogError(fmt.Sprintf("PopTaskId Error getting list length: %v", err))
			return err
		}
		if listLength == 0 {
			logger.SlogInfo("PopTaskId list is empty.")
			return err
		}
		// 使用 LRANGE 获取列表中的所有值
		values, err := redis.RedisClient.LRange(context.Background(), TaskNodeName, 0, listLength-1)
		if err != nil {
			logger.SlogError(fmt.Sprintf("PopTaskId Error fetching list values: %v", err))
			return err
		}
		// 遍历列表，检查是否包含目标字符串
		for _, value := range values {
			if strings.Contains(value, id) {
				// 删除包含目标字符串的元素
				err := redis.RedisClient.LRem(context.Background(), TaskNodeName, 0, value)
				if err != nil {
					logger.SlogError(fmt.Sprintf("PopTaskId removing value: %v", err))
					return err
				} else {
				}
			}
		}
	}
	return nil
}

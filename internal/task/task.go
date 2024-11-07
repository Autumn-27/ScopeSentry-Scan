// task-------------------------------------
// @file      : task.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 21:08
// -------------------------------------------

package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/bigcache"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handler"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pebbledb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/runner"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	goRedis "github.com/redis/go-redis/v9"
	"os"
	"strings"
	"sync"
	"time"
)

func GetTask() {
	// 运行本地缓存的任务
	RunPebbledbTask()
	// 从redis获取任务
	RunRedisTask()
}

// RunPebbledbTask 运行本地缓存任务
func RunPebbledbTask() {
	prefix := "task:"
	keys, err := pebbledb.PebbleStore.GetKeysWithPrefix(prefix)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("pebbledb get task error: %v", err))
		os.Exit(0)
	}
	if len(keys) > 0 {
		// 打印所有以 "task:" 开头的键值对
		for key, value := range keys {
			var wg sync.WaitGroup
			logger.SlogInfoLocal(fmt.Sprintf("get PebbleStore task: %v", string(value)))
			var runnerOption options.TaskOptions
			err = utils.Tools.JSONToStruct(value, &runnerOption)
			// 任务增加全局上下文
			contextmanager.GlobalContextManagers.AddContext(runnerOption.ID)
			// 设置为本地获取的任务
			runnerOption.IsRestart = true
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("task JSONToStruct error: %v - %v", string(value), err))
				err = pebbledb.PebbleStore.Delete([]byte(fmt.Sprintf("task:%s", value)))
				if err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("PebbleStore delete error: %v", value))
				}
				continue
			}
			prefix = fmt.Sprintf("%s:", runnerOption.ID)
			targets, err := pebbledb.PebbleStore.GetKeysWithPrefix(prefix)
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("pebbledb get task targets error: %v", err))
				err = pebbledb.PebbleStore.Delete([]byte(fmt.Sprintf("task:%s", value)))
				if err != nil {
					logger.SlogErrorLocal(fmt.Sprintf("PebbleStore delete error: %v", value))
				}
				continue
			}
			for idTarget, _ := range targets {
				wg.Add(1)
				// 创建 runnerOption 的副本
				optionCopy := runnerOption
				target := strings.SplitN(idTarget, ":", 2)
				optionCopy.Target = target[1]
				// 使用局部变量创建闭包
				taskFunc := func(op options.TaskOptions) func() {
					return func() {
						defer wg.Done()
						select {
						case <-contextmanager.GlobalContextManagers.GetContext(op.ID).Done():
							// 任务取消直接返回
							return
						default:
							err := runner.Run(op)
							if err != nil {
								// 说明该任务取消了，直接返回不进行删除目标
								return
							} else {
								// 目标运行完毕删除目标
								DeletePebbleTarget(pebbledb.PebbleStore, []byte(op.ID+":"+op.Target))
							}
						}
					}
				}(optionCopy)

				// 提交任务
				err := pool.PoolManage.SubmitTask("task", taskFunc)
				if err != nil {
					logger.SlogError(fmt.Sprintf("task pool error: %v", err))
					// 如果提交任务失败，手动减少计数
					wg.Done()
				}
			}
			wg.Wait()
			// 任务运行完毕删除任务
			err = pebbledb.PebbleStore.Delete([]byte("task:" + key))
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("PebbleStore Delete %v error: %v", key, err))
			}
			// 删除任务上下文
			contextmanager.GlobalContextManagers.DeleteContext(runnerOption.ID)
			logger.SlogInfoLocal(fmt.Sprintf("PebbleStore task run end: %v", runnerOption.ID))

		}
		// 关闭nuclei引擎
		handler.CloseNucleiEngine()
	}
}

// RunRedisTask 从redis中获取任务
func RunRedisTask() {
	ticker := time.Tick(3 * time.Second)
	for {
		<-ticker
		TaskNodeName := "NodeTask:" + global.AppConfig.NodeName
		exists, err := redis.RedisClient.Exists(context.Background(), TaskNodeName)

		if err != nil {
			logger.SlogError(fmt.Sprintf("GetTask Error: %v", err))
			continue
		}
		if exists {
			var wg sync.WaitGroup

			taskInfo, err := redis.RedisClient.GetFirstFromList(context.Background(), TaskNodeName)
			if err != nil {
				logger.SlogError(fmt.Sprintf("GetTask info error: %v", err))
				continue
			}
			logger.SlogInfo(fmt.Sprintf("Get a new task: %v", taskInfo))
			var runnerOption options.TaskOptions
			err = json.Unmarshal([]byte(taskInfo), &runnerOption)
			if err != nil {
				logger.SlogError(fmt.Sprintf("Task parse error: %s", err))
				continue
			}
			// 将任务配置写入本地
			runnerOption.IsRestart = false
			taskKey := fmt.Sprintf("task:%v", runnerOption.ID)
			err = pebbledb.PebbleStore.Put([]byte(taskKey), []byte(taskInfo))
			if err != nil {
				logger.SlogError(fmt.Sprintf("PebbleStore.Put Task error: %s", err))
				continue
			}
			logger.SlogInfo(fmt.Sprintf("Task begin: %v", runnerOption.ID))
			if runnerOption.Type == "page_monitoring" {
				// 运行页面监控程序
				for {
					targets, err := redis.RedisClient.BatchGetAndDelete(context.Background(), "TaskInfo:"+runnerOption.ID, 50)
					if len(targets) == 0 {
						break
					}
					if err != nil {
						// 如果 err 不为空，并且不是 redis.Nil 错误，则打印错误信息
						if !errors.Is(err, goRedis.Nil) {
							logger.SlogError(fmt.Sprintf("GetRedisTask BatchGetAndDelete error: %v", err))
							// 如果获取任务出错了 直接退出 防止删除本地任务 重启之后重新获取本地任务开始执行
							os.Exit(0)
						}
						break
					}
					runner.PageMonitoringRunner(targets)
				}
			} else {
				// 任务增加全局上下文
				contextmanager.GlobalContextManagers.AddContext(runnerOption.ID)
				for {
					target, err := redis.RedisClient.PopFromListR(context.Background(), "TaskInfo:"+runnerOption.ID)
					if err != nil {
						// 如果 err 不为空，并且不是 redis.Nil 错误，则打印错误信息
						if !errors.Is(err, goRedis.Nil) {
							logger.SlogError(fmt.Sprintf("GetRedisTask redis error: %v", err))
							// 如果获取任务出错了 直接退出 防止删除本地任务 重启之后重新获取本地任务开始执行
							os.Exit(0)
						}
						break
					}
					wg.Add(1)
					optionCopy := runnerOption
					optionCopy.Target = target
					taskFunc := func(op options.TaskOptions) func() {
						return func() {
							defer wg.Done()
							select {
							case <-contextmanager.GlobalContextManagers.GetContext(op.ID).Done():
								// 任务取消直接返回
								return
							default:
								err := runner.Run(op)
								if err != nil {
									// 说明该任务取消了，直接返回不进行删除目标
									return
								} else {
									// 目标运行完毕删除目标
									DeletePebbleTarget(pebbledb.PebbleStore, []byte(op.ID+":"+op.Target))
								}
							}
						}
					}(optionCopy)
					// 将任务目标写入本地
					err = pebbledb.PebbleStore.Put([]byte(fmt.Sprintf("%v:%v", runnerOption.ID, target)), []byte(""))
					if err != nil {
						logger.SlogError(fmt.Sprintf("PebbleStore.Put target error: %v", err))
					}
					// 提交任务
					err = pool.PoolManage.SubmitTask("task", taskFunc)
					if err != nil {
						logger.SlogError(fmt.Sprintf("task pool error: %v", err))
						wg.Done()
					}
					logger.SlogInfoLocal(fmt.Sprintf("task target pool running goroutines: %v", pool.PoolManage.GetModuleRunningGoroutines("task")))
				}
				wg.Wait()
				// 删除任务上下文
				contextmanager.GlobalContextManagers.DeleteContext(runnerOption.ID)
			}
			logger.SlogInfo(fmt.Sprintf("Task end: %v", runnerOption.ID))
			handler.CloseNucleiEngine()
			// 目标运行完毕 删除任务信息
			// 删除本地缓存任务信息
			err = pebbledb.PebbleStore.Delete([]byte(taskKey))
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("PebbleStore Delete %v error: %v", taskKey, err))
			}
			// 任务结束重新初始化花奴才能
			err = bigcache.Initialize()
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("bigcache Initialize error: %v", err))
			}
			// 删除redis信息
			_, err = redis.RedisClient.PopFirstFromList(context.Background(), TaskNodeName)
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("RemoveFirstFromList Delete %v error: %v", taskKey, err))
			}
		}
	}
}

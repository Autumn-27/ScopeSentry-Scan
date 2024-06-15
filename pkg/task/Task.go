package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/runner"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"strings"
	"sync"
	"time"
)

type opTask struct {
	TaskId            string
	SubdomainScan     bool
	Subfinder         bool
	Ksubdomain        bool
	UrlScan           bool
	SensitiveInfoScan bool
	PageMonitoring    string
	CrawlerScan       bool
	VulScan           bool
	VulList           []string
	Duplicates        string
	PortScan          bool
	Ports             string
	DirScan           bool
	Waybackurl        bool
	Type              string
}
type taskType struct {
	target string
	tp     string
	op     runner.Option
}

func GetTask() {
	MaxTaskNumInt, err := strconv.Atoi(system.AppConfig.System.MaxTaskNum)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	tasks := make(chan taskType, 1)
	var wg sync.WaitGroup
	var mu sync.Mutex
	// 启动初始数量的工作goroutine
	for i := 1; i <= MaxTaskNumInt; i++ {
		wg.Add(1)
		go ParseTask("线程"+strconv.Itoa(i)+" begin", tasks, &wg)
	}
	go func() {
		ticker := time.Tick(3 * time.Second)
		for {
			<-ticker
			errorm := system.RedisClient.Ping(context.Background())
			if errorm != nil {
				system.GetRedisClient()
			}
			if system.AppConfig.System.State == "2" {
				for {
					time.Sleep(3 * time.Second)
					if system.AppConfig.System.State == "1" {
						break
					}
				}
			}
			//redisClientInstance.HMSet(context.Background())
			TaskNodeName := "NodeTask:" + system.AppConfig.System.NodeName
			exists, err := system.RedisClient.Exists(context.Background(), TaskNodeName)
			if err != nil {
				myLog := system.CustomLog{
					Status: "Error",
					Msg:    fmt.Sprintf("GetTask Error", err),
				}
				system.PrintLog(myLog)
				continue
			}
			if exists {
				system.SlogInfo("Get a new task~")
				r, err := system.RedisClient.PopFromListR(context.Background(), TaskNodeName)
				if err != nil {
					system.SlogError(fmt.Sprintf("GetTask Error 2", err))
					continue
				}
				taskOpt := opTask{}
				err2 := json.Unmarshal([]byte(r), &taskOpt)
				if err2 != nil {
					system.SlogError(fmt.Sprintf("Task parse error: %s", err))
					continue
				}
				for {
					newMaxTaskNumInt, err := strconv.Atoi(system.AppConfig.System.MaxTaskNum)
					if err != nil {
						fmt.Println("err:", err)
						return
					}
					if MaxTaskNumInt != newMaxTaskNumInt {
						mu.Lock()
						diff := newMaxTaskNumInt - MaxTaskNumInt
						if newMaxTaskNumInt > MaxTaskNumInt {
							for i := 0; i < diff; i++ {
								wg.Add(1)
								go ParseTask("新增线程"+strconv.Itoa(i), tasks, &wg)
							}
						} else {
							for i := 0; i < -diff; i++ {
								tmpTask := taskType{target: "STOP"}
								tasks <- tmpTask // 发送一个停止信号来停止一个goroutine
							}
						}
						MaxTaskNumInt = newMaxTaskNumInt
						mu.Unlock()
					}
					scanTarget, err := system.RedisClient.PopFromListR(context.Background(), "TaskInfo:"+taskOpt.TaskId)
					if err != nil {
						// 如果 err 不为空，并且不是 redis.Nil 错误，则打印错误信息
						if !errors.Is(err, redis.Nil) {
							system.SlogError(fmt.Sprintf("GetTask Error: %s", err))
						}
						// 如果 err 是 redis.Nil 错误，则说明列表已经为空，退出循环
						break
					}
					scanTask := taskType{
						target: scanTarget,
						tp:     taskOpt.Type,
						op: runner.Option{
							SubfinderEnabled:      taskOpt.Subfinder,
							SubdomainScanEnabled:  taskOpt.SubdomainScan,
							KsubdomainScanEnabled: taskOpt.Ksubdomain,
							PortScanEnabled:       taskOpt.PortScan,
							DirScanEnabled:        taskOpt.DirScan,
							CrawlerEnabled:        taskOpt.CrawlerScan,
							Ports:                 taskOpt.Ports,
							WaybackurlEnabled:     taskOpt.Waybackurl,
							UrlScan:               taskOpt.UrlScan,
							Cookie:                "",
							Header:                []string{},
							Duplicates:            taskOpt.Duplicates,
							TaskId:                taskOpt.TaskId,
							SensitiveInfoScan:     taskOpt.SensitiveInfoScan,
							VulScan:               taskOpt.VulScan,
							VulList:               taskOpt.VulList,
							PageMonitoring:        taskOpt.PageMonitoring,
						},
					}
					tasks <- scanTask
					system.SlogDebugLocal(fmt.Sprintf("当前 tasks 通道内的元素数量：%d", len(tasks)))
				}
			}
			errorm = system.MongoClient.Ping()
			if errorm != nil {
				system.GetMongbClient()
			}
		}
	}()

	wg.Wait()
	close(tasks)
}

func ParseTask(msg string, tasks <-chan taskType, wg *sync.WaitGroup) {
	defer wg.Done()
	if msg != "" {
		system.SlogDebugLocal(msg)
	}
	for task := range tasks {
		if task.target == "STOP" {
			system.SlogDebugLocal("Task Worker received stop signal")
			return
		}
		system.SlogDebugLocal(fmt.Sprintf("接收到扫描任务: %+v", task))
		if task.tp == "page_monitoring" {
			if task.target != "" {
				runner.PageMRun(task.target)
			}
		} else {
			if task.op.PortScanEnabled {
				if task.op.Ports != "" {
					ports, err := getPortListById(task.op.Ports)
					if err != nil {
						for _, p := range system.PortDict {
							if p.ID == task.op.Ports {
								task.op.Ports = parsePort(p.Value)
							}
						}
					}
					task.op.Ports = parsePort(ports)
				}
			}
			runner.Process(task.target, task.op)
		}
	}
}

func getPortListById(objectIDString string) (string, error) {
	var result struct {
		Value string `bson:"value"`
	}
	id, err := primitive.ObjectIDFromHex(objectIDString)
	if err != nil {
		system.SlogError(fmt.Sprintf("Invalid ObjectID: %v", err))
	}
	if err := system.MongoClient.FindOne("PortDict", bson.M{"_id": id}, bson.M{"_id": 0, "value": 1}, &result); err != nil {
		system.SlogError(fmt.Sprintf("getPortListById error: %s", err))
		return "", err
	}
	return result.Value, nil
}

func parsePort(ports string) string {
	parts := strings.Split(ports, ",")
	tmp := ""
	for _, p := range parts {
		if p == "80" || p == "443" {
			continue
		}
		if strings.Contains(p, "-") {
			s := strings.Split(p, "-")
			start, _ := strconv.Atoi(s[0])
			end, _ := strconv.Atoi(s[1])
			for i := start; i <= end; i++ {
				character := strconv.Itoa(i)
				tmp += character + ","
			}
			continue
		}
		tmp += p + ","
	}
	return tmp[:len(tmp)-1]
}

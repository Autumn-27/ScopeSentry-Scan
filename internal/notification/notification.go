// notification-------------------------------------
// @file      : notification.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/19 19:54
// -------------------------------------------

package notification

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"strings"
	"time"
)

const (
	batchSize     = 20
	flushInterval = 2 * time.Second // 每2秒从队列中取数据
)

type NotificationQueue struct {
	Queue chan string
}

var NotificationQueues = make(map[string]*NotificationQueue)

func InitializeNotification() {
	// 模块列表
	modules := []string{
		"SubdomainScan", "SubdomainSecurity",
		"AssetMapping", "PortScan", "URLScan",
		"URLSecurity", "DirScan", "WebCrawler", "VulnerabilityScan", "PageMonitor",
	}
	// 初始化模块队列和 Goroutine
	for _, module := range modules {
		NotificationQueues[module] = &NotificationQueue{
			Queue: make(chan string, 50), // 缓存队列大小可以大于 batchSize
		}
		go processQueue(module, NotificationQueues[module])
	}
}

func processQueue(module string, mq *NotificationQueue) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		// 每2秒触发一次处理逻辑
		case <-ticker.C:
			processBatch(module, mq)
		}
	}
}

// processBatch 从队列中取出最多 batchSize 条数据进行处理
func processBatch(module string, mq *NotificationQueue) {
	var buffer = ""
	num := 0
	// 尝试取出 batchSize 条数据，如果不足则取剩下的所有数据
	for i := 0; i < batchSize; i++ {
		select {
		case msg := <-mq.Queue:
			num += 1
			buffer += msg
		default:
			// 如果队列中没有更多数据，跳出循环
			break
		}
	}

	if num > 0 {
		FlushBuffer(module, &buffer)
	}
}

// FlushBuffer 模拟处理队列中的数据
func FlushBuffer(module string, buffer *string) {
	// 处理消息
	*buffer = "[" + global.AppConfig.NodeName + "][" + module + "]results:\n" + *buffer
	for _, api := range global.NotificationApi {
		uri := strings.Replace(api.Url, "*msg*", *buffer, -1)
		if api.Method == "GET" {
			_, err := utils.Requests.HttpGet(uri)
			if err != nil {
				logger.SlogError(fmt.Sprintf("SendNotification %v HTTP Get Error: %v", uri, err))
			}
		} else {
			data := strings.Replace(api.Data, "*msg*", *buffer, -1)
			err, _ := utils.Requests.HttpPost(uri, []byte(data), api.ContentType)
			if err != nil {
				logger.SlogError(fmt.Sprintf("SendNotification %v HTTP Post Error: %v", uri, err))
			}
		}
	}
	// 清空缓冲区
	*buffer = ""
}

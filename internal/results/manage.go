// results-------------------------------------
// @file      : manage.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/17 20:59
// -------------------------------------------

package results

import (
	"time"
)

const (
	batchSize     = 50
	flushInterval = 30 * time.Second
)

type ResultQueue struct {
	Queue   chan interface{}
	CloseCh chan struct{}
}

var ResultQueues = make(map[string]*ResultQueue)

func InitializeResultQueue() {
	// 模块列表
	modules := []string{
		"SubdomainScan", "SubdomainSecurity",
		"AssetChangeLog", "URLScan",
		"WebCrawler", "VulnerabilityScan",
		"SensitiveResult", "DirScan",
	}
	// 初始化模块队列和 Goroutine
	for _, module := range modules {
		ResultQueues[module] = &ResultQueue{
			Queue:   make(chan interface{}, batchSize),
			CloseCh: make(chan struct{}),
		}
		go processQueue(module, ResultQueues[module])
	}

	// 初始化去重模块
	InitializeDuplicate()
	// 初始化结果处理模块
	InitializeHandler()
	// 初始化结果插入模块
	InitializeResults()
}

func processQueue(module string, mq *ResultQueue) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	var buffer []interface{}

	for {
		select {
		case batch := <-mq.Queue:
			if batch != nil {
				buffer = append(buffer, batch)
				if len(buffer) >= batchSize {
					flushBuffer(module, &buffer)
				}
			}

		case <-ticker.C:
			if len(buffer) > 0 {
				flushBuffer(module, &buffer)
			}
		case <-mq.CloseCh:
			// 处理关闭信号
			if len(buffer) > 0 {
				flushBuffer(module, &buffer)
			}
			return
		}
	}
}

func flushBuffer(module string, buffer *[]interface{}) {
	if len(*buffer) == 0 {
		return
	}
	var name string
	switch module {
	case "SubdomainScan":
		name = "subdomain"
	case "SubdomainSecurity":
		name = "SubdoaminTakerResult"
	case "AssetChangeLog":
		name = "AssetChangeLog"
	case "URLScan":
		name = "UrlScan"
	case "SensitiveResult":
		name = "SensitiveResult"
	case "WebCrawler":
		name = "crawler"
	case "VulnerabilityScan":
		name = "vulnerability"
	case "DirScan":
		name = "DirScanResult"

	}
	Results.Insert(name, buffer)
	*buffer = nil
}

func Close() {
	for _, mq := range ResultQueues {
		close(mq.Queue)   // 关闭队列
		close(mq.CloseCh) // 发送关闭信号
	}
}

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
	batchSize     = 60
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
		"SensitiveResult", "DirScan", "PageMonitoring", "PageMonitoringBody", "SensitiveBody",
	}
	// 初始化模块队列和 Goroutine
	for _, module := range modules {
		if module == "PageMonitoringUrl" {
			ResultQueues[module] = &ResultQueue{
				Queue:   make(chan interface{}, 100),
				CloseCh: make(chan struct{}),
			}
		} else if module == "URLScan" {
			ResultQueues[module] = &ResultQueue{
				Queue:   make(chan interface{}, 250),
				CloseCh: make(chan struct{}),
			}
		} else if module == "SensitiveResult" {
			ResultQueues[module] = &ResultQueue{
				Queue:   make(chan interface{}, 120),
				CloseCh: make(chan struct{}),
			}
		} else if module == "SensitiveBody" {
			ResultQueues[module] = &ResultQueue{
				Queue:   make(chan interface{}, 15),
				CloseCh: make(chan struct{}),
			}
		} else {
			ResultQueues[module] = &ResultQueue{
				Queue:   make(chan interface{}, batchSize),
				CloseCh: make(chan struct{}),
			}
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
	if module == "URLScan" {
		ticker = time.NewTicker(60 * time.Second)
	}
	defer ticker.Stop()

	var buffer []interface{}

	for {
		select {
		case batch := <-mq.Queue:
			if batch != nil {
				buffer = append(buffer, batch)
				if len(buffer) >= batchSize-2 {
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
	if module == "SensitiveBody" {
		Results.Update(buffer)
	} else {
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
		case "PageMonitoring":
			name = "PageMonitoring"
		case "PageMonitoringBody":
			name = "PageMonitoringBody"
		}
		Results.Insert(name, buffer)
	}
	*buffer = nil
}

func Close() {
	for _, mq := range ResultQueues {
		close(mq.Queue)   // 关闭队列
		close(mq.CloseCh) // 发送关闭信号
	}
}

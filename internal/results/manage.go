// results-------------------------------------
// @file      : manage.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/17 20:59
// -------------------------------------------

package results

import (
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

const (
	batchSize     = 50
	flushInterval = 30 * time.Second
)

type ResultQueue struct {
	queue chan []interface{}
}

var ResultQueues = make(map[string]*ResultQueue)

func Initialize() {
	// 模块列表
	modules := []string{
		"SubdomainScan", "SubdomainSecurity",
		"AssetMapping", "PortScan", "AssetResultHandl", "URLScan",
		"URLSecurity", "WebCrawler", "VulnerabilityScan",
	}
	// 初始化模块队列和 Goroutine
	for _, module := range modules {
		ResultQueues[module] = &ResultQueue{
			queue: make(chan []interface{}, batchSize),
		}
		//go processQueue(module, ResultQueues[module])
	}

	//初始化去重模块
	InitializeDuplicate()
}

func processQueue(module string, mq *ResultQueue) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	var buffer []interface{}

	for {
		select {
		case batch := <-mq.queue:
			buffer = append(buffer, batch...)
			if len(buffer) >= batchSize {
				//flushBuffer(module, mq.collection, &buffer)
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				//flushBuffer(module, mq.collection, &buffer)
			}
		}
	}
}

func flushBuffer(module string, collection *mongo.Collection, buffer *[]interface{}) {
	if len(*buffer) == 0 {
		return
	}

	*buffer = nil
}

// global-------------------------------------
// @file      : passivescan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/2/9 20:12
// -------------------------------------------

package global

import (
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/passivescan"
	"sync"
)

var TaskPassiveScanGlobal = make(map[string]chan interface{})
var mu sync.Mutex // 声明一个互斥锁
var PassiveScanWgMap = make(map[string]*sync.WaitGroup)

func SetPassiveScanChan(id string, op *options.TaskOptions) {
	mu.Lock()         // 获取锁，确保只有一个 goroutine 能修改 map
	defer mu.Unlock() // 函数结束时释放锁

	// 检查 id 是否已经存在
	if _, exists := TaskPassiveScanGlobal[id]; !exists {
		// 如果不存在，则创建一个新的 chan 并添加到字典
		vulnerabilityInputChan := make(chan interface{}, 100)
		TaskPassiveScanGlobal[id] = vulnerabilityInputChan
		passivescanModule := passivescan.NewRunner(op, nil)
		passivescanModule.SetInput(vulnerabilityInputChan)
		go func() {
			PassiveScanWgMap[id].Add(1)
			err := passivescanModule.ModuleRun()
			if err != nil {

			}
		}()
	}
}

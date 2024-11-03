// task-------------------------------------
// @file      : runner.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/9 21:47
// -------------------------------------------

package runner

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handler"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/modules"
	"sync"
	"time"
)

func Run(op options.TaskOptions) {
	var wg sync.WaitGroup
	var start time.Time
	var end time.Time
	start = time.Now()
	handler.TaskHandle.ProgressStart("scan", op.Target, op.ID, 1)
	op.ModuleRunWg = &wg
	op.TargetParser = append(op.TargetParser, "7bbaec6487f51a9aafeff4720c7643f0")
	process := modules.CreateScanProcess(&op)
	ch := make(chan interface{})
	process.SetInput(ch)
	go func() {
		err := process.ModuleRun()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}()
	ch <- op.Target
	close(ch)
	time.Sleep(10 * time.Second)
	wg.Wait()
	end = time.Now()
	duration := end.Sub(start)
	handler.TaskHandle.ProgressEnd("scan", op.Target, op.ID, 1, duration)
	handler.TaskHandle.TaskEnd(op.Target, op.ID)
}

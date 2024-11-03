// task-------------------------------------
// @file      : runner.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/9 21:47
// -------------------------------------------

package runner

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/modules"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"sync"
	"time"
)

func Run(op options.TaskOptions) {
	var wg sync.WaitGroup
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
	logger.SlogInfoLocal(fmt.Sprintf("ModuleRun completed: %v %v", op.ID, op.Target))
}

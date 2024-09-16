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
	wg.Add(1)
	op.TargetParser = append(op.TargetParser, "TargetParser")
	process := modules.CreateScanProcess(op)
	ch := make(chan interface{})
	process.SetInput(ch)
	go func() {
		defer wg.Done()
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
	logger.SlogInfoLocal(fmt.Sprintf("ModuleRun completed: %v %v\n", op.ID, op.Target))
}

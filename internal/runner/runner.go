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
	"time"
)

func Run(op options.TaskOptions) {
	fmt.Println(op.Target)
	time.Sleep(10 * time.Second)
	modules.CreateScanProcess(op)
}

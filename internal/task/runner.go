// task-------------------------------------
// @file      : runner.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/9 21:47
// -------------------------------------------

package task

import (
	"fmt"
	"time"
)

func Run(op Options) {
	fmt.Println(op.Target)
	time.Sleep(10 * time.Second)
}

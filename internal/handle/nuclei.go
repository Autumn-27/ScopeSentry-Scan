// handle-------------------------------------
// @file      : nuclei.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/20 14:48
// -------------------------------------------

package handle

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	nuclei "github.com/projectdiscovery/nuclei/v3/lib"
	"sync"
)

var NucleiEngine *nuclei.ThreadSafeNucleiEngine
var NucleiEngineWg sync.WaitGroup

func NewNucleiEngine() {
	ctx := context.Background()
	ne, err := nuclei.NewThreadSafeNucleiEngineCtx(ctx)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("NewNucleiEngine error: %v", err))
		return
	}
	NucleiEngine = ne
}

func CloseNucleiEngine() {
	NucleiEngineWg.Wait()
	if NucleiEngine != nil {
		NucleiEngine.Close()
	}
}

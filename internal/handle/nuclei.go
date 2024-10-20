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

var NucleiEngines []*nuclei.ThreadSafeNucleiEngine
var NucleiEngineWg sync.WaitGroup
var mu sync.Mutex

func NewNucleiEngine() *nuclei.ThreadSafeNucleiEngine {
	ctx := context.Background()
	ne, err := nuclei.NewThreadSafeNucleiEngineCtx(ctx, nuclei.DisableUpdateCheck())
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("NewNucleiEngine error: %v", err))
		return nil
	}
	mu.Lock()
	NucleiEngines = append(NucleiEngines, ne)
	mu.Unlock()
	return ne
}

func CloseNucleiEngine() {
	NucleiEngineWg.Wait()
	for _, ne := range NucleiEngines {
		if ne != nil {
			ne.Close()
		}
	}
}

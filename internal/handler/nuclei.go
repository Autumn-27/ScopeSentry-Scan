// handle-------------------------------------
// @file      : nuclei.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/20 14:48
// -------------------------------------------

package handler

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	nuclei "github.com/projectdiscovery/nuclei/v3/lib"
	"github.com/projectdiscovery/nuclei/v3/pkg/templates"
	"sync"
)

var NucleiEngines []*nuclei.ThreadSafeNucleiEngine
var NucleiEngineWg sync.WaitGroup
var mu sync.Mutex
var Parser *templates.Parser

func NewNucleiEngine() *nuclei.ThreadSafeNucleiEngine {
	mu.Lock()
	defer mu.Unlock()
	if Parser == nil {
		Parser = templates.NewParser()
	}
	ctx := context.Background()
	ne, err := nuclei.NewThreadSafeNucleiEngineCtx(ctx, Parser, nuclei.DisableUpdateCheck())
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("NewNucleiEngine error: %v", err))
		return nil
	}
	NucleiEngines = append(NucleiEngines, ne)
	return ne
}

func CloseNucleiEngine() {
	NucleiEngineWg.Wait()
	for _, ne := range NucleiEngines {
		if ne != nil {
			ne.Close()
		}
	}
	Parser = nil
}

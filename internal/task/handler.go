// task-------------------------------------
// @file      : handler.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/11/1 21:22
// -------------------------------------------

package task

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pebbledb"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
)

func DeletePebbleTarget(PebbleStore *pebbledb.PebbleDB, targetKey []byte) {
	err := PebbleStore.Delete(targetKey)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("PebbleStore Delete error: %v", err))
	}
}

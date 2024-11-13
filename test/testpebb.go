// main-------------------------------------
// @file      : testpebb.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/11/13 20:42
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pebbledb"
	"path/filepath"
)

func main() {
	config.Initialize()
	pebbledbSetting := pebbledb.Settings{
		DBPath:       filepath.Join(global.AbsolutePath, "data", "pebbledb"),
		CacheSize:    64 << 20,
		MaxOpenFiles: 500,
	}
	pebbledbOption := pebbledb.GetPebbleOptions(&pebbledbSetting)
	if !global.AppConfig.Debug {
		pebbledbOption.Logger = nil
	}
	pedb, _ := pebbledb.NewPebbleDB(pebbledbOption, pebbledbSetting.DBPath)
	pebbledb.PebbleStore = pedb
	pebbledb.PebbleStore.Delete([]byte("task:672f56b023162c195f6e0933"))
	prefix := "task:"
	keys, _ := pebbledb.PebbleStore.GetKeysWithPrefix(prefix)
	if len(keys) > 0 {
		// 打印所有以 "task:" 开头的键值对
		for key, value := range keys {
			fmt.Println(key, value)
		}
	}
}

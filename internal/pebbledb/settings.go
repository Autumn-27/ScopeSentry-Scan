// pebbledb-------------------------------------
// @file      : settings.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 21:30
// -------------------------------------------

package pebbledb

import (
	"github.com/cockroachdb/pebble"
	"time"
)

// Settings 用于 PebbleDB 的配置
type Settings struct {
	DBPath       string
	CacheSize    int64         // 缓存大小
	MaxOpenFiles int           // 最大打开文件数
	L0Compaction int           // L0 层的压缩触发条件
	Timeout      time.Duration // 操作超时时间
}

// DefaultSettings 返回默认的 PebbleDB 设置
func DefaultSettings() *Settings {
	return &Settings{
		DBPath:       "./data/pebbledb",
		CacheSize:    64 << 20, // 64MB
		MaxOpenFiles: 500,
		L0Compaction: 4,
		Timeout:      30 * time.Second,
	}
}

// GetPebbleOptions 根据 Settings 生成 pebble.Options
func GetPebbleOptions(settings *Settings) *pebble.Options {
	return &pebble.Options{
		Cache:         pebble.NewCache(settings.CacheSize),
		MaxOpenFiles:  settings.MaxOpenFiles,
		LBaseMaxBytes: 64 << 20, // 64MB
		Levels: []pebble.LevelOptions{
			{Compression: pebble.NoCompression, TargetFileSize: 2 << 20}, // 2MB target file size
		},
	}
}

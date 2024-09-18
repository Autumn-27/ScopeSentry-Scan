// bigcache-------------------------------------
// @file      : bigcache.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/18 20:18
// -------------------------------------------

package bigcache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"time"
)

// BigCacheWrapper 封装 BigCache 的实现
type BigCacheWrapper struct {
	cache *bigcache.BigCache
}

var BigCache *BigCacheWrapper

// NewBigCacheWrapper 创建一个 BigCache 实例
func Initialize() error {
	config := bigcache.Config{
		LifeWindow:       10 * 365 * 24 * time.Hour, // 缓存生命周期长
		Shards:           1024,
		MaxEntrySize:     500, // 每个缓存项最大大小
		HardMaxCacheSize: 100, // 最大缓存大小 100 MB
		Verbose:          true,
		CleanWindow:      5 * time.Minute,
	}
	bigCache, err := bigcache.New(context.Background(), config)
	if err != nil {
		return err
	}
	BigCache = &BigCacheWrapper{cache: bigCache}
	return nil
}

// Set 将数据存储到缓存
func (b *BigCacheWrapper) Set(key string, value []byte) error {
	return b.cache.Set(key, value)
}

// Get 从缓存中获取数据
func (b *BigCacheWrapper) Get(key string) ([]byte, error) {
	return b.cache.Get(key)
}

// Delete 从缓存中删除数据
func (b *BigCacheWrapper) Delete(key string) error {
	return b.cache.Delete(key)
}

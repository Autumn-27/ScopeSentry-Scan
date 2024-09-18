// results-------------------------------------
// @file      : duplicates.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/18 20:04
// -------------------------------------------

package results

import (
	"context"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/bigcache"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type duplicate struct {
}

var Duplicate *duplicate

func InitializeDuplicate() {
	Duplicate = &duplicate{}
}

// SubdomainInTask 本地缓存taskid:subdomain 是否存在，不存在则存入本地缓存，从redis查询是否重复，如果开启了子域名去重，则查询mongdob中是否存在子域名。
func (d *duplicate) SubdomainInTask(result *types.SubdomainResult) bool {
	key := result.TaskId + ":subdomain:" + result.Host
	_, err := bigcache.BigCache.Get(key)
	if err != nil {
		// bigcache中不存在，设置缓存
		err := bigcache.BigCache.Set(key, []byte{})
		if err != nil {
			logger.SlogError(fmt.Sprintf("Set BigCache error: %v - %v", key, err))
		}
		keyRedis := "duplicates:domain:" + result.TaskId
		ctx := context.Background()
		exists, err := redis.RedisClient.SIsMember(ctx, keyRedis, result.Host)
		if err != nil {
			logger.SlogError(fmt.Sprintf("SubdomainInTask Deduplication error %v", err))
			// 如果查询redis出错 直接认为不存在重复的
			return true
		}
		if exists {
			// 如果redis中已经存在子域名了，表示其他节点或该节点之前已经在扫描该子域名了，返回false跳过此域名
			return false
		} else {
			_, err = system.RedisClient.SAdd(ctx, keyRedis, result.Host)
			if err != nil {
				logger.SlogError(fmt.Sprintf("SubdomainInTask Deduplication sadd error %v", err))
			}
			// 子域名在redis中不存在，表示该子域名还没有进行扫描，返回true开始扫描
			return true
		}
	}
	return true
}

func (d *duplicate) SubdomainInMongoDb(result *types.SubdomainResult) bool {
	var resultDoc bson.M
	err := system.MongoClient.FindOne("subdomain", bson.M{"host": result.Host}, bson.M{"_id": 1}, &resultDoc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// 从mongodb中没有查到子域名，返回true表示开始该子域名的扫描
			return true
		}
		// mongodb查询错误，忽略该错误直接开始扫描该域名
		logger.SlogErrorLocal(fmt.Sprintf("SubdomainInMongoDb error :%s\n", err))
		return true
	}
	// 从mongodb中找到了该域名，进行去重，返回false，不进行该子域名的扫描
	return false
}

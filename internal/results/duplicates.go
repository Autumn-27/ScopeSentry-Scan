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
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/url"
	"strings"
)

type duplicate struct {
}

var Duplicate *duplicate

func InitializeDuplicate() {
	Duplicate = &duplicate{}
}

// SubdomainInTask 本地缓存taskid:subdomain 是否存在，不存在则存入本地缓存，从redis查询是否重复，如果开启了子域名去重，则查询mongdob中是否存在子域名。
func (d *duplicate) SubdomainInTask(taskId string, host string, isRestart bool) bool {
	key := "duplicates:" + taskId + ":subdomain:" + host
	flag := d.DuplicateLocalCache(key)
	if isRestart {
		return flag
	} else {
		if flag {
			keyRedis := "duplicates:" + taskId + ":domain"
			valueRedis := host
			flag = d.DuplicateRedisCache(keyRedis, valueRedis)
			return flag
		}
	}
	// 本地缓存中存在，返回false
	return false
}

func (d *duplicate) SubdomainInMongoDb(result *types.SubdomainResult) bool {
	var resultDoc bson.M
	err := mongodb.MongodbClient.FindOne("subdomain", bson.M{"host": result.Host}, bson.M{"_id": 1}, &resultDoc)
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

func (d *duplicate) PortIntask(taskId string, host string, port string, isRestart bool) bool {
	key := "duplicates:" + taskId + ":port:" + host + ":" + port
	flag := d.DuplicateLocalCache(key)
	if isRestart {
		return flag
	}
	if flag {
		// 本地缓存中不存在，从redis中查找
		keyRedis := "duplicates:" + taskId + ":port"
		valueRedis := host + "-" + port
		flag = d.DuplicateRedisCache(keyRedis, valueRedis)
		return flag
	}
	return false
}

// 返回true 表示 内存中不存在  返回false 表示存在 重复
func (d *duplicate) DuplicateLocalCache(key string) bool {
	_, err := bigcache.BigCache.Get(key)
	if err != nil {
		// bigcache中不存在，设置缓存
		err := bigcache.BigCache.Set(key, []byte{})
		if err != nil {
			logger.SlogError(fmt.Sprintf("Set BigCache error: %v - %v", key, err))
		}
		return true
	}
	// 本地缓存中存在，则重复
	return false
}

// DuplicateRedisCache 在key中查找是否存在value来进行去重，返回true 表示不存在 不重复 返回false 表示已经存在了 重复
func (d *duplicate) DuplicateRedisCache(key string, value string) bool {
	ctx := context.Background()
	exists, err := redis.RedisClient.SIsMember(ctx, key, value)
	if err != nil {
		logger.SlogError(fmt.Sprintf("PortIntask Deduplication error %v", err))
		// 如果查询redis出错 直接认为不存在重复的
		return true
	}
	if exists {
		// 如果redis中已经存在了，表示其他节点或该节点之前已经在扫描该端口了，返回false跳过此域名
		return false
	} else {
		_, err = redis.RedisClient.SAdd(ctx, key, value)
		if err != nil {
			logger.SlogError(fmt.Sprintf("PortIntask Deduplication sadd error %v", err))
		}
		// 子域名在redis中不存在，表示该域名-端口还没有进行扫描，返回true开始扫描
		return true
	}
}

func (d *duplicate) AssetInMongodb(host string, port string) (bool, string, bson.M) {
	var result bson.M
	err := mongodb.MongodbClient.FindOne("asset", bson.M{"host": host, "port": port}, nil, &result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// 说明在mongodb中不存在
			return false, "", nil
		}
		// 其他错误也认为在mongodb中不存在
		logger.SlogErrorLocal(fmt.Sprintf("AssetInMongodb error :%s\n", err))
		return false, "", nil
	} else {
		// 获取并删除 _id 字段，转换为字符串
		var id string
		if objId, ok := result["_id"].(primitive.ObjectID); ok {
			id = objId.Hex()      // 将 ObjectID 转为字符串
			delete(result, "_id") // 从 result 中删除 _id 字段
		}
		// 在mongodb中找到了这个资产记录
		return true, id, result
	}
}

func (d *duplicate) URL(rawUrl string, taskId string) bool {
	dupKey := d.URLParams(rawUrl)
	key := "duplicates:" + taskId + ":url:" + dupKey
	return d.DuplicateLocalCache(key)
}

func (d *duplicate) Crawler(value string, taskId string) bool {
	key := "duplicates:" + taskId + ":crawler:" + value
	return d.DuplicateLocalCache(key)
}

func (d *duplicate) URLParams(rawUrl string) string {
	parsedURL, err := url.Parse(rawUrl)
	dupKey := utils.Tools.CalculateMD5(strings.TrimLeft(strings.TrimLeft(strings.TrimSuffix(rawUrl, "/"), "http://"), "https://"))
	if err != nil {
	} else {
		queryParams := parsedURL.Query()
		if len(queryParams) > 0 {
			paramskey := fmt.Sprintf("%s%s", parsedURL.Host, strings.TrimSuffix(parsedURL.Path, "/"))
			for key, _ := range queryParams {
				paramskey += key
			}
			dupKey = utils.Tools.CalculateMD5(paramskey)
		}
	}
	fmt.Println(dupKey)
	return dupKey
}

func (d *duplicate) SensitiveBody(md5 string, taskId string) bool {
	key := "duplicates:" + taskId + ":SensitiveBody:" + md5
	return d.DuplicateLocalCache(key)
}

func (d *duplicate) DuplicateUrlFileKey(filename string, taskId string) bool {
	key := "duplicates:" + taskId + ":urlfile:" + filename
	if d.DuplicateLocalCache(key) {
		keyRedis := "duplicates:" + taskId + ":urlfile"
		valueRedis := filename
		return d.DuplicateRedisCache(keyRedis, valueRedis)
	} else {
		return false
	}
}

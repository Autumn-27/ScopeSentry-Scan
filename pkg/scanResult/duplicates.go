// Package result-----------------------------
// @file      : runner.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/04/06 20:20
// -------------------------------------------
package scanResult

import (
	"context"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"github.com/cloudflare/cfssl/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/url"
	"strings"
	"time"
)

func SubdomainRedisDeduplication(domain string, taskId string) bool {
	topDomain, subDomain := getTopDomain(domain)
	if topDomain == "" {
		return false
	}
	keyDomain := ""
	if subDomain == "" {
		keyDomain = topDomain + ":" + topDomain
	} else {
		keyDomain = topDomain + ":" + subDomain
	}
	errorm := system.RedisClient.Ping(context.Background())
	if errorm != nil {
		system.GetRedisClient()
	}

	key := "duplicates:domain:" + taskId
	ctx := context.Background()
	exists, err := system.RedisClient.SIsMember(ctx, key, keyDomain)
	if err != nil {
		system.SlogError(fmt.Sprintf("URLRedisDeduplication error %v", err))
		return false
	}
	expirationDuration := 5 * time.Hour
	err = system.RedisClient.Expire(context.Background(), key, expirationDuration)
	if err != nil {
		system.SlogError(fmt.Sprintf("Error subdomain %v setting expiration:%v", taskId, err))
	}
	if exists {
		return true
	} else {
		_, err = system.RedisClient.SAdd(ctx, key, keyDomain)
		if err != nil {
			system.SlogError(fmt.Sprintf("URLRedisDeduplication sadd error %v", err))
			return false
		}
		return false
	}
}

func URLRedisDeduplication(url string, taskId string) bool {
	errorm := system.RedisClient.Ping(context.Background())
	if errorm != nil {
		system.GetRedisClient()
	}
	key := "duplicates:url:" + taskId
	ctx := context.Background()
	url = util.CalculateMD5(url)
	exists, err := system.RedisClient.SIsMember(ctx, key, url)
	if err != nil {
		system.SlogError(fmt.Sprintf("URLRedisDeduplication error %v", err))
		return false
	}
	expirationDuration := 5 * time.Hour
	err = system.RedisClient.Expire(context.Background(), key, expirationDuration)
	if err != nil {
		system.SlogError(fmt.Sprintf("Error subdomain %v setting expiration:%v", taskId, err))
	}
	if exists {
		return true
	} else {
		_, err = system.RedisClient.SAdd(ctx, key, url)
		if err != nil {
			system.SlogError(fmt.Sprintf("URLRedisDeduplication sadd error %v", err))
			return false
		}
		return false
	}
}
func CrawRedisDeduplication(url string, taskId string) bool {
	errorm := system.RedisClient.Ping(context.Background())
	if errorm != nil {
		system.GetRedisClient()
	}
	key := "duplicates:craw:" + taskId
	ctx := context.Background()
	url = util.CalculateMD5(url)
	exists, err := system.RedisClient.SIsMember(ctx, key, url)
	if err != nil {
		system.SlogError(fmt.Sprintf("CrawRedisDeduplication error %v", err))
		return false
	}
	expirationDuration := 5 * time.Hour
	err = system.RedisClient.Expire(context.Background(), key, expirationDuration)
	if err != nil {
		system.SlogError(fmt.Sprintf("Error CrawRedisDeduplication %v setting expiration:%v", taskId, err))
	}
	if exists {
		return true
	} else {
		_, err = system.RedisClient.SAdd(ctx, key, url)
		if err != nil {
			system.SlogError(fmt.Sprintf("CrawRedisDeduplication sadd error %v", err))
			return false
		}
		return false
	}
}
func getTopDomain(domain string) (string, string) {
	u, _ := url.Parse("http://" + domain)
	parts := strings.Split(u.Hostname(), ".")
	if len(parts) < 2 {
		return "", ""
	}

	topDomain := strings.Join(parts[len(parts)-2:], ".")
	subDomain := strings.Join(parts[:len(parts)-2], ".")
	return topDomain, subDomain
}

type tmpRes struct {
	ID primitive.ObjectID `bson:"_id"`
}

func SubdomainMongoDbDeduplication(domain string) bool {
	tmp := tmpRes{}
	errorm := system.MongoClient.Ping()
	if errorm != nil {
		system.GetMongbClient()
	}
	erro := system.MongoClient.FindOne("subdomain", bson.M{"host": domain}, bson.M{"_id": 1}, &tmp)
	if erro != nil {
		if errors.Is(erro, mongo.ErrNoDocuments) {
			return false
		}
		log.Error(fmt.Sprintf("SubdomainMongoDbDeduplication error :%s\n", erro))
		return false
	}
	return true
}

type port struct {
	Port string `bson:"port"`
}

func GetPortByHost(host string) string {
	ports := []port{}
	errorm := system.MongoClient.Ping()
	if errorm != nil {
		system.GetMongbClient()
	}
	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.M{
			"$or": []bson.M{
				{"host": host},
				{"url": "http://" + host},
				{"url": "https://" + host},
			},
		}}},
		bson.D{{"$group", bson.M{
			"_id":  "$port",
			"port": bson.M{"$first": "$port"},
		}}},
		bson.D{{"$project", bson.M{
			"_id":  0,
			"port": 1,
		}}},
	}
	err := system.MongoClient.Aggregate("asset", pipeline, &ports)
	if err != nil {
		return ""
	}

	var result string
	for _, p := range ports {
		result += p.Port + ","
	}
	if strings.HasSuffix(result, ",") {
		// 去掉最后一个逗号
		result = result[:len(result)-1]
	}
	return result
}

func PageMonitoringMongoDbDeduplication(url string) (bool, types.PageMonitResult) {
	tmp := types.PageMonitResult{}
	errorm := system.MongoClient.Ping()
	if errorm != nil {
		system.GetMongbClient()
	}
	erro := system.MongoClient.FindOne("PageMonitoring", bson.M{"url": url}, bson.M{"_id": 1, "content": 1, "hash": 1, "diff": 1}, &tmp)
	if erro != nil {
		if errors.Is(erro, mongo.ErrNoDocuments) {
			return false, types.PageMonitResult{}
		}
		log.Error(fmt.Sprintf("PageMonitoringMongoDbDeduplication error :%s\n", erro))
		return false, types.PageMonitResult{}
	}
	return true, tmp
}

func SensRedisDeduplication(md5 string, taskId string) bool {
	errorm := system.RedisClient.Ping(context.Background())
	if errorm != nil {
		system.GetRedisClient()
	}
	key := "duplicates:sensresp:" + taskId
	ctx := context.Background()
	exists, err := system.RedisClient.SIsMember(ctx, key, md5)
	if err != nil {
		system.SlogError(fmt.Sprintf("SensRedisDeduplication error %v", err))
		return false
	}
	expirationDuration := 5 * time.Hour
	err = system.RedisClient.Expire(context.Background(), key, expirationDuration)
	if err != nil {
		system.SlogError(fmt.Sprintf("Error SensRedisDeduplication %v setting expiration:%v", taskId, err))
	}
	if exists {
		return true
	} else {
		_, err = system.RedisClient.SAdd(ctx, key, md5)
		if err != nil {
			system.SlogError(fmt.Sprintf("SensRedisDeduplication sadd error %v", err))
			return false
		}
		return false
	}
}

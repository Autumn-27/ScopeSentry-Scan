// configupdater-------------------------------------
// @file      : configupdater.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/7 20:24
// -------------------------------------------

package configupdater

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
)

// UpdateGlobalModulesConfig 拉取server的全局模块配置
func UpdateGlobalModulesConfig() {
	logger.SlogInfoLocal("system config load begin")
	var result ConfigResult
	err := mongodb.MongodbClient.FindOne("config", bson.M{"name": "ModulesConfig"}, bson.M{"_id": 0, "value": 1}, &result)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("system config load error: %v", err))
		return
	}
	err = utils.WriteContentFile(config.ModulesConfigPath, result.Value)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("system config writing file error: %v", err))
		return
	}
	logger.SlogInfoLocal("system config load end")
}

func UpdateNodeModulesConfig() {
	logger.SlogInfoLocal("node config load begin")
	redisNodeName := "node:" + config.AppConfig.NodeName
	// 从 Redis 中获取 nodeName 的值
	modulesConfigString, err := redis.RedisClient.HGet(context.Background(), redisNodeName, "modulesConfig")
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("node config load error: %v", err))
		return
	}
	err = utils.WriteContentFile(config.ModulesConfigPath, modulesConfigString)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("node config writing file error: %v", err))
		return
	}
	logger.SlogInfoLocal("node config load end")
}

func Initialize() {
	UpdateGlobalModulesConfig()
	if !config.FirstRun {
		UpdateNodeModulesConfig()
	}
}

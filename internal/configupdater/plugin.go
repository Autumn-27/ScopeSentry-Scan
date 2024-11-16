// configupdater-------------------------------------
// @file      : plugin.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/11/16 20:40
// -------------------------------------------

package configupdater

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"path/filepath"
)

// PluginInfo 存储插件信息的结构体
type PluginInfo struct {
	Module string `bson:"module"`
	Hash   string `bson:"hash"`
	Source string `bson:"source"`
}

func InstallPlugin(id string) {
	var result PluginInfo
	logger.SlogInfoLocal(fmt.Sprintf("update plugin:%v", id))
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("invalid plugin id: %v", err))
		return
	}

	err = mongodb.MongodbClient.FindOne("plugins", bson.M{"_id": objectID}, bson.M{"module": 1, "hash": 1, "source": 1}, &result)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("find plugin error: %v", err))
		return
	}
	plgPath := filepath.Join(global.PluginDir, result.Module, fmt.Sprintf("%v.go", result.Hash))
	err = utils.Tools.WriteContentFile(plgPath, result.Source)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("WriteContentFile plugin error: %v", err))
		return
	}
	logger.SlogInfoLocal(fmt.Sprintf("write plugin end:%v", id))

	logger.SlogInfoLocal(fmt.Sprintf("load plugin:%v", id))

	plugin, err := plugins.LoadCustomPlugin(plgPath, result.Module, result.Hash)
	plugins.GlobalPluginManager.RegisterPlugin(plugin.GetModule(), plugin.GetPluginId(), plugin)
	nodePlgInfokey := fmt.Sprintf("NodePlg:%v", global.AppConfig.NodeName)
	plgInfo := map[string]interface{}{
		plugin.GetPluginId(): 0,
	}
	err = plugin.Install()
	if err != nil {
		plgInfo[plugin.GetPluginId()] = 1
		plgInfoErr := redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
		if plgInfoErr != nil {
			logger.SlogErrorLocal(fmt.Sprintf("send plginfo error 1: %s", plgInfoErr))
		}
		logger.SlogErrorLocal(fmt.Sprintf("plugin install func error: %v", err))
		return
	}
	err = plugin.Check()
	if err != nil {
		plgInfo[plugin.GetPluginId()] = 3
		plgInfoErr := redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
		if plgInfoErr != nil {
			logger.SlogErrorLocal(fmt.Sprintf("send plginfo error 3: %s", plgInfoErr))
		}
		logger.SlogErrorLocal(fmt.Sprintf("plugin check func error: %v", err))
		return
	}
	plgInfo[plugin.GetPluginId()] = 4
	plgInfoErr := redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
	if plgInfoErr != nil {
		logger.SlogErrorLocal(fmt.Sprintf("send plginfo error 4: %s", plgInfoErr))
	}

	logger.SlogInfoLocal(fmt.Sprintf("load plugin end:%v", id))
}

func DeletePlugin(hash string) {
	var result PluginInfo
	logger.SlogInfoLocal(fmt.Sprintf("delete plugin:%v", hash))
	err := mongodb.MongodbClient.FindOne("plugins", bson.M{"hash": hash}, bson.M{"module": 1, "hash": 1}, &result)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("find plugin error: %v", err))
		return
	}
	plg, flag := plugins.GlobalPluginManager.GetPlugin(result.Module, result.Hash)
	if !flag {
		logger.SlogInfoLocal(fmt.Sprintf("plugin %v not found", hash))
		return
	}
	err = plg.UnInstall()
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("plugin UnInstall func error: %v", err))
		return
	}
	plgPath := filepath.Join(global.PluginDir, result.Module, fmt.Sprintf("%v.go", result.Hash))
	utils.Tools.DeleteFile(plgPath)
	logger.SlogInfoLocal(fmt.Sprintf("delete plugin end:%v", hash))
}

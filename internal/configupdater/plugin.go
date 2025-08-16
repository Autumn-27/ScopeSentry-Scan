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
	"strings"
)

// PluginInfo 存储插件信息的结构体

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
	if err != nil {
		logger.SlogError(fmt.Sprintf("load plugin error: %v %v %v %v", plgPath, result.Module, result.Hash, err))
		return
	}
	plugins.GlobalPluginManager.RegisterPlugin(plugin.GetModule(), plugin.GetPluginId(), plugin)
	nodePlgInfokey := fmt.Sprintf("NodePlg:%v", global.AppConfig.NodeName)
	plgInfo := map[string]interface{}{
		plugin.GetPluginId() + "_install": 0,
		plugin.GetPluginId() + "_check":   0,
	}
	err = plugin.Install()
	if err != nil {
		plgInfoErr := redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
		if plgInfoErr != nil {
			logger.SlogErrorLocal(fmt.Sprintf("send plginfo error 1: %s", plgInfoErr))
		}
		logger.SlogErrorLocal(fmt.Sprintf("plugin install func error: %v", err))
		return
	}
	plgInfo[plugin.GetPluginId()+"_install"] = 1
	err = plugin.Check()
	if err != nil {
		plgInfoErr := redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
		if plgInfoErr != nil {
			logger.SlogErrorLocal(fmt.Sprintf("send plginfo error 3: %s", plgInfoErr))
		}
		logger.SlogErrorLocal(fmt.Sprintf("plugin check func error: %v", err))
		return
	}
	plgInfo[plugin.GetPluginId()+"_check"] = 1
	plgInfoErr := redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
	if plgInfoErr != nil {
		logger.SlogErrorLocal(fmt.Sprintf("send plginfo error 4: %s", plgInfoErr))
	}

	logger.SlogInfoLocal(fmt.Sprintf("load plugin end:%v", plugin.GetName()))
}

func DeletePlugin(data string) {
	parts := strings.Split(data, "_")
	hash := parts[0]
	module := parts[1]
	logger.SlogInfoLocal(fmt.Sprintf("delete plugin:%v", hash))

	plg, flag := plugins.GlobalPluginManager.GetPlugin(module, hash)
	if !flag {
		logger.SlogInfoLocal(fmt.Sprintf("plugin %v not found", hash))
		return
	}
	err := plg.UnInstall()
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("plugin UnInstall func error: %v", err))
		return
	}
	plgPath := filepath.Join(global.PluginDir, module, fmt.Sprintf("%v.go", hash))
	utils.Tools.DeleteFile(plgPath)
	nodePlgInfokey := fmt.Sprintf("NodePlg:%v", global.AppConfig.NodeName)
	plgInfoErr := redis.RedisClient.HDel(context.Background(), nodePlgInfokey, hash+"_install", hash+"_check")
	if plgInfoErr != nil {
	}
	logger.SlogInfoLocal(fmt.Sprintf("delete plugin end:%v", hash))
}

func ReInstall(data string) {
	parts := strings.Split(data, "_")
	hash := parts[0]
	module := parts[1]
	plgInfo := map[string]interface{}{
		hash + "_install": 0,
	}
	nodePlgInfokey := fmt.Sprintf("NodePlg:%v", global.AppConfig.NodeName)
	plgInfoErr := redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
	if plgInfoErr != nil {
		logger.SlogErrorLocal(fmt.Sprintf("ReInstall send plginfo error 3: %s", plgInfoErr))
	}
	plg, flag := plugins.GlobalPluginManager.GetPlugin(module, hash)
	if !flag {
		logger.SlogWarnLocal(fmt.Sprintf("module %v hash %v not found", module, hash))
		return
	}
	err := plg.Install()
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("module %v hash %v install error: %v", module, hash, err))
		return
	}
	plgInfo[hash+"_install"] = 1
	plgInfoErr = redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
	if plgInfoErr != nil {
		logger.SlogErrorLocal(fmt.Sprintf("ReInstall send plginfo error 3: %s", plgInfoErr))
	}
	logger.SlogInfoLocal(fmt.Sprintf("plugin reinstall success: %v", data))
}

func ReCheck(data string) {
	parts := strings.Split(data, "_")
	hash := parts[0]
	module := parts[1]
	plgInfo := map[string]interface{}{
		hash + "_check": 0,
	}
	nodePlgInfokey := fmt.Sprintf("NodePlg:%v", global.AppConfig.NodeName)
	plgInfoErr := redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
	if plgInfoErr != nil {
		logger.SlogErrorLocal(fmt.Sprintf("ReCheck send plginfo error 3: %s", plgInfoErr))
	}
	plg, flag := plugins.GlobalPluginManager.GetPlugin(module, hash)
	if !flag {
		logger.SlogWarnLocal(fmt.Sprintf("module %v hash %v not found", module, hash))
		return
	}
	err := plg.Check()
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("module %v hash %v check error: %v", module, hash, err))
		return
	}
	plgInfo[hash+"_check"] = 1
	plgInfoErr = redis.RedisClient.HMSet(context.Background(), nodePlgInfokey, plgInfo)
	if plgInfoErr != nil {
		logger.SlogErrorLocal(fmt.Sprintf("ReCheck send plginfo error 3: %s", plgInfoErr))
	}

	logger.SlogInfoLocal(fmt.Sprintf("plugin recheck success: %v", data))
}

func Uninstall(data string) {
	parts := strings.Split(data, "_")
	hash := parts[0]
	module := parts[1]
	plg, flag := plugins.GlobalPluginManager.GetPlugin(module, hash)
	if !flag {
		logger.SlogWarnLocal(fmt.Sprintf("module %v hash %v not found", module, hash))
		return
	}
	err := plg.UnInstall()
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("module %v hash %v UnInstall error: %v", module, hash, err))
		return
	}
	logger.SlogInfoLocal(fmt.Sprintf("plugin UnInstall success: %v", data))
}

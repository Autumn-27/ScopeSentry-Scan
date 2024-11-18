// configupdater-------------------------------------
// @file      : configupdater.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/7 20:24
// -------------------------------------------

package configupdater

import (
	"context"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	goRedis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"path/filepath"
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
	err = utils.Tools.WriteContentFile(config.ModulesConfigPath, result.Value)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("system config writing file error: %v", err))
		return
	}
	logger.SlogInfoLocal("system config load end")
}

func UpdateNodeModulesConfig() {
	logger.SlogInfoLocal("node config load begin")
	redisNodeName := "node:" + global.AppConfig.NodeName
	// 从 Redis 中获取 nodeName 的值
	modulesConfigString, err := redis.RedisClient.HGet(context.Background(), redisNodeName, "modulesConfig")
	if err != nil {
		if !errors.Is(err, goRedis.Nil) {
			logger.SlogErrorLocal(fmt.Sprintf("node config load error: %v", err))

		}
		return
	}
	err = utils.Tools.WriteContentFile(config.ModulesConfigPath, modulesConfigString)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("node config writing file error: %v", err))
		return
	}
	logger.SlogInfoLocal("node config load end")
}

// UpdateDictionary 更新字典文件
func UpdateDictionary(id string) {
	logger.SlogInfoLocal("UpdateDictionary load begin")
	var results map[string][]byte // 根据实际数据结构定义类型
	var err error
	if id == "all" {
		results, err = mongodb.MongodbClient.FindFilesByPattern(".*")
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("UpdateDictionary load error: %v", err))
			return
		}
	} else {
		results, err = mongodb.MongodbClient.FindFilesByPattern(id)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("UpdateDictionary load error: %v", err))
			return
		}
	}
	for id, content := range results {
		filePath := filepath.Join(global.DictPath, id)
		err = utils.Tools.EnsureFilePathExists(filePath)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("SubDomainDic create file folder error: %v - %v", id, err))
			return
		}
		err = utils.Tools.WriteByteContentFile(filePath, content)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("SubDomainDic writing file error: %v - %v", id, err))
		}
	}
	logger.SlogInfoLocal("UpdateDictionary load end")
}

func UpdateSubfinderApiConfig() {
	logger.SlogInfoLocal("UpdateSubfinderApiConfig load begin")
	//var err error
	var result struct {
		Value string `bson:"value"`
	}
	err := mongodb.MongodbClient.FindOne("config", bson.M{"name": "SubfinderApiConfig"}, bson.M{"_id": 0, "value": 1}, &result)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("UpdateSubfinderApiConfig load error: %v", err))
		return
	}
	subfinderConfigPath := filepath.Join(global.ConfigDir, "subfinderConfig.yaml")
	err = utils.Tools.WriteContentFile(subfinderConfigPath, result.Value)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("Subfinder writing file error: %v", err))
	}
	logger.SlogInfoLocal("UpdateSubfinderApiConfig load end")
	return
}

func UpdateRadConfig() {
	logger.SlogInfoLocal("UpdateRadConfig load begin")
	var result struct {
		Value string `bson:"value"`
	}
	radConfigPath := filepath.Join(global.ExtDir, "rad", "rad_config.yml")
	err := mongodb.MongodbClient.FindOne("config", bson.M{"name": "RadConfig"}, bson.M{"_id": 0, "value": 1}, &result)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("UpdateRadConfig load error: %v", err))
		return
	}
	err = utils.Tools.WriteContentFile(radConfigPath, result.Value)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("Rad writing file error: %v", err))
	}
	logger.SlogInfoLocal("UpdateRadConfig load end")
	return
}

type tmpSensitive struct {
	ID      primitive.ObjectID `bson:"_id"`
	Name    string             `bson:"name"`
	Regular string             `bson:"regular"`
	State   bool               `bson:"state"`
	Color   string             `bson:"color"`
}

func UpdateSensitive() {
	logger.SlogInfoLocal("sens rule load begin")
	var tmpRule []tmpSensitive
	if err := mongodb.MongodbClient.FindAll("SensitiveRule", bson.M{"state": true}, bson.M{"_id": 1, "regular": 1, "state": 1, "color": 1, "name": 1}, &tmpRule); err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("Get Sensitive error: %v", err))
		return
	}
	global.SensitiveRules = []types.SensitiveRule{}
	for _, rule := range tmpRule {
		var r types.SensitiveRule
		r.ID = rule.ID.Hex()
		r.Regular = rule.Regular
		r.State = rule.State
		r.Color = rule.Color
		r.Name = rule.Name
		global.SensitiveRules = append(global.SensitiveRules, r)
	}
	logger.SlogInfoLocal("sens rule load end")
	return
}

type tmpProject struct {
	ID          primitive.ObjectID `bson:"_id"`
	RootDomains []string           `bson:"root_domains"`
	Ignore      string             `bson:"ignore"`
}

func UpdateProject() {
	logger.SlogInfoLocal("project load begin")
	var tmpProjects []tmpProject
	if err := mongodb.MongodbClient.FindAll("project", bson.M{}, bson.M{"_id": 1, "root_domains": 1, "ignore": 1}, &tmpProjects); err != nil {
		return
	}
	global.Projects = []types.Project{}
	for _, tmpProj := range tmpProjects {
		// 创建一个 types.Project 类型的值
		var proj types.Project
		// 将 tmpProject 的值赋给 types.Project 的对应字段
		proj.ID = tmpProj.ID.Hex()
		proj.Target = tmpProj.RootDomains
		ignoreList, regexList, err := utils.Tools.GenerateIgnore(tmpProj.Ignore)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("GenerateIgnore error: %v", err))
			return
		}
		proj.IgnoreList = ignoreList
		proj.IgnoreRegexList = regexList
		global.Projects = append(global.Projects, proj)
	}
	logger.SlogInfoLocal("project load end")
}

type tmpPoc struct {
	ID      primitive.ObjectID `bson:"_id"`
	Hash    string             `bson:"hash"`
	Content string             `bson:"content"`
	Name    string             `bson:"name"`
	Level   string             `bson:"level"`
}

func LoadPoc(id []string) {
	logger.SlogInfoLocal("poc load begin")
	filePath := filepath.Join(global.PocDir, "lock")
	_, err := os.Stat(filePath)
	existFlag := false
	if err == nil {
		existFlag = true // 文件存在
	}
	var tmpPocR []tmpPoc
	// 加载所有poc
	if len(id) == 0 {
		if !existFlag {
			if err := mongodb.MongodbClient.FindAll("PocList", bson.M{}, bson.M{"_id": 1, "content": 1, "name": 1, "level": 1}, &tmpPocR); err != nil {
				logger.SlogError(fmt.Sprintf("Get Poc List error: %s", err))
				return
			}
		}
	} else {
		var objectIDs []primitive.ObjectID
		for _, strID := range id {
			objID, err := primitive.ObjectIDFromHex(strID)
			if err != nil {
				logger.SlogError(fmt.Sprintf("Invalid ObjectID format: %s", strID))
				return
			}
			objectIDs = append(objectIDs, objID)
		}

		// 查询指定的 id
		filter := bson.M{"_id": bson.M{"$in": objectIDs}}
		if err := mongodb.MongodbClient.FindAll("PocList", filter, bson.M{"_id": 1, "content": 1, "name": 1, "level": 1}, &tmpPocR); err != nil {
			logger.SlogError(fmt.Sprintf("Get Poc List by ID error: %s", err))
			return
		}
	}

	if len(tmpPocR) != 0 {
		for _, poc := range tmpPocR {
			id := poc.ID.Hex()
			err := utils.Tools.WriteContentFile(filepath.Join(global.PocDir, string(id)+".yaml"), poc.Content)
			if err != nil {
				logger.SlogError(fmt.Sprintf("Failed to write poc %s: %s", poc.Hash, err))
			}
		}
		if !existFlag {
			err := utils.Tools.WriteContentFile(filePath, "true")
			if err != nil {
				logger.SlogError(fmt.Sprintf("Failed to write poc lock: %s", err))
			}
		}
	}
	logger.SlogInfoLocal("poc load end")
}

type tmpWebFinger struct {
	ID      primitive.ObjectID `bson:"_id"`
	Express []string           `bson:"express"`
	State   bool               `bson:"state"`
	Name    string             `bson:"name"`
}

func UpdateWebFinger() {
	logger.SlogInfoLocal("WebFinger load begin")
	var tmpWebF []tmpWebFinger
	if err := mongodb.MongodbClient.FindAll("FingerprintRules", bson.M{}, bson.M{"_id": 1, "express": 1, "state": 1, "name": 1}, &tmpWebF); err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("WebFinger load error: %v", err))
		return
	}
	global.WebFingers = []types.WebFinger{}
	for _, f := range tmpWebF {
		var wf types.WebFinger
		wf.ID = f.ID.Hex() // 将 ObjectId 转换为字符串
		wf.Express = f.Express
		wf.State = f.State
		wf.Name = f.Name
		global.WebFingers = append(global.WebFingers, wf)
	}
	logger.SlogInfoLocal("WebFinger load end")
}

func UpdateNotification() {
	logger.SlogInfoLocal("Notification load begin")
	if err := mongodb.MongodbClient.FindAll("notification", bson.M{"state": true}, bson.M{"_id": 0, "method": 1, "url": 1, "contentType": 1, "data": 1, "state": 1}, &global.NotificationApi); err != nil {
		logger.SlogError(fmt.Sprintf("UpdateNotification error notification api: %s", err))
		return
	}
	if err := mongodb.MongodbClient.FindOne("config", bson.M{"name": "notification"}, bson.M{"_id": 0, "dirScanNotification": 1, "portScanNotification": 1, "sensitiveNotification": 1, "subdomainTakeoverNotification": 1, "pageMonNotification": 1, "subdomainNotification": 1, "vulNotification": 1, "vulLevel": 1}, &global.NotificationConfig); err != nil {
		logger.SlogError(fmt.Sprintf("UpdateNotification error notification config: %s", err))
		return
	}
	logger.SlogInfoLocal("Notification load end")
}

func LoadPlugin() {
	logger.SlogInfoLocal("load plugin load begin")
	var result []PluginInfo
	err := mongodb.MongodbClient.FindAll("plugins", bson.M{"isSystem": false}, bson.M{"module": 1, "hash": 1, "source": 1}, &result)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("find plugin error: %v", err))
		return
	}
	for _, r := range result {
		plgPath := filepath.Join(global.PluginDir, r.Module, fmt.Sprintf("%v.go", r.Hash))
		err = utils.Tools.WriteContentFile(plgPath, r.Source)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("WriteContentFile plugin %v error: %v", r.Hash, err))
			continue
		}
	}
	logger.SlogInfoLocal("load plugin load end")
}

func Initialize() {
	UpdateSubfinderApiConfig()
	UpdateRadConfig()
	LoadPoc([]string{})
	UpdateDictionary("all")
	UpdateGlobalModulesConfig()
	UpdateNodeModulesConfig()
	UpdateSensitive()
	UpdateProject()
	UpdateWebFinger()
	UpdateNotification()
	LoadPlugin()
}

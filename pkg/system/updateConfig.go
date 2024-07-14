// Package system -----------------------------
// @file      : updateConfig.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/1/10 11:49
// -------------------------------------------
package system

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"path/filepath"
)

func UpdateSubfinderApiConfig() bool {
	//var err error
	var result struct {
		Value string `bson:"value"`
	}
	erro := MongoClient.FindOne("config", bson.M{"name": "SubfinderApiConfig"}, bson.M{"_id": 0, "value": 1}, &result)
	if erro != nil {
		return false
	}
	subfinderConfigPath := filepath.Join(ConfigDir, "subfinderConfig.yaml")
	flag := util.WriteContentFile(subfinderConfigPath, result.Value)
	if !flag {
		fmt.Printf("Write target file error")
		return false
	}
	return true
}

func UpdateDomainDicConfig() bool {
	SlogInfoLocal("domain dict load begin")
	content, erro := MongoClient.FindFile("DomainDic")
	if erro != nil {
		return false
	}
	domainDicConfigPath := filepath.Join(ConfigDir, "domainDic")
	flag := util.WriteContentFile(domainDicConfigPath, string(content))
	if !flag {
		fmt.Printf("Write target file error")
		return false
	}
	SlogInfoLocal("domain dict load end")
	return true
}

func UpdateDirDicConfig() bool {
	//SlogInfoLocal("dir dict load begin")
	//content, erro := MongoClient.FindFile("dirdict")
	//if erro != nil {
	//	return false
	//}
	//dirDicConfigPath := filepath.Join(ConfigDir, "dirdict")
	//flag := util.WriteContentFile(dirDicConfigPath, string(content))
	//if !flag {
	//	fmt.Printf("Write dirdict file error")
	//	return false
	//}
	//SlogInfoLocal("dir dict load end")
	return true
}

func UpdateSystemConfig(flag bool) bool {
	SlogInfoLocal("system config load begin")
	// 定义一个通用的结果结构体
	type ConfigResult struct {
		Value string `bson:"value"`
	}

	// 定义一个通用的函数，用于查询并更新配置
	updateConfig := func(name string, target *string) bool {
		var result ConfigResult
		if err := MongoClient.FindOne("config", bson.M{"name": name}, bson.M{"_id": 0, "value": 1}, &result); err != nil {
			SlogInfoLocal(fmt.Sprintf("system config load err: %v", err))
			return false
		}
		*target = result.Value
		return true
	}

	// 查询并更新具体的配置项
	if !updateConfig("timezone", &AppConfig.System.TimeZoneName) {
		return false
	}
	if !updateConfig("MaxTaskNum", &AppConfig.System.MaxTaskNum) {
		return false
	}
	if !updateConfig("DirscanThread", &AppConfig.System.DirscanThread) {
		return false
	}
	if !updateConfig("PortscanThread", &AppConfig.System.PortscanThread) {
		return false
	}
	if !updateConfig("CrawlerThread", &AppConfig.System.CrawlerThread) {
		return false
	}
	if !updateConfig("UrlThread", &AppConfig.System.UrlThread) {
		return false
	}
	if !updateConfig("UrlMaxNum", &AppConfig.System.UrlMaxNum) {
		return false
	}
	err := WriteYamlConfigToFile(filepath.Join(ConfigDir, "ScopeSentryConfig.yaml"), AppConfig)
	nodeInfo := map[string]interface{}{
		"maxTaskNum":     AppConfig.System.MaxTaskNum,
		"dirscanThread":  AppConfig.System.DirscanThread,
		"portscanThread": AppConfig.System.PortscanThread,
		"crawlerThread":  AppConfig.System.CrawlerThread,
		"urlThread":      AppConfig.System.UrlThread,
		"urlMaxNum":      AppConfig.System.UrlMaxNum,
	}
	errorm := RedisClient.Ping(context.Background())
	if errorm != nil {
		GetRedisClient()
	}
	key := "node:" + AppConfig.System.NodeName
	err = RedisClient.HMSet(context.Background(), key, nodeInfo)
	if err != nil {
		SlogErrorLocal(fmt.Sprintf("Error setting initial values: %s", err))
	}
	if flag {
		CrawlerThreadUpdateFlag <- true
	}
	if err != nil {
		SlogInfoLocal(fmt.Sprintf("system config load err: %v", err))
		return false
	}
	SlogInfoLocal("system config load end")
	return true
}

func UpdateRadConfig() bool {
	SlogInfoLocal("rad config load begin")
	var result struct {
		Value string `bson:"value"`
	}
	radPath := filepath.Join(ExtPath, "rad")
	radConfigPath := filepath.Join(radPath, "rad_config.yml")
	erro := MongoClient.FindOne("config", bson.M{"name": "RadConfig"}, bson.M{"_id": 0, "value": 1}, &result)
	if erro != nil {
		SlogError("Get RadConfig from mongodb error")
		return false
	}
	flag := util.WriteContentFile(radConfigPath, result.Value)
	if !flag {
		fmt.Printf("Write target file error")
		return false
	}
	SlogInfoLocal("dir dict load end")
	return true
}

type tmpSensitive struct {
	ID      primitive.ObjectID `bson:"_id"`
	Name    string             `bson:"name"`
	Regular string             `bson:"regular"`
	State   bool               `bson:"state"`
	Color   string             `bson:"color"`
}

func UpdateSensitive() bool {
	SlogInfoLocal("sens rule load begin")
	var tmpRule []tmpSensitive
	if err := MongoClient.FindAll("SensitiveRule", bson.M{}, bson.M{"_id": 1, "regular": 1, "state": 1, "color": 1, "name": 1}, &tmpRule); err != nil {
		SlogError(fmt.Sprintf("Get Sensitive error: %s", err))
		return false
	}
	SensitiveRules = []types.SensitiveRule{}
	for _, rule := range tmpRule {
		var r types.SensitiveRule
		r.ID = rule.ID.Hex()
		r.Regular = rule.Regular
		r.State = rule.State
		r.Color = rule.Color
		r.Name = rule.Name
		SensitiveRules = append(SensitiveRules, r)
	}
	SlogInfoLocal("sens rule load end")
	return true
}

func UpdateNodeName(name string) {
	oldName := AppConfig.System.NodeName
	AppConfig.System.NodeName = name
	key := "node:" + oldName
	err := RedisClient.HDel(context.Background(), key)
	if err != nil {
		return
	}
	err = WriteYamlConfigToFile(filepath.Join(ConfigDir, "ScopeSentryConfig.yaml"), AppConfig)
	if err != nil {
		return
	}
}

func UpdateNode(flag bool) {
	SlogInfoLocal("node config load begin")
	if !ConfigFileExists {
		return
	}
	RedisNodeName := "node:" + AppConfig.System.NodeName
	// 从 Redis 中获取 nodeName 的值
	maxTaskNum, err := RedisClient.HGet(context.Background(), RedisNodeName, "maxTaskNum")
	if err != nil {
		// 处理错误
		fmt.Println("Failed to get maxTaskNum:", err)
		return
	}

	state, err := RedisClient.HGet(context.Background(), RedisNodeName, "state")
	if err != nil {
		// 处理错误
		fmt.Println("Failed to get state:", err)
		return
	}
	dirscanThread, err := RedisClient.HGet(context.Background(), RedisNodeName, "dirscanThread")
	if err != nil {
		// 处理错误
		fmt.Println("Failed to get dirscanThread:", err)
		return
	}
	portscanThread, err := RedisClient.HGet(context.Background(), RedisNodeName, "portscanThread")
	if err != nil {
		// 处理错误
		fmt.Println("Failed to get dirscanThread:", err)
		return
	}
	crawlerThread, err := RedisClient.HGet(context.Background(), RedisNodeName, "crawlerThread")
	if err != nil {
		// 处理错误
		fmt.Println("Failed to get crawlerThread:", err)
		return
	}
	UrlThread, err := RedisClient.HGet(context.Background(), RedisNodeName, "urlThread")
	if err != nil {
		// 处理错误
		fmt.Println("Failed to get UrlThread:", err)
		return
	}
	UrlMaxNum, err := RedisClient.HGet(context.Background(), RedisNodeName, "urlMaxNum")
	if err != nil {
		// 处理错误
		fmt.Println("Failed to get UrlThread:", err)
		return
	}
	AppConfig.System.MaxTaskNum = maxTaskNum
	AppConfig.System.State = state
	AppConfig.System.DirscanThread = dirscanThread
	AppConfig.System.PortscanThread = portscanThread
	AppConfig.System.CrawlerThread = crawlerThread
	AppConfig.System.UrlThread = UrlThread
	AppConfig.System.UrlMaxNum = UrlMaxNum
	err = WriteYamlConfigToFile(filepath.Join(ConfigDir, "ScopeSentryConfig.yaml"), AppConfig)
	if flag {
		CrawlerThreadUpdateFlag <- true
	}
	if err != nil {
		SlogInfoLocal("node config load end")
		return
	}
}

type tmpProject struct {
	ID          primitive.ObjectID `bson:"_id"`
	RootDomains []string           `bson:"root_domains"`
}
type tmpPortDict struct {
	ID    primitive.ObjectID `bson:"_id"`
	Value string             `bson:"value"`
}

func UpdateProject() {
	SlogInfoLocal("project load begin")
	var tmpProjects []tmpProject
	if err := MongoClient.FindAll("project", bson.M{}, bson.M{"_id": 1, "root_domains": 1}, &tmpProjects); err != nil {
		return
	}
	Projects = make([]types.Project, 0)
	for _, tmpProj := range tmpProjects {
		// 创建一个 types.Project 类型的值
		var proj types.Project
		// 将 tmpProject 的值赋给 types.Project 的对应字段
		proj.ID = tmpProj.ID.Hex()
		proj.Target = tmpProj.RootDomains
		Projects = append(Projects, proj)
	}
	SlogInfoLocal("project load end")
}

func UpdatePort() {
	SlogInfoLocal("port load begin")
	var tmpPort []tmpPortDict
	if err := MongoClient.FindAll("PortDict", bson.M{}, bson.M{"_id": 1, "value": 1}, &tmpPort); err != nil {
		return
	}
	PortDict = []types.PortDict{}
	for _, p := range tmpPort {
		var pt types.PortDict
		pt.ID = p.ID.Hex() // 将 ObjectId 转换为字符串
		pt.Value = p.Value
		PortDict = append(PortDict, pt)
	}
	err := WriteYamlConfigToFile(filepath.Join(ConfigDir, "ports.yaml"), PortDict)
	SlogInfoLocal("port load end")
	if err != nil {
		return
	}
}

type tmpPoc struct {
	ID      primitive.ObjectID `bson:"_id"`
	Hash    string             `bson:"hash"`
	Content string             `bson:"content"`
	Name    string             `bson:"name"`
	Level   int                `bson:"level"`
}

func UpdatePoc(flag bool) {
	if !flag {
		return
	}
	SlogInfoLocal("poc load begin")
	var tmpPocR []tmpPoc
	if err := MongoClient.FindAll("PocList", bson.M{}, bson.M{"_id": 1, "content": 1, "name": 1, "level": 1}, &tmpPocR); err != nil {
		SlogError(fmt.Sprintf("Get Poc List error: %s", err))
		return
	}
	levelDict := map[int]string{1: "unknown", 2: "info", 3: "low", 4: "medium", 5: "high", 6: "critical"}
	for _, p := range tmpPocR {
		PocList[p.ID.Hex()] = types.PocData{
			Name:  p.Name,
			Level: levelDict[p.Level],
		}
	}
	if len(tmpPocR) != 0 {
		path := filepath.Join(ConfigDir, "/poc")
		_ = os.RemoveAll(path)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			SlogError(fmt.Sprintf("Failed to create poc folder:", err))
		}
		for _, poc := range tmpPocR {
			id := poc.ID.Hex()
			err := WriteToFile(filepath.Join(path, string(id)+".yaml"), []byte(poc.Content))
			if err != nil {
				SlogError(fmt.Sprintf("Failed to write poc %s: %s", poc.Hash, err))
			}
		}
	}
	SlogInfoLocal("poc load end")
}

type tmpWebFinger struct {
	ID      primitive.ObjectID `bson:"_id"`
	Express []string           `bson:"express"`
	State   bool               `bson:"state"`
}

func UpdateWebFinger() {
	SlogInfoLocal("WebFinger load begin")
	var tmpWebF []tmpWebFinger
	if err := MongoClient.FindAll("FingerprintRules", bson.M{}, bson.M{"_id": 1, "express": 1, "state": 1}, &tmpWebF); err != nil {
		return
	}
	WebFingers = []types.WebFinger{}
	for _, f := range tmpWebF {
		var wf types.WebFinger
		wf.ID = f.ID.Hex() // 将 ObjectId 转换为字符串
		wf.Express = f.Express
		wf.State = f.State
		WebFingers = append(WebFingers, wf)
	}
	SlogInfoLocal("WebFinger load end")
}
func UpdateNotification() {
	SlogInfoLocal("Notification load begin")
	if err := MongoClient.FindAll("notification", bson.M{"state": true}, bson.M{"_id": 0, "method": 1, "url": 1, "contentType": 1, "data": 1, "state": 1}, &NotificationApi); err != nil {
		SlogError(fmt.Sprintf("UpdateNotification error notification api: %s", err))
		return
	}
	if err := MongoClient.FindOne("config", bson.M{"name": "notification"}, bson.M{"_id": 0, "dirScanNotification": 1, "portScanNotification": 1, "sensitiveNotification": 1, "subdomainTakeoverNotification": 1, "pageMonNotification": 1, "subdomainNotification": 1, "vulNotification": 1}, &NotificationConfig); err != nil {
		SlogError(fmt.Sprintf("UpdateNotification error notification config: %s", err))
		return
	}
	SlogInfoLocal("Notification load end")
}

func UpdateSetUp() {
	UpdateSystemConfig(false)
	UpdateDomainDicConfig()
	UpdateDirDicConfig()
	UpdateSubfinderApiConfig()
	UpdateRadConfig()
	UpdateSensitive()
	UpdateProject()
	UpdatePort()
	UpdatePoc(true)
	UpdateWebFinger()
	UpdateNode(false)
	UpdateNotification()
}

func UpdateSystem() {
	//UpdateSystemFlag <- true
}

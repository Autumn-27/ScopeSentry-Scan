// Package node -----------------------------
// @file      : node.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/2/24 19:03
// -------------------------------------------
package node

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/mongdbClient"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"github.com/shirou/gopsutil/v3/mem"
	"go.mongodb.org/mongo-driver/bson"
	"path/filepath"
	"time"
)

func Register() {
	defer system.RecoverPanic("Node Register")
	nodeName := system.AppConfig.System.NodeName
	if nodeName == "" {
		nodeName = util.GenerateRandomString(6)
	}
	//if system.ConfigFileExists == false {
	//	key = "node:" + nodeName
	//	exists, err := redisClientInstance.Exists(context.Background(), key)
	//	if err != nil {
	//		fmt.Printf("Error checking key existence: %v\n", err)
	//		return
	//	}
	//	if exists {
	//		nodeName = nodeName + "_repeat_" + util.GenerateRandomString(6)
	//	}
	//}
	key := "node:" + nodeName
	if nodeName != system.AppConfig.System.NodeName {
		system.AppConfig.System.NodeName = nodeName
		err := system.WriteYamlConfigToFile(filepath.Join(system.ConfigDir, "ScopeSentryConfig.yaml"), system.AppConfig)
		if err != nil {
			return
		}
	}
	firstRegister := true
	ticker := time.Tick(20 * time.Second)
	for {
		if firstRegister {
			memInfo, _ := mem.VirtualMemory()
			system.AppConfig.System.Running = 0
			system.AppConfig.System.Finished = 0
			nodeInfo := map[string]interface{}{
				"updateTime":     system.GetTimeNow(),
				"running":        system.AppConfig.System.Running,
				"finished":       system.AppConfig.System.Finished,
				"cpuNum":         0,
				"maxTaskNum":     system.AppConfig.System.MaxTaskNum,
				"dirscanThread":  system.AppConfig.System.DirscanThread,
				"portscanThread": system.AppConfig.System.PortscanThread,
				"crawlerThread":  system.AppConfig.System.CrawlerThread,
				"urlThread":      system.AppConfig.System.UrlThread,
				"urlMaxNum":      system.AppConfig.System.UrlMaxNum,
				"TotleMem":       float64(memInfo.Total) / 1024 / 1024,
				"memNum":         0,
				"state":          1, //1运行中 2暂停 3未连接
				"version":        system.VERSION,
			}
			err := system.RedisClient.HMSet(context.Background(), key, nodeInfo)
			if err != nil {
				system.SlogErrorLocal(fmt.Sprintf("Error setting initial values: %s", err))
				return
			}
			system.SlogInfo(fmt.Sprintf("Register Success:%v - version %v", nodeName, system.VERSION))
			firstRegister = false
		} else {
			cpuNum, memNum := util.GetSystemUsage()
			run, fin := system.GetRunFin()
			nodeInfo := map[string]interface{}{
				"updateTime": system.GetTimeNow(),
				"cpuNum":     cpuNum,
				"memNum":     memNum,
				"maxTaskNum": system.AppConfig.System.MaxTaskNum,
				"running":    run,
				"finished":   fin,
				"state":      1,
				"version":    system.VERSION,
			}
			errorm := system.RedisClient.Ping(context.Background())
			if errorm != nil {
				system.GetRedisClient()
			}
			err := system.RedisClient.HMSet(context.Background(), key, nodeInfo)
			if err != nil {
				system.SlogErrorLocal(fmt.Sprintf("Error setting initial values: %s", err))
				continue
			}
		}
		<-ticker
	}
}

func Register2(mongoClient *mongdbClient.MongoDBClient) {
	nodeName := system.AppConfig.System.NodeName
	if nodeName == "" {
		nodeName = util.GenerateRandomString(6)
	}

	filter := bson.M{"node_name": nodeName}
	var result bson.M
	err := mongoClient.FindOne("node", filter, nil, &result)
	if err == nil {
		// Key exists, generate a unique node name
		nodeName = nodeName + "_repeat_" + util.GenerateRandomString(6)
		filter = bson.M{"node_name": nodeName}
	}

	firstRegister := true
	for {
		if firstRegister {
			nodeInfo := bson.M{
				"node_name":  nodeName,
				"updateTime": system.GetTimeNow(),
				"running":    0,
				"finished":   0,
				"cpuNum":     0,
				"memNum":     0,
				"state":      1, //1运行中 2暂停 3未连接
			}
			_, err := mongoClient.Upsert("node", filter, bson.M{"$set": nodeInfo})
			if err != nil {
				fmt.Printf("Error upserting initial values: %v\n", err)
				return
			}
			firstRegister = false
		} else {
			cpuNum, memNum := util.GetSystemUsage()
			update := bson.M{
				"$set": bson.M{
					"last_time": system.GetTimeNow(),
					"cpuNum":    cpuNum,
					"memNum":    memNum,
				},
			}
			_, err := mongoClient.Update("node", filter, update)
			if err != nil {
				fmt.Printf("Error updating node info: %v\n", err)
				return
			}
		}
		time.Sleep(30 * time.Second)
	}
}

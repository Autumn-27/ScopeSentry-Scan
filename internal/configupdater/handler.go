// configupdater-------------------------------------
// @file      : handler.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/11/2 15:11
// -------------------------------------------

package configupdater

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handler"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"os"
	"strings"
	"time"
)

type Message struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func handleRefreshConfigMessage(jsonData Message) {
	switch jsonData.Type {
	case "system":
		UpdateSystemConfig(jsonData.Content)
	case "dictionary":
		Updatedictionary(jsonData.Content)
	case "subfinder":
		UpdateSubfinderApiConfig()
	case "rad":
		UpdateRadConfig()
	case "sensitive":
		UpdateSensitive()
	case "nodeConfig":
		UpdateNode(jsonData.Content)
	case "project":
		UpdateProject()
	case "notification":
		UpdateNotification()
	case "stop_task":
		handler.TaskHandle.StopTask(jsonData.Content)
	case "delete_task":
		handler.TaskHandle.DeleteTask(jsonData.Content)
	case "install_plugin":
		InstallPlugin(jsonData.Content)
	case "delete_plugin":
		DeletePlugin(jsonData.Content)
	case "re_install_plugin":
		ReInstall(jsonData.Content)
	case "re_check_plugin":
		ReCheck(jsonData.Content)
	case "uninstall_plugin":
		Uninstall(jsonData.Content)
	case "UpdateSystem":
		SystemUpdate(jsonData.Content)
	case "restart":
		os.Exit(0)
	}
}

func RefreshConfig() {
	ticker := time.Tick(3 * time.Second)
	for {
		<-ticker
		RefreshConfigNodeName := "refresh_config:" + global.AppConfig.NodeName
		exists, err := redis.RedisClient.Exists(context.Background(), RefreshConfigNodeName)
		if err != nil {
			logger.SlogError(fmt.Sprintf("RefreshConfig Error: %v", err))
			continue
		}
		if !exists {
			continue
		}
		listLength, err := redis.RedisClient.LLen(context.Background(), RefreshConfigNodeName)
		if err != nil {
			logger.SlogError(fmt.Sprintf("RefreshConfig Error getting list length: %v", err))
			continue
		}
		if listLength == 0 {
			continue
		}
		msgs, err := redis.RedisClient.BatchGetAndDelete(context.Background(), RefreshConfigNodeName, listLength)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("RefreshConfig Error batch get: %v", err))
			continue
		}

		var pocAddIDs, pocDeleteIDs []string
		var needUpdateFinger bool

		for _, msg := range msgs {
			logger.SlogInfo(fmt.Sprintf("recv RefreshConfig: %s", msg))
			jsonData := Message{}
			err2 := json.Unmarshal([]byte(msg), &jsonData)
			if err2 != nil {
				logger.SlogErrorLocal(fmt.Sprintf("Task parse error: %v", err))
				continue
			}
			if jsonData.Name != "all" && jsonData.Name != global.AppConfig.NodeName {
				continue
			}
			switch jsonData.Type {
			case "poc":
				addIDs, deleteIDs := parsePocContent(jsonData.Content)
				pocAddIDs = append(pocAddIDs, addIDs...)
				pocDeleteIDs = append(pocDeleteIDs, deleteIDs...)
			case "finger":
				needUpdateFinger = true
			default:
				handleRefreshConfigMessage(jsonData)
			}
		}

		if len(pocDeleteIDs) > 0 {
			UpdatePoc("delete:" + strings.Join(uniqueStrings(pocDeleteIDs), ","))
		}
		if len(pocAddIDs) > 0 {
			UpdatePoc("add:" + strings.Join(uniqueStrings(pocAddIDs), ","))
		}
		if needUpdateFinger {
			UpdateWebFinger()
		}
	}
}

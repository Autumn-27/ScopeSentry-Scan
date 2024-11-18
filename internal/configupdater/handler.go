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
	"time"
)

type Message struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
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
		if exists {
			msg, err := redis.RedisClient.PopFirstFromList(context.Background(), RefreshConfigNodeName)
			logger.SlogInfo(fmt.Sprintf("recv RefreshConfig: %s", msg))
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("RefreshConfig Error 2:%v", err))
				continue
			}
			jsonData := Message{}
			err2 := json.Unmarshal([]byte(msg), &jsonData)
			if err2 != nil {
				logger.SlogErrorLocal(fmt.Sprintf("Task parse error: %v", err))
				continue
			}
			if jsonData.Name == "all" || jsonData.Name == global.AppConfig.NodeName {
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
				case "poc":
					UpdatePoc(jsonData.Content)
				case "finger":
					UpdateWebFinger()
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
				}
			}
		}
	}
}

// results-------------------------------------
// @file      : handler.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/19 19:10
// -------------------------------------------

package results

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/notification"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type handler struct {
}

var Handler *handler

func InitializeHandler() {
	Handler = &handler{}
}

func (h *handler) GetAssetProject(host string) string {
	for _, p := range global.Projects {
		for _, t := range p.Target {
			if host == t {
				return p.ID
			}
		}
	}
	return ""
}

func (h *handler) Subdomain(result *types.SubdomainResult) {
	var interfaceSlice interface{}
	rootDomain, err := utils.Tools.GetRootDomain(result.Host)
	if err != nil {
		logger.SlogInfoLocal(fmt.Sprintf("%v GetRootDomain error: %v", result.Host, err))
	}
	result.RootDomain = rootDomain
	result.Project = h.GetAssetProject(rootDomain)
	interfaceSlice = &result
	if global.NotificationConfig.SubdomainScan {
		NotificationMsg := fmt.Sprintf("%v - %v\n", result.Host, result.IP)
		notification.NotificationQueues["SubdomainScan"].Queue <- NotificationMsg
	}
	ResultQueues["SubdomainScan"].Queue <- interfaceSlice
}

func (h *handler) SubdomainTakeover(result *types.SubTakeResult) {
	var interfaceSlice interface{}
	rootDomain, err := utils.Tools.GetRootDomain(result.Input)
	if err != nil {
		logger.SlogInfoLocal(fmt.Sprintf("%v GetRootDomain error: %v", result.Input, err))
	}
	result.RootDomain = rootDomain
	result.Project = h.GetAssetProject(rootDomain)
	interfaceSlice = &result
	if global.NotificationConfig.SubdomainTakeoverNotification {
		NotificationMsg := fmt.Sprintf("Subdomain Takeover:\n%v - %v\n", result.Input, result.Cname)
		notification.NotificationQueues["SubdomainSecurity"].Queue <- NotificationMsg
	}
	ResultQueues["SubdomainScan"].Queue <- interfaceSlice
}

func (h *handler) AssetChangeLog(result *types.AssetChangeLog) {
	var interfaceSlice interface{}
	interfaceSlice = &result
	ResultQueues["AssetChangeLog"].Queue <- interfaceSlice
}

func (h *handler) AssetUpdate(id string, updateData interface{}) {
	// 资产比较少，并且需要实时获取历史记录，所以不采用多条插入的方式
	objectID, err := primitive.ObjectIDFromHex(id) // 将字符串 ID 转换为 ObjectID
	if err != nil {
		logger.SlogError(fmt.Sprintf("AssetUpdate %v ObjectIDFromHex error:%v", id, err))
		return
	}
	selector := bson.M{"_id": objectID} // 创建选择器，匹配指定的 ObjectID
	_, err = mongodb.MongodbClient.Update("asset", selector, bson.M{"$set": updateData})
	if err != nil {
		logger.SlogError(fmt.Sprintf("AssetUpdate %v error:%v", id, err))
	} // 使用 $set 更新字段
}

func (h *handler) AssetOtherInsert(result *types.AssetOther) {
	// 资产比较少，并且需要实时获取历史记录，所以不采用多条插入的方式
	var interfaceSlice interface{}
	rootDomain, err := utils.Tools.GetRootDomain(result.Host)
	if err != nil {
		logger.SlogInfoLocal(fmt.Sprintf("%v GetRootDomain error: %v", result.Host, err))
	}
	result.RootDomain = rootDomain
	result.Project = h.GetAssetProject(rootDomain)
	interfaceSlice = &result
	_, err = mongodb.MongodbClient.InsertOne("asset", interfaceSlice)
	if err != nil {
		logger.SlogError(fmt.Sprintf("AssetOtherInsert error:%v", err))
		return
	}
}

func (h *handler) AssetHttpInsert(result *types.AssetHttp) {
	// 资产比较少，并且需要实时获取历史记录，所以不采用多条插入的方式
	var interfaceSlice interface{}
	rootDomain, err := utils.Tools.GetRootDomain(result.Host)
	if err != nil {
		logger.SlogInfoLocal(fmt.Sprintf("%v GetRootDomain error: %v", result.Host, err))
	}
	result.RootDomain = rootDomain
	result.Project = h.GetAssetProject(rootDomain)
	interfaceSlice = &result
	_, err = mongodb.MongodbClient.InsertOne("asset", interfaceSlice)
	if err != nil {
		logger.SlogError(fmt.Sprintf("AssetOtherInsert error:%v", err))
		return
	}
}

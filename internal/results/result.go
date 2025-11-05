// results-------------------------------------
// @file      : result.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/17 21:12
// -------------------------------------------

package results

import (
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type result struct {
}

var Results *result

func InitializeResults() {
	Results = &result{}
}

//func (r *result) Subdoamin(result *[]interface{}) bool {
//	_, err := mongodb.MongodbClient.InsertMany("subdomain", *result)
//	if err != nil {
//		logger.SlogError(fmt.Sprintf("SubdoaminResult error: %s", err))
//		return false
//	}
//	return true
//}

func (r *result) Insert(name string, result *[]interface{}) bool {
	_, err := mongodb.MongodbClient.InsertMany(name, *result)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			if name == "PageMonitoring" || name == "PageMonitoringBody" {
				for _, wErr := range writeException.WriteErrors {
					logger.SlogWarnLocal(fmt.Sprintf("插入失败的文档: %v, 错误: %v\n", wErr))
				}
			}
		}
		logger.SlogWarnLocal(fmt.Sprintf("insert %v error: %s", name, err))
		return false
	}
	return true
}

func (r *result) Update(result *[]interface{}, name string) bool {
	// 将 *[]interface{} 转换为 []types.BulkUpdateOperation
	var operations []mongo.WriteModel

	// 遍历传入的 result，逐一进行类型断言并构建 WriteModel
	for _, item := range *result {
		// 类型断言将 item 转换为 types.BulkUpdateOperation
		op, ok := item.(types.BulkUpdateOperation)
		if !ok {
			fmt.Println("Type assertion failed: item is not of type types.BulkUpdateOperation")
			return false
		}

		// 将 BulkUpdateOperation 转换为 mongo.UpdateOneModel 并添加到 operations 中
		updateModel := mongo.NewUpdateOneModel().
			SetFilter(op.Selector).
			SetUpdate(op.Update).
			SetUpsert(true) // 设置 Upsert 为 true，如果没有匹配文档，则插入新的文档

		operations = append(operations, updateModel)
	}
	// 调用 BulkWrite 批量写入或更新操作
	_, err := mongodb.MongodbClient.BulkWrite(name, operations)
	if err != nil {
		fmt.Println("Error during bulk write:", err)
		return false
	}
	return true
}

func (r *result) UpdateNow(op types.BulkUpdateOperation, name string) bool {
	// 将 *[]interface{} 转换为 []types.BulkUpdateOperation
	var operations []mongo.WriteModel

	// 将 BulkUpdateOperation 转换为 mongo.UpdateOneModel 并添加到 operations 中
	updateModel := mongo.NewUpdateOneModel().
		SetFilter(op.Selector).
		SetUpdate(op.Update).
		SetUpsert(true) // 设置 Upsert 为 true，如果没有匹配文档，则插入新的文档

	operations = append(operations, updateModel)
	// 调用 BulkWrite 批量写入或更新操作
	_, err := mongodb.MongodbClient.BulkWrite(name, operations)
	if err != nil {
		fmt.Println("Error during bulk write:", err)
		return false
	}
	return true
}

func (r *result) InsertVulnerabilityScan(result *[]interface{}) {
	var newVulnResults []interface{}
	var vulDetails []interface{}

	for _, v := range *result {
		var vulnResult *types.VulnResult

		// 第一种：*types.VulnResult
		if vr, ok := v.(*types.VulnResult); ok {
			vulnResult = vr
		}

		// 第二种：**types.VulnResult
		if vrPtr, ok := v.(**types.VulnResult); ok && vrPtr != nil {
			vulnResult = *vrPtr
		}

		if vulnResult == nil {
			fmt.Println("类型断言失败:", reflect.TypeOf(v))
			continue
		}

		// 生成唯一 Hash
		vulnResult.Hash = utils.Tools.GenerateRandomString(16)

		// 保存详细信息
		vulDetails = append(vulDetails, types.VulnerabilityDetail{
			Hash:     vulnResult.Hash,
			Request:  vulnResult.Request,
			Response: vulnResult.Response,
		})

		// 清空请求与响应
		vulnResult.Request = ""
		vulnResult.Response = ""

		// 放回新的结果集
		newVulnResults = append(newVulnResults, vulnResult)
	}

	// 写入数据库
	r.Insert("vulnerability", &newVulnResults)
	r.Insert("vulnerabilityDetail", &vulDetails)
}

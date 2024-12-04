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
	"go.mongodb.org/mongo-driver/mongo"
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

func (r *result) Update(result *[]interface{}) bool {
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
	_, err := mongodb.MongodbClient.BulkWrite("SensitiveBody", operations)
	if err != nil {
		fmt.Println("Error during bulk write:", err)
		return false
	}
	return true
}

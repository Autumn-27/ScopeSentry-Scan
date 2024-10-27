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
					logger.SlogErrorLocal(fmt.Sprintf("插入失败的文档: %v, 错误: %v\n", wErr))
				}
			}
		}
		logger.SlogError(fmt.Sprintf("insert %v error: %s", name, err))
		return false
	}
	return true
}

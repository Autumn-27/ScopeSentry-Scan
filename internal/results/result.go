// results-------------------------------------
// @file      : result.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/17 21:12
// -------------------------------------------

package results

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
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
		logger.SlogError(fmt.Sprintf("insert %v error: %s", name, err))
		return false
	}
	return true
}

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

func SubdoaminResultHandler(result *[]interface{}) bool {
	_, err := mongodb.MongodbClient.InsertMany("subdomain", *result)
	if err != nil {
		logger.SlogError(fmt.Sprintf("SubdoaminResult error: %s", err))
		return false
	}
	return true
}

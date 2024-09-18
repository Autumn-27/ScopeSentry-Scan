// results-------------------------------------
// @file      : result.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/17 21:12
// -------------------------------------------

package results

import (
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
)

func SubdoaminResultHandler(result *types.SubdomainResult) bool {

	//util.GetAssetOwner(result)
	//NotificationMsg := "SubdomainScan Result"
	//var interfaceSlice []interface{}
	//for _, r := range result {
	//	project := GetAssetOwner(r.Host)
	//	r.Project = project
	//	r.TaskId = taskId
	//	interfaceSlice = append(interfaceSlice, r)
	//	NotificationMsg += fmt.Sprintf("%v - %v\n", r.Host, r.IP)
	//}
	//if len(interfaceSlice) != 0 {
	//	//if system.NotificationConfig.SubdomainNotification {
	//	//	go system.SendNotification(NotificationMsg)
	//	//}
	//	errorm := system.MongoClient.Ping()
	//	if errorm != nil {
	//		system.GetMongbClient()
	//	}
	//	_, err := system.MongoClient.InsertMany("subdomain", interfaceSlice)
	//	if err != nil {
	//		system.SlogError(fmt.Sprintf("SubdoaminResult error: %s", err))
	//		return false
	//	}
	//}
	return true
}

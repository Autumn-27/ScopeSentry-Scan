// scanResult-------------------------------------
// @file      : result.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/9 21:04
// -------------------------------------------

package scanResult

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/httpxMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/sensitiveMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"net/url"
	"sync"
)

func SubdoaminResult(result []types.SubdomainResult, taskId string) bool {

	//util.GetAssetOwner(result)
	NotificationMsg := "SubdomainScan Result"
	var interfaceSlice []interface{}
	for _, r := range result {
		project := GetAssetOwner(r.Host)
		r.Project = project
		r.TaskId = taskId
		interfaceSlice = append(interfaceSlice, r)
		NotificationMsg += fmt.Sprintf("%v - %v\n", r.Host, r.IP)
	}
	if len(interfaceSlice) != 0 {
		if system.NotificationConfig.SubdomainNotification {
			go system.SendNotification(NotificationMsg)
		}
		errorm := system.MongoClient.Ping()
		if errorm != nil {
			system.GetMongbClient()
		}
		_, err := system.MongoClient.InsertMany("subdomain", interfaceSlice)
		if err != nil {
			system.SlogError(fmt.Sprintf("SubdoaminResult error: %s", err))
			return false
		}
	}
	return true
}

func ParseUrlToDomain(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		system.SlogError(fmt.Sprintf("解析URL时出错: %s", err))
		return urlStr
	}
	domain := u.Hostname()
	return domain
}
func AssetResult(httpRestlts []types.AssetHttp, assetOthers []types.AssetOther, taskId string) bool {
	if len(system.WebFingers) == 0 {
		system.UpdateWebFinger()
	}
	var interfaceSlice []interface{}
	for _, h := range httpRestlts {
		project := GetAssetOwner(h.URL)
		h.Project = project
		h.Domain = ParseUrlToDomain(h.URL)
		h.TaskId = taskId
		hs := WebFingerScan(h)
		interfaceSlice = append(interfaceSlice, hs)
	}
	for _, a := range assetOthers {
		project := GetAssetOwner(a.Host)
		a.Project = project
		a.TaskId = taskId
		interfaceSlice = append(interfaceSlice, a)
	}
	if len(interfaceSlice) != 0 {
		errorm := system.MongoClient.Ping()
		if errorm != nil {
			system.GetMongbClient()
		}
		_, err := system.MongoClient.InsertMany("asset", interfaceSlice)
		if err != nil {
			myLog := system.CustomLog{
				Status: "Error",
				Msg:    fmt.Sprintf("SubdoaminResult error: %s", err),
			}
			system.PrintLog(myLog)
			return false
		}
	}
	return true
}

func SubTakerResult(result []types.SubTakeResult, taskId string) {
	var interfaceSlice []interface{}
	NotificationMsg := "SubTaker Result:\n"
	for _, r := range result {
		project := GetAssetOwner(r.Input)
		r.Project = project
		r.TaskId = taskId
		interfaceSlice = append(interfaceSlice, r)
		NotificationMsg += fmt.Sprintf("%v - %v\n resp: %v\n", r.Input, r.Cname, r.Response)
	}
	if len(interfaceSlice) != 0 {
		if system.NotificationConfig.SubdomainTakeoverNotification {
			go system.SendNotification(NotificationMsg)
		}
		errorm := system.MongoClient.Ping()
		if errorm != nil {
			system.GetMongbClient()
		}
		_, err := system.MongoClient.InsertMany("SubdoaminTakerResult", interfaceSlice)
		if err != nil {
			myLog := system.CustomLog{
				Status: "Error",
				Msg:    fmt.Sprintf("SubdoaminTakerResult error: %s", err),
			}
			system.PrintLog(myLog)
		}
	}
}

func DirResult(result []types.DirResult) {
	var interfaceSlice []interface{}
	for _, r := range result {
		project := GetAssetOwner(r.Url)
		r.Project = project
		interfaceSlice = append(interfaceSlice, r)
	}
	if len(interfaceSlice) != 0 {
		errorm := system.MongoClient.Ping()
		if errorm != nil {
			system.GetMongbClient()
		}
		_, err := system.MongoClient.InsertMany("DirScanResult", interfaceSlice)
		if err != nil {
			myLog := system.CustomLog{
				Status: "Error",
				Msg:    fmt.Sprintf("dirScanResult error: %s", err),
			}
			system.PrintLog(myLog)
		}
	}

}

func SensitiveResult(result []types.SensitiveResult, taskId string) {
	var interfaceSlice []interface{}
	for _, r := range result {
		project := GetAssetOwner(r.Url)
		r.Project = project
		r.TaskId = taskId
		interfaceSlice = append(interfaceSlice, r)
	}
	if len(interfaceSlice) != 0 {
		errorm := system.MongoClient.Ping()
		if errorm != nil {
			system.GetMongbClient()
		}
		_, err := system.MongoClient.InsertMany("SensitiveResult", interfaceSlice)
		if err != nil {
			system.SlogError(fmt.Sprintf("SensitiveResult error: %s", err))
		}
	}
}

func UrlResult(result []types.UrlResult, taskId string, secretFlag bool) {
	var interfaceSlice []interface{}
	var wg sync.WaitGroup
	for _, r := range result {
		StatusCode, ContentLength, RespBody, err := httpxMode.HttpSurvival(r.Output)
		if err != nil {
			system.SlogDebugLocal(fmt.Sprintf("Url %v HttpSurvival eror: %v", r.Output, err))
			continue
		}
		if secretFlag {
			if !system.IsMatchingFilter(system.DisallowedURLFilters, []byte(r.Output)) {
				if secretFlag {
					wg.Add(1)
					go func(url string, msg string) {
						defer func() {
							wg.Done()
						}()
						respMd5 := util.CalculateMD5(msg)
						if !SensRedisDeduplication(respMd5, taskId) {
							sensitiveMode.Scan(url, msg, respMd5, taskId)
						}
					}(r.Output, RespBody)
				}
			}
		}
		r.Status = StatusCode
		r.Length = ContentLength
		project := GetAssetOwner(r.Input)
		r.Project = project
		r.TaskId = taskId
		interfaceSlice = append(interfaceSlice, r)
	}
	if len(interfaceSlice) != 0 {
		errorm := system.MongoClient.Ping()
		if errorm != nil {
			system.GetMongbClient()
		}
		_, err := system.MongoClient.InsertMany("UrlScan", interfaceSlice)
		if err != nil {
			myLog := system.CustomLog{
				Status: "Error",
				Msg:    fmt.Sprintf("UrlScan error: %s", err),
			}
			system.PrintLog(myLog)
		}
	}
	wg.Wait()
}

func CrawlerResult(result []types.CrawlerResult, taskId string) {
	var interfaceSlice []interface{}
	for _, r := range result {
		project := GetAssetOwner(r.Url)
		r.Project = project
		r.TaskId = taskId
		interfaceSlice = append(interfaceSlice, r)
	}
	if len(interfaceSlice) != 0 {
		errorm := system.MongoClient.Ping()
		if errorm != nil {
			system.GetMongbClient()
		}
		_, err := system.MongoClient.InsertMany("crawler", interfaceSlice)
		if err != nil {
			myLog := system.CustomLog{
				Status: "Error",
				Msg:    fmt.Sprintf("crawler error: %s", err),
			}
			system.PrintLog(myLog)
		}
	}
}

func VulnResult(result []types.VulnResult) {
	var interfaceSlice []interface{}
	for _, r := range result {
		project := ""
		if r.Url != "" {
			project = GetAssetOwner(r.Url)
		} else {
			project = GetAssetOwner(r.Matched)
		}
		r.Project = project
		interfaceSlice = append(interfaceSlice, r)
		system.SlogInfo(fmt.Sprintf("Found vulnerable: %v - %v", r.Url, r.VulName))
	}
	if len(interfaceSlice) != 0 {
		errorm := system.MongoClient.Ping()
		if errorm != nil {
			system.GetMongbClient()
		}
		_, err := system.MongoClient.InsertMany("vulnerability", interfaceSlice)
		if err != nil {
			myLog := system.CustomLog{
				Status: "Error",
				Msg:    fmt.Sprintf("vulnerability error: %s", err),
			}
			system.PrintLog(myLog)
		}
	}
}

type TmpPageMonitResultWithoutIdkey struct {
	Url     string   `bson:"url"`
	Content []string `bson:"content"`
	Hash    []string `bson:"hash"`
	Diff    []string `bson:"diff"`
	State   int      `bson:"state"`
	Project string   `bson:"project"`
	Time    string   `bson:"time"`
	TaskId  string   `bson:"taskId"`
}

func PageMonitoringInitResult(result []types.TmpPageMonitResult, taskId string) {
	var interfaceSlice []interface{}
	for _, r := range result {
		flag, tmp := PageMonitoringMongoDbDeduplication(r.Url)
		nHash := util.CalculateMD5(r.Content)
		if flag {
			if r.Content == "" {
				continue
			}
			if len(tmp.Hash) == 0 || len(tmp.Content) == 0 {
				tmp.Content = append(tmp.Content, r.Content)
				tmp.Hash = append(tmp.Hash, nHash)
				updateFields := bson.M{
					"content": tmp.Content,
					"hash":    tmp.Hash,
					"diff":    tmp.Diff,
					"time":    "",
					"project": GetAssetOwner(r.Url),
				}
				errorm := system.MongoClient.Ping()
				if errorm != nil {
					system.GetMongbClient()
				}
				_, err := system.MongoClient.Update("PageMonitoring", bson.M{"_id": tmp.ID}, bson.M{"$set": updateFields})
				if err != nil {
					system.SlogError(fmt.Sprintf("PageMonitoringInitResult Update Error%s", err))
				}
				continue
			}
			tmpHash := tmp.Hash[len(tmp.Hash)-1]
			if nHash != tmpHash {
				diff := DiffContent(tmp.Content[len(tmp.Content)-1], r.Content)
				tmp.Content = append(tmp.Content, r.Content)
				if len(tmp.Content) > 2 {
					tmp.Content = tmp.Content[len(tmp.Content)-2:]
				}
				tmp.Hash = append(tmp.Hash, nHash)
				tmp.Diff = append(tmp.Diff, diff)
				updateFields := bson.M{
					"content": tmp.Content,
					"hash":    tmp.Hash,
					"diff":    tmp.Diff,
					"time":    system.GetTimeNow(),
				}
				errorm := system.MongoClient.Ping()
				if errorm != nil {
					system.GetMongbClient()
				}
				_, err := system.MongoClient.Update("PageMonitoring", bson.M{"_id": tmp.ID}, bson.M{"$set": updateFields})
				if err != nil {
					system.SlogError(fmt.Sprintf("PageMonitoringInitResult DiffContent Error%s", err))
				}
				if system.NotificationConfig.PageMonNotification {
					NotificationMsg := fmt.Sprintf("PageMonitoring Result:\n%v \n Diff:%v", tmp.Url, tmp.Diff)
					go system.SendNotification(NotificationMsg)
				}
			}
		} else {
			if r.Url != "" {
				tmpR := TmpPageMonitResultWithoutIdkey{
					Url:     r.Url,
					Project: GetAssetOwner(r.Url),
					Content: []string{},
					Hash:    []string{},
					Diff:    []string{},
					State:   1,
					TaskId:  taskId,
				}
				interfaceSlice = append(interfaceSlice, tmpR)
			}
		}
	}
	if len(interfaceSlice) != 0 {
		errorm := system.MongoClient.Ping()
		if errorm != nil {
			system.GetMongbClient()
		}
		_, err := system.MongoClient.InsertMany("PageMonitoring", interfaceSlice)
		if err != nil {
			myLog := system.CustomLog{
				Status: "Error",
				Msg:    fmt.Sprintf("PageMonitoring error: %s", err),
			}
			system.PrintLog(myLog)
		}
	}
}

func TaskEnds(target string, taskId string) {
	errorm := system.RedisClient.Ping(context.Background())
	if errorm != nil {
		system.GetRedisClient()
	}
	key := "TaskInfo:time:" + taskId
	err := system.RedisClient.Set(context.Background(), key, system.GetTimeNow())
	if err != nil {
		myLog := system.CustomLog{
			Status: "Error",
			Msg:    fmt.Sprintf("TaskEnds push redis error: %s", err),
		}
		system.PrintLog(myLog)
		return
	}
	key = "TaskInfo:tmp:" + taskId
	_, err = system.RedisClient.AddToList(context.Background(), key, target)
	if err != nil {
		myLog := system.CustomLog{
			Status: "Error",
			Msg:    fmt.Sprintf("TaskEnds push redis error: %s", err),
		}
		system.PrintLog(myLog)
		return
	}

}

func ProgressStart(typ string, target string, taskId string) {
	//system.SlogDebugLocal("ProgressStart begin")
	key := "TaskInfo:progress:" + taskId + ":" + target
	ty := typ + "_start"
	ProgressInfo := map[string]interface{}{
		ty: system.GetTimeNow(),
	}
	errorm := system.RedisClient.Ping(context.Background())
	if errorm != nil {
		system.GetRedisClient()
	}
	err := system.RedisClient.HMSet(context.Background(), key, ProgressInfo)
	if err != nil {
		system.SlogError(fmt.Sprintf("ProgressStart redis error: %s", err))
		return
	}
	//system.SlogDebugLocal("ProgressStart end")
}

func ProgressEnd(typ string, target string, taskId string) {

	key := "TaskInfo:progress:" + taskId + ":" + target
	ty := typ + "_end"
	ProgressInfo := map[string]interface{}{
		ty: system.GetTimeNow(),
	}
	errorm := system.RedisClient.Ping(context.Background())
	if errorm != nil {
		system.GetRedisClient()
	}
	err := system.RedisClient.HMSet(context.Background(), key, ProgressInfo)
	if err != nil {
		myLog := system.CustomLog{
			Status: "Error",
			Msg:    fmt.Sprintf("ProgressEnd redis error: %s", err),
		}
		system.PrintLog(myLog)
		return
	}
}

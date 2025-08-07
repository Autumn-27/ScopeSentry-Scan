// runner-------------------------------------
// @file      : pagemonitoring.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/11/5 20:51
// -------------------------------------------

package runner

import (
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"github.com/panjf2000/ants/v2"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"sync"
	"time"
)

func PageMonitoringRunner(targets []string) {
	pool, err := ants.NewPool(30)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("创建池失败: %v", err))
		return
	}
	defer pool.Release() // 确保池被释放
	var wg sync.WaitGroup
	for _, target := range targets {
		wg.Add(1)
		var pageMonitResult types.PageMonit
		err := json.Unmarshal([]byte(target), &pageMonitResult)
		if err != nil {
			logger.SlogWarnLocal(fmt.Sprintf("PageMonitoringRunner json.Unmarshal error: %v", err))
			continue
		}
		pageMonitResultCopy := pageMonitResult
		err = pool.Submit(func() {
			Handler(pageMonitResultCopy, &wg)
		})
		if err != nil {
			logger.SlogWarnLocal(fmt.Sprintf("任务提交失败: %v", err))
		}
	}
	time.Sleep(3 * time.Second)
	wg.Wait()
}

func Handler(pageMonitResult types.PageMonit, wg *sync.WaitGroup) {
	defer wg.Done()
	response, err := utils.Requests.HttpGet(pageMonitResult.Url)
	if err != nil {
		return
	}

	flag := utils.Tools.IsSuffixURL(pageMonitResult.Url, ".js")
	if flag {
		if strings.Contains(response.Body, "<!DOCTYPE html>") {
			response.StatusCode = 0
			response.Body = ""
		}
	}
	var newHash string
	if response.Body == "" {
		newHash = ""
	} else {
		newHash = utils.Tools.HashXX64String(response.Body)
	}
	if len(pageMonitResult.Hash) == 0 {
		rootDomain, err := utils.Tools.GetRootDomain(pageMonitResult.Url)
		if err != nil {
			return
		}
		pageMonitResult.StatusCode = []int{response.StatusCode}
		pageMonitResult.Hash = []string{newHash}
		mongodb.MongodbClient.Update("PageMonitoring", bson.M{"md5": pageMonitResult.Md5}, bson.M{"$set": bson.M{"hash": pageMonitResult.Hash,
			"time":       utils.Tools.GetTimeNow(),
			"statusCode": pageMonitResult.StatusCode,
			"rootDomain": rootDomain,
		}})
		mongodb.MongodbClient.Upsert("PageMonitoringBody", bson.M{"md5": pageMonitResult.Md5}, bson.M{"$set": bson.M{"content": []string{response.Body}}})
		return
	}
	// 如果状态码相同，比较响应体hash是否相同，不同则计算两个响应体的相似度，记录相似度的值
	// 并且判断 响应体hash是一个还是两个，如果是一个，那直接加入hash 存入body
	// 如果是两个，那么删除第一个hash和对应的body，增加第二个hash和对应的body， 这里考虑一下存储body是使用body的hash还是url的hash
	if pageMonitResult.StatusCode[len(pageMonitResult.StatusCode)-1] == response.StatusCode {
		if response.StatusCode == 0 {
			return
		}
		if pageMonitResult.Hash[len(pageMonitResult.Hash)-1] != newHash {
			tmp := types.PageMonitBody{}
			err := mongodb.MongodbClient.FindOne("PageMonitoringBody", bson.M{"md5": pageMonitResult.Md5}, bson.M{"content": 1}, &tmp)
			if err != nil {
				logger.SlogErrorLocal(fmt.Sprintf("PageMonitoringBody findone error: %v", err))
				return
			}
			similarity, err := utils.Tools.CompareContentSimilarity(tmp.Content[len(tmp.Content)-1], response.Body)
			if err != nil {
				logger.SlogWarnLocal(fmt.Sprintf("CompareContentSimilarity error: %v", err))
				return
			}
			pageMonitResult.Similarity = similarity
			if len(pageMonitResult.Hash) == 2 {
				pageMonitResult.Hash = pageMonitResult.Hash[1:]
			}
			pageMonitResult.Hash = append(pageMonitResult.Hash, newHash)

			if len(tmp.Content) == 2 {
				tmp.Content = tmp.Content[1:]
			}
			tmp.Content = append(tmp.Content, response.Body)
			mongodb.MongodbClient.Update("PageMonitoringBody", bson.M{"md5": pageMonitResult.Md5}, bson.M{"$set": bson.M{"content": tmp.Content}})
			mongodb.MongodbClient.Update("PageMonitoring", bson.M{"md5": pageMonitResult.Md5}, bson.M{"$set": bson.M{"hash": pageMonitResult.Hash,
				"time":       utils.Tools.GetTimeNow(),
				"similarity": pageMonitResult.Similarity,
			}})
		}
	} else {
		// 状态码不相同，记录状态码
		if len(pageMonitResult.StatusCode) == 2 {
			pageMonitResult.StatusCode = pageMonitResult.StatusCode[1:]
		}
		pageMonitResult.StatusCode = append(pageMonitResult.StatusCode, response.StatusCode)
		tmp := types.PageMonitBody{}
		err := mongodb.MongodbClient.FindOne("PageMonitoringBody", bson.M{"md5": pageMonitResult.Md5}, bson.M{"content": 1}, &tmp)
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("PageMonitoringBody2 findone error: %v", err))
			return
		}
		if len(tmp.Content) == 2 {
			tmp.Content = tmp.Content[1:]
		}
		if len(pageMonitResult.Hash) == 2 {
			pageMonitResult.Hash = pageMonitResult.Hash[1:]
		}
		pageMonitResult.Hash = append(pageMonitResult.Hash, newHash)
		tmp.Content = append(tmp.Content, response.Body)
		mongodb.MongodbClient.Update("PageMonitoringBody", bson.M{"md5": pageMonitResult.Md5}, bson.M{"$set": bson.M{"content": tmp.Content}})
		mongodb.MongodbClient.Update("PageMonitoring", bson.M{"md5": pageMonitResult.Md5}, bson.M{"$set": bson.M{"hash": pageMonitResult.Hash, "statusCode": pageMonitResult.StatusCode, "time": utils.Tools.GetTimeNow()}})
	}

}

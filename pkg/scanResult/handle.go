// scanResult-------------------------------------
// @file      : handle.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/9 21:37
// -------------------------------------------

package scanResult

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"github.com/sergi/go-diff/diffmatchpatch"
	"strconv"
	"strings"
	"sync"
)

func GetAssetOwner(domain string) string {
	for _, p := range system.Projects {
		for _, t := range p.Target {
			domainRoot, err := util.GetRootDomain(domain)
			if err != nil {
				domainRoot = domain
			}
			if strings.Contains(domainRoot, t) {
				return p.ID
			}
		}
	}
	return ""
}

func popLastTwoBool(slice []bool) (bool, bool, []bool) {
	if len(slice) < 2 {
		return false, false, slice // 如果切片长度小于2，直接返回原切片
	}

	// 获取最后两个元素
	lastIndex := len(slice) - 1
	last := slice[lastIndex]
	secondLast := slice[lastIndex-1]

	// 使用切片操作去除最后两个元素
	slice = slice[:lastIndex-1]

	return secondLast, last, slice
}

func WebFingerScan(httpResult types.AssetHttp) types.AssetHttp {
	var wg sync.WaitGroup
	var mu sync.Mutex
	MaxTaskNumInt, err := strconv.Atoi(system.AppConfig.System.MaxTaskNum)
	maxWorkers := 10
	if err == nil {
		maxWorkersTmp := MaxTaskNumInt - system.AppConfig.System.Running
		if maxWorkersTmp > 10 {
			maxWorkers = maxWorkersTmp
		}
	}
	semaphore := make(chan struct{}, maxWorkers)

	for _, finger := range system.WebFingers {
		semaphore <- struct{}{} // 占用一个槽，限制并发数量
		wg.Add(1)
		go func(finger types.WebFinger) {
			defer func() {
				<-semaphore // 释放一个槽，允许新的goroutine开始
				wg.Done()
			}()
			tmpExp := []bool{}
			for _, exp := range finger.Express {
				key := ""
				value := ""
				if exp != "||" && exp != "&&" {
					r := strings.SplitN(exp, "=", 2)
					if len(r) != 2 {
						continue
					}
					key = r[0]
					value = strings.Trim(r[1], `"`)
				} else {
					key = exp
				}
				switch key {
				case "title", "title!":
					if strings.Contains(httpResult.Title, value) {
						if key == "title" {
							tmpExp = append(tmpExp, true)
						} else { // key == "title!"
							tmpExp = append(tmpExp, false)
						}
					} else {
						if key == "title!" {
							tmpExp = append(tmpExp, true)
						} else { // key == "title!"
							tmpExp = append(tmpExp, false)
						}
					}
				case "body", "body!":
					if strings.Contains(httpResult.ResponseBody, value) {
						if key == "body" {
							tmpExp = append(tmpExp, true)
						} else { // key == "title!"
							tmpExp = append(tmpExp, false)
						}
					} else {
						if key == "body!" {
							tmpExp = append(tmpExp, true)
						} else { // key == "title!"
							tmpExp = append(tmpExp, false)
						}
					}
				case "header", "header!":
					if strings.Contains(httpResult.RawHeaders, value) {
						if key == "header" {
							tmpExp = append(tmpExp, true)
						} else { // key == "title!"
							tmpExp = append(tmpExp, false)
						}
					} else {
						if key == "header!" {
							tmpExp = append(tmpExp, true)
						} else { // key == "title!"
							tmpExp = append(tmpExp, false)
						}
					}
				case "banner", "banner!":
					if strings.Contains(httpResult.RawHeaders, value) {
						if key == "banner" {
							tmpExp = append(tmpExp, true)
						} else { // key == "title!"
							tmpExp = append(tmpExp, false)
						}
					} else {
						if key == "banner!" {
							tmpExp = append(tmpExp, true)
						} else { // key == "title!"
							tmpExp = append(tmpExp, false)
						}
					}
				case "||":
					secondLast, last, slice := popLastTwoBool(tmpExp)
					r := last || secondLast
					slice = append(slice, r)
					tmpExp = slice
				case "&&":
					secondLast, last, slice := popLastTwoBool(tmpExp)
					r := last && secondLast
					slice = append(slice, r)
					tmpExp = slice
				default:
					tmpExp = append(tmpExp, false)
				}
			}

			if len(tmpExp) != 1 {
				myLog := system.CustomLog{
					Status: "Error",
					Msg:    fmt.Sprintf("WebFingerScan error: %s - %s", finger.ID, httpResult.URL),
				}
				system.PrintLog(myLog)
				return
			}

			flag := tmpExp[0]
			if flag {
				mu.Lock()
				httpResult.WebFinger = append(httpResult.WebFinger, finger.ID)
				mu.Unlock()
			}
		}(finger)
	}
	wg.Wait()
	return httpResult
}

func DiffContent(str1 string, str2 string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(str1, str2, false)
	msg := ""
	diffCount := len(diffs)
	for i, diff := range diffs {
		formatType := fmt.Sprintf("%v", diff.Type)
		if formatType == "Equal" {
			continue
		}
		msg += fmt.Sprintf("%v:[*][*] %v\n", diff.Type, diff.Text)
		if i < diffCount-1 {
			msg += "---------------------------------------------\n"
		}
	}
	return msg
}

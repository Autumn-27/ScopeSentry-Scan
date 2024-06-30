// Package urlScanMode -----------------------------
// @file      : urlScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/15 19:43
// -------------------------------------------
package urlScanMode

import (
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/sensitiveMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"github.com/jaeles-project/gospider/core"
	"github.com/sirupsen/logrus"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Option struct {
	Cookie  string
	Headers []string
}

func isMatchingFilter(fs []*regexp.Regexp, d []byte) bool {
	for _, r := range fs {
		if r.Match(d) {
			return true
		}
	}
	return false
}

func isContentType(s string) bool {
	mediaTypes := []string{
		"text/html",
		"multipart/form-data",
		"application/x-www-form-urlencoded",
		"text/plain",
		"text/xml",
		"image/gif",
		"image/jpeg",
		"image/png",
		"application/xml",
		"application/json",
		"application/octet-stream",
		"application/xhtml+xml",
		"application/atom+xml",
		"application/pdf",
		"application/msword",
	}
	for _, ty := range mediaTypes {
		if ty == s {
			return true
		}
	}
	return false
}

func Run(op Option, siteList []string, secretFlag bool, pageMonitoring string, taskId string) []types.UrlResult {
	defer system.RecoverPanic("urlScanMode")
	core.Logger.SetLevel(logrus.PanicLevel)
	// Check again to make sure at least one site in slice
	if len(siteList) == 0 {
		system.SlogInfo("URL Scan No site in list")
		return nil
	}
	MaxUrlNum, err := strconv.Atoi(system.AppConfig.System.UrlMaxNum)
	if err != nil {
		MaxUrlNum = 500
	}
	threads, err := strconv.Atoi(system.AppConfig.System.UrlThread)
	if err != nil {
		threads = 5
	}
	sitemap := true
	linkfinder := true
	robots := true
	PageMonitoringInitResults := []types.TmpPageMonitResult{}
	urlInfos := []types.UrlResult{}
	var DisallowedURLFilters []*regexp.Regexp
	disallowedRegex := `(?i)\.(png|apng|bmp|gif|ico|cur|jpg|jpeg|jfif|pjp|pjpeg|svg|tif|tiff|webp|xbm|3gp|aac|flac|mpg|mpeg|mp3|mp4|m4a|m4v|m4p|oga|ogg|ogv|mov|wav|webm|eot|woff|woff2|ttf|otf|css)(?:\?|#|$)`
	DisallowedURLFilters = append(DisallowedURLFilters, regexp.MustCompile(disallowedRegex))
	var seenUrls sync.Map
	urlScanResultHandler := func(msg string) {
		urlInfo := types.UrlResult{}
		err := json.Unmarshal([]byte(msg), &urlInfo)
		if err != nil {
			system.SlogErrorLocal(fmt.Sprintf("Url scan parse json err: %s", err))
			return
		}
		//system.SlogDebugLocal(fmt.Sprintf("urlScanResultHandler: %s", msg))
		urlInfo.Time = system.GetTimeNow()
		var url string
		url = urlInfo.Output
		teFlag := isContentType(url)
		if teFlag {
			return
		}
		if !isMatchingFilter(DisallowedURLFilters, []byte(url)) {
			urlInfo.Output = strings.TrimSpace(urlInfo.Output)
			inputDomainRoot, err := util.GetRootDomain(urlInfo.Input)
			if err != nil {
				system.SlogError(fmt.Sprintf("inputDomainRoot %s error: %s", urlInfo.Input, err))
			}
			hostParts := strings.Split(urlInfo.Output, ".")
			if len(hostParts) < 2 {
				urlInfo.Output = urlInfo.Input + "/" + urlInfo.Output
			}
			outputDomainRoot, err := util.GetRootDomain(urlInfo.Output)
			if inputDomainRoot == outputDomainRoot {
				if _, seen := seenUrls.Load(url); !seen {
					seenUrls.Store(url, struct{}{})
					flag := scanResult.URLRedisDeduplication(urlInfo.Output, taskId)
					if !flag {
						scanResult.UrlResult([]types.UrlResult{urlInfo}, taskId)
						urlInfos = append(urlInfos, urlInfo)
					}
				}
			}
			if pageMonitoring == "JS" {
				if idx := strings.Index(url, "?"); idx != -1 {
					url = url[:idx]
				}
				if strings.HasSuffix(url, ".js") {
					tmpPage := types.TmpPageMonitResult{
						Url:     url,
						Content: "",
					}
					PageMonitoringInitResults = append(PageMonitoringInitResults, tmpPage)
					scanResult.PageMonitoringInitResult([]types.TmpPageMonitResult{tmpPage}, taskId)
				}
			}
			if pageMonitoring == "All" {
				tmpPage := types.TmpPageMonitResult{
					Url:     url,
					Content: "",
				}
				PageMonitoringInitResults = append(PageMonitoringInitResults, tmpPage)
				scanResult.PageMonitoringInitResult([]types.TmpPageMonitResult{tmpPage}, taskId)
			}
		}
	}
	var wg sync.WaitGroup
	respBodyHandler := func(url string, msg string) {
		if secretFlag || pageMonitoring != "None" {
			if !isMatchingFilter(DisallowedURLFilters, []byte(url)) {
				if secretFlag {
					wg.Add(1)
					go func(url string, msg string) {
						defer func() {
							wg.Done()
						}()
						respMd5 := util.CalculateMD5(msg)
						if !scanResult.SensRedisDeduplication(respMd5, taskId) {
							sensitiveMode.Scan(url, msg, respMd5, taskId)
						}
					}(url, msg)
				}
			}
		}
	}

	inputChan := make(chan string, threads)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for rawSite := range inputChan {
			site, err := url.Parse(rawSite)
			if err != nil {
				system.SlogErrorLocal(fmt.Sprintf("Failed to parse %s: %s", rawSite, err))
				continue
			}
			if site.Hostname() == "" {
				system.SlogDebugLocal(fmt.Sprintf("site %v host is null %v", rawSite, siteList))
				continue
			}
			var siteWg sync.WaitGroup
			crawler := core.NewCrawler(site, op.Cookie, op.Headers, urlScanResultHandler, respBodyHandler, MaxUrlNum)
			if crawler == nil {
				continue
			}
			siteWg.Add(1)
			go func() {
				defer siteWg.Done()
				crawler.Start(linkfinder)
			}()

			// Brute force Sitemap path
			if sitemap {
				siteWg.Add(1)
				go core.ParseSiteMap(site, crawler, crawler.C, &siteWg)
			}

			// Find Robots.txt
			if robots {
				siteWg.Add(1)
				go core.ParseRobots(site, crawler, crawler.C, &siteWg)
			}

			siteWg.Wait()
			crawler.C.Wait()
			crawler.LinkFinderCollector.Wait()
		}
	}()
	for _, site := range siteList {
		if !isMatchingFilter(DisallowedURLFilters, []byte(site)) {
			inputChan <- site
		}
	}
	time.Sleep(5 * time.Second)
	close(inputChan)
	wg.Wait()
	system.SlogInfo(fmt.Sprintf("Get Url result %v, get page monitoring result %v", len(urlInfos), len(PageMonitoringInitResults)))
	return urlInfos

}

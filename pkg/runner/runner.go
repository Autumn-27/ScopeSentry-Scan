// Package runner -----------------------------
// @file      : runner.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/9 20:20
// -------------------------------------------
package runner

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/dirScanMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/httpxMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/portScanMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/subdomainMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/subdomainTakeoverMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/subfinderMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/urlScanMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/vulnMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/waybackMode"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

type Option struct {
	SubfinderEnabled      bool
	SubdomainScanEnabled  bool
	KsubdomainScanEnabled bool
	PortScanEnabled       bool
	DirScanEnabled        bool
	CrawlerEnabled        bool
	Ports                 string
	WaybackurlEnabled     bool
	SensitiveInfoScan     bool
	UrlScan               bool
	Cookie                string
	Header                []string
	Duplicates            string
	TaskId                string
	VulScan               bool
	VulList               []string
	PageMonitoring        string
}

func isValidIP(ipStr string) bool {
	// IP 地址的正则表达式模式（可能带有端口）
	ipPattern := `^(\d{1,3}\.){3}\d{1,3}(:\d+)?$`

	// 编译正则表达式
	regexpIP := regexp.MustCompile(ipPattern)

	// 匹配 IP 地址
	return regexpIP.MatchString(ipStr)
}

func Process(Host string, op Option) {
	defer system.RecoverPanic("Task Runner Process")
	system.StartTask()
	system.SlogInfo(fmt.Sprintf("target %s starts scanning", Host))
	scanResult.ProgressStart("scan", Host, op.TaskId)
	system.SlogDebugLocal("target Parse begin")
	normalizedHttp := ""
	if !strings.HasPrefix(Host, "http://") && !strings.HasPrefix(Host, "https://") {
		normalizedHttp = "http://" + Host
	} else {
		normalizedHttp = Host
	}
	parsedURL, err := url.Parse(normalizedHttp)
	if err != nil {
		system.SlogError(fmt.Sprintf("%s: %s\n", Host, "parse url error"))
		scanResult.TaskEnds(Host, op.TaskId)
		system.AppConfig.System.Running = system.AppConfig.System.Running - 1
		return
	}
	hostParts := strings.Split(parsedURL.Host, ":")
	hostWithoutPort := hostParts[0]
	ipIsValid := isValidIP(hostWithoutPort)
	port := parsedURL.Port()
	SubDomainResults := []types.SubdomainResult{}
	domainList := []string{}
	if ipIsValid {
		if port != "" {
			tmp := Host
			dotIndex := strings.Index(Host, "*.")
			if dotIndex != -1 {
				substr := Host[dotIndex+2:]
				tmp = substr
			}
			domainList = append(domainList, tmp)
		}
		tmp := hostWithoutPort
		dotIndex := strings.Index(hostWithoutPort, "*.")
		if dotIndex != -1 {
			substr := hostWithoutPort[dotIndex+2:]
			tmp = substr
		}
		domainList = append(domainList, tmp)
	} else {
		// 子域名扫描
		domainDnsResult := subdomainMode.Verify([]string{hostWithoutPort})
		SubDomainResults = append(SubDomainResults, domainDnsResult...)
		if len(SubDomainResults) != 0 {
			if port != "" {
				tmp := Host
				dotIndex := strings.Index(hostWithoutPort, "*.")
				if dotIndex != -1 {
					substr := hostWithoutPort[dotIndex+2:]
					tmp = substr
				}
				domainList = append(domainList, tmp)
			}
		} else {
			tmp := Host
			dotIndex := strings.Index(Host, "*.")
			if dotIndex != -1 {
				substr := Host[dotIndex+2:]
				tmp = substr
			}
			domainList = append(domainList, tmp)
		}
		system.SlogDebugLocal("target Parse done")
		if op.SubdomainScanEnabled {
			system.SlogInfo(fmt.Sprintf("target %s subdomain enumeration begins", Host))
			scanResult.ProgressStart("subdomain", Host, op.TaskId)
			if op.SubfinderEnabled {
				subfinderDomain := hostWithoutPort
				dotIndex := strings.Index(hostWithoutPort, "*.")
				if dotIndex != -1 {
					substr := hostWithoutPort[dotIndex+2:]
					subfinderDomain = substr
				}
				subfinderResult := subfinderMode.SubfinderScan(subfinderDomain)
				SubDomainResults = append(SubDomainResults, subfinderResult...)
			}
			if op.SubdomainScanEnabled {
				// 判断是否泛解析，跳过泛解析
				if !IsWildCard(hostWithoutPort) {
					subDomainResult := subdomainMode.SubDomainRunner(hostWithoutPort)
					SubDomainResults = append(SubDomainResults, subDomainResult...)
				}
			}
			system.SlogInfo(fmt.Sprintf("target %s subdomain enumeration ends", Host))
			scanResult.ProgressEnd("subdomain", Host, op.TaskId)
		}
		uniqueSubDomainResults := []types.SubdomainResult{}
		seenHosts := make(map[string]struct{})
		for _, result := range SubDomainResults {
			if _, seen := seenHosts[result.Host]; seen {
				continue
			}
			flag := scanResult.SubdomainRedisDeduplication(result.Host, op.TaskId)
			if flag {
				continue
			}
			if op.Duplicates == "subdomain" {
				flag = scanResult.SubdomainMongoDbDeduplication(result.Host)
				if flag {
					continue
				}
			}
			seenHosts[result.Host] = struct{}{}
			alreadyInDomainList := false
			for _, domain := range domainList {
				if domain == result.Host {
					alreadyInDomainList = true
					break
				}
			}
			if !alreadyInDomainList {
				tmp := result.Host
				dotIndex := strings.Index(result.Host, "*.")
				if dotIndex != -1 {
					substr := result.Host[dotIndex+2:]
					tmp = substr
				}
				domainList = append(domainList, tmp)
			}
			uniqueSubDomainResults = append(uniqueSubDomainResults, result)
		}
		system.SlogInfo(fmt.Sprintf("Get the number of %s unique subdomains as %v, raw subdoamins %v", Host, len(uniqueSubDomainResults), len(SubDomainResults)))
		if len(uniqueSubDomainResults) != 0 {
			f := scanResult.SubdoaminResult(uniqueSubDomainResults, op.TaskId)
			if !f {
				system.SlogError(fmt.Sprintf("Insert subdomain error"))
				scanResult.TaskEnds(Host, op.TaskId)
				scanResult.ProgressEnd("scan", Host, op.TaskId)
				system.EndTask()
				return
			}
		}
		// 子域名接管
		system.SlogInfo(fmt.Sprintf("target %s subdomainTakeover begins", Host))
		scanResult.ProgressStart("subdomainTakeover", Host, op.TaskId)
		for _, sd := range uniqueSubDomainResults {
			if sd.Type == "CNAME" {
				subdomainTakeoverMode.Scan(sd.Host, sd.Value, op.TaskId)
			}
		}
		system.SlogInfo(fmt.Sprintf("target %s subdomainTakeover ends", Host))
		scanResult.ProgressEnd("subdomainTakeover", Host, op.TaskId)
	}
	system.SlogInfo(fmt.Sprintf("target %s asset mapping begins", Host))
	scanResult.ProgressStart("assetMapping", Host, op.TaskId)
	// 资产、端口扫描
	var httpxResults []types.AssetHttp
	var httpxResultsMutex sync.Mutex
	httpxResultsHandler := func(r types.AssetHttp) {
		httpxResultsMutex.Lock()
		httpxResults = append(httpxResults, r)
		httpxResultsMutex.Unlock()
	}
	httpxMode.HttpxScan(domainList, httpxResultsHandler)
	assetOthers := []types.AssetOther{}
	// 端口扫描
	if op.PortScanEnabled {
		system.SlogInfo(fmt.Sprintf("%s port scan start", Host))
		scanResult.ProgressStart("portScan", Host, op.TaskId)
		var mutex sync.Mutex
		for _, domain := range domainList {
			assetHttpTemp, assetOtherTemp := portScanMode.PortScan(domain, op.Ports, op.Duplicates)
			mutex.Lock()
			httpxResults = append(httpxResults, assetHttpTemp...)
			assetOthers = append(assetOthers, assetOtherTemp...)
			mutex.Unlock()
		}
		system.SlogInfo(fmt.Sprintf("%s port scan ends", Host))
		scanResult.ProgressEnd("portScan", Host, op.TaskId)
	}
	// 将HTTP资产进行去重
	existing := make(map[string]bool)
	var uniqueHttpResults []types.AssetHttp
	for _, result := range httpxResults {
		key := result.URL + result.Port
		if !existing[key] {
			flag := scanResult.URLRedisDeduplication(result.URL, op.TaskId)
			if flag {
				continue
			}
			existing[key] = true
			uniqueHttpResults = append(uniqueHttpResults, result)
		}
	}
	// 非HTTP资产去重
	existingOther := make(map[string]bool)
	var uniqueassetOtherResults []types.AssetOther
	for _, result := range assetOthers {
		key := result.Host + ":" + result.Port
		if !existingOther[key] {
			existingOther[key] = true
			uniqueassetOtherResults = append(uniqueassetOtherResults, result)
		}
	}
	// 存储资产结果
	scanResult.AssetResult(uniqueHttpResults, uniqueassetOtherResults, op.TaskId)
	system.SlogInfo(fmt.Sprintf("target %s asset mapping completed", Host))
	scanResult.ProgressEnd("assetMapping", Host, op.TaskId)
	//缓存污染 todo

	// url扫描、js信息泄露
	var urlResu []types.UrlResult
	urlResu = []types.UrlResult{}
	domainUrlScanList := []string{}
	for _, httpxResult := range uniqueHttpResults {

		domainUrlScanList = append(domainUrlScanList, httpxResult.URL)
	}
	system.SlogInfo(fmt.Sprintf("Get target %v http url number %v", Host, len(domainUrlScanList)))
	if op.UrlScan {
		system.SlogInfo(fmt.Sprintf("target %s url scan begins", Host))
		scanResult.ProgressStart("urlScan", Host, op.TaskId)
		urlScanOption := urlScanMode.Option{
			Cookie:  op.Cookie,
			Headers: op.Header,
		}
		if op.SensitiveInfoScan {
			system.SlogInfo(fmt.Sprintf("target %s SensitiveInfo scan begins", Host))
			scanResult.ProgressStart("sensitive", Host, op.TaskId)
		}
		urlResu = urlScanMode.Run(urlScanOption, domainUrlScanList, op.SensitiveInfoScan, op.PageMonitoring, op.TaskId)
		if op.WaybackurlEnabled {
			system.SlogInfo(fmt.Sprintf("target %s waybackurl scan start", Host))
			waybackMode.Runner(domainUrlScanList, op.TaskId)
			system.SlogInfo(fmt.Sprintf("target %s waybackurl scan end", Host))
		}
		system.SlogInfo(fmt.Sprintf("targety %v URLScan get result %v", Host, len(urlResu)))
		system.SlogInfo(fmt.Sprintf("target %s url scan completed", Host))
		scanResult.ProgressEnd("urlScan", Host, op.TaskId)
		if op.SensitiveInfoScan {
			system.SlogInfo(fmt.Sprintf("target %s SensitiveInfo scan completed", Host))
			scanResult.ProgressEnd("sensitive", Host, op.TaskId)
		}
	}
	var crawWg sync.WaitGroup
	// 爬虫
	if op.CrawlerEnabled {
		system.SlogInfo(fmt.Sprintf("target %s crawler scan begins", Host))
		scanResult.ProgressStart("crawler", Host, op.TaskId)
		var targetList []string
		if len(urlResu) == 0 {
			targetList = append(targetList, domainUrlScanList...)
		} else {
			for _, u := range urlResu {
				targetList = append(targetList, u.Output)
			}
		}
		//crawlerMode.CrawlerScan(targetList)
		crawWg.Add(1)
		system.CrawlerTarget <- types.CrawlerTask{Target: targetList, Host: Host, Id: op.TaskId, Wg: &crawWg}
	}
	//目录扫描
	if op.DirScanEnabled {
		scanResult.ProgressStart("dirScan", Host, op.TaskId)
		system.SlogInfo(fmt.Sprintf("target %s directory scan begins", Host))
		dirScanMode.Scan(domainUrlScanList, op.TaskId)
		system.SlogInfo(fmt.Sprintf("target %s directory scan completed", Host))
		scanResult.ProgressEnd("dirScan", Host, op.TaskId)
	}

	//漏洞扫描
	if op.VulScan {
		system.SlogInfo(fmt.Sprintf("target %s vulnerability scan begins", Host))
		scanResult.ProgressStart("vulnerability", Host, op.TaskId)
		template := []string{}
		if len(op.VulList) != 0 {
			if op.VulList[0] == "All Poc" {
				template = append(template, "*")
			} else {
				for _, vul := range op.VulList {
					template = append(template, fmt.Sprintf("%s.yaml", vul))
				}
			}
			if len(domainUrlScanList) != 0 {
				vulnMode.Scan(domainUrlScanList, template, op.TaskId)
			}
		}
		system.SlogInfo(fmt.Sprintf("target %s vulnerability scan completed", Host))
		scanResult.ProgressEnd("vulnerability", Host, op.TaskId)
	}
	crawWg.Wait()
	system.SlogInfo(fmt.Sprintf("target %s scan completed", Host))
	scanResult.TaskEnds(Host, op.TaskId)
	scanResult.ProgressEnd("scan", Host, op.TaskId)
	system.EndTask()
}

func IsWildCard(domain string) bool {
	targets := []string{}
	for i := 0; i < 2; i++ {
		dotIndex := strings.Index(domain, "*.")
		subdomain := util.GenerateRandomString(6) + "." + domain
		if dotIndex != -1 {
			subdomain = strings.Replace(domain, "*", util.GenerateRandomString(6), -1)
		}
		targets = append(targets, subdomain)
	}
	result := subdomainMode.Verify(targets)
	if len(result) == 0 {
		return false
	}
	return true
}

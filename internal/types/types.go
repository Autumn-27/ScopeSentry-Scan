// Package types -----------------------------
// @file      : type.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/11 10:05
// -------------------------------------------
package types

import (
	"encoding/json"
	"github.com/dlclark/regexp2"
	"github.com/projectdiscovery/katana/pkg/navigation"
	"github.com/projectdiscovery/tlsx/pkg/tlsx/clients"
	"go.mongodb.org/mongo-driver/bson"
	"regexp"
	"sync"
	"time"
)

type SubdomainResult struct {
	Host       string
	Type       string
	Value      []string
	IP         []string
	Time       string
	Tags       []string `bson:"tags"`
	Project    string
	TaskName   string `bson:"taskName"`
	RootDomain string `bson:"rootDomain"`
}

type AssetHttp struct {
	Time          string                 `bson:"time" csv:"time"`
	LastScanTime  string                 `bson:"lastScanTime"`
	TLSData       *clients.Response      `bson:"tls" csv:"tls"`
	Hashes        map[string]interface{} `bson:"hash" csv:"hash"`
	CDNName       string                 `bson:"cdnname" csv:"cdn_name"`
	Port          string                 `bson:"port" csv:"port"`
	URL           string                 `bson:"url" csv:"url"`
	Title         string                 `bson:"title" csv:"title"`
	Type          string                 `bson:"type" csv:"type"`
	Error         string                 `bson:"error" csv:"error"`
	ResponseBody  string                 `bson:"body" csv:"body"`
	Host          string                 `bson:"host" csv:"host"`
	IP            string                 `bson:"ip"`
	Screenshot    string                 `bson:"screenshot"`
	FavIconMMH3   string                 `bson:"faviconmmh3" csv:"favicon"`
	FaviconPath   string                 `bson:"faviconpath" csv:"favicon_path"`
	RawHeaders    string                 `bson:"rawheaders" csv:"raw_header"`
	Jarm          string                 `bson:"jarm" csv:"jarm"`
	Technologies  []string               `bson:"technologies" csv:"tech"`
	StatusCode    int                    `bson:"statuscode" csv:"status_code"`
	ContentLength int                    `bson:"contentlength" csv:"content_length"`
	CDN           bool                   `bson:"cdn" csv:"cdn"`
	Webcheck      bool                   `bson:"webcheck" csv:"webcheck"`
	Project       string                 `bson:"project" csv:"project"`
	IconContent   string                 `bson:"iconcontent"`
	Domain        string                 `bson:"domain"`
	TaskName      []string               `bson:"taskName"`
	WebServer     string                 `bson:"webServer"`
	Service       string                 `bson:"service"`
	RootDomain    string                 `bson:"rootDomain"`
	Tags          []string               `bson:"tags"`
}

type PortAlive struct {
	Host string `bson:"host"`
	IP   string `bson:"ip"`
	Port string `bson:"port"`
}
type Project struct {
	ID              string           `bson:"id"`
	Target          []string         `bson:"target"`
	IgnoreList      []string         `bson:"ignoreList"`
	IgnoreRegexList []*regexp.Regexp `yaml:"ignoreRegexList"`
}

type AssetOther struct {
	Time         string          `bson:"time" csv:"time"`
	LastScanTime string          `bson:"lastScanTime"`
	Host         string          `bson:"host"`
	IP           string          `bson:"ip"`
	Port         string          `bson:"port"`
	Service      string          `bson:"service"`
	TLS          bool            `bson:"tls"`
	Transport    string          `bson:"transport"`
	Version      string          `bson:"version"`
	Raw          json.RawMessage `bson:"metadata"`
	Project      string          `bson:"project"`
	Type         string          `bson:"type"`
	Tags         []string        `bson:"tags"`
	TaskName     []string        `bson:"taskName"`
	RootDomain   string          `bson:"rootDomain"`
	UrlPath      string          `bson:"urlPath"`
}

type ChangeLog struct {
	FieldName string `json:"fieldName"`
	Old       string `json:"old"`
	New       string `json:"new"`
}

type AssetChangeLog struct {
	AssetId   string `json:"assetId"`
	Timestamp string `json:"timestamp" csv:"timestamp"`
	Change    []ChangeLog
}

type UrlResult struct {
	Input      string `json:"input"`
	Source     string `json:"source"`
	OutputType string `json:"type"`
	Output     string `json:"output"`
	Status     int    `json:"status"`
	Length     int    `json:"length"`
	Time       string `json:"time"`
	Body       string `bson:"body"`
	Ext        string `bson:"ext"`
	Project    string
	TaskName   string   `bson:"taskName"`
	ResultId   string   `bson:"resultId"`
	RootDomain string   `bson:"rootDomain"`
	Tags       []string `bson:"tags"`
}

type UrlFile struct {
	Filepath string
}

type SecretResults struct {
	Url   string
	Kind  string
	Key   string
	Value string
}

type CrawlerResult struct {
	Url        string
	Method     string
	Body       string
	Project    string
	TaskName   string   `bson:"taskName"`
	ResultId   string   `bson:"resultId"`
	RootDomain string   `bson:"rootDomain"`
	Time       string   `json:"time"`
	Tags       []string `bson:"tags"`
}

type PortDict struct {
	ID    string `bson:"id"`
	Value string `bson:"value"`
}

type SubdomainTakerFinger struct {
	Name     string
	Cname    []string
	Response []string
}

type SubTakeResult struct {
	Input      string
	Value      string
	Cname      string
	Response   string
	Project    string
	TaskName   string   `bson:"taskName"`
	RootDomain string   `bson:"rootDomain"`
	Tags       []string `bson:"tags"`
}

type DirResult struct {
	Url        string
	Status     int
	Msg        string
	Project    string
	Length     int
	TaskName   string   `bson:"taskName"`
	RootDomain string   `bson:"rootDomain"`
	Tags       []string `bson:"tags"`
}

type SensitiveRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	State       bool     `json:"enabled"`
	Regular     string   `json:"pattern"`
	Color       string   `bson:"color"`
	Tags        []string `bson:"tags"`
	RuleCompile *regexp2.Regexp
}

type SensitiveResult struct {
	Url        string
	UrlId      string
	SID        string
	Match      []string
	Project    string
	Color      string
	Time       string
	Md5        string
	TaskName   string   `bson:"taskName"`
	RootDomain string   `bson:"rootDomain"`
	Tags       []string `bson:"tags"`
	Status     int      `bson:"status"` // 1表示未处理 2表示处理中 3表示忽略 4表示疑似 5表示确认
}

type VulnResult struct {
	Url        string
	VulnId     string
	VulName    string
	Matched    string
	Project    string
	Level      string
	Time       string
	Request    string
	Response   string
	TaskName   string   `bson:"taskName"`
	RootDomain string   `yaml:"rootDomain"`
	Tags       []string `bson:"tags"`
	Status     int      `bson:"status"` // 1表示未处理 2表示处理中 3表示忽略 4表示疑似 5表示确认
}
type TmpPageMonitResult struct {
	Url      string
	Content  string
	TaskName string `bson:"taskName"`
}

type WebFinger struct {
	ID      string
	Express []string
	Name    string
	State   bool
}

type HttpSample struct {
	Url        string
	StatusCode int
	Body       string
	Msg        string
}

type HttpResponse struct {
	Url           string
	StatusCode    int
	Body          string
	ContentLength int
	Redirect      string
	Title         string
}

type NotificationConfig struct {
	SubdomainScan                 bool   `bson:"subdomainScan"`
	DirScanNotification           bool   `bson:"dirScanNotification"`
	PortScanNotification          bool   `bson:"portScanNotification"`
	SensitiveNotification         bool   `bson:"sensitiveNotification"`
	SubdomainNotification         bool   `bson:"subdomainNotification"`
	SubdomainTakeoverNotification bool   `bson:"subdomainTakeoverNotification"`
	PageMonNotification           bool   `bson:"pageMonNotification"`
	VulNotification               bool   `bson:"vulNotification"`
	VulLevel                      string `bson:"vulLevel"`
}

type NotificationApi struct {
	Url         string `bson:"url"`
	Method      string `bson:"method"`
	ContentType string `bson:"contentType"`
	Data        string `bson:"data"`
	State       bool   `bson:"state"`
}

type PocData struct {
	Name  string
	Level string
}

type CrawlerTask struct {
	Target []string
	Host   string
	Id     string
	Wg     *sync.WaitGroup
}

type DomainSkip struct {
	Domain string
	Skip   bool
	IP     []string
	CIDR   bool
}
type DomainResolve struct {
	Domain string
	IP     []string
}

type KatanaResult struct {
	Timestamp        time.Time
	Request          *navigation.Request
	Response         *navigation.Response
	PassiveReference *navigation.PassiveReference
	Error            string
}

type PageMonit struct {
	Url        string   `bson:"url"`
	Hash       []string `bson:"hash"`
	Md5        string   `bson:"md5"`
	Length     []int    `bson:"length"`
	StatusCode []int    `bson:"statusCode"`
	Similarity float64  `bson:"similarity"`
	State      int      `bson:"state"`
	Project    string   `bson:"project"`
	Time       string   `bson:"time"`
	TaskName   string   `bson:"taskName"`
	RootDomain string   `bson:"rootDomain"`
	Tags       []string `bson:"tags"`
}

type PageMonitBody struct {
	Content []string `bson:"content"`
	Md5     string   `bson:"md5"`
}

type BulkUpdateOperation struct {
	Selector bson.M // 条件选择器数组
	Update   bson.M // 更新内容数组
}

type RootDomain struct {
	Domain   string   `bson:"domain"`
	ICP      string   `bson:"icp"`
	Company  string   `bson:"company"`
	Tags     []string `bson:"tags"`
	TaskName string   `bson:"taskName"`
	Project  string   `bson:"project"`
	Time     string   `bson:"time"`
}

type APP struct {
	Name        string   `bson:"name"`
	Version     string   `bson:"version"`
	Url         string   `bson:"url"`
	ICP         string   `bson:"icp"`
	FilePath    string   `bson:"-"`
	Company     string   `bson:"company"`
	BundleID    string   `bson:"bundleID"`
	Category    string   `bson:"category"`
	Description string   `bson:"description"`
	APK         string   `bson:"apk"`
	Logo        string   `bson:"logo"`
	Tags        []string `bson:"tags"`
	TaskName    string   `bson:"taskName"`
	Project     string   `bson:"project"`
	Time        string   `bson:"time"`
}

type MP struct {
	Name        string   `bson:"name"`
	Url         string   `bson:"url"`
	ICP         string   `bson:"icp"`
	Description string   `bson:"description"`
	Company     string   `bson:"company"`
	FilePath    string   `bson:"-"`
	Tags        []string `bson:"tags"`
	TaskName    string   `bson:"taskName"`
	Project     string   `bson:"project"`
	Time        string   `bson:"time"`
}

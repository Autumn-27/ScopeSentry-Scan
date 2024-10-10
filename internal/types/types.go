// Package types -----------------------------
// @file      : type.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/11 10:05
// -------------------------------------------
package types

import (
	"encoding/json"
	"github.com/projectdiscovery/tlsx/pkg/tlsx/clients"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sync"
)

type SubdomainResult struct {
	Host       string
	Type       string
	Value      []string
	IP         []string
	Time       string
	Project    string
	TaskId     string `bson:"taskId"`
	RootDomain string `bson:"rootDomain"`
}

type AssetHttp struct {
	Timestamp     string                 `bson:"timestamp,omitempty" csv:"timestamp"`
	LastScanTime  string                 `bson:"lastScanTime"`
	TLSData       *clients.Response      `bson:"tls,omitempty" csv:"tls"`
	Hashes        map[string]interface{} `bson:"hash,omitempty" csv:"hash"`
	CDNName       string                 `bson:"cdn_name,omitempty" csv:"cdn_name"`
	Port          string                 `bson:"port,omitempty" csv:"port"`
	URL           string                 `bson:"url,omitempty" csv:"url"`
	Title         string                 `bson:"title,omitempty" csv:"title"`
	Type          string                 `bson:"Type,omitempty" csv:"Type"`
	Error         string                 `bson:"error,omitempty" csv:"error"`
	ResponseBody  string                 `bson:"body,omitempty" csv:"body"`
	Host          string                 `bson:"host,omitempty" csv:"host"`
	IP            string                 `bson:"ip"`
	FavIconMMH3   string                 `bson:"favicon,omitempty" csv:"favicon"`
	FaviconPath   string                 `bson:"favicon_path,omitempty" csv:"favicon_path"`
	RawHeaders    string                 `bson:"raw_header,omitempty" csv:"raw_header"`
	Jarm          string                 `bson:"jarm,omitempty" csv:"jarm"`
	Technologies  []string               `bson:"tech,omitempty" csv:"tech"`
	StatusCode    int                    `bson:"status_code,omitempty" csv:"status_code"`
	ContentLength int                    `bson:"content_length,omitempty" csv:"content_length"`
	CDN           bool                   `bson:"cdn,omitempty" csv:"cdn"`
	Webcheck      bool                   `bson:"webcheck,omitempty" csv:"webcheck"`
	Project       string                 `bson:"project,omitempty" csv:"project"`
	WebFinger     []string               `bson:"web_finger,omitempty" csv:"web_finger"`
	IconContent   string                 `bson:"iconContent"`
	Domain        string                 `bson:"domain"`
	TaskId        string                 `bson:"taskId"`
	WebServer     string                 `bson:"webServer"`
	Service       string                 `bson:"service"`
	RootDomain    string                 `bson:"rootDomain"`
}

type PortAlive struct {
	Host string `json:"Host,omitempty"`
	IP   string `json:"Host,omitempty"`
	Port string `json:"Port,omitempty"`
}
type Project struct {
	ID     string   `bson:"id"`
	Target []string `bson:"target"`
}
type AssetOther struct {
	Timestamp    string          `bson:"timestamp,omitempty" csv:"timestamp"`
	LastScanTime string          `bson:"lastScanTime"`
	Host         string          `bson:"host,omitempty"`
	IP           string          `bson:"ip"`
	Port         string          `bson:"port"`
	Service      string          `bson:"service"`
	TLS          bool            `bson:"tls"`
	Transport    string          `bson:"transport"`
	Version      string          `bson:"version,omitempty"`
	Raw          json.RawMessage `bson:"metadata"`
	Project      string          `bson:"project"`
	Type         string
	TaskId       string `bson:"taskId"`
	RootDomain   string `bson:"rootDomain"`
}

type ChangeLog struct {
	FieldName string `json:"fieldName"`
	Old       string `json:"old"`
	New       string `json:"new"`
}

type AssetChangeLog struct {
	AssetId   string `json:"assetId"`
	Timestamp string `json:"timestamp,omitempty" csv:"timestamp"`
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
	Project    string
	TaskId     string `bson:"taskId"`
}

type SecretResults struct {
	Url   string
	Kind  string
	Key   string
	Value string
}

type CrawlerResult struct {
	Url     string
	Method  string
	Body    string
	Project string
	TaskId  string `bson:"taskId"`
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
	TaskId     string `bson:"taskId"`
	RootDomain string
}

type DirResult struct {
	Url     string
	Status  int
	Msg     string
	Project string
	Length  int
	TaskId  string `bson:"taskId"`
}

type SensitiveRule struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	State   bool   `json:"enabled"`
	Regular string `json:"pattern"`
	Color   string `bson:"color"`
}

type SensitiveResult struct {
	Url     string
	SID     string
	Match   []string
	Project string
	Body    string
	Color   string
	Time    string
	Md5     string
	TaskId  string `bson:"taskId"`
}

type VulnResult struct {
	Url      string
	VulnId   string
	VulName  string
	Matched  string
	Project  string
	Level    string
	Time     string
	Request  string
	Response string
	TaskId   string `bson:"taskId"`
}
type TmpPageMonitResult struct {
	Url     string
	Content string
	TaskId  string `bson:"taskId"`
}
type PageMonitResult struct {
	ID      primitive.ObjectID `bson:"_id"`
	Url     string             `bson:"url"`
	Content []string           `bson:"content"`
	Hash    []string           `bson:"hash"`
	Diff    []string           `bson:"diff"`
	State   int                `bson:"state"`
	Project string             `bson:"project"`
	Time    string             `bson:"time"`
	TaskId  string             `bson:"taskId"`
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
	SubdomainScan                 bool `bson:"subdomainScan"`
	DirScanNotification           bool `bson:"dirScanNotification"`
	PortScanNotification          bool `bson:"portScanNotification"`
	SensitiveNotification         bool `bson:"sensitiveNotification"`
	SubdomainNotification         bool `bson:"subdomainNotification"`
	SubdomainTakeoverNotification bool `bson:"subdomainTakeoverNotification"`
	PageMonNotification           bool `bson:"pageMonNotification"`
	VulNotification               bool `bson:"vulNotification"`
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
}
type DomainResolve struct {
	Domain string
	IP     []string
}

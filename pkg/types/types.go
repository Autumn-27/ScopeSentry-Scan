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
	Host    string
	Type    string
	Value   []string
	IP      []string
	Time    string
	Project string
}

type AssetHttp struct {
	Timestamp     string                 `json:"timestamp,omitempty" csv:"timestamp"`
	TLSData       *clients.Response      `json:"tls,omitempty" csv:"tls"`
	Hashes        map[string]interface{} `json:"hash,omitempty" csv:"hash"`
	CDNName       string                 `json:"cdn_name,omitempty" csv:"cdn_name"`
	Port          string                 `json:"port,omitempty" csv:"port"`
	URL           string                 `json:"url,omitempty" csv:"url"`
	Title         string                 `json:"title,omitempty" csv:"title"`
	Type          string                 `json:"Type,omitempty" csv:"Type"`
	Error         string                 `json:"error,omitempty" csv:"error"`
	ResponseBody  string                 `json:"body,omitempty" csv:"body"`
	Host          string                 `json:"host,omitempty" csv:"host"`
	FavIconMMH3   string                 `json:"favicon,omitempty" csv:"favicon"`
	FaviconPath   string                 `json:"favicon_path,omitempty" csv:"favicon_path"`
	RawHeaders    string                 `json:"raw_header,omitempty" csv:"raw_header"`
	Jarm          string                 `json:"jarm,omitempty" csv:"jarm"`
	Technologies  []string               `json:"tech,omitempty" csv:"tech"`
	StatusCode    int                    `json:"status_code,omitempty" csv:"status_code"`
	ContentLength int                    `json:"content_length,omitempty" csv:"content_length"`
	CDN           bool                   `json:"cdn,omitempty" csv:"cdn"`
	Webcheck      bool                   `json:"webcheck,omitempty" csv:"webcheck"`
	Project       string                 `json:"project,omitempty" csv:"project"`
	WebFinger     []string               `json:"web_finger,omitempty" csv:"web_finger"`
	IconContent   string                 `json:"iconContent"`
	Domain        string                 `json:"domain"`
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
	Timestamp string          `json:"timestamp,omitempty" csv:"timestamp"`
	Host      string          `json:"host,omitempty"`
	IP        string          `json:"ip"`
	Port      string          `json:"port"`
	Protocol  string          `json:"protocol"`
	TLS       bool            `json:"tls"`
	Transport string          `json:"transport"`
	Version   string          `json:"version,omitempty"`
	Raw       json.RawMessage `json:"metadata"`
	Project   string          `json:"project"`
	Type      string
}

type UrlResult struct {
	Input      string `json:"input"`
	Source     string `json:"source"`
	OutputType string `json:"type"`
	Output     string `json:"output"`
	StatusCode int    `json:"status"`
	Length     int    `json:"length"`
	Time       string `json:"time"`
	Project    string
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
	Input    string
	Value    string
	Cname    string
	Response string
	Project  string
}

type DirResult struct {
	Url     string
	Status  int
	Msg     string
	Project string
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
}
type TmpPageMonitResult struct {
	Url     string
	Content string
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
}
type WebFinger struct {
	ID      string
	Express []string
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

// task-------------------------------------
// @file      : types.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/9 21:31
// -------------------------------------------

package options

import "sync"

type TaskOptions struct {
	ID                  string                       //任务ID
	TaskName            string                       `bson:"TaskName" json:"TaskName"`                       // 任务名称
	Target              string                       `bson:"target" json:"target"`                           //目标
	Type                string                       `bson:"type" json:"type"`                               //任务类型
	IsStart             bool                         `bson:"isStart" json:"isStart"`                         // 是否为暂停后开始
	TargetHandler       []string                     `bson:"TargetHandler" json:"TargetHandler"`             // 目标解析模块
	SubdomainScan       []string                     `bson:"SubdomainScan" json:"SubdomainScan"`             // 子域名扫描模块
	SubdomainSecurity   []string                     `bson:"SubdomainSecurity" json:"SubdomainSecurity"`     // 子域名安全检测模块
	AssetMapping        []string                     `bson:"AssetMapping" json:"AssetMapping"`               // 资产测绘模块
	AssetHandle         []string                     `bson:"AssetHandle" json:"AssetHandle"`                 // 资产处理模块
	PortScanPreparation []string                     `bson:"PortScanPreparation" json:"PortScanPreparation"` // 端口扫描预处理模块
	PortScan            []string                     `bson:"PortScan" json:"PortScan"`                       // 端口扫描模块
	PortFingerprint     []string                     `bson:"PortFingerprint" json:"PortFingerprint"`         // 端口指纹识别模块
	URLScan             []string                     `bson:"URLScan" json:"URLScan"`                         // URL扫描模块
	URLSecurity         []string                     `bson:"URLSecurity" json:"URLSecurity"`                 // URL安全检测模块
	WebCrawler          []string                     `bson:"WebCrawler" json:"WebCrawler"`                   // 爬虫模块
	DirScan             []string                     `bson:"DirScan" json:"DirScan"`                         //目录扫描模块
	VulnerabilityScan   []string                     `bson:"VulnerabilityScan" json:"VulnerabilityScan"`     //漏洞扫描模块
	PassiveScan         []string                     `bson:"PassiveScan" json:"PassiveScan"`                 // 被动扫描模块
	Parameters          map[string]map[string]string `bson:"Parameters" json:"Parameters"`                   // 各个插件的参数
	IsRestart           bool                         // 是否为重启后从本地获取缓存中获取的目标
	Duplicates          string                       `bson:"duplicates" json:"duplicates"` // 是否忽略已经存储在mongodb中的子域名
	InputChan           map[string]chan interface{}  // 每个模块的输入
	ModuleRunWg         *sync.WaitGroup              // 总的WaitGroup
	SubdomainFilename   string                       // 子域名扫描字典
	ProtRangeId         string                       // 端口范围在数据库中的id
	PortRange           string                       // 端口范围
}

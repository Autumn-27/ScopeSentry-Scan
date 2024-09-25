// task-------------------------------------
// @file      : types.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/9 21:31
// -------------------------------------------

package options

import "sync"

type TaskOptions struct {
	ID                  string                            //任务ID
	TaskName            string                            // 任务名称
	Target              string                            //目标
	TargetParser        []string                          // 目标解析模块
	SubdomainScan       []string                          // 子域名扫描模块
	SubdomainSecurity   []string                          // 子域名安全检测模块
	AssetMapping        []string                          // 资产测绘模块
	AssetHandle         []string                          // 资产处理模块
	PortScanPreparation []string                          // 端口扫描预处理模块
	PortScan            []string                          // 端口扫描模块
	URLScan             []string                          // URL扫描模块
	URLSecurity         []string                          // URL安全检测模块
	WebCrawler          []string                          // 爬虫模块
	VulnerabilityScan   []string                          //漏洞扫描模块
	Parameters          map[string]map[string]interface{} // 各个插件的参数
	IsRestart           bool                              // 是否为重启后从本地获取缓存中获取的目标
	IgnoreOldSubdomains bool                              // 是否忽略已经存储在mongodb中的子域名
	InputChan           map[string]chan interface{}       // 每个模块的输入
	ModuleRunWg         *sync.WaitGroup                   // 总的WaitGroup
	SubdomainFilename   string                            // 子域名扫描字典
	ProtRangeId         string                            // 端口范围在数据库中的id
	PortRange           string                            // 端口范围
}

// task-------------------------------------
// @file      : types.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/9 21:31
// -------------------------------------------

package task

type Options struct {
	ID                    string
	TaskName              string
	Target                string
	SubdomainScan         []string
	SubdomainResultHandle []string
	AssetMapping          []string
	PortScan              []string
	AssetResultHandle     []string
	URLScan               []string
	URLScanResultHandle   []string
	WebCrawler            []string
	VulnerabilityScan     []string
}

// task-------------------------------------
// @file      : types.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/9 21:31
// -------------------------------------------

package options

type TaskOptions struct {
	ID                string
	TaskName          string
	Target            string
	TargetParser      []string
	SubdomainScan     []string
	SubdomainSecurity []string
	AssetMapping      []string
	AssetHandle       []string
	PortScan          []string
	URLScan           []string
	URLSecurity       []string
	WebCrawler        []string
	VulnerabilityScan []string
	Parameters        map[string]map[string]interface{}
}

// modules-------------------------------------
// @file      : manage.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 21:51
// -------------------------------------------

package modules

import (
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/assethandle"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/assetmapping"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/portscan"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/subdomain"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/subdomainsecurity"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/targethandler"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlscan"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlsecurity"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/vulnerabilityscan"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/webcrawler"
)

func CreateScanProcess(op options.TaskOptions) interfaces.ModuleRunner {
	vulnerabilityModule := vulnerabilityscan.NewRunner(&op, nil)
	webCrawlerModule := webcrawler.NewRunner(&op, vulnerabilityModule)
	urlSecurityModule := urlsecurity.NewRunner(&op, webCrawlerModule)
	urlScanModule := urlscan.NewRunner(&op, urlSecurityModule)
	assetHandleModule := assethandle.NewRunner(&op, urlScanModule)
	assetMappingModule := assetmapping.NewRunner(&op, assetHandleModule)
	portscanModule := portscan.NewRunner(&op, assetMappingModule)
	subdomainSecurityModule := subdomainsecurity.NewRunner(&op, portscanModule)
	subdomainModule := subdomain.NewRunner(&op, subdomainSecurityModule)
	targetHandlerModule := targethandler.NewRunner(&op, subdomainModule)
	return targetHandlerModule
}

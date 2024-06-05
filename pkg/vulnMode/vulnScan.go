// vulnMode-------------------------------------
// @file      : vulnScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/12 21:14
// -------------------------------------------

package vulnMode

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	nuclei "github.com/projectdiscovery/nuclei/v3/lib"
	"github.com/projectdiscovery/nuclei/v3/pkg/catalog/config"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"
	"os"
	"path/filepath"
	"strings"
)

func Scan(target []string, template []string) {
	defer system.RecoverPanic("vulnMode")
	_, err := os.Stat(filepath.Join(system.ConfigDir, "/poc"))
	if err != nil {
		system.UpdatePoc(true)
	}
	config.DefaultConfig.DisableUpdateCheck()
	config.DefaultConfig.TemplatesDirectory = filepath.Join(system.ConfigDir, "/poc")
	ne, err := nuclei.NewNucleiEngine(
		nuclei.WithTemplatesOrWorkflows(nuclei.TemplateSources{Templates: template}),
		nuclei.WithTemplateUpdateCallback(true, func(newVersion string) {}),
	)
	var vulResults []types.VulnResult
	if err != nil {
		system.SlogError(fmt.Sprintf("Nuclei to err: %s", err))
	}
	callBackFunc := func(event *output.ResultEvent) {
		vulName := system.PocList[strings.TrimSuffix(filepath.Base(event.TemplatePath), ".yaml")].Name
		level := system.PocList[strings.TrimSuffix(filepath.Base(event.TemplatePath), ".yaml")].Level
		if vulName == "" {
			vulName = event.Info.Name
		}
		if level == "" {
			level = "unknown"
		}
		tmpResult := types.VulnResult{
			Url:      event.URL,
			VulnId:   strings.TrimSuffix(filepath.Base(event.TemplatePath), ".yaml"),
			VulName:  vulName,
			Matched:  event.Matched,
			Level:    level,
			Time:     system.GetTimeNow(),
			Request:  event.Request,
			Response: event.Response,
		}
		vulResults = append(vulResults, tmpResult)
	}
	// load targets and optionally probe non http/https targets
	ne.LoadTargets(target, false)
	err = ne.ExecuteWithCallback(callBackFunc)
	if err != nil {
		system.SlogError(fmt.Sprintf("Nuclei to 2err: %s", err))
	}
	defer ne.Close()
	defer scanResult.VulnResult(vulResults)

}

//func Scan2(target []string, template []string) {
//	config.DefaultConfig.DisableUpdateCheck()
//options := &nucleiTypes.Options{
//	RateLimit:                     10,
//	BulkSize:                      10,
//	TemplateThreads:               10,
//	HeadlessBulkSize:              10,
//	HeadlessTemplateThreads:       10,
//	Timeout:                       5,
//	Retries:                       1,
//	MaxHostError:                  2,
//	NoColor:                       true,
//	Validate:                      false,
//	UpdateTemplates:               false,
//	Debug:                         false,
//	Verbose:                       false,
//	EnableProgressBar:             false,
//	Silent:                        true,
//	Headless:                      false,
//	PublicTemplateDisableDownload: true,
//	CustomHeaders:                 []string{"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.64"},
//	VerboseVerbose:                false,
//}

//}

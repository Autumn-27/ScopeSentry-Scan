// sensitiveMode-------------------------------------
// @file      : sensitiveScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/11 19:55
// -------------------------------------------

package sensitiveMode

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"github.com/dlclark/regexp2"
)

func Scan(url string, resp string, resMd5 string, taskId string) {
	defer system.RecoverPanic("sensitiveMode")
	if len(system.SensitiveRules) == 0 {
		system.UpdateSensitive()
	}
	chunkSize := 5120
	overlapSize := 100
	NotificationMsg := "SensitiveScan Result:\n"
	//allstart := time.Now()
	findFlag := false
	for _, rule := range system.SensitiveRules {
		//start := time.Now()
		if rule.State {
			r, err := regexp2.Compile(rule.Regular, 0)
			if err != nil {
				system.SlogError(fmt.Sprintf("Error compiling sensitive regex pattern: %s - %s", err, rule.ID))
				continue
			}
			resultChan, errorChan := processInChunks(r, resp, chunkSize, overlapSize)
			for matches := range resultChan {
				if len(matches) != 0 {
					var tmpResult types.SensitiveResult
					if findFlag {
						tmpResult = types.SensitiveResult{Url: url, SID: rule.Name, Match: matches, Body: "", Time: system.GetTimeNow(), Color: rule.Color, Md5: fmt.Sprintf("md5==%v", resMd5)}
					} else {
						tmpResult = types.SensitiveResult{Url: url, SID: rule.Name, Match: matches, Body: resp, Time: system.GetTimeNow(), Color: rule.Color, Md5: resMd5}
					}
					scanResult.SensitiveResult([]types.SensitiveResult{tmpResult}, taskId)
					NotificationMsg += fmt.Sprintf("%v\n%v:%v", url, rule.Name, matches)
					findFlag = true
				}
			}
			if err := <-errorChan; err != nil {
				system.SlogError(fmt.Sprintf("Error processing chunks: %s", err))
			}
		}
		//elapsed := time.Since(start)
		//if elapsed > 1*time.Second {
		//	fmt.Printf("%s  Regex performance: %s\n", rule.Name, elapsed)
		//}
	}
	//allelapsed := time.Since(allstart)
	//fmt.Printf("all Regex performance: %s\n", allelapsed)
	if system.NotificationConfig.SensitiveNotification && len(NotificationMsg) > 25 {
		go system.SendNotification(NotificationMsg)
	}
}

func processInChunks(regex *regexp2.Regexp, text string, chunkSize int, overlapSize int) (chan []string, chan error) {
	resultChan := make(chan []string, 10)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)
		for start := 0; start < len(text); start += chunkSize {
			end := start + chunkSize
			if end > len(text) {
				end = len(text)
			}

			chunkEnd := end
			if end+overlapSize < len(text) {
				chunkEnd = end + overlapSize
			}

			matches, err := findMatchesInChunk(regex, text[start:chunkEnd])
			if err != nil {
				errorChan <- err
				return
			}

			if len(matches) > 0 {
				resultChan <- matches
			}
		}
	}()

	return resultChan, errorChan
}

func findMatchesInChunk(regex *regexp2.Regexp, text string) ([]string, error) {
	var matches []string
	m, _ := regex.FindStringMatch(text)
	for m != nil {
		matches = append(matches, m.String())
		m, _ = regex.FindNextMatch(m)
	}
	return matches, nil
}

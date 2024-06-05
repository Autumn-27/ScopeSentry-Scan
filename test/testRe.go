// main-------------------------------------
// @file      : testRe.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/5/27 18:56
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"github.com/dlclark/regexp2"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

func main() {
	go func() {
		_ = http.ListenAndServe("0.0.0.0:6060", nil)
	}()
	testMsg := ""
	system.InitDb()
	system.UpdateSensitive()
	var wg sync.WaitGroup
	fmt.Println("begin test")
	for i := 1; i <= 1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			go Scan2("ddd", testMsg)
		}()
	}
	time.Sleep(5 * time.Second)
	wg.Wait()
}

func Scan(url string, resp string) {
	defer system.RecoverPanic("sensitiveMode")
	if len(system.SensitiveRules) == 0 {
		system.UpdateSensitive()
	}
	chunkSize := 5120
	overlapSize := 100
	NotificationMsg := "SensitiveScan Result:\n"
	allstart := time.Now()
	Sresults := []types.SensitiveResult{}
	for _, rule := range system.SensitiveRules {
		start := time.Now()
		if rule.State {
			r, err := regexp2.Compile(rule.Regular, 0)
			if err != nil {
				system.SlogError(fmt.Sprintf("Error compiling sensitive regex pattern: %s - %s", err, rule.ID))
				continue
			}
			resultChan, errorChan := processInChunks(r, resp, chunkSize, overlapSize)
			for matches := range resultChan {
				if len(matches) != 0 {
					tmpResult := types.SensitiveResult{Url: url, SID: rule.ID, Match: matches, Body: resp, Time: system.GetTimeNow(), Color: rule.Color}
					Sresults = append(Sresults, tmpResult)
					NotificationMsg += fmt.Sprintf("%v\n%v:%v", url, rule.Name, matches)
				}
			}
			if err := <-errorChan; err != nil {
				system.SlogError(fmt.Sprintf("Error processing chunks: %s", err))
			}
		}
		elapsed := time.Since(start)
		fmt.Printf("%s  Regex performance: %s\n", rule.Name, elapsed)
	}
	time.Sleep(5 * time.Second)
	allelapsed := time.Since(allstart)
	fmt.Printf("all Regex performance: %s\n", allelapsed)
}

func Scan2(url string, resp string) {
	defer system.RecoverPanic("sensitiveMode")
	if len(system.SensitiveRules) == 0 {
		system.UpdateSensitive()
	}
	chunkSize := 5120
	overlapSize := 100
	NotificationMsg := "SensitiveScan Result:\n"
	Sresults := []types.SensitiveResult{}
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	mu := sync.Mutex{}
	allstart := time.Now()
	for _, rule := range system.SensitiveRules {
		if rule.State {
			wg.Add(1)
			go func(rule types.SensitiveRule) {

				defer wg.Done()
				sem <- struct{}{} // 获取一个令牌
				defer func() {
					<-sem
				}() // 释放一个令牌
				start := time.Now()
				r, err := regexp2.Compile(rule.Regular, 0)
				if err != nil {
					system.SlogError(fmt.Sprintf("Error compiling sensitive regex pattern: %s - %s", err, rule.ID))
					return
				}
				resultChan, errorChan := processInChunks(r, resp, chunkSize, overlapSize)
				for matches := range resultChan {
					if len(matches) != 0 {
						tmpResult := types.SensitiveResult{Url: url, SID: rule.ID, Match: matches, Body: resp, Time: system.GetTimeNow(), Color: rule.Color}
						mu.Lock()
						Sresults = append(Sresults, tmpResult)
						NotificationMsg += fmt.Sprintf("%v\n%v:%v", url, rule.Name, matches)
						mu.Unlock()
					}
				}
				if err := <-errorChan; err != nil {
					system.SlogError(fmt.Sprintf("Error processing chunks: %s", err))
				}
				elapsed := time.Since(start)
				fmt.Printf("%s  Regex performance: %s\n", rule.Name, elapsed)
			}(rule)
		}
	}
	wg.Wait()
	time.Sleep(5 * time.Second)
	allelapsed := time.Since(allstart)
	fmt.Printf("all Regex performance: %s\n", allelapsed)
}
func processInChunks(regex *regexp2.Regexp, text string, chunkSize int, overlapSize int) (chan []string, chan error) {
	resultChan := make(chan []string, 100)
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

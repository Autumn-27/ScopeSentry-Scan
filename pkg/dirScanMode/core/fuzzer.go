// Package core -----------------------------
// @file      : Fuzzer.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/28 23:00
// -------------------------------------------
package core

import (
	"bufio"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type Fuzzer struct {
	Dictionary         string
	Threads            int
	BasePath           string
	Scanners           map[string]map[string]*Scanner
	Request            Request
	Options            Options
	MaxSameLen         int
	Mu                 sync.Mutex
	RCLMu              sync.Mutex
	ResponseCodeLength map[string]int
}

func (f *Fuzzer) Start() {
	err := f.SetupScanners()
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, f.Options.Thread)
	flag := 0
	var mu sync.Mutex
	file, err := os.Open(f.Dictionary) // 打开文件
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		path := scanner.Text()
		semaphore <- struct{}{}
		wg.Add(1)
		if flag >= MaxRetries {
			return
		}
		go func(path string, flag *int) {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			mu.Lock()
			if *flag >= MaxRetries {
				mu.Unlock()
				return
			}
			mu.Unlock()
			scanners := f.GetScannersFor(path)
			err := f.Scan(f.BasePath+path, scanners)
			if err != nil {
				mu.Lock()
				*flag += 1
				if *flag >= MaxRetries {
					mu.Unlock()
					return
				}
				mu.Unlock()
				//if strings.Contains(fmt.Sprintf("%v", err), "timed out") || strings.Contains(fmt.Sprintf("%v", err), "the server closed connection") {
				//	mu.Lock()
				//	*flag += 1
				//	if *flag >= MaxRetries {
				//		mu.Unlock()
				//		return
				//	}
				//	mu.Unlock()
				//}
			} else {
				mu.Lock()
				if *flag > 0 {
					*flag -= 1
				}
				mu.Unlock()
			}
		}(path, &flag)
	}
	time.Sleep(time.Second * 5)
	wg.Wait()
}

func (f *Fuzzer) SetupScanners() error {
	scanner, err := (&Scanner{Request: f.Request, Path: f.BasePath}).SetUp()
	if err != nil {
		return err
	}
	f.Scanners = make(map[string]map[string]*Scanner)
	f.Scanners["default"] = make(map[string]*Scanner)
	f.Scanners["default"]["index"] = scanner
	scanner, err = (&Scanner{Request: f.Request, Path: f.BasePath + PlaceholderMarkers}).SetUp()
	if err != nil {
		return err
	}
	f.Scanners["default"]["random"] = scanner
	f.Scanners["suffixes"] = make(map[string]*Scanner)
	for _, suff := range f.Options.Extensions {
		scanner, err = (&Scanner{Request: f.Request, Path: f.BasePath + PlaceholderMarkers + "." + suff}).SetUp()
		if err != nil {
			return err
		}
		f.Scanners["suffixes"][suff] = scanner
	}
	return nil
}

func (f *Fuzzer) Scan(path string, scanners []*Scanner) error {
	response, err := f.Request.Request(path)
	if err != nil {
		return err
	}
	if f.IsExcluded(response) {
		return nil
	}

	for _, scanner := range scanners {
		if !scanner.Check(path, response, &f.MaxSameLen, &f.Mu) {
			return nil
		}
	}
	key := fmt.Sprintf("%d:%d", response.StatusCode, response.ContentLength)
	if f.ResponseCodeLength[key] > 20 {
		return nil
	}
	f.Options.MatchCallback(response)
	f.RCLMu.Lock()
	f.ResponseCodeLength[key]++
	f.RCLMu.Unlock()
	return nil
}

func (f *Fuzzer) GetScannersFor(path string) []*Scanner {
	path = CleanPath(path)
	var scanners []*Scanner
	for suffix, scanner := range f.Scanners["suffixes"] {
		if strings.HasSuffix(path, suffix) {
			scanners = append(scanners, scanner)
		}
	}
	for _, scanner := range f.Scanners["default"] {
		scanners = append(scanners, scanner)
	}
	return scanners
}

func (f *Fuzzer) IsExcluded(response types.HttpResponse) bool {
	index := sort.SearchInts(f.Options.IncludeStatusCodes, response.StatusCode)

	// 判断目标字符串是否在数组中
	if index < len(f.Options.IncludeStatusCodes) && f.Options.IncludeStatusCodes[index] == response.StatusCode {
		return false
	} else {
		return true
	}
}

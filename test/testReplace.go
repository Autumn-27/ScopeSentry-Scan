// Package main -----------------------------
// @file      : testReplace.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/5/28 13:33
// -------------------------------------------
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"strings"
	"sync"
	"time"
)

func main() {
	s := ""
	go func() {
		_ = http.ListenAndServe("0.0.0.0:6060", nil)
	}()
	var wg sync.WaitGroup
	fmt.Println("begin test")
	for i := 1; i <= 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			go DecodeChars2(s)
		}()
	}
	time.Sleep(5 * time.Second)
	wg.Wait()
}

func DecodeChars(s string) string {
	source, err := url.QueryUnescape(s)
	if err == nil {
		s = source
	}

	// In case json encoded chars
	replacer := strings.NewReplacer(
		`\u002f`, "/",
		`\u0026`, "&",
	)
	s = replacer.Replace(s)
	return s
}

func DecodeChars2(s string) string {
	source, err := url.QueryUnescape(s)
	if err == nil {
		s = source
	}

	s = strings.Replace(s, `\u002f`, "/", -1)
	s = strings.Replace(s, `\u0026`, "&", -1)
	return s
}

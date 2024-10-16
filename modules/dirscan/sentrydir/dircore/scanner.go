// Package core -----------------------------
// @file      : scanner.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/28 23:24
// -------------------------------------------
package dircore

import (
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/dirscan/sentrydir/dirutils"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

type Scanner struct {
	Path                  string
	Request               Request
	Tested                map[string]map[string]Scanner
	Response              types.HttpResponse
	WildcardRedirectRegex string
	ContentParser         *DynamicContentParser
}

func (s *Scanner) SetUp() (*Scanner, error) {
	firstPath := strings.Replace(s.Path, PlaceholderMarkers, dirutils.RandomString(7), -1)
	firstResponse, err := s.Request.Request(firstPath)
	if err != nil {
		return s, err
	}
	s.Response = firstResponse
	secondPath := strings.Replace(s.Path, PlaceholderMarkers, dirutils.RandomString(7), -1)
	secondResponse, err := s.Request.Request(secondPath)
	if err != nil {
		return s, err
	}
	s.WildcardRedirectRegex = ""
	if firstResponse.Redirect != "" && secondResponse.Redirect != "" {
		s.WildcardRedirectRegex = generateRedirectRegex(CleanPath(firstResponse.Redirect), firstPath, CleanPath(secondResponse.Redirect), secondPath)
	}
	s.ContentParser = NewDynamicContentParser(firstResponse.Body, secondResponse.Body)
	return s, nil
}

func (s *Scanner) Check(path string, response types.HttpResponse, maxSameLen *int, mu *sync.Mutex) bool {
	if s.Response.StatusCode != response.StatusCode {
		return true
	}
	if s.WildcardRedirectRegex != "" && response.Redirect != "" {
		path = Unquote(CleanPath(path))
		redirect := Unquote(CleanPath(response.Redirect))
		regexToCompare := strings.ReplaceAll(s.WildcardRedirectRegex, ReplaceMarkers, regexp.QuoteMeta(path))
		isWildcardRedirect, _ := regexp.MatchString(regexToCompare, redirect)
		if !isWildcardRedirect {
			return true
		} else {
			return false
		}
	}
	if len(response.Body) == len(s.Response.Body) {
		if *maxSameLen <= 0 {
			return false
		}
		mu.Lock()
		*maxSameLen -= 1
		defer mu.Unlock()
	}
	if s.IsWildcard(response) {
		return false
	}
	return true
}

func (s *Scanner) IsWildcard(response types.HttpResponse) bool {
	if response.Body == "" && s.Response.Body == "" {
		return response.Body == s.Response.Body
	}

	return s.ContentParser.CompareTo(response.Body)
}

func CleanPath(path string) string {
	path = strings.SplitN(path, "#", 2)[0]
	path = strings.SplitN(path, "?", 2)[0]
	return path
}

func Unquote(s string) string {
	unquoted, err := url.QueryUnescape(s)
	if err != nil {
		// 如果解码出现错误，返回原始数据
		return s
	}
	return unquoted
}

func generateRedirectRegex(firstLoc string, firstPath string, secondLoc string, secondPath string) string {
	if firstPath != "" {
		decodedURL, _ := url.QueryUnescape(firstLoc)
		if decodedURL != "" {
			firstLoc = strings.Replace(decodedURL, firstPath, ReplaceMarkers, -1)
		}
	}
	if secondPath != "" {
		decodedURL, _ := url.QueryUnescape(secondLoc)
		if decodedURL != "" {
			secondLoc = strings.Replace(decodedURL, secondPath, ReplaceMarkers, -1)
		}
	}
	return generateMatchingRegex(firstLoc, secondLoc)
}

func generateMatchingRegex(string1 string, string2 string) string {
	start := "^"
	end := "$"
	var minLen int
	if len(string1) < len(string2) {
		minLen = len(string1)
	} else {
		minLen = len(string2)
	}
	for i := 0; i < minLen; i++ {
		if string1[i] != string2[i] {
			start += ".*"
			break
		}
		start += regexp.QuoteMeta(string1[i : i+1])
	}

	if len(start) > 1 && start[len(start)-2:] == ".*" {
		for i := minLen - 1; i >= 0; i-- {
			if string1[i] != string2[i] {
				break
			}
			end = regexp.QuoteMeta(string1[i:i+1]) + end
		}
	}
	return start + end
}

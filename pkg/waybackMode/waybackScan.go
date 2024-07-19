// waybackMode-------------------------------------
// @file      : waybackScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/11 21:06
// -------------------------------------------

package waybackMode

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

func Runner(domains []string, taskId string) {
	var urlInfos []types.UrlResult
	if len(domains) == 0 {
		system.SlogInfo("target waybackMode in null")
		return
	}
	for _, u := range domains {
		system.SlogInfo(fmt.Sprintf("waybackMode runner target %v", u))
		urlWithoutHTTP := strings.TrimPrefix(u, "http://")
		urlWithoutHTTPS := strings.TrimPrefix(urlWithoutHTTP, "https://")
		waybackResult := Scan(urlWithoutHTTPS)
		system.SlogInfo(fmt.Sprintf("target %v waybackMode get %v result", u, len(waybackResult)))
		nowTime := system.GetTimeNow()
		for _, wurl := range waybackResult {
			urlInfos = append(urlInfos, types.UrlResult{Input: u, Source: "waybackurls", Output: wurl, Time: nowTime})
		}
	}
	var Wg sync.WaitGroup
	tmpseenUrls := make(map[string]struct{})
	for _, result := range urlInfos {
		if _, seen := tmpseenUrls[result.Output]; seen {
			continue
		}
		flag := scanResult.URLRedisDeduplication(result.Output, taskId)
		if flag {
			continue
		}
		Wg.Add(1)
		go func(result types.UrlResult) {
			defer Wg.Done()
			scanResult.UrlResult([]types.UrlResult{result}, taskId)
		}(result)
		tmpseenUrls[result.Output] = struct{}{}
	}
	Wg.Wait()
	system.SlogInfo(fmt.Sprintf("target[0] %v waybackMode get %v unique result", domains[0], len(tmpseenUrls)))
}
func Scan(domain string) []string {
	defer system.RecoverPanic("waybackMode")
	var result []string
	var mu sync.Mutex
	var noSubs bool
	noSubs = false
	fetchFns := []fetchFn{
		getWaybackURLs,
		getCommonCrawlURLs,
	}
	//system.SlogDebugLocal(fmt.Sprintf("waybackMode runner target %v", len(domains)))
	wurls := make(chan string)

	var wg sync.WaitGroup

	for _, fn := range fetchFns {
		fetch := fn
		resp, err := fetch(domain, noSubs)
		if err != nil {
			system.SlogDebugLocal(fmt.Sprintf("waybackMode %s error: %v", domain, err))
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			const maxLines = 1000
			lineCount := 0
			for _, r := range resp {
				lineCount++
				if lineCount >= maxLines {
					return
				}
				if noSubs && isSubdomain(r, domain) {
					continue
				}
				wurls <- r
			}
		}()
	}

	go func() {
		wg.Wait()
		close(wurls)
	}()

	seen := make(map[string]struct{})
	for w := range wurls {
		mu.Lock()
		if _, ok := seen[w]; !ok {
			seen[w] = struct{}{}
			if w != "" {
				result = append(result, w)
			}
		}
		mu.Unlock()
	}
	//system.SlogDebugLocal(fmt.Sprintf("waybackMode get %v result", len(result)))
	return result
}

type fetchFn func(string, bool) ([]string, error)

func getWaybackURLs(domain string, noSubs bool) ([]string, error) {
	subsWildcard := "*."
	if noSubs {
		subsWildcard = ""
	}

	//res, err := http.Get(
	//	fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s%s/*&output=json&collapse=urlkey", subsWildcard, domain),
	//)
	//res, err := util.HttpGetByte(fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s%s/*&output=json&collapse=urlkey", subsWildcard, domain))
	//if err != nil {
	//	return []wurl{}, err
	//}
	//
	//var wrapper [][]string
	//err = json.Unmarshal(res, &wrapper)
	res, err := http.Get(
		fmt.Sprintf("https://web.archive.org/cdx/search/cdx?url=%s%s/*&output=text&collapse=urlkey&fl=original", subsWildcard, domain),
	)
	//res, err := util.HttpGetByte(fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s%s/*&output=text&collapse=urlkey&fl=original", subsWildcard, domain))
	if err != nil {
		return []string{}, err
	}
	if res.StatusCode != 200 {
		return []string{}, nil
	}
	defer res.Body.Close()
	sc := bufio.NewScanner(res.Body)
	buffseSize := 4 * 1024
	buf := make([]byte, buffseSize)
	sc.Buffer(buf, buffseSize)
	const maxLines = 1000
	var out []string
	lineCount := 0
	for sc.Scan() {
		out = append(out, sc.Text())
		lineCount++
		if lineCount >= maxLines {
			break
		}
	}
	return out, nil
}

func getCommonCrawlURLs(domain string, noSubs bool) ([]string, error) {
	subsWildcard := "*."
	if noSubs {
		subsWildcard = ""
	}

	res, err := http.Get(
		fmt.Sprintf("https://index.commoncrawl.org/CC-MAIN-2024-18-index?url=%s%s/*&output=text&fl=url", subsWildcard, domain),
	)
	if err != nil {
		return []string{}, err
	}
	if res.StatusCode != 200 {
		return []string{}, nil
	}

	defer res.Body.Close()
	sc := bufio.NewScanner(res.Body)

	const maxLines = 1000
	buffseSize := 4 * 1024
	buf := make([]byte, buffseSize)
	sc.Buffer(buf, buffseSize)
	var out []string
	lineCount := 0
	for sc.Scan() {
		out = append(out, sc.Text())
		lineCount++
		if lineCount >= maxLines {
			break
		}
	}
	return out, nil
}

//func getVirusTotalURLs(domain string, noSubs bool) ([]wurl, error) {
//	out := make([]wurl, 0)
//
//	apiKey := os.Getenv("VT_API_KEY")
//	if apiKey == "" {
//		// no API key isn't an error,
//		// just don't fetch
//		return out, nil
//	}
//
//	fetchURL := fmt.Sprintf(
//		"https://www.virustotal.com/vtapi/v2/domain/report?apikey=%s&domain=%s",
//		apiKey,
//		domain,
//	)
//
//	resp, err := http.Get(fetchURL)
//	if err != nil {
//		return out, err
//	}
//	defer resp.Body.Close()
//
//	wrapper := struct {
//		URLs []struct {
//			URL string `json:"url"`
//			// TODO: handle VT date format (2018-03-26 09:22:43)
//			//Date string `json:"scan_date"`
//		} `json:"detected_urls"`
//	}{}
//
//	dec := json.NewDecoder(resp.Body)
//
//	err = dec.Decode(&wrapper)
//
//	for _, u := range wrapper.URLs {
//		out = append(out, wurl{url: u.URL})
//	}
//
//	return out, nil
//
//}

func isSubdomain(rawUrl, domain string) bool {
	u, err := url.Parse(rawUrl)
	if err != nil {
		// we can't parse the URL so just
		// err on the side of including it in output
		return false
	}

	return strings.ToLower(u.Hostname()) != strings.ToLower(domain)
}

func getVersions(u string) ([]string, error) {
	out := make([]string, 0)

	resp, err := http.Get(fmt.Sprintf(
		"http://web.archive.org/cdx/search/cdx?url=%s&output=json", u,
	))

	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	r := [][]string{}

	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&r)
	if err != nil {
		return out, err
	}

	first := true
	seen := make(map[string]bool)
	for _, s := range r {

		// skip the first element, it's the field names
		if first {
			first = false
			continue
		}

		// fields: "urlkey", "timestamp", "original", "mimetype", "statuscode", "digest", "length"
		if seen[s[5]] {
			continue
		}
		seen[s[5]] = true
		out = append(out, fmt.Sprintf("https://web.archive.org/web/%sif_/%s", s[1], s[2]))
	}

	return out, nil
}

// source-------------------------------------
// @file      : waybackarchive.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/14 20:46
// -------------------------------------------

package source

import (
	"bufio"
	"fmt"
	"net/http"
	"time"
)

func WaybackarchiveRun(rootUrl string, result chan Result) int {
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置超时时间
	}
	res, err := client.Get(
		fmt.Sprintf("https://web.archive.org/cdx/search/cdx?url=*.%s/*&output=text&collapse=urlkey&fl=original", rootUrl),
	)
	//res, err := util.HttpGetByte(fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s%s/*&output=text&collapse=urlkey&fl=original", subsWildcard, domain))
	if err != nil {
		return 0
	}
	if res.StatusCode != 200 {
		return 0
	}
	defer res.Body.Close()
	sc := bufio.NewScanner(res.Body)
	buffseSize := 4 * 1024
	buf := make([]byte, buffseSize)
	sc.Buffer(buf, buffseSize)
	lineCount := 0
	for sc.Scan() {
		result <- Result{
			URL:    sc.Text(),
			Source: "waybackarchive",
		}
		lineCount++
	}
	return lineCount
}

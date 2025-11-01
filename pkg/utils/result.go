// utils-------------------------------------
// @file      : result.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/9 21:21
// -------------------------------------------

package utils

import (
	"bytes"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/cckuailong/simHtml/simHtml"
	"strconv"
	"strings"
)

type Result struct {
}

var Results *Result

func InitializeResults() {
	Results = &Result{}
}

func (r *Result) CompareAssetOther(old, new types.AssetOther) types.AssetChangeLog {
	var Change types.AssetChangeLog
	if old.TLS != new.TLS {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "TLS",
			Old:       fmt.Sprintf("%t", old.TLS),
			New:       fmt.Sprintf("%t", new.TLS),
		})
	}
	if old.IP != new.IP {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "IP",
			Old:       old.IP,
			New:       new.IP,
		})
	}
	if old.Service != new.Service {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "Service",
			Old:       old.Service,
			New:       new.Service,
		})
	}

	if old.Version != new.Version {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "Version",
			Old:       old.Version,
			New:       new.Version,
		})
	}
	if old.Transport != new.Transport {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "Transport",
			Old:       old.Transport,
			New:       new.Transport,
		})
	}
	if !bytes.Equal([]byte(old.Banner), []byte(new.Banner)) {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "Banner",
			Old:       old.Banner,
			New:       new.Banner,
		})
	}
	if len(Change.Change) != 0 {
		Change.Timestamp = new.Time
		return Change
	} else {
		return types.AssetChangeLog{}
	}
}

func compareTechnologies(tech1, tech2 []string) (changes string) {
	// 用来存储检查过的技术
	checked := make([]bool, len(tech1))
	change := ""
	// 查找增加
	for _, tech2Item := range tech2 {
		found := false
		for i, tech1Item := range tech1 {
			if strings.EqualFold(tech1Item, tech2Item) {
				found = true
				checked[i] = true // 标记为已检查
				break
			}
		}
		if !found {
			change += "+" + tech2Item + "\n"
		}
	}

	// 查找减少
	for i, tech1Item := range tech1 {
		if !checked[i] { // 只有未检查的才是减少的
			change += "-" + tech1Item + "\n"
		}
	}
	return strings.TrimSuffix(change, "\n")
}

func (r *Result) CompareAssetHttp(old, new types.AssetHttp) types.AssetChangeLog {
	var Change types.AssetChangeLog
	if old.StatusCode != new.StatusCode {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "StatusCode",
			Old:       strconv.Itoa(old.StatusCode),
			New:       strconv.Itoa(new.StatusCode),
		})
	}

	if old.Title != new.Title {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "Title",
			Old:       old.Title,
			New:       new.Title,
		})
	}
	if old.Service != new.Service {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "Service",
			Old:       old.Service,
			New:       new.Service,
		})
	}
	if old.IP != new.IP {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "IP",
			Old:       old.IP,
			New:       new.IP,
		})
	}
	if old.ContentLength != old.ContentLength {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "ContentLength",
			Old:       strconv.Itoa(old.ContentLength),
			New:       strconv.Itoa(new.ContentLength),
		})
	}
	if old.WebServer != new.WebServer {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "WebServer",
			Old:       old.WebServer,
			New:       new.WebServer,
		})
	}
	tecCompStr := compareTechnologies(old.Technologies, new.Technologies)
	if tecCompStr != "" {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "Technologies",
			Old:       "",
			New:       tecCompStr,
		})
	}
	if old.CDN != new.CDN {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "CDN",
			Old:       fmt.Sprintf("%t", old.CDN),
			New:       fmt.Sprintf("%t", new.CDN),
		})
	}
	if old.Screenshot != new.Screenshot {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "Screenshot",
			Old:       fmt.Sprintf("%v", old.Screenshot),
			New:       fmt.Sprintf("%v", new.Screenshot),
		})
	}
	if old.ResponseBodyHash != new.ResponseBodyHash {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "BodyHash",
			Old:       old.ResponseBodyHash,
			New:       new.ResponseBodyHash,
		})
		simNum := simHtml.GetSimFromStr(old.ResponseBody, new.ResponseBody)
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "DOM similarity",
			Old:       "",
			New:       fmt.Sprintf("%.2f%%", simNum*100),
		})
	}
	if len(Change.Change) != 0 {
		Change.Timestamp = new.Time
		return Change
	} else {
		return types.AssetChangeLog{}
	}
}

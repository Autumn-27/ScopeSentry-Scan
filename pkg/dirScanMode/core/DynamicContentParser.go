// Package core -----------------------------
// @file      : DynamicContentParser.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/5/7 11:45
// -------------------------------------------
package core

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/dirScanMode/utils"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type DynamicContentParser struct {
	StaticPatterns []string
	IsStatic       bool
	Differ         *diffmatchpatch.DiffMatchPatch
	BaseContent    string
}

var MaxMatchRatio = 0.9

func NewDynamicContentParser(content1 string, content2 string) *DynamicContentParser {
	tmp := &DynamicContentParser{
		BaseContent: content1,
	}
	if content1 == content2 {
		tmp.IsStatic = true
	} else {
		tmp.IsStatic = false
		tmp.Differ = diffmatchpatch.New()
		tmp.Differ.DiffTimeout = 0
		if content1 == "" || content2 == "" {
			tmp.StaticPatterns = []string{}
		} else {
			diffs := tmp.Differ.DiffMain(content1, content2, false)
			tmp.StaticPatterns = GetStaticPatterns(diffs)
		}
	}
	return tmp
}
func (d *DynamicContentParser) CompareTo(content string) bool {
	if d.IsStatic {
		return content == d.BaseContent
	}
	var contentStaticPatterns []string
	if content == "" || d.BaseContent == "" {
		contentStaticPatterns = []string{}
	} else {
		diffs := d.Differ.DiffMain(content, d.BaseContent, false)
		contentStaticPatterns = GetStaticPatterns(diffs)
	}
	if slicesEqual(contentStaticPatterns, d.StaticPatterns) {
		return true
	}
	sm := utils.NewSequenceMatcher(d.BaseContent, content)
	ration := sm.Ratio()
	return ration > MaxMatchRatio
}

func slicesEqual(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}
func GetStaticPatterns(diffs []diffmatchpatch.Diff) []string {
	var tmp []string
	for _, d := range diffs {
		formatType := fmt.Sprintf("%v", d.Type)
		if formatType == "Equal" {
			tmp = append(tmp, d.Text)
		}
	}
	return tmp
}

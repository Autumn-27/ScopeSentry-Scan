// utils-------------------------------------
// @file      : result.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/9 21:21
// -------------------------------------------

package utils

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"reflect"
)

type Result struct {
}

var Results *Result

func InitializeResults() {
	Results = &Result{}
}

func (r *Result) CompareAssetOther(a1, a2 types.AssetOther) map[string][]string {
	differences := make(map[string][]string)

	v1 := reflect.ValueOf(a1)
	v2 := reflect.ValueOf(a2)

	for i := 0; i < v1.NumField(); i++ {
		fieldName := v1.Type().Field(i).Name

		// 跳过 TaskId 字段
		if fieldName == "TaskId" {
			continue
		}

		val1 := v1.Field(i).Interface()
		val2 := v2.Field(i).Interface()

		if !reflect.DeepEqual(val1, val2) {
			differences[fieldName] = []string{
				fmt.Sprintf("a1: %v", val1),
				fmt.Sprintf("a2: %v", val2),
			}
		}
	}
	return differences
}

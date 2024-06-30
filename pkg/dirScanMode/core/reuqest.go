// Package core -----------------------------
// @file      : reuqest.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/28 23:51
// -------------------------------------------
package core

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
)

type Request struct {
	Url string
}

func (r *Request) Request(path string) (types.HttpResponse, error) {
	response, err := util.HttpGet(r.Url + path)
	if err != nil {
		for i := 0; i < MaxRetries-5; i++ {
			response, err = util.HttpGet(r.Url + path)
			if err != nil {
				system.SlogDebugLocal(fmt.Sprintf("dirScan request error: %s", err))
				continue
			}
			return response, nil
		}
	}
	return response, err
}

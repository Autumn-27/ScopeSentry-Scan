// Package dircore-----------------------------
// @file      : reuqest.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/4/28 23:51
// -------------------------------------------
package dircore

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"net/url"
)

type Request struct {
	Url string
}

func (r *Request) Request(path string) (types.HttpResponse, error) {
	uri := r.Url + url.QueryEscape(path)
	response, err := utils.Requests.HttpGet(uri)
	if err != nil {
		for i := 0; i < MaxRetries-5; i++ {
			response, err = utils.Requests.HttpGet(uri)
			if err != nil {
				logger.SlogWarnLocal(fmt.Sprintf("Senstrydir target %s request error: %s", uri, err))
				continue
			}
			return response, nil
		}
	}
	return response, err
}

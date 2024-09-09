// Package core -----------------------------
// @file      : types.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/5/7 15:30
// -------------------------------------------
package core

import (
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
)

type Options struct {
	Extensions         []string
	Thread             int
	IncludeStatusCodes []int
	MatchCallback      func(response types.HttpResponse)
}

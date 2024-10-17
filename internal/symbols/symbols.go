// symbols-------------------------------------
// @file      : symbols.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/17 19:03
// -------------------------------------------

//go:generate go install github.com/traefik/yaegi/cmd/yaegi@v0.15.0
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/pkg/logger
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/pkg/utils
package symbols

import "reflect"

var Symbols = map[string]map[string]reflect.Value{}

// symbols-------------------------------------
// @file      : symbols.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/17 19:03
// -------------------------------------------

//go:generate go install github.com/traefik/yaegi/cmd/yaegi@v0.15.0
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/pkg/logger
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/pkg/utils
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/options
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/bigcache
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/config
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/global
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/notification
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/plugins
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/pool
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/redis
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/results
//go:generate yaegi extract github.com/Autumn-27/ScopeSentry-Scan/internal/types
package symbols

import "reflect"

var Symbols = map[string]map[string]reflect.Value{}

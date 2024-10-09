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
)

type Result struct {
}

var Results *Result

func InitializeResults() {
	Results = &Result{}
}

func (r *Result) CompareAssetOther(old, new types.AssetOther) types.ChangeLogAssetOther {
	var Change types.ChangeLogAssetOther
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
	if !bytes.Equal(old.Raw, new.Raw) {
		Change.Change = append(Change.Change, types.ChangeLog{
			FieldName: "Raw",
			Old:       string(old.Raw),
			New:       string(new.Raw),
		})
	}
	if len(Change.Change) != 0 {
		Change.Timestamp = new.Timestamp
		return Change
	} else {
		return types.ChangeLogAssetOther{}
	}
}

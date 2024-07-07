package filekv

import "errors"

var (
	ErrItemExists   = errors.New("item already exist")
	ErrItemFiltered = errors.New("item filtered")
)

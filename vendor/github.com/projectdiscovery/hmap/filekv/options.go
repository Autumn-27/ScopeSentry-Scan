package filekv

var (
	BufferSize = 50 * 1024 * 1024 // 50Mb
	Separator  = ";;;"
	NewLine    = "\n"
	FpRatio    = 0.0001
	MaxItems   = uint(250000)
)

type Options struct {
	Path           string
	Compress       bool
	MaxItems       uint
	Cleanup        bool
	SkipEmpty      bool
	FilterCallback func(k, v []byte) bool
	Dedupe         Strategy
}

type Stats struct {
	NumberOfFilteredItems uint
	NumberOfAddedItems    uint
	NumberOfDupedItems    uint
	NumberOfItems         uint
}

var DefaultOptions Options = Options{
	Compress:  false,
	Cleanup:   true,
	Dedupe:    MemoryLRU,
	SkipEmpty: true,
}

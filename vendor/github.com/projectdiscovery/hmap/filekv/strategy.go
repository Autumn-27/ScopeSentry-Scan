package filekv

type Strategy uint8

const (
	None Strategy = iota
	// Uses go standard map without eviction - grows linearly with the items number
	MemoryMap
	// MemoryLRU keeps in memory the last x items and filter with a look-back probabilistic window with fixed size
	MemoryLRU
	// MemoryFilter uses bitset in-memory filters to remove duplicates
	MemoryFilter
	// Use full disk kv store to remove all duplicates - it should have low heap memory footprint but lots of I/O interactions
	DiskFilter
)

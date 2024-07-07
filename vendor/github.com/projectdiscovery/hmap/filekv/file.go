package filekv

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"io"
	"os"
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
	lru "github.com/hashicorp/golang-lru/v2"
	fileutil "github.com/projectdiscovery/utils/file"
	permissionutil "github.com/projectdiscovery/utils/permission"
	"github.com/syndtr/goleveldb/leveldb"
)

// FileDB - represents a file db implementation
type FileDB struct {
	stats       Stats
	options     Options
	tmpDbName   string
	tmpDb       *os.File
	tmpDbWriter io.WriteCloser
	db          *os.File
	dbWriter    io.WriteCloser

	// todo: refactor into independent package
	mapdb   map[string]struct{}
	mdb     *lru.Cache[string, struct{}] // lru cache
	bdb     *bloom.BloomFilter           // bloom filter
	ddb     *leveldb.DB                  // disk based filter
	ddbName string

	sync.RWMutex
}

// Open a new file based db
func Open(options Options) (*FileDB, error) {
	db, err := os.OpenFile(options.Path, os.O_RDWR|os.O_CREATE|os.O_APPEND, permissionutil.ConfigFilePermission)
	if err != nil {
		return nil, err
	}

	tmpFileName, err := fileutil.GetTempFileName()
	if err != nil {
		return nil, err
	}
	tmpDb, err := os.OpenFile(tmpFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, permissionutil.TempFilePermission)
	if err != nil {
		return nil, err
	}

	fdb := &FileDB{
		tmpDbName: tmpFileName,
		options:   options,
		db:        db,
		tmpDb:     tmpDb,
	}

	if options.Compress {
		fdb.tmpDbWriter = zlib.NewWriter(fdb.tmpDb)
		fdb.dbWriter = zlib.NewWriter(fdb.db)
	} else {
		fdb.tmpDbWriter = fdb.tmpDb
		fdb.dbWriter = fdb.db
	}

	return fdb, nil
}

// Process added files/slices/elements
func (fdb *FileDB) Process() error {
	// Closes the temporary file
	if fdb.options.Compress {
		// close the writer
		if err := fdb.tmpDbWriter.Close(); err != nil {
			return err
		}
	}

	// closes the file to flush to disk and reopen it
	_ = fdb.tmpDb.Close()
	var err error
	fdb.tmpDb, err = os.Open(fdb.tmpDbName)
	if err != nil {
		return err
	}

	var maxItems uint
	switch {
	case fdb.options.MaxItems > 0:
		maxItems = fdb.options.MaxItems
	case fdb.stats.NumberOfAddedItems < MaxItems:
		maxItems = fdb.stats.NumberOfAddedItems
	default:
		maxItems = MaxItems
	}

	// size the filter according to the number of input items
	switch fdb.options.Dedupe {
	case MemoryMap:
		fdb.mapdb = make(map[string]struct{}, maxItems)
	case MemoryLRU:
		fdb.mdb, err = lru.New[string, struct{}](int(maxItems))
		if err != nil {
			return err
		}
	case MemoryFilter:
		fdb.bdb = bloom.NewWithEstimates(maxItems, FpRatio)
	case DiskFilter:
		// using executable name so the same app using hmap will remove the files after a certain amount of time
		fdb.ddbName, err = os.MkdirTemp("", fileutil.ExecutableName())
		if err != nil {
			return err
		}
		fdb.ddb, err = leveldb.OpenFile(fdb.ddbName, nil)
		if err != nil {
			return err
		}
	}

	var tmpDbReader io.Reader
	if fdb.options.Compress {
		var err error
		tmpDbReader, err = zlib.NewReader(fdb.tmpDb)
		if err != nil {
			return err
		}
	} else {
		tmpDbReader = fdb.tmpDb
	}

	sc := bufio.NewScanner(tmpDbReader)
	buf := make([]byte, BufferSize)
	sc.Buffer(buf, BufferSize)
	for sc.Scan() {
		_ = fdb.Set(sc.Bytes(), nil)
	}

	fdb.tmpDb.Close()

	// flush to disk
	fdb.dbWriter.Close()
	fdb.db.Close()

	// cleanup filters
	switch fdb.options.Dedupe {
	case MemoryMap:
		fdb.mapdb = nil
	case MemoryLRU:
		fdb.mdb.Purge()
	case MemoryFilter:
		fdb.bdb.ClearAll()
	case DiskFilter:
		fdb.ddb.Close()
		os.RemoveAll(fdb.ddbName)
	}

	return nil
}

// Reset the db
func (fdb *FileDB) Reset() error {
	// clear the cache
	switch fdb.options.Dedupe {
	case MemoryMap:
		fdb.mapdb = nil
	case MemoryLRU:
		fdb.mdb.Purge()
	case MemoryFilter:
		fdb.bdb.ClearAll()
	case DiskFilter:
		// close - remove - reopen
		fdb.ddb.Close()
		os.RemoveAll(fdb.ddbName)
		var err error
		fdb.ddb, err = leveldb.OpenFile(fdb.ddbName, nil)
		if err != nil {
			return err
		}
	}

	// reset the tmp file
	fdb.tmpDb.Close()
	var err error
	fdb.tmpDb, err = os.Create(fdb.tmpDbName)
	if err != nil {
		return err
	}

	// reset the target file
	fdb.db.Close()
	fdb.db, err = os.Create(fdb.tmpDbName)
	if err != nil {
		return err
	}

	if fdb.options.Compress {
		fdb.tmpDbWriter = zlib.NewWriter(fdb.tmpDb)
		fdb.dbWriter = zlib.NewWriter(fdb.db)
	} else {
		fdb.tmpDbWriter = fdb.tmpDb
		fdb.dbWriter = fdb.db
	}

	return nil
}

// Size - returns the size of the database in bytes
func (fdb *FileDB) Size() int64 {
	osstat, err := fdb.db.Stat()
	if err != nil {
		return 0
	}
	return osstat.Size()
}

// Close ...
func (fdb *FileDB) Close() {
	tmpDBFilename := fdb.tmpDb.Name()
	_ = fdb.tmpDb.Close()
	os.RemoveAll(tmpDBFilename)

	_ = fdb.db.Close()
	dbFilename := fdb.db.Name()
	if fdb.options.Cleanup {
		os.RemoveAll(dbFilename)
	}

	if fdb.ddbName != "" {
		fdb.ddb.Close()
		os.RemoveAll(fdb.ddbName)
	}
}

func (fdb *FileDB) set(k, v []byte) error {
	var s bytes.Buffer
	s.Write(k)
	s.WriteString(Separator)
	s.Write(v)
	s.WriteString(NewLine)
	_, err := fdb.dbWriter.Write(s.Bytes())
	if err != nil {
		return err
	}
	fdb.stats.NumberOfItems++
	return nil
}

func (fdb *FileDB) Set(k, v []byte) error {
	// check for duplicates
	switch fdb.options.Dedupe {
	case MemoryMap:
		if _, ok := fdb.mapdb[string(k)]; ok {
			fdb.stats.NumberOfDupedItems++
			return ErrItemExists
		}
		fdb.mapdb[string(k)] = struct{}{}
	case MemoryLRU:
		if ok, _ := fdb.mdb.ContainsOrAdd(string(k), struct{}{}); ok {
			fdb.stats.NumberOfDupedItems++
			return ErrItemExists
		}
	case MemoryFilter:
		if ok := fdb.bdb.TestOrAdd(k); ok {
			fdb.stats.NumberOfDupedItems++
			return ErrItemExists
		}
	case DiskFilter:
		if ok, err := fdb.ddb.Has(k, nil); err == nil && ok {
			fdb.stats.NumberOfDupedItems++
			return ErrItemExists
		} else if err == nil && !ok {
			_ = fdb.ddb.Put(k, []byte{}, nil)
		}
	}

	if fdb.shouldSkip(k, v) {
		fdb.stats.NumberOfFilteredItems++
		return ErrItemFiltered
	}

	fdb.stats.NumberOfItems++
	return fdb.set(k, v)
}

// Scan - iterate over the whole store using the handler function
func (fdb *FileDB) Scan(handler func([]byte, []byte) error) error {
	// open the db and scan
	dbCopy, err := os.Open(fdb.options.Path)
	if err != nil {
		return err
	}
	defer dbCopy.Close()

	var dbReader io.ReadCloser
	if fdb.options.Compress {
		dbReader, err = zlib.NewReader(dbCopy)
		if err != nil {
			return err
		}
	} else {
		dbReader = dbCopy
	}

	sc := bufio.NewScanner(dbReader)
	buf := make([]byte, BufferSize)
	sc.Buffer(buf, BufferSize)
	for sc.Scan() {
		tokens := bytes.SplitN(sc.Bytes(), []byte(Separator), 2)
		var k, v []byte
		if len(tokens) > 0 {
			k = tokens[0]
		}
		if len(tokens) > 1 {
			v = tokens[1]
		}
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

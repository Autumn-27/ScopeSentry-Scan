// pebbledb-------------------------------------
// @file      : pebbledb.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 21:27
// -------------------------------------------

package pebbledb

import (
	"github.com/cockroachdb/pebble"
)

type PebbleDB struct {
	db *pebble.DB
}

var PebbleStore *PebbleDB

// NewPebbleDB 初始化 PebbleDB 实例
func NewPebbleDB(options *pebble.Options, dbPath string) (*PebbleDB, error) {
	db, err := pebble.Open(dbPath, options)
	if err != nil {
		return nil, err
	}
	return &PebbleDB{db: db}, nil
}

// Put 将键值对存入数据库
func (p *PebbleDB) Put(key, value []byte) error {
	return p.db.Set(key, value, pebble.Sync)
}

// Get 从数据库中获取指定键的值
func (p *PebbleDB) Get(key []byte) ([]byte, error) {
	value, closer, err := p.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	return value, nil
}

// Delete 删除数据库中的键
func (p *PebbleDB) Delete(key []byte) error {
	return p.db.Delete(key, pebble.Sync)
}

// Close 关闭数据库连接
func (p *PebbleDB) Close() error {
	return p.db.Close()
}

// BatchWrite 批量写入键值对
func (p *PebbleDB) BatchWrite(pairs map[string]string) error {
	batch := p.db.NewBatch()
	defer batch.Close()

	for key, value := range pairs {
		if err := batch.Set([]byte(key), []byte(value), nil); err != nil {
			return err
		}
	}

	if err := batch.Commit(pebble.Sync); err != nil {
		return err
	}

	return nil
}

// Compact 强制压缩数据库，释放空间
func (p *PebbleDB) Compact() error {
	start := []byte("")               // 从最开始位置压缩
	end := []byte("zzzzzzzzzzzzzzzz") // 直到最大键的后面
	// 调用 Compact 方法进行压缩
	return p.db.Compact(start, end, true)
}

func (p *PebbleDB) GetKeysWithPrefix(prefix string) (map[string][]byte, error) {
	result := make(map[string][]byte)

	// 创建迭代器并设置范围
	lowerBound := []byte(prefix)          // 设置下界为指定前缀
	upperBound := []byte(prefix + "\xff") // 设置上界为指定前缀后加上一个字符

	iter, _ := p.db.NewIter(&pebble.IterOptions{
		LowerBound: lowerBound,
		UpperBound: upperBound,
	})
	defer func(iter *pebble.Iterator) {
		err := iter.Close()
		if err != nil {

		}
	}(iter) // 确保在函数结束时关闭迭代器

	// 遍历所有符合条件的键
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()
		result[string(key)] = value
	}

	if err := iter.Error(); err != nil {
		return nil, err
	}

	return result, nil
}

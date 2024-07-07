package filekv

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

func (f *FileDB) Merge(items ...interface{}) (uint, error) {
	var count uint
	for _, item := range items {
		switch itemData := item.(type) {
		case [][]byte:
			for _, data := range itemData {
				if _, err := f.tmpDbWriter.Write(data); err != nil {
					return 0, err
				}
				if _, err := f.tmpDbWriter.Write([]byte(NewLine)); err != nil {
					return 0, err
				}
				count++
				f.stats.NumberOfAddedItems++
			}
		case []string:
			for _, data := range itemData {
				_, err := f.tmpDbWriter.Write([]byte(data + NewLine))
				if err != nil {
					return 0, err
				}
				count++
				f.stats.NumberOfAddedItems++
			}
		case io.Reader:
			c, err := f.MergeReader(itemData)
			if err != nil {
				return 0, err
			}
			count += c
		case string:
			c, err := f.MergeFile(itemData)
			if err != nil {
				return 0, err
			}
			count += c
		}
	}
	return count, nil
}

func (f *FileDB) shouldSkip(k, v []byte) bool {
	if f.options.SkipEmpty && len(k) == 0 {
		return true
	}

	if f.options.FilterCallback != nil {
		return f.options.FilterCallback(k, v)
	}

	return false
}

func (f *FileDB) MergeFile(filename string) (uint, error) {
	newF, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer newF.Close()

	return f.MergeReader(newF)
}

func (f *FileDB) MergeReader(reader io.Reader) (uint, error) {
	var count uint
	sc := bufio.NewScanner(reader)
	buf := make([]byte, BufferSize)
	sc.Buffer(buf, BufferSize)
	var itemToWrite bytes.Buffer
	for sc.Scan() {
		itemToWrite.Write(sc.Bytes())
		itemToWrite.WriteString(NewLine)
		if _, err := f.tmpDbWriter.Write(itemToWrite.Bytes()); err != nil {
			return 0, err
		}
		itemToWrite.Reset()
		count++
		f.stats.NumberOfAddedItems++
	}
	return count, nil
}

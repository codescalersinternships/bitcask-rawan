package bitcask

import (
	"os"
	"path"
)

func Get(b *BitcaskHandle, key []byte) ([]byte, error) {

	var fEntry *KeydirEntry
	fEntry = nil
	for k, entry := range b.Keydir {
		if k == string(key) {
			fEntry = &entry
		}
	}

	if fEntry == nil {
		return nil, ErrKeyNotFound
	}

	file, err := os.Open(path.Join(b.Dir, fEntry.FileId))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	valueFromFile := make([]byte, fEntry.ValueSize)
	_, err = file.ReadAt(valueFromFile, fEntry.ValuePos)
	if err != nil {
		return nil, err
	}

	return valueFromFile, nil
}

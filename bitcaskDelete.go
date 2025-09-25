package bitcask

import (
	"encoding/binary"
	"path/filepath"
	"time"
)

func Delete(b *BitcaskHandle, key []byte) error {

	if !b.Opts.ReadWrite {
		return ErrUnuthorizedPut
	}

	file := b.active

	// write tombstone record, value size = 0
	header := make([]byte, 20)
	ts := uint64(time.Now().Unix())
	binary.BigEndian.PutUint32(header[0:4], calculateCRC(ts, len(key), 0, key, nil))
	binary.BigEndian.PutUint64(header[4:12], ts)
	binary.BigEndian.PutUint32(header[12:16], uint32(len(key)))
	binary.BigEndian.PutUint32(header[16:20], uint32(0))

	if _, err := file.Write(header); err != nil {
		return err
	}

	if _, err := file.Write(key); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.Keydir[string(key)] = KeydirEntry{
		Timestamp: ts,
		FileId:    filepath.Base(file.Name()),
		ValueSize: 0,
		ValuePos: 0,
		Tombstone: true,
	}

	return nil
}

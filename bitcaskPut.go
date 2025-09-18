package bitcask

import (
	"encoding/binary"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"time"
)

func calculateCRC(ts uint64, keySize, valueSize int, key string, value []byte) uint32 {
	header := make([]byte, 16)
	binary.BigEndian.PutUint64(header[0:8], ts)
	binary.BigEndian.PutUint32(header[8:12], uint32(keySize))
	binary.BigEndian.PutUint32(header[12:16], uint32(valueSize))

	data := append(header, []byte(key)...)
	data = append(data, value...)
	return crc32.ChecksumIEEE(data)
}

func Put(handle *BitcaskHandle, key []byte, value []byte) error {
	if !handle.Opts.ReadWrite {
		return ErrUnuthorizedPut
	}

	file, err := os.OpenFile(handle.active.Name(), os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	pos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	header := make([]byte, 20)
	ts := uint64(time.Now().Unix())
	binary.BigEndian.PutUint32(header[0:4], calculateCRC(ts, len(key), len(value), string(key), value))
	binary.BigEndian.PutUint64(header[4:12], ts)
	binary.BigEndian.PutUint32(header[12:16], uint32(len(key)))
	binary.BigEndian.PutUint32(header[16:20], uint32(len(value)))

	file.Write(header)
	file.Write(key)
	file.Write(value)

	handle.mu.Lock()
	defer handle.mu.Unlock()
	handle.Keydir[string(key)] = KeydirEntry{
		Timestamp: ts,
		FileId:    filepath.Base(handle.active.Name()),
		ValueSize: len(value),
		ValuePos:  pos + 20 + int64(len(key)),
	}

	return nil
}

package bitcask

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func getNextFileId(current string) string {
	base := strings.TrimSuffix(current, ".data")
	n, _ := strconv.Atoi(base)
	return fmt.Sprintf("%010d.data", n+1)
}

func Put(handle *BitcaskHandle, key []byte, value []byte) error {
	if !handle.Opts.ReadWrite {
		return ErrUnuthorizedPut
	}

	header := make([]byte, 20)
	ts := uint64(time.Now().Unix())
	binary.BigEndian.PutUint32(header[0:4], calculateCRC(ts, len(key), len(value), string(key), value))
	binary.BigEndian.PutUint64(header[4:12], ts)
	binary.BigEndian.PutUint32(header[12:16], uint32(len(key)))
	binary.BigEndian.PutUint32(header[16:20], uint32(len(value)))

	var pos int64

	info, err := handle.active.Stat()
	if err != nil {
		return err
	}

	if info.Size() >= MaxFileSize {
		if err := handle.active.Close(); err != nil {
			return err
		}

		newFileId := getNextFileId(filepath.Base(handle.active.Name()))
		newFilePath := filepath.Join(handle.Dir, newFileId)

		newFile, err := os.OpenFile(newFilePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}

		handle.active = newFile
	}

	if _, err := handle.active.Write(header); err != nil {
		return err
	}
	if _, err := handle.active.Write(key); err != nil {
		return err
	}

	pos, err = handle.active.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	if _, err := handle.active.Write(value); err != nil {
		return err
	}

	handle.mu.Lock()
	defer handle.mu.Unlock()
	handle.Keydir[string(key)] = KeydirEntry{
		Timestamp: ts,
		FileId:    filepath.Base(handle.active.Name()),
		ValueSize: len(value),
		ValuePos:  pos,
	}

	return nil
}

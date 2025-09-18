package bitcask

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const MaxFileSize = 100 * 1024 * 1024 // 100MB

var ErrIncorrectCrc = errors.New("incorrect crc")

type FileEntry struct {
	Crc       uint32
	Timestamp uint64
	KeySize   int
	ValueSize int
	Key       string
	Value     any
}

type KeydirEntry struct {
	Timestamp uint64
	FileId    string
	ValueSize int
	ValuePos  int64
}

type Options struct {
	ReadWrite bool
	SyncOnPut bool
}

type BitcaskHandle struct {
	Dir    string
	Keydir map[string]KeydirEntry
	Opts   Options
	mu     sync.RWMutex
	active *os.File
}

// -----header-----
// crc	ts	ks	vs		key		value
// 4	8	4	4		ksz		vsz
func readRecord(f *os.File) (*FileEntry, int64, error) {
	header := make([]byte, 20)
	_, err := io.ReadFull(f, header)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, 0, io.EOF
		}
		return nil, 0, err
	}

	crc := binary.BigEndian.Uint32(header[0:4])
	ts := binary.BigEndian.Uint64(header[4:12])
	kSize := int(binary.BigEndian.Uint32(header[12:16]))
	vSize := int(binary.BigEndian.Uint32(header[16:20]))

	key := make([]byte, kSize)
	if _, err := io.ReadFull(f, key); err != nil {
		return nil, 0, err
	}
	val := make([]byte, vSize)
	if _, err := io.ReadFull(f, val); err != nil {
		return nil, 0, err
	}

	crcBytes := append(header[4:], key...)
	crcBytes = append(crcBytes, val...)

	if crc32.ChecksumIEEE(crcBytes) != crc {
		return nil, 0, ErrIncorrectCrc
	}

	entry := &FileEntry{
		Crc:       crc,
		Timestamp: ts,
		KeySize:   kSize,
		ValueSize: vSize,
		Key:       string(key),
		Value:     val,
	}

	totalSize := int64(len(header) + kSize + vSize)
	return entry, totalSize, nil
}

func Open(dir string, opts Options) (*BitcaskHandle, error) {

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	b := &BitcaskHandle{
		Dir:    dir,
		Keydir: make(map[string]KeydirEntry),
		Opts:   opts,
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.data"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		var offset int64
		for {
			entry, size, err := readRecord(f)
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF{
					break
				}
				f.Close()
				return nil, err
			}
			b.Keydir[entry.Key] = KeydirEntry{
				Timestamp: entry.Timestamp,
				FileId:    filepath.Base(file),
				ValueSize: entry.ValueSize,
				ValuePos:  offset + 20 + int64(entry.KeySize), // to stop right before the value
			}

			offset += size
		}
		f.Close()
	}
	return b, nil
}

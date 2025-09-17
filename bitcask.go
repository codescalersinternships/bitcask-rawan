package bitcask

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const MaxFileSize = 100 * 1024 * 1024 // 100MB

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

func readRecords(f *os.File) ([]FileEntry, error) {
	var entries []FileEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 6)

		crc, _ := strconv.Atoi(parts[0])
		ts, _ := strconv.ParseInt(parts[1], 10, 64)
		ksz, _ := strconv.Atoi(parts[2])
		vsz, _ := strconv.Atoi(parts[3])
		key := parts[4]
		value := parts[5]
		entries = append(entries, FileEntry{
			Crc:       uint32(crc),
			Timestamp: uint64(ts),
			KeySize:   ksz,
			ValueSize: vsz,
			Key:       key,
			Value:     value,
		})
	}
	return entries, scanner.Err()
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
		defer f.Close()

		for {
			entries, err := readRecords(f)
			if err != nil {
				return nil, err
			}
			for _, entry := range entries {
				b.Keydir[entry.Key] = KeydirEntry{
					Timestamp: entry.Timestamp,
					FileId:    filepath.Base(file),
					ValueSize: entry.ValueSize,
				}
			}
		}
	}
	return b, nil
}

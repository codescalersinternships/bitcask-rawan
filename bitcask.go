package bitcask

import (
	"os"
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


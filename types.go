package bitcask

import (
	"errors"
	"os"
	"sync"
)

const MaxFileSize = 100 // 100 byte

var ErrIncorrectCrc = errors.New("incorrect crc")
var ErrUnuthorizedPut = errors.New("this process in unauthorized to write in this bitcask store")
var ErrKeyNotFound = errors.New("key not found")
type FileEntry struct {
	Crc       uint32
	Timestamp uint64
	KeySize   int
	ValueSize int
	Key       string
	Value     []byte
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

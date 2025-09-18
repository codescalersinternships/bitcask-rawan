package bitcask

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func createTestDataFile(filename string, entries []*FileEntry) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, entry := range entries {
		header := make([]byte, 20)
		binary.BigEndian.PutUint32(header[0:4], entry.Crc)
		binary.BigEndian.PutUint64(header[4:12], entry.Timestamp)
		binary.BigEndian.PutUint32(header[12:16], uint32(entry.KeySize))
		binary.BigEndian.PutUint32(header[16:20], uint32(entry.ValueSize))

		file.Write(header)
		file.Write([]byte(entry.Key))
		file.Write(entry.Value.([]byte))
	}

	return nil
}

func TestOpen(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testdata")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	entries := []*FileEntry{
		{
			Timestamp: 1234567890,
			Key:       "key1",
			Value:     []byte("value1"),
		},
		{
			Timestamp: 24681012,
			Key:       "key2",
			Value:     []byte("value2"),
		},
	}
	for _, entry := range entries {
		entry.KeySize = len(entry.Key)
		entry.ValueSize = len(entry.Value.([]byte))
		entry.Crc = calculateCRC(entry.Timestamp, entry.KeySize, entry.ValueSize, entry.Key, entry.Value.([]byte))
	}

	path := filepath.Join(tempDir, "0000000001.data")
	err = createTestDataFile(path, entries)
	if err != nil {
		t.Fatalf("Failed to create test file %v", err)
	}

	b, err := Open(tempDir, Options{})
	if err != nil {
		t.Fatalf("failed to open bitcask %v", err)
	}
	expectedKeyDir := map[string]KeydirEntry{
		"key1": {
			Timestamp: 1234567890,
			FileId:    "0000000001.data",
			ValueSize: 6,
			ValuePos:  20 + 4,
		},

		"key2": {
			Timestamp: 24681012,
			FileId:    "0000000001.data",
			ValueSize: 6,
			ValuePos:  (20 + 4 + 6) + 20 + 4, //first entry + second entry until value
		},
	}

	if len(expectedKeyDir) != len(b.Keydir) {
		t.Errorf("expected %d keys, got %d", len(expectedKeyDir), len(b.Keydir))
	}

	if !reflect.DeepEqual(expectedKeyDir, b.Keydir) {
		t.Errorf("expected keydir %v, got %v", expectedKeyDir, b.Keydir)
	}
}

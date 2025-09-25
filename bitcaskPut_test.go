package bitcask

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPutNoData(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testdata")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	b, err := Open(tempDir, Options{ReadWrite: true})
	if err != nil {
		t.Fatalf("failed to open bitcask %v", err)
	}

	key := []byte("putTestKey")
	value := []byte("1")

	err = Put(b, key, value)

	if err != nil {
		t.Fatalf("error putting the key %v", err)
	}

	keyStr := string(key)
	// key exist in keydir
	entry, exists := b.Keydir[keyStr]

	if !exists {
		t.Fatalf("key %s not found in keydir", keyStr)
	}

	// value size
	if entry.ValueSize != len(value) {
		t.Errorf("expected value size %d, got %d", len(value), entry.ValueSize)
	}

	// value written at the correct position
	file, err := os.Open(filepath.Join(tempDir, entry.FileId))
	if err != nil {
		t.Fatalf("failed to open data file: %v", err)
	}
	defer file.Close()

	valueFromFile := make([]byte, entry.ValueSize)
	_, err = file.ReadAt(valueFromFile, entry.ValuePos)
	if err != nil {
		t.Fatalf("failed to read value from file: %v", err)
	}

	if string(valueFromFile) != string(value) {
		t.Errorf("value from file mismatch: expected %s, got %s", string(value), string(valueFromFile))
	}
}

func createTestDataFiles(t *testing.T, tempDir string) {
	t.Helper()

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
		entry.ValueSize = len(entry.Value)
		entry.Crc = calculateCRC(entry.Timestamp, entry.KeySize, entry.ValueSize, []byte(entry.Key), entry.Value)
	}

	path := filepath.Join(tempDir, "0000000001.data")
	err := createTestDataFile(path, entries)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}

func TestPutWithData(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testdata")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	createTestDataFiles(t, tempDir)

	b, err := Open(tempDir, Options{ReadWrite: true})
	if err != nil {
		t.Fatalf("failed to open bitcask %v", err)
	}

	key := []byte("putTestKey")
	value := []byte("1")

	err = Put(b, key, value)
	if err != nil {
		t.Fatalf("error putting the key %v", err)
	}

	keyStr := string(key)
	// key exist in keydir
	entry, exists := b.Keydir[keyStr]
	if !exists {
		t.Fatalf("key %s not found in keydir", keyStr)
	}

	// value size
	if entry.ValueSize != len(value) {
		t.Errorf("expected value size %d, got %d", len(value), entry.ValueSize)
	}

	// value written at the correct position
	file, err := os.Open(filepath.Join(tempDir, entry.FileId))
	if err != nil {
		t.Fatalf("failed to open data file: %v", err)
	}
	defer file.Close()

	valueFromFile := make([]byte, entry.ValueSize)
	_, err = file.ReadAt(valueFromFile, entry.ValuePos)
	if err != nil {
		t.Fatalf("failed to read value from file: %v", err)
	}

	if string(valueFromFile) != string(value) {
		t.Errorf("value from file mismatch: expected %s, got %s", string(value), string(valueFromFile))
	}
}

func TestPutReadOnly(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testdata")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	b, err := Open(tempDir, Options{ReadWrite: false})
	if err != nil {
		t.Fatalf("failed to open bitcask: %v", err)
	}

	err = Put(b, []byte("testKey"), []byte("testValue"))
	if err != ErrUnuthorizedPut {
		t.Errorf("expected ErrUnuthorizedPut, got %v", err)
	}
}

func TestPutFileThreshold(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testdata")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	b, err := Open(tempDir, Options{ReadWrite: true})
	if err != nil {
		t.Fatalf("failed to open bitcask %v", err)
	}

	key := []byte("k1")
	value := []byte(strings.Repeat("a", 90))

	err = Put(b, key, value)
	if err != nil {
		t.Fatalf("error putting the key %v", err)
	}

	key = []byte("k2")
	value = []byte(strings.Repeat("a", 90))

	err = Put(b, key, value)
	if err != nil {
		t.Fatalf("error putting the key %v", err)
	}

	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}

	dataFiles := []string{}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".data") {
			dataFiles = append(dataFiles, f.Name())
		}
	}

	if len(dataFiles) < 2 {
		t.Errorf("expected more than 2 files due to file threshold, got %v", dataFiles)
	}

	keyStr := string(key)
	// key exist in keydir
	entry, exists := b.Keydir[keyStr]
	if !exists {
		t.Fatalf("key %s not found in keydir", keyStr)
	}

	//correct fileid check
	if entry.FileId != dataFiles[len(dataFiles)-1] {
		t.Errorf("wrong active file")
	}

	// value written at the correct position
	file, err := os.Open(filepath.Join(tempDir, entry.FileId))
	if err != nil {
		t.Fatalf("failed to open data file: %v", err)
	}
	defer file.Close()

	valueFromFile := make([]byte, entry.ValueSize)
	_, err = file.ReadAt(valueFromFile, entry.ValuePos)
	if err != nil {
		t.Fatalf("failed to read value from file: %v", err)
	}

	if string(valueFromFile) != string(value) {
		t.Errorf("value from file mismatch: expected %s, got %s", string(value), string(valueFromFile))
	}
}

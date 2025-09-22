package bitcask

import (
	"os"
	"testing"
)

func TestBitcaskGet(t *testing.T) {
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

	value, err := Get(b, []byte("key1"))

	if err != nil {
		t.Fatalf("failed to get key %v", err)
	}
	expectedValue := []byte("value1")

	if string(value) != string(expectedValue) {
		t.Errorf("value mismatch, expected %s, got %s", string(expectedValue), string(value))
	}

}



func TestBitcaskGetNotFound(t *testing.T) {
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

	value, err := Get(b, []byte("NotFoundKey"))

	if err != ErrKeyNotFound{
		t.Errorf("error mismatch, should have got key not found error")
	}

	if value!=nil{
		t.Errorf("value mismatch, value should be nil")
	}

}
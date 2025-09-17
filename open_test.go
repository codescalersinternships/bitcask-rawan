package bitcask

import (
	"os"
	"testing"
)

func TestReadRecord(t *testing.T) {
	f, err := os.Open("testdata/test.data")
	if err != nil {
		t.Fatalf("Failed to open test data file: %v", err)
	}
	defer f.Close()

	entries, err := readRecords(f)
	if err != nil {
		t.Fatalf("Failed to read records: %v", err)
	}

	if len(entries) != 10 {
		t.Fatalf("Expected 10 entries, got %d", len(entries))
	}

	expectedFileEntry := FileEntry{
		Crc:       1234,
		Timestamp: 56789,
		Key:       "foo",
		Value:     "bar",
		KeySize:   3,
		ValueSize: 3,
	}

	for _, entry := range entries {
		if entry != expectedFileEntry {
			t.Errorf("Expected %+v, got %+v", expectedFileEntry, entry)
		}
	}
}

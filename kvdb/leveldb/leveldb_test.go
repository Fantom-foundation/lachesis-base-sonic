package leveldb

import (
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

func TestBatchDeleteRange_KeysInRangeGetDeleted(t *testing.T) {
	db := newTestDB(t)

	// Insert test keys: "a", "b", "c", "d", "e"
	keys := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")}
	for _, key := range keys {
		if err := db.Put(key, []byte("val-"+string(key)), nil); err != nil {
			t.Fatalf("failed to put key %s: %v", key, err)
		}
	}

	batch := &batch{
		db: db,
		b:  new(leveldb.Batch),
	}

	// Delete keys in range ["b", "d")
	err := batch.DeleteRange([]byte("b"), []byte("d"))
	if err != nil {
		t.Fatalf("DeleteRange failed: %v", err)
	}

	// Write the batch
	if err := batch.Write(); err != nil {
		t.Fatalf("batch write failed: %v", err)
	}

	// Check which keys remain
	tests := []struct {
		key      []byte
		expected bool
	}{
		{[]byte("a"), true},
		{[]byte("b"), false},
		{[]byte("c"), false},
		{[]byte("d"), true},
		{[]byte("e"), true},
	}
	for _, test := range tests {
		ok, err := db.Has(test.key, nil)
		if err != nil {
			t.Fatalf("db.Has(%s) failed: %v", test.key, err)
		}
		if ok != test.expected {
			t.Errorf("key %s: expected %v, got %v", test.key, test.expected, ok)
		}
	}
}

func TestBatchDeleteRange_NoKeysInRange(t *testing.T) {
	db := newTestDB(t)

	// Insert keys "a", "b"
	keys := [][]byte{[]byte("a"), []byte("b")}
	for _, key := range keys {
		if err := db.Put(key, []byte("val"), nil); err != nil {
			t.Fatalf("failed to put key: %v", err)
		}
	}

	batch := &batch{
		db: db,
		b:  new(leveldb.Batch),
	}

	// Delete range ["c", "d") -- no keys in this range
	err := batch.DeleteRange([]byte("c"), []byte("d"))
	if err != nil {
		t.Fatalf("DeleteRange failed: %v", err)
	}
	if err := batch.Write(); err != nil {
		t.Fatalf("batch write failed: %v", err)
	}

	for _, key := range keys {
		ok, err := db.Has(key, nil)
		if err != nil {
			t.Fatalf("db.Has failed: %v", err)
		}
		if !ok {
			t.Errorf("key %s should not be deleted", key)
		}
	}
}

func newTestDB(t *testing.T) *leveldb.DB {
	dir := t.TempDir()
	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close test db: %v", err)
		}
	})
	return db
}

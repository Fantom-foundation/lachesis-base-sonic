package pebble

import (
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/stretchr/testify/require"
)

func TestBatchDeleteRange_KeysInRangeGetDeleted(t *testing.T) {
	db := newTestDB(t)

	// Insert some keys
	keys := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
		[]byte("d"),
		[]byte("e"),
	}
	for _, k := range keys {
		db.Set(k, []byte("value-"+string(k)), nil)
	}

	batch := batch{
		db:   db,
		b:    db.NewBatch(),
		size: 0,
	}
	// Delete keys in range ["b", "d")
	err := batch.DeleteRange([]byte("b"), []byte("d"))
	require.NoError(t, err)

	// Write the batch
	err = batch.Write()
	require.NoError(t, err)

	// Check which keys remain
	testKeys := [][]byte{
		[]byte("a"),
		[]byte("d"),
		[]byte("e"),
	}

	got := [][]byte{}
	for _, key := range testKeys {
		_, closer, err := db.Get(key)
		if err == nil {
			got = append(got, key)
		}
		closer.Close()
	}
	require.ElementsMatch(t, got, testKeys, "Keys not deleted in range")
}

func TestBatchDeleteRange_NoKeysInRange(t *testing.T) {
	db := newTestDB(t)

	// Insert some keys
	keys := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}
	for _, k := range keys {
		db.Set(k, []byte("value-"+string(k)), nil)
	}

	batch := batch{
		db:   db,
		b:    db.NewBatch(),
		size: 0,
	}
	// Delete keys in range ["d", "e")
	err := batch.DeleteRange([]byte("d"), []byte("e"))
	require.NoError(t, err)

	// Write the batch
	err = batch.Write()
	require.NoError(t, err)

	// Check which keys remain
	testKeys := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}

	got := [][]byte{}
	for _, key := range testKeys {
		_, closer, err := db.Get(key)
		if err == nil {
			got = append(got, key)
		}
		closer.Close()
	}
	require.ElementsMatch(t, got, testKeys, "Keys not deleted in range")
}

func newTestDB(t *testing.T) *pebble.DB {
	dir := t.TempDir()
	db, err := pebble.Open(dir+"testDB", &pebble.Options{})
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

package database

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgraph-io/badger/v3"
)

type BadgerDB struct {
	db *badger.DB
}

// InitDB initializes Badger database
func InitDB() (*BadgerDB, error) {
	dbPath := "./badger_data"

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable logging
	opts.SyncWrites = false
	// Removed unsupported option in current Badger version:
	// opts.MaxTableSize = 64 << 20 // 64MB

	// NEW: allow overriding SyncWrites via environment variable
	if v := os.Getenv("BADGER_SYNC_WRITES"); v != "" {
		switch {
		case v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes"):
			opts.SyncWrites = true
		case v == "0" || strings.EqualFold(v, "false") || strings.EqualFold(v, "no"):
			opts.SyncWrites = false
		}
	}

	// Keep value log file size tuning (this option exists in v3)
	opts.ValueLogFileSize = 256 << 20 // 256MB

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	return &BadgerDB{db: db}, nil
}

// Close closes the database
func (b *BadgerDB) Close() error {
	return b.db.Close()
}

// Set stores a key-value pair
func (b *BadgerDB) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// Get retrieves a value by key
func (b *BadgerDB) Get(key string) ([]byte, error) {
	var result []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			result = make([]byte, len(val))
			copy(result, val)
			return nil
		})
	})
	return result, err
}

// GetJSON retrieves and unmarshals a value by key
func (b *BadgerDB) GetJSON(key string, v interface{}) error {
	data, err := b.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Delete removes a key from the database
func (b *BadgerDB) Delete(key string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// Exists checks if a key exists
func (b *BadgerDB) Exists(key string) (bool, error) {
	exists := false
	err := b.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		exists = true
		return nil
	})
	return exists, err
}

// GetAll retrieves all values with a prefix
func (b *BadgerDB) GetAll(prefix string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())
			var val []byte
			err := item.Value(func(v []byte) error {
				val = make([]byte, len(v))
				copy(val, v)
				return nil
			})
			if err != nil {
				return err
			}
			result[key] = val
		}
		return nil
	})
	return result, err
}

// GenerateHash creates a SHA256 hash from data
func GenerateHash(data ...string) string {
	h := sha256.New()
	for _, d := range data {
		h.Write([]byte(d))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// GC runs garbage collection
func (b *BadgerDB) GC() error {
	return b.db.RunValueLogGC(0.5)
}

// Backup creates a backup
func (b *BadgerDB) Backup(backupPath string) error {
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return err
	}

	backupFile := filepath.Join(backupPath, "backup.db")
	f, err := os.Create(backupFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = b.db.Backup(f, 0)
	return err
}

// IterateWithPrefix iterates over all items with a prefix and applies a function
func (b *BadgerDB) IterateWithPrefix(prefix string, fn func(key string, value []byte) error) error {
	return b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())
			var val []byte
			err := item.Value(func(v []byte) error {
				val = make([]byte, len(v))
				copy(val, v)
				return nil
			})
			if err != nil {
				return err
			}
			if err := fn(key, val); err != nil {
				return err
			}
		}
		return nil
	})
}

// CountByPrefix returns the number of keys starting with the given prefix.
// This is used by stats summary to count contacts/issues/prs/repos.
func (b *BadgerDB) CountByPrefix(prefix string) (int, error) {
	count := 0
	err := b.db.View(func(txn *badger.Txn) error {
		itrOpts := badger.DefaultIteratorOptions
		itrOpts.PrefetchValues = false
		it := txn.NewIterator(itrOpts)
		defer it.Close()

		p := []byte(prefix)
		for it.Seek(p); it.ValidForPrefix(p); it.Next() {
			count++
		}
		return nil
	})
	return count, err
}

// INSERT: iterate over all keys with a given prefix
func (b *BadgerDB) IteratePrefix(prefix string, fn func(k []byte, v []byte) error) error {
	return b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		p := []byte(prefix)
		for it.Seek(p); it.ValidForPrefix(p); it.Next() {
			item := it.Item()
			k := item.KeyCopy(nil)
			if err := item.Value(func(v []byte) error {
				// pass a copy of value to callback
				val := make([]byte, len(v))
				copy(val, v)
				return fn(k, val)
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

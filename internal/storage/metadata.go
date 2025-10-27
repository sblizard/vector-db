package storage

import (
	"github.com/dgraph-io/badger/v4"
)

type MetaStore struct {
	DB *badger.DB
}

func NewMetaStore(path string) *MetaStore {
	db, err := OpenMetaStore(path)
	if err != nil {
		panic(err)
	}
	return db
}

// OpenMetaStore opens a BadgerDB instance at the specified path.
func OpenMetaStore(path string) (*MetaStore, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &MetaStore{DB: db}, nil
}

// Close closes the BadgerDB instance.
func (m *MetaStore) Close() error {
	return m.DB.Close()
}

// Put stores a key-value pair in the MetaStore.
func (m *MetaStore) Put(key string, value []byte) error {
	return m.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

// Get retrieves the value for a given key from the MetaStore.
func (m *MetaStore) Get(key string) ([]byte, error) {
	var valCopy []byte
	err := m.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		valCopy, err = item.ValueCopy(nil)
		return err
	})
	return valCopy, err
}

// Delete removes a key-value pair from the MetaStore.
func (m *MetaStore) Delete(key string) error {
	return m.DB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// GetAll retrieves all key-value pairs from the MetaStore.
func (m *MetaStore) GetAll() (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := m.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := string(item.Key())
			valCopy, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			result[k] = valCopy
		}
		return nil
	})
	return result, err
}

// DeleteAllMetadata removes all metadata entries from the MetaStore.
func (m *MetaStore) DeleteAllMetadata() error {
	return m.DB.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(indexPrefix)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			if err := txn.Delete(item.Key()); err != nil {
				return err
			}
		}
		return nil
	})
}

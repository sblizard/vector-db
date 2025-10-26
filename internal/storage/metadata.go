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

func OpenMetaStore(path string) (*MetaStore, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &MetaStore{DB: db}, nil
}

func (m *MetaStore) Close() error {
	return m.DB.Close()
}

func (m *MetaStore) Put(key string, value []byte) error {
	return m.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

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

func (m *MetaStore) Delete(key string) error {
	return m.DB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

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

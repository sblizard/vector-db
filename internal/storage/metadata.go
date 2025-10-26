package storage

import (
	"github.com/dgraph-io/badger/v4"
)

type MetaStore struct {
	DB *badger.DB
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

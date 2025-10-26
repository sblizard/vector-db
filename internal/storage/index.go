package storage

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

type VectorIndex struct {
	Position int64 `json:"position"`
	Dim      int   `json:"dim"`
}

const indexPrefix = "_idx_"

func (m *MetaStore) PutIndex(id string, index VectorIndex) error {
	data, err := json.Marshal(index)
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	return m.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(indexPrefix+id), data)
	})
}

func (m *MetaStore) GetIndex(id string) (*VectorIndex, error) {
	var index VectorIndex
	err := m.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(indexPrefix + id))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &index)
		})
	})

	if err != nil {
		return nil, err
	}
	return &index, nil
}

func (m *MetaStore) DeleteIndex(id string) error {
	return m.DB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(indexPrefix + id))
	})
}

func (m *MetaStore) GetAllIndices() (map[string]VectorIndex, error) {
	result := make(map[string]VectorIndex)
	err := m.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(indexPrefix)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := string(item.Key()[len(indexPrefix):])

			err := item.Value(func(val []byte) error {
				var index VectorIndex
				if err := json.Unmarshal(val, &index); err != nil {
					return err
				}
				result[k] = index
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return result, err
}

func (m *MetaStore) GetNextPosition(dim int) (int64, error) {
	indices, err := m.GetAllIndices()
	if err != nil {
		return 0, err
	}

	if len(indices) == 0 {
		return 0, nil
	}

	var maxPos int64 = -1
	var maxDim int
	for _, idx := range indices {
		if idx.Position > maxPos {
			maxPos = idx.Position
			maxDim = idx.Dim
		}
	}

	bytesPerFloat := int64(4)
	return maxPos + int64(maxDim)*bytesPerFloat, nil
}

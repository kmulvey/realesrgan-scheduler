package cache

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/kmulvey/path"
)

type Cache struct {
	*badger.DB
}

func NewCache(cachePath string) (Cache, error) {
	db, err := badger.Open(badger.DefaultOptions(cachePath))
	if err != nil {
		return Cache{}, fmt.Errorf("error opening badger db: %w", err)
	}

	return Cache{db}, nil
}

func (c *Cache) Close() error {
	return c.DB.Close()
}

func (c *Cache) AddImage(image path.Entry) error {
	return c.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(image.AbsolutePath), nil)
	})
}

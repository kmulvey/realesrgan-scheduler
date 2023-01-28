package cache

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/kmulvey/path"
)

type Cache struct {
	*badger.DB
}

func New(cachePath string) (Cache, error) {
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

func (c *Cache) Contains(image path.Entry) bool {

	var found bool
	if err := c.DB.View(func(txn *badger.Txn) error {

		var _, err = txn.Get([]byte(image.AbsolutePath))
		if err == nil {
			found = true
			return nil
		}
		return err

	}); err != nil {
		return false
	}

	return found
}

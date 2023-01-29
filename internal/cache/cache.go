package cache

import (
	"fmt"
	"strings"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/kmulvey/path"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type Cache struct {
	*badger.DB
}

func New(cachePath string) (Cache, error) {
	var l = logrus.New()
	l.SetLevel(log.ErrorLevel)

	var opts = badger.DefaultOptions(cachePath)
	opts.Logger = l

	db, err := badger.Open(opts)
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

func (c *Cache) ListKeys(searchTerm string, images chan string) error {

	return c.DB.View(func(txn *badger.Txn) error {

		var opts = badger.DefaultIteratorOptions
		opts.PrefetchSize = 20
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			var key = string(it.Item().Key())
			if strings.Contains(key, searchTerm) {
				images <- key
			}
		}

		it.Close()
		close(images)
		return nil
	})
}

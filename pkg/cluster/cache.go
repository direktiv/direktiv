package cluster

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/direktiv/direktiv/pkg/dlog"
	"go.uber.org/zap"
)

type Cache struct {
	logger *zap.SugaredLogger
	config CacheConfig

	db *badger.DB
}

const cacheTopic = "cache"

type CacheConfig struct {
	Prefix string

	Node *Node
}

var prefixList sync.Map

func NewCache(config CacheConfig) (*Cache, error) {

	_, exists := prefixList.LoadOrStore(config.Prefix, config.Prefix)
	if exists {
		return nil, fmt.Errorf("prefix %s already exists", config.Prefix)
	}

	logger, err := dlog.ApplicationLogger("cache")
	if err != nil {
		return nil, err
	}

	if config.Prefix == "" || strings.Contains(config.Prefix, "-") {
		return nil, fmt.Errorf("no prefix set or contains -")
	}

	opt := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}

	cache := &Cache{
		db:     db,
		logger: logger,
		config: config,
	}

	if config.Node != nil {
		_, err = config.Node.Subscribe(cacheTopic, cache.invalidateInternal)
	}

	// TODO: logging
	// serfConfig.Logger = zap.NewStdLog(logger.Desugar())

	// run garbage collector
	go gc(db)

	return cache, nil

}

func DefaultCacheConfig() CacheConfig {

	return CacheConfig{
		Prefix: "dummy",
	}

}

func (c *Cache) keyForPrefix(key string) []byte {
	return []byte(fmt.Sprintf("%s-%s", c.config.Prefix, key))
}

func (c *Cache) Set(key string, value []byte, ttl time.Duration) error {

	c.logger.Debugf("setting key %s", string(c.keyForPrefix(key)))

	return c.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(c.keyForPrefix(key), value).WithTTL(ttl)
		if ttl == 0 {
			e = badger.NewEntry(c.keyForPrefix(key), value)
		}
		return txn.SetEntry(e)
	})
}

func (c *Cache) GetFunction(key string, fetch func(string) ([]byte, error),
	ttl time.Duration) ([]byte, error) {

	if fetch == nil {
		return nil, fmt.Errorf("function in cache not set")
	}

	value, err := c.Get(key)

	// checking for key not found error
	// if not found we run the function
	// and set the value
	if errors.Is(err, badger.ErrKeyNotFound) {
		// run function to fetch
		value, err = fetch(key)
		if err != nil {
			return nil, err
		}
		err = c.Set(key, value, ttl)
		if err != nil {
			return nil, err
		}

	} else if err != nil {
		return nil, err
	}

	return value, nil
}

func (c *Cache) Get(key string) ([]byte, error) {

	var value []byte

	c.logger.Debugf("getting key %s", string(c.keyForPrefix(key)))

	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(c.keyForPrefix(key))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			value = append([]byte{}, val...)
			return nil
		})

		return err
	})

	return value, err
}

func (c *Cache) Invalidate(key string) error {
	return c.db.Update(func(txn *badger.Txn) error {

		// tell cluster to invalidate

		c.logger.Debugf("invalidating key %s", string(c.keyForPrefix(key)))

		// send to bus if set
		if c.config.Node != nil {
			msg := fmt.Sprintf("invalidate-%s", string(c.keyForPrefix(key)))
			err := c.config.Node.Publish(cacheTopic, []byte(msg))
			if err != nil {
				c.logger.Errorf("can not publish invalidate message %s", msg)
			}
		}

		return txn.Delete(c.keyForPrefix(key))
	})
}

func (c *Cache) InvalidateAll() error {

	keysToDelete := make([]string, 0)

	c.logger.Debugf("invalidate all with prefix %s", c.config.Prefix)

	// send to bus if set
	if c.config.Node != nil {
		msg := fmt.Sprintf("invalidateAll-%s", c.config.Prefix)
		err := c.config.Node.Publish(cacheTopic, []byte(msg))
		if err != nil {
			c.logger.Errorf("can not publish invalidate message %s", msg)
		}
	}

	err := c.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			keysToDelete = append(keysToDelete, string(item.Key()))
		}

		return nil
	})

	if err != nil {
		return err
	}

	err = c.db.Update(func(txn *badger.Txn) error {
		for i := range keysToDelete {
			k := keysToDelete[i]
			e := txn.Delete([]byte(k))
			if e != nil {
				return e
			}
		}

		return nil
	})

	return err
}

func gc(db *badger.DB) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
	again:
		err := db.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}
}

func (c *Cache) invalidateInternal(key []byte) error {

	keyIn := string(key)

	split := strings.SplitN(keyIn, "-", 3)
	if len(split) != 3 {
		return fmt.Errorf("invalid key for cluster cache")
	}

	fmt.Printf("%s - %s - %s\n", split[0], split[1], split[2])

	// // messages come in in format <cmd>-<cache-name>-<key>
	switch split[0] {
	case "invalidate":
		return c.db.Update(func(txn *badger.Txn) error {
			return txn.Delete(c.keyForPrefix(split[2]))
		})
	case "invalidateAll":
	default:
	}

	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!111")
	fmt.Println(keyIn)
	return nil
}

package secrets

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"
)

type Driver interface {
	ConstructSource(data []byte) Source
	RedactConfig(data []byte) ([]byte, error)
	ValidateConfig(data []byte) error
}

var drivers sync.Map

func RegisterDriver(name string, d Driver) error {
	_, defined := drivers.Load(name)
	if defined {
		return NewJSONMarshalableError(fmt.Errorf("duplicate driver: '%s'", name))
	}

	drivers.Store(name, d)

	return nil
}

func GetDriver(name string) (Driver, error) {
	x, defined := drivers.Load(name)
	if defined {
		d, ok := x.(Driver)
		if !ok {
			panic(errors.New("bad type stored in driver registry"))
		}

		return d, nil
	}

	return nil, NewJSONMarshalableError(fmt.Errorf("driver not found: %s", name))
}

var controllers sync.Map

func RegisterController(namespace string, c Controller) error {
	_, defined := controllers.Load(namespace)
	if defined {
		return NewJSONMarshalableError(fmt.Errorf("duplicate controller: '%s'", namespace))
	}

	controllers.Store(namespace, c)

	return nil
}

func GetController(namespace string) (Controller, error) {
	x, defined := controllers.Load(namespace)
	if defined {
		c, ok := x.(Controller)
		if !ok {
			panic(errors.New("bad type stored in controller registry"))
		}

		return c, nil
	}

	// attempt to automatically load controller
	c, err := LoadController(namespace)
	if err != nil {
		return nil, NewJSONMarshalableError(fmt.Errorf("failed to load controller: %w", err))
	}

	// this function can be called in parallel, so we back off if another thread beat us to the punch
	y, _ := controllers.LoadOrStore(namespace, c)

	c, ok := y.(Controller)
	if !ok {
		panic(errors.New("bad type stored in controller registry"))
	}

	return c, nil
}

func DeleteController(namespace string) {
	x, loaded := controllers.LoadAndDelete(namespace)
	if loaded {
		c, ok := x.(Controller)
		if !ok {
			panic(errors.New("bad type stored in controller registry"))
		}

		if err := c.Delete(); err != nil {
			panic(err)
		}
	}
}

var configGetter = func(namespace string) (*Config, error) {
	return nil, errors.New("default config getter undefined")
}

func GetDefaultConfigGetter() func(namespace string) (*Config, error) {
	return configGetter
}

func SetDefaultConfigGetter(fn func(namespace string) (*Config, error)) {
	configGetter = fn
}

var defaultCacheFactory = func(namespace string) (Cache, error) {
	return &MemoryCache{}, nil
}

func GetDefaultCacheFactory() func(namespace string) (Cache, error) {
	return defaultCacheFactory
}

func SetDefaultCacheFactory(fn func(namespace string) (Cache, error)) {
	defaultCacheFactory = fn
}

func LoadController(namespace string) (Controller, error) {
	config, err := GetDefaultConfigGetter()(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace secrets config: %w", err)
	}

	cache, err := GetDefaultCacheFactory()(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace secrets cache: %w", err)
	}

	return NewController(config, cache), nil
}

// The memory cache exists primarily as a helper for unit testing.
// It will also be the default in case NATS doesn't exist, but we
// intend that in all real use this gets unused because Direktiv
// should call SetDefaultCache to replace this.
type MemoryCache struct {
	lock sync.RWMutex
	list List
}

func (c *MemoryCache) List(_ context.Context) (List, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	list := make(List, len(c.list))
	copy(list, c.list)

	return list, nil
}

func (c *MemoryCache) Insert(_ context.Context, secret Secret) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if secret.ID == "" {
		secret.ID = uuid.New().String()
	}

	for _, entry := range c.list {
		if entry.Path == secret.Path && entry.Source == secret.Source {
			return ErrKeyExists
		}
	}

	list := make(List, len(c.list)+1)

	copy(list, c.list)

	list[len(c.list)] = secret

	sort.Sort(list)

	c.list = list

	return nil
}

func (c *MemoryCache) Delete() error {
	// don't need to do anything here. Garbage collector should handle everything

	return nil
}

package natscache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/direktiv/direktiv/pkg/secrets"
	"github.com/nats-io/nats.go/jetstream"
)

type Cache struct {
	namespace string
	js        jetstream.JetStream
	kv        jetstream.KeyValue
}

func (c *Cache) List(ctx context.Context) (secrets.List, error) {
	keys, err := c.kv.ListKeys(ctx)
	if err != nil {
		return nil, err
	}

	list := make(secrets.List, 0)

	for key := range keys.Keys() {
		list = append(list, secrets.Secret{
			Path: key,
		})
	}

	for idx := range list {
		key := list[idx].Path

		entry, err := c.kv.Get(ctx, key)
		if err != nil {
			return nil, err
		}

		value := entry.Value()

		err = json.Unmarshal(value, &list[idx])
		if err != nil {
			return nil, err
		}
	}

	sort.Sort(list)

	return list, nil
}

func (c *Cache) Insert(ctx context.Context, secret secrets.Secret) error {
	data, _ := json.Marshal(secret)

	_, err := c.kv.Create(ctx, secret.Path, data)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyExists) {
			return secrets.ErrKeyExists
		}

		return err
	}

	return nil
}

func (c *Cache) Delete() error {
	err := c.js.DeleteKeyValue(context.Background(), bucketName(c.namespace))
	if err != nil {
		if errors.Is(err, jetstream.ErrBucketNotFound) {
			return nil
		}

		return err
	}

	return nil
}

func bucketName(namespace string) string {
	return fmt.Sprintf("secrets-%s", namespace)
}

func New(js jetstream.JetStream, namespace string) (secrets.Cache, error) {
	c := &Cache{
		namespace: namespace,
		js:        js,
	}

	kv, err := c.js.CreateKeyValue(context.Background(), jetstream.KeyValueConfig{
		Bucket:      bucketName(c.namespace),
		Description: fmt.Sprintf("Acts as a distributed cache to store looked-up secrets for namespace '%s'", c.namespace),
		TTL:         time.Minute, // TODO: make this configurable
		Storage:     jetstream.MemoryStorage,
	})
	if err != nil {
		if !errors.Is(err, jetstream.ErrBucketExists) {
			return nil, err
		}

		kv, err = c.js.KeyValue(context.Background(), bucketName(c.namespace))
		if err != nil {
			return nil, err
		}
	}

	c.kv = kv

	return c, nil
}

// TODO: unit testing

// Package consumer manages the consumers of the gateway. It can be only updated with a
// set of consumers and not individual consumers. If an individual consumer is getting
// created, updated or deleted this package updates all the consumers.
package consumer

import (
	"log/slog"
	"sync"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

var consumers = newConsumerList()

// consumerList holds different `views` on the consumers for faster lookup.
// That means apiKey is an unique key as well. Duplicate api keys are not allowed.
type consumerList struct {
	apiKeyView   map[string]*core.Consumer
	usernameView map[string]*core.Consumer
	listView     []*core.Consumer

	lock sync.RWMutex
}

func newConsumerList() *consumerList {
	return &consumerList{
		apiKeyView:   make(map[string]*core.Consumer, 0),
		usernameView: make(map[string]*core.Consumer, 0),
		listView:     make([]*core.Consumer, 0),
	}
}

// GetConsumers returns a list of all consumers in the system.
func GetConsumers() []*core.Consumer {
	consumers.lock.RLock()
	defer consumers.lock.RUnlock()

	return consumers.listView
}

// SetConsumers set a new list of consumers in the system. The new list
// is getting swapped out at the end of processing.
func SetConsumer(consumerList []*core.Consumer) {
	newConsumer := newConsumerList()

	for i := range consumerList {
		c := consumerList[i]

		// username is the primary key.
		if c.Username == "" {
			slog.Warn("consumer name empty")

			continue
		}

		// skip duplicates
		_, ok := newConsumer.usernameView[c.Username]
		if ok {
			slog.Warn("consumer already defined",
				slog.String("consumer", c.Username))

			continue
		}

		slog.Info("adding consumer",
			slog.String("user", c.Username))

		// add to list view
		newConsumer.listView = append(newConsumer.listView, c)

		// add to username view
		newConsumer.usernameView[c.Username] = c

		// add to api key view
		if c.APIKey != "" {
			newConsumer.apiKeyView[c.APIKey] = c
		}
	}

	// replace with new consumer list
	consumers.lock.Lock()
	defer consumers.lock.Unlock()
	consumers = newConsumer
}

// FindByUser returns a consumer with the provided name, nil if not found.
func FindByUser(user string) *core.Consumer {
	consumers.lock.RLock()
	defer consumers.lock.RUnlock()

	c, ok := consumers.usernameView[user]
	if !ok {
		return nil
	}

	return c
}

// FindByAPIKey returns a consumer with the provided key, nil if not found.
func FindByAPIKey(key string) *core.Consumer {
	consumers.lock.RLock()
	defer consumers.lock.RUnlock()

	c, ok := consumers.apiKeyView[key]
	if !ok {
		return nil
	}

	return c
}

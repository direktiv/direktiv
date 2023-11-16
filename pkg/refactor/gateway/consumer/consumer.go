package consumer

import (
	"log/slog"
	"sync"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

var consumers = newConsumerList()

type consumerList struct {
	apiKeyView   map[string]*core.Consumer
	usernameView map[string]*core.Consumer

	lock sync.RWMutex
}

func newConsumerList() *consumerList {
	return &consumerList{
		apiKeyView:   make(map[string]*core.Consumer, 0),
		usernameView: make(map[string]*core.Consumer, 0),
	}
}

func SetConsumer(consumerList []*core.Consumer) {
	newConsumer := newConsumerList()

	for i := range consumerList {
		c := consumerList[i]

		_, ok := newConsumer.usernameView[c.Username]
		if ok {
			slog.Warn("consumer already defined",
				slog.String("consumer", c.Username))
		}

		slog.Info("adding consumer",
			slog.String("user", c.Username))

		newConsumer.usernameView[c.Username] = c

		// set api key lookup
		if c.APIKey != "" {
			newConsumer.apiKeyView[c.APIKey] = c
		}
	}

	// replace with new consumer list
	consumers.lock.Lock()
	defer consumers.lock.Unlock()
	consumers = newConsumer
}

func FindByUser(user string) *core.Consumer {
	consumers.lock.RLock()
	defer consumers.lock.RUnlock()

	c, ok := consumers.usernameView[user]
	if !ok {
		return nil
	}

	return c
}

func FindByAPIKey(key string) *core.Consumer {
	consumers.lock.RLock()
	defer consumers.lock.RUnlock()

	c, ok := consumers.apiKeyView[key]
	if !ok {
		return nil
	}

	return c
}

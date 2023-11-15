package consumer

import (
	"fmt"
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

	fmt.Printf("SETTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT %v\n", consumerList)

	newConsumer := newConsumerList()

	for i := range consumerList {
		c := consumerList[i]

		fmt.Println("SETTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT")

		_, ok := newConsumer.usernameView[c.Username]
		if ok {
			slog.Warn("consumer already defined",
				slog.String("consumer", c.Username))
		}

		slog.Info("adding consumer",
			slog.String("user", c.Username))

		newConsumer.usernameView[c.Username] = c

		// set api key lookup
		if c.APIkey != "" {
			newConsumer.apiKeyView[c.APIkey] = c
		}

	}

	// replace with new consumer list
	consumers.lock.Lock()
	defer consumers.lock.Unlock()
	consumers = newConsumer

}

func ConsumerByUser(user string) *core.Consumer {
	consumers.lock.RLock()
	defer consumers.lock.RUnlock()

	fmt.Println(consumers)

	c, ok := consumers.usernameView[user]
	if !ok {
		return nil
	}

	return c
}

func ConsumerByAPIKey(key string) {
	consumers.lock.RLock()
	defer consumers.lock.RUnlock()
}

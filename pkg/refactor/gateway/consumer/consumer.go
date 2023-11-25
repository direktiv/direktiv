// Package consumer manages the consumers of the gateway. It can be only updated with a
// set of consumers and not individual consumers. If an individual consumer is getting
// created, updated or deleted this package updates all the consumers.
package consumer

import (
	"log/slog"
	"sync"

	"github.com/direktiv/direktiv/pkg/refactor/spec"
)

// consumerList holds different `views` on the consumers for faster lookup.
// That means apiKey is an unique key as well. Duplicate api keys are not allowed.
type List struct {
	apiKeyView   map[string]*spec.ConsumerFile
	usernameView map[string]*spec.ConsumerFile
	listView     []*spec.ConsumerFile

	lock sync.RWMutex
}

func NewConsumerList() *List {
	return &List{
		apiKeyView:   make(map[string]*spec.ConsumerFile, 0),
		usernameView: make(map[string]*spec.ConsumerFile, 0),
		listView:     make([]*spec.ConsumerFile, 0),
	}
}

// GetConsumers returns a list of all consumers in the system.
func (cl *List) GetConsumers() []*spec.ConsumerFile {
	cl.lock.RLock()
	defer cl.lock.RUnlock()

	return cl.listView
}

// SetConsumers set a new lists of consumers in the system. The new lists
// is getting swapped out at the end of processing.
func (cl *List) SetConsumers(consumerList []*spec.ConsumerFile) {
	apiKeyView := make(map[string]*spec.ConsumerFile, 0)
	usernameView := make(map[string]*spec.ConsumerFile, 0)
	listView := make([]*spec.ConsumerFile, 0)

	for i := range consumerList {
		c := consumerList[i]

		// empty it for API response
		c.DirektivAPI = ""

		// username is the primary key.
		if c.Username == "" {
			slog.Warn("consumer name empty")

			continue
		}

		// skip duplicates
		_, ok := usernameView[c.Username]
		if ok {
			slog.Warn("consumer already defined",
				slog.String("consumer", c.Username))

			continue
		}

		slog.Info("adding consumer",
			slog.String("user", c.Username))

		// add to list view
		listView = append(listView, c)

		// add to username view
		usernameView[c.Username] = c

		// add to api key view
		if c.APIKey != "" {
			apiKeyView[c.APIKey] = c
		}
	}

	// replace with new consumer lists
	cl.lock.Lock()
	defer cl.lock.Unlock()

	cl.apiKeyView = apiKeyView
	cl.listView = listView
	cl.usernameView = usernameView
}

// FindByUser returns a consumer with the provided name, nil if not found.
func (cl *List) FindByUser(user string) *spec.ConsumerFile {
	cl.lock.RLock()
	defer cl.lock.RUnlock()

	c, ok := cl.usernameView[user]
	if !ok {
		return nil
	}

	return c
}

// FindByAPIKey returns a consumer with the provided key, nil if not found.
func (cl *List) FindByAPIKey(key string) *spec.ConsumerFile {
	cl.lock.RLock()
	defer cl.lock.RUnlock()

	c, ok := cl.apiKeyView[key]
	if !ok {
		return nil
	}

	return c
}

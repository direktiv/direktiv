package databus

import (
	"sync"

	"github.com/direktiv/direktiv/internal/engine"
	"github.com/google/uuid"
)

type instanceNotifier struct {
	mu       sync.Mutex
	watchers map[string]chan<- *engine.InstanceStatus
}

func newInstanceNotifier() *instanceNotifier {
	return &instanceNotifier{watchers: make(map[string]chan<- *engine.InstanceStatus)}
}

func (n *instanceNotifier) Add(id uuid.UUID, ch chan<- *engine.InstanceStatus) {
	n.mu.Lock()
	n.watchers[id.String()] = ch
	n.mu.Unlock()
}

func (n *instanceNotifier) Notify(id uuid.UUID, status *engine.InstanceStatus) {
	n.mu.Lock()
	ch, ok := n.watchers[id.String()]
	if ok {
		delete(n.watchers, id.String())
	}
	n.mu.Unlock()

	if ok && ch != nil {
		ch <- status
		close(ch)
	}
}

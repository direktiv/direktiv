package pubsub

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
)

type Bus struct {
	coreBus CoreBus

	subscribers  sync.Map
	fingerprints sync.Map
}

const defaultDebouncePublishDuration = 200 * time.Millisecond

func NewBus(coreBus CoreBus) *Bus {
	return &Bus{
		coreBus: coreBus,
	}
}

func (p *Bus) Start(circuit *core.Circuit) {
	p.coreBus.Loop(circuit.Done(), func(channel string, data string) {
		p.subscribers.Range(func(key, f any) bool {
			k, _ := key.(string)
			h, _ := f.(func(data string))

			if strings.HasPrefix(k, channel) {
				go h(data)
			}

			return true
		})
	})
}

func (p *Bus) Publish(channel string, data string) error {
	return p.coreBus.Publish(channel, data)
}

func (p *Bus) debouncedPublishWithInterval(i time.Duration, channel string, data string) error {
	// This function works by associating input with a signature, sleep for a duration and nly publish the message
	// when the signature matches.

	input := fmt.Sprintf("%d_%s_%s", i, channel, data)
	signature := uuid.New()
	p.fingerprints.Store(input, signature)

	go func() {
		time.Sleep(i)
		currentSignature, _ := p.fingerprints.Load(input)
		// When signature matches, this means no later async publish was recorded.
		if signature == currentSignature {
			_ = p.coreBus.Publish(channel, data)
		}
	}()

	return nil
}

// DebouncedPublish prevents multiple concussive publishes of the same input during an interval.
func (p *Bus) DebouncedPublish(channel string, data string) error {
	return p.debouncedPublishWithInterval(defaultDebouncePublishDuration, channel, data)
}

func (p *Bus) Subscribe(handler func(data string), channels ...string) {
	for _, channel := range channels {
		p.subscribers.Store(fmt.Sprintf("%s_%s", channel, uuid.New().String()), handler)
	}
}

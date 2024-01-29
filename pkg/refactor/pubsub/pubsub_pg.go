package pubsub

import (
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Bus struct {
	coreBus CoreBus

	subscribers sync.Map
	logger      *zap.SugaredLogger
}

func NewBus(logger *zap.SugaredLogger, coreBus CoreBus) *Bus {
	return &Bus{
		coreBus: coreBus,
		logger:  logger,
	}
}

func (p *Bus) Start(done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	p.coreBus.Loop(done, p.logger, func(channel string, data string) {
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

func (p *Bus) Subscribe(handler func(data string), channels ...string) {
	for _, channel := range channels {
		p.subscribers.Store(fmt.Sprintf("%s_%s", channel, uuid.New().String()), handler)
	}
}

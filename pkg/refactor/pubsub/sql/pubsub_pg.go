package sql

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PostgresBus struct {
	coreBus *postgresCoreBus

	subscribers sync.Map
	logger      *zap.SugaredLogger
}

func NewPostgresBus(logger *zap.SugaredLogger, db *sql.DB, listenConnectionString string) (*PostgresBus, error) {
	coreBus, err := newPostgresCoreBus(db, listenConnectionString)
	if err != nil {
		return nil, err
	}

	p := &PostgresBus{
		coreBus: coreBus,
		logger:  logger,
	}

	return p, nil
}

func (p *PostgresBus) Start(done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	p.coreBus.forEachMessage(done, p.logger, func(channel string, data string) {
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

func (p *PostgresBus) Publish(channel string, data string) error {
	return p.coreBus.publish(channel, data)
}

func (p *PostgresBus) Subscribe(handler func(data string), channels ...string) {
	for _, channel := range channels {
		p.subscribers.Store(fmt.Sprintf("%s_%s", channel, uuid.New().String()), handler)
	}
}

var _ pubsub.Bus = &PostgresBus{}

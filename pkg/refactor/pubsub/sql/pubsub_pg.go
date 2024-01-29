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
	for {
		select {
		case msg := <-p.coreBus.NotifyChannel():
			channel, data, err := splitNotificationText(msg.Extra)
			if err != nil {
				p.logger.Error("parsing notify message", "msg", msg.Extra, "err", err)
			} else {
				p.subscribers.Range(func(key, f any) bool {
					k, _ := key.(string)
					h, _ := f.(func(data string))

					if strings.HasPrefix(k, channel) {
						go h(data)
					}

					return true
				})
			}
		case <-done:
			return
		}
	}
}

func (p *PostgresBus) Publish(channel string, data string) error {
	return p.coreBus.Publish(channel, data)
}

func (p *PostgresBus) Subscribe(handler func(data string), channels ...string) {
	for _, channel := range channels {
		p.subscribers.Store(fmt.Sprintf("%s_%s", channel, uuid.New().String()), handler)
	}
}

func splitNotificationText(text string) (string, string, error) {
	firstSpaceIndex := strings.IndexAny(text, " ")
	if firstSpaceIndex < 0 {
		return "", "", fmt.Errorf("no space in message: text: >%s<", text)
	}

	return text[:firstSpaceIndex], text[firstSpaceIndex+1:], nil
}

var _ pubsub.Bus = &PostgresBus{}

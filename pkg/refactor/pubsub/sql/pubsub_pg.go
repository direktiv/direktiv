package sql

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const globalPostgresChannel = "direktiv_pubsub_events"

type PostgresBus struct {
	listener    *pq.Listener
	db          *sql.DB
	subscribers sync.Map

	logger *zap.SugaredLogger
}

func NewPostgresBus(logger *zap.SugaredLogger, db *sql.DB, listenConnectionString string) (*PostgresBus, error) {
	p := &PostgresBus{
		db:     db,
		logger: logger,
	}

	p.listener = pq.NewListener(listenConnectionString, time.Second, time.Second,
		func(event pq.ListenerEventType, err error) {
			// do nothing.
		})

	var err error
	// try ping up to 10 times.
	for i := 0; i < 10; i++ {
		err = p.listener.Ping()
		if err != nil {
			time.Sleep(time.Second)

			continue
		}

		break
	}
	if err != nil {
		return nil, fmt.Errorf("ping connection, err: %w", err)
	}
	err = p.listener.Listen(globalPostgresChannel)
	if err != nil {
		return nil, fmt.Errorf("listen to direktiv_pubsub_events channel, err: %w", err)
	}

	return p, nil
}

func (p *PostgresBus) Start(done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case msg := <-p.listener.Notify:
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
	if channel == "" || strings.Contains(channel, " ") {
		return fmt.Errorf("channel name is empty or has spaces: >%s<", channel)
	}
	_, err := p.db.Exec(fmt.Sprintf("NOTIFY %s, '%s %s'", globalPostgresChannel, channel, data))
	if err != nil {
		return fmt.Errorf("send notify command, channel: %s, data: %v, err: %w", channel, data, err)
	}

	return nil
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

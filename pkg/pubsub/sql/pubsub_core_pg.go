package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/lib/pq"
)

type postgresBus struct {
	listener  *pq.Listener
	errorChan chan error
	db        *sql.DB
}

func NewPostgresCoreBus(db *sql.DB, listenConnectionString string) (pubsub.CoreBus, error) {
	p := &postgresBus{
		db:        db,
		errorChan: make(chan error),
	}

	p.listener = pq.NewListener(listenConnectionString, time.Second, time.Second,
		func(event pq.ListenerEventType, err error) {
			p.errorChan <- err
		})

	var err error
	// try ping up to 10 times.
	for range 10 {
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

	return p, nil
}

func (p *postgresBus) Publish(channel string, data string) error {
	if channel == "" || strings.Contains(channel, " ") {
		return fmt.Errorf("channel name is empty or has spaces: >%s<", channel)
	}
	_, err := p.db.Exec(fmt.Sprintf("NOTIFY %s, '%s'", channel, data))
	if err != nil {
		return fmt.Errorf("send notify command, channel: %s, data: %v, err: %w", channel, data, err)
	}

	return nil
}

func (p *postgresBus) Listen(channel string) error {
	err := p.listener.Listen(channel)
	if !errors.Is(err, pq.ErrChannelAlreadyOpen) {
		return err
	}

	return nil
}

func (p *postgresBus) Loop(done <-chan struct{}, handler func(channel string, data string)) error {
	for {
		select {
		case msg := <-p.listener.Notify:
			slog.Debug("pubsub core: received notify message", "channel", msg.Channel, "msg", ">"+msg.Extra+"<")
			handler(msg.Channel, msg.Extra)
		case <-done:
			return nil
		case err := <-p.errorChan:
			if err != nil {
				return fmt.Errorf("database connection, err: %w", err)
			}
		}
	}
}

var _ pubsub.CoreBus = &postgresBus{}

package sql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

const globalPostgresChannel = "direktiv_pubsub_events"

type postgresCoreBus struct {
	listener *pq.Listener
	db       *sql.DB
}

func newPostgresCoreBus(db *sql.DB, listenConnectionString string) (*postgresCoreBus, error) {
	p := &postgresCoreBus{
		db: db,
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

func (p *postgresCoreBus) NotifyChannel() <-chan *pq.Notification {
	return p.listener.Notify
}

func (p *postgresCoreBus) Publish(channel string, data string) error {
	if channel == "" || strings.Contains(channel, " ") {
		return fmt.Errorf("channel name is empty or has spaces: >%s<", channel)
	}
	_, err := p.db.Exec(fmt.Sprintf("NOTIFY %s, '%s %s'", globalPostgresChannel, channel, data))
	if err != nil {
		return fmt.Errorf("send notify command, channel: %s, data: %v, err: %w", channel, data, err)
	}

	return nil
}

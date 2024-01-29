package sql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const globalPostgresChannel = "direktiv_pubsub_events"

type postgresBus struct {
	listener *pq.Listener
	db       *sql.DB
}

func NewPostgresCoreBus(db *sql.DB, listenConnectionString string) (pubsub.CoreBus, error) {
	p := &postgresBus{
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

func (p *postgresBus) Publish(channel string, data string) error {
	if channel == "" || strings.Contains(channel, " ") {
		return fmt.Errorf("channel name is empty or has spaces: >%s<", channel)
	}
	_, err := p.db.Exec(fmt.Sprintf("NOTIFY %s, '%s %s'", globalPostgresChannel, channel, data))
	if err != nil {
		return fmt.Errorf("send notify command, channel: %s, data: %v, err: %w", channel, data, err)
	}

	return nil
}

func (p *postgresBus) Loop(done <-chan struct{}, logger *zap.SugaredLogger, handler func(channel string, data string)) {
	for {
		select {
		case msg := <-p.listener.Notify:
			channel, data, err := splitNotificationText(msg.Extra)
			if err != nil {
				logger.Error("parsing notify message", "msg", msg.Extra, "err", err)
			} else {
				handler(channel, data)
			}
		case <-done:
			return
		}
	}
}

func splitNotificationText(text string) (string, string, error) {
	firstSpaceIndex := strings.IndexAny(text, " ")
	if firstSpaceIndex < 0 {
		return "", "", fmt.Errorf("no space in message: text: >%s<", text)
	}

	return text[:firstSpaceIndex], text[firstSpaceIndex+1:], nil
}

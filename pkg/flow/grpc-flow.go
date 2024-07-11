package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/database"
)

type flow struct {
	*server
}

const srv = "server"

func initFlowServer(ctx context.Context, srv *server) (*flow, error) {
	var err error

	flow := &flow{server: srv}

	go func() { //nolint:contextcheck
		// instance garbage collector
		ctx := context.Background()
		<-time.After(2 * time.Minute)

		for {
			<-time.After(time.Hour)
			t := time.Now().UTC().Add(time.Hour * -24)

			tx, err := srv.flow.beginSQLTx(ctx)
			if err != nil {
				slog.Error("garbage collector", "error", fmt.Errorf("failed to get transaction to cleanup old instances: %w", err))
				continue
			}

			err = tx.InstanceStore().DeleteOldInstances(ctx, t)
			if err != nil {
				tx.Rollback()
				slog.Error("garbage collector", "error", fmt.Errorf("failed to cleanup old instances: %w", err))

				continue
			}

			err = tx.Commit(ctx)
			if err != nil {
				slog.Error("garbage collector", "error", fmt.Errorf("failed to commit tx to cleanup old instances: %w", err))

				continue
			}

			// TODO: alan: cleanup old instance variables.
		}
	}()

	go func() {
		// logs garbage collector
		<-time.After(3 * time.Minute)
		for {
			<-time.After(time.Hour)
			t := time.Now().UTC().Add(time.Hour * -48) // TODO make this a config option.
			slog.Debug("deleting all logs since", "since", t)
			err = srv.flow.runSQLTx(ctx, func(tx *database.SQLStore) error {
				return tx.DataStore().NewLogs().DeleteOldLogs(ctx, t)
			})
			if err != nil {
				slog.Error("garbage collector", "error", fmt.Errorf("failed to cleanup old logs: %w", err))
				continue
			}
		}
	}()

	go func() { //nolint:contextcheck
		// timed-out instance retrier
		<-time.After(1 * time.Minute)
		ticker := time.NewTicker(5 * time.Minute)
		for {
			<-ticker.C
			go flow.kickExpiredInstances()
		}
	}()

	srv.pBus.Subscribe(configureRouterMessage{}, flow.configureRouterHandler)

	return flow, nil
}

func (flow *flow) kickExpiredInstances() {
	ctx := context.Background()

	tx, err := flow.beginSQLTx(ctx)
	if err != nil {
		slog.Error("Failed to begin SQL transaction in kickExpiredInstances.", "error", err)
		return
	}
	defer tx.Rollback()

	list, err := tx.InstanceStore().GetHangingInstances(ctx)
	if err != nil {
		slog.Error("Failed to retrieve hanging instances.", "error", err)
		return
	}

	for i := range list {
		data, err := json.Marshal(&retryMessage{
			InstanceID: list[i].ID.String(),
		})
		if err != nil {
			slog.Error("Failed to marshal retry message for instance.", "error", err)
			panic(err) // TODO ?
		}

		flow.engine.retryWakeup(data)
	}
}

func (flow *flow) GetAttributes() map[string]string {
	tags := make(map[string]string)
	tags["recipientType"] = srv

	return tags
}

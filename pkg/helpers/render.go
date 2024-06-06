package helpers

import (
	"context"
	slog2 "log/slog"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/filestore"
)

func RenderGatewayFiles(db *database.SQLStore, manager core.GatewayManager) {
	ctx := context.Background()
	slog := slog2.With("subscriber", "gateway file watcher")

	fStore, dStore := db.FileStore(), db.DataStore()

	nsList, err := dStore.Namespaces().GetAll(ctx)
	if err != nil {
		slog.Error("listing namespaces", "err", err)

		return
	}

	var consumers []core.ConsumerV2
	var endpoints []core.EndpointV2

	for _, ns := range nsList {
		slog = slog.With("namespace", ns.Name)
		files, err := fStore.ForNamespace(ns.Name).ListDirektivFilesWithData(ctx)
		if err != nil {
			slog.Error("listing direktiv files", "err", err)

			continue
		}
		for _, file := range files {
			if file.Typ == filestore.FileTypeConsumer {
				consumers = append(consumers, core.ParseConsumerFileV2(ns.Name, file.Path, file.Data))
			} else if file.Typ == filestore.FileTypeEndpoint {
				endpoints = append(endpoints, core.ParseEndpointFileV2(ns.Name, file.Path, file.Data))
			}
		}
	}
	manager.SetEndpoints(endpoints, consumers)
}

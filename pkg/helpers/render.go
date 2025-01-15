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

	var consumers []core.Consumer
	var endpoints []core.Endpoint
	var gateways []core.Gateway

	for _, ns := range nsList {
		slog = slog.With("namespace", ns.Name)
		files, err := fStore.ForNamespace(ns.Name).ListDirektivFilesWithData(ctx)
		if err != nil {
			slog.Error("listing direktiv files", "err", err)

			continue
		}
		for _, file := range files {
			if file.Typ == filestore.FileTypeConsumer {
				consumers = append(consumers, core.ParseConsumerFile(ns.Name, file.Path, file.Data))
			} else if file.Typ == filestore.FileTypeEndpoint {
				// ep := core.ParseEndpointFile(ns.Name, file.Path, file.Data)
				// err = validateEndpoint(&ep, ns.Name, fStore)
				// if err != nil {
				// 	ep.Errors = append(ep.Errors, err.Error())
				// }
				// endpoints = append(endpoints, ep)
				endpoints = append(endpoints, core.ParseEndpointFile(ns.Name, file.Path, file.Data))
			} else if file.Typ == filestore.FileTypeGateway {
				gateways = append(gateways, core.ParseGatewayFile(ns.Name, file.Path, file.Data))
			}
		}
	}
	err = manager.SetEndpoints(endpoints, consumers, gateways)
	if err != nil {
		slog.Error("render gateway files", "err", err)
	}
}

package server

import (
	"context"
	"log/slog"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/filestore/filesql"
	"gorm.io/gorm"
)

func renderGatewayFiles(db *gorm.DB, manager core.GatewayManager) {
	ctx := context.Background()
	sLog := slog.With("subscriber", "gateway file watcher")
	fStore, dStore := filesql.NewStore(db), datasql.NewStore(db)

	nsList, err := dStore.Namespaces().GetAll(ctx)
	if err != nil {
		slog.Error("listing namespaces", "err", err)

		return
	}

	var consumers []core.Consumer
	var endpoints []core.Endpoint
	var gateways []core.Gateway

	for _, ns := range nsList {
		sLog = sLog.With("namespace", ns.Name)
		files, dataList, err := fStore.ForRoot(ns.Name).ListDirektivFilesWithData(ctx)
		if err != nil {
			slog.Error("listing direktiv files", "err", err)

			continue
		}
		for i, file := range files {
			data := dataList[i]
			//nolint:exhaustive
			switch file.Typ {
			case filestore.FileTypeConsumer:
				consumers = append(consumers, core.ParseConsumerFile(ns.Name, file.Path, data))
			case filestore.FileTypeEndpoint:
				endpoints = append(endpoints, core.ParseEndpointFile(ns.Name, file.Path, data))
			case filestore.FileTypeGateway:
				gateways = append(gateways, core.ParseGatewayFile(ns.Name, file.Path, data))
			}
		}
	}
	err = manager.SetEndpoints(endpoints, consumers, gateways)
	if err != nil {
		slog.Error("render gateway files", "err", err)
	}
}

func renderServiceFiles(db *gorm.DB, serviceManager core.ServiceManager) {
	// TODO: fix nil data params below.
	return
}

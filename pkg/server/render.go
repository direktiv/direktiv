package server

import (
	"context"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/filestore"
)

func renderGatewayFiles(db *database.DB, manager core.GatewayManager) {
	//TODO: fix nil data params below.
	return
	ctx := context.Background()
	sLog := slog.With("subscriber", "gateway file watcher")

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
		sLog = sLog.With("namespace", ns.Name)
		files, err := fStore.ForNamespace(ns.Name).ListDirektivFilesWithData(ctx)
		if err != nil {
			slog.Error("listing direktiv files", "err", err)

			continue
		}
		for _, file := range files {
			//nolint:exhaustive
			switch file.Typ {
			case filestore.FileTypeConsumer:
				consumers = append(consumers, core.ParseConsumerFile(ns.Name, file.Path, nil))
			case filestore.FileTypeEndpoint:
				endpoints = append(endpoints, core.ParseEndpointFile(ns.Name, file.Path, nil))
			case filestore.FileTypeGateway:
				gateways = append(gateways, core.ParseGatewayFile(ns.Name, file.Path, nil))
			}
		}
	}
	err = manager.SetEndpoints(endpoints, consumers, gateways)
	if err != nil {
		slog.Error("render gateway files", "err", err)
	}
}

func renderServiceFiles(db *database.DB, serviceManager core.ServiceManager) {
	//TODO: fix nil data params below.
	return
	ctx := context.Background()
	slog := slog.With("subscriber", "services file watcher")

	fStore, dStore := db.FileStore(), db.DataStore()

	nsList, err := dStore.Namespaces().GetAll(ctx)
	if err != nil {
		slog.Error("listing namespaces", "error", err)

		return
	}

	funConfigList := []*core.ServiceFileData{}

	for _, ns := range nsList {
		slog = slog.With("namespace", ns.Name)
		files, err := fStore.ForNamespace(ns.Name).ListDirektivFilesWithData(ctx)
		if err != nil {
			slog.Error("listing direktiv files", "error", err)

			continue
		}
		for _, file := range files {
			if file.Typ == filestore.FileTypeService {
				serviceDef, err := core.ParseServiceFile(nil)
				if err != nil {
					slog.Error("parse service file", "error", err)

					continue
				}
				typ := core.ServiceTypeNamespace
				if ns.Name == core.SystemNamespace {
					typ = core.ServiceTypeSystem
				}
				funConfigList = append(funConfigList, &core.ServiceFileData{
					Typ:         typ,
					Name:        "",
					Namespace:   ns.Name,
					FilePath:    file.Path,
					ServiceFile: *serviceDef,
				})
			}
		}
	}
	serviceManager.SetServices(funConfigList)
}

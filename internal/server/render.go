package server

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/direktiv/direktiv/internal/cluster/cache"
	"github.com/direktiv/direktiv/internal/compiler"
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

func renderServiceFiles(db *gorm.DB, serviceManager core.ServiceManager,
	cacheManager cache.Manager, secretsManager core.SecretsManager) {
	ctx := context.Background()
	dStore := datasql.NewStore(db)

	namespaces, err := dStore.Namespaces().GetAll(ctx)
	if err != nil {
		slog.Error("cannot render files", slog.Any("error", err))
		return
	}

	fStore := filesql.NewStore(db)

	funConfigList := []*core.ServiceFileData{}
	for i := range namespaces {
		ns := namespaces[i]
		files, err := fStore.ForRoot(ns.Name).ListAllFiles(ctx)
		if err != nil {
			slog.Error("cannot get namespace",
				slog.String("name", ns.Name), slog.Any("error", err))

			continue
		}

		for a := range files {
			f := files[a]

			switch f.Typ {
			case filestore.FileTypeService:
				f, err := filesql.NewStore(db).ForRoot(ns.Name).GetFile(ctx, f.Path)
				if err != nil {
					slog.Error("cannot find service file", slog.String("path", f.Path),
						slog.Any("error", err))

					continue
				}

				svc, err := filesql.NewStore(db).ForFile(f).GetData(ctx)
				if err != nil {
					slog.Error("cannot load service file", slog.String("path", f.Path),
						slog.Any("error", err))

					continue
				}

				var ac core.ActionConfig
				err = json.Unmarshal(svc, &ac)
				if err != nil {
					slog.Error("cannot marshal service file", slog.String("path", f.Path),
						slog.Any("error", err))

					continue
				}

				if ac.Image == "" {
					slog.Error("no image defined in service file", slog.String("path", f.Path))

					continue
				}

				funConfigList = append(funConfigList, svcFile(ac, ns.Name, f.Path))

			case filestore.FileTypeWorkflow:
				c, err := compiler.NewCompiler(db, cacheManager.FlowCache())
				if err != nil {
					slog.Error("cannot get compiler for workflow",
						slog.String("namespace", ns.Name),
						slog.String("path", f.Path), slog.Any("error", err))

					continue
				}
				s, err := c.FetchScript(ctx, ns.Name, f.Path)
				if err != nil {
					slog.Error("cannot generate script",
						slog.String("namespace", ns.Name),
						slog.String("path", f.Path), slog.Any("error", err))

					continue
				}

				// setup secrets
				for i := range s.Config.Secrets {
					secret := s.Config.Secrets[i]

					// we create the secrets. if they exists it fails and we ignore the error
					// if they don't we set them as empty
					_, err := secretsManager.Create(ctx, ns.Name, &core.Secret{
						Name: secret,
						Data: []byte{},
					})
					if err != nil {
						slog.Warn("could not create secret", slog.Any("error", err))
					}
				}

				// to make it unique for flow actions, we use a hash as name
				for k := range s.Config.Actions {
					action := s.Config.Actions[k]

					sf := core.ServiceFile{
						Image: action.Image,
						Cmd:   action.Cmd,
						Size:  action.Size,
						Envs:  action.Envs,
						// Patches: action.Patches,
						// TODO: this need to be set to zero to enable zero scaling.
						Scale: 1,
					}

					sd := &core.ServiceFileData{
						Typ:         core.ServiceTypeWorkflow,
						Name:        "",
						Namespace:   ns.Name,
						FilePath:    f.Path,
						ServiceFile: sf,
					}

					// set name for workflow action
					sd.Name = sd.GetValueHash()

					funConfigList = append(funConfigList, sd)
				}
			}
		}
	}

	serviceManager.SetServices(funConfigList)
}

func svcFile(action core.ActionConfig, namespace, path string) *core.ServiceFileData {
	sf := core.ServiceFile{
		Image: action.Image,
		Cmd:   action.Cmd,
		Size:  action.Size,
		Envs:  action.Envs,
		Scale: 1,
	}

	return &core.ServiceFileData{
		Typ:         core.ServiceTypeWorkflow,
		Name:        "",
		Namespace:   namespace,
		FilePath:    path,
		ServiceFile: sf,
	}
}

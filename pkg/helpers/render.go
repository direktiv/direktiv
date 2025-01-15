package helpers

import (
	"context"
	slog2 "log/slog"
	"net/url"
	"path/filepath"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/getkin/kin-openapi/openapi3"
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

func validateEndpoint(ep *core.Endpoint, ns string, fileStore filestore.FileStore) error {
	l := openapi3.NewLoader()
	l.IsExternalRefsAllowed = true
	l.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		// get relative from file

		// p, err := filepath.Rel(filepath.Dir(ep.FilePath), url.String())
		// fmt.Println(ep.FilePath)
		// fmt.Println(url.String())

		path := url.String()

		// if not absolute we need to calculate path
		if !filepath.IsAbs(url.String()) {
			p, err := filepath.Rel(filepath.Dir(ep.FilePath),
				filepath.Join(filepath.Dir(ep.FilePath), url.String()))
			if err != nil {
				return nil, err
			}
			path = p
		}

		file, err := fileStore.ForNamespace(ns).GetFile(context.Background(), path)
		if err != nil {
			return nil, err
		}
		return fileStore.ForFile(file).GetData(context.Background())
	}

	// create fake doc for validation
	doc := &openapi3.T{
		Paths:   openapi3.NewPaths(openapi3.WithPath(ep.Config.Path, &ep.RenderedPathItem)),
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "dummy",
			Version: "1.0.0",
		},
	}

	err := l.ResolveRefsIn(doc, nil)
	if err != nil {
		return err
	}

	// validate the whole thing
	return ep.RenderedPathItem.Validate(context.Background())
}

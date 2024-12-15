package api

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
)

type gatewayDocsController struct {
	fstore filestore.FileStore
}

func (c *gatewayDocsController) mountRouter(r chi.Router) {
	r.Get("/", c.get) // list the openapi docs for the gateway routes of the current namepsace
}

func (c *gatewayDocsController) get(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	paths := map[string]core.PathItem{}
	files, err := c.fstore.ForNamespace(ns.Name).ListDirektivFilesWithData(r.Context())
	if err != nil {
		writeDataStoreError(w, err)
	}
	for _, file := range files {
		if file.Typ == filestore.FileTypeAPIPath {
			var p core.PathItem
			err := yaml.Unmarshal(file.Data, &p)
			if err != nil {
				writeDataStoreError(w, err)
			}
			paths["/api/v2/namespaces/"+ns.Name+core.ExtractAPIPath(&p)] = p
		}
	}

	spec := core.OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: core.OpenAPIInfo{
			Title:       "Example API",
			Description: "This is an example API to test OpenAPI specs",
			Version:     "1.0.0",
		},
		Servers: []core.OpenAPIServer{
			{
				URL:         "https://api.example.com/v1",
				Description: "Production server",
			},
		},
		Paths: paths,
	}

	writeJSON(w, spec)
}

package tsengine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/tsengine/tsservice"
	"golang.org/x/exp/slog"
)

// Manager handles compilation and generation of TypeScript workflow services.
type Manager struct {
	db     *database.SQLStore
	config core.Config
}

// NewManager creates a new Manager instance.
func NewManager(db *database.SQLStore, config core.Config) *Manager {
	return &Manager{
		db:     db,
		config: config,
	}
}

// Run iterates over namespaces and TypeScript workflows, compiling and generating service files.
func (m Manager) Run(cir *core.Circuit) error {
	ctx := cir.Context()

	nsList, err := m.db.DataStore().Namespaces().GetAll(ctx)
	if err != nil {
		return fmt.Errorf("listing namespaces: %w", err)
	}

	for _, ns := range nsList {
		log := slog.With("namespace", ns.Name)

		files, err := m.db.FileStore().ForNamespace(ns.Name).ListDirektivFilesWithData(ctx)
		if err != nil {
			log.Error("listing direktiv files", "err", err)
			continue
		}

		for _, file := range files {
			if file.Typ != filestore.FileTypeTSWorkflow {
				continue
			}

			if err := m.processTSFile(ctx, ns.Name, file); err != nil {
				log.Error("processing ts file", "path", file.Path, "err", err)
			}
		}
	}

	return nil
}

// processTSFile handles the compilation and generation of service files for a single TypeScript workflow.
func (m *Manager) processTSFile(ctx context.Context, namespace string, file *filestore.File) error {
	log := slog.With("namespace", namespace, "file", file.Path)

	compiler, err := tsservice.NewTSServiceCompiler(namespace, file.Path, string(file.Data))
	if err != nil {
		return fmt.Errorf("creating tsfile compiler: %w", err)
	}

	flowInfo, err := compiler.CompileFlow()
	if err != nil {
		return fmt.Errorf("compiling tsfile: %w", err)
	}

	if len(flowInfo.Definition.Scale) == 0 {
		return fmt.Errorf("returned bad scale")
	}

	scale := flowInfo.Definition.Scale[0]
	serviceData := &core.ServiceFileData{
		ID:        flowInfo.ID,
		Typ:       "service/v1",
		Namespace: namespace,
		FilePath:  file.Path,
		Error:     nil,

		Name: flowInfo.ID,
		ServiceFile: core.ServiceFile{
			DirektivAPI: "service/v1",
			Scale:       scale.Min,
			Image:       m.config.KnativeSidecar,
			Cmd:         "",
			Size:        "small",
			Patches:     []core.ServicePatch{},
			Envs: []core.EnvironmentVariable{
				{
					Name:  "DIREKTIV_APP",
					Value: "tsengine",
				},
				{
					Name:  "DIREKTIV_JSENGINE_NAMESPACE",
					Value: namespace,
				},
				{
					Name:  "DIREKTIV_JSENGINE_WORKFLOW_PATH",
					Value: file.Path,
				},
				{
					Name:  "DIREKTIV_DB",
					Value: m.config.DB,
				},
				{
					Name:  "DIREKTIV_SECRET_KEY",
					Value: m.config.ApiKey,
				},
			},
			// TODO: fill all fields.
		},
	}

	data, err := json.Marshal(serviceData)
	if err != nil {
		return fmt.Errorf("marshalling service data: %w", err)
	}

	_, err = m.db.FileStore().ForNamespace(namespace).CreateFile(
		ctx, file.Path+".yaml", filestore.FileTypeService, "application/yaml", data,
	)
	if err != nil {
		return fmt.Errorf("creating yaml file: %w", err)
	}

	log.Info("successfully processed ts file")

	return nil
}

func GenerateBasicServiceFile(path, ns string) *core.ServiceFileData {
	return &core.ServiceFileData{
		Typ:       core.ServiceTypeTypescript,
		Name:      path,
		Namespace: ns,
		FilePath:  path,
	}
}

package tsengine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/tsengine/tsservice"
)

type Manager struct {
	db     *database.SQLStore
	config core.Config
}

func NewManager(db *database.SQLStore, config core.Config) *Manager {
	return &Manager{
		db:     db,
		config: config,
	}
}

func (m Manager) Create(cir *core.Circuit, namespace string, filePath string, fileType string) error {
	if fileType != string(filestore.FileTypeTSWorkflow) {
		return nil // Not a TypeScript workflow file, no action needed
	}

	ctx := cir.Context()

	file, err := m.db.FileStore().ForNamespace(namespace).GetFile(ctx, filePath)
	if err != nil {
		return err
	}
	if err := m.processTSFile(ctx, namespace, file); err != nil {
		return err
	}

	return nil
}

func (m Manager) Update(cir *core.Circuit, namespace string, filePath string, fileType string) error {
	if fileType != string(filestore.FileTypeTSWorkflow) {
		return nil // no action here
	}
	err := m.Delete(cir, namespace, filePath, fileType)
	if err != nil {
		return err
	}
	err = m.Create(cir, namespace, filePath, fileType)
	if err != nil {
		return err
	}

	return nil
}

func (m Manager) Delete(cir *core.Circuit, namespace string, filePath string, fileType string) error {
	if fileType != string(filestore.FileTypeTSWorkflow) {
		return nil // no action here
	}

	ctx := cir.Context()

	file, err := m.db.FileStore().ForNamespace(namespace).GetFile(ctx, filePath+".yaml")
	if err != nil {
		return err
	}

	err = m.db.FileStore().ForFile(file).Delete(ctx, true)
	if err != nil {
		return err
	}

	return nil
}

// processTSFile handles the core logic of compiling the TypeScript file,
// extracting flow information, and creating the associated service file.
func (m *Manager) processTSFile(ctx context.Context, namespace string, file *filestore.File) error {
	log := slog.With("namespace", namespace, "file", file.Path)

	compiler, err := tsservice.NewTSServiceCompiler(namespace, file.Path, string(file.Data))
	if err != nil {
		return fmt.Errorf("creating tsfile compiler: %w", err)
	}

	flowInfo, err := compiler.Parse()
	if err != nil {
		return fmt.Errorf("extracting flow information: %w", err)
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

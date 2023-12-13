package mirror

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
)

type Applyer interface {
	apply(ctx context.Context, callbacks Callbacks, proc *Process, parser *Parser, notes map[string]string) error
}

type DryrunApplyer struct{}

func (o *DryrunApplyer) apply(_ context.Context, _ Callbacks, _ *Process, _ *Parser, _ map[string]string) error {
	return nil
}

type DirektivApplyer struct {
	NamespaceID uuid.UUID
	log         FormatLogger
	callbacks   Callbacks
	proc        *Process
	parser      *Parser

	rootID uuid.UUID
	notes  map[string]string
}

func (o *DirektivApplyer) apply(ctx context.Context, callbacks Callbacks, proc *Process, parser *Parser, notes map[string]string) error {
	o.log = newPIDFormatLogger(callbacks.ProcessLogger(), proc.ID)
	o.callbacks = callbacks
	o.proc = proc
	o.parser = parser
	o.notes = notes

	o.rootID = uuid.New()

	root, err := callbacks.FileStore().CreateTempRoot(ctx, o.rootID)
	if err != nil {
		return fmt.Errorf("failed to create new filesystem root: %w", err)
	}

	err = o.copyFilesIntoRoot(ctx)
	if err != nil {
		return fmt.Errorf("failed to copy files into new filesystem root: %w", err)
	}

	err = o.copyServicesIntoRoot(ctx)
	if err != nil {
		return fmt.Errorf("failed to copy services into new filesystem root: %w", err)
	}

	err = o.copyWorkflowsIntoRoot(ctx)
	if err != nil {
		return fmt.Errorf("failed to copy workflows into new filesystem root: %w", err)
	}

	err = o.copyEndpointsIntoRoot(ctx)
	if err != nil {
		return fmt.Errorf("failed to copy endpoints into new filesystem root: %w", err)
	}

	err = o.copyConsumersIntoRoot(ctx)
	if err != nil {
		return fmt.Errorf("failed to copy consumers into new filesystem root: %w", err)
	}

	err = o.copyDeprecatedVariables(ctx)
	if err != nil {
		return fmt.Errorf("failed to copy deprecated variables: %w", err)
	}

	err = o.createAnnotations(ctx)
	if err != nil {
		return fmt.Errorf("failed to create annotations: %w", err)
	}

	// TODO: join the next two operations into a single atomic SQL operation?
	err = callbacks.FileStore().ForNamespace(proc.Namespace).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete old filesystem root: %w", err)
	}

	err = callbacks.FileStore().ForRootID(root.ID).SetNamespace(ctx, proc.Namespace)
	if err != nil {
		return fmt.Errorf("failed to delete old filesystem root: %w", err)
	}

	err = o.configureWorkflows(ctx)
	if err != nil {
		return fmt.Errorf("failed to configure workflows: %w", err)
	}

	err = o.copyEventFilters(ctx)
	if err != nil {
		return fmt.Errorf("failed to copy event filters: %w", err)
	}

	err = o.updateConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to configure update config: %w", err)
	}

	return nil
}

func (o *DirektivApplyer) copyFilesIntoRoot(ctx context.Context) error {
	paths, err := o.parser.ListFiles()
	if err != nil {
		return err
	}

	for _, path := range paths {
		actual := filepath.Join(o.parser.tempDir, path)

		fi, err := os.Stat(actual)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			_, _, err = o.callbacks.FileStore().ForRootID(o.rootID).CreateFile(ctx, path, filestore.FileTypeDirectory, "", nil)
			if err != nil {
				return err
			}

			o.log.Debugf("Created directory in database: %s", path)

			continue
		}

		data, err := os.ReadFile(actual)
		if err != nil {
			return err
		}

		mt := mimetype.Detect(data)

		_, _, err = o.callbacks.FileStore().ForRootID(o.rootID).CreateFile(ctx, path, filestore.FileTypeFile, strings.Split(mt.String(), ";")[0], data)
		if err != nil {
			return err
		}

		o.log.Debugf("Created file in database: %s", path)
	}

	return nil
}

func (o *DirektivApplyer) copyWorkflowsIntoRoot(ctx context.Context) error {
	paths := []string{}
	for k := range o.parser.Workflows {
		paths = append(paths, k)
	}

	sort.Strings(paths)

	for _, path := range paths {
		data := o.parser.Workflows[path]
		_, _, err := o.callbacks.FileStore().ForRootID(o.rootID).CreateFile(ctx, path, filestore.FileTypeWorkflow, "application/direktiv", data)
		if err != nil {
			return err
		}

		o.log.Debugf("Created workflow in database: %s", path)
	}

	return nil
}

func (o *DirektivApplyer) copyServicesIntoRoot(ctx context.Context) error {
	paths := []string{}
	for k := range o.parser.Services {
		paths = append(paths, k)
	}

	sort.Strings(paths)

	for _, path := range paths {
		data := o.parser.Services[path]
		_, _, err := o.callbacks.FileStore().ForRootID(o.rootID).CreateFile(ctx, path, filestore.FileTypeService, "application/direktiv", data)
		if err != nil {
			return err
		}

		o.log.Debugf("Created service in database: %s", path)
	}

	return nil
}

func (o *DirektivApplyer) copyEndpointsIntoRoot(ctx context.Context) error {
	paths := []string{}
	for k := range o.parser.Endpoints {
		paths = append(paths, k)
	}

	sort.Strings(paths)

	for _, path := range paths {
		data := o.parser.Endpoints[path]
		_, _, err := o.callbacks.FileStore().ForRootID(o.rootID).CreateFile(ctx, path, filestore.FileTypeEndpoint, "application/direktiv", data)
		if err != nil {
			return err
		}

		o.log.Debugf("Created endpoint in database: %s", path)
	}

	return nil
}

func (o *DirektivApplyer) copyConsumersIntoRoot(ctx context.Context) error {
	paths := []string{}
	for k := range o.parser.Consumers {
		paths = append(paths, k)
	}

	sort.Strings(paths)

	for _, path := range paths {
		data := o.parser.Consumers[path]
		_, _, err := o.callbacks.FileStore().ForRootID(o.rootID).CreateFile(ctx, path, filestore.FileTypeConsumer, "application/direktiv", data)
		if err != nil {
			return err
		}

		o.log.Debugf("Created consumer in database: %s", path)
	}

	return nil
}

func (o *DirektivApplyer) configureWorkflows(ctx context.Context) error {
	paths := []string{}
	for k := range o.parser.Workflows {
		paths = append(paths, k)
	}

	sort.Strings(paths)

	for _, path := range paths {
		file, err := o.callbacks.FileStore().ForRootID(o.rootID).GetFile(ctx, path)
		if err != nil {
			return err
		}

		err = o.callbacks.ConfigureWorkflowFunc(ctx, o.NamespaceID, o.proc.Namespace, file)
		if err != nil {
			return err
		}

		o.log.Debugf("Configured workflow in database: %s", path)
	}

	return nil
}

func (o *DirektivApplyer) copyDeprecatedVariables(ctx context.Context) error {
	for k, v := range o.parser.DeprecatedNamespaceVars {
		mt := mimetype.Detect(v)
		mtString := strings.Split(mt.String(), ";")

		_, err := o.callbacks.VarStore().Set(ctx,
			&core.RuntimeVariable{
				Namespace: o.proc.Namespace,
				Name:      k,
				MimeType:  mtString[0],
				Data:      v,
			})
		if err != nil {
			return fmt.Errorf("failed to save namespace variable '%s': %w", k, err)
		}
	}

	for path, m := range o.parser.DeprecatedWorkflowVars {
		file, err := o.callbacks.FileStore().ForRootID(o.rootID).GetFile(ctx, path)
		if err != nil {
			return err
		}

		for k, v := range m {
			mt := mimetype.Detect(v)
			mtString := strings.Split(mt.String(), ";")

			_, err := o.callbacks.VarStore().Set(ctx,
				&core.RuntimeVariable{
					Namespace:    o.proc.Namespace,
					WorkflowPath: file.Path,
					Name:         k,
					MimeType:     mtString[0],
					Data:         v,
				})
			if err != nil {
				return fmt.Errorf("failed to save workflow variable '%s' '%s': %w", path, k, err)
			}
		}
	}

	return nil
}

func (o *DirektivApplyer) createAnnotations(ctx context.Context) error {
	f, err := o.callbacks.FileStore().ForRootID(o.rootID).GetFile(ctx, "/")
	if err != nil {
		return err
	}

	err = o.callbacks.FileAnnotationsStore().Set(ctx, &core.FileAnnotations{
		FileID: f.ID,
		Data:   o.notes,
	})
	if err != nil {
		return err
	}

	return nil
}

func (o *DirektivApplyer) updateConfig(ctx context.Context) error {
	cfg, err := o.callbacks.Store().GetConfig(ctx, o.proc.Namespace)
	if err != nil {
		return err
	}

	cfg.UpdatedAt = time.Now().UTC()

	if v, ok := o.notes["commit_hash"]; ok {
		cfg.GitCommitHash = v

		return nil
	}

	_, err = o.callbacks.Store().UpdateConfig(ctx, cfg)
	if err != nil {
		return err
	}

	return nil
}

func (o *DirektivApplyer) copyEventFilters(ctx context.Context) error {
	filters, err := o.callbacks.EventFilterStore().GetAll(ctx, o.NamespaceID)
	if err != nil {
		return err
	}

	for _, filter := range filters {
		err = o.callbacks.EventFilterStore().Delete(ctx, o.NamespaceID, filter.Name)
		if err != nil {
			return err
		}
	}

	for name, script := range o.parser.Filters {
		err = o.callbacks.EventFilterStore().Create(ctx, o.NamespaceID, name, string(script))
		if err != nil {
			return err
		}
	}

	return nil
}

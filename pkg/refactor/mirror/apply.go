package mirror

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
)

type Applyer interface {
	apply(context.Context, Callbacks, *Process, *Parser) error
}

type DryrunApplyer struct {
}

func (o *DryrunApplyer) apply(ctx context.Context, _ Callbacks, _ *Process, parser *Parser) error {
	return nil
}

type DirektivApplyer struct {
	log       FormatLogger
	callbacks Callbacks
	proc      *Process
	parser    *Parser

	rootID uuid.UUID
}

func (o *DirektivApplyer) apply(ctx context.Context, callbacks Callbacks, proc *Process, parser *Parser) error {
	o.log = newPIDFormatLogger(callbacks.ProcessLogger(), proc.ID)
	o.callbacks = callbacks
	o.proc = proc
	o.parser = parser

	oldRoot, err := callbacks.FileStore().GetRoot(ctx, proc.RootID)
	if err != nil {
		return fmt.Errorf("failed to get old filesystem root: %w", err)
	}

	o.rootID = uuid.New()

	root, err := callbacks.FileStore().CreateRoot(ctx, o.rootID, proc.NamespaceID, fmt.Sprintf("%s-sync", oldRoot.Name))
	if err != nil {
		return fmt.Errorf("failed to create new filesystem root: %w", err)
	}

	err = o.copyFilesIntoRoot(ctx)
	if err != nil {
		return fmt.Errorf("failed to copy files into new filesystem root: %w", err)
	}

	err = o.copyWorkflowsIntoRoot(ctx)
	if err != nil {
		return fmt.Errorf("failed to copy workflows into new filesystem root: %w", err)
	}

	err = o.copyDeprecatedVariables(ctx)
	if err != nil {
		return fmt.Errorf("failed to copy deprecated variables: %w", err)
	}

	// TODO: copy commit hash metadata
	// TODO: copy filters
	// TODO: copy services

	// TODO: join the next two operations into a single atomic SQL operation?
	err = callbacks.FileStore().ForRootID(oldRoot.ID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete old filesystem root: %w", err)
	}

	err = callbacks.FileStore().ForRootID(root.ID).Rename(ctx, oldRoot.Name)
	if err != nil {
		return fmt.Errorf("failed to delete old filesystem root: %w", err)
	}

	err = o.configureWorkflows(ctx)
	if err != nil {
		return fmt.Errorf("failed to configure workflows: %w", err)
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

		rdr := bytes.NewReader(data)

		_, _, err = o.callbacks.FileStore().ForRootID(o.rootID).CreateFile(ctx, path, filestore.FileTypeFile, mt.String(), rdr)
		if err != nil {
			return err
		}

		o.log.Debugf("Created file in database: %s", path)
	}

	return nil
}

func (o *DirektivApplyer) copyWorkflowsIntoRoot(ctx context.Context) error {
	var paths []string
	for k := range o.parser.Workflows {
		paths = append(paths, k)
	}

	sort.Strings(paths)

	for _, path := range paths {
		data := o.parser.Workflows[path]

		rdr := bytes.NewReader(data)

		_, _, err := o.callbacks.FileStore().ForRootID(o.rootID).CreateFile(ctx, path, filestore.FileTypeWorkflow, "application/direktiv", rdr)
		if err != nil {
			return err
		}

		o.log.Debugf("Created workflow in database: %s", path)
	}

	return nil
}

func (o *DirektivApplyer) configureWorkflows(ctx context.Context) error {
	var paths []string
	for k := range o.parser.Workflows {
		paths = append(paths, k)
	}

	sort.Strings(paths)

	for _, path := range paths {
		file, err := o.callbacks.FileStore().ForRootID(o.rootID).GetFile(ctx, path)
		if err != nil {
			return err
		}

		err = o.callbacks.ConfigureWorkflowFunc(ctx, o.proc.NamespaceID, file)
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

		_, err := o.callbacks.VarStore().Set(ctx,
			&core.RuntimeVariable{
				NamespaceID: o.proc.NamespaceID,
				Name:        k,
				MimeType:    mt.String(),
				Data:        v,
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

			_, err := o.callbacks.VarStore().Set(ctx,
				&core.RuntimeVariable{
					NamespaceID:  o.proc.NamespaceID,
					WorkflowPath: file.Path,
					Name:         k,
					MimeType:     mt.String(),
					Data:         v,
				})
			if err != nil {
				return fmt.Errorf("failed to save workflow variable '%s' '%s': %w", path, k, err)
			}
		}
	}

	return nil
}

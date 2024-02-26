package mirror

import (
	"context"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
)

type Callbacks interface {
	// ConfigureWorkflowFunc is a hookup function the gets called for every new or updated workflow file.
	ConfigureWorkflowFunc(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error

	Store() datastore.MirrorStore
	FileStore() filestore.FileStore
	VarStore() datastore.RuntimeVariablesStore
}

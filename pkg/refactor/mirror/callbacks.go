package mirror

import (
	"context"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
)

type Callbacks interface {
	// ConfigureWorkflowFunc is a hookup function the gets called for every new or updated workflow file.
	// TODO: alan, can we remove/replace this in favor for pubsub?
	ConfigureWorkflowFunc(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error

	Store() Store
	FileStore() filestore.FileStore
	VarStore() core.RuntimeVariablesStore
}

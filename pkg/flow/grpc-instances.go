package flow

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	enginerefactor "github.com/direktiv/direktiv/pkg/engine"
	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/google/uuid"
)

func (srv *server) getInstance(ctx context.Context, namespace, instanceID string) (*enginerefactor.Instance, error) {
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, err
	}

	tx, err := srv.flow.beginSQLTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, namespace)
	if err != nil {
		return nil, err
	}

	idata, err := tx.InstanceStore().ForInstanceID(id).GetSummary(ctx)
	if err != nil {
		return nil, err
	}

	if ns.ID != idata.NamespaceID {
		return nil, os.ErrNotExist
	}

	instance, err := enginerefactor.ParseInstanceData(idata)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (engine *engine) StartWorkflow(ctx context.Context, namespace, path string, input []byte) (*instancestore.InstanceData, error) {
	var err error
	var ns *datastore.Namespace

	err = engine.runSQLTx(ctx, func(tx *database.SQLStore) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, namespace)
		return err
	})
	if err != nil {
		return nil, err
	}

	// TODO tracing
	// TODO logging
	ctx, end, err := tracing.NewSpan(ctx, "starting workflow")
	if err != nil {
		slog.Warn("failed tracing.NewSpan()", "error", fmt.Errorf("StartWorkflow %w", err))
	}
	defer end()
	calledAs := path
	traceParent, err := tracing.ExtractTraceParent(ctx)
	if err != nil {
		slog.Warn("failed tracing.ExtractTraceParent", "error", fmt.Errorf("StartWorkflow %w", err))
	}

	if input == nil {
		input = make([]byte, 0)
	}

	args := &newInstanceArgs{
		ID:        uuid.New(),
		Namespace: ns,
		CalledAs:  calledAs,
		Input:     input,
		Invoker:   apiCaller,
		TelemetryInfo: &enginerefactor.InstanceTelemetryInfo{
			TraceParent:   traceParent,
			NamespaceName: ns.Name,
		},
	}

	im, err := engine.NewInstance(ctx, args)
	if err != nil {
		return nil, err
	}

	go engine.start(im) //nolint:contextcheck

	return im.instance.Instance, nil
}

func (engine *engine) CancelInstance(ctx context.Context, namespace, instanceID string) error {
	instance, err := engine.getInstance(ctx, namespace, instanceID)
	if err != nil {
		return err
	}

	engine.cancelInstance(instance.Instance.ID.String(), "direktiv.cancels.api", "cancelled by api request", false) //nolint:contextcheck

	return nil
}

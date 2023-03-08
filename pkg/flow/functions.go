package flow

import (
	"context"
	"errors"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/functions"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/lib/pq"
)

func (flow *flow) functionsHeartbeat() {
	ctx := context.Background()

	clients := flow.edb.Clients(ctx)

	nss, err := clients.Namespace.Query().All(ctx)
	if err != nil {
		flow.sugar.Error(err)
		return
	}

	for _, ns := range nss {

		wfs, err := ns.QueryWorkflows().All(ctx)
		if err != nil {
			flow.sugar.Error(err)
			continue
		}

		for _, wf := range wfs {

			tuples := make([]*functions.HeartbeatTuple, 0)
			checksums := make(map[string]bool)

			cached := new(database.CacheData)
			err = flow.database.Workflow(ctx, cached, wf.ID)
			if err != nil {
				flow.sugar.Error(err)
				continue
			}

			revs, err := wf.QueryRevisions().WithWorkflow().All(ctx)
			if err != nil {
				flow.sugar.Error(err)
				continue
			}

			for _, rev := range revs {

				x := &database.Revision{
					ID:        rev.ID,
					CreatedAt: rev.CreatedAt,
					Hash:      rev.Hash,
					Source:    rev.Source,
					Metadata:  rev.Metadata,
					Workflow:  rev.Edges.Workflow.ID,
				}

				w, err := loadSource(x)
				if err != nil {
					continue
				}

				fns := w.GetFunctions()

				for i := range fns {

					fn := fns[i]

					if fn.GetType() != model.ReusableContainerFunctionType {
						continue
					}

					def, ok := fn.(*model.ReusableFunctionDefinition)
					if !ok {
						continue
					}

					tuple := &functions.HeartbeatTuple{
						NamespaceName:      ns.Name,
						NamespaceID:        ns.ID.String(),
						WorkflowPath:       cached.Path(),
						WorkflowID:         cached.Workflow.ID.String(),
						Revision:           rev.Hash,
						FunctionDefinition: def,
					}

					csum := bytedata.Checksum(tuple)

					if _, exists := checksums[csum]; !exists {
						checksums[csum] = true
						tuples = append(tuples, tuple)
					}

				}

			}

			flow.flushHeartbeatTuples(tuples)

		}

	}
}

const heartbeatMessageLimit = 4096 // some evidence that we could get away with a limit of 8000, so I've set it here to be safe

func (flow *flow) flushHeartbeatTuples(tuples []*functions.HeartbeatTuple) {
	l := len(tuples)

	if l == 0 {
		return
	}

	msg := bytedata.Marshal(tuples)

	if len(msg) > heartbeatMessageLimit {

		if l == 1 {
			flow.sugar.Errorf("Single heartbeat entry exceeds maximum heartbeat size.")
			return
		}

		x := l / 2

		flow.flushHeartbeatTuples(tuples[:x])
		flow.flushHeartbeatTuples(tuples[x:])
		return

	}

	ctx := context.Background()

	conn, err := flow.edb.DB().Conn(ctx)
	if err != nil {
		flow.sugar.Error(err)
		return
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "SELECT pg_notify($1, $2)", functions.FunctionsChannel, msg)
	perr := new(pq.Error)
	if errors.As(err, &perr) {

		flow.sugar.Errorf("db notification failed: %v", perr)
		return

	}
}

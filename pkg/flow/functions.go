package flow

import (
	"context"

	"github.com/direktiv/direktiv/pkg/functions"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/lib/pq"
)

func (flow *flow) functionsHeartbeat() {

	ctx := context.Background()

	nss, err := flow.db.Namespace.Query().All(ctx)
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

			var tuples = make([]*functions.HeartbeatTuple, 0)
			checksums := make(map[string]bool)

			d, err := flow.reverseTraverseToWorkflow(ctx, wf.ID.String())
			if err != nil {
				flow.sugar.Error(err)
				continue
			}

			revs, err := wf.QueryRevisions().All(ctx)
			if err != nil {
				flow.sugar.Error(err)
				continue
			}

			for _, rev := range revs {

				w, err := loadSource(rev)
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
						WorkflowPath:       d.path,
						WorkflowID:         wf.ID.String(),
						Revision:           rev.Hash,
						FunctionDefinition: def,
					}

					csum := checksum(tuple)

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

	msg := marshal(tuples)

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

	conn, err := flow.db.DB().Conn(ctx)
	if err != nil {
		flow.sugar.Error(err)
		return
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "SELECT pg_notify($1, $2)", functions.FunctionsChannel, msg)
	if err, ok := err.(*pq.Error); ok {

		flow.sugar.Errorf("db notification failed: %v", err)
		return

	}

}

package flow

import (
	"context"

	"github.com/vorteil/direktiv/pkg/functions"
	"github.com/vorteil/direktiv/pkg/model"
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

func (flow *flow) flushHeartbeatTuples(tuples []*functions.HeartbeatTuple) {

	if len(tuples) == 0 {
		return
	}

	s := marshal(tuples)

	conn := flow.server.redis.Get()
	// TODO: do we need to flush or close this conn?

	_, err := conn.Do("PUBLISH", functions.FunctionsChannel, s)
	if err != nil {
		flow.sugar.Error(err)
	}

}

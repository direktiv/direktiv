package flow

import (
	"context"

	"github.com/vorteil/direktiv/pkg/model"
)

type heartbeatTuple struct {
	NamespaceName      string
	NamespaceID        string
	WorkflowPath       string
	WorkflowID         string
	FunctionDefinition *model.ReusableFunctionDefinition
}

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

			var tuples = make([]*heartbeatTuple, 0)
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

					tuple := &heartbeatTuple{
						NamespaceName:      ns.Name,
						NamespaceID:        ns.ID.String(),
						WorkflowPath:       d.path,
						WorkflowID:         wf.ID.String(),
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

func (flow *flow) flushHeartbeatTuples(tuples []*heartbeatTuple) {

}

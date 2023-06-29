package flow

import (
	"context"
	"errors"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/functions"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/lib/pq"
)

func (flow *flow) functionsHeartbeat() {
	ctx := context.Background()

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		flow.sugar.Error(err)
		return
	}
	defer tx.Rollback()

	nss, err := tx.DataStore().Namespaces().GetAll(ctx)
	if err != nil {
		flow.sugar.Error(err)
		return
	}

	for _, ns := range nss {
		files, err := tx.FileStore().ForRootID(ns.ID).ListAllFiles(ctx)
		if err != nil {
			flow.sugar.Error(err)
			return
		}

		for _, file := range files {
			if file.Typ != filestore.FileTypeWorkflow {
				continue
			}

			tuples := make([]*functions.HeartbeatTuple, 0)
			checksums := make(map[string]bool)

			revs, err := tx.FileStore().ForFile(file).GetAllRevisions(ctx)
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
						WorkflowPath:       file.Path,
						WorkflowID:         file.ID.String(),
						Revision:           rev.Checksum,
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
	conn, err := flow.rawDB.Conn(ctx)
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

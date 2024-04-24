package flow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
)

func (flow *flow) NamespaceLint(ctx context.Context, req *grpc.NamespaceLintRequest) (*grpc.NamespaceLintResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	tx, err := flow.beginSQLTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetName())
	if err != nil {
		return nil, err
	}

	secretIssues, err := flow.lintSecrets(ctx, tx, ns)
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceLintResponse

	resp.Namespace = bytedata.ConvertNamespaceToGrpc(ns)
	resp.Issues = make([]*grpc.LinterIssue, 0)
	resp.Issues = append(resp.Issues, secretIssues...)

	return &resp, nil
}

func (flow *flow) lintSecrets(ctx context.Context, tx *database.SQLStore, ns *datastore.Namespace) ([]*grpc.LinterIssue, error) {
	secrets, err := tx.DataStore().Secrets().GetAll(ctx, ns.Name)
	if err != nil {
		return nil, err
	}

	issues := make([]*grpc.LinterIssue, 0)

	for _, secret := range secrets {
		if secret.Data == nil {
			issues = append(issues, &grpc.LinterIssue{
				Level: "critical",
				Type:  "secret",
				Id:    secret.Name,
				Issue: fmt.Sprintf(`secret '%s' has not been initialized`, secret.Name),
			})
		}
	}

	return issues, nil
}

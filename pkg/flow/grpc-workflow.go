// TODO: yassir, need refactor.
package flow

import (
	"bytes"
	"context"
	"fmt"
	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"io"
	"time"
)

/*
import (

	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	entref "github.com/direktiv/direktiv/pkg/flow/ent/ref"
	entrev "github.com/direktiv/direktiv/pkg/flow/ent/revision"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

)

	func (srv *server) traverseToWorkflow(ctx context.Context, namespace, path string) (*database.CacheData, error) {
		cached, err := srv.traverseToInode(ctx, namespace, path)
		if err != nil {
			return nil, err
		}

		err = srv.database.Workflow(ctx, cached, cached.Inode().Workflow)
		if err != nil {
			return nil, err
		}

		return cached, nil
	}

	func (srv *server) reverseTraverseToWorkflow(ctx context.Context, workflow string) (*database.CacheData, error) {
		id, err := uuid.Parse(workflow)
		if err != nil {
			return nil, err
		}

		cached := new(database.CacheData)

		err = srv.database.Workflow(ctx, cached, id)
		if err != nil {
			return nil, err
		}

		return cached, nil
	}

	func (srv *server) traverseToRef(ctx context.Context, namespace, path, reference string) (*database.CacheData, error) {
		if reference == "" {
			reference = latest
		}

		cached, err := srv.traverseToWorkflow(ctx, namespace, path)
		if err != nil {
			srv.sugar.Debugf("%s failed to resolve workflow: %v", parent(), err)
			return nil, err
		}

		var ref *database.Ref

		for i := range cached.Workflow.Refs {
			x := cached.Workflow.Refs[i]
			if x.Name == reference {
				ref = x
				break
			}
		}

		cached.Ref = ref

		err = srv.database.Revision(ctx, cached, ref.Revision)
		if err != nil {
			return nil, err
		}

		return cached, nil
	}

	func (flow *flow) ResolveWorkflowUID(ctx context.Context, req *grpc.ResolveWorkflowUIDRequest) (*grpc.WorkflowResponse, error) {
		flow.sugar.Debugf("Handling gRPC request: %s", this())

		id, err := uuid.Parse(req.GetId())
		if err != nil {
			return nil, err
		}

		cached := new(database.CacheData)
		err = flow.database.Workflow(ctx, cached, id)
		if err != nil {
			return nil, err
		}

		var resp grpc.WorkflowResponse

		err = bytedata.ConvertDataForOutput(cached.Inode(), &resp.Node)
		if err != nil {
			return nil, err
		}

		if resp.Node.ExpandedType == "" {
			resp.Node.ExpandedType = resp.Node.Type
		}

		resp.Namespace = cached.Namespace.Name
		resp.Node.Parent = cached.Dir()
		resp.Node.Path = cached.Path()
		resp.Oid = cached.Workflow.ID.String()

		return &resp, nil
	}
*/
func (flow *flow) Workflow(ctx context.Context, req *grpc.WorkflowRequest) (*grpc.WorkflowResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	file, err := flow.fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	revision, err := flow.fStore.ForFile(file).GetCurrentRevision(ctx)
	if err != nil {
		return nil, err
	}
	dataReader, err := flow.fStore.ForRevision(revision).GetData(ctx)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}

	resp := bytedata.NewWorkflowResponse(file, revision)
	resp.Namespace = ns.Name
	resp.Revision.Source = data
	resp.EventLogging = ""
	resp.Oid = file.ID.String()

	return resp, nil
}

func (flow *flow) WorkflowStream(req *grpc.WorkflowRequest, srv grpc.Flow_WorkflowStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return err
	}

	file, err := flow.fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return err
	}
	revision, err := flow.fStore.ForFile(file).GetCurrentRevision(ctx)
	if err != nil {
		return err
	}
	dataReader, err := flow.fStore.ForRevision(revision).GetData(ctx)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(dataReader)
	if err != nil {
		return err
	}

	resp := bytedata.NewWorkflowResponse(file, revision)
	resp.Namespace = ns.Name
	resp.Revision.Source = data
	resp.EventLogging = ""
	resp.Oid = file.ID.String()

	err = srv.Send(resp)
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 10)

	return nil
}

/*
type createWorkflowArgs struct {
	ns         *database.Namespace
	pino       *database.Inode
	path       string
	super      bool
	data       []byte
	noValidate bool
}

func (flow *flow) createWorkflow(ctx context.Context, args *createWorkflowArgs) (*database.Workflow, *database.Inode, error) {
	if !args.super && args.pino.ReadOnly {
		return nil, nil, errors.New("cannot write into read-only directory")
	}

	hash, err := bytedata.ComputeHash(args.data)
	if err != nil {
		return nil, nil, err
	}

	dir, base := filepath.Split(args.path)

	pcached := new(database.CacheData)

	err = flow.database.Inode(ctx, pcached, args.pino.ID)
	if err != nil {
		return nil, nil, err
	}

	cached, err := flow.database.CreateCompleteWorkflow(ctx, &database.CreateCompleteWorkflowArgs{
		Name:     base,
		ReadOnly: args.pino.ReadOnly,
		Parent:   pcached,
		Hash:     hash,
		Source:   args.data,
		Metadata: make(map[string]interface{}),
	})
	if err != nil {
		return nil, nil, err
	}

	flags := rcfNoPriors
	if args.noValidate {
		flags |= rcfNoValidate
	}

	err = flow.configureRouter(ctx, cached, flags,
		func() error {
			return nil
		},
		func() error {
			return nil
		},
		// tx.Commit,
	)
	if err != nil {
		return nil, nil, err
	}

	metricsWf.WithLabelValues(cached.Namespace.Name, cached.Namespace.Name).Inc()
	metricsWfUpdated.WithLabelValues(cached.Namespace.Name, args.path, cached.Namespace.Name).Inc()

	flow.logToNamespace(ctx, time.Now(), cached, "Created workflow '%s'.", args.path)
	flow.pubsub.NotifyInode(cached.Inode())

	err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeCreate,
		broadcastWorkflowInput{
			Name:   base,
			Path:   args.path,
			Parent: dir,
			Live:   true,
		}, cached)

	if err != nil {
		return nil, nil, err
	}

	return cached.Workflow, cached.Inode(), nil
}
*/

func (flow *flow) CreateWorkflow(ctx context.Context, req *grpc.CreateWorkflowRequest) (*grpc.CreateWorkflowResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	file, revision, err := flow.fStore.ForRootID(ns.ID).CreateFile(ctx, req.GetPath(), filestore.FileTypeWorkflow, bytes.NewReader(req.GetSource()))
	if err != nil {
		fmt.Printf(">>>>> >%s<", req.GetSource())
		return nil, fmt.Errorf("ddddddd: %s", err)
	}
	dataReader, err := flow.fStore.ForRevision(revision).GetData(ctx)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}
	resp := bytedata.NewCreateWorkflowResponse(file, revision)
	resp.Namespace = ns.Name
	resp.Revision.Source = data

	//defer rollback(tx)

	// CREATE HERE

	//metricsWf.WithLabelValues(cached.Namespace.Name, cached.Namespace.Name).Inc()
	//metricsWfUpdated.WithLabelValues(cached.Namespace.Name, path, cached.Namespace.Name).Inc()

	//flow.logToNamespace(ctx, time.Now(), cached, "Created workflow '%s'.", path)
	//flow.pubsub.NotifyInode(cached.Inode())

	//err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeCreate,
	//	broadcastWorkflowInput{
	//		Name:   resp.Node.Name,
	//		Path:   resp.Node.Path,
	//		Parent: resp.Node.Parent,
	//		Live:   true,
	//	}, cached)
	//
	//if err != nil {
	//	return nil, err
	//}

	return resp, nil
}

/*
type updateWorkflowArgs struct {
	cached     *database.CacheData
	path       string
	super      bool
	data       []byte
	noValidate bool
}


func (flow *flow) updateWorkflow(ctx context.Context, args *updateWorkflowArgs) (*database.Revision, error) {
	data := args.data

	hash, err := bytedata.ComputeHash(data)
	if err != nil {
		return nil, err
	}

	if !args.super && args.cached.Inode().ReadOnly {
		return nil, errors.New("cannot write into read-only directory")
	}

	var ref *database.Ref

	for i := range args.cached.Workflow.Refs {
		x := args.cached.Workflow.Refs[i]
		if x.Name == latest {
			ref = x
			break
		}
	}

	args.cached.Ref = ref

	err = flow.database.Revision(ctx, args.cached, ref.Revision)
	if err != nil {
		return nil, err
	}

	oldrev := args.cached.Revision

	var k int
	var rev *database.Revision

	if oldrev.Hash == hash {
		// gracefully abort if hash matches latest
		return oldrev, nil
	}

	// flags := rcfNoPriors
	flags := rcfBreaking
	if args.noValidate {
		flags |= rcfNoValidate
	}

	err = flow.configureRouter(ctx, args.cached, flags,
		func() error {
			rev, err = flow.database.CreateRevision(ctx, &database.CreateRevisionArgs{
				Workflow: args.cached.Workflow.ID,
				Hash:     hash,
				Source:   data,
				Metadata: make(map[string]interface{}),
			})

			clients := flow.edb.Clients(ctx)

			// change latest tag
			err = clients.Ref.UpdateOneID(ref.ID).SetRevisionID(rev.ID).Exec(ctx)
			if err != nil {
				return err
			}

			k, err = clients.Ref.Query().Where(entref.HasRevisionWith(entrev.ID(oldrev.ID))).Count(ctx)
			if err != nil {
				return err
			}

			// ??? if hash matches non-latest

			if k == 0 {
				// delete previous latest if untagged
				err = clients.Revision.DeleteOneID(oldrev.ID).Exec(ctx)
				if err != nil {
					return err
				}
			}

			return nil
		},
		func() error { return nil },
	)
	if err != nil {
		return nil, err
	}

	metricsWfUpdated.WithLabelValues(args.cached.Namespace.Name, args.cached.Path(), args.cached.Namespace.Name).Inc()

	flow.logToWorkflow(ctx, time.Now(), args.cached, "Updated workflow.")
	flow.pubsub.NotifyWorkflow(args.cached.Workflow)

	err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeUpdate,
		broadcastWorkflowInput{
			Name:   args.cached.Inode().Name,
			Path:   args.cached.Path(),
			Parent: args.cached.Dir(),
			Live:   args.cached.Workflow.Live,
		}, args.cached)

	if err != nil {
		return nil, err
	}

	return rev, nil
}

func (flow *flow) UpdateWorkflow(ctx context.Context, req *grpc.UpdateWorkflowRequest) (*grpc.UpdateWorkflowResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToWorkflow(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	rev, err := flow.updateWorkflow(tctx, &updateWorkflowArgs{
		cached: cached,
		path:   cached.Path(),
		data:   req.GetSource(),
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.database.InvalidateWorkflow(ctx, cached, false)

	var resp grpc.UpdateWorkflowResponse

	err = bytedata.ConvertDataForOutput(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = cached.Dir()
	resp.Node.Path = cached.Path()

	err = bytedata.ConvertDataForOutput(rev, &resp.Revision)
	if err != nil {
		return nil, err
	}

	resp.Revision = &grpc.Revision{
		Name: rev.ID.String(),
	}

	return &resp, nil
}

func (flow *flow) SaveHead(ctx context.Context, req *grpc.SaveHeadRequest) (*grpc.SaveHeadResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToRef(tctx, req.GetNamespace(), req.GetPath(), "")
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	k, err := clients.Ref.Query().Where(entref.HasRevisionWith(entrev.ID(cached.Revision.ID))).Count(ctx)
	if err != nil {
		return nil, err
	}

	metadata := req.GetMetadata()

	if k > 1 {
		// already saved, gracefully back out
		rollback(tx)
		goto respond
	}

	if len(metadata) != 0 {
		obj := make(map[string]interface{})

		err := json.Unmarshal(metadata, &obj)
		if err != nil {
			return nil, err
		}

		_, err = clients.Revision.UpdateOneID(cached.Revision.ID).SetMetadata(obj).Save(ctx)
		if err != nil {
			return nil, err
		}
	}

	err = clients.Ref.Create().SetImmutable(true).SetName(cached.Revision.ID.String()).SetRevisionID(cached.Revision.ID).SetWorkflowID(cached.Workflow.ID).Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.database.InvalidateWorkflow(ctx, cached, false)

	flow.logToWorkflow(ctx, time.Now(), cached, "Saved workflow: %s.", cached.Revision.ID.String())
	flow.pubsub.NotifyWorkflow(cached.Workflow)

respond:

	var resp grpc.SaveHeadResponse

	err = bytedata.ConvertDataForOutput(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = cached.Dir()
	resp.Node.Path = cached.Path()

	err = bytedata.ConvertDataForOutput(cached.Revision, &resp.Revision)
	if err != nil {
		return nil, err
	}

	resp.Revision.Name = cached.Revision.ID.String()

	return &resp, nil
}

func (flow *flow) DiscardHead(ctx context.Context, req *grpc.DiscardHeadRequest) (*grpc.DiscardHeadResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToRef(tctx, req.GetNamespace(), req.GetPath(), "")
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	if len(cached.Workflow.Revisions) == 1 || len(cached.Workflow.Refs) > 1 {
		// already saved, or not discardable, gracefully back out
		rollback(tx)
		goto respond
	}

	err = flow.configureRouter(tctx, cached, rcfBreaking,
		func() error {
			err = clients.Ref.UpdateOneID(cached.Ref.ID).SetRevisionID(cached.Workflow.Revisions[1].ID).Exec(tctx)
			if err != nil {
				return err
			}

			err = clients.Revision.DeleteOneID(cached.Revision.ID).Exec(tctx)
			if err != nil {
				return err
			}

			err = flow.database.Revision(tctx, cached, cached.Workflow.Revisions[1].ID)
			if err != nil {
				return err
			}

			return nil
		},
		tx.Commit,
	)
	if err != nil {
		return nil, err
	}

	flow.database.InvalidateWorkflow(ctx, cached, false)

	metricsWfUpdated.WithLabelValues(cached.Namespace.Name, cached.Path(), cached.Namespace.Name).Inc()

	flow.logToWorkflow(ctx, time.Now(), cached, "Discard unsaved changes to workflow.")
	flow.pubsub.NotifyWorkflow(cached.Workflow)

respond:

	var resp grpc.DiscardHeadResponse

	err = bytedata.ConvertDataForOutput(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = cached.Dir()
	resp.Node.Path = cached.Path()

	err = bytedata.ConvertDataForOutput(cached.Revision, &resp.Revision)
	if err != nil {
		return nil, err
	}

	resp.Revision.Name = cached.Revision.ID.String()

	return &resp, nil
}

func (flow *flow) ToggleWorkflow(ctx context.Context, req *grpc.ToggleWorkflowRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToWorkflow(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	if cached.Workflow.Live == req.GetLive() {
		rollback(tx)
		return &resp, nil
	}

	clients := flow.edb.Clients(tctx)

	err = flow.configureRouter(tctx, cached, rcfBreaking,
		func() error {
			wf, err := clients.Workflow.UpdateOneID(cached.Workflow.ID).SetLive(req.GetLive()).Save(tctx)
			if err != nil {
				return err
			}
			cached.Workflow.Live = wf.Live
			return nil
		},
		tx.Commit,
	)
	if err != nil {
		return nil, err
	}

	flow.database.InvalidateWorkflow(ctx, cached, false)

	live := "disabled"
	if cached.Workflow.Live {
		live = "enabled"
	}

	err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeUpdate,
		broadcastWorkflowInput{
			Name:   cached.Inode().Name,
			Path:   cached.Path(),
			Parent: cached.Dir(),
			Live:   cached.Workflow.Live,
		}, cached)

	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), cached, "Workflow is now %s", live)
	flow.pubsub.NotifyWorkflow(cached.Workflow)

	return &resp, nil
}

func (flow *flow) SetWorkflowEventLogging(ctx context.Context, req *grpc.SetWorkflowEventLoggingRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToWorkflow(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	_, err = clients.Workflow.UpdateOneID(cached.Workflow.ID).SetLogToEvents(req.GetLogger()).Save(tctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.database.InvalidateWorkflow(ctx, cached, false)

	flow.logToWorkflow(ctx, time.Now(), cached, "Workflow now logging to cloudevents: %s", req.GetLogger())
	flow.pubsub.NotifyWorkflow(cached.Workflow)
	var resp emptypb.Empty

	return &resp, nil
}
*/

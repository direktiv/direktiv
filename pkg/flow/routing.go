package flow

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getRouter(ctx context.Context, tx *sqlTx, file *filestore.File) (*core.FileAnnotations, *routerData, error) {
	router := &routerData{
		Enabled: true,
		Routes:  make(map[string]int),
	}

	annotations, err := tx.DataStore().FileAnnotations().Get(ctx, file.ID)
	if err != nil {
		if errors.Is(err, core.ErrFileAnnotationsNotSet) {
			t := time.Now()
			annotations := &core.FileAnnotations{
				FileID:    file.ID,
				Data:      make(core.FileAnnotationsData),
				CreatedAt: t,
				UpdatedAt: t,
			}
			return annotations, router, nil
		}
		return nil, nil, err
	} else {
		s := annotations.Data.GetEntry(routerAnnotationKey)
		if s != "" && s != `""` && s != `\"\"` {
			err = json.Unmarshal([]byte(s), &router)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return annotations, router, nil
}

type muxStart struct {
	Enabled  bool                         `json:"enabled"`
	Type     string                       `json:"type"`
	Cron     string                       `json:"cron"`
	Events   []model.StartEventDefinition `json:"events"`
	Lifespan string                       `json:"lifespan"`
}

func newMuxStart(workflow *model.Workflow) *muxStart {
	ms := new(muxStart)

	def := workflow.GetStartDefinition()
	ms.Enabled = true
	ms.Type = def.GetType().String()
	ms.Events = def.GetEvents()

	switch def.GetType() {
	case model.StartTypeDefault:
	case model.StartTypeEvent:
	case model.StartTypeEventsAnd:
		x := def.(*model.EventsAndStart)
		ms.Lifespan = x.LifeSpan
	case model.StartTypeEventsXor:
	case model.StartTypeScheduled:
		x := def.(*model.ScheduledStart)
		ms.Cron = x.Cron
	default:
		panic(fmt.Errorf("unexpected start type: %v", def.GetType()))
	}

	return ms
}

func (ms *muxStart) Hash() string {
	if ms == nil {
		ms = new(muxStart)
		ms.Enabled = true
		ms.Type = model.StartTypeDefault.String()
	}

	return bytedata.Checksum(ms)
}

const routerAnnotationKey = "router"

type routerData struct {
	Enabled bool
	Routes  map[string]int
}

func (r *routerData) Marshal() string {
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (srv *server) validateRouter(ctx context.Context, tx *sqlTx, file *filestore.File) (*muxStart, error, error) {
	_, router, err := getRouter(ctx, tx, file)
	if err != nil {
		return nil, nil, err
	}

	if len(router.Routes) == 0 {
		rev, err := tx.FileStore().ForFile(file).GetCurrentRevision(ctx)
		if err != nil {
			return nil, nil, err
		}

		workflow := new(model.Workflow)

		err = workflow.Load(rev.Data)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error()), nil
		}

		ms := newMuxStart(workflow)
		ms.Enabled = router.Enabled

		return ms, nil, nil
	}

	var startHash string
	var startRef string

	var ms *muxStart

	for ref := range router.Routes {
		rev, err := tx.FileStore().ForFile(file).GetRevision(ctx, ref)
		if err != nil {
			return nil, nil, err
		}

		workflow := new(model.Workflow)

		err = workflow.Load(rev.Data)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("route to '%s' invalid because revision fails to compile: %v", ref, err)), nil
		}

		ms = newMuxStart(workflow)
		ms.Enabled = router.Enabled

		hash := ms.Hash() // checksum(workflow.Start)
		if startHash == "" {
			startHash = hash
			startRef = ref
		} else {
			if startHash != hash {
				return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("incompatible start definitions between refs '%s' and '%s'", startRef, ref)), nil
			}
		}
	}

	return ms, nil, nil
}

func (engine *engine) getAmbiguousFile(ctx context.Context, tx *sqlTx, ns *database.Namespace, path string) (*filestore.File, error) {
	file, err := tx.FileStore().ForRootID(ns.ID).GetFile(ctx, path)
	if err != nil {
		if errors.Is(err, filestore.ErrNotFound) { // try as-is, then '.yaml', then '.yml'
			if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
				var err2 error
				file, err2 = tx.FileStore().ForRootID(ns.ID).GetFile(ctx, path+".yaml")
				if err2 != nil {
					if errors.Is(err2, filestore.ErrNotFound) {
						file, err2 = tx.FileStore().ForRootID(ns.ID).GetFile(ctx, path+".yml")
						if err2 != nil {
							if !errors.Is(err2, filestore.ErrNotFound) {
								err = err2
							}
							return nil, err
						}
					} else {
						err = err2
						return nil, err
					}
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return file, nil
}

func (engine *engine) mux(ctx context.Context, ns *database.Namespace, calledAs string) (*filestore.File, *filestore.Revision, error) {
	// TODO: Alan, fix for the new filestore.(*Revision).GetRevision() api.
	uriElems := strings.SplitN(calledAs, ":", 2)
	path := uriElems[0]
	ref := ""
	if len(uriElems) > 1 {
		ref = uriElems[1]
	}

	tx, err := engine.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	file, err := engine.getAmbiguousFile(ctx, tx, ns, path)
	if err != nil {
		return nil, nil, err
	}

	_, router, err := getRouter(ctx, tx, file)
	if err != nil {
		return nil, nil, err
	}

	if !router.Enabled {
		return nil, nil, errors.New("cannot execute disabled workflow")
	}

	var rev *filestore.Revision

	if ref == filestore.Latest {
		rev, err = tx.FileStore().ForFile(file).GetCurrentRevision(ctx)
		if err != nil {
			return nil, nil, err
		}
	} else if ref != "" {
		rev, err = tx.FileStore().ForFile(file).GetRevision(ctx, ref)
		if err != nil {
			return nil, nil, err
		}
	} else {
		if len(router.Routes) == 0 {
			rev, err = tx.FileStore().ForFile(file).GetCurrentRevision(ctx)
			if err != nil {
				return nil, nil, err
			}
		} else {
			totalWeights := 0
			allRevs := make([]*filestore.Revision, 0)
			allWeights := make([]int, 0)

			for k, v := range router.Routes {
				id, err := uuid.Parse(k)
				if err == nil {
					rev, err = tx.FileStore().ForFile(file).GetRevision(ctx, id.String())
					if err == nil {
						totalWeights += v
						allRevs = append(allRevs, rev)
						allWeights = append(allWeights, v)
					}
				} else if k == filestore.Latest {
					rev, err = tx.FileStore().ForFile(file).GetCurrentRevision(ctx)
					if err == nil {
						totalWeights += v
						allRevs = append(allRevs, rev)
						allWeights = append(allWeights, v)
					}
				} else {
					rev, err = tx.FileStore().ForFile(file).GetRevision(ctx, k)
					if err == nil {
						totalWeights += v
						allRevs = append(allRevs, rev)
						allWeights = append(allWeights, v)
					}
				}
			}

			if totalWeights < 1 {
				rev, err = tx.FileStore().ForFile(file).GetCurrentRevision(ctx)
				if err != nil {
					return nil, nil, err
				}
			} else {
				x, err := rand.Int(rand.Reader, big.NewInt(int64(totalWeights)))
				if err != nil {
					return nil, nil, err
				}
				choice := int(x.Int64())
				var idx int
				for idx = 0; choice > 0; idx++ {
					choice -= allWeights[idx]
					if choice <= 0 {
						break
					}
				}
				rev = allRevs[idx]
			}
		}
	}

	// if cached.Ref == nil {
	// 	// TODO: yassir, how are we going to fake these fields?
	// 	cached.Ref = &database.Ref{
	// 		// ID: ,
	// 		// Immutable: ,
	// 		// Name: ,
	// 		// CreatedAt: ,
	// 		Revision: rev.ID,
	// 	}
	// }

	return file, rev, nil
}

const (
	rcfNone       = 0
	rcfNoPriors   = 1 << iota
	rcfBreaking   // don't throw a router validation error if the router was already invalid before the change
	rcfNoValidate // skip validation of new workflow if old router has one or fewer routes
)

func hasFlag(flags, flag int) bool {
	return flags&flag != 0
}

func (flow *flow) configureRouterHandler(req *pubsub.PubsubUpdate) {
	msg := new(pubsub.ConfigureRouterMessage)

	err := json.Unmarshal([]byte(req.Key), msg)
	if err != nil {
		flow.sugar.Error(err)
		return
	}

	if msg.Cron == "" || !msg.Enabled {
		flow.timers.deleteCronForWorkflow(msg.ID)
	}

	if msg.Cron != "" && msg.Enabled {
		err = flow.timers.addCron(msg.ID, wfCron, msg.Cron, []byte(msg.ID))
		if err != nil {
			flow.sugar.Error(err)
			return
		}
	}
}

func (flow *flow) cronHandler(data []byte) {
	ctx := context.Background()

	id, err := uuid.Parse(string(data))
	if err != nil {
		flow.sugar.Error(err)
		return
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		flow.sugar.Error(err)
		return
	}
	defer tx.Rollback()

	file, err := tx.FileStore().GetFile(ctx, id)
	if err != nil {
		if errors.Is(err, filestore.ErrNotFound) {
			flow.sugar.Infof("Cron failed to find workflow. Deleting cron.")
			flow.timers.deleteCronForWorkflow(id.String())
			return
		}
		flow.sugar.Error(err)
		return
	}

	ns, err := tx.DataStore().Namespaces().GetByID(ctx, file.RootID)
	if err != nil {
		flow.sugar.Error(err)
		return
	}

	tx.Rollback()

	ctx, conn, err := flow.engine.lock(id.String(), defaultLockWait)
	if err != nil {
		flow.sugar.Error(err)
		return
	}
	defer flow.engine.unlock(id.String(), conn)

	err = tx.InstanceStore().AssertNoParallelCron(ctx, file.Path)
	if errors.Is(err, instancestore.ErrParallelCron) {
		// already triggered
		return
	} else if err != nil {
		flow.sugar.Error(err)
		return
	}

	span := trace.SpanFromContext(ctx)

	args := &newInstanceArgs{
		ID:        uuid.New(),
		Namespace: ns,
		CalledAs:  file.Path,
		Input:     nil,
		Invoker:   instancestore.InvokerCron,
		TelemetryInfo: &enginerefactor.InstanceTelemetryInfo{
			TraceID: span.SpanContext().TraceID().String(),
			SpanID:  span.SpanContext().SpanID().String(),
			// TODO: alan, CallPath: ,
			NamespaceName: ns.Name,
		},
	}

	im, err := flow.engine.NewInstance(ctx, args)
	if err != nil {
		flow.sugar.Error("Error returned to gRPC request %s: %v", this(), err)
		return
	}

	flow.engine.queue(im)
}

func (flow *flow) configureWorkflowStarts(ctx context.Context, tx *sqlTx, nsID uuid.UUID, file *filestore.File, router *routerData, strict bool) error {
	ms, verr, err := flow.validateRouter(ctx, tx, file)
	if err != nil {
		return err
	}
	if verr != nil {
		if strict {
			return verr
		}
	}

	if ms == nil {
		ms = &muxStart{
			Enabled: false,
			Type:    model.StartTypeDefault.String(),
		}
	}

	err = flow.events.processWorkflowEvents(ctx, nsID, file, ms)
	if err != nil {
		return err
	}

	flow.pubsub.ConfigureRouterCron(file.ID.String(), ms.Cron, ms.Enabled)

	return nil
}

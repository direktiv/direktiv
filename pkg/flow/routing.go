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
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entirt "github.com/direktiv/direktiv/pkg/flow/ent/instanceruntime"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	data, _ := json.Marshal(r)
	return string(data)
}

func (srv *server) validateRouter(ctx context.Context, fStore filestore.FileStore, store datastore.Store, file *filestore.File) (*muxStart, error, error) {

	router := new(routerData)

	annotations, err := store.FileAnnotations().Get(ctx, file.ID)
	if err != nil {
		if !errors.Is(err, core.ErrFileAnnotationsNotSet) {
			return nil, nil, err
		}
	} else {
		s := annotations.Data.GetEntry(routerAnnotationKey)
		if s != "" {
			err = json.Unmarshal([]byte(s), router)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	if len(router.Routes) == 0 {

		rev, err := fStore.ForFile(file).GetCurrentRevision(ctx)
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

		var rev *filestore.Revision

		uid, err := uuid.Parse(ref)
		if err == nil {
			rev, err = fStore.ForFile(file).GetRevision(ctx, uid)
		}
		if err != nil {
			rev, err = fStore.ForFile(file).GetRevisionByTag(ctx, ref)
		}
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

func (engine *engine) mux(ctx context.Context, namespace, path, ref string) (*database.CacheData, error) {
	ns, err := engine.edb.NamespaceByName(ctx, namespace)
	if err != nil {
		return nil, err
	}
	fStore, store, _, rollback, err := engine.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx)

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, path)
	if err != nil {
		if errors.Is(err, filestore.ErrNotFound) { // try as-is, then '.yaml', then '.yml'
			if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
				var err2 error
				file, err2 = fStore.ForRootID(ns.ID).GetFile(ctx, path+".yaml")
				if err2 != nil {
					if errors.Is(err2, filestore.ErrNotFound) {
						file, err2 = fStore.ForRootID(ns.ID).GetFile(ctx, path+".yml")
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

	var rev *filestore.Revision

	if ref == "latest" {
		rev, err = fStore.ForFile(file).GetCurrentRevision(ctx)
		if err != nil {
			return nil, err
		}
	} else if ref != "" {
		rev, err = fStore.ForFile(file).GetRevisionByTag(ctx, ref)
		if err != nil {
			return nil, err
		}
	} else {
		router := new(routerData)
		annotations, err := store.FileAnnotations().Get(ctx, file.ID)
		if err != nil {
			if !errors.Is(err, core.ErrFileAnnotationsNotSet) {
				return nil, err
			}
		} else {
			s := annotations.Data.GetEntry(routerAnnotationKey)
			err = json.Unmarshal([]byte(s), &router)
			if err != nil {
				return nil, err
			}
		}

		if len(router.Routes) == 0 {
			rev, err = fStore.ForFile(file).GetCurrentRevision(ctx)
			if err != nil {
				return nil, err
			}
		} else {
			totalWeights := 0
			allRevs := make([]*filestore.Revision, 0)
			allWeights := make([]int, 0)

			for k, v := range router.Routes {
				id, err := uuid.Parse(k)
				if err == nil {
					rev, err = fStore.ForFile(file).GetRevision(ctx, id)
					if err == nil {
						totalWeights += v
						allRevs = append(allRevs, rev)
						allWeights = append(allWeights, v)
					}
				} else {
					rev, err = fStore.ForFile(file).GetRevisionByTag(ctx, k)
					if err == nil {
						totalWeights += v
						allRevs = append(allRevs, rev)
						allWeights = append(allWeights, v)
					}
				}
			}

			if totalWeights < 1 {
				rev, err = fStore.ForFile(file).GetCurrentRevision(ctx)
				if err != nil {
					return nil, err
				}
			} else {
				x, err := rand.Int(rand.Reader, big.NewInt(int64(totalWeights)))
				if err != nil {
					return nil, err
				}
				choice := int(x.Int64())
				var idx int
				for idx = 0; choice > 0; idx++ {
					choice -= allWeights[idx]
				}
				rev = allRevs[idx]
			}
		}
	}

	cached := new(database.CacheData)
	cached.Namespace = ns
	cached.File = file
	cached.Revision = rev

	if cached.Ref == nil {
		// TODO: yassir, how are we going to fake these fields?
		cached.Ref = &database.Ref{
			// ID: ,
			// Immutable: ,
			// Name: ,
			// CreatedAt: ,
			Revision: rev.ID,
		}
	}

	return cached, nil
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

	fStore, _, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		flow.sugar.Error(err)
		return
	}
	defer rollback(ctx)

	file, err := fStore.GetFile(ctx, id)
	if err != nil {
		if errors.Is(err, filestore.ErrNotFound) {
			flow.sugar.Infof("Cron failed to find workflow. Deleting cron.")
			flow.timers.deleteCronForWorkflow(id.String())
			return
		}
		flow.sugar.Error(err)
		return
	}
	rollback(ctx)

	ns, err := flow.edb.Namespace(ctx, file.RootID)
	if err != nil {
		flow.sugar.Error(err)
		return
	}

	ctx, conn, err := flow.engine.lock(id.String(), defaultLockWait)
	if err != nil {
		flow.sugar.Error(err)
		return
	}
	defer flow.engine.unlock(id.String(), conn)

	cached := new(database.CacheData)
	cached.Namespace = ns
	cached.File = file

	clients := flow.edb.Clients(ctx)

	k, err := clients.Instance.Query().Where(entinst.WorkflowID(cached.File.ID)).Where(entinst.CreatedAtGT(time.Now().Add(-time.Second*30)), entinst.HasRuntimeWith(entirt.CallerData(util.CallerCron))).Count(ctx)
	if err != nil {
		flow.sugar.Error(err)
		return
	}

	if k != 0 {
		// already triggered
		return
	}

	args := new(newInstanceArgs)
	args.Namespace = cached.Namespace.Name
	args.Path = cached.File.Path
	args.Ref = ""
	args.Input = nil
	args.Caller = util.CallerCron
	args.CallerData = util.CallerCron

	im, err := flow.engine.NewInstance(ctx, args)
	if err != nil {
		flow.sugar.Error("Error returned to gRPC request %s: %v", this(), err)
		return
	}

	flow.engine.queue(im)
}

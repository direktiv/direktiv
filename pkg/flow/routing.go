package flow

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entirt "github.com/direktiv/direktiv/pkg/flow/ent/instanceruntime"
	entwf "github.com/direktiv/direktiv/pkg/flow/ent/workflow"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
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

func (srv *server) validateRouter(ctx context.Context, cached *database.CacheData) (*muxStart, error, error) {
	if len(cached.Workflow.Routes) == 0 {

		// latest
		var ref *database.Ref
		for i := range cached.Workflow.Refs {
			if cached.Workflow.Refs[i].Name == latest {
				ref = cached.Workflow.Refs[i]
				break
			}
		}

		if ref == nil {
			return nil, nil, os.ErrNotExist
		}

		cached.Ref = ref

		err := srv.database.Revision(ctx, cached, cached.Ref.Revision)
		if err != nil {
			return nil, nil, err
		}

		workflow, err := loadSource(cached.Revision)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error()), nil
		}

		ms := newMuxStart(workflow)
		ms.Enabled = cached.Workflow.Live

		return ms, nil, nil

	} else {
		for i := range cached.Workflow.Routes {

			route := cached.Workflow.Routes[i]
			if route.Ref == nil {
				return nil, nil, &derrors.NotFoundError{
					Label: "ref not found",
				}
			}

		}
	}

	var startHash string
	var startRef string

	var ms *muxStart

	for _, route := range cached.Workflow.Routes {

		cached.Ref = route.Ref

		err := srv.database.Revision(ctx, cached, cached.Ref.Revision)
		if err != nil {
			return nil, nil, err
		}

		workflow, err := loadSource(cached.Revision)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("route to '%s' invalid because revision fails to compile: %v", route.Ref.Name, err)), nil
		}

		ms = newMuxStart(workflow)
		ms.Enabled = cached.Workflow.Live

		hash := ms.Hash() // checksum(workflow.Start)
		if startHash == "" {
			startHash = hash
			startRef = route.Ref.Name
		} else {
			if startHash != hash {
				return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("incompatible start definitions between refs '%s' and '%s'", startRef, route.Ref.Name)), nil
			}
		}

	}

	return ms, nil, nil
}

func (engine *engine) mux(ctx context.Context, namespace, path, ref string) (*database.CacheData, error) {
	cached, err := engine.traverseToWorkflow(ctx, namespace, path)
	if err != nil {
		return nil, fmt.Errorf("workflow multiplexer failed to resolve workflow: %w", err)
	}

	if ref == "" {
		// use router to select version

		if len(cached.Workflow.Routes) == 0 {
			ref = latest
		} else {

			weight := 0

			for _, route := range cached.Workflow.Routes {
				weight += route.Weight
			}

			cn, err := rand.Int(rand.Reader, big.NewInt(int64(weight)))
			if err != nil {
				return nil, err
			}

			n := int(cn.Int64())

			n = n % weight

			var route *database.Route

			for idx := range cached.Workflow.Routes {
				route = cached.Workflow.Routes[idx]
				n -= route.Weight
				if n < 0 {
					break
				}
			}

			cached.Ref = route.Ref

		}
	}

	if cached.Ref == nil {
		for idx := range cached.Workflow.Refs {
			x := cached.Workflow.Refs[idx]
			if x.Name == ref {
				cached.Ref = x
				break
			}
		}
	}

	err = engine.database.Revision(ctx, cached, cached.Ref.Revision)
	if err != nil {
		return nil, fmt.Errorf("workflow multiplexer failed to resolve workflow revision matching ref '%s' (UUID: %s): %w", cached.Ref.Name, cached.Ref.Revision, err)
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

func (flow *flow) configureRouter(ctx context.Context, cached *database.CacheData, flags int, changer, commit func() error) error {
	var err error
	var muxErr1 error
	var ms1 *muxStart

	if !hasFlag(flags, rcfNoPriors) {
		// NOTE: we check router valid before deleting because there's no sense failing the
		// operation for resulting in an invalid router if the router was already invalid.
		ms1, muxErr1, err = flow.validateRouter(ctx, cached)
		if err != nil {
			return err
		}
	}

	err = changer()
	if err != nil {
		return err
	}

	ms2, muxErr2, err := flow.validateRouter(ctx, cached)
	if err != nil {
		return err
	}

	if muxErr2 != nil {
		if hasFlag(flags, rcfNoValidate) && len(cached.Workflow.Routes) <= 1 {
			// no need to do anything here?
		} else if muxErr1 == nil || !hasFlag(flags, rcfBreaking) {
			return muxErr2
		}
	}

	if ms2 == nil {
		ms2 = new(muxStart)
	}

	mustReconfigureRouter := ms1.Hash() != ms2.Hash() || hasFlag(flags, rcfNoPriors)

	if mustReconfigureRouter {
		err = flow.preCommitRouterConfiguration(ctx, cached, ms2)
		if err != nil {
			return err
		}
	}

	err = commit()

	if err != nil {
		return err
	}

	if mustReconfigureRouter {
		flow.postCommitRouterConfiguration(cached.Workflow.ID.String(), ms2)
	}

	return nil
}

func (flow *flow) preCommitRouterConfiguration(ctx context.Context, cached *database.CacheData, ms *muxStart) error {
	err := flow.events.processWorkflowEvents(ctx, cached, ms)
	if err != nil {
		return err
	}

	return nil
}

func (flow *flow) postCommitRouterConfiguration(id string, ms *muxStart) {
	flow.pubsub.ConfigureRouterCron(id, ms.Cron, ms.Enabled)
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
	id, err := uuid.Parse(string(data))
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
	err = flow.database.Workflow(ctx, cached, id)
	if err != nil {

		if derrors.IsNotFound(err) {
			flow.sugar.Infof("Cron failed to find workflow. Deleting cron.")
			flow.timers.deleteCronForWorkflow(id.String())
			return
		}

		flow.sugar.Error(err)
		return

	}

	clients := flow.edb.Clients(ctx)

	k, err := clients.Instance.Query().Where(entinst.HasWorkflowWith(entwf.ID(cached.Workflow.ID))).Where(entinst.CreatedAtGT(time.Now().Add(-time.Second*30)), entinst.HasRuntimeWith(entirt.CallerData(util.CallerCron))).Count(ctx)
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
	args.Path = cached.Path()
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

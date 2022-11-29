package flow

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entirt "github.com/direktiv/direktiv/pkg/flow/ent/instanceruntime"
	entref "github.com/direktiv/direktiv/pkg/flow/ent/ref"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/util"
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

	return checksum(ms)

}

func validateRouter(ctx context.Context, wf *ent.Workflow) (*muxStart, error, error) {

	routes, err := wf.QueryRoutes().WithRef(func(q *ent.RefQuery) {
		q.WithRevision()
	}).All(ctx)
	if err != nil {
		return nil, nil, err
	}

	if len(routes) == 0 {

		// latest
		ref, err := wf.QueryRefs().Where(entref.NameEQ(latest)).WithRevision().Only(ctx)
		if err != nil {
			return nil, nil, err
		}

		if ref.Edges.Revision == nil {
			err = &derrors.NotFoundError{
				Label: "revision not found",
			}
			return nil, nil, err
		}

		workflow, err := loadSource(ref.Edges.Revision)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error()), nil
		}

		ms := newMuxStart(workflow)
		ms.Enabled = wf.Live

		return ms, nil, nil

	} else {

		for i := range routes {

			route := routes[i]
			if route.Edges.Ref == nil {
				err = &derrors.NotFoundError{
					Label: "ref not found",
				}
				return nil, nil, err
			}

			if route.Edges.Ref.Edges.Revision == nil {
				err = &derrors.NotFoundError{
					Label: "revision not found",
				}
				return nil, nil, err
			}

		}

	}

	var startHash string
	var startRef string

	var ms *muxStart

	for _, route := range routes {

		workflow, err := loadSource(route.Edges.Ref.Edges.Revision)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("route to '%s' invalid because revision fails to compile: %v", route.Edges.Ref.Name, err)), nil
		}

		ms = newMuxStart(workflow)
		ms.Enabled = wf.Live

		hash := ms.Hash() // checksum(workflow.Start)
		if startHash == "" {
			startHash = hash
			startRef = route.Edges.Ref.Name
		} else {
			if startHash != hash {
				return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("incompatible start definitions between refs '%s' and '%s'", startRef, route.Edges.Ref.Name)), nil
			}
		}

	}

	return ms, nil, nil

}

func (engine *engine) mux(ctx context.Context, nsc *ent.NamespaceClient, namespace, path, ref string) (*refData, error) {

	wd, err := engine.traverseToWorkflow(ctx, nsc, namespace, path)
	if err != nil {
		return nil, err
	}

	d := new(refData)
	d.wfData = wd

	var query *ent.RefQuery

	if ref == "" {

		// use router to select version

		routes, err := d.wf.QueryRoutes().All(ctx)
		if err != nil {
			return nil, err
		}

		if len(routes) == 0 {

			ref = latest

		} else {

			weight := 0

			for _, route := range routes {
				weight += route.Weight
			}

			n := rand.Int()

			n = n % weight

			var route *ent.Route

			for idx := range routes {
				route = routes[idx]
				n -= route.Weight
				if n < 0 {
					break
				}
			}

			query = route.QueryRef()

		}

	}

	if query == nil {
		query = d.wf.QueryRefs().Where(entref.NameEQ(ref))
	}

	d.ref, err = query.WithRevision().Only(ctx)
	if err != nil {
		return nil, err
	}

	if d.ref.Edges.Revision == nil {
		err = &derrors.NotFoundError{
			Label: "revision not found",
		}
		return nil, err
	}

	d.ref.Edges.Workflow = d.wf
	d.ref.Edges.Revision.Edges.Workflow = d.wf

	return d, nil

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

func (flow *flow) configureRouter(ctx context.Context, evc *ent.EventsClient, wf **ent.Workflow, flags int, changer, commit func() error) error {

	var err error
	var muxErr1 error
	var ms1 *muxStart
	var existingRoutes int

	if !hasFlag(flags, rcfNoPriors) {
		// NOTE: we check router valid before deleting because there's no sense failing the
		// operation for resulting in an invalid router if the router was already invalid.
		ms1, muxErr1, err = validateRouter(ctx, *wf)
		if err != nil {
			return err
		}

		existingRoutes, err = (*wf).QueryRoutes().Count(ctx)
		if err != nil {
			return err
		}
	}

	err = changer()
	if err != nil {
		return err
	}

	ms2, muxErr2, err := validateRouter(ctx, *wf)
	if err != nil {
		return err
	}

	if muxErr2 != nil {

		if hasFlag(flags, rcfNoValidate) && existingRoutes <= 1 {
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
		err = flow.preCommitRouterConfiguration(ctx, evc, *wf, ms2)
		if err != nil {
			return err
		}
	}

	err = commit()

	if err != nil {
		return err
	}

	if mustReconfigureRouter {
		flow.postCommitRouterConfiguration((*wf).ID.String(), ms2)
	}

	return nil

}

func (flow *flow) preCommitRouterConfiguration(ctx context.Context, evc *ent.EventsClient, wf *ent.Workflow, ms *muxStart) error {

	err := flow.events.processWorkflowEvents(ctx, evc, wf, ms)
	if err != nil {
		return err
	}

	return nil

}

func (flow *flow) postCommitRouterConfiguration(id string, ms *muxStart) {

	flow.pubsub.ConfigureRouterCron(id, ms.Cron, ms.Enabled)

}

func (flow *flow) configureRouterHandler(req *PubsubUpdate) {

	msg := new(configureRouterMessage)

	err := unmarshal(req.Key, msg)
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

	id := string(data)

	ctx, conn, err := flow.engine.lock(id, defaultLockWait)
	if err != nil {
		flow.sugar.Error(err)
		return
	}
	defer flow.engine.unlock(id, conn)

	d, err := flow.reverseTraverseToWorkflow(ctx, id)
	if err != nil {

		if derrors.IsNotFound(err) {
			flow.sugar.Infof("Cron failed to find workflow. Deleting cron.")
			flow.timers.deleteCronForWorkflow(id)
			return
		}

		flow.sugar.Error(err)
		return

	}

	k, err := d.wf.QueryInstances().Where(entinst.CreatedAtGT(time.Now().Add(-time.Second*30)), entinst.HasRuntimeWith(entirt.CallerData(util.CallerCron))).Count(ctx)
	if err != nil {
		flow.sugar.Error(err)
		return
	}

	if k != 0 {
		// already triggered
		return
	}

	args := new(newInstanceArgs)
	args.Namespace = d.namespace()
	args.Path = d.path
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

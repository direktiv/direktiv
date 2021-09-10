package flow

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/vorteil/direktiv/pkg/flow/ent"
	entref "github.com/vorteil/direktiv/pkg/flow/ent/ref"
)

func validateRouter(ctx context.Context, wf *ent.Workflow) (error, error) {

	routes, err := wf.QueryRoutes().WithRef(func(q *ent.RefQuery) {
		q.WithRevision()
	}).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(routes) == 0 {

		// latest
		ref, err := wf.QueryRefs().Where(entref.NameEQ(latest)).WithRevision().Only(ctx)
		if err != nil {
			return nil, err
		}

		_, err = loadSource(ref.Edges.Revision)
		if err != nil {
			return err, nil
		}

		return nil, nil

	}

	var startHash string
	var startRef string

	for _, route := range routes {

		workflow, err := loadSource(route.Edges.Ref.Edges.Revision)
		if err != nil {
			return fmt.Errorf("route to '%s' invalid because revision fails to compile: %v", route.Edges.Ref.Name, err), nil
		}

		hash := checksum(workflow.Start)
		if startHash == "" {
			startHash = hash
			startRef = route.Edges.Ref.Name
		} else {
			if startHash != hash {
				return fmt.Errorf("incompatible start definitions between refs '%s' and '%s'", startRef, route.Edges.Ref.Name), nil
			}
		}

	}

	return nil, nil

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

	d.ref.Edges.Workflow = d.wf
	d.ref.Edges.Revision.Edges.Workflow = d.wf

	return d, nil

}

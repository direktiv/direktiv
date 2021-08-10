package direktiv

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/isolates"
	igrpc "github.com/vorteil/direktiv/pkg/isolates/grpc"
	"github.com/vorteil/direktiv/pkg/model"
)

func cancelJob(ctx context.Context, client igrpc.IsolatesServiceClient,
	actionID string) {

	log.Debugf("cancelling job %v", actionID)

	cr := igrpc.CancelPodRequest{
		ActionID: &actionID,
	}

	_, err := client.CancelIsolatePod(ctx, &cr)
	if err != nil {
		log.Errorf("can not cancel job %s: %v", actionID, err)
	}

}

func addPodFunction(ctx context.Context,
	client igrpc.IsolatesServiceClient, ir *isolateRequest) (string, error) {

	sz := int32(ir.Container.Size)
	scale := int32(ir.Container.Scale)
	step := int64(ir.Workflow.Step)

	cr := igrpc.CreatePodRequest{
		Info: &igrpc.BaseInfo{
			Name:      &ir.Container.ID,
			Namespace: &ir.Workflow.Namespace,
			Workflow:  &ir.Workflow.ID,
			Image:     &ir.Container.Image,
			Cmd:       &ir.Container.Cmd,
			Size:      &sz,
			MinScale:  &scale,
		},
		ActionID:   &ir.ActionID,
		InstanceID: &ir.Workflow.InstanceID,
		Step:       &step,
	}

	r, err := client.CreateIsolatePod(ctx, &cr)
	return r.GetIp(), err

}

func isKnativeFunction(client igrpc.IsolatesServiceClient,
	name, namespace, workflow string) bool {

	// search annotations
	a := make(map[string]string)
	a[isolates.ServiceHeaderName] = name
	a[isolates.ServiceHeaderNamespace] = namespace
	a[isolates.ServiceHeaderWorkflow] = workflow
	a[isolates.ServiceHeaderScope] = isolates.PrefixService

	log.Debugf("knative function search: %v", a)

	l, err := client.ListIsolates(context.Background(), &igrpc.ListIsolatesRequest{
		Annotations: a,
	})

	if err != nil {
		log.Errorf("can not list knative service: %v", err)
		return false
	}

	if len(l.Isolates) > 0 {
		return true
	}

	return false
}

func createKnativeFunction(client igrpc.IsolatesServiceClient,
	ir *isolateRequest) error {

	sz := int32(ir.Container.Size)
	scale := int32(ir.Container.Scale)

	cr := igrpc.CreateIsolateRequest{
		Info: &igrpc.BaseInfo{
			Name:      &ir.Container.ID,
			Namespace: &ir.Workflow.Namespace,
			Workflow:  &ir.Workflow.ID,
			Image:     &ir.Container.Image,
			Cmd:       &ir.Container.Cmd,
			Size:      &sz,
			MinScale:  &scale,
		},
	}

	_, err := client.CreateIsolate(context.Background(), &cr)

	return err

}

func createKnativeFunctions(client igrpc.IsolatesServiceClient, wfm model.Workflow, ns string) error {

	for _, f := range wfm.GetFunctions() {

		// only build workflow based isolates
		if f.GetType() != model.ReusableContainerFunctionType {
			continue
		}

		fn := f.(*model.ReusableFunctionDefinition)

		// create services async
		go func(fd *model.ReusableFunctionDefinition,
			model model.Workflow, name, namespace string) {

			sz := int32(fd.Size)
			scale := int32(fd.Scale)

			cr := igrpc.CreateIsolateRequest{
				Info: &igrpc.BaseInfo{
					Name:      &name,
					Namespace: &namespace,
					Workflow:  &model.ID,
					Image:     &fd.Image,
					Cmd:       &fd.Cmd,
					Size:      &sz,
					MinScale:  &scale,
				},
			}

			_, err := client.CreateIsolate(context.Background(), &cr)
			if err != nil {
				log.Errorf("can not create knative service: %v", err)
			}

		}(fn, wfm, fn.ID, ns)

	}

	return nil
}

func deleteKnativeFunctions(client igrpc.IsolatesServiceClient,
	ns, wf, name string) error {

	annotations := make(map[string]string)

	scope := isolates.PrefixService

	if ns != "" {
		annotations[isolates.ServiceHeaderNamespace] = ns
		scope = isolates.PrefixNamespace
	}

	if wf != "" {
		annotations[isolates.ServiceHeaderWorkflow] = wf
		scope = isolates.PrefixWorkflow
	}

	if name != "" {
		annotations[isolates.ServiceHeaderName] = name
		scope = isolates.PrefixService
	}
	annotations[isolates.ServiceHeaderScope] = scope

	dr := igrpc.ListIsolatesRequest{
		Annotations: annotations,
	}

	_, err := client.DeleteIsolates(context.Background(), &dr)
	if err != nil {
		log.Errorf("can not delete knative service: %v", err)
	}

	return nil

}

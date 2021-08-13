package direktiv

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/functions"
	igrpc "github.com/vorteil/direktiv/pkg/functions/grpc"
	"github.com/vorteil/direktiv/pkg/model"
)

func cancelJob(ctx context.Context, client igrpc.FunctionsServiceClient,
	actionID string) {

	log.Debugf("cancelling job %v", actionID)

	cr := igrpc.CancelPodRequest{
		ActionID: &actionID,
	}

	_, err := client.CancelFunctionsPod(ctx, &cr)
	if err != nil {
		log.Errorf("can not cancel job %s: %v", actionID, err)
	}

}

func addPodFunction(ctx context.Context,
	client igrpc.FunctionsServiceClient, ir *functionRequest) (string, string, error) {

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

	r, err := client.CreateFunctionsPod(ctx, &cr)
	return r.GetHostname(), r.GetIp(), err

}

func isKnativeFunction(client igrpc.FunctionsServiceClient,
	name, namespace, workflow string) bool {

	// search annotations
	a := make(map[string]string)
	a[functions.ServiceHeaderName] = name
	a[functions.ServiceHeaderNamespace] = namespace
	a[functions.ServiceHeaderWorkflow] = workflow
	a[functions.ServiceHeaderScope] = functions.PrefixService

	log.Debugf("knative function search: %v", a)

	l, err := client.ListFunctions(context.Background(), &igrpc.ListFunctionsRequest{
		Annotations: a,
	})

	if err != nil {
		log.Errorf("can not list knative service: %v", err)
		return false
	}

	if len(l.Functions) > 0 {
		return true
	}

	return false
}

func createKnativeFunction(client igrpc.FunctionsServiceClient,
	ir *functionRequest) error {

	sz := int32(ir.Container.Size)
	scale := int32(ir.Container.Scale)

	cr := igrpc.CreateFunctionRequest{
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

	_, err := client.CreateFunction(context.Background(), &cr)

	return err

}

func createKnativeFunctions(client igrpc.FunctionsServiceClient,
	wfm model.Workflow, ns string) error {

	for _, f := range wfm.GetFunctions() {

		// only build workflow based functions
		if f.GetType() != model.ReusableContainerFunctionType {
			continue
		}

		fn := f.(*model.ReusableFunctionDefinition)

		// create services async
		go func(fd *model.ReusableFunctionDefinition,
			model model.Workflow, name, namespace string) {

			sz := int32(fd.Size)
			scale := int32(fd.Scale)

			cr := igrpc.CreateFunctionRequest{
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

			_, err := client.CreateFunction(context.Background(), &cr)
			if err != nil {
				log.Errorf("can not create knative service: %v", err)
			}

		}(fn, wfm, fn.ID, ns)

	}

	return nil
}

func deleteKnativeFunctions(client igrpc.FunctionsServiceClient,
	ns, wf, name string) error {

	annotations := make(map[string]string)

	scope := functions.PrefixService

	if ns != "" {
		annotations[functions.ServiceHeaderNamespace] = ns
		scope = functions.PrefixNamespace
	}

	if wf != "" {
		annotations[functions.ServiceHeaderWorkflow] = wf
		scope = functions.PrefixWorkflow
	}

	if name != "" {
		annotations[functions.ServiceHeaderName] = name
		scope = functions.PrefixService
	}
	annotations[functions.ServiceHeaderScope] = scope

	dr := igrpc.ListFunctionsRequest{
		Annotations: annotations,
	}

	_, err := client.DeleteFunctions(context.Background(), &dr)
	if err != nil {
		log.Errorf("can not delete knative service: %v", err)
	}

	return nil

}

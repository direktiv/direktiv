package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bradfitz/slice"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	"github.com/direktiv/direktiv/pkg/flow/ent/predicate"

	"github.com/direktiv/direktiv/pkg/flow/ent/services"

	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/google/uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"

	v1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/serving/pkg/client/clientset/versioned"
)

func (is *functionsServer) storeService(ctx context.Context, info *igrpc.BaseInfo) error {

	svcName, _, _ := GenerateServiceName(info)

	uid, err := uuid.Parse(info.GetNamespace())
	if err != nil {
		return err
	}

	// check if it exists
	logger.Infof("adding/updating service %s to namespace %s", info.GetName(), uid)

	b, err := json.Marshal(info)
	if err != nil {
		return err
	}

	svc, err := is.db.Services.Query().Where(services.And(
		services.Name(info.GetName()),
		services.HasNamespaceWith(
			predicate.Namespace(namespace.ID(uid)),
		),
	)).Only(ctx)

	if err != nil && ent.IsNotFound(err) {

		logger.Infof("creating service %v", info.GetName())
		return is.db.Services.Create().
			SetNamespaceID(uid).
			SetName(info.GetName()).
			SetURL(svcName).
			SetData(string(b)).
			Exec(ctx)

	} else if err != nil {
		return err
	}

	logger.Infof("updating service %v", info.GetName())

	return svc.Update().
		SetData(string(b)).
		SetURL(svcName).
		Exec(ctx)

}

func (is *functionsServer) CreateFunction(ctx context.Context, in *igrpc.CreateFunctionRequest) (*emptypb.Empty, error) {

	logger.Infof("storing function %s", in.GetInfo().GetName())

	err := validateLabel(in.GetInfo().GetName())
	if err != nil {
		logger.Errorf("can not create knative service: %v", err)
		return &empty, status.Error(codes.InvalidArgument, err.Error())
	}

	// save if namespace scoped
	if in.Info.GetWorkflow() == "" {
		err = is.storeService(ctx, in.GetInfo())
		if err != nil {
			logger.Errorf("can not store knative service: %v", err)
			return &empty, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	// create ksvc
	_, err = createKnativeFunction(in.GetInfo())
	if err != nil {
		logger.Errorf("can not create knative service: %v", err)
		return &empty, k8sToGRPCError(err)
	}

	return &empty, nil

}

func fetchServiceAPI() (*versioned.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Errorf("error getting api: %v", err)
		return nil, err
	}
	return versioned.NewForConfig(config)
}

// Available prefixes for different scopes
const (
	PrefixWorkflow  = "workflow"
	PrefixNamespace = "namespace"
)

func (is *functionsServer) DeleteRevision(ctx context.Context,
	in *igrpc.DeleteRevisionRequest) (*emptypb.Empty, error) {

	logger.Debugf("delete revision %v", in.GetRevision())
	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return &empty, err
	}

	err = cs.ServingV1().Revisions(functionsConfig.Namespace).
		Delete(context.Background(), in.GetRevision(), metav1.DeleteOptions{})
	if err != nil {
		logger.Errorf("error delete knative revision %s: %v", in.GetRevision(), err)
		return &empty, err
	}

	return &empty, nil
}

func (is *functionsServer) DeleteFunctions(ctx context.Context,
	in *igrpc.ListFunctionsRequest) (*emptypb.Empty, error) {

	logger.Debugf("deleting functions %v", in.GetAnnotations())

	svcList, err := deleteKnativeFunctions(in.GetAnnotations())
	if err != nil {
		logger.Errorf("error delete knative services %s: %w", in.GetAnnotations(), err)
		return &empty, err
	}

	// Delete Database records
	logger.Debugf("deleting database records %v", in.GetAnnotations())

	conditions := make([]predicate.Services, 0)
	for i := range svcList {
		conditions = append(conditions, services.Name(svcList[i]))
	}

	deleteRecord := is.db.Services.Delete().Where(services.Or(conditions...))
	recordCount, err := deleteRecord.Exec(ctx)
	if err != nil {
		logger.Errorf("error delete knative database record %s: %w", in.GetAnnotations(), err)
		return &empty, err
	}

	logger.Debugf("deleted %v database records", recordCount)

	return &empty, err

}

func (is *functionsServer) GetFunction(ctx context.Context,
	in *igrpc.GetFunctionRequest) (*igrpc.GetFunctionResponse, error) {

	name := SanitizeLabel(in.GetServiceName())

	logger.Debugf("get function %v", name)

	var resp *igrpc.GetFunctionResponse

	if name == "" {
		return resp, fmt.Errorf("service name can not be nil")
	}

	return getKnativeFunction(name)

}

// ListPods returns pods based on label filter
func (is *functionsServer) ListPods(ctx context.Context,
	in *igrpc.ListPodsRequest) (*igrpc.ListPodsResponse, error) {

	var resp igrpc.ListPodsResponse

	items, err := listPods(in.GetAnnotations())
	if err != nil {
		return &resp, err
	}

	resp.Pods = items
	return &resp, nil
}

// ListFunctions returns isoaltes based on label filter
func (is *functionsServer) ListFunctions(ctx context.Context,
	in *igrpc.ListFunctionsRequest) (*igrpc.ListFunctionsResponse, error) {

	var resp igrpc.ListFunctionsResponse

	items, err := listKnativeFunctions(in.GetAnnotations())
	if err != nil {
		return &resp, err
	}

	resp.Functions = items

	var ms = int32(functionsConfig.MaxScale)
	resp.Config = &igrpc.FunctionsConfig{
		Maxscale: &ms,
	}

	return &resp, nil

}

func (is *functionsServer) ReconstructFunction(ctx context.Context,
	in *igrpc.ReconstructFunctionRequest) (*emptypb.Empty, error) {

	name := in.GetName()
	logger.Infof("reconstructing functions %s", name)

	if name == "" {
		return &empty, fmt.Errorf("name can not be nil")
	}

	err := is.reconstructService(name, ctx)
	if err != nil {
		logger.Errorf("could not recreate service: %v", err)

		// Service backup record not found in database
		if ent.IsNotFound(err) {
			return &empty, status.Error(codes.NotFound, "could not recreate service")
		}
		return &empty, fmt.Errorf("could not recreate service")
	}

	return &empty, nil

}

func deleteKnativeFunction(name string) error {

	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return err
	}

	return cs.ServingV1().Services(functionsConfig.Namespace).Delete(context.Background(),
		name, metav1.DeleteOptions{})

}

func (is *functionsServer) DeleteFunction(ctx context.Context,
	in *igrpc.GetFunctionRequest) (*emptypb.Empty, error) {

	logger.Debugf("deleting function %v", in.GetServiceName())

	err := deleteKnativeFunction(in.GetServiceName())
	if err != nil {
		logger.Errorf("can not delete knative service: %v", err)
		return &empty, err
	}

	if strings.HasPrefix(in.GetServiceName(), "namespace-") {
		deleteRecord := is.db.Services.Delete().Where(services.And(
			services.Name(in.GetServiceName()),
		))
		_, err := deleteRecord.Exec(ctx)

		if err != nil {
			logger.Errorf("successfully delete service, but could not delete backup record: %v", err)
			return &empty, fmt.Errorf("successfully delete service, but could not delete backup record: %v", err)
		}
		logger.Infof("Successfully deleted knative service and record")
	}

	return &empty, nil

}

func (is *functionsServer) UpdateFunction(ctx context.Context,
	in *igrpc.UpdateFunctionRequest) (*emptypb.Empty, error) {

	logger.Infof("updating function %s", in.GetServiceName())

	if in.GetInfo() == nil {
		return &empty, fmt.Errorf("info can not be nil")
	}

	// create ksvc service
	_, err := updateKnativeFunction(in.GetServiceName(), in.GetInfo())
	if err != nil {
		logger.Errorf("can not update knative service: %v", err)
		return &empty, err
	}

	if in.GetInfo().GetWorkflow() == "" {
		err = is.storeService(ctx, in.GetInfo())
		if err != nil {
			logger.Errorf("can not store knative service: %v", err)
			return &empty, err
		}
	}

	return &empty, nil
}

func listKnativeFunctions(annotations map[string]string) ([]*igrpc.FunctionsInfo, error) {

	var b []*igrpc.FunctionsInfo

	logger.Debugf("list annotations: %s", labels.Set(annotations).String())

	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return b, err
	}

	lo := metav1.ListOptions{LabelSelector: labels.Set(annotations).String()}
	l, err := cs.ServingV1().Services(functionsConfig.Namespace).List(context.Background(), lo)

	if err != nil {
		logger.Errorf("error getting functions list: %v", err)
		return b, err
	}

	logger.Debugf("%d functions", len(l.Items))

	for i := range l.Items {

		svc := l.Items[i]
		status, conds := statusFromCondition(svc.Status.Conditions)

		ii := &igrpc.FunctionsInfo{
			Info:        serviceBaseInfo(&svc),
			ServiceName: &svc.Name,
			Status:      &status,
			Conditions:  conds,
		}

		b = append(b, ii)
	}

	return b, nil
}

func listPods(annotations map[string]string) ([]*igrpc.PodsInfo, error) {

	var b []*igrpc.PodsInfo

	logger.Debugf("list annotations: %s", labels.Set(annotations).String())

	cs, err := getClientSet()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return b, err
	}

	lo := metav1.ListOptions{LabelSelector: labels.Set(annotations).String()}
	l, err := cs.CoreV1().Pods(functionsConfig.Namespace).List(context.Background(), lo)

	if err != nil {
		logger.Errorf("error getting functions list: %v", err)
		return b, err
	}

	for i := range l.Items {

		pod := l.Items[i]
		sn := pod.Labels[ServiceKnativeHeaderName]
		sr := pod.Labels[ServiceKnativeHeaderRevision]
		ii := &igrpc.PodsInfo{
			Name:            &pod.Name,
			Status:          (*string)(&pod.Status.Phase),
			ServiceName:     &sn,
			ServiceRevision: &sr,
		}

		b = append(b, ii)
	}

	logger.Debugf("list done")

	return b, nil
}

func getKnativeFunction(name string) (*igrpc.GetFunctionResponse, error) {

	var (
		revs []*igrpc.Revision
	)

	logger.Infof("fetching knative service %s", name)

	resp := &igrpc.GetFunctionResponse{}

	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return resp, k8sToGRPCError(err)
	}

	svc, err := cs.ServingV1().Services(functionsConfig.Namespace).Get(context.Background(),
		name, metav1.GetOptions{})

	if err != nil {
		logger.Errorf("error getting knative service: %v", err)
		return resp, k8sToGRPCError(err)
	}

	n := svc.Labels[ServiceHeaderName]
	nsID := svc.Labels[ServiceHeaderNamespaceID]
	nsName := svc.Labels[ServiceHeaderNamespaceName]
	workflow := svc.Labels[ServiceHeaderWorkflowID]
	path := svc.Labels[ServiceHeaderPath]
	revision := svc.Labels[ServiceHeaderRevision]

	resp.Name = &n
	resp.Namespace = &nsName
	resp.NamespaceID = &nsID
	resp.Workflow = &workflow
	resp.Path = &path
	resp.WorkflowRevision = &revision
	resp.Scope = &strings.Split(name, "-")[0]

	rs, err := cs.ServingV1().Revisions(functionsConfig.Namespace).List(context.Background(),
		metav1.ListOptions{LabelSelector: fmt.Sprintf("serving.knative.dev/service=%s", name)})
	if err != nil {
		logger.Errorf("error getting knative service: %v", err)
		return resp, k8sToGRPCError(err)
	}

	fn := func(rev v1.Revision) *igrpc.Revision {
		info := &igrpc.Revision{}

		// size and scale
		var sz, scale int32
		var gen int64
		fmt.Sscan(rev.Annotations[ServiceHeaderSize], &sz)
		fmt.Sscan(rev.Annotations["autoscaling.knative.dev/minScale"], &scale)
		fmt.Sscan(rev.Labels[ServiceTemplateGeneration], &gen)

		info.Size = &sz
		info.MinScale = &scale
		info.Generation = &gen

		// set status
		status, conds := statusFromCondition(rev.Status.Conditions)
		info.Status = &status
		info.Conditions = conds

		img, cmd := containerFromList(rev.Spec.Containers)
		info.Image = &img
		info.Cmd = &cmd

		// name
		svn := rev.Name
		info.Name = &svn

		ss := strings.Split(rev.Name, "-")
		info.Rev = &ss[len(ss)-1]

		// replicas
		if rev.Status.ActualReplicas != nil {
			info.ActualReplicas = int64(*rev.Status.ActualReplicas)
		}

		if rev.Status.DesiredReplicas != nil {
			info.DesiredReplicas = int64(*rev.Status.DesiredReplicas)
		}

		// creation date
		var t int64 = rev.CreationTimestamp.Unix()
		info.Created = &t

		return info
	}

	// get details
	for i := range rs.Items {
		r := rs.Items[i]
		info := fn(r)
		revs = append(revs, info)
	}

	slice.Sort(revs[:], func(i, j int) bool {
		return *revs[i].Generation > *revs[j].Generation
	})

	resp.Revisions = revs

	// add config
	var ms = int32(functionsConfig.MaxScale)
	resp.Config = &igrpc.FunctionsConfig{
		Maxscale: &ms,
	}
	return resp, nil

}

func deleteKnativeFunctions(annotations map[string]string) ([]string, error) {

	logger.Debugf("delete annotations: %s", labels.Set(annotations).String())

	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return nil, err
	}

	annotations = fixAnnotations(annotations)

	lo := metav1.ListOptions{LabelSelector: labels.Set(annotations).String()}

	// Get list of services will be deleted
	svcList, err := cs.ServingV1().Services(functionsConfig.Namespace).List(context.Background(), lo)
	if err != nil {
		logger.Errorf("error getting service list from knative: %v", err)
		return nil, err
	}

	servicesToDelete := make([]string, 0)
	for i := range svcList.Items {
		servicesToDelete = append(servicesToDelete, svcList.Items[i].Name)
	}

	err = cs.ServingV1().Services(functionsConfig.Namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, lo)
	return servicesToDelete, err
}

func updateKnativeFunction(svn string, info *igrpc.BaseInfo) (*v1.Service, error) {

	containers, err := makeContainers(info.GetImage(), info.GetCmd(),
		int(info.GetSize()), info.GetEnvs())
	if err != nil {
		logger.Errorf("can not update service: %v", err)
		return nil, err
	}

	// adjust traffic for new revision
	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return nil, err
	}

	spec := metav1.ObjectMeta{
		Annotations: make(map[string]string),
		Labels:      make(map[string]string),
	}

	spec.Annotations[ServiceHeaderSize] = fmt.Sprintf("%d", info.GetSize())
	spec.Annotations["autoscaling.knative.dev/minScale"] =
		fmt.Sprintf("%d", info.GetMinScale())

	svc := &v1.Service{
		Spec: v1.ServiceSpec{
			ConfigurationSpec: v1.ConfigurationSpec{
				Template: v1.RevisionTemplateSpec{
					ObjectMeta: spec,
					Spec: v1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: containers,
							Volumes:    createVolumes(),
						},
					},
				},
			},
		},
	}

	b, err := json.MarshalIndent(*svc, "", "    ")
	if err != nil {
		logger.Errorf("error marshalling new services: %v", err)
		return nil, err
	}

	logger.Debugf("patching service %s", svn)

	// lock for updates and deletes
	l, err := locksmgr.lock(svn, false)
	if err != nil {
		return nil, err
	}
	defer locksmgr.unlock(svn, l)

	svc, err = cs.ServingV1().Services(functionsConfig.Namespace).Patch(context.Background(),
		svn, types.MergePatchType, b, metav1.PatchOptions{})

	if err != nil {
		logger.Errorf("can not patch service %s: %v", svn, err)
		return nil, err
	}

	return svc, nil
}

func createPullSecrets(namespace string) []corev1.LocalObjectReference {
	var lo []corev1.LocalObjectReference

	secrets := listRegistriesNames(namespace)
	for _, s := range secrets {
		logger.Debugf("adding pull secret: %v", s)
		lo = append(lo, corev1.LocalObjectReference{
			Name: s,
		})
	}

	return lo
}

func (is *functionsServer) reconstructService(name string, ctx context.Context) error {

	// Get backed up service from database
	dbSvc, err := is.db.Services.Query().Where(services.URL(name)).First(ctx)
	if err != nil {
		return err
	}

	var info igrpc.BaseInfo
	err = json.Unmarshal([]byte(dbSvc.Data), &info)
	if err != nil {
		return err
	}

	// create ksvc
	_, err = createKnativeFunction(&info)
	if err != nil {
		logger.Errorf("can not create knative service: %v", err)
		return err
	}

	return nil
}

// recretae all services on startup
func (is *functionsServer) reconstructServices(ctx context.Context) error {

	svcs, err := is.db.Services.Query().All(ctx)
	if err != nil {
		logger.Error("failed to get services from database: %w", err)
		return err
	}

	for s := range svcs {
		svc := svcs[s]

		var info igrpc.BaseInfo
		err = json.Unmarshal([]byte(svc.Data), &info)
		if err != nil {
			logger.Errorf("could not recreate service on startup: %v", err)
			continue
		}

		// create ksvc
		_, err = createKnativeFunction(&info)
		if err != nil {
			logger.Errorf("could not recreate service on startup: %v", err)
			continue
		}

	}

	return nil
}

func (is *functionsServer) CancelWorfklow(ctx context.Context, in *igrpc.CancelWorkflowRequest) (*emptypb.Empty, error) {

	label := "serving.knative.dev/service"

	svn := SanitizeLabel(in.GetServiceName())
	aid := in.GetActionID()

	if svn == "" || aid == "" {
		return &empty, fmt.Errorf("service name or action id can not be empty")
	}

	logger.Infof("cancelling action %s on %s", aid, svn)

	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{
		label: svn,
	}}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	cs, err := getClientSet()
	if err != nil {
		logger.Errorf("error getting client set: %v", err)
		return &empty, err
	}

	podList, err := cs.CoreV1().Pods(functionsConfig.Namespace).List(context.Background(),
		listOptions)

	if err != nil {
		logger.Errorf("could not get cancel list: %v", err)
		return &empty, err
	}

	for i := range podList.Items {

		service := podList.Items[i].ObjectMeta.Labels[label]

		// cancel request to pod
		go func(name, ns, svc string) {
			logger.Infof("cancelling %v", name)
			addr := fmt.Sprintf("http://%s.%s/cancel", svc, ns)

			req, err := http.NewRequest(http.MethodPost, addr, nil)
			if err != nil {
				logger.Errorf("error creating delete request: %v", err)
				return
			}
			req.Header.Add("Direktiv-ActionID", aid)

			client := http.Client{
				Timeout: 60 * time.Second,
			}
			_, err = client.Do(req)
			if err != nil {
				logger.Errorf("error sending delete request: %v", err)
			}

		}(podList.Items[i].Name, functionsConfig.Namespace, service)

	}

	return &empty, nil

}

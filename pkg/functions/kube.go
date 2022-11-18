package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bradfitz/slice"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/mitchellh/hashstructure/v2"
	hash "github.com/mitchellh/hashstructure/v2"

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

const (
	httpsProxy = "HTTPS_PROXY"
	httpProxy  = "HTTP_PROXY"
	noProxy    = "NO_PROXY"

	containerUser    = "direktiv-container"
	containerSidecar = "direktiv-sidecar"
)

// ---------------------------------------------------------------------------------------

type fnType struct {
	prefix string
	full   string
}

var (
	wfType = fnType{
		prefix: "wf",
		full:   "workflow",
	}
	nsType = fnType{
		prefix: "ns",
		full:   "namespace",
	}
)

type functionInformation struct {
	Cmd      string            `json:"cmd,omitempty"`
	Image    string            `json:"image,omitempty"`
	Name     string            `json:"name,omitempty"`
	Size     int32             `json:"size,omitempty"`
	MinScale int32             `json:"minScale,omitempty"`
	Envs     map[string]string `json:"envs"`

	NamespaceOID string
	WorkflowPath string
}

func generateServiceName(fn *functionInformation) (string, error) {

	hash, err := hashstructure.Hash(fn, hashstructure.FormatV2, nil)
	if err != nil {
		return "", err
	}

	scope := wfType
	if fn.NamespaceOID != "" {
		scope = nsType
	}

	return fmt.Sprintf("%s-%d", scope.prefix, hash), nil

}

// func GenerateServiceName(info *igrpc.BaseInfo /* ns, wf, n string*/) (string, string, string) {

// 	var name, scope, hash string

// 	if info.GetWorkflow() != "" {
// 		scope = PrefixWorkflow
// 		name, hash = GenerateWorkflowServiceName(info)
// 	} else if info.GetNamespace() != "" {
// 		scope = PrefixNamespace
// 		name = fmt.Sprintf("%s-%s-%s", PrefixNamespace, info.GetNamespaceName(), info.GetName())
// 		hash = ""
// 	} else {
// 		scope = PrefixGlobal
// 		name = fmt.Sprintf("%s-%s", PrefixGlobal, info.GetName())
// 		hash = ""
// 	}

// 	return name, scope, hash

// }

// GenerateWorkflowServiceName generates a knative name based on workflow details
// func GenerateWorkflowServiceName(info *igrpc.BaseInfo) (string, string) {

// 	wf := SanitizeLabel(info.GetWorkflow())
// 	fndef := fndefFromBaseInfo(info)

// 	// NOTE: fndef.Files can be safely excluded

// 	var strs []string
// 	strs = []string{fndef.Cmd, fndef.ID, fndef.Image, fmt.Sprintf("%v", fndef.Size), fmt.Sprintf("%v", fndef.Type)}

// 	def, err := json.Marshal(strs)
// 	if err != nil {
// 		panic(err)
// 	}

// 	svn := SanitizeLabel(fndef.ID)

// 	h, err := hash.Hash(fmt.Sprintf("%s-%s", wf, def), hash.FormatV2, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	name := fmt.Sprintf("%s-%d-%s", PrefixWorkflow, h, svn)

// 	return name, fmt.Sprintf("%v", h)

// }

func (is *functionsServer) storeService(ctx context.Context, fn *functionInformation) error {

	// svn, err := generateServiceName(fn)
	// if err != nil {
	// 	return err
	// }

	// uid, err := uuid.Parse(fn.NamespaceOID)
	// if err != nil {
	// 	return err
	// }

	// // check if it exists
	// logger.Infof("adding/updating service %s to workflow %s", svn, uid)

	// b, err := json.Marshal(fn)
	// if err != nil {
	// 	return err
	// }

	// svc, err := is.db.Services.Query().Where(entservices.And(
	// 	entservices.Name(fn.Name),
	// 	entservices.HasNamespaceWith(
	// 		predicate.Namespace(namespace.ID(uid)),
	// 	),
	// )).Only(ctx)

	// if err != nil && ent.IsNotFound(err) {

	// 	logger.Infof("creating service %v", svn)
	// 	return is.db.Services.Create().
	// 		SetNamespaceID(uid).
	// 		SetName(fn.Name).
	// 		SetURL(svn).
	// 		SetData(string(b)).
	// 		Exec(ctx)

	// } else if err != nil {
	// 	return err
	// }

	// // only update if names are different
	// // we need to delete ksvc too
	// if svc.URL != svn {

	// 	logger.Infof("updating service %v", svn)

	// 	// DELETE KSVC HERE
	// 	return svc.Update().
	// 		SetData(string(b)).
	// 		SetURL(svn).
	// 		Exec(ctx)
	// }

	return nil

}

// StoreFunctions saves or updates functions which means creating knative services
// based on the provided configuration
func (is *functionsServer) CreateFunction(ctx context.Context, in *igrpc.CreateFunctionRequest) (*emptypb.Empty, error) {

	logger.Infof("storing function %s", in.GetInfo().GetName())

	err := validateLabel(in.GetInfo().GetName())
	if err != nil {
		logger.Errorf("can not create knative service: %v", err)
		return &empty, status.Error(codes.InvalidArgument, err.Error())
	}

	// save if namespace scoped
	if in.Info.GetNamespaceName() != "" {
		fn := functionInformation{
			Cmd:      in.Info.GetCmd(),
			Image:    in.Info.GetImage(),
			Name:     in.Info.GetName(),
			Size:     in.Info.GetSize(),
			MinScale: in.Info.GetMinScale(),
			Envs:     in.Info.GetEnvs(),

			WorkflowPath: in.Info.GetPath(),
			NamespaceOID: in.Info.GetNamespace(),
		}
		err = is.storeService(ctx, &fn)

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

	// _, err = newRecord.Save(context.Background())

	// // create ksvc service
	// svc, err := createKnativeFunction(in.GetInfo())
	// if err != nil {
	// 	logger.Errorf("can not create knative service: %v", err)
	// 	return &empty, k8sToGRPCError(err)
	// }

	// // backup service if not a workflow service
	// if svc.ObjectMeta.Labels[ServiceHeaderWorkflowID] == "" {
	// 	if err := is.backupService(svc.Name, backupServiceOptions{}); err != nil {
	// 		logger.Errorf("can not backup knative service: %v", err)
	// 		return &empty, err
	// 	}
	// }

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

// ------------------------------------------------------------------------------------------------------

// Available prefixes for different scopes
const (
	PrefixWorkflow  = "workflow"
	PrefixNamespace = "namespace"
	PrefixGlobal    = "global"
	PrefixService   = "service" // unused, only if a one item list is requested
)

// const (
// 	serviceType   = iota
// 	workflowType  = iota
// 	namespaceType = iota
// 	globalType    = iota
// 	invalidType   = iota
// )

type serviceExportInfo struct {
	Service   *v1.Service
	Revisions []v1.Revision
}

type backupServiceOptions struct {
	previousRevisionName string
	patch                bool
}

func (is *functionsServer) DeleteRevision(ctx context.Context,
	in *igrpc.DeleteRevisionRequest) (*emptypb.Empty, error) {

	logger.Debugf("delete revision %v", in.GetRevision())
	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return &empty, err
	}

	r, err := cs.ServingV1().Revisions(functionsConfig.Namespace).Get(context.Background(),
		in.GetRevision(), metav1.GetOptions{})
	if err != nil {
		logger.Errorf("error getting revision for %v: %v", in.GetRevision(), err)
		return &empty, err
	}

	svcName := r.Labels[ServiceKnativeHeaderConfiguration]
	resp, err := getKnativeFunction(svcName)
	if err != nil {
		logger.Errorf("error getting svc for %v: %v", svcName, err)
		return &empty, err
	}

	for i := range resp.Revisions {
		rr := resp.Revisions[i]
		if rr.Name != nil && rr.GetName() == in.GetRevision() && rr.GetTraffic() > 0 {
			logger.Errorf("revisions with traffic can not be deleted")
			return &empty, fmt.Errorf("revision %s still has traffic assigned: %d%%",
				in.GetRevision(), rr.GetTraffic())
		}
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

	_, err := deleteKnativeFunctions(in.GetAnnotations())
	if err != nil {
		logger.Errorf("error delete knative services %s: %w", in.GetAnnotations(), err)
		return &empty, err
	}

	// Delete Database records
	// logger.Debugf("deleting database records %v", in.GetAnnotations())

	// deleteRecord := is.db.Services.Delete()
	// conditions := make([]predicate.Services, 0)
	// for i := range svcList {
	// 	conditions = append(conditions, entservices.Name(svcList[i]))
	// }

	// deleteRecord = deleteRecord.Where(entservices.Or(conditions...))
	// recordCount, err := deleteRecord.Exec(ctx)
	// if err != nil {
	// 	logger.Errorf("error delete knative database record %s: %w", in.GetAnnotations(), err)
	// 	return &empty, err
	// }

	// logger.Debugf("deleted %v database records", recordCount)

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

	logger.Debugf("***********************************************8list pods %v", in.GetAnnotations())

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

	// logger.Infof("reconstructing functions %s", in.GetName())

	// if in.GetName() == "" {
	// 	return &empty, fmt.Errorf("name can not be nil")
	// }

	// name := in.GetName()

	// err := is.reconstructService(name, ctx)
	// if err != nil {
	// 	logger.Errorf("could not recreate service: %v", err)

	// 	// Service backup record not found in database
	// 	if ent.IsNotFound(err) {
	// 		return &empty, status.Error(codes.NotFound, "could not recreate service")
	// 	}

	// 	return &empty, fmt.Errorf("could not recreate service")
	// }

	return &empty, nil

}

func (is *functionsServer) WatchLogs(in *igrpc.WatchLogsRequest, out igrpc.FunctionsService_WatchLogsServer) error {

	if in.GetPodName() == "" {
		return fmt.Errorf("pod name can not be nil")
	}

	cs, err := getClientSet()
	if err != nil {
		return fmt.Errorf("could not create fetch client: %v", err)
	}

	req := cs.CoreV1().Pods(functionsConfig.Namespace).GetLogs(*in.PodName, &corev1.PodLogOptions{
		Container: "direktiv-container",
		Follow:    true,
	})

	plogs, err := req.Stream(context.Background())
	if err != nil {
		return fmt.Errorf("could not get logs: %v", err)
	}
	defer plogs.Close()

	var done bool

	// Make sure stream is closed if client disconnects
	go func() {
		<-out.Context().Done()
		plogs.Close()
		done = true
	}()

	for {
		if done {
			break
		}
		buf := make([]byte, 2000)
		numBytes, err := plogs.Read(buf)
		if numBytes == 0 {
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		message := string(buf[:numBytes])
		resp := igrpc.WatchLogsResponse{
			Data: &message,
		}

		err = out.Send(&resp)
		if err != nil {
			return fmt.Errorf("log watcher failed to send event: %v", err)
		}
	}

	return nil
}

func (is *functionsServer) SetFunctionsTraffic(ctx context.Context,
	in *igrpc.SetTrafficRequest) (*emptypb.Empty, error) {

	// svc, err := trafficKnativeFunctions(in.GetName(), in.GetTraffic())
	// if err != nil {
	// 	logger.Errorf("can not set traffic: %v", err)
	// 	return &empty, err
	// }

	// // backup service
	// if svc.ObjectMeta.Labels[ServiceHeaderWorkflowID] == "" {
	// 	if err := is.backupService(svc.Name, backupServiceOptions{
	// 		patch: true,
	// 	}); err != nil {
	// 		logger.Errorf("can not backup knative service: %v", err)
	// 		return &empty, err
	// 	}
	// }

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

	logger.Debugf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!deleting function %v", in.GetServiceName())

	err := deleteKnativeFunction(in.GetServiceName())
	if err != nil {
		logger.Errorf("can not delete knative service: %v", err)
		return &empty, err
	}

	// deleteRecord := is.db.Services.Delete().Where(entservices.Name(in.GetServiceName()))
	// recordCount, err := deleteRecord.Exec(ctx)
	// if err != nil {
	// 	logger.Errorf("successfully delete service, but could not delete backup record: %v", err)
	// 	return &empty, fmt.Errorf("successfully delete service, but could not delete backup record: %v", err)
	// }

	// logger.With("service", in.GetServiceName(), "deleted-records", recordCount).Debug("Successfully deleted knative service and record")
	return &empty, nil

}

func (is *functionsServer) UpdateFunction(ctx context.Context,
	in *igrpc.UpdateFunctionRequest) (*emptypb.Empty, error) {

	logger.Infof("updating function!!!!!!!!!!!!!!!!!!!!!!!!!!! %s", in.GetServiceName())

	if in.GetInfo() == nil {
		return &empty, fmt.Errorf("info can not be nil")
	}

	// Get Last Created Revision
	// cs, err := fetchServiceAPI()
	// if err != nil {
	// 	logger.Errorf("error getting clientset for knative: %v", err)
	// 	return nil, err
	// }

	// previousSvc, err := cs.ServingV1().Services(functionsConfig.Namespace).Get(context.Background(),
	// 	in.GetServiceName(), metav1.GetOptions{})
	// if err != nil {
	// 	logger.Errorf("error getting knative service: %v", err)
	// 	return nil, k8sToGRPCError(err)
	// }

	// create ksvc service
	_, err := updateKnativeFunction(in.GetServiceName(), in.GetInfo())
	if err != nil {
		logger.Errorf("can not update knative service: %v", err)
		return &empty, err
	}

	// // backup service
	// if svc.ObjectMeta.Labels[ServiceHeaderWorkflowID] == "" {
	// 	if err := is.backupService(svc.Name, backupServiceOptions{
	// 		previousRevisionName: previousSvc.Status.LatestCreatedRevisionName,
	// 		patch:                true,
	// 	}); err != nil {
	// 		logger.Errorf("can not backup knative service: %v", err)
	// 		return &empty, err
	// 	}
	// }

	return &empty, nil
}

func listKnativeFunctions(annotations map[string]string) ([]*igrpc.FunctionsInfo, error) {

	var b []*igrpc.FunctionsInfo

	// filtered := filterLabels(annotations)
	// if len(filtered) == 0 {
	// 	return b, fmt.Errorf("request labels are invalid")
	// }

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

// AssembleWorkflowServiceName generates a knative name based on workflow details
func AssembleWorkflowServiceName(wf, svn string, hash uint64) string {

	wf = SanitizeLabel(wf)

	svn = SanitizeLabel(svn)

	name := fmt.Sprintf("%s-%d-%s", PrefixWorkflow, hash, svn)

	return name

}

// GenerateWorkflowServiceName generates a knative name based on workflow details
func GenerateWorkflowServiceName(info *igrpc.BaseInfo) (string, string) {

	wf := SanitizeLabel(info.GetWorkflow())
	fndef := fndefFromBaseInfo(info)

	var strs []string
	strs = []string{fndef.Cmd, fndef.ID, fndef.Image, fmt.Sprintf("%v", fndef.Size), fmt.Sprintf("%v", fndef.Type)}

	def, err := json.Marshal(strs)
	if err != nil {
		panic(err)
	}

	svn := SanitizeLabel(fndef.ID)

	h, err := hash.Hash(fmt.Sprintf("%s-%s", wf, def), hash.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	name := fmt.Sprintf("%s-%d-%s", PrefixWorkflow, h, svn)

	return name, fmt.Sprintf("%v", h)

}

func fndefFromBaseInfo(info *igrpc.BaseInfo) *model.ReusableFunctionDefinition {
	fndef := new(model.ReusableFunctionDefinition)
	fndef.Cmd = info.GetCmd()
	fndef.ID = info.GetName()
	fndef.Image = info.GetImage()
	size := int(info.GetSize())
	fndef.Size = model.Size(size)
	fndef.Type = model.ReusableContainerFunctionType
	return fndef
}

// GenerateServiceName generates a knative name based on workflow details
func GenerateServiceName(info *igrpc.BaseInfo /* ns, wf, n string*/) (string, string, string) {

	var name, scope, hash string

	if info.GetWorkflow() != "" {
		scope = PrefixWorkflow
		name, hash = GenerateWorkflowServiceName(info)
	} else if info.GetNamespace() != "" {
		scope = PrefixNamespace
		name = fmt.Sprintf("%s-%s-%s", PrefixNamespace, SanitizeLabel(info.GetNamespaceName()), info.GetName())
		hash = ""
	} else {
		scope = PrefixGlobal
		name = fmt.Sprintf("%s-%s", PrefixGlobal, info.GetName())
		hash = ""
	}

	return name, scope, hash

}

func getKnativeFunction(name string) (*igrpc.GetFunctionResponse, error) {

	var (
		revs []*igrpc.Revision
	)

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
		int(info.GetSize()))
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

// reconstructService : Reconstructs a service and its revisions from a backed up service database record
// This is done in two steps:
// 1) Create service with earliest recorded revision
// 2) For each other revision create them in asceneding order by patching the exisiting service.
func (is *functionsServer) reconstructService(name string, ctx context.Context) error {

	// cs, err := fetchServiceAPI()
	// if err != nil {
	// 	logger.Errorf("error getting clientset for knative: %v", err)
	// 	return err
	// }

	// // Get backed up service from database
	// dbSVC, err := is.db.Services.Query().Where(entservices.Name(name)).First(ctx)
	// if err != nil {
	// 	return err
	// }

	// l, err := locksmgr.lock(name, false)
	// if err != nil {
	// 	return err
	// }
	// defer locksmgr.unlock(name, l)

	// var recoveredSVC serviceExportInfo
	// err = json.Unmarshal([]byte(dbSVC.Data), &recoveredSVC)
	// if err != nil {
	// 	return err
	// }

	// logger.With("service", name).Debug("reconstructing service")

	// // Recreate Service
	// svc, err := cs.ServingV1().Services(functionsConfig.Namespace).Create(ctx, recoveredSVC.Service, metav1.CreateOptions{})
	// if err != nil {
	// 	logger.Errorf("failed creating service: %w", err)
	// 	return fmt.Errorf("failed creating service: %w", err)
	// }

	// // Recreate Revisions
	// for i := 0; i < len(recoveredSVC.Revisions); i++ {
	// 	tmpRev := recoveredSVC.Revisions[i].DeepCopy()

	// 	// Recover template generation
	// 	gen, err := getTemplateMetaGeneration(&tmpRev.ObjectMeta)
	// 	if err != nil {
	// 		logger.With("service", name, "revision", tmpRev.ObjectMeta.Name).Warnf("error getting direktiv service generation: %w", err)
	// 		logger.With("service", name, "revision", tmpRev.ObjectMeta.Name).Debugf("attempting to get template generation from revision name")

	// 		// If the template generation label is not set, attempt to recover generation from name.
	// 		gen, err = getTemplateMetaGenerationFromName(&tmpRev.ObjectMeta)
	// 		if err != nil {
	// 			logger.With("service", name, "revision", tmpRev.ObjectMeta.Name).Errorf("error getting direktiv service name: %w", err)
	// 			return fmt.Errorf("error getting direktiv service generation")
	// 		}
	// 	}

	// 	logger.With("service", name, "revision", tmpRev.ObjectMeta.Name, "template-generation", gen).Debug("reconstructing service revision")
	// 	revPatch := &v1.Service{
	// 		Spec: v1.ServiceSpec{
	// 			ConfigurationSpec: v1.ConfigurationSpec{
	// 				Template: v1.RevisionTemplateSpec{
	// 					ObjectMeta: metav1.ObjectMeta{
	// 						Annotations: tmpRev.Annotations,
	// 						Labels: map[string]string{
	// 							ServiceTemplateGeneration: fmt.Sprint(gen),
	// 						},
	// 						Name: tmpRev.ObjectMeta.Name,
	// 					},
	// 					Spec: tmpRev.Spec,
	// 				},
	// 			},
	// 		},
	// 	}

	// 	b, err := json.MarshalIndent(*revPatch, "", "    ")
	// 	if err != nil {
	// 		logger.Errorf("error marshalling new services: %v", err)
	// 		return err
	// 	}

	// 	// Patch Service to create new revision
	// 	_, err = cs.ServingV1().Services(functionsConfig.Namespace).Patch(context.Background(),
	// 		svc.Name, types.MergePatchType, b, metav1.PatchOptions{})

	// 	if err != nil {
	// 		logger.Errorf("can not patch service %s: %v", svc.Name, err)
	// 		return err
	// 	}

	// 	// Wait until revision is created
	// 	for i := 0; i < 60; i++ {
	// 		tmpSVC, err := cs.ServingV1().Services(functionsConfig.Namespace).Get(context.Background(), svc.Name, metav1.GetOptions{})
	// 		if err != nil {
	// 			logger.Errorf("error getting service info: %w", err)
	// 			return err
	// 		}

	// 		if tmpSVC.Status.LatestCreatedRevisionName == tmpRev.GetObjectMeta().GetName() {
	// 			break
	// 		}

	// 		time.Sleep(250 * time.Millisecond)
	// 	}
	// }

	return nil
}

// reconstructServices : Checks to see if there are any records in the database of
// backed up services that are missing. If any missing services are found, they
// are reconstructed
func (is *functionsServer) reconstructServices(ctx context.Context) error {

	// cs, err := fetchServiceAPI()
	// if err != nil {
	// 	logger.Errorf("error getting clientset for knative: %v", err)
	// 	return err
	// }

	// // Get Current Namespace and Global Services
	// lblScope, err := labels.NewRequirement(ServiceHeaderScope, selection.In, []string{PrefixNamespace, PrefixGlobal})
	// if err != nil {
	// 	logger.Errorf("invalid label: %v", err)
	// 	return err
	// }

	// svcList, err := cs.ServingV1().Services(functionsConfig.Namespace).List(ctx, metav1.ListOptions{
	// 	LabelSelector: lblScope.String(),
	// })
	// if err != nil {
	// 	logger.Errorf("error getting services", err)
	// 	return err
	// }

	// logger.Infof("Reconstruct Services - Found %v services", len(svcList.Items))

	// // Get missing services that exist in database
	// query := is.db.Services.Query()
	// conditions := make([]predicate.Services, 0)
	// for _, svc := range svcList.Items {
	// 	conditions = append(conditions, entservices.Not(entservices.Name(svc.Name)))
	// }
	// query = query.Where(entservices.And(conditions...))

	// recoverSVCList, err := query.All(ctx)
	// if err != nil {
	// 	logger.Error("failed to get services from database: %w", err)
	// 	return err
	// }

	// logger.Infof("Reconstruct Services - Found %v backup records to recover", len(recoverSVCList))

	// for i := range recoverSVCList {
	// 	logger.Infof("Reconstruct Services - Reconstructing %s ", recoverSVCList[i].Name)
	// 	err := is.reconstructService(recoverSVCList[i].Name, ctx)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// logger.Infof("Reconstruct Services - Successfully recovered %v services", len(recoverSVCList))
	return nil
}

func (is *functionsServer) CancelWorfklow(ctx context.Context, in *igrpc.CancelWorkflowRequest) (*emptypb.Empty, error) {

	label := "serving.knative.dev/service"

	svn := in.GetServiceName()
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

package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bradfitz/slice"
	shellwords "github.com/mattn/go-shellwords"
	hash "github.com/mitchellh/hashstructure/v2"
	igrpc "github.com/vorteil/direktiv/pkg/functions/grpc"
	"github.com/vorteil/direktiv/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/apis"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

const (
	httpsProxy = "HTTPS_PROXY"
	httpProxy  = "HTTP_PROXY"
	noProxy    = "NO_PROXY"

	containerUser    = "direktiv-container"
	containerSidecar = "direktiv-sidecar"

	generationHeader = "serving.knative.dev/configurationGeneration"
)

// Headers for knative services
const (
	ServiceHeaderName      = "direktiv.io/name"
	ServiceHeaderNamespace = "direktiv.io/namespace"
	ServiceHeaderWorkflow  = "direktiv.io/workflow"
	ServiceHeaderSize      = "direktiv.io/size"
	ServiceHeaderScale     = "direktiv.io/scale"
	// ServiceHeaderScale = "autoscaling.knative.dev/minScale"
	ServiceHeaderScope = "direktiv.io/scope"
)

// Available prefixes for different scopes
const (
	PrefixWorkflow  = "workflow"
	PrefixNamespace = "namespace"
	PrefixGlobal    = "global"
	PrefixService   = "service" // unused, only if a one item list is requested
)

const (
	serviceType   = iota
	workflowType  = iota
	namespaceType = iota
	globalType    = iota
	invalidType   = iota
)

const (
	watcherTimeout = 60 * time.Minute
)

var (
	mtx sync.Mutex
)

func (is *functionsServer) DeleteRevision(ctx context.Context,
	in *igrpc.DeleteRevisionRequest) (*emptypb.Empty, error) {

	logger.Debugf("delete revision %v", in.GetRevision())
	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return &empty, err
	}

	// check if there is traffic on it
	// decline if there is still traffic on it
	r, err := cs.ServingV1().Revisions(functionsConfig.Namespace).Get(context.Background(),
		in.GetRevision(), metav1.GetOptions{})
	if err != nil {
		logger.Errorf("error getting revision for %v: %v", in.GetRevision(), err)
		return &empty, err
	}

	svcName := r.Labels["serving.knative.dev/configuration"]
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

	err := deleteKnativeFunctionss(in.GetAnnotations())

	return &empty, err
}

func (is *functionsServer) GetFunction(ctx context.Context,
	in *igrpc.GetFunctionRequest) (*igrpc.GetFunctionResponse, error) {

	logger.Debugf("get function %v", in.GetServiceName())

	var resp *igrpc.GetFunctionResponse

	if in.GetServiceName() == "" {
		return resp, fmt.Errorf("service name can not be nil")
	}

	return getKnativeFunction(in.GetServiceName())

}

// ListPods returns pods based on label filter
func (is *functionsServer) ListPods(ctx context.Context,
	in *igrpc.ListPodsRequest) (*igrpc.ListPodsResponse, error) {

	var resp igrpc.ListPodsResponse

	logger.Debugf("list pods %v", in.GetAnnotations())

	items, err := listPods(in.GetAnnotations())
	if err != nil {
		return &resp, err
	}

	resp.Pods = items
	return &resp, nil
}

func (is *functionsServer) WatchPods(in *igrpc.WatchPodsRequest, out igrpc.FunctionsService_WatchPodsServer) error {

	if in.GetServiceName() == "" {
		return fmt.Errorf("service name can not be nil")
	}

	cs, err := getClientSet()
	if err != nil {
		return fmt.Errorf("could not create fetch client: %v", err)
	}

	l := map[string]string{
		"serving.knative.dev/service": *in.ServiceName,
	}

	if in.GetRevisionName() != "" {
		l["serving.knative.dev/revision"] = *in.RevisionName
	}

	labels := labels.Set(l).String()

	for {
		if done, err := is.watcherPods(cs, labels, out); err != nil {
			logger.Errorf("pod watcher channel failed to restart: %s", err.Error())
			return err
		} else if done {
			// connection has ended
			return nil
		}
		logger.Debugf("pod watcher channel has closed, attempting to restart")
		time.Sleep(5 * time.Second)
	}
}

func (is *functionsServer) watcherPods(cs *kubernetes.Clientset, labels string, out igrpc.FunctionsService_WatchPodsServer) (bool, error) {
	timeout := int64(watcherTimeout.Seconds())
	watch, err := cs.CoreV1().Pods(functionsConfig.Namespace).Watch(context.Background(), metav1.ListOptions{
		LabelSelector:  labels,
		TimeoutSeconds: &timeout,
	})
	if err != nil {
		return false, fmt.Errorf("could start watcher: %v", err)
	}

	for {
		select {
		case event := <-watch.ResultChan():
			p, ok := event.Object.(*corev1.Pod)
			if !ok {
				return false, nil
			}

			svc := p.Labels["serving.knative.dev/service"]
			srev := p.Labels["serving.knative.dev/revision"]

			pod := igrpc.PodsInfo{
				Name:            &p.Name,
				Status:          (*string)(&p.Status.Phase),
				ServiceName:     &svc,
				ServiceRevision: &srev,
			}

			resp := igrpc.WatchPodsResponse{
				Event: (*string)(&event.Type),
				Pod:   &pod,
			}

			err = out.Send(&resp)
			if err != nil {
				return false, fmt.Errorf("failed to send event: %v", err)
			}

		case <-time.After(watcherTimeout):
			return false, nil
		case <-out.Context().Done():
			logger.Debug("pod watcher server event connection closed")
			watch.Stop()
			return true, nil
		}
	}
}

// ListFunctionss returns isoaltes based on label filter
func (is *functionsServer) ListFunctions(ctx context.Context,
	in *igrpc.ListFunctionsRequest) (*igrpc.ListFunctionsResponse, error) {

	var resp igrpc.ListFunctionsResponse

	logger.Debugf("list functions %v", in.GetAnnotations())

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

// StoreFunctions saves or updates functionss which means creating knative services
// baes on the provided configuration
func (is *functionsServer) CreateFunction(ctx context.Context,
	in *igrpc.CreateFunctionRequest) (*emptypb.Empty, error) {

	logger.Infof("storing functions %s", in.GetInfo().GetName())

	if in.GetInfo() == nil {
		return &empty, fmt.Errorf("info can not be nil")
	}

	// create ksvc service
	err := createKnativeFunction(in.GetInfo())
	if err != nil {
		logger.Errorf("can not create knative service: %v", err)
		return &empty, err
	}

	return &empty, nil

}

func (is *functionsServer) WatchFunctions(in *igrpc.WatchFunctionsRequest, out igrpc.FunctionsService_WatchFunctionsServer) error {

	annotations := in.GetAnnotations()

	l := filterLabels(annotations)
	if len(l) == 0 {
		return fmt.Errorf("request labels are invalid")
	}

	cs, err := fetchServiceAPI()
	if err != nil {
		return fmt.Errorf("could not create fetch client: %v", err)
	}

	labels := labels.Set(l).String()

	for {
		if done, err := is.watcherFunctions(cs, labels, out); err != nil {
			logger.Errorf("function watcher channel failed to restart: %s", err.Error())
			return err
		} else if done {
			// connection has ended
			return nil
		}
		logger.Debugf("function watcher channel has closed, attempting to restart")
		time.Sleep(5 * time.Second)
	}

}

func (is *functionsServer) watcherFunctions(cs *versioned.Clientset, labels string, out igrpc.FunctionsService_WatchFunctionsServer) (bool, error) {
	timeout := int64(watcherTimeout.Seconds())
	watch, err := cs.ServingV1().Services(functionsConfig.Namespace).Watch(context.Background(), metav1.ListOptions{
		LabelSelector:  labels,
		TimeoutSeconds: &timeout,
	})
	if err != nil {
		return false, fmt.Errorf("could start watcher: %v", err)
	}

	for {
		select {
		case event := <-watch.ResultChan():
			s, ok := event.Object.(*v1.Service)
			if !ok {
				return false, nil
			}

			status, conds := statusFromCondition(s.Status.Conditions)
			resp := igrpc.WatchFunctionsResponse{
				Event: (*string)(&event.Type),
				Function: &igrpc.FunctionsInfo{
					Info:        serviceBaseInfo(s),
					Status:      &status,
					Conditions:  conds,
					ServiceName: &s.Name,
				},
			}

			// traffic map
			tm := make(map[string]*int64)
			for i := range s.Status.Traffic {
				tt := s.Status.Traffic[i]
				// sometimes knative routes between the same revisions
				// in this case we just add the percents
				if p, ok := tm[tt.RevisionName]; ok {
					newp := *p + *tt.Percent
					tm[tt.RevisionName] = &newp
				} else {
					tm[tt.RevisionName] = tt.Percent

				}
			}

			resp.Traffic = make([]*igrpc.Traffic, 0)
			for r, p := range tm {
				name := r
				t := new(igrpc.Traffic)
				t.RevisionName = &name
				t.Traffic = p

				// Get Generation
				i, e := strconv.ParseInt(name[strings.LastIndex(name, "-")+1:], 10, 64)
				if e != nil {
					logger.Errorf("could get generation from revision name %v", e)
				}
				t.Generation = &i

				resp.Traffic = append(resp.Traffic, t)
			}

			// Sort by Generation
			sort.Slice(resp.Traffic[:], func(i, j int) bool {
				return *resp.Traffic[i].Generation > *resp.Traffic[j].Generation
			})

			err = out.Send(&resp)
			if err != nil {
				return false, fmt.Errorf("failed to send event: %v", err)
			}

		case <-time.After(watcherTimeout):
			return false, nil
		case <-out.Context().Done():
			logger.Debug("function watcher server event connection closed")
			watch.Stop()
			return true, nil
		}
	}
}

func (is *functionsServer) WatchRevisions(in *igrpc.WatchRevisionsRequest, out igrpc.FunctionsService_WatchRevisionsServer) error {

	var revisionFilter string

	if in.GetServiceName() == "" {
		return fmt.Errorf("service name can not be nil")
	}

	cs, err := fetchServiceAPI()
	if err != nil {
		return fmt.Errorf("could not create fetch client: %v", err)
	}

	l := map[string]string{
		"serving.knative.dev/service": *in.ServiceName,
	}

	if in.GetRevisionName() != "" {
		revisionFilter = *in.RevisionName
	}

	labels := labels.Set(l).String()

	for {
		if done, err := is.watcherRevisions(cs, labels, revisionFilter, out); err != nil {
			logger.Errorf("revision watcher channel failed to restart: %s", err.Error())
			return err
		} else if done {
			// connection has ended
			return nil
		}
		logger.Debugf("revision watcher channel has closed, attempting to restart")
		time.Sleep(5 * time.Second)
	}

}

func (is *functionsServer) watcherRevisions(cs *versioned.Clientset, labels string, revisionFilter string, out igrpc.FunctionsService_WatchRevisionsServer) (bool, error) {
	timeout := int64(watcherTimeout.Seconds())
	watch, err := cs.ServingV1().Revisions(functionsConfig.Namespace).Watch(context.Background(), metav1.ListOptions{
		LabelSelector:  labels,
		TimeoutSeconds: &timeout,
	})
	if err != nil {
		return false, fmt.Errorf("could start watcher: %v", err)
	}

	for {
		select {
		case event := <-watch.ResultChan():
			rev, ok := event.Object.(*v1.Revision)
			if !ok {
				return false, nil
			} else if revisionFilter != "" && rev.Name != revisionFilter {
				continue // skip
			}

			info := &igrpc.Revision{}

			// size and scale
			var sz, scale int32
			var gen int64
			fmt.Sscan(rev.Annotations[ServiceHeaderSize], &sz)
			fmt.Sscan(rev.Annotations["autoscaling.knative.dev/minScale"], &scale)
			fmt.Sscan(rev.Labels[generationHeader], &gen)

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

			// replicas
			info.ActualReplicas = int64(rev.Status.ActualReplicas)
			info.DesiredReplicas = int64(rev.Status.DesiredReplicas)

			// creation date
			var t int64 = rev.CreationTimestamp.Unix()
			info.Created = &t

			resp := igrpc.WatchRevisionsResponse{
				Event:    (*string)(&event.Type),
				Revision: info,
			}

			err = out.Send(&resp)
			if err != nil {
				return false, fmt.Errorf("failed to send event: %v", err)
			}

		case <-time.After(watcherTimeout):
			return false, nil
		case <-out.Context().Done():
			logger.Debug("revision watcher server event connection closed")
			watch.Stop()
			return true, nil
		}
	}
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

	err := trafficKnativeFunctions(in.GetName(), in.GetTraffic())
	if err != nil {
		logger.Errorf("can not set traffic: %v", err)
		return &empty, err
	}

	return &empty, nil

}

func (is *functionsServer) DeleteFunction(ctx context.Context,
	in *igrpc.GetFunctionRequest) (*emptypb.Empty, error) {

	err := deleteKnativeFunctions(in.GetServiceName())
	if err != nil {
		logger.Errorf("can not delete knative service: %v", err)
		return &empty, err
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
	err := updateKnativeFunction(in.GetServiceName(),
		in.GetInfo(), in.GetTrafficPercent())
	if err != nil {
		logger.Errorf("can not update knative service: %v", err)
		return &empty, err
	}

	return &empty, nil
}

func containerFromList(containers []corev1.Container) (string, string) {

	var img, cmd string

	for a := range containers {
		c := containers[a]

		if c.Name == containerUser {
			img = c.Image
			cmd = strings.Join(c.Command, ", ")
		}

	}

	return img, cmd
}

func serviceBaseInfo(s *v1.Service) *igrpc.BaseInfo {

	var sz, scale int32
	fmt.Sscan(s.Annotations[ServiceHeaderSize], &sz)
	fmt.Sscan(s.Annotations[ServiceHeaderScale], &scale)
	n := s.Labels[ServiceHeaderName]
	ns := s.Labels[ServiceHeaderNamespace]
	wf := s.Labels[ServiceHeaderWorkflow]
	img, cmd := containerFromList(s.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers)

	info := &igrpc.BaseInfo{}
	info.Name = &n
	info.Namespace = &ns
	info.Workflow = &wf
	info.Size = &sz
	info.MinScale = &scale
	info.Image = &img
	info.Cmd = &cmd

	return info
}

func filterLabels(annotations map[string]string) map[string]string {

	var (
		setter uint8
	)

	// filter out invalid annotations
	a := make(map[string]string)
	for k, v := range annotations {
		if strings.HasPrefix(k, "direktiv.io/") {
			a[k] = v
		}

		if k == ServiceHeaderName && len(v) > 0 {
			setter = setter | 1
		} else if k == ServiceHeaderWorkflow && len(v) > 0 {
			setter = setter | 2
		} else if k == ServiceHeaderNamespace && len(v) > 0 {
			setter = setter | 4
		}
	}

	var (
		scope string
		ok    bool
	)
	if scope, ok = annotations[ServiceHeaderScope]; !ok {
		logger.Errorf("scope not set for list")
		return make(map[string]string)
	}

	t := invalidType
	switch setter {
	case 7:
		t = serviceType
		if scope != PrefixService {
			t = invalidType
		}
	case 6:
		t = workflowType
		if scope != PrefixWorkflow {
			t = invalidType
		}
	case 5:
		t = namespaceType
		if scope != PrefixNamespace {
			t = invalidType
		}
	case 4:
		t = namespaceType
		if scope != PrefixNamespace {
			t = invalidType
		}
	case 0:
		t = globalType
		if scope != PrefixGlobal {
			t = invalidType
		}
	}

	logger.Debugf("request type: %v", setter)

	// Skip invalid check if only scope and name are given
	if setter != 1 && t == invalidType {
		logger.Errorf("wrong labels for search")
		return make(map[string]string)
	}

	// the search is actually on workflow scope
	if a[ServiceHeaderScope] == PrefixService {
		a[ServiceHeaderScope] = PrefixWorkflow
	}

	return a
}

func listKnativeFunctions(annotations map[string]string) ([]*igrpc.FunctionsInfo, error) {

	var b []*igrpc.FunctionsInfo

	filtered := filterLabels(annotations)
	if len(filtered) == 0 {
		return b, fmt.Errorf("request labels are invalid")
	}

	logger.Debugf("list annotations: %s", labels.Set(filtered).String())

	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return b, err
	}

	lo := metav1.ListOptions{LabelSelector: labels.Set(filtered).String()}
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

	filtered := filterLabels(annotations)
	if len(filtered) == 0 {
		return b, fmt.Errorf("request labels are invalid")
	}

	logger.Debugf("list annotations: %s", labels.Set(filtered).String())

	cs, err := getClientSet()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return b, err
	}

	lo := metav1.ListOptions{LabelSelector: labels.Set(filtered).String()}
	l, err := cs.CoreV1().Pods(functionsConfig.Namespace).List(context.Background(), lo)

	if err != nil {
		logger.Errorf("error getting functions list: %v", err)
		return b, err
	}

	for i := range l.Items {

		pod := l.Items[i]
		sn := pod.Labels["serving.knative.dev/service"]
		sr := pod.Labels["serving.knative.dev/revision"]
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

func metaSpec(net string, min, max int, ns, wf, name, scope string, size int) metav1.ObjectMeta {

	metaSpec := metav1.ObjectMeta{
		Namespace:   functionsConfig.Namespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	metaSpec.Labels["direktiv-app"] = "direktiv"

	if len(net) > 0 {
		metaSpec.Annotations["kubernetes.io/ingress-bandwidth"] = net
		metaSpec.Annotations["kubernetes.io/egress-bandwidth"] = net
	}

	metaSpec.Annotations["autoscaling.knative.dev/minScale"] = fmt.Sprintf("%d", min)
	metaSpec.Annotations["autoscaling.knative.dev/maxScale"] = fmt.Sprintf("%d", max)

	metaSpec.Labels[ServiceHeaderName] = name
	if len(wf) > 0 {
		metaSpec.Labels[ServiceHeaderWorkflow] = wf
	}

	metaSpec.Labels[ServiceHeaderNamespace] = ns
	metaSpec.Labels[ServiceHeaderScope] = scope

	metaSpec.Annotations[ServiceHeaderScale] = fmt.Sprintf("%d", min)
	metaSpec.Annotations[ServiceHeaderSize] = fmt.Sprintf("%d", size)

	return metaSpec

}

func meta(svn, name, ns, wf string, scale, size int, scope string) metav1.ObjectMeta {

	meta := metav1.ObjectMeta{
		Name:        svn,
		Namespace:   functionsConfig.Namespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}

	meta.Labels["networking.knative.dev/visibility"] = "cluster-local"

	meta.Labels[ServiceHeaderName] = name
	if len(wf) > 0 {
		meta.Labels[ServiceHeaderWorkflow] = wf
	}

	meta.Labels[ServiceHeaderNamespace] = ns
	meta.Labels[ServiceHeaderScope] = scope

	meta.Annotations[ServiceHeaderScale] = fmt.Sprintf("%d", scale)
	meta.Annotations[ServiceHeaderSize] = fmt.Sprintf("%d", size)

	// crashes knative sometimes
	meta.Annotations["serving.knative.dev/rolloutDuration"] =
		fmt.Sprintf("%ds", functionsConfig.RolloutDuration)

	return meta
}

func proxyEnvs(withGrpc bool) []corev1.EnvVar {

	proxyEnvs := []corev1.EnvVar{}
	if len(functionsConfig.Proxy.HTTP) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  httpProxy,
			Value: functionsConfig.Proxy.HTTP,
		})
	}
	if len(functionsConfig.Proxy.HTTPS) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  httpsProxy,
			Value: functionsConfig.Proxy.HTTPS,
		})
	}
	if len(functionsConfig.Proxy.No) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  noProxy,
			Value: functionsConfig.Proxy.No,
		})
	}

	// add debug if there is an env
	if len(os.Getenv(util.DirektivDebug)) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivDebug,
			Value: os.Getenv(util.DirektivDebug),
		})
	}

	// disable tcp logging
	proxyEnvs = append(proxyEnvs, corev1.EnvVar{
		Name:  util.DirektivFluentbitTCP,
		Value: "true",
	})

	if withGrpc {

		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivFlowEndpoint,
			Value: util.FlowEndpoint(),
		})

		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivMaxServerRcv,
			Value: fmt.Sprintf("%d", util.GrpcCfg().MaxRcvServer),
		})

		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivMaxClientRcv,
			Value: fmt.Sprintf("%d", util.GrpcCfg().MaxRcvClient),
		})

		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivMaxServerSend,
			Value: fmt.Sprintf("%d", util.GrpcCfg().MaxSendServer),
		})

		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivMaxClientSend,
			Value: fmt.Sprintf("%d", util.GrpcCfg().MaxSendClient),
		})

		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivFlowTLS,
			Value: util.GrpcCfg().FlowTLS,
		})

		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivFlowMTLS,
			Value: util.GrpcCfg().FlowMTLS,
		})

	}

	return proxyEnvs
}

func generateResourceLimits(size int) (corev1.ResourceRequirements, error) {

	var (
		m int
		c float64
	)

	switch size {
	case 1:
		m = functionsConfig.Memory.Medium
		c = functionsConfig.CPU.Medium
	case 2:
		m = functionsConfig.Memory.Large
		c = functionsConfig.CPU.Large
	default:
		m = functionsConfig.Memory.Small
		c = functionsConfig.CPU.Small
	}

	qcpu, err := resource.ParseQuantity(fmt.Sprintf("%v", c))
	if err != nil {
		return corev1.ResourceRequirements{}, err
	}

	qcpuHigh, err := resource.ParseQuantity(fmt.Sprintf("%v", c*2))
	if err != nil {
		return corev1.ResourceRequirements{}, err
	}

	qmem, err := resource.ParseQuantity(fmt.Sprintf("%vM", m))
	if err != nil {
		return corev1.ResourceRequirements{}, err
	}

	qmemHigh, err := resource.ParseQuantity(fmt.Sprintf("%vM", m*2))
	if err != nil {
		return corev1.ResourceRequirements{}, err
	}

	ephemeral, err := resource.ParseQuantity(fmt.Sprintf("%dMi", functionsConfig.Storage))
	if err != nil {
		return corev1.ResourceRequirements{}, err
	}

	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			"cpu":    qcpu,
			"memory": qmem,
		},
		Limits: corev1.ResourceList{
			"cpu":               qcpuHigh,
			"memory":            qmemHigh,
			"ephemeral-storage": ephemeral,
		},
	}, nil

}

func makeContainers(img, cmd string, size int) ([]corev1.Container, error) {

	res, err := generateResourceLimits(size)
	if err != nil {
		logger.Errorf("can not parse requests limits")
		return []corev1.Container{}, err
	}

	// user container
	uc := corev1.Container{
		Name:      containerUser,
		Image:     img,
		Env:       proxyEnvs(false),
		Resources: res,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/mnt/shared",
			},
		},
	}

	if len(cmd) > 0 {
		args, err := shellwords.Parse(cmd)
		if err != nil {
			return []corev1.Container{}, err
		}
		uc.Command = args
	}

	proxy := proxyEnvs(true)

	// append db info
	proxy = append(proxy, corev1.EnvVar{
		Name: util.DBConn,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: functionsConfig.SidecarDb,
				},
				Key: "db",
			},
		},
	})

	vmounts := []corev1.VolumeMount{
		{
			Name:      "workdir",
			MountPath: "/mnt/shared",
		},
	}

	if util.GrpcCfg().FlowTLS != "" && util.GrpcCfg().FlowTLS != "none" {
		tlsVolumeMount := corev1.VolumeMount{}
		tlsVolumeMount.Name = "flowcerts"
		tlsVolumeMount.MountPath = "/etc/direktiv/certs/flow"
		tlsVolumeMount.ReadOnly = true
		vmounts = append(vmounts, tlsVolumeMount)
	}

	// direktiv sidecar
	ds := corev1.Container{
		Name:         containerSidecar,
		Image:        functionsConfig.Sidecar,
		Env:          proxy,
		VolumeMounts: vmounts,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: 8890,
			},
		},
	}

	c := []corev1.Container{uc, ds}

	for i := range functionsConfig.AdditionalContainers {
		container := functionsConfig.AdditionalContainers[i]
		c = append(c, container)
	}

	return c, nil

}

func fetchServiceAPI() (*versioned.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Errorf("error getting api: %v", err)
		return nil, err
	}
	return versioned.NewForConfig(config)
}

// GenerateServiceName generates a knative name based on workflow details
func GenerateServiceName(ns, wf, n string) (string, string, error) {

	h, err := hash.Hash(fmt.Sprintf("%s-%s-%s", ns, wf, n), hash.FormatV2, nil)
	if err != nil {
		return "", "", err
	}

	// get scope and create name
	// workflow
	name := fmt.Sprintf("%s-%d", PrefixWorkflow, h)
	scope := PrefixWorkflow
	if ns == "" {
		// global
		name = fmt.Sprintf("%s-%s", PrefixGlobal, n)
		scope = PrefixGlobal
	} else if wf == "" {
		//namespace
		scope = PrefixNamespace
		name = fmt.Sprintf("%s-%s-%s", PrefixNamespace, ns, n)
	}

	return name, scope, nil

}

func statusFromCondition(conditions []apis.Condition) (string, []*igrpc.Condition) {
	// status and status messages
	status := fmt.Sprintf("%s", corev1.ConditionUnknown)

	var condList []*igrpc.Condition

	for m := range conditions {
		cond := conditions[m]

		if cond.Type == v1.RevisionConditionReady {
			status = fmt.Sprintf("%s", cond.Status)
		}

		ct := string(cond.Type)
		st := string(cond.Status)
		c := &igrpc.Condition{
			Name:    &ct,
			Status:  &st,
			Reason:  &cond.Reason,
			Message: &cond.Message,
		}
		condList = append(condList, c)
	}

	return status, condList

}

func getKnativeFunction(name string) (*igrpc.GetFunctionResponse, error) {

	var (
		revs []*igrpc.Revision
	)

	resp := &igrpc.GetFunctionResponse{}

	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return resp, err
	}

	svc, err := cs.ServingV1().Services(functionsConfig.Namespace).Get(context.Background(),
		name, metav1.GetOptions{})
	if err != nil {
		logger.Errorf("error getting knative service: %v", err)
		return resp, err
	}

	// traffic map
	tm := make(map[string]*int64)

	for i := range svc.Status.Traffic {
		tt := svc.Status.Traffic[i]
		// sometimes knative routes between the same revisions
		// in this case we just add the percents
		if p, ok := tm[tt.RevisionName]; ok {
			newp := *p + *tt.Percent
			tm[tt.RevisionName] = &newp
		} else {
			tm[tt.RevisionName] = tt.Percent
		}
	}

	n := svc.Labels[ServiceHeaderName]
	namespace := svc.Labels[ServiceHeaderNamespace]
	workflow := svc.Labels[ServiceHeaderWorkflow]

	resp.Name = &n
	resp.Namespace = &namespace
	resp.Workflow = &workflow

	rs, err := cs.ServingV1().Revisions(functionsConfig.Namespace).List(context.Background(),
		metav1.ListOptions{LabelSelector: fmt.Sprintf("serving.knative.dev/service=%s", name)})
	if err != nil {
		logger.Errorf("error getting knative service: %v", err)
		return resp, err
	}

	fn := func(rev v1.Revision) *igrpc.Revision {
		info := &igrpc.Revision{}

		// size and scale
		var sz, scale int32
		var gen int64
		fmt.Sscan(rev.Annotations[ServiceHeaderSize], &sz)
		fmt.Sscan(rev.Annotations["autoscaling.knative.dev/minScale"], &scale)
		fmt.Sscan(rev.Labels[generationHeader], &gen)

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

		// replicas
		info.ActualReplicas = int64(rev.Status.ActualReplicas)
		info.DesiredReplicas = int64(rev.Status.DesiredReplicas)

		// creation date
		var t int64 = rev.CreationTimestamp.Unix()
		info.Created = &t

		// set traffic
		var p int64
		if percent, ok := tm[rev.Name]; ok {
			info.Traffic = percent
		} else {
			info.Traffic = &p
		}

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

func deleteKnativeFunctionss(annotations map[string]string) error {

	filtered := filterLabels(annotations)
	if len(filtered) == 0 {
		return fmt.Errorf("request labels are invalid")
	}

	logger.Debugf("delete annotations: %s", labels.Set(filtered).String())

	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return err
	}

	lo := metav1.ListOptions{LabelSelector: labels.Set(filtered).String()}
	return cs.ServingV1().Services(functionsConfig.Namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, lo)

}

func createVolumes() []corev1.Volume {

	volumes := []corev1.Volume{
		{
			Name: "workdir",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	if util.GrpcCfg().FlowTLS != "" && util.GrpcCfg().FlowTLS != "none" {
		tlsVolume := corev1.Volume{
			Name: "flowcerts",
		}
		tlsVolume.Secret = &corev1.SecretVolumeSource{
			SecretName: util.GrpcCfg().FlowTLS,
		}
		volumes = append(volumes, tlsVolume)
	}
	return volumes

}

func updateKnativeFunction(svn string, info *igrpc.BaseInfo, percent int64) error {

	logger.Debugf("update knative function %s", svn)

	containers, err := makeContainers(info.GetImage(), info.GetCmd(),
		int(info.GetSize()))
	if err != nil {
		logger.Errorf("can not update service: %v", err)
		return err
	}

	spec := metav1.ObjectMeta{
		Annotations: make(map[string]string),
	}

	spec.Annotations[ServiceHeaderSize] = fmt.Sprintf("%d", info.GetSize())
	spec.Annotations["serving.knative.dev/rolloutDuration"] =
		fmt.Sprintf("%ds", functionsConfig.RolloutDuration)
	spec.Annotations["autoscaling.knative.dev/minScale"] =
		fmt.Sprintf("%d", info.GetMinScale())

	// adjust traffic for new revision
	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return err
	}

	// get all revisions

	s, err := cs.ServingV1().Services(functionsConfig.Namespace).Get(context.Background(),
		svn, metav1.GetOptions{})
	if err != nil {
		logger.Errorf("error getting knative service: %v", err)
		return err
	}

	useLatest := true

	tr := []v1.TrafficTarget{}
	tt := v1.TrafficTarget{
		LatestRevision: &useLatest,
		Percent:        &percent,
	}
	tr = append(tr, tt)

	for _, trafficInfo := range s.Status.Traffic {
		if trafficInfo.Percent != nil {
			newPercent := math.Round(float64(*trafficInfo.Percent) *
				(100.0 - float64(percent)) / 100.0)
			if newPercent != 0 {
				p := int64(newPercent)
				logger.Debugf("setting existing traffic percent for '%s' to '%d' (was '%d')",
					trafficInfo.RevisionName, p, *trafficInfo.Percent)
				tr = append(tr, v1.TrafficTarget{
					RevisionName: trafficInfo.RevisionName,
					Percent:      &p,
				})
			}
		}
	}

	mountToken := false
	svc := v1.Service{
		Spec: v1.ServiceSpec{
			RouteSpec: v1.RouteSpec{
				Traffic: tr,
			},
			ConfigurationSpec: v1.ConfigurationSpec{
				Template: v1.RevisionTemplateSpec{
					ObjectMeta: spec,
					Spec: v1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							AutomountServiceAccountToken: &mountToken,
							Containers:                   containers,
							Volumes:                      createVolumes(),
						},
					},
				},
			},
		},
	}

	b, err := json.MarshalIndent(svc, "", "    ")
	if err != nil {
		logger.Errorf("error marshalling new services: %v", err)
		return nil
	}

	logger.Debugf("patching service %s", svn)

	// lock for updates and deletes
	l, err := kubeLock(svn, false)
	if err != nil {
		return err
	}
	defer kubeUnlock(l)

	_, err = cs.ServingV1().Services(functionsConfig.Namespace).Patch(context.Background(),
		svn, types.MergePatchType, b, metav1.PatchOptions{})

	if err != nil {
		logger.Errorf("can not patch service %s: %v", svn, err)
		return err
	}

	// remove older revisions
	rs, err := cs.ServingV1().Revisions(functionsConfig.Namespace).List(context.Background(),
		metav1.ListOptions{LabelSelector: fmt.Sprintf("serving.knative.dev/service=%s", svn)})
	if err != nil {
		logger.Errorf("error getting old revisions: %v", err)
		return err
	}

	slice.Sort(rs.Items[:], func(i, j int) bool {
		var gen1, gen2 int64
		fmt.Sscan(rs.Items[i].Labels[generationHeader], &gen1)
		fmt.Sscan(rs.Items[j].Labels[generationHeader], &gen2)
		return gen1 < gen2
	})

	logger.Debugf("removing old revisions for %s (%d)", svn, (len(rs.Items) - functionsConfig.KeepRevisions))

	// delete old revisions
	for i := 0; i < (len(rs.Items) - functionsConfig.KeepRevisions); i++ {
		logger.Debugf("deleting %v", rs.Items[i].Name)
		err := cs.ServingV1().Revisions(functionsConfig.Namespace).Delete(context.Background(), rs.Items[i].Name, metav1.DeleteOptions{})
		if err != nil {
			logger.Errorf("error deleting old revisions: %v", err)
		}
	}

	return nil
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

func createKnativeFunction(info *igrpc.BaseInfo) error {

	var (
		concurrency int64 = 100
		timeoutSec  int64 = 60
	)

	name, scope, err := GenerateServiceName(info.GetNamespace(),
		info.GetWorkflow(), info.GetName())
	if err != nil {
		logger.Errorf("can not create service name: %v", err)
		return err
	}

	l, err := kubeLock(name, false)
	if err != nil {
		return err
	}
	defer kubeUnlock(l)

	logger.Debugf("creating knative service %s", name)

	logger.Debugf("functions namespace %s", functionsConfig.Namespace)

	// check if min scale is not beyond max
	min := int(info.GetMinScale())
	if min > functionsConfig.MaxScale {
		min = functionsConfig.MaxScale
	}

	// TODO: gcp db, pullimagesecrets

	if functionsConfig.Concurrency > 0 {
		concurrency = int64(functionsConfig.Concurrency)
	}

	containers, err := makeContainers(info.GetImage(), info.GetCmd(),
		int(info.GetSize()))
	if err != nil {
		logger.Errorf("can not make containers: %v", err)
		return err
	}

	mountToken := false
	svc := v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "serving.knative.dev/v1",
			Kind:       "Service",
		},
		ObjectMeta: meta(name, info.GetName(),
			info.GetNamespace(), info.GetWorkflow(), min, int(info.GetSize()), scope),
		Spec: v1.ServiceSpec{
			ConfigurationSpec: v1.ConfigurationSpec{
				Template: v1.RevisionTemplateSpec{
					ObjectMeta: metaSpec(functionsConfig.NetShape, min, functionsConfig.MaxScale,
						info.GetNamespace(), info.GetWorkflow(), info.GetName(), scope, int(info.GetSize())),
					Spec: v1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							AutomountServiceAccountToken: &mountToken,
							ImagePullSecrets:             createPullSecrets(info.GetNamespace()),
							ServiceAccountName:           functionsConfig.ServiceAccount,
							Containers:                   containers,
							Volumes:                      createVolumes(),
						},
						ContainerConcurrency: &concurrency,
						TimeoutSeconds:       &timeoutSec,
					},
				},
			},
		},
	}

	if len(functionsConfig.Runtime) > 0 && functionsConfig.Runtime != "default" {
		logger.Debugf("setting runtime class %v", functionsConfig.Runtime)
		svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.RuntimeClassName = &functionsConfig.Runtime
	}

	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return err
	}

	_, err = cs.ServingV1().Services(functionsConfig.Namespace).Create(context.Background(), &svc, metav1.CreateOptions{})
	if err != nil {
		logger.Errorf("error creating knative service: %v", err)
		return err
	}

	return nil
}

func deleteKnativeFunctions(name string) error {

	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return err
	}

	return cs.ServingV1().Services(functionsConfig.Namespace).Delete(context.Background(),
		name, metav1.DeleteOptions{})

}

func trafficKnativeFunctions(name string, tv []*igrpc.TrafficValue) error {

	logger.Debugf("setting traffic for %s", name)

	if len(tv) == 0 {
		return fmt.Errorf("no traffic defined")
	}

	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return err
	}

	r, err := getKnativeFunction(name)
	if err != nil {
		logger.Errorf("error getting service: %v", err)
		return err
	}

	latestRevision := ""

	// should not happen
	if len(r.Revisions) > 0 {
		latestRevision = r.Revisions[0].GetName()
	}

	// check if that rev exists
	revExists := func(revs []*igrpc.Revision, name string) bool {
		for a := range revs {
			if revs[a].GetName() == name {
				return true
			}
		}
		return false
	}

	logger.Debugf("latest revision name: %s", latestRevision)

	tr := []v1.TrafficTarget{}
	for i := range tv {

		if !revExists(r.Revisions, tv[i].GetRevision()) {
			return fmt.Errorf("unknown revision %s", tv[i].GetRevision())
		}

		isLatest := latestRevision == tv[i].GetRevision()

		logger.Debugf("check revision: %s, %v", tv[i].GetRevision(), isLatest)

		if tv[i].GetPercent() > 0 {
			tt := v1.TrafficTarget{
				LatestRevision: &isLatest,
				Percent:        tv[i].Percent,
			}

			logger.Debugf("setting traffic %v: %v",
				tv[i].GetRevision(), tv[i].GetPercent())

			if !isLatest {
				tt.RevisionName = tv[i].GetRevision()
			}

			tr = append(tr, tt)
		}
	}

	var nr v1.Route
	nr.Spec.Traffic = tr

	b, err := json.MarshalIndent(nr, "", "    ")
	if err != nil {
		logger.Errorf("error marshalling new services: %v", err)
	}

	_, err = cs.ServingV1().Services(functionsConfig.Namespace).Patch(context.Background(),
		name, types.MergePatchType, b, metav1.PatchOptions{})

	if err != nil {
		logger.Errorf("error setting traffic: %v", err)
	}

	return err

}

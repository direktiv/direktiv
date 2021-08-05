package isolates

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/bradfitz/slice"
	shellwords "github.com/mattn/go-shellwords"
	hash "github.com/mitchellh/hashstructure/v2"
	log "github.com/sirupsen/logrus"
	igrpc "github.com/vorteil/direktiv/pkg/isolates/grpc"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/apis"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

const (
	httpsProxy = "HTTPS_PROXY"
	httpProxy  = "HTTP_PROXY"
	noProxy    = "NO_PROXY"

	envNS    = "DIREKTIV_KUBERNETES_NAMESPACE"
	envDebug = "DIREKTIV_DEBUG"
	envFlow  = "DIREKTIV_FLOW_ENDPOINT"
	envDB    = "DIREKTIV_DB"

	containerUser    = "direktiv-container"
	containerSidecar = "direktiv-sidecar"

	generationHeader = "serving.knative.dev/configurationGeneration"

	prefixWorkflow  = "w"
	prefixNamespace = "ns"
	prefixGlobal    = "g"
	prefixService   = "s" // unused, only if a one item list is requested
)

const (
	serviceType   = iota
	workflowType  = iota
	namespaceType = iota
	globalType    = iota
	invalidType   = iota
)

var (
	mtx sync.Mutex
)

func conatinerFromList(containers []corev1.Container) (string, string) {

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

func filterLabels(annotations map[string]string) map[string]string {

	var (
		setter uint8
	)

	// filter out invalid annotations
	a := make(map[string]string)
	for k, v := range annotations {
		if strings.HasPrefix(k, "direktiv.io/") && k != ServiceHeaderScope {
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
		log.Errorf("scope not set for list")
		return make(map[string]string)
	}

	t := invalidType
	switch setter {
	case 7:
		t = serviceType
		if scope != prefixService {
			t = invalidType
		}
	case 6:
		t = workflowType
		if scope != prefixWorkflow {
			t = invalidType
		}
	case 4:
		t = namespaceType
		if scope != prefixNamespace {
			t = invalidType
		}
	case 0:
		t = globalType
		if scope != prefixGlobal {
			t = invalidType
		}
	}

	log.Debugf("request type: %v", setter)

	if t == invalidType {
		log.Errorf("wrong labels for search")
		return make(map[string]string)
	}

	return a
}

func listKnativeIsolates(annotations map[string]string) ([]*igrpc.IsolateInfo, error) {

	var b []*igrpc.IsolateInfo

	filtered := filterLabels(annotations)
	if len(filtered) == 0 {
		return b, fmt.Errorf("request labels are invalid")
	}

	log.Debugf("list annotations: %s", labels.Set(filtered).String())

	cs, err := fetchServiceAPI()
	if err != nil {
		log.Errorf("error getting clientset for knative: %v", err)
		return b, err
	}

	lo := metav1.ListOptions{LabelSelector: labels.Set(filtered).String()}
	l, err := cs.ServingV1().Services(ns()).List(context.Background(), lo)

	if err != nil {
		log.Errorf("error getting isolate list: %v", err)
		return b, err
	}

	for i := range l.Items {

		svc := l.Items[i]
		n := svc.Labels[ServiceHeaderName]
		ns := svc.Labels[ServiceHeaderNamespace]
		wf := svc.Labels[ServiceHeaderWorkflow]

		info := &igrpc.BaseInfo{}

		info.Name = &n
		info.Namespace = &ns
		info.Workflow = &wf

		var sz, scale int32
		fmt.Sscan(svc.Annotations[ServiceHeaderSize], &sz)
		fmt.Sscan(svc.Annotations[ServiceHeaderScale], &scale)

		info.Size = &sz
		info.MinScale = &scale

		status, statusMsg := statusFromCondition(svc.Status.Conditions)

		img, cmd := conatinerFromList(svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers)
		info.Image = &img
		info.Cmd = &cmd

		svn := svc.Name

		ii := &igrpc.IsolateInfo{
			Info:          info,
			ServiceName:   &svn,
			Status:        &status,
			StatusMessage: &statusMsg,
		}

		b = append(b, ii)

	}

	return b, nil
}

func metaSpec(net string, min, max int, ns, wf, name, scope string) metav1.ObjectMeta {

	metaSpec := metav1.ObjectMeta{
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

	return metaSpec

}

func meta(svn, name, namespace, ns, wf string, scale, size int, scope string) metav1.ObjectMeta {

	meta := metav1.ObjectMeta{
		Name:        svn,
		Namespace:   namespace,
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
	meta.Annotations["serving.knative.dev/rolloutDuration"] =
		fmt.Sprintf("%ds", isolateConfig.RolloutDuration)

	return meta
}

func proxyEnvs() []corev1.EnvVar {

	proxyEnvs := []corev1.EnvVar{}
	if len(isolateConfig.Proxy.HTTP) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  httpProxy,
			Value: isolateConfig.Proxy.HTTP,
		})
	}
	if len(isolateConfig.Proxy.HTTPS) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  httpsProxy,
			Value: isolateConfig.Proxy.HTTPS,
		})
	}
	if len(isolateConfig.Proxy.No) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  noProxy,
			Value: isolateConfig.Proxy.No,
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
		m = isolateConfig.Memory.Medium
		c = isolateConfig.CPU.Medium
	case 2:
		m = isolateConfig.Memory.Large
		c = isolateConfig.CPU.Large
	default:
		m = isolateConfig.Memory.Small
		c = isolateConfig.CPU.Small
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

	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			"cpu":    qcpu,
			"memory": qmem,
		},
		Limits: corev1.ResourceList{
			"cpu":    qcpuHigh,
			"memory": qmemHigh,
		},
	}, nil

}

func makeContainers(img, cmd string, size int) ([]corev1.Container, error) {

	proxy := proxyEnvs()

	res, err := generateResourceLimits(size)
	if err != nil {
		log.Errorf("can not parse requests limits")
		return []corev1.Container{}, err
	}

	// user container
	uc := corev1.Container{
		Name:      containerUser,
		Image:     img,
		Env:       proxy,
		Resources: res,
	}

	if len(cmd) > 0 {
		args, err := shellwords.Parse(cmd)
		if err != nil {
			return []corev1.Container{}, err
		}
		uc.Command = args
	}

	// add debug if there is an env
	if len(os.Getenv(envDebug)) > 0 {
		proxy = append(proxy, corev1.EnvVar{
			Name:  envDebug,
			Value: "true",
		})
	}

	proxy = append(proxy, corev1.EnvVar{
		Name:  envFlow,
		Value: os.Getenv(envFlow),
	})

	// append db info
	proxy = append(proxy, corev1.EnvVar{
		Name: envDB,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: isolateConfig.SidecarDb,
				},
				Key: "db",
			},
		},
	})

	// direktiv sidecar
	ds := corev1.Container{
		Name:  containerSidecar,
		Image: isolateConfig.Sidecar,
		Env:   proxy,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: 8890,
			},
		},
	}

	c := []corev1.Container{uc, ds}

	return c, nil

}

func fetchServiceAPI() (*versioned.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("error getting api: %v", err)
		return nil, err
	}
	return versioned.NewForConfig(config)
}

// GenerateServiceName generates a knative name based on workflow details
func GenerateServiceName(ns, wf, n string) (string, string, error) {

	log.Debugf("service name: %s %s %s", ns, wf, n)

	h, err := hash.Hash(fmt.Sprintf("%s-%s-%s", ns, wf, n), hash.FormatV2, nil)
	if err != nil {
		return "", "", err
	}

	// get scope and create name
	// workflow
	name := fmt.Sprintf("%s-%d", prefixWorkflow, h)
	scope := "wf"
	if ns == "" {
		// global
		name = fmt.Sprintf("%s-%s", prefixGlobal, n)
		scope = prefixGlobal
	} else if wf == "" {
		//namespace
		scope = prefixNamespace
		name = fmt.Sprintf("%s-%s-%s", prefixNamespace, ns, n)
	}

	return name, scope, nil

}

func ns() string {
	// get namespace to deploy in
	configNamespace := os.Getenv(envNS)
	if len(configNamespace) == 0 {
		configNamespace = "default"
	}

	return configNamespace
}

func statusFromCondition(conditions []apis.Condition) (string, string) {
	// status and status message
	status := fmt.Sprintf("%s", corev1.ConditionUnknown)
	var statusMsg string

	for m := range conditions {
		cond := conditions[m]
		if cond.Type == v1.RevisionConditionReady {
			status = fmt.Sprintf("%s", cond.Status)
		} else if cond.Type == v1.RevisionConditionResourcesAvailable ||
			cond.Type == v1.RevisionConditionContainerHealthy {
			// these types can report errors
			statusMsg = fmt.Sprintf("%s %s", statusMsg, cond.Message)
		}
	}

	return status, strings.TrimSpace(statusMsg)

}

func getKnativeIsolate(name string) (*igrpc.GetIsolateResponse, error) {

	var (
		revs []*igrpc.Revision
	)

	resp := &igrpc.GetIsolateResponse{}

	cs, err := fetchServiceAPI()
	if err != nil {
		log.Errorf("error getting clientset for knative: %v", err)
		return resp, err
	}

	svc, err := cs.ServingV1().Services(ns()).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		log.Errorf("error getting knative service: %v", err)
		return resp, err
	}

	// traffic map
	tm := make(map[string]*int64)

	for i := range svc.Status.Traffic {
		tt := svc.Status.Traffic[i]
		tm[tt.RevisionName] = tt.Percent
	}

	n := svc.Labels[ServiceHeaderName]
	namespace := svc.Labels[ServiceHeaderNamespace]
	workflow := svc.Labels[ServiceHeaderWorkflow]

	resp.Name = &n
	resp.Namespace = &namespace
	resp.Workflow = &workflow

	rs, err := cs.ServingV1().Revisions(ns()).List(context.Background(),
		metav1.ListOptions{LabelSelector: fmt.Sprintf("serving.knative.dev/service=%s", name)})
	if err != nil {
		log.Errorf("error getting knative service: %v", err)
		return resp, err
	}

	fn := func(rev v1.Revision) *igrpc.Revision {
		info := &igrpc.Revision{}

		// size and scale
		var sz, scale int32
		var gen int64
		fmt.Sscan(rev.Annotations[ServiceHeaderSize], &sz)
		fmt.Sscan(rev.Annotations[ServiceHeaderScale], &scale)
		fmt.Sscan(rev.Labels[generationHeader], &gen)
		info.Size = &sz
		info.MinScale = &scale
		info.Generation = &gen

		// set status
		status, statusMsg := statusFromCondition(rev.Status.Conditions)
		info.Status = &status
		info.StatusMessage = &statusMsg

		img, cmd := conatinerFromList(rev.Spec.Containers)
		info.Image = &img
		info.Cmd = &cmd

		// name
		svn := rev.Name
		info.Name = &svn

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

	return resp, nil

}

func deleteKnativeIsolates(annotations map[string]string) error {

	filtered := filterLabels(annotations)
	if len(filtered) == 0 {
		return fmt.Errorf("request labels are invalid")
	}

	log.Debugf("delete annotations: %s", labels.Set(filtered).String())

	cs, err := fetchServiceAPI()
	if err != nil {
		log.Errorf("error getting clientset for knative: %v", err)
		return err
	}

	lo := metav1.ListOptions{LabelSelector: labels.Set(filtered).String()}
	return cs.ServingV1().Services(ns()).DeleteCollection(context.Background(), metav1.DeleteOptions{}, lo)

}

func updateKnativeIsolate(svn string, info *igrpc.BaseInfo) error {

	containers, err := makeContainers(info.GetImage(), info.GetCmd(),
		int(info.GetSize()))
	if err != nil {
		log.Errorf("can not update service: %v", err)
		return err
	}

	spec := metav1.ObjectMeta{
		Annotations: make(map[string]string),
	}

	spec.Annotations["autoscaling.knative.dev/minScale"] =
		fmt.Sprintf("%d", info.GetMinScale())

	svc := v1.Service{
		Spec: v1.ServiceSpec{
			ConfigurationSpec: v1.ConfigurationSpec{
				Template: v1.RevisionTemplateSpec{
					ObjectMeta: spec,
					Spec: v1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: containers,
						},
					},
				},
			},
		},
	}

	b, err := json.MarshalIndent(svc, "", "    ")
	if err != nil {
		log.Errorf("error marshalling new services: %v", err)
		return nil
	}

	cs, err := fetchServiceAPI()
	if err != nil {
		log.Errorf("error getting clientset for knative: %v", err)
		return err
	}

	log.Debugf("patching service %s", svn)

	_, err = cs.ServingV1().Services(ns()).Patch(context.Background(),
		svn, types.MergePatchType, b, metav1.PatchOptions{})

	if err != nil {
		log.Errorf("can not patch service %s: %v", svn, err)
		return err
	}

	// remove older revisions
	rs, err := cs.ServingV1().Revisions(ns()).List(context.Background(),
		metav1.ListOptions{LabelSelector: fmt.Sprintf("serving.knative.dev/service=%s", svn)})
	if err != nil {
		log.Errorf("error getting old revisions: %v", err)
		return err
	}

	slice.Sort(rs.Items[:], func(i, j int) bool {
		var gen1, gen2 int64
		fmt.Sscan(rs.Items[i].Labels[generationHeader], &gen1)
		fmt.Sscan(rs.Items[j].Labels[generationHeader], &gen2)
		return gen1 < gen2
	})

	log.Debugf("removing old revisions for %s (%d)", svn, (len(rs.Items) - isolateConfig.KeepRevisions))

	// delete old revisions
	for i := 0; i < (len(rs.Items) - isolateConfig.KeepRevisions); i++ {
		log.Debugf("deleting %v", rs.Items[i].Name)
		err := cs.ServingV1().Revisions(ns()).Delete(context.Background(), rs.Items[i].Name, metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("error deleting old revisions: %v", err)
		}
	}

	return nil
}

func createKnativeIsolate(info *igrpc.BaseInfo) error {

	var (
		concurrency int64 = 100
		timeoutSec  int64 = 60
	)

	name, scope, err := GenerateServiceName(info.GetNamespace(),
		info.GetWorkflow(), info.GetName())
	if err != nil {
		log.Errorf("can not create service name: %v", err)
		return err
	}

	log.Debugf("creating knative service %s", name)

	// get namespace to deploy in
	configNamespace := ns()

	log.Debugf("isolate namespace %s", configNamespace)

	// check if min scale is not beyond max
	min := int(info.GetMinScale())
	if min > isolateConfig.MaxScale {
		min = isolateConfig.MaxScale
	}

	// TODO: gcp db, pullimagesecrets, Proxy,

	if isolateConfig.Concurrency > 0 {
		concurrency = int64(isolateConfig.Concurrency)
	}

	containers, err := makeContainers(info.GetImage(), info.GetCmd(),
		int(info.GetSize()))
	if err != nil {
		log.Errorf("can not make containers: %v", err)
		return err
	}

	svc := v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "serving.knative.dev/v1",
			Kind:       "Service",
		},
		ObjectMeta: meta(name, info.GetName(), configNamespace,
			info.GetNamespace(), info.GetWorkflow(), min, int(info.GetSize()), scope),
		Spec: v1.ServiceSpec{
			ConfigurationSpec: v1.ConfigurationSpec{
				Template: v1.RevisionTemplateSpec{
					ObjectMeta: metaSpec(isolateConfig.NetShape, min, isolateConfig.MaxScale,
						info.GetNamespace(), info.GetWorkflow(), info.GetName(), scope),
					Spec: v1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: containers,
						},
						ContainerConcurrency: &concurrency,
						TimeoutSeconds:       &timeoutSec,
					},
				},
			},
		},
	}

	if len(isolateConfig.Runtime) > 0 && isolateConfig.Runtime != "default" {
		log.Debugf("setting runtime class %v", isolateConfig.Runtime)
		svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.RuntimeClassName = &isolateConfig.Runtime
	}

	cs, err := fetchServiceAPI()
	if err != nil {
		log.Errorf("error getting clientset for knative: %v", err)
		return err
	}

	_, err = cs.ServingV1().Services(configNamespace).Create(context.Background(), &svc, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("error creating knative service: %v", err)
		return err
	}

	return nil
}

func deleteKnativeIsolate(name string) error {

	cs, err := fetchServiceAPI()
	if err != nil {
		log.Errorf("error getting clientset for knative: %v", err)
		return err
	}

	return cs.ServingV1().Services(ns()).Delete(context.Background(),
		name, metav1.DeleteOptions{})

}

func trafficKnativeIsolate(name string, tv []*igrpc.TrafficValue) error {

	log.Debugf("setting traffic for %s", name)

	if len(tv) == 0 {
		return fmt.Errorf("no traffic defined")
	}

	cs, err := fetchServiceAPI()
	if err != nil {
		log.Errorf("error getting clientset for knative: %v", err)
		return err
	}

	tr := []v1.TrafficTarget{}
	for i := range tv {
		tt := v1.TrafficTarget{
			RevisionName: tv[i].GetRevision(),
			Percent:      tv[i].Percent,
		}
		tr = append(tr, tt)
	}

	var nr v1.Route
	nr.Spec.Traffic = tr

	b, err := json.MarshalIndent(nr, "", "    ")
	if err != nil {
		log.Errorf("error marshalling new services: %v", err)
	}
	fmt.Printf("%s", string(b))

	_, err = cs.ServingV1().Services(ns()).Patch(context.Background(),
		name, types.MergePatchType, b, metav1.PatchOptions{})

	if err != nil {
		log.Errorf("error setting traffic: %v", err)
	}

	return err

}

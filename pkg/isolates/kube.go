package isolates

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

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

var (
	kubeAPIKServiceURL         = "https://kubernetes.default.svc/apis/serving.knative.dev/v1/namespaces/%s/services"
	kubeAPIKServiceURLSpecific = "https://kubernetes.default.svc/apis/serving.knative.dev/v1/namespaces/%s/services/%s"
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

func listKnativeIsolates(annotations map[string]string) ([]*igrpc.IsolateInfo, error) {

	var b []*igrpc.IsolateInfo

	if len(annotations) == 0 {
		return b, fmt.Errorf("annotations empty")
	}

	log.Debugf("list annotations: %s", labels.Set(annotations).String())

	cs, err := fetchServiceAPI()
	if err != nil {
		log.Errorf("error getting clientset for knative: %v", err)
		return b, err
	}

	lo := metav1.ListOptions{LabelSelector: labels.Set(annotations).String()}
	l, err := cs.ServingV1().Services(ns()).List(context.Background(), lo)

	if err != nil {
		log.Errorf("error getting isolate list: %v", err)
		return b, err
	}

	for i := range l.Items {

		log.Debugf("ITEM %+v", l.Items[i])

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

func metaSpec(net string, min, max int, ns, wf, name string) metav1.ObjectMeta {

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

	if len(ns) > 0 {
		metaSpec.Labels[ServiceHeaderNamespace] = ns
	} else {
		metaSpec.Labels[ServiceHeaderNamespace] = "global"
	}

	return metaSpec

}

func meta(svn, name, namespace, ns, wf string, scale, size int) metav1.ObjectMeta {

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

	if len(ns) > 0 {
		meta.Labels[ServiceHeaderNamespace] = ns
	} else {
		meta.Labels[ServiceHeaderNamespace] = "global"
	}

	meta.Annotations[ServiceHeaderScale] = fmt.Sprintf("%d", scale)
	meta.Annotations[ServiceHeaderSize] = fmt.Sprintf("%d", size)

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
func GenerateServiceName(ns, wf, n string) (string, error) {

	log.Debugf("service name: %s %s %s", ns, wf, n)

	h, err := hash.Hash(fmt.Sprintf("%s-%s-%s", ns, wf, n), hash.FormatV2, nil)
	if err != nil {
		return "", err
	}

	// get scope and create name
	// workflow
	name := fmt.Sprintf("w-%d", h)
	if ns == "" {
		// global
		name = fmt.Sprintf("g-%s", n)
	} else if wf == "" {
		//namespace
		name = fmt.Sprintf("ns-%s-%s", ns, n)
	}

	return name, nil

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

	return status, statusMsg

}

func getKnativeIsolate(name string) error {

	var (
		revs []*igrpc.Revision
	)

	resp := &igrpc.GetIsolateResponse{
		Revisions: revs,
	}

	cs, err := fetchServiceAPI()
	if err != nil {
		log.Errorf("error getting clientset for knative: %v", err)
		// return &resp, err
	}

	svc, err := cs.ServingV1().Services(ns()).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		log.Errorf("error getting knative service: %v", err)
		// return &resp, err
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
		// return &resp, err
	}

	fn := func(rev v1.Revision) *igrpc.Revision {
		info := &igrpc.Revision{}

		// size and scale
		var sz, scale int32
		var gen int64
		fmt.Sscan(rev.Annotations[ServiceHeaderSize], &sz)
		fmt.Sscan(rev.Annotations[ServiceHeaderScale], &scale)
		fmt.Sscan(rev.Labels["serving.knative.dev/configurationGeneration"], &gen)
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

		return info
	}

	// get details
	for i := range rs.Items {

		r := rs.Items[i]
		info := fn(r)
		revs = append(revs, info)

	}

	b1, err := json.MarshalIndent(revs, "", "    ")
	if err != nil {
		log.Errorf("error marshalling new services: %v", err)
		return nil
	}
	fmt.Printf("%s", string(b1))

	// log.Debugf("GET %+v", b)

	return nil

}

func deleteIsolates(annotations map[string]string) error {

	if len(annotations) == 0 {
		return fmt.Errorf("annotations empty")
	}

	log.Debugf("delete annotations: %s", labels.Set(annotations).String())

	cs, err := fetchServiceAPI()
	if err != nil {
		log.Errorf("error getting clientset for knative: %v", err)
		return err
	}

	lo := metav1.ListOptions{LabelSelector: labels.Set(annotations).String()}
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

	spec.Annotations["autoscaling.knative.dev/minScale"] = "0"

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

	_, err = cs.ServingV1().Services(ns()).Patch(context.Background(),
		svn, types.MergePatchType, b, metav1.PatchOptions{})

	if err != nil {
		log.Errorf("can not patch service %s: %v", svn, err)
		return err
	}

	// remove older replicas

	return nil
}

func createKnativeIsolate(info *igrpc.BaseInfo) error {

	var (
		concurrency int64 = 100
		timeoutSec  int64 = 60
	)

	name, err := GenerateServiceName(info.GetNamespace(),
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

	// serving.knative.dev/rolloutDuration: "380s"

	svc := v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "serving.knative.dev/v1",
			Kind:       "Service",
		},
		ObjectMeta: meta(name, info.GetName(), configNamespace,
			info.GetNamespace(), info.GetWorkflow(), min, int(info.GetSize())),
		Spec: v1.ServiceSpec{
			ConfigurationSpec: v1.ConfigurationSpec{
				Template: v1.RevisionTemplateSpec{
					ObjectMeta: metaSpec(isolateConfig.NetShape, min, isolateConfig.MaxScale,
						info.GetNamespace(), info.GetWorkflow(), info.GetName()),
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

	mtx.Lock()
	defer mtx.Unlock()

	_, err = cs.ServingV1().Services(configNamespace).Create(context.Background(), &svc, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("error creating knative service: %v", err)
		return err
	}

	return nil
}

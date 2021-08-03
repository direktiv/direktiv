package isolates

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	shellwords "github.com/mattn/go-shellwords"
	hash "github.com/mitchellh/hashstructure/v2"
	log "github.com/sirupsen/logrus"
	igrpc "github.com/vorteil/direktiv/pkg/isolates/grpc"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
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
)

// pullSecrets add/list/remove

func getServices(ns, wf string) {

	log.Infof("getting knative services")
	// v1 "knative.dev/serving/pkg/apis/serving/v1"
}

func metaSpec(net string, min, max int) metav1.ObjectMeta {

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

	return metaSpec

}

func meta(name, ns string, external bool) metav1.ObjectMeta {

	meta := metav1.ObjectMeta{
		Name:      name,
		Namespace: ns,
		Labels:    make(map[string]string),
	}

	if !external {
		meta.Labels["networking.knative.dev/visibility"] = "cluster-local"
	}

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
		Limits: corev1.ResourceList{
			"cpu":    qcpu,
			"memory": qmem,
		},
		Requests: corev1.ResourceList{
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
		Name:      "direktiv-container",
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
		Name:  "direktiv-container",
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

// func updateServiceKube(name, ns, wf string, conf map[string]string, external bool) error {
//
// 	// storeServiceKube(name, ns, wf, conf, external, true)
//
// 	configNamespace := "default"
// 	// netShape := "1M"
// 	// min := "0"
// 	// max := "3"
//
// 	containers := makeContainers("localhost:5000/jens:v2", "")
//
// 	svc := v1.Revision{
// 		TypeMeta: metav1.TypeMeta{
// 			APIVersion: "serving.knative.dev/v1",
// 			Kind:       "Revision",
// 		},
// 		ObjectMeta: meta(name, configNamespace, external),
// 		Spec: v1.RevisionSpec{
// 			PodSpec: corev1.PodSpec{
// 				Containers: containers,
// 			},
// 		},
// 	}
//
// 	b, err := json.MarshalIndent(svc, "", "    ")
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil
// 	}
// 	fmt.Println(string(b))
//
// 	// u := fmt.Sprintf(kubeAPIKServiceURL, os.Getenv(direktivWorkflowNamespace))
//
// 	u := fmt.Sprintf(kubeAPIKServiceURL, "default")
// 	log.Debugf(">>>>>>>>>>>>>>>>>>>>>>> %v", u)
// 	// url := fmt.Sprintf("%s/%s", u, name)
// 	// log.Debugf(">>>>>>>>>>>>>>>>>>>>>>> %v", url)
// 	resp, err := direktiv.SendKuberequest(http.MethodPatch, u, bytes.NewBufferString(string(b)))
//
// 	if err != nil {
// 		log.Errorf("can not create knative service: %v", err)
// 		return err
// 	}
// 	log.Debugf("AD %v", resp)
//
// 	return nil
// }

func fetchServiceAPI() (*versioned.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("ERRORRORORO %v", err)
		return nil, err
	}

	return versioned.NewForConfig(config)

}

func generateServiceName(info *igrpc.BaseInfo) (string, error) {

	h, err := hash.Hash(fmt.Sprintf("%s-%s-%s", info.GetNamespace(),
		info.GetWorkflow(), info.GetName()), hash.FormatV2, nil)
	if err != nil {
		return "", err
	}

	// get scope and create name
	// workflow
	name := fmt.Sprintf("w-%d", h)
	if info.GetNamespace() == "" {
		// global
		name = fmt.Sprintf("g-%s", info.GetName())
	} else if info.GetWorkflow() == "" {
		//namespace
		name = fmt.Sprintf("ns-%s-%s", info.GetNamespace(), info.GetName())
	}

	return name, nil

}

func createKnativeIsolate(info *igrpc.BaseInfo, conf *igrpc.Config, external bool) error {

	var (
		concurrency int64 = 100
		timeoutSec  int64 = 60
	)

	if external {
		return fmt.Errorf("external services not supported")
	}

	name, err := generateServiceName(info)
	if err != nil {
		log.Errorf("can not create service name: %v", err)
		return err
	}

	log.Debugf("creating knative service %s", name)

	// get namespace to deploy in
	configNamespace := os.Getenv(envNS)
	if len(configNamespace) == 0 {
		configNamespace = "default"
	}

	log.Debugf("isolate namespace %s", configNamespace)

	// check if min scale is not beyond max
	min := int(conf.GetMinScale())
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
		ObjectMeta: meta(name, configNamespace, external),
		Spec: v1.ServiceSpec{
			ConfigurationSpec: v1.ConfigurationSpec{
				Template: v1.RevisionTemplateSpec{
					ObjectMeta: metaSpec(isolateConfig.NetShape, min, isolateConfig.MaxScale),
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

	if len(isolateConfig.Runtime) > 0 {
		log.Debugf("setting runtime class %v", isolateConfig.Runtime)
		svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.RuntimeClassName = &isolateConfig.Runtime
	}

	b, err := json.MarshalIndent(svc, "", "    ")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(string(b))

	// u := fmt.Sprintf(kubeAPIKServiceURL, os.Getenv(direktivWorkflowNamespace))
	// u := fmt.Sprintf(kubeAPIKServiceURL, "default")

	// if patch {
	//
	// 	u := fmt.Sprintf(kubeAPIKServiceURL, "default")
	// 	log.Debugf(">>>>>>>>>>>>>>>>>>>>>>> %v", u)
	// 	url := fmt.Sprintf("%s/%s", u, name)
	// 	log.Debugf(">>>>>>>>>>>>>>>>>>>>>>> %v", url)
	// 	resp, err := direktiv.SendKuberequest(http.MethodPatch, url, bytes.NewBufferString(string(b)))
	// 	if err != nil {
	// 		log.Errorf("can not create knative service: %v", err)
	// 		return err
	// 	}
	// 	log.Debugf("PATCH %v", resp)
	// } else {metav1.CreateOptions

	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	log.Errorf("ERRORRORORO %v", err)
	// 	return err
	// }
	// cs, err := versioned.NewForConfig(config)
	// if err != nil {
	// 	log.Errorf("ERRORRORORO2 %v", err)
	// 	return err
	// }
	// //
	// // Create(ctx context.Context, service *v1.Service, opts metav1.CreateOptions) (*v1.Service, error)

	if true {
		return nil
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

	//
	// svc2 := v1.Service{
	// 	ObjectMeta: meta(name, configNamespace, external),
	// }
	//
	// log.Infof("1 %v", svc2.Spec)
	// log.Infof("2%v", svc2.Spec.ConfigurationSpec)
	// log.Infof("3 %v", svc2.Spec.ConfigurationSpec.Template)
	// log.Infof("4 %v", svc2.Spec.ConfigurationSpec.Template.Spec.PodSpec)
	//
	// svc2.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers = makeContainers("localhost:5000/jens:v1", "")
	//
	// svc2.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0].Image = "localhost:5000/jens:v1"
	// svc2.Spec.ConfigurationSpec.Template.ObjectMeta.Annotations = make(map[string]string)
	// svc2.Spec.ConfigurationSpec.Template.ObjectMeta.Annotations["autoscaling.knative.dev/minScale"] = "1"
	//
	// log.Info("???????????????????????????")
	// b2, err := json.MarshalIndent(svc2, "", "    ")
	// if err != nil {
	// 	log.Errorf("ERRORRORORO4 %v", err)
	// 	return err
	// }
	//
	// fmt.Println(string(b2))
	// svc3, err := cs.ServingV1().Services("default").Patch(context.Background(), name, types.MergePatchType, b2, metav1.PatchOptions{})
	// if err != nil {
	// 	log.Errorf("ERRORRORORO4 %v", err)
	// 	return err
	// }
	//
	// log.Infof("HH %v", svc3)
	//
	// // svc.ObjectMeta.ResourceVersion = sc.ObjectMeta.ResourceVersion
	//
	// // Create(ctx context.Context, service *v1.Service, opts metav1.CreateOptions) (*v1.Service, error)
	// // sc, err = cs.ServingV1().Services("default").Update(context.Background(), &svc, metav1.UpdateOptions{})
	// // if err != nil {
	// // 	log.Errorf("ERRORRORORO3 %v", err)
	// // 	return err
	// // }
	// // log.Debugf(">>123 %+v", sc)
	//
	// // cs.ServingV1().Revisions("default")
	//
	// l, err := cs.ServingV1().Revisions("default").List(context.Background(), metav1.ListOptions{})
	// if err != nil {
	// 	log.Errorf("ERRORRORORO3 %v", err)
	// 	return err
	// }
	//
	// log.Debugf(">> %+v", l)
	// resp, err := direktiv.SendKuberequest(http.MethodPost, u, bytes.NewBufferString(string(b)))
	// if err != nil {
	// 	log.Errorf("can not create knative service: %v", err)
	// 	return err
	// }
	// log.Debugf("AD %v", resp)
	// }
	// url := fmt.Sprintf("%s/%s", u, fmt.Sprintf("%s-%s", namespace, ah))

	return nil
}

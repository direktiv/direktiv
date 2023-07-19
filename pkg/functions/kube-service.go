package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/mattn/go-shellwords"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	httpsProxy = "HTTPS_PROXY"
	httpProxy  = "HTTP_PROXY"
	noProxy    = "NO_PROXY"

	containerUser    = "direktiv-container"
	containerSidecar = "direktiv-sidecar"
)

// Headers for knative services.
const (
	// Direktiv Headers.
	ServiceHeaderName          = "direktiv.io/name"
	ServiceHeaderNamespaceID   = "direktiv.io/namespace-id"
	ServiceHeaderNamespaceName = "direktiv.io/namespace-name"
	ServiceHeaderWorkflowID    = "direktiv.io/workflow-id"
	ServiceHeaderPath          = "direktiv.io/workflow-name"
	ServiceHeaderRevision      = "direktiv.io/revision"
	ServiceHeaderSize          = "direktiv.io/size"
	ServiceHeaderScale         = "direktiv.io/scale"
	ServiceTemplateGeneration  = "direktiv.io/templateGeneration"
	ServiceHeaderScope         = "direktiv.io/scope"

	// Serving Headers.
	ServiceKnativeHeaderName            = "serving.knative.dev/service"
	ServiceKnativeHeaderConfiguration   = "serving.knative.dev/configuration"
	ServiceKnativeHeaderGeneration      = "serving.knative.dev/configurationGeneration"
	ServiceKnativeHeaderRevision        = "serving.knative.dev/revision"
	ServiceKnativeHeaderRolloutDuration = "serving.knative.dev/rolloutDuration"
)

func createKnativeFunction(info *igrpc.FunctionsBaseInfo) (*v1.Service, error) {
	name, scope, hash := GenerateServiceName(info)

	l, err := locksmgr.lock(name, false)
	if err != nil {
		return nil, err
	}
	defer locksmgr.unlock(name, l)

	logger.Debugf("creating knative service %s in %s", name, functionsConfig.Namespace)

	// check if min scale is not beyond max
	min := int(info.GetMinScale())
	if min > functionsConfig.MaxScale {
		min = functionsConfig.MaxScale
	}

	containers, err := makeContainers(info.GetImage(), info.GetCmd(),
		int(info.GetSize()), info.GetEnvs())
	if err != nil {
		logger.Errorf("can not make containers: %v", err)
		return nil, err
	}

	n := functionsConfig.knativeAffinity.DeepCopy()
	reqAffinity := n.RequiredDuringSchedulingIgnoredDuringExecution
	if reqAffinity != nil {
		terms := &reqAffinity.NodeSelectorTerms
		if len(*terms) > 0 {
			expressions := &(*terms)[0].MatchExpressions
			if len(*expressions) > 0 {
				expression := &(*expressions)[0]
				if expression.Key == "direktiv.io/namespace" {
					expression.Operator = corev1.NodeSelectorOpIn
					expression.Values = []string{*info.NamespaceName}
				}
			}
		}
	}

	svc := v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "serving.knative.dev/v1",
			Kind:       "Service",
		},
		ObjectMeta: generateServiceMeta(name, scope, hash, min, info),
		Spec: v1.ServiceSpec{
			ConfigurationSpec: v1.ConfigurationSpec{
				Template: v1.RevisionTemplateSpec{
					ObjectMeta: generatePodMeta(name, scope, hash, min, info),
					Spec: v1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							ServiceAccountName: functionsConfig.ServiceAccount,
							Containers:         containers,
							Volumes:            createVolumes(),
							Affinity: &corev1.Affinity{
								NodeAffinity: n,
							},
						},
					},
				},
			},
		},
	}

	// Set Registry Secrets
	secrets := createPullSecrets(info.GetNamespaceName())
	svc.Spec.ConfigurationSpec.Template.Spec.ImagePullSecrets = secrets
	svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.ImagePullSecrets = secrets

	if len(functionsConfig.Runtime) > 0 && functionsConfig.Runtime != "default" {
		logger.Debugf("setting runtime class %v", functionsConfig.Runtime)
		svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.RuntimeClassName = &functionsConfig.Runtime
	}

	cs, err := fetchServiceAPI()
	if err != nil {
		logger.Errorf("error getting clientset for knative: %v", err)
		return nil, err
	}

	newSvc, err := cs.ServingV1().Services(functionsConfig.Namespace).Create(context.Background(), &svc, metav1.CreateOptions{})
	if err != nil {
		logger.Errorf("error creating knative service: %v", err)
		return nil, err
	}

	return newSvc, nil
}

func trimRevisionSuffix(s string) string {
	if i := strings.LastIndex(s, ":"); i > 0 {
		s = s[:i]
	}

	return s
}

func marshal(x interface{}) string {
	data, _ := json.MarshalIndent(x, "", "  ")
	return string(data)
}

func generateServiceMeta(svn, scope, hash string, size int, info *igrpc.FunctionsBaseInfo) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:        svn,
		Namespace:   functionsConfig.Namespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}

	meta.Labels["networking.knative.dev/visibility"] = "cluster-local"

	meta.Labels[ServiceHeaderName] = SanitizeLabel(info.GetName())
	if len(info.GetWorkflow()) > 0 {
		meta.Labels[ServiceHeaderWorkflowID] = SanitizeLabel(info.GetWorkflow())
		meta.Labels[ServiceHeaderPath] = trimRevisionSuffix(SanitizeLabel(filepath.Base(info.GetPath())))
		meta.Labels[ServiceHeaderRevision] = SanitizeLabel(hash)
	}

	meta.Labels[ServiceHeaderNamespaceID] = SanitizeLabel(info.GetNamespace())
	meta.Labels[ServiceHeaderNamespaceName] = SanitizeLabel(info.GetNamespaceName())
	meta.Labels[ServiceHeaderScope] = scope

	meta.Annotations[ServiceHeaderScale] = fmt.Sprintf("%d", int(info.GetMinScale()))
	meta.Annotations[ServiceHeaderSize] = fmt.Sprintf("%d", size)

	meta.Annotations["networking.knative.dev/ingress.class"] = functionsConfig.IngressClass

	return meta
}

func generatePodMeta(svn, scope, hash string, size int, info *igrpc.FunctionsBaseInfo) metav1.ObjectMeta {
	metaSpec := metav1.ObjectMeta{
		Namespace:   functionsConfig.Namespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	metaSpec.Labels["direktiv-app"] = "direktiv"

	if len(functionsConfig.NetShape) > 0 {
		metaSpec.Annotations["kubernetes.io/ingress-bandwidth"] = functionsConfig.NetShape
		metaSpec.Annotations["kubernetes.io/egress-bandwidth"] = functionsConfig.NetShape
	}

	metaSpec.Annotations["autoscaling.knative.dev/minScale"] = fmt.Sprintf("%d", info.GetMinScale())
	metaSpec.Annotations["autoscaling.knative.dev/maxScale"] = fmt.Sprintf("%d", functionsConfig.MaxScale)

	metaSpec.Labels[ServiceHeaderName] = SanitizeLabel(info.GetName())
	if len(info.GetWorkflow()) > 0 {
		metaSpec.Labels[ServiceHeaderWorkflowID] = SanitizeLabel(info.GetWorkflow())
		metaSpec.Labels[ServiceHeaderPath] = trimRevisionSuffix(SanitizeLabel(filepath.Base(info.GetPath())))
		metaSpec.Labels[ServiceHeaderRevision] = SanitizeLabel(hash)
	}

	metaSpec.Labels[ServiceHeaderNamespaceID] = SanitizeLabel(info.GetNamespace())
	metaSpec.Labels[ServiceHeaderNamespaceName] = SanitizeLabel(info.GetNamespaceName())
	metaSpec.Labels[ServiceHeaderScope] = scope

	metaSpec.Annotations[ServiceHeaderScale] = fmt.Sprintf("%d", int(info.GetSize()))
	metaSpec.Annotations[ServiceHeaderSize] = fmt.Sprintf("%d", size)

	return metaSpec
}

func makeContainers(img, cmd string, size int,
	envs map[string]string,
) ([]corev1.Container, error) {
	res, err := generateResourceLimits(size)
	if err != nil {
		logger.Errorf("can not parse requests limits")
		return []corev1.Container{}, err
	}

	logger.Debugf("resource limits: %+v", res)

	// user container
	uc := corev1.Container{
		Name:      containerUser,
		Image:     img,
		Env:       proxyEnvs(false, envs),
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

	proxy := proxyEnvs(true, make(map[string]string))

	vmounts := []corev1.VolumeMount{
		{
			Name:      "workdir",
			MountPath: "/mnt/shared",
		},
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

	c = append(c, functionsConfig.extraContainers...)

	return c, nil
}

func proxyEnvs(withGrpc bool, envs map[string]string) []corev1.EnvVar {
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

	proxyEnvs = append(proxyEnvs, corev1.EnvVar{
		Name:  util.DirektivOpentelemetry,
		Value: functionsConfig.OpenTelemetryBackend,
	})

	proxyEnvs = append(proxyEnvs, corev1.EnvVar{
		Name:  util.DirektivLogJSON,
		Value: functionsConfig.Logging,
	})

	if withGrpc {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivFlowEndpoint,
			Value: functionsConfig.FlowService,
		})
	}

	for k, v := range envs {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	proxyEnvs = append(proxyEnvs, corev1.EnvVar{
		Name:  "DIREKTIV_APP",
		Value: "sidecar",
	})

	return proxyEnvs
}

func generateResourceLimits(size int) (corev1.ResourceRequirements, error) {
	var (
		m int
		c string
		d int
	)

	baseCPU, _ := resource.ParseQuantity("0.1")
	baseMem, _ := resource.ParseQuantity("64M")
	baseDisk, _ := resource.ParseQuantity("64M")

	switch size {
	case 1:
		m = functionsConfig.Memory.Medium
		c = functionsConfig.CPU.Medium
		d = functionsConfig.Disk.Medium
	case 2:
		m = functionsConfig.Memory.Large
		c = functionsConfig.CPU.Large
		d = functionsConfig.Disk.Large
	default:
		m = functionsConfig.Memory.Small
		c = functionsConfig.CPU.Small
		d = functionsConfig.Disk.Small
	}

	// just in case for old helm charts
	if d == 0 {
		d = 4096
	}

	ephemeralHigh, err := resource.ParseQuantity(fmt.Sprintf("%dM", d))
	if err != nil {
		return corev1.ResourceRequirements{}, err
	}

	rl := corev1.ResourceList{
		"ephemeral-storage": ephemeralHigh,
	}

	if m != 0 {
		qmem, err := resource.ParseQuantity(fmt.Sprintf("%dM", m))
		if err != nil {
			return corev1.ResourceRequirements{}, err
		}
		rl["memory"] = qmem
	}

	if c != "" {
		qcpu, err := resource.ParseQuantity(c)
		if err != nil {
			return corev1.ResourceRequirements{}, err
		}
		rl["cpu"] = qcpu
	}

	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			"cpu":               baseCPU,
			"memory":            baseMem,
			"ephemeral-storage": baseDisk,
		},
		Limits: rl,
	}, nil
}

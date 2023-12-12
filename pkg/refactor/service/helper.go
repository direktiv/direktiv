package service

import (
	"fmt"
	"strconv"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/mattn/go-shellwords"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func buildService(c *core.Config, sv *core.ServiceConfig) (*servingv1.Service, error) {
	containers, err := buildContainers(c, sv)
	if err != nil {
		return nil, err
	}

	// nolint
	//n := functionsConfig.knativeAffinity.DeepCopy()
	//reqAffinity := n.RequiredDuringSchedulingIgnoredDuringExecution
	//if reqAffinity != nil {
	//	terms := &reqAffinity.NodeSelectorTerms
	//	if len(*terms) > 0 {
	//		expressions := &(*terms)[0].MatchExpressions
	//		if len(*expressions) > 0 {
	//			expression := &(*expressions)[0]
	//			if expression.Key == "direktiv.io/namespace" {
	//				expression.Operator = corev1.NodeSelectorOpIn
	//				expression.Values = []string{*info.NamespaceName}
	//			}
	//		}
	//	}
	//}

	svc := &servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "serving.knative.dev/v1",
			Kind:       "Service",
		},
		ObjectMeta: buildServiceMeta(c, sv),
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: buildPodMeta(c, sv),
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							ServiceAccountName: c.KnativeServiceAccount,
							Containers:         containers,
							Volumes:            buildVolumes(c, sv),
							Affinity:           &corev1.Affinity{
								// NodeAffinity: n,
							},
						},
					},
				},
			},
		},
	}

	// nolint
	// Set Registry Secrets
	//secrets := createPullSecrets(info.GetNamespaceName())
	//svc.Spec.ConfigurationSpec.Template.Spec.ImagePullSecrets = secrets
	//svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.ImagePullSecrets = secrets
	//if len(functionsConfig.Runtime) > 0 && functionsConfig.Runtime != "default" {
	//	logger.Debugf("setting runtime class %v", functionsConfig.Runtime)
	//	svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.RuntimeClassName = &functionsConfig.Runtime
	//}

	return svc, nil
}

func buildServiceMeta(c *core.Config, sv *core.ServiceConfig) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:        sv.GetID(),
		Namespace:   c.KnativeNamespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}

	meta.Annotations["direktiv.io/inputHash"] = sv.GetValueHash()
	meta.Labels["networking.knative.dev/visibility"] = "cluster-local"
	meta.Annotations["networking.knative.dev/ingress.class"] = c.KnativeIngressClass

	return meta
}

func buildPodMeta(c *core.Config, sv *core.ServiceConfig) metav1.ObjectMeta {
	metaSpec := metav1.ObjectMeta{
		Namespace:   c.KnativeNamespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	metaSpec.Labels["direktiv-app"] = "direktiv"

	metaSpec.Annotations["autoscaling.knative.dev/minScale"] = strconv.Itoa(sv.Scale)
	metaSpec.Annotations["autoscaling.knative.dev/maxScale"] = strconv.Itoa(c.KnativeMaxScale)
	metaSpec.Annotations["autoscaling.knative.dev/minScale"] = strconv.Itoa(sv.Scale)

	if len(c.KnativeNetShape) > 0 {
		metaSpec.Annotations["kubernetes.io/ingress-bandwidth"] = c.KnativeNetShape
		metaSpec.Annotations["kubernetes.io/egress-bandwidth"] = c.KnativeNetShape
	}

	return metaSpec
}

// nolint
func buildVolumes(c *core.Config, sv *core.ServiceConfig) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "workdir",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	// volumes = append(volumes, c.extraVolumes...)

	return volumes
}

func buildContainers(c *core.Config, sv *core.ServiceConfig) ([]corev1.Container, error) {
	// set resource limits.
	rl, err := buildResourceLimits(c, sv)
	if err != nil {
		return nil, err
	}

	// user container
	uc := corev1.Container{
		Name:      containerUser,
		Image:     sv.Image,
		Env:       buildEnvVars(false, c, sv),
		Resources: *rl,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/mnt/shared",
			},
		},
	}

	if len(sv.CMD) > 0 {
		args, err := shellwords.Parse(sv.CMD)
		if err != nil {
			return []corev1.Container{}, err
		}
		uc.Command = args
	}

	vMounts := []corev1.VolumeMount{
		{
			Name:      "workdir",
			MountPath: "/mnt/shared",
		},
	}

	// direktiv sidecar
	sc := corev1.Container{
		Name:         containerSidecar,
		Image:        c.KnativeSidecar,
		Env:          buildEnvVars(true, c, sv),
		VolumeMounts: vMounts,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: containerSidecarPort,
			},
		},
	}

	return []corev1.Container{uc, sc}, nil
}

// nolint
func buildResourceLimits(cf *core.Config, sv *core.ServiceConfig) (*corev1.ResourceRequirements, error) {
	var (
		m int
		c string
		d int
	)

	switch sv.Size {
	case "small":
		m = cf.KnativeSizeMemorySmall
		c = cf.KnativeSizeCPUSmall
		d = cf.KnativeSizeDiskSmall
	case "medium":
		m = cf.KnativeSizeMemoryMedium
		c = cf.KnativeSizeCPUMedium
		d = cf.KnativeSizeDiskMedium
	case "large":
		m = cf.KnativeSizeMemoryLarge
		c = cf.KnativeSizeCPULarge
		d = cf.KnativeSizeDiskLarge
	default:
		return nil, fmt.Errorf("service size: '%s' is invalid, expected value: ['small', 'medium', 'large']", sv.Size)
	}

	ephemeralHigh, err := resource.ParseQuantity(fmt.Sprintf("%dM", d))
	if err != nil {
		return nil, err
	}

	rl := corev1.ResourceList{
		"ephemeral-storage": ephemeralHigh,
	}

	if m != 0 {
		qmem, err := resource.ParseQuantity(fmt.Sprintf("%dM", m))
		if err != nil {
			return nil, err
		}
		rl["memory"] = qmem
	}

	if c != "" {
		qcpu, err := resource.ParseQuantity(c)
		if err != nil {
			return nil, err
		}
		rl["cpu"] = qcpu
	}

	baseCPU, _ := resource.ParseQuantity("0.1")
	baseMem, _ := resource.ParseQuantity("64M")
	baseDisk, _ := resource.ParseQuantity("64M")

	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			"cpu":               baseCPU,
			"memory":            baseMem,
			"ephemeral-storage": baseDisk,
		},
		Limits: rl,
	}, nil
}

// nolint
func buildEnvVars(withGrpc bool, c *core.Config, sv *core.ServiceConfig) []corev1.EnvVar {
	proxyEnvs := []corev1.EnvVar{}

	if len(c.KnativeProxyHTTP) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivProxyHTTP,
			Value: c.KnativeProxyHTTP,
		})
	}

	if len(c.KnativeProxyHTTPS) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivProxyHTTPS,
			Value: c.KnativeProxyHTTPS,
		})
	}

	if len(c.KnativeProxyNo) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivProxyNO,
			Value: c.KnativeProxyNo,
		})
	}

	// add debug if there is an env
	if c.LogDebug {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivDebug,
			Value: "true",
		})
	}

	proxyEnvs = append(proxyEnvs, corev1.EnvVar{
		Name:  util.DirektivOpentelemetry,
		Value: c.OpenTelemetry,
	})

	proxyEnvs = append(proxyEnvs, corev1.EnvVar{
		Name:  util.DirektivLogFormat,
		Value: c.LogFormat,
	})

	if withGrpc {
		namespace := c.DirektivNamespace

		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivFlowEndpoint,
			Value: fmt.Sprintf("direktiv-flow.%s", namespace),
		})

		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  "DIREKTIV_APP",
			Value: "sidecar",
		})
	} else {
		for _, v := range sv.Envs {
			proxyEnvs = append(proxyEnvs, corev1.EnvVar{
				Name:  v.Name,
				Value: v.Value,
			})
		}
	}

	return proxyEnvs
}

package service

import (
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/mattn/go-shellwords"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func buildService(c *core.Config, cfg *core.ServiceConfig) (*servingv1.Service, error) {
	containers, err := buildContainers(c, cfg)
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
		ObjectMeta: buildServiceMeta(c, cfg),
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: buildPodMeta(c, cfg),
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							ServiceAccountName: c.KnativeServiceAccount,
							Containers:         containers,
							Volumes:            buildVolumes(c, cfg),
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

func buildServiceMeta(c *core.Config, cfg *core.ServiceConfig) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:        cfg.GetID(),
		Namespace:   c.KnativeNamespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}

	meta.Annotations["direktiv.io/inputHash"] = cfg.GetValueHash()
	meta.Labels["networking.knative.dev/visibility"] = "cluster-local"
	meta.Annotations["networking.knative.dev/ingress.class"] = c.KnativeIngressClass

	return meta
}

func buildPodMeta(c *core.Config, cfg *core.ServiceConfig) metav1.ObjectMeta {
	metaSpec := metav1.ObjectMeta{
		Namespace:   c.KnativeNamespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	metaSpec.Labels["direktiv-app"] = "direktiv"

	metaSpec.Annotations["autoscaling.knative.dev/minScale"] = fmt.Sprintf("%d", cfg.Scale)
	metaSpec.Annotations["autoscaling.knative.dev/maxScale"] = fmt.Sprintf("%d", c.KnativeMaxScale)
	metaSpec.Annotations["autoscaling.knative.dev/minScale"] = fmt.Sprintf("%d", cfg.Scale)

	metaSpec.Annotations["kubernetes.io/egress-bandwidth"] = "10M"
	metaSpec.Annotations["kubernetes.io/ingress-bandwidth"] = "10M"

	return metaSpec
}

// nolint
func buildVolumes(c *core.Config, cfg *core.ServiceConfig) []corev1.Volume {
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

func buildContainers(c *core.Config, cfg *core.ServiceConfig) ([]corev1.Container, error) {
	// TODO: yassir, we appear to have lost envs
	envs := make(map[string]string)

	// set resource limits.
	rl, err := buildResourceLimits(c, cfg)
	if err != nil {
		return nil, err
	}

	// user container
	uc := corev1.Container{
		Name:      containerUser,
		Image:     cfg.Image,
		Env:       buildEnvVars(false, c, cfg, envs),
		Resources: rl,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/mnt/shared",
			},
		},
	}

	if len(cfg.CMD) > 0 {
		args, err := shellwords.Parse(cfg.CMD)
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
		Env:          buildEnvVars(true, c, cfg, envs),
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
func buildResourceLimits(c *core.Config, cfg *core.ServiceConfig) (corev1.ResourceRequirements, error) {
	return corev1.ResourceRequirements{}, nil
}

// nolint
func buildEnvVars(withGrpc bool, c *core.Config, cfg *core.ServiceConfig, envs map[string]string) []corev1.EnvVar {
	proxyEnvs := []corev1.EnvVar{}

	// TODO: yassir
	// if len(functionsConfig.Proxy.HTTP) > 0 {
	// 	proxyEnvs = append(proxyEnvs, corev1.EnvVar{
	// 		Name:  httpProxy,
	// 		Value: functionsConfig.Proxy.HTTP,
	// 	})
	// }

	// TODO: yassir
	// if len(functionsConfig.Proxy.HTTPS) > 0 {
	// 	proxyEnvs = append(proxyEnvs, corev1.EnvVar{
	// 		Name:  httpsProxy,
	// 		Value: functionsConfig.Proxy.HTTPS,
	// 	})
	// }

	// TODO: yassir
	// if len(functionsConfig.Proxy.No) > 0 {
	// 	proxyEnvs = append(proxyEnvs, corev1.EnvVar{
	// 		Name:  noProxy,
	// 		Value: functionsConfig.Proxy.No,
	// 	})
	// }

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
		Name:  util.DirektivLogJSON,
		Value: c.LogFormat,
	})

	if withGrpc {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  util.DirektivFlowEndpoint,
			Value: "direktiv-flow.direktiv", // TODO: alan
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

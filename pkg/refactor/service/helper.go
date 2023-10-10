// nolint
package service

import (
	"fmt"

	"github.com/mattn/go-shellwords"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func validateConfig(c *ClientConfig) (*ClientConfig, error) {
	if c.MaxScale > 9 || c.MaxScale < 1 {
		c.MaxScale = 5
	}

	return c, nil
}

func buildService(c *ClientConfig, cfg *Config) (*servingv1.Service, error) {
	containers, err := buildContainers(c, cfg)
	if err != nil {
		return nil, err
	}

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
							ServiceAccountName: c.ServiceAccount,
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

func buildServiceMeta(c *ClientConfig, cfg *Config) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:        cfg.getID(),
		Namespace:   c.Namespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}

	meta.Annotations["direktiv.io/inputHash"] = cfg.getValueHash()
	meta.Labels["networking.knative.dev/visibility"] = "cluster-local"
	meta.Annotations["networking.knative.dev/ingress.class"] = c.IngressClass

	return meta
}

func buildPodMeta(c *ClientConfig, cfg *Config) metav1.ObjectMeta {
	metaSpec := metav1.ObjectMeta{
		Namespace:   c.Namespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	metaSpec.Labels["direktiv-app"] = "direktiv"

	metaSpec.Annotations["autoscaling.knative.dev/minScale"] = fmt.Sprintf("%d", cfg.Scale)
	metaSpec.Annotations["autoscaling.knative.dev/maxScale"] = fmt.Sprintf("%d", c.MaxScale)

	metaSpec.Annotations["kubernetes.io/egress-bandwidth"] = "10M"
	metaSpec.Annotations["kubernetes.io/ingress-bandwidth"] = "10M"

	return metaSpec
}

func buildVolumes(c *ClientConfig, cfg *Config) []corev1.Volume {
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

func buildContainers(c *ClientConfig, cfg *Config) ([]corev1.Container, error) {
	// set resource limits.
	rl, err := buildResourceLimits(c, cfg)
	if err != nil {
		return nil, err
	}

	// user container
	uc := corev1.Container{
		Name:      containerUser,
		Image:     cfg.Image,
		Env:       buildEnvVars(c, cfg),
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
	_ = corev1.Container{
		Name:         containerSidecar,
		Image:        c.Sidecar,
		Env:          buildEnvVars(c, cfg),
		VolumeMounts: vMounts,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: 8890,
			},
		},
	}

	return []corev1.Container{uc}, nil
}

func buildResourceLimits(c *ClientConfig, cfg *Config) (corev1.ResourceRequirements, error) {
	return corev1.ResourceRequirements{}, nil
}

func buildEnvVars(c *ClientConfig, cfg *Config) []corev1.EnvVar {
	proxyEnvs := []corev1.EnvVar{}

	proxyEnvs = append(proxyEnvs, corev1.EnvVar{
		Name:  "DIREKTIV_APP",
		Value: "sidecar",
	})

	return proxyEnvs
}

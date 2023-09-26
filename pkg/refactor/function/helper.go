// nolint
package function

import (
	"github.com/mattn/go-shellwords"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func buildService(c *ClientConfig, cfg *FunctionConfig) (*servingv1.Service, error) {
	containers, err := buildContainers(c, cfg)
	if err != nil {
		return nil, err
	}
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
	//cs, err := fetchServiceAPI()
	//if err != nil {
	//	logger.Errorf("error getting clientset for knative: %v", err)
	//	return nil, err
	//}

	return svc, nil
}

func buildServiceMeta(c *ClientConfig, cfg *FunctionConfig) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:        cfg.id(),
		Namespace:   c.Namespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}

	meta.Annotations["direktiv.io/input_hash"] = cfg.hash()

	return meta
}

func buildPodMeta(c *ClientConfig, cfg *FunctionConfig) metav1.ObjectMeta {
	metaSpec := metav1.ObjectMeta{
		Namespace:   c.Namespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	metaSpec.Labels["direktiv-app"] = "direktiv"

	return metaSpec
}

func buildVolumes(c *ClientConfig, cfg *FunctionConfig) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "workdir",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	// volumes = append(volumes, functionsConfig.extraVolumes...)

	return volumes
}

func buildContainers(c *ClientConfig, cfg *FunctionConfig) ([]corev1.Container, error) {
	// set resource limits.

	// user container
	uc := corev1.Container{
		Name:  containerUser,
		Image: cfg.Config.Image,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/mnt/shared",
			},
		},
	}

	if len(cfg.Config.CMD) > 0 {
		args, err := shellwords.Parse(cfg.Config.CMD)
		if err != nil {
			return []corev1.Container{}, err
		}
		uc.Command = args
	}

	return []corev1.Container{uc}, nil
}

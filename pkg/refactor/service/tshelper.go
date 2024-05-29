package service

import (
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func buildTypescriptService(c *core.Config, sv *core.ServiceFileData, registrySecrets []corev1.LocalObjectReference) (*servingv1.Service, error) {

	fmt.Println(string(sv.TypescriptFile))

	nonRoot := false

	containers, err := buildTypescriptContainers(c, sv)
	if err != nil {
		return nil, err
	}

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
							SecurityContext: &corev1.PodSecurityContext{
								RunAsNonRoot: &nonRoot,
								SeccompProfile: &corev1.SeccompProfile{
									// should we change it to runtime?
									Type: corev1.SeccompProfileTypeUnconfined,
								},
							},
							ServiceAccountName: c.KnativeServiceAccount,
							Containers:         containers,
							// InitContainers:     initContainers,
							Volumes:  buildVolumes(c, sv),
							Affinity: &corev1.Affinity{},
						},
					},
				},
			},
		},
	}

	return svc, nil

}

func buildTypescriptContainers(c *core.Config, sv *core.ServiceFileData) ([]corev1.Container, error) {

	allowPrivilegeEscalation := true
	secContext := &corev1.SecurityContext{
		AllowPrivilegeEscalation: &allowPrivilegeEscalation,
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				// corev1.Capability("ALL"),
			},
		},
	}

	uc := corev1.Container{
		Name:  containerUser,
		Image: sv.Image,
		Env:   buildEnvVars(false, c, sv),
		// Resources: *rl,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/mnt/shared",
			},
		},
		SecurityContext: secContext,
	}

	return []corev1.Container{uc}, nil
}

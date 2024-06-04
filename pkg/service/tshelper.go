package service

import (
	"fmt"

	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/direktiv/direktiv/pkg/core"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func buildTypescriptService(c *core.Config, sv *core.ServiceFileData, registrySecrets []corev1.LocalObjectReference) (*servingv1.Service, error) {

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

	compiler, err := compiler.New(sv.FilePath, string(sv.TypescriptFile))
	if err != nil {
		return nil, err
	}

	flowInformation, err := compiler.CompileFlow()
	if err != nil {
		return nil, err
	}

	fmt.Println("SECRETS")
	fmt.Println(flowInformation.Secrets)
	fmt.Println("FILES")
	fmt.Println(flowInformation.Files)
	fmt.Println("FUNCTIONS")
	fmt.Println(flowInformation.Functions)

	basicPort := 8081

	userContainerBasicEnvs := buildEnvVars(false, c, sv)
	for k := range flowInformation.Functions {
		fmt.Println("APPEND!!!!!")
		userContainerBasicEnvs = append(userContainerBasicEnvs, corev1.EnvVar{
			Name:  k,
			Value: fmt.Sprintf("http://localhost:%d", basicPort),
		})
		basicPort++
	}

	fmt.Println("ENVS")
	fmt.Println(userContainerBasicEnvs)

	uc := corev1.Container{
		Name:  containerUser,
		Image: c.KnativeSidecar,
		Args:  []string{"tsengine"},
		Env:   userContainerBasicEnvs,
		// Resources: *rl,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/mnt/shared",
			},
		},
		SecurityContext: secContext,
	}
	containers := []corev1.Container{uc}

	// add function containers
	basicPort = 8081
	for k, v := range flowInformation.Functions {

		// v.Cmd
		// v.Envs
		// v.Size
		// v.GetID()

		fnContainer := corev1.Container{

			Name:  k,
			Image: v.Image,
			Ports: []corev1.ContainerPort{
				{
					ContainerPort: int32(basicPort),
				},
			},
			// Env:   buildEnvVars(false, c, sv),
			// Resources: *rl,

			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "workdir",
					MountPath: "/mnt/shared",
				},
			},
			SecurityContext: secContext,
		}

		basicPort++
		containers = append(containers, fnContainer)

	}

	fmt.Println("LENGTH!!!!")
	fmt.Println(len(containers))

	return containers, nil
}

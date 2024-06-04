package service

import (
	"fmt"

	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/direktiv/direktiv/pkg/core"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func buildTypescriptService(c *core.Config, sv *core.ServiceFileData, registrySecrets []corev1.LocalObjectReference) (*servingv1.Service, error) {

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

	// set scale from configuration
	if len(flowInformation.Definition.Scale) > 0 {
		sv.Scale = flowInformation.Definition.Scale[0].Min
	}

	nonRoot := false

	containers, err := buildTypescriptContainers(c, sv, flowInformation)
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

func buildTypescriptContainers(c *core.Config, sv *core.ServiceFileData, flowInformation *compiler.FlowInformation) ([]corev1.Container, error) {

	rl, err := buildResourceLimits(c, sv.Size)
	if err != nil {
		return nil, err
	}

	allowPrivilegeEscalation := true
	secContext := &corev1.SecurityContext{
		AllowPrivilegeEscalation: &allowPrivilegeEscalation,
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				// corev1.Capability("ALL"),
			},
		},
	}

	// add engine
	basicPort := 8081
	userContainerBasicEnvs := buildEnvVars(false, c, sv)
	for k := range flowInformation.Functions {
		userContainerBasicEnvs = append(userContainerBasicEnvs, corev1.EnvVar{
			Name:  k,
			Value: fmt.Sprintf("http://localhost:%d", basicPort),
		})
		basicPort++
	}
	userContainerBasicEnvs = append(userContainerBasicEnvs,
		corev1.EnvVar{
			Name:  "DIREKTIV_JSENGINE_BASEDIR",
			Value: "/mnt/shared",
		},
		corev1.EnvVar{
			Name:  "DIREKTIV_APP",
			Value: "tsengine",
		},
		corev1.EnvVar{
			Name:  "DIREKTIV_JSENGINE_NAMESPACE",
			Value: sv.Namespace,
		},
		corev1.EnvVar{
			Name:  "DIREKTIV_JSENGINE_FLOWPATH",
			Value: sv.FilePath,
		},
	)

	db := corev1.EnvVar{
		Name: "DIREKTIV_DB",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: "db",
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "direktiv-secrets-functions",
				},
			},
		},
	}
	userContainerBasicEnvs = append(userContainerBasicEnvs, db)

	key := corev1.EnvVar{
		Name: "DIREKTIV_SECRET_KEY",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: "key",
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "direktiv-secrets-functions",
				},
			},
		},
	}
	userContainerBasicEnvs = append(userContainerBasicEnvs, key)

	// LogLevel  string `env:"DIREKTIV_JS_ENGINE_LOGLEVEL" envDefault:"info"`
	// SelfCopy string `env:"DIREKTIV_JSENGINE_SELFCOPY"`

	baseCPU, _ := resource.ParseQuantity("4")
	baseMem, _ := resource.ParseQuantity("4096M")
	// baseDisk, _ := resource.ParseQuantity("64M")

	rr := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			"cpu":    baseCPU,
			"memory": baseMem,
			// "ephemeral-storage": baseDisk,
		},
		Limits: corev1.ResourceList{
			"cpu":    baseCPU,
			"memory": baseMem,
			// "ephemeral-storage": baseDisk,
		},
	}

	uc := corev1.Container{
		Name:  containerUser,
		Image: c.KnativeSidecar,
		Env:   userContainerBasicEnvs,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/mnt/shared",
			},
		},
		SecurityContext: secContext,
		Resources:       rr,
	}
	containers := []corev1.Container{uc}

	// add function containers
	basicPort = 8081
	for k, v := range flowInformation.Functions {

		// only workflow functions
		if v.Image == "" {
			continue
		}

		// v.Cmd
		// v.Envs
		// v.Size

		fnContainer := corev1.Container{

			Name:  k,
			Image: v.Image,
			Ports: []corev1.ContainerPort{
				{
					ContainerPort: int32(basicPort),
				},
			},
			// Env:   buildEnvVars(false, c, sv),
			Resources: *rl,

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

package service

import (
	"fmt"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/tsengine/compiler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func buildTypescriptService(c *core.Config, sv *core.ServiceFileData, _ []corev1.LocalObjectReference) (*servingv1.Service, error) {
	compiler, err := compiler.New(sv.FilePath, string(sv.TypescriptFile))
	if err != nil {
		return nil, err
	}

	flowInformation, err := compiler.CompileFlow()
	if err != nil {
		return nil, err
	}

	// check if special command is used in a function
	useSpecialCommands := false
	for k := range flowInformation.Functions {
		fn := flowInformation.Functions[k]
		if fn.Cmd == "direktiv" {
			useSpecialCommands = true
			break
		}
	}

	// add init container if a function requires it
	initContainers := []corev1.Container{}
	if useSpecialCommands {
		initContainers = append(initContainers, buildInitContainer(c.KnativeSidecar))
	}

	userContainerBasicEnvs := buildEnvVars(false, c, sv)

	// build engine container
	engineContainer := buildEngineContainer(c, sv, flowInformation.Functions,
		userContainerBasicEnvs)

	// set scale from configuration
	if len(flowInformation.Definition.Scale) > 0 {
		sv.Scale = flowInformation.Definition.Scale[0].Min
	}

	containers, err := buildFunctionContainers(c, sv, flowInformation)
	if err != nil {
		return nil, err
	}

	containers = append(containers, engineContainer)

	nonRoot := false
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
									Type: corev1.SeccompProfileTypeUnconfined,
								},
							},
							ServiceAccountName: c.KnativeServiceAccount,
							Containers:         containers,
							InitContainers:     initContainers,
							Volumes:            buildVolumes(c, sv),
							Affinity:           &corev1.Affinity{},
						},
					},
				},
			},
		},
	}

	return svc, nil
}

func buildInitContainer(sidecar string) corev1.Container {
	return corev1.Container{
		Name:  "init",
		Image: sidecar,
		Env: []corev1.EnvVar{
			{
				Name:  "DIREKTIV_APP",
				Value: "tsengine",
			},
			{
				Name:  "DIREKTIV_JSENGINE_SELFCOPY",
				Value: "/mnt/shared/direktiv",
			},
			{
				Name:  "DIREKTIV_JSENGINE_SELFCOPY_EXIT",
				Value: "true",
			},
			{
				Name:  "DIREKTIV_JSENGINE_FLOWPATH",
				Value: "dummy",
			},
			{
				Name:  "DIREKTIV_JSENGINE_NAMESPACE",
				Value: "dummy",
			},
			{
				Name:  "DIREKTIV_SECRET_KEY",
				Value: "dummy",
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/mnt/shared/",
			},
		},
	}
}

func buildEngineContainer(c *core.Config, sv *core.ServiceFileData, functions map[string]compiler.Function, basicEnvs []corev1.EnvVar) corev1.Container {
	basicPort := 8081

	for k := range functions {
		basicEnvs = append(basicEnvs, corev1.EnvVar{
			Name:  k,
			Value: fmt.Sprintf("http://localhost:%d", basicPort),
		})
		basicPort++
	}

	basicEnvs = append(basicEnvs,
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
		corev1.EnvVar{
			Name: "DIREKTIV_DB",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "db",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "direktiv-secrets-functions",
					},
				},
			},
		},
		corev1.EnvVar{
			Name: "DIREKTIV_SECRET_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "key",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "direktiv-secrets-functions",
					},
				},
			},
		},
	)

	return corev1.Container{
		Name:  containerUser,
		Image: c.KnativeSidecar,
		Env:   basicEnvs,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: int32(8080),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/mnt/shared",
			},
		},
		SecurityContext: getSecurityContext(),
	}
}

func getSecurityContext() *corev1.SecurityContext {
	allowPrivilegeEscalation := true
	return &corev1.SecurityContext{
		AllowPrivilegeEscalation: &allowPrivilegeEscalation,
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{},
		},
	}
}

func buildFunctionContainers(c *core.Config, sv *core.ServiceFileData, flowInformation *compiler.FlowInformation) ([]corev1.Container, error) {
	rl, err := buildResourceLimits(c, sv)
	if err != nil {
		return nil, err
	}

	containers := make([]corev1.Container, 0)

	// add function containers
	basicPort := 8081
	for k, v := range flowInformation.Functions {
		// only workflow functions
		if v.Image == "" {
			continue
		}

		// create envs
		fnContainerBasicEnvs := buildEnvVars(false, c, sv)
		const DIREKTIV = "direktiv"
		cmd := v.Cmd
		if v.Cmd == DIREKTIV {
			cmd = "/mnt/shared/direktiv"

			// add direktiv app env
			fnContainerBasicEnvs = append(fnContainerBasicEnvs, corev1.EnvVar{
				Name:  "DIREKTIV_APP",
				Value: "cmdserver",
			})
			fnContainerBasicEnvs = append(fnContainerBasicEnvs, corev1.EnvVar{
				Name:  "DIREKTIV_PORT",
				Value: fmt.Sprintf("%d", basicPort),
			})
		}

		// v.Envs
		// v.Size

		fnContainer := corev1.Container{
			Name:      k,
			Image:     v.Image,
			Command:   []string{cmd},
			Env:       fnContainerBasicEnvs,
			Resources: *rl,

			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "workdir",
					MountPath: "/mnt/shared",
				},
			},
			SecurityContext: getSecurityContext(),
		}

		basicPort++
		containers = append(containers, fnContainer)
	}

	return containers, nil
}

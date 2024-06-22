package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/mattn/go-shellwords"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	direktivCmdExecValue  = "/usr/share/direktiv/direktiv-cmd"
	direktivProxyHTTPS    = "HTTPS_PROXY"
	direktivProxyHTTP     = "HTTP_PROXY"
	direktivProxyNO       = "NO_PROXY"
	direktivOpentelemetry = "DIREKTIV_OTLP"
	direktivFlowEndpoint  = "DIREKTIV_FLOW_ENDPOINT"
	direktivDebug         = "DIREKTIV_DEBUG"
)

func buildService(c *core.Config, sv *core.ServiceFileData, registrySecrets []corev1.LocalObjectReference) (*servingv1.Service, error) {
	containers, err := buildContainers(c, sv)
	if err != nil {
		return nil, err
	}

	nonRoot := false

	initContainers := []corev1.Container{}
	if sv.Cmd == direktivCmdExecValue {
		initContainers = append(initContainers, corev1.Container{
			Name:  "init",
			Image: c.KnativeSidecar,
			Env: []corev1.EnvVar{
				{
					Name:  "DIREKTIV_APP",
					Value: "init",
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "bindir",
					MountPath: "/usr/share/direktiv/",
				},
			},
		})
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
							InitContainers:     initContainers,
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

	// Set Registry Secrets
	svc.Spec.ConfigurationSpec.Template.Spec.ImagePullSecrets = registrySecrets
	svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.ImagePullSecrets = registrySecrets

	return svc, nil
}

func buildServiceMeta(c *core.Config, sv *core.ServiceFileData) metav1.ObjectMeta {
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

func buildPodMeta(c *core.Config, sv *core.ServiceFileData) metav1.ObjectMeta {
	metaSpec := metav1.ObjectMeta{
		Namespace:   c.KnativeNamespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	metaSpec.Labels["direktiv-app"] = "direktiv"

	metaSpec.Annotations["autoscaling.knative.dev/minScale"] = strconv.Itoa(sv.Scale)
	metaSpec.Annotations["autoscaling.knative.dev/maxScale"] = strconv.Itoa(c.KnativeMaxScale)

	if len(c.KnativeNetShape) > 0 {
		metaSpec.Annotations["kubernetes.io/ingress-bandwidth"] = c.KnativeNetShape
		metaSpec.Annotations["kubernetes.io/egress-bandwidth"] = c.KnativeNetShape
	}

	return metaSpec
}

func buildVolumes(_ *core.Config, sv *core.ServiceFileData) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "workdir",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	// add extra folder if bin required
	if sv.Cmd == direktivCmdExecValue {
		volumes = append(volumes, corev1.Volume{
			Name: "bindir",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	return volumes
}

func buildContainers(c *core.Config, sv *core.ServiceFileData) ([]corev1.Container, error) {
	// set resource limits.
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
		SecurityContext: secContext,
	}

	// add volume for binary or add command
	if sv.Cmd == direktivCmdExecValue {
		uc.VolumeMounts = append(uc.VolumeMounts, corev1.VolumeMount{
			Name:      "bindir",
			MountPath: "/usr/share/direktiv/",
		})
	}

	if len(sv.Cmd) > 0 {
		args, err := shellwords.Parse(sv.Cmd)
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
	sidecarEnvs := buildEnvVars(true, c, sv)
	sidecarEnvs = append(sidecarEnvs, corev1.EnvVar{Name: "API_KEY", Value: c.ApiKey})
	sc := corev1.Container{
		Name:         containerSidecar,
		Image:        c.KnativeSidecar,
		Env:          sidecarEnvs,
		VolumeMounts: vMounts,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: containerSidecarPort,
			},
		},
		SecurityContext: secContext,
	}

	return []corev1.Container{uc, sc}, nil
}

func buildResourceLimits(cf *core.Config, size string) (*corev1.ResourceRequirements, error) {
	var (
		m int
		c string
		d int
	)

	switch size {
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
		return nil, fmt.Errorf("service size: '%s' is invalid, expected value: ['small', 'medium', 'large']", size)
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

func buildEnvVars(forSidecar bool, c *core.Config, sv *core.ServiceFileData) []corev1.EnvVar {
	proxyEnvs := []corev1.EnvVar{}

	if len(c.KnativeProxyHTTP) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  direktivProxyHTTP,
			Value: c.KnativeProxyHTTP,
		})
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  strings.ToLower(direktivProxyHTTP),
			Value: c.KnativeProxyHTTP,
		})
	}

	if len(c.KnativeProxyHTTPS) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  direktivProxyHTTPS,
			Value: c.KnativeProxyHTTPS,
		})
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  strings.ToLower(direktivProxyHTTPS),
			Value: c.KnativeProxyHTTPS,
		})
	}

	if len(c.KnativeProxyNo) > 0 {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  direktivProxyNO,
			Value: c.KnativeProxyNo,
		})
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  strings.ToLower(direktivProxyNO),
			Value: c.KnativeProxyNo,
		})
	}

	// add debug if there is an env
	if c.LogDebug {
		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  direktivDebug,
			Value: "true",
		})
	}

	proxyEnvs = append(proxyEnvs, corev1.EnvVar{
		Name:  direktivOpentelemetry,
		Value: c.OpenTelemetry,
	})

	if forSidecar {
		namespace := c.DirektivNamespace

		proxyEnvs = append(proxyEnvs, corev1.EnvVar{
			Name:  direktivFlowEndpoint,
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

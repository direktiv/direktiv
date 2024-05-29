package compiler

import (
	// 	"fmt"

	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
)

// const (
// 	PodServiceHashKey = "direktiv-hash"
// 	PodServiceIDKey   = "direktiv-id"
// )

// // Config:
// // - Namespace
// // - Size for medium, small, large

// // TODO: PATCH!!!

func (fi *FlowInformation) CompilePod(name, path, ns string) (*corev1.Pod, error) {

	containers, err := buildContainers(fi.Functions, path)
	if err != nil {
		return nil, err
	}

	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: buildServiceMeta(name, ns),
		Spec: corev1.PodSpec{
			Volumes: buildVolumes(),
			// ServiceAccountName: ,
			// SecurityContext: ,
			// ImagePullSecrets: ,
			Containers: containers,
		},
	}

	// p, _ := yaml.Marshal(pod)
	// fmt.Println(pod.String())

	// os.WriteFile("/tmp/pod.yaml", p, 0777)

	newFile, err := os.Create("/tmp/pod.yaml")
	if err != nil {
		panic(err)
	}
	defer newFile.Close()
	y := printers.YAMLPrinter{}
	y.PrintObj(pod, newFile)

	return pod, nil

}

// 	containers, err := buildContainers(fi.Body().Functions())
// 	if err != nil {
// 		return nil, err
// 	}

// 	pod := &corev1.Pod{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "Pod",
// 			APIVersion: "v1",
// 		},
// 		ObjectMeta: buildServiceMeta(),
// 		Spec: corev1.PodSpec{
// 			Volumes: buildVolumes(),
// 			// ServiceAccountName: ,
// 			// SecurityContext: ,
// 			// ImagePullSecrets: ,
// 			Containers: containers,
// 		},
// 	}

// 	// b, err := json.MarshalIndent(pod, "", "   ")
// 	// // fmt.Println(err)
// 	// // fmt.Println(string(b))

// 	// f, _ := os.OpenFile("../../delme.yaml", os.O_RDWR|os.O_CREATE, 0644)

// 	// json2yaml.Convert(f, bytes.NewReader(b))

// 	return pod, nil
// }

func buildContainers(fns map[string]Function, path string) ([]corev1.Container, error) {

	startPort := 8080

	vMounts := []corev1.VolumeMount{
		{
			Name:      "shared",
			MountPath: "/direktiv/shared",
		},
		{
			Name:      "instances",
			MountPath: "/direktiv/instances",
		},
		{
			Name:      "bin",
			MountPath: "/direktiv/bin",
		},
	}

	engineContainer := corev1.Container{
		Name:         "engine",
		Image:        "localhost:5000/engine",
		Command:      []string{"/engine"},
		Args:         []string{"server"},
		VolumeMounts: vMounts,
		Env: []corev1.EnvVar{
			{
				Name:  "DIREKTIV_PORT",
				Value: fmt.Sprintf("%d", startPort),
			},
			{
				Name:  "DIREKTIV_JSENGINE_SELFCOPY",
				Value: "/direktiv/bin/engine",
			},
			{
				Name:  "DIREKTIV_JSENGINE_FLOWPATH",
				Value: path,
			},
		},
	}

	// base := []corev1.Container{engineContainer}
	base := []corev1.Container{}

	for k, v := range fns {
		startPort = startPort + 1
		fnContainer := corev1.Container{
			Name:  k,
			Image: v.Image,
			// Command:      []string{"/engine"},
			VolumeMounts: vMounts,
			Env: []corev1.EnvVar{
				{
					Name:  "DIREKTIV_PORT",
					Value: fmt.Sprintf("%d", startPort),
				},
			},
		}

		if v.Cmd == composeStaticCmd {
			fnContainer.Command = []string{"/direktiv/bin/engine", "cmd"}
			// cmd := make(types.ShellCommand, 2)
			// cmd[0] = "/direktiv/bin/engine"
			// cmd[1] = "cmd"

		}

		base = append(base, fnContainer)

		engineContainer.Env = append(engineContainer.Env,
			corev1.EnvVar{
				Name:  k,
				Value: fmt.Sprintf("http://127.0.0.1:%d", startPort),
			},
		)
	}

	base = append(base, engineContainer)

	// 	// stroe the user containers in a hash to avoid duiplicates
	// 	// if they are the same within the workflow
	// 	userContainers := make(map[string]corev1.Container)

	// 	// iterate all user functions and add those containers
	// 	// for i := range fns {
	// 	// 	fn := fns[i]

	// 	// 	// add one to port so they are no overlapping ports
	// 	// 	// for the functions
	// 	// 	startPort = startPort + 1

	// 	// 	hash := fn.GetID()

	// 	// 	// basic env isport
	// 	// 	envs := []corev1.EnvVar{
	// 	// 		{
	// 	// 			Name:  "DIREKTIV_PORT",
	// 	// 			Value: fmt.Sprintf("%d", startPort),
	// 	// 		},
	// 	// 	}

	// 	// 	for k, v := range fn.Envs {
	// 	// 		envs = append(envs, corev1.EnvVar{
	// 	// 			Name:  k,
	// 	// 			Value: v,
	// 	// 		})
	// 	// 	}

	// 	// 	sc := corev1.Container{
	// 	// 		Name:         hash,
	// 	// 		Image:        fn.Image,
	// 	// 		VolumeMounts: vMounts,
	// 	// 		Env:          envs,
	// 	// 	}

	// 	// 	userContainers[hash] = sc
	// 	// }

	// 	// add individual containers withou duplicates
	// for _, v := range userContainers {
	// 	base = append(base, v)
	// }

	return base, nil
}

func buildServiceMeta(name, ns string) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:        name,
		Namespace:   ns,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	meta.Labels["direktiv.io/flow"] = "true"
	meta.Annotations["direktiv.io/id"] = "123"

	return meta
}

func buildVolumes() []corev1.Volume {

	// shared filesystem engine and functions
	volumes := []corev1.Volume{
		{
			Name: "shared",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	// bin dir for direktiv executable
	volumes = append(volumes, corev1.Volume{
		Name: "instances",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	volumes = append(volumes, corev1.Volume{
		Name: "bin",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	return volumes
}

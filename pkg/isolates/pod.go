package isolates

import (
	"context"
	"fmt"
	"os"
	"time"

	shellwords "github.com/mattn/go-shellwords"
	log "github.com/sirupsen/logrus"
	igrpc "github.com/vorteil/direktiv/pkg/isolates/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	pullPolicy = v1.PullAlways
)

// Pod env vars
const (
	PodEnvNamespace  = "DIREKTIV_NAMESPACE"
	PodEnvActionID   = "DIREKTIV_ACTIONID"
	PodEnvInstanceID = "DIREKTIV_INSTANCEID"
	PodEnvStep       = "DIREKTIV_STEP"
	PodEnvEndpoint   = "DIREKTIV_FLOW_ENDPOINT"
)

func createUserContainer(size int, image, cmd string) (v1.Container, error) {

	res, err := generateResourceLimits(size)
	if err != nil {
		log.Errorf("can not parse requests limits")
		return v1.Container{}, err
	}

	userContainer := v1.Container{
		ImagePullPolicy: pullPolicy,
		Resources:       res,
		Name:            "direktiv-container",
		Image:           image,
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/direktiv-data",
			},
		},
		Env: proxyEnvs(),
	}

	if len(cmd) > 0 {
		args, err := shellwords.Parse(cmd)
		if err != nil {
			return userContainer, err
		}
		userContainer.Command = args
	}

	return userContainer, nil

}

func commonEnvs(in *igrpc.CreatePodRequest) []v1.EnvVar {

	return []v1.EnvVar{
		{
			Name:  PodEnvNamespace,
			Value: in.GetInfo().GetNamespace(),
		},
		{
			Name:  PodEnvActionID,
			Value: in.GetActionID(),
		},
		{
			Name:  PodEnvInstanceID,
			Value: in.GetInstanceID(),
		},
		{
			Name:  PodEnvStep,
			Value: fmt.Sprintf("%d", in.GetStep()),
		},
		{
			Name:  PodEnvEndpoint,
			Value: os.Getenv(PodEnvEndpoint),
		},
	}

}

func (is *isolateServer) CancelIsolatePod(ctx context.Context,
	in *igrpc.CancelPodRequest) (*emptypb.Empty, error) {

	log.Debugf("cancel pod %v", in.GetActionID())

	clientset, err := getClientSet()
	if err != nil {
		log.Errorf("could not get client set: %v", err)
		return &empty, err
	}

	jobs := clientset.BatchV1().Jobs(isolateConfig.Namespace)

	fg := metav1.DeletePropagationBackground
	var gp int64 = 30
	opts := metav1.DeleteOptions{
		PropagationPolicy:  &fg,
		GracePeriodSeconds: &gp,
	}

	err = jobs.DeleteCollection(context.Background(), opts,
		metav1.ListOptions{LabelSelector: fmt.Sprintf("direktiv.io/action-id=%s", in.GetActionID())})

	if err != nil {
		log.Errorf("can not delete job %s: %v", in.GetActionID(), err)
	}

	return &empty, err

}

func (is *isolateServer) CreateIsolatePod(ctx context.Context,
	in *igrpc.CreatePodRequest) (*igrpc.CreatePodResponse, error) {

	log.Debugf("creating pod %v", in.GetInfo().GetName())

	var resp igrpc.CreatePodResponse

	info := in.GetInfo()

	// ttl for kubernetes 1.20+
	var ttl int32 = 60

	clientset, err := getClientSet()
	if err != nil {
		log.Errorf("could not get client set: %v", err)
		return &resp, err
	}

	jobs := clientset.BatchV1().Jobs(isolateConfig.Namespace)

	userContainer, err := createUserContainer(int(info.GetSize()),
		info.GetImage(), info.GetCmd())
	if err != nil {
		log.Errorf("can not create user container: %v", err)
		return &resp, err
	}

	labels := make(map[string]string)
	labels["direktiv.io/action-id"] = in.GetActionID()
	labels["direktiv.io/job"] = "true"

	labels[ServiceHeaderName] = info.GetName()
	labels[ServiceHeaderWorkflow] = info.GetName()
	labels[ServiceHeaderNamespace] = info.GetNamespace()

	commonJobVars := commonEnvs(in)

	initJobVars := append(commonJobVars, v1.EnvVar{
		Name:  "DIREKTIV_LIFECYCLE",
		Value: "init",
	})

	sidecarJobVars := append(commonJobVars,
		v1.EnvVar{
			Name:  "DIREKTIV_LIFECYCLE",
			Value: "run",
		})

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "direktiv-job-",
			Namespace:    isolateConfig.Namespace,
			Labels:       labels,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					ImagePullSecrets: createPullSecrets(info.GetNamespace()),
					Volumes: []v1.Volume{
						{
							Name: "workdir",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
					InitContainers: []v1.Container{
						{
							ImagePullPolicy: pullPolicy,
							Name:            "init-container",
							Image:           isolateConfig.InitPod,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "workdir",
									MountPath: "/direktiv-data",
								},
							},
							Env: initJobVars,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8890,
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							ImagePullPolicy: pullPolicy,
							Name:            "direktiv-sidecar",
							Image:           isolateConfig.InitPod,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "workdir",
									MountPath: "/direktiv-data",
								},
							},
							Env: sidecarJobVars,
						},
						userContainer,
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}

	j, err := jobs.Create(context.TODO(), jobSpec, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("failed to create job: %v", err)
		return &resp, err
	}

	watch, err := clientset.CoreV1().Pods(isolateConfig.Namespace).Watch(context.Background(),
		metav1.ListOptions{LabelSelector: fmt.Sprintf("job-name=%s", j.ObjectMeta.Name)},
	)
	if err != nil {
		log.Errorf("can not watch job pod: %v", err)
		// whatever happend, we try to delet the pod
		jobs.Delete(context.TODO(), j.ObjectMeta.Name, metav1.DeleteOptions{})
		return &resp, err
	}

	waitFn := func(job string) (string, error) {

		var (
			p  *v1.Pod
			ok bool
		)

		for {
			select {
			case event := <-watch.ResultChan():
				p, ok = event.Object.(*v1.Pod)
				if !ok {
					continue
				}

				// as soon is reachable we break
				pip := p.Status.PodIP
				if len(pip) > 0 {
					log.Debugf("ip for pod %s", pip)
					return pip, nil
				}

			case <-time.After(30 * time.Second):
				jobs.Delete(context.TODO(), job, metav1.DeleteOptions{})
				if p != nil {
					// delete the pod too if possible
					clientset.CoreV1().Pods(isolateConfig.Namespace).Delete(context.TODO(),
						p.Name, metav1.DeleteOptions{})
				}
				return "", fmt.Errorf("timeout for pod")
			}
		}

	}

	ip, err := waitFn(j.ObjectMeta.Name)
	if err != nil {
		log.Errorf("timeout for job pod creation %s", j.ObjectMeta.Name)
		return &resp, err
	}

	resp.Ip = &ip
	return &resp, nil

}

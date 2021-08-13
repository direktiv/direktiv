package functions

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	shellwords "github.com/mattn/go-shellwords"
	log "github.com/sirupsen/logrus"
	igrpc "github.com/vorteil/direktiv/pkg/functions/grpc"
	"github.com/vorteil/direktiv/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watchv1 "k8s.io/apimachinery/pkg/watch"
)

const (
	pullPolicy = v1.PullAlways
)

// Pod env vars
const (
	PodEnvActionID   = "DIREKTIV_ACTIONID"
	PodEnvInstanceID = "DIREKTIV_INSTANCEID"
	PodEnvStep       = "DIREKTIV_STEP"
)

var namespaceCounter map[string]int64

func runPodRequestLimiter(echan chan error) {

	namespaceCounter = make(map[string]int64)
	var mtx sync.Mutex

	// opts for clean job
	fg := metav1.DeletePropagationBackground
	var gp int64 = 30
	opts := metav1.DeleteOptions{
		PropagationPolicy:  &fg,
		GracePeriodSeconds: &gp,
	}

	clientset, err := getClientSet()
	if err != nil {
		log.Errorf("could not get client set: %v", err)
		echan <- err
		return
	}

	jobs := clientset.BatchV1().Jobs(functionsConfig.Namespace)

	watch, err := jobs.Watch(context.Background(),
		metav1.ListOptions{LabelSelector: "direktiv.io/job=true"},
	)
	if err != nil {
		log.Errorf("can not create job watcher: %v", err)
		echan <- err
		return
	}

	echan <- nil

	for {
		select {
		case event := <-watch.ResultChan():
			j, ok := event.Object.(*batchv1.Job)
			if !ok {
				continue
			}

			mtx.Lock()

			if ns, ok := j.Labels["direktiv.io/namespace"]; ok {

				if _, ok := namespaceCounter[ns]; !ok {
					namespaceCounter[ns] = 0
				}
				if event.Type == watchv1.Deleted {
					namespaceCounter[ns]--
					log.Debugf("job counter for ns %s: %d", ns, namespaceCounter[ns])
					if namespaceCounter[ns] <= 0 {
						delete(namespaceCounter, ns)
					}
				} else if event.Type == watchv1.Added { // empty string is ADDED
					namespaceCounter[ns]++
					log.Debugf("job counter for ns %s: %d", ns, namespaceCounter[ns])
				}
			}

			mtx.Unlock()

		case <-time.After(60 * time.Second):

			if functionsConfig.PodCleaner {

				log.Debugf("run pod cleaner")
				lock, err := kubeLock("podclean", true)
				if err != nil {
					log.Debugf("can not get pod cleaner lock: %v", err)
					continue
				}

				l, err := jobs.List(context.Background(), metav1.ListOptions{LabelSelector: "direktiv.io/job=true"})
				if err != nil {
					kubeUnlock(lock)
					log.Errorf("can not list jobs: %v", err)
					continue
				}

				jobs := clientset.BatchV1().Jobs(functionsConfig.Namespace)

				for i := range l.Items {
					j := l.Items[i]

					// if nothing is runing and at least one succeeded or failed
					if j.Status.Active == 0 && (j.Status.Succeeded > 0 || j.Status.Failed > 0) {
						log.Debugf("deleting job %v", j.ObjectMeta.Name)
						err = jobs.Delete(context.Background(), j.ObjectMeta.Name, opts)
						if err != nil {
							log.Errorf("could not delete job: %v", err)
						}
					}
				}

				kubeUnlock(lock)
			}

		}
	}

}

func createUserContainer(size int, image, cmd string) (v1.Container, error) {

	res, err := generateResourceLimits(size)
	if err != nil {
		log.Errorf("can not parse requests limits")
		return v1.Container{}, err
	}

	userContainer := v1.Container{
		ImagePullPolicy: pullPolicy,
		Resources:       res,
		Name:            containerUser,
		Image:           image,
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/direktiv-data",
			},
		},
		Env: proxyEnvs(false),
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

func commonEnvs(in *igrpc.CreatePodRequest, ns string) []v1.EnvVar {

	e := proxyEnvs(true)

	add := []v1.EnvVar{
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
			Name:  util.DirektivNamespace,
			Value: ns,
		},
	}

	return append(e, add...)

}

func (is *functionsServer) CancelFunctionsPod(ctx context.Context,
	in *igrpc.CancelPodRequest) (*emptypb.Empty, error) {

	log.Debugf("cancel pod %v", in.GetActionID())

	clientset, err := getClientSet()
	if err != nil {
		log.Errorf("could not get client set: %v", err)
		return &empty, err
	}

	jobs := clientset.BatchV1().Jobs(functionsConfig.Namespace)

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

func (is *functionsServer) CreateFunctionsPod(ctx context.Context,
	in *igrpc.CreatePodRequest) (*igrpc.CreatePodResponse, error) {

	log.Debugf("creating pod %v", in.GetInfo().GetName())

	var resp igrpc.CreatePodResponse

	info := in.GetInfo()

	// if MaxJobs
	var (
		c  int64
		ok bool
	)
	if c, ok = namespaceCounter[info.GetNamespace()]; ok {
		if c >= int64(functionsConfig.MaxJobs) {
			return &resp, fmt.Errorf("max job number exceeded")
		}
	}

	// ttl for kubernetes 1.20+
	var ttl int32 = 60

	clientset, err := getClientSet()
	if err != nil {
		log.Errorf("could not get client set: %v", err)
		return &resp, err
	}

	jobs := clientset.BatchV1().Jobs(functionsConfig.Namespace)

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

	commonJobVars := commonEnvs(in, info.GetNamespace())

	annotations := make(map[string]string)
	annotations["kubernetes.io/ingress-bandwidth"] = functionsConfig.NetShape
	annotations["kubernetes.io/egress-bandwidth"] = functionsConfig.NetShape

	initJobVars := make([]v1.EnvVar, len(commonJobVars))
	copy(initJobVars, commonJobVars)
	initJobVars = append(initJobVars, v1.EnvVar{
		Name:  "DIREKTIV_LIFECYCLE",
		Value: "init",
	})

	sidecarJobVars := append(commonJobVars,
		v1.EnvVar{
			Name:  "DIREKTIV_LIFECYCLE",
			Value: "run",
		})

	// if flow uses tls or mtls we need the certificate
	// needs to have the same name in direktiv's namespace as it has
	// in the service namespace

	tlsVolumeMount := v1.VolumeMount{}

	volumes := []v1.Volume{
		{
			Name: "workdir",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}

	volumeMounts := []v1.VolumeMount{
		{
			Name:      "workdir",
			MountPath: "/direktiv-data",
		},
	}

	if util.GrpcCfg().FlowTLS != "" && util.GrpcCfg().FlowTLS != "none" {

		certName := "flowcerts"
		tlsVolume := v1.Volume{}
		tlsVolume.Name = certName
		tlsVolume.Secret = &v1.SecretVolumeSource{
			SecretName: util.GrpcCfg().FlowTLS,
		}
		volumes = append(volumes, tlsVolume)

		tlsVolumeMount.Name = certName
		tlsVolumeMount.MountPath = "/etc/direktiv/certs/flow"
		tlsVolumeMount.ReadOnly = true
		volumeMounts = append(volumeMounts, tlsVolumeMount)
	}

	if functionsConfig.InitPodCertificate != "none" &&
		functionsConfig.InitPodCertificate != "" {

		certName := "podcerts"
		tlsVolume := v1.Volume{}
		tlsVolume.Name = certName
		tlsVolume.Secret = &v1.SecretVolumeSource{
			SecretName: functionsConfig.InitPodCertificate,
		}
		volumes = append(volumes, tlsVolume)

		tlsVolumeMount.Name = certName
		tlsVolumeMount.MountPath = "/etc/direktiv/certs/http"
		tlsVolumeMount.ReadOnly = true
		volumeMounts = append(volumeMounts, tlsVolumeMount)

	}

	mountToken := false
	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "direktiv-job-",
			Namespace:    functionsConfig.Namespace,
			Labels:       labels,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: v1.PodSpec{
					AutomountServiceAccountToken: &mountToken,
					ImagePullSecrets:             createPullSecrets(info.GetNamespace()),
					Volumes:                      volumes,
					InitContainers: []v1.Container{
						{
							ImagePullPolicy: pullPolicy,
							Name:            "init-container",
							Image:           functionsConfig.InitPod,
							VolumeMounts:    volumeMounts,
							Env:             initJobVars,
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
							Name:            containerSidecar,
							Image:           functionsConfig.InitPod,
							VolumeMounts:    volumeMounts,
							Env:             sidecarJobVars,
						},
						userContainer,
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}

	if len(functionsConfig.Runtime) > 0 && functionsConfig.Runtime != "default" {
		log.Debugf("setting runtime class %v", functionsConfig.Runtime)
		jobSpec.Spec.Template.Spec.RuntimeClassName = &functionsConfig.Runtime
	}

	j, err := jobs.Create(context.TODO(), jobSpec, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("failed to create job: %v", err)
		return &resp, err
	}

	watch, err := clientset.CoreV1().Pods(functionsConfig.Namespace).Watch(context.Background(),
		metav1.ListOptions{LabelSelector: fmt.Sprintf("job-name=%s", j.ObjectMeta.Name)},
	)
	if err != nil {
		log.Errorf("can not watch job pod: %v", err)
		// whatever happend, we try to delet the pod
		jobs.Delete(context.TODO(), j.ObjectMeta.Name, metav1.DeleteOptions{})
		return &resp, err
	}

	waitFn := func(job string) (string, string, error) {

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
				hostname := fmt.Sprintf("%s.%s.pod", strings.ReplaceAll(pip, ".", "-"), p.Namespace)

				// 172-17-0-3.default.pod.cluster.local

				if len(pip) > 0 {
					log.Debugf("ip for pod %s, hostname %s", pip, hostname)
					return pip, hostname, nil
				}

			case <-time.After(30 * time.Second):
				jobs.Delete(context.TODO(), job, metav1.DeleteOptions{})
				if p != nil {
					// delete the pod too if possible
					clientset.CoreV1().Pods(functionsConfig.Namespace).Delete(context.TODO(),
						p.Name, metav1.DeleteOptions{})
				}
				return "", "", fmt.Errorf("timeout for pod")
			}
		}

	}

	ip, hostname, err := waitFn(j.ObjectMeta.Name)
	if err != nil {
		log.Errorf("timeout for job pod creation %s", j.ObjectMeta.Name)
		return &resp, err
	}

	resp.Ip = &ip
	resp.Hostname = &hostname

	return &resp, nil

}

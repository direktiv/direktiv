package direktiv

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	shellwords "github.com/mattn/go-shellwords"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/isolates"
	igrpc "github.com/vorteil/direktiv/pkg/isolates/grpc"
	"github.com/vorteil/direktiv/pkg/model"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	kubeAPIKServiceURL         = "https://kubernetes.default.svc/apis/serving.knative.dev/v1/namespaces/%s/services"
	kubeAPIKServiceURLSpecific = "https://kubernetes.default.svc/apis/serving.knative.dev/v1/namespaces/%s/services/%s"

	annotationNamespace = "direktiv.io/namespace"
	annotationURL       = "direktiv.io/url"
	annotationURLHash   = "direktiv.io/urlhash"

	httpsProxy = "HTTPS_PROXY"
	httpProxy  = "HTTP_PROXY"
	noProxy    = "NO_PROXY"

	pullPolicy      = v1.PullAlways
	cleanupInterval = 60
	dbLockID        = 123456

	prefixNamespace = "ns:"
	prefixGlobal    = "g:"
)

const (
	k8sNamespaceVar = "DIREKTIV_KUBERNETES_NAMESPACE"
	secretsPrefix   = "direktiv-secret"
)

type kubeRequest struct {
	serviceTempl string
	sidecar      string

	apiConfig *rest.Config
	mtx       sync.Mutex
}

var (
	gracePeriod int64 = 10
	kubeReq           = kubeRequest{}

	knativeMtx sync.Mutex
)

var (
	kubeCounter          = 0
	kubeCounterDelta     = 0
	kubeCounterTimestamp time.Time
	kubeLockQueue        chan *kubeLockRequest
)

type kubeLockRequest struct {
	ch chan bool
}

func initKubeLock() {
	// kubeLockQueue = make(chan *kubeLockRequest, 1024)
	// pollKubeLock()
	// go runKubeLock()
}

func pollKubeLock() {

	for {
		// clientset, kns, err := getClientSet()
		// if err != nil {
		// 	log.Errorf("could not get client set: %v", err)
		// 	time.Sleep(time.Second)
		// 	continue
		// }
		//
		// jobs := clientset.BatchV1().Jobs(kns)
		//
		// l, err := jobs.List(context.Background(), metav1.ListOptions{LabelSelector: "direktiv.io/job=true"})
		// if err != nil {
		// 	log.Errorf("can not list jobs: %v", err)
		// 	time.Sleep(time.Second)
		// 	continue
		// }
		//
		// kubeCounterTimestamp = time.Now()
		// kubeCounter = len(l.Items)
		// kubeCounterDelta = 0
		//
		// log.Infof("kubelock polling discovered %v pods", kubeCounter)

		return
	}

}

func runKubeLock() {

	// for {
	// min := kubeCounterTimestamp.Add(time.Second)
	// max := kubeCounterTimestamp.Add(time.Minute)
	// if kubeCounter > 100 || kubeCounterDelta > 10 {
	// 	t := time.Now()
	// 	if !t.After(min) {
	// 		<-time.After(min.Sub(time.Now()))
	// 	}
	// 	pollKubeLock()
	// 	continue
	// }
	//
	// select {
	// case r := <-kubeLockQueue:
	// 	kubeCounter++
	// 	kubeCounterDelta++
	// 	close(r.ch)
	// case <-time.After(max.Sub(time.Now())):
	// 	pollKubeLock()
	// }
	// }

}

func enqueueKubeLock(ch chan bool) {
	kubeLockQueue <- &kubeLockRequest{
		ch: ch,
	}
}

func queueForKubeLock() (<-chan bool, error) {
	ch := make(chan bool)
	select {
	case kubeLockQueue <- &kubeLockRequest{
		ch: ch,
	}:
		return ch, nil
	default:
		close(ch)
		return nil, errors.New("kube lock overload")
	}

}

func getKubeLock(ctx context.Context) error {

	ch, err := queueForKubeLock()
	if err != nil {
		return err
	}

	for {
		select {
		case <-ch:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

}

func deleteJob(name string) error {

	clientset, kns, err := getClientSet()
	if err != nil {
		return err
	}

	jobs := clientset.BatchV1().Jobs(kns)

	fg := metav1.DeletePropagationBackground
	opts := metav1.DeleteOptions{
		PropagationPolicy:  &fg,
		GracePeriodSeconds: &gracePeriod,
	}
	log.Debugf("deleting job with name %v", name)

	return jobs.Delete(context.Background(), name, opts)

}

// TTL is beta so e.g. GKE doesn't have it anabled in 1.20 clusters
// it is configurable to turn it off
func completedJobsCleaner(db *dbManager) error {

	log.Infof("starting pod cleaner")

	clientset, kns, err := getClientSet()
	if err != nil {
		log.Errorf("could not get client set: %v", err)
		return err
	}

	jobs := clientset.BatchV1().Jobs(kns)

	for {
		time.Sleep(cleanupInterval * time.Second)

		lock, conn, err := db.tryLockDB(dbLockID)
		if err != nil {
			continue
		}

		if lock {

			l, err := jobs.List(context.Background(), metav1.ListOptions{LabelSelector: "direktiv.io/job=true"})
			if err != nil {
				log.Errorf("can not list jobs: %v", err)
				db.unlockDB(dbLockID, conn)
				continue
			}

			for i := range l.Items {
				j := l.Items[i]

				// we clean up after 1 minute
				// if nothing is runing and at least one succeeded or failed:
				if j.Status.Active == 0 && (j.Status.Succeeded > 0 || j.Status.Failed > 0) &&
					time.Now().After(j.Status.CompletionTime.Add(1*time.Minute)) {

					err := deleteJob(j.ObjectMeta.Name)
					if err != nil {
						log.Errorf("could not delete job: %v", err)
					}

				}
			}

			db.unlockDB(dbLockID, conn)
		}

	}

}

func cancelJob(ctx context.Context, actionID string) {

	clientset, kns, err := getClientSet()
	if err != nil {
		log.Errorf("could not get client set: %v", err)
	}

	jobs := clientset.BatchV1().Jobs(kns)
	opts := metav1.ListOptions{LabelSelector: fmt.Sprintf("direktiv.io/action-id=%s", actionID)}
	jl, err := jobs.List(context.Background(), opts)
	if err != nil {
		log.Errorf("could not list jobs: %v", err)
	}

	if len(jl.Items) > 0 {
		for i := range jl.Items {
			j := jl.Items[i]

			err := deleteJob(j.ObjectMeta.Name)
			if err != nil {
				log.Errorf("could not delete job: %v", err)
			}

		}

	}

}

func createResourceLimits(size int) v1.ResourceList {

	cpu, mem := containerSizeCalc(size)
	rl := make(v1.ResourceList)
	c, _ := resource.ParseQuantity(fmt.Sprintf("%v", cpu))
	rl[v1.ResourceCPU] = c
	c, _ = resource.ParseQuantity(fmt.Sprintf("%vMiB", mem))
	rl[v1.ResourceMemory] = c

	return rl

}

func createUserContainer(size int, image, cmd string) (v1.Container, error) {

	proxyEnvs := []v1.EnvVar{}

	if len(os.Getenv(httpProxy)) > 0 || len(os.Getenv(httpsProxy)) > 0 {

		proxyEnvs = []v1.EnvVar{}
		for _, e := range []string{httpProxy, httpsProxy, noProxy} {
			proxyEnvs = append(proxyEnvs, v1.EnvVar{
				Name:  e,
				Value: os.Getenv(e),
			})
		}

	}

	// Resources ResourceRequirements
	userContainer := v1.Container{
		ImagePullPolicy: pullPolicy,
		Resources: v1.ResourceRequirements{
			Limits: createResourceLimits(size),
		},
		Name:  "direktiv-container",
		Image: image,
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/direktiv-data",
			},
		},
		Env: proxyEnvs,
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

func addPodFunction(ctx context.Context, ah string, ar *isolateRequest) (string, error) {

	err := getKubeLock(ctx)
	if err != nil {
		return "", err
	}

	log.Infof("adding pod function %s", ah)

	clientset, kns, err := getClientSet()
	if err != nil {
		log.Errorf("could not get client set: %v", err)
		return "", err
	}

	jobs := clientset.BatchV1().Jobs(kns)

	var finishSeconds int32 = 60
	size, _ := resource.ParseQuantity("10Mi")

	userContainer, err := createUserContainer(int(ar.Container.Size),
		ar.Container.Image, ar.Container.Cmd)
	if err != nil {
		log.Errorf("can not create user container: %v", err)
		return "", err
	}

	labels := make(map[string]string)
	labels["direktiv.io/action-id"] = ar.ActionID
	labels["direktiv.io/job"] = "true"

	commonJobVars := []v1.EnvVar{
		{
			Name:  "DIREKTIV_NAMESPACE",
			Value: ar.Workflow.Namespace,
		},
		{
			Name:  "DIREKTIV_ACTIONID",
			Value: ar.ActionID,
		},
		{
			Name:  "DIREKTIV_INSTANCEID",
			Value: ar.Workflow.InstanceID,
		},
		{
			Name:  "DIREKTIV_STEP",
			Value: fmt.Sprintf("%d", int64(ar.Workflow.Step)),
		},
		{
			Name:  "DIREKTIV_FLOW_ENDPOINT",
			Value: os.Getenv("DIREKTIV_FLOW_ENDPOINT"),
		},
	}

	initJobVars := append(commonJobVars, v1.EnvVar{
		Name:  "DIREKTIV_LIFECYCLE",
		Value: "init",
	})

	sidecarJobVars := append(commonJobVars, v1.EnvVar{
		Name:  "DIREKTIV_LIFECYCLE",
		Value: "run",
	})

	// generate pull secrets
	secrets, err := kubernetesListRegistriesNames(ar.Workflow.Namespace)
	if err != nil {
		return "", err
	}

	var lo []v1.LocalObjectReference
	for _, s := range secrets {
		lo = append(lo, v1.LocalObjectReference{
			Name: s,
		})
	}

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%v-", ah),
			Namespace:    kns,
			Labels:       labels,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &finishSeconds,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					ImagePullSecrets: lo,
					Volumes: []v1.Volume{
						{
							Name: "workdir",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{
									SizeLimit: &size,
								},
							},
						},
					},
					InitContainers: []v1.Container{
						{
							ImagePullPolicy: pullPolicy,
							Name:            "init-container",
							Image:           os.Getenv("DIREKTIV_FLOW_INITPOD"),
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
							Image:           os.Getenv("DIREKTIV_FLOW_INITPOD"),
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
		return "", err
	}

	log.Debugf("creating job %v", j.ObjectMeta.Name)

	var ip string
	for i := 0; i < 50; i++ {
		podList, err := clientset.CoreV1().Pods(kns).List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("job-name=%s", j.ObjectMeta.Name)})

		if err != nil {
			log.Errorf("can not get clientset for pods: %v", err)
			continue
		}

		if len(podList.Items) == 0 {
			log.Infof("waiting for pod: %s", j.ObjectMeta.Name)
			continue
		}

		if len(podList.Items) > 1 {
			log.Infof("more than one job with than name.")
		}

		pod := podList.Items[0]
		ip = pod.Status.PodIP
		if len(ip) > 0 {
			break
		}

		log.Infof("waiting for pod ip: %s", j.ObjectMeta.Name)
		time.Sleep(500 * time.Millisecond)
	}

	if len(ip) == 0 {
		return "", fmt.Errorf("could not create pod")
	}

	log.Debugf("pod cluster ip: %v", ip)
	return ip, nil

}

func kubernetesListRegistriesNames(namespace string) ([]string, error) {

	var registries []string

	clientset, kns, err := getClientSet()
	if err != nil {
		return registries, err
	}

	var lo metav1.ListOptions
	secrets, err := clientset.CoreV1().Secrets(kns).List(context.Background(), lo)
	if err != nil {
		return registries, err
	}

	for _, s := range secrets.Items {
		if s.Annotations[annotationNamespace] == namespace {
			registries = append(registries, s.Name)
		}
	}

	return registries, nil

}

func kubernetesListRegistries(namespace string) ([]string, error) {

	var registries []string

	clientset, kns, err := getClientSet()
	if err != nil {
		return registries, err
	}

	var lo metav1.ListOptions
	secrets, err := clientset.CoreV1().Secrets(kns).List(context.Background(), lo)
	if err != nil {
		return registries, err
	}

	for _, s := range secrets.Items {
		if s.Annotations[annotationNamespace] == namespace {
			registries = append(registries, fmt.Sprintf("%s###%s",
				s.Annotations[annotationURL], s.Annotations[annotationURLHash]))
		}
	}

	return registries, nil

}

func kubernetesDeleteSecret(name, namespace string) error {

	log.Debugf("deleting secret %s (%s)", name, namespace)

	clientset, kns, err := getClientSet()
	if err != nil {
		return err
	}

	var lo metav1.ListOptions
	secrets, err := clientset.CoreV1().Secrets(kns).List(context.Background(), lo)
	if err != nil {
		return err
	}

	for _, s := range secrets.Items {

		if s.Annotations[annotationNamespace] == namespace &&
			s.Annotations[annotationURL] == name {

			u, err := url.Parse(name)
			if err != nil {
				return err
			}
			secretName := fmt.Sprintf("%s-%s-%s", secretsPrefix, namespace, u.Hostname())

			return clientset.CoreV1().Secrets(kns).Delete(context.Background(), secretName, metav1.DeleteOptions{})

		}

	}

	return fmt.Errorf("no registry with name %s found", name)

}

// func kubernetesAddSecret(name, namespace string, data []byte) error {
//
// 	log.Debugf("adding secret %s (%s)", name, namespace)
//
// 	clientset, kns, err := getClientSet()
// 	if err != nil {
// 		return err
// 	}
//
// 	u, err := url.Parse(name)
// 	if err != nil {
// 		return err
// 	}
//
// 	secretName := fmt.Sprintf("%s-%s-%s", secretsPrefix, namespace, u.Hostname())
//
// 	kubernetesDeleteSecret(name, namespace)
//
// 	sa := &v1.Secret{
// 		Data: make(map[string][]byte),
// 	}
//
// 	sa.Annotations = make(map[string]string)
// 	sa.Annotations[annotationNamespace] = namespace
// 	sa.Annotations[annotationURL] = name
// 	sa.Annotations[annotationURLHash] = base64.StdEncoding.EncodeToString([]byte(name))
//
// 	sa.Name = secretName
// 	sa.Data[".dockerconfigjson"] = data
// 	sa.Type = "kubernetes.io/dockerconfigjson"
//
// 	_, err = clientset.CoreV1().Secrets(kns).Create(context.Background(), sa, metav1.CreateOptions{})
//
// 	return err
//
// }

func getClientSet() (*kubernetes.Clientset, string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, "", err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, "", err
	}

	kns := os.Getenv(direktivWorkflowNamespace)
	if kns == "" {
		kns = "default"
	}

	return clientset, kns, nil
}

func isKnativeFunction(client igrpc.IsolatesServiceClient,
	name, namespace, workflow string) bool {

	// search annotations
	a := make(map[string]string)
	a[isolates.ServiceHeaderName] = name
	a[isolates.ServiceHeaderNamespace] = namespace
	a[isolates.ServiceHeaderWorkflow] = workflow
	a[isolates.ServiceHeaderScope] = isolates.PrefixService

	l, err := client.ListIsolates(context.Background(), &igrpc.ListIsolatesRequest{
		Annotations: a,
	})

	if err != nil {
		log.Errorf("can not list knative service: %v", err)
		return false
	}

	if len(l.Isolates) > 0 {
		return true
	}

	return false
}

func createKnativeFunction(client igrpc.IsolatesServiceClient,
	ir *isolateRequest) error {

	sz := int32(ir.Container.Size)
	scale := int32(ir.Container.Scale)

	cr := igrpc.CreateIsolateRequest{
		Info: &igrpc.BaseInfo{
			Name:      &ir.Container.ID,
			Namespace: &ir.Workflow.Namespace,
			Workflow:  &ir.Workflow.ID,
			Image:     &ir.Container.Image,
			Cmd:       &ir.Container.Cmd,
			Size:      &sz,
			MinScale:  &scale,
		},
	}

	_, err := client.CreateIsolate(context.Background(), &cr)

	return err

}

func createKnativeFunctions(client igrpc.IsolatesServiceClient, wfm model.Workflow, ns string) error {

	for _, f := range wfm.GetFunctions() {

		// only build workflow based isolates
		if f.GetType() != model.ReusableContainerFunctionType {
			continue
		}

		fn := f.(*model.ReusableFunctionDefinition)

		// create services async
		go func(fd *model.ReusableFunctionDefinition,
			model model.Workflow, name, namespace string) {

			sz := int32(fd.Size)
			scale := int32(fd.Scale)

			cr := igrpc.CreateIsolateRequest{
				Info: &igrpc.BaseInfo{
					Name:      &name,
					Namespace: &namespace,
					Workflow:  &model.ID,
					Image:     &fd.Image,
					Cmd:       &fd.Cmd,
					Size:      &sz,
					MinScale:  &scale,
				},
			}

			_, err := client.CreateIsolate(context.Background(), &cr)
			if err != nil {
				log.Errorf("can not create knative service: %v", err)
			}

		}(fn, wfm, fn.ID, ns)

	}

	return nil
}

func deleteKnativeFunctions(client igrpc.IsolatesServiceClient,
	ns, wf, name string) error {

	annotations := make(map[string]string)

	scope := isolates.PrefixService

	if ns != "" {
		annotations[isolates.ServiceHeaderNamespace] = ns
		scope = isolates.PrefixNamespace
	}

	if wf != "" {
		annotations[isolates.ServiceHeaderWorkflow] = wf
		scope = isolates.PrefixWorkflow
	}

	if name != "" {
		annotations[isolates.ServiceHeaderName] = name
		scope = isolates.PrefixService
	}
	annotations[isolates.ServiceHeaderScope] = scope

	dr := igrpc.ListIsolatesRequest{
		Annotations: annotations,
	}

	_, err := client.DeleteIsolates(context.Background(), &dr)
	if err != nil {
		log.Errorf("can not create knative service: %v", err)
	}

	return nil

}

// func getKnativeFunction(isolateClient igrpc.IsolatesServiceClient, svn string) error {
//
// 	r := igrpc.GetIsolateRequest{
// 		Name: &svn,
// 	}
// 	// GetIsolate(ctx context.Context, in *GetIsolateRequest, opts ...grpc.CallOption)
// 	_, err := isolateClient.GetIsolate(context.Background(), &r)
//
// 	if err != nil {
// 		log.Debugf("err %v", err)
// 	}
// 	// u := fmt.Sprintf(kubeAPIKServiceURL, os.Getenv(direktivWorkflowNamespace))
// 	//
// 	// url := fmt.Sprintf("%s/%s", u, svc)
// 	// resp, err := SendKuberequest(http.MethodGet, url, nil)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	//
// 	// if resp.StatusCode != 200 {
// 	// 	return fmt.Errorf("service does not exists")
// 	// }
//
// 	return err
// }

func cmdToCommand(s string) (string, error) {

	args, err := shellwords.Parse(s)
	if err != nil {
		return "", err
	}

	argsQuote := []string{}
	for _, a := range args {
		argsQuote = append(argsQuote, strconv.Quote(a))
	}

	return strings.Join(argsQuote, ", "), nil
}

func containerSizeCalc(size int) (float64, int) {
	var (
		cpu float64
		mem int
	)

	switch size {
	case 1:
		cpu = 1
		mem = 512
	case 2:
		cpu = 2
		mem = 1024
	default:
		cpu = 0.5
		mem = 256
	}

	return cpu, mem

}

// func addKnativeFunction(ir *isolateRequest) error {
//
// 	log.Debugf("adding knative service")
//
// 	namespace := ir.Workflow.Namespace
//
// 	ah, err := serviceToHash(ir)
// 	if err != nil {
// 		return err
// 	}
//
// 	log.Debugf("adding knative service hash %v", ah)
//
// 	cpu, mem := containerSizeCalc(int(ir.Container.Size))
//
// 	u := fmt.Sprintf(kubeAPIKServiceURL, os.Getenv(direktivWorkflowNamespace))
//
// 	// get imagePullSecrets
// 	secrets, err := kubernetesListRegistriesNames(namespace)
// 	if err != nil {
// 		return err
// 	}
//
// 	var sstrings []string
// 	for _, s := range secrets {
// 		sstrings = append(sstrings, fmt.Sprintf("{ \"name\": \"%s\"}", s))
// 	}
//
// 	cmd, err := cmdToCommand(ir.Container.Cmd)
// 	if err != nil {
// 		return err
// 	}
//
// 	svc := fmt.Sprintf(kubeReq.serviceTempl, fmt.Sprintf("%s-%s", namespace, ah), ir.Container.Scale,
// 		strings.Join(sstrings, ","),
// 		ir.Container.Image, cmd, cpu, fmt.Sprintf("%dM", mem), cpu*2, fmt.Sprintf("%dM", mem*2),
// 		kubeReq.sidecar)
//
// 	fmt.Printf("%v\n", svc)
//
// 	resp, err := SendKuberequest(http.MethodPost, u, bytes.NewBufferString(svc))
// 	if err != nil {
// 		log.Errorf("can not send kube request: %v", err)
// 		return err
// 	}
//
// 	if resp.StatusCode != 200 && resp.StatusCode != 201 {
// 		b, _ := ioutil.ReadAll(resp.Body)
// 		defer resp.Body.Close()
// 		return fmt.Errorf("can not add knative service: %v", string(b))
// 	}
//
// 	return nil
//
// }

// func SendKuberequest(method, url string, data io.Reader) (*http.Response, error) {
//
// 	if kubeReq.apiConfig == nil {
// 		config, err := rest.InClusterConfig()
// 		if err != nil {
// 			return nil, err
// 		}
// 		rest.LoadTLSFiles(config)
// 		kubeReq.apiConfig = config
// 	}
//
// 	caCertPool := x509.NewCertPool()
// 	caCertPool.AppendCertsFromPEM(kubeReq.apiConfig.CAData)
//
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			TLSClientConfig: &tls.Config{
// 				RootCAs:    caCertPool,
// 				MinVersion: tls.VersionTLS12,
// 			},
// 		},
// 	}
//
// 	req, err := http.NewRequestWithContext(context.Background(), method, url, data)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Accept", "application/json")
// 	req.Header.Add("Authorization",
// 		fmt.Sprintf("Bearer %s", kubeReq.apiConfig.BearerToken))
//
// 	return client.Do(req)
//
// }

// func k8sNamespace() string {
// 	return os.Getenv(k8sNamespaceVar)
// }

// func serviceToHash(ar *isolateRequest) (string, error) {
// 	re := regexp.MustCompile(`[_,.;'!@#$%^&*()\s]+`)
//
// 	h, err := hash.Hash(fmt.Sprintf("%s-%s-%s", ar.Workflow.Namespace,
// 		ar.Workflow.ID, ar.Container.ID), hash.FormatV2, nil)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	suffix := fmt.Sprintf("-%d", h)
// 	maxLen := 64 - len(fmt.Sprintf("%s.%s", suffix, k8sNamespace())) - (len(ar.Workflow.Namespace) + 1)
//
// 	prefix := fmt.Sprintf("%s-%s", ar.Workflow.ID, ar.Container.ID)
// 	if len(prefix) > maxLen {
// 		prefix = prefix[:maxLen]
// 	}
//
// 	newHash := re.ReplaceAllString(fmt.Sprintf("%s%s", prefix, suffix), "-")
// 	return newHash, nil
//
// }

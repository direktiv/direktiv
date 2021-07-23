package direktiv

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	shellwords "github.com/mattn/go-shellwords"
	hash "github.com/mitchellh/hashstructure/v2"
	log "github.com/sirupsen/logrus"
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

var kubeReq = kubeRequest{}

func cancelJob(ctx context.Context, actionId string) {

}

func addPodFunction(ctx context.Context, ah string, ar *isolateRequest) (string, error) {

	log.Infof("adding pod function %s", ah)

	// pullPolicy := v1.PullIfNotPresent
	pullPolicy := v1.PullAlways

	clientset, kns, err := getClientSet()
	if err != nil {
		log.Errorf("could not get client set: %v", err)
		return "", err
	}

	jobs := clientset.BatchV1().Jobs(kns)

	var finishSeconds int32 = 60

	size, _ := resource.ParseQuantity("10Mi")

	proxyEnvs := []v1.EnvVar{}

	if len(os.Getenv("HTTP_PROXY")) > 0 || len(os.Getenv("HTTPS_PROXY")) > 0 {

		proxyEnvs = []v1.EnvVar{
			{
				Name:  "HTTP_PROXY",
				Value: os.Getenv("HTTP_PROXY"),
			},
			{
				Name:  "HTTPS_PROXY",
				Value: os.Getenv("HTTPS_PROXY"),
			},
			{
				Name:  "NO_PROXY",
				Value: os.Getenv("NO_PROXY"),
			},
		}

	}

	userContainer := v1.Container{
		ImagePullPolicy: pullPolicy,
		Name:            "direktiv-container",
		Image:           ar.Container.Image,
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/direktiv-data",
			},
		},
		Env: proxyEnvs,
	}

	if len(ar.Container.Cmd) > 0 {
		args, err := shellwords.Parse(ar.Container.Cmd)
		if err != nil {
			return "", err
		}
		userContainer.Command = args
	}

	annotations := make(map[string]string)
	annotations["direktiv.io/action-id"] = ar.ActionID
	annotations["direktiv.io/instance-id"] = ar.Workflow.InstanceID
	annotations["direktiv.io/container-id"] = ar.Container.ID

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

	// req.Header.Add(DirektivDeadlineHeader, deadline.Format(time.RFC3339))

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%v-", ah),
			Namespace:    kns,
			Annotations:  annotations,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &finishSeconds,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					ServiceAccountName: "direktiv-sidecar",
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
		podList, err := clientset.CoreV1().Pods(kns).List(context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("job-name=%s", j.ObjectMeta.Name)})

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

func kubernetesAddSecret(name, namespace string, data []byte) error {

	log.Debugf("adding secret %s (%s)", name, namespace)

	clientset, kns, err := getClientSet()
	if err != nil {
		return err
	}

	u, err := url.Parse(name)
	if err != nil {
		return err
	}

	secretName := fmt.Sprintf("%s-%s-%s", secretsPrefix, namespace, u.Hostname())

	kubernetesDeleteSecret(name, namespace)

	sa := &v1.Secret{
		Data: make(map[string][]byte),
	}

	sa.Annotations = make(map[string]string)
	sa.Annotations[annotationNamespace] = namespace
	sa.Annotations[annotationURL] = name
	sa.Annotations[annotationURLHash] = base64.StdEncoding.EncodeToString([]byte(name))

	sa.Name = secretName
	sa.Data[".dockerconfigjson"] = data
	sa.Type = "kubernetes.io/dockerconfigjson"

	_, err = clientset.CoreV1().Secrets(kns).Create(context.Background(), sa, metav1.CreateOptions{})

	return err

}

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

func deleteKnativeFunctions(uid string, db *dbManager) error {

	log.Debugf("delete functions for %v", uid)

	var wf model.Workflow

	wfdb, err := db.getWorkflowByUid(context.Background(), uid)
	if err != nil {
		return err
	}

	// no need to error check, it passed the save check
	wf.Load(wfdb.Workflow)
	namespace := wfdb.Edges.Namespace.ID

	log.Debugf("delete functions for %v, %s", wfdb.Name, namespace)

	for _, f := range wf.GetFunctions() {

		var ir isolateRequest

		ir.Workflow.Namespace = namespace
		ir.Workflow.Name = wf.Name
		ir.Workflow.ID = wf.ID

		ir.Container.Type = f.Type
		ir.Container.Image = f.Image
		ir.Container.Cmd = f.Cmd
		ir.Container.Size = f.Size
		ir.Container.Scale = f.Scale
		ir.Container.ID = f.ID

		ah, err := serviceToHash(&ir)
		if err != nil {
			return err
		}

		u := fmt.Sprintf(kubeAPIKServiceURL, os.Getenv(direktivWorkflowNamespace))
		url := fmt.Sprintf("%s/%s", u, fmt.Sprintf("%s-%s", namespace, ah))

		log.Debugf("deleting url %v", url)

		_, err = sendKuberequest(http.MethodDelete, url, nil)
		if err != nil {
			log.Errorf("can not delete function: %v", err)
		}

		// wait till the service is 100 percent gone
		// this is needed for the engine to create a new one
		// otherwise it might be in terminated stage and can get a request
		for {
			err := getKnativeFunction(url)
			log.Debugf("err while waiting: %v", err)
			if err != nil {
				break
			}
		}

	}

	return nil

}

func getKnativeFunction(svc string) error {

	u := fmt.Sprintf(kubeAPIKServiceURL, os.Getenv(direktivWorkflowNamespace))

	url := fmt.Sprintf("%s/%s", u, svc)
	resp, err := sendKuberequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("service does not exists")
	}

	return nil
}

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

func addKnativeFunction(ir *isolateRequest) error {

	log.Debugf("adding knative service")

	namespace := ir.Workflow.Namespace

	ah, err := serviceToHash(ir)
	if err != nil {
		return err
	}

	log.Debugf("adding knative service hash %v", ah)

	var (
		cpu float64
		mem int
	)

	switch ir.Container.Size {
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

	u := fmt.Sprintf(kubeAPIKServiceURL, os.Getenv(direktivWorkflowNamespace))

	// get imagePullSecrets
	secrets, err := kubernetesListRegistriesNames(namespace)
	if err != nil {
		return err
	}

	var sstrings []string
	for _, s := range secrets {
		sstrings = append(sstrings, fmt.Sprintf("{ \"name\": \"%s\"}", s))
	}

	cmd, err := cmdToCommand(ir.Container.Cmd)
	if err != nil {
		return err
	}

	svc := fmt.Sprintf(kubeReq.serviceTempl, fmt.Sprintf("%s-%s", namespace, ah), ir.Container.Scale,
		strings.Join(sstrings, ","),
		ir.Container.Image, cmd, cpu, fmt.Sprintf("%dM", mem), cpu*2, fmt.Sprintf("%dM", mem*2),
		kubeReq.sidecar)

	fmt.Printf("%v\n", svc)

	resp, err := sendKuberequest(http.MethodPost, u, bytes.NewBufferString(svc))
	if err != nil {
		log.Errorf("can not send kube request: %v", err)
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		b, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		return fmt.Errorf("can not add knative service: %v", string(b))
	}

	return nil

}

func sendKuberequest(method, url string, data io.Reader) (*http.Response, error) {

	if kubeReq.apiConfig == nil {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		rest.LoadTLSFiles(config)
		kubeReq.apiConfig = config
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(kubeReq.apiConfig.CAData)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    caCertPool,
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, data)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization",
		fmt.Sprintf("Bearer %s", kubeReq.apiConfig.BearerToken))

	return client.Do(req)

}

func k8sNamespace() string {
	return os.Getenv(k8sNamespaceVar)
}

func serviceToHash(ar *isolateRequest) (string, error) {

	h, err := hash.Hash(fmt.Sprintf("%s-%s-%s", ar.Workflow.Namespace,
		ar.Workflow.ID, ar.Container.ID), hash.FormatV2, nil)
	if err != nil {
		return "", err
	}

	suffix := fmt.Sprintf("-%d", h)
	maxLen := 64 - len(fmt.Sprintf("%s.%s", suffix, k8sNamespace())) - (len(ar.Workflow.Namespace) + 1)

	prefix := fmt.Sprintf("%s-%s", ar.Workflow.ID, ar.Container.ID)
	if len(prefix) > maxLen {
		prefix = prefix[:maxLen]
	}

	return fmt.Sprintf("%s%s", prefix, suffix), nil

}

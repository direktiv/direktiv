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
	"sync"

	hash "github.com/mitchellh/hashstructure/v2"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/model"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
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
	serviceAccountPrefix = "direktiv-sa"
	secretsPrefix        = "direktiv-secret"
)

type kubeRequest struct {
	serviceTempl string
	sidecar      string

	apiConfig *rest.Config
	mtx       sync.Mutex
}

var kubeReq = kubeRequest{}

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

func kubernetesActionServiceAccount(name string, create bool) error {

	clientset, kns, err := getClientSet()
	if err != nil {
		return err
	}

	log.Debugf("kubernetes service account: %s (create: %v)", name, create)
	sa := &v1.ServiceAccount{}
	sa.Name = fmt.Sprintf("%s-%s", serviceAccountPrefix, name)

	if !create {
		var opt metav1.GetOptions
		sa, err = clientset.CoreV1().ServiceAccounts(kns).Get(context.Background(), sa.Name, opt)
		if err != nil {
			return err
		}

		for _, ps := range sa.ImagePullSecrets {
			err = clientset.CoreV1().Secrets(kns).Delete(context.Background(), ps.Name, metav1.DeleteOptions{})
			if err != nil {
				// we can keep going
				log.Errorf("can not delete secret for sa: %v", err)
			}
		}
	}

	// we delete the account if it is there
	clientset.RbacV1().RoleBindings(kns).Delete(context.Background(), fmt.Sprintf("%s-binding",
		sa.Name), metav1.DeleteOptions{})
	err = clientset.CoreV1().ServiceAccounts(kns).Delete(context.Background(), sa.Name, metav1.DeleteOptions{})

	if create {

		sbj := rbac.Subject{
			Kind:      "ServiceAccount",
			Name:      sa.Name,
			Namespace: kns,
		}

		rb := &rbac.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-binding", sa.Name),
				Namespace: kns,
			},
			Subjects: []rbac.Subject{sbj},
			RoleRef: rbac.RoleRef{
				Kind: "Role",
				Name: "sidecar-role",
			},
		}

		_, err = clientset.CoreV1().ServiceAccounts(kns).Create(context.Background(), sa, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("can not create service account: %v", err)
			return err
		}

		_, err = clientset.RbacV1().RoleBindings(kns).Create(context.Background(), rb, metav1.CreateOptions{})

	}

	return err

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

			err = clientset.CoreV1().Secrets(kns).Delete(context.Background(), secretName, metav1.DeleteOptions{})
			if err != nil {
				return err
			}

			// detach it from service account
			return kubernetesSecretFromServiceAccount(namespace, secretName)

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
	if err != nil {
		return err
	}

	// attach to service account
	kubernetesSecretToServiceAccount(fmt.Sprintf("%s-%s", serviceAccountPrefix, namespace),
		fmt.Sprintf("%s-%s-%s", secretsPrefix, namespace, u.Hostname()))

	return err

}

// updating service account to
func kubernetesSecretFromServiceAccount(name, secret string) error {

	clientset, kns, err := getClientSet()
	if err != nil {
		return err
	}

	var opt metav1.GetOptions
	sa, err := clientset.CoreV1().ServiceAccounts(kns).Get(context.Background(), fmt.Sprintf("%s-%s", serviceAccountPrefix, name), opt)
	if err != nil {
		return err
	}

	// iterate and skip the secret we want to remove
	var r []v1.LocalObjectReference
	for _, ps := range sa.ImagePullSecrets {
		if ps.Name != secret {
			r = append(r, ps)
		}
	}

	sa.ImagePullSecrets = r

	_, err = clientset.CoreV1().ServiceAccounts(kns).Update(context.Background(), sa, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil

}

// updating service account to
func kubernetesSecretToServiceAccount(name, secret string) error {

	clientset, kns, err := getClientSet()
	if err != nil {
		return err
	}

	var opt metav1.GetOptions
	sa, err := clientset.CoreV1().ServiceAccounts(kns).Get(context.Background(), name, opt)
	if err != nil {
		return err
	}

	sa.ImagePullSecrets = append(sa.ImagePullSecrets, v1.LocalObjectReference{
		Name: secret,
	})

	_, err = clientset.CoreV1().ServiceAccounts(kns).Update(context.Background(), sa, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil

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
		url := fmt.Sprintf("%s/%s", u, fmt.Sprintf("%s-%d", namespace, ah))

		log.Debugf("deleting url %v", url)

		_, err = sendKuberequest(http.MethodDelete, url, nil)
		if err != nil {
			log.Errorf("can not delete function: %v", err)
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

	svc := fmt.Sprintf(kubeReq.serviceTempl, fmt.Sprintf("%s-%d", namespace, ah), ir.Container.Scale,
		fmt.Sprintf("%s-%s", serviceAccountPrefix, namespace),
		ir.Container.Image, cpu, fmt.Sprintf("%dM", mem), cpu*2, fmt.Sprintf("%dM", mem*2),
		kubeReq.sidecar)

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
				RootCAs: caCertPool,
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

func serviceToHash(ar *isolateRequest) (uint64, error) {

	return hash.Hash(fmt.Sprintf("%s-%s-%s", ar.Workflow.Namespace,
		ar.Workflow.ID, ar.Container.ID), hash.FormatV2, nil)

}

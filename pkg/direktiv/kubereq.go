package direktiv

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/vorteil/direktiv/pkg/model"

	hash "github.com/mitchellh/hashstructure/v2"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	serviceAccountPrefix = "direktiv-sa"
	secretsPrefix        = "direktiv-secret"
)

func kubernetesListRegistries(namespace string) ([]string, error) {

	var registries []string

	clientset, kns, err := getClientSet()
	if err != nil {
		return registries, err
	}

	var lo metav1.ListOptions
	secrets, err := clientset.CoreV1().Secrets(kns).List(lo)
	if err != nil {
		return registries, err
	}

	for _, s := range secrets.Items {
		if s.Annotations[annotationNamespace] == namespace {
			registries = append(registries, s.Annotations[annotationURL])
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
		sa, err = clientset.CoreV1().ServiceAccounts(kns).Get(sa.Name, opt)
		if err != nil {
			return err
		}

		for _, ps := range sa.ImagePullSecrets {
			err = clientset.CoreV1().Secrets(kns).Delete(ps.Name, &metav1.DeleteOptions{})
			if err != nil {
				// we can keep going
				log.Errorf("can not delete secret for sa: %v", err)
			}
		}
	}

	// we delete the account if it is there
	err = clientset.CoreV1().ServiceAccounts(kns).Delete(sa.Name, nil)

	if create {
		_, err = clientset.CoreV1().ServiceAccounts(kns).Create(sa)
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
	secrets, err := clientset.CoreV1().Secrets(kns).List(lo)
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

			err = clientset.CoreV1().Secrets(kns).Delete(secretName, &metav1.DeleteOptions{})
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

	sa.Name = secretName
	sa.Data[".dockerconfigjson"] = data
	sa.Type = "kubernetes.io/dockerconfigjson"

	_, err = clientset.CoreV1().Secrets(kns).Create(sa)
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
	sa, err := clientset.CoreV1().ServiceAccounts(kns).Get(fmt.Sprintf("%s-%s", serviceAccountPrefix, name), opt)
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

	_, err = clientset.CoreV1().ServiceAccounts(kns).Update(sa)
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
	sa, err := clientset.CoreV1().ServiceAccounts(kns).Get(name, opt)
	if err != nil {
		return err
	}

	sa.ImagePullSecrets = append(sa.ImagePullSecrets, v1.LocalObjectReference{
		Name: secret,
	})

	_, err = clientset.CoreV1().ServiceAccounts(kns).Update(sa)
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

	kns := os.Getenv(flowNamespace)
	if kns == "" {
		kns = "default"
	}

	return clientset, kns, nil
}

func (is *ingressServer) deleteKnativeFunctions(uid string) error {

	if is.wfServer.config.MockupMode == 1 {
		return nil
	}

	var wf model.Workflow

	wfdb, err := is.wfServer.dbManager.getWorkflowByUid(context.Background(), uid)
	if err != nil {
		return err
	}

	// no need to error check, it passed the save check
	wf.Load(wfdb.Workflow)
	namespace := wfdb.Edges.Namespace.ID

	for _, f := range wf.GetFunctions() {

		ah, err := serviceToHash(namespace, f.Image, f.Cmd, f.Size)
		if err != nil {
			return err
		}

		svcName := fmt.Sprintf("%s-%d", namespace, ah)
		url := fmt.Sprintf("%s/%s", kubeAPIKServiceURL, svcName)

		err = is.sendKuberequest(http.MethodDelete, url, nil)
		if err != nil {
			return err
		}

	}

	return nil

}

func (is *ingressServer) addKnativeFunctions(namespace string, workflow *model.Workflow) error {

	if is.wfServer.config.MockupMode == 1 {
		return nil
	}

	for _, f := range workflow.GetFunctions() {

		ah, err := serviceToHash(namespace, f.Image, f.Cmd, f.Size)
		if err != nil {
			return err
		}

		log.Debugf("deleting isolate: %d", ah)

		var (
			cpu float64
			mem int
		)

		switch f.Size {
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

		svc := fmt.Sprintf(is.serviceTmpl, fmt.Sprintf("%s-%d", namespace, ah),
			fmt.Sprintf("%s-%s", serviceAccountPrefix, namespace),
			f.Image, cpu, fmt.Sprintf("%dM", mem), cpu*2, fmt.Sprintf("%dM", mem*2),
			is.wfServer.config.FlowAPI.Sidecar)

		err = is.sendKuberequest(http.MethodPost, kubeAPIKServiceURL,
			bytes.NewBufferString(svc))
		if err != nil {
			return err
		}

	}

	return nil

}

func (is *ingressServer) sendKuberequest(method, url string, data io.Reader) error {

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(is.kubeCA)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), method, url, data)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", string(is.kubeToken)))

	_, err := client.Do(req)
	return err

}

func serviceToHash(ns, img, cmd string, size model.Size) (uint64, error) {

	return hash.Hash(fmt.Sprintf("%s-%s-%s-%d", ns, img,
		cmd, size), hash.FormatV2, nil)

}

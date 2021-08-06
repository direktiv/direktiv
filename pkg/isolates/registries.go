package isolates

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	igrpc "github.com/vorteil/direktiv/pkg/isolates/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	secretsPrefix = "direktiv-secret"

	annotationNamespace = "direktiv.io/namespace"
	annotationURL       = "direktiv.io/url"
	annotationURLHash   = "direktiv.io/urlhash"
)

func getClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func kubernetesDeleteRegistry(name, namespace string) error {

	log.Debugf("deleting registry %s (%s)", name, namespace)

	clientset, err := getClientSet()
	if err != nil {
		return err
	}

	secrets, err := clientset.CoreV1().Secrets(isolateConfig.Namespace).
		List(context.Background(), metav1.ListOptions{})
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

			return clientset.CoreV1().Secrets(isolateConfig.Namespace).
				Delete(context.Background(), secretName, metav1.DeleteOptions{})

		}

	}

	return fmt.Errorf("no registry with name %s found", name)

}

func (is *isolateServer) DeleteRegistry(ctx context.Context, in *igrpc.DeleteRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty
	return &resp, kubernetesDeleteRegistry(in.GetName(), in.GetNamespace())
}

func (is *isolateServer) StoreRegistry(ctx context.Context, in *igrpc.StoreRegistryRequest) (*emptypb.Empty, error) {

	// create secret data, needs to be attached to service account
	userToken := strings.SplitN(string(in.Data), ":", 2)
	if len(userToken) != 2 {
		return nil, fmt.Errorf("invalid username/token format")
	}

	tmpl := `{
	"auths": {
		"%s": {
			"username": "%s",
			"password": "%s",
			"auth": "%s"
		}
	}
	}`

	auth := fmt.Sprintf(tmpl, in.GetName(), userToken[0], userToken[1],
		base64.StdEncoding.EncodeToString(in.Data))

	log.Debugf("adding secret %s (%s)", in.GetName(), in.GetNamespace())

	clientset, err := getClientSet()
	if err != nil {
		return &empty, err
	}

	// make sure it is URL format
	u, err := url.Parse(in.GetName())
	if err != nil {
		return &empty, err
	}

	secretName := fmt.Sprintf("%s-%s-%s", secretsPrefix, in.GetNamespace(), u.Hostname())

	kubernetesDeleteRegistry(in.GetName(), in.GetNamespace())

	sa := &v1.Secret{
		Data: make(map[string][]byte),
	}

	sa.Annotations = make(map[string]string)
	sa.Annotations[annotationNamespace] = in.GetNamespace()
	sa.Annotations[annotationURL] = in.GetName()
	sa.Annotations[annotationURLHash] = base64.StdEncoding.EncodeToString([]byte(in.GetName()))

	sa.Name = secretName
	sa.Data[".dockerconfigjson"] = []byte(auth)
	sa.Type = "kubernetes.io/dockerconfigjson"

	_, err = clientset.CoreV1().Secrets(isolateConfig.Namespace).Create(context.Background(),
		sa, metav1.CreateOptions{})

	return &empty, err

}

func (is *isolateServer) GetRegistries(ctx context.Context, in *igrpc.GetRegistriesRequest) (*igrpc.GetRegistriesResponse, error) {

	resp := &igrpc.GetRegistriesResponse{
		Registries: []*igrpc.GetRegistriesResponse_Registry{},
	}

	clientset, err := getClientSet()
	if err != nil {
		return resp, err
	}

	secrets, err := clientset.CoreV1().Secrets(isolateConfig.Namespace).
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return resp, err
	}

	for _, s := range secrets.Items {
		if s.Annotations[annotationNamespace] == in.GetNamespace() {
			u := s.Annotations[annotationURL]
			h := s.Annotations[annotationURLHash]
			resp.Registries = append(resp.Registries, &igrpc.GetRegistriesResponse_Registry{
				Name: &u,
				Id:   &h,
			})
		}
	}

	return resp, nil

}

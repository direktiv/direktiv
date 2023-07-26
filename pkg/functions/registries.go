package functions

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	hash "github.com/mitchellh/hashstructure/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	secretsPrefix = "direktiv-secret"

	annotationNamespace = "direktiv.io/namespace"
	annotationURL       = "direktiv.io/url"
	annotationURLHash   = "direktiv.io/urlhash"

	// Registry Types.
	annotationRegistryTypeKey            = "direktiv.io/registry-type"
	annotationRegistryTypeNamespaceValue = "namespace"
	annotationRegistryObfuscatedUser     = "direktiv.io/obf-user"
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

func kubernetesDeleteRegistry(ctx context.Context, name, namespace string) error {
	logger.Debugf("deleting registry %s (%s)", name, namespace)

	clientset, err := getClientSet()
	if err != nil {
		return err
	}

	fo := make(map[string]string)
	fo[annotationNamespace] = namespace
	h, _ := hash.Hash(name, hash.FormatV2, nil)
	fo[annotationURLHash] = fmt.Sprintf("%d", h)

	lo := metav1.ListOptions{LabelSelector: labels.Set(fo).String()}
	secrets, err := clientset.CoreV1().Secrets(functionsConfig.Namespace).List(ctx, lo)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("could not retrieve registry: %s", err))
	}

	if len(secrets.Items) == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("registry '%s' does not exist", name))
	}

	return clientset.CoreV1().Secrets(functionsConfig.Namespace).Delete(ctx, secrets.Items[0].Name, metav1.DeleteOptions{})
}

func listRegistriesNames(namespace string) []string {
	logger.Debugf("getting registries for namespace %s", namespace)
	var registries []string

	clientset, err := getClientSet()
	if err != nil {
		logger.Errorf("can not get clientset: %v", err)
		return registries
	}

	secrets, err := clientset.CoreV1().Secrets(functionsConfig.Namespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationNamespace, namespace)})
	if err != nil {
		logger.Errorf("can not list secrets: %v", err)
		return registries
	}

	for _, s := range secrets.Items {
		registries = append(registries, s.Name)
	}

	logger.Debugf("registries for namespace: %+v", registries)

	return registries
}

func (is *functionsServer) DeleteRegistry(ctx context.Context, in *igrpc.FunctionsDeleteRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty
	return &resp, kubernetesDeleteRegistry(ctx, in.GetName(), in.GetNamespace())
}

func (is *functionsServer) StoreRegistry(ctx context.Context, in *igrpc.FunctionsStoreRegistryRequest) (*emptypb.Empty, error) {
	// create secret data, needs to be attached to service account
	userToken := strings.SplitN(string(in.Data), ":", 2)
	if len(userToken) != 2 {
		logger.Errorf("invalid username/token format for registry")
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

	logger.Debugf("adding secret %s (%s)", in.GetName(), in.GetNamespace())

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

	err = kubernetesDeleteRegistry(ctx, in.GetName(), in.GetNamespace())
	if err != nil {
		// delete the old registry is just a safety measure
		logger.Debugf("ignoring error")
	}

	sa := prepareNewRegistrySecret(secretName, in.GetName(), auth)

	sa.Labels[annotationNamespace] = in.GetNamespace()
	sa.Labels[annotationRegistryTypeKey] = annotationRegistryTypeNamespaceValue

	sa.Annotations[annotationRegistryObfuscatedUser] = obfuscateUser(userToken[0])

	_, err = clientset.CoreV1().Secrets(functionsConfig.Namespace).Create(context.Background(),
		&sa, metav1.CreateOptions{})

	return &empty, err
}

func (is *functionsServer) GetRegistries(ctx context.Context, in *igrpc.FunctionsGetRegistriesRequest) (*igrpc.FunctionsGetRegistriesResponse, error) {
	resp := &igrpc.FunctionsGetRegistriesResponse{
		Registries: []*igrpc.FunctionsRegistry{},
	}

	clientset, err := getClientSet()
	if err != nil {
		return resp, err
	}

	secrets, err := clientset.CoreV1().Secrets(functionsConfig.Namespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationNamespace, in.GetNamespace())})
	if err != nil {
		return resp, err
	}

	for _, s := range secrets.Items {
		u := s.Annotations[annotationURL]
		h := s.Annotations[annotationURLHash]
		user := s.Annotations[annotationRegistryObfuscatedUser]
		resp.Registries = append(resp.Registries, &igrpc.FunctionsRegistry{
			Name: &u,
			Id:   &h,
			User: &user,
		})
	}

	return resp, nil
}

func obfuscateUser(user string) string {
	switch len(user) {
	case 1, 2, 3:
		user = fmt.Sprintf("%s***", string(user[0]))
	case 4, 5:
		user = fmt.Sprintf("%s***%s", string(user[0]), string(user[len(user)-1]))
	default:
		user = fmt.Sprintf("%s***%s", user[:2], user[len(user)-2:])
	}

	return user
}

func prepareNewRegistrySecret(name, url, authConfig string) v1.Secret {
	sa := v1.Secret{
		Data: make(map[string][]byte),
	}

	sa.Labels = make(map[string]string)

	h, err := hash.Hash(url, hash.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	sa.Labels[annotationURLHash] = fmt.Sprintf("%d", h)

	sa.Annotations = make(map[string]string)
	sa.Annotations[annotationURL] = url
	sa.Annotations[annotationURLHash] = base64.StdEncoding.EncodeToString([]byte(url))

	sa.Name = name
	sa.Data[".dockerconfigjson"] = []byte(authConfig)
	sa.Type = "kubernetes.io/dockerconfigjson"

	return sa
}

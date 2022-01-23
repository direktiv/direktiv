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
	secretsPrefix              = "direktiv-secret"
	secretsGlobalPrefix        = "direktiv-global-secret"
	secretsGlobalPrivatePrefix = "direktiv-global-private-secret"

	annotationNamespace = "direktiv.io/namespace"
	annotationURL       = "direktiv.io/url"
	annotationURLHash   = "direktiv.io/urlhash"

	// Registry Types
	annotationRegistryTypeKey                = "direktiv.io/registry-type"
	annotationRegistryTypeGlobalValue        = "global"
	annotationRegistryTypeGlobalPrivateValue = "global-private"
	annotationRegistryTypeNamespaceValue     = "namespace"
	annotationRegistryObfuscatedUser         = "direktiv.io/obf-user"
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

func kubernetesDeleteGlobalRegistry(ctx context.Context, name, globalAnnotation string) error {

	logger.Debugf("deleting global registry %s (%s)", name, globalAnnotation)

	clientset, err := getClientSet()
	if err != nil {
		return err
	}

	fo := make(map[string]string)
	fo[annotationRegistryTypeKey] = globalAnnotation
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

func listGlobalRegistriesNames() []string {

	logger.Debugf("getting public global registries")
	var registries []string

	clientset, err := getClientSet()
	if err != nil {
		logger.Errorf("can not get clientset: %v", err)
		return registries
	}

	secrets, err := clientset.CoreV1().Secrets(functionsConfig.Namespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationRegistryTypeKey, annotationRegistryTypeGlobalValue)})
	if err != nil {
		logger.Errorf("can not list secrets: %v", err)
		return registries
	}

	for _, s := range secrets.Items {
		registries = append(registries, s.Name)
	}

	logger.Debugf("public global registries : %+v", registries)

	return registries

}

func listGlobalPrivateRegistriesNames() []string {

	logger.Debugf("getting private global registries")
	var registries []string

	clientset, err := getClientSet()
	if err != nil {
		logger.Errorf("can not get clientset: %v", err)
		return registries
	}

	secrets, err := clientset.CoreV1().Secrets(functionsConfig.Namespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationRegistryTypeKey, annotationRegistryTypeGlobalPrivateValue)})
	if err != nil {
		logger.Errorf("can not list secrets: %v", err)
		return registries
	}

	for _, s := range secrets.Items {
		registries = append(registries, s.Name)
	}

	logger.Debugf("private global registries : %+v", registries)

	return registries

}

// namespace
func (is *functionsServer) DeleteRegistry(ctx context.Context, in *igrpc.DeleteRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty
	return &resp, kubernetesDeleteRegistry(ctx, in.GetName(), in.GetNamespace())
}

func (is *functionsServer) StoreRegistry(ctx context.Context, in *igrpc.StoreRegistryRequest) (*emptypb.Empty, error) {

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

	kubernetesDeleteRegistry(ctx, in.GetName(), in.GetNamespace())

	sa := prepareNewRegistrySecret(secretName, in.GetName(), auth)
	sa.Labels[annotationNamespace] = in.GetNamespace()
	sa.Labels[annotationRegistryTypeKey] = annotationRegistryTypeNamespaceValue

	// var un string
	// ut := userToken[0]
	// switch len(ut) {
	// case 1, 2, 3:
	// 	un = fmt.Sprintf("%s***", string(ut[0]))
	// case 4, 5:
	// 	un = fmt.Sprintf("%s***%s", string(ut[0]), string(ut[len(ut)-1]))
	// default:
	// 	un = fmt.Sprintf("%s***%s", ut[:2], ut[len(ut)-2:])
	// }

	sa.Annotations[annotationRegistryObfuscatedUser] = obfuscateUser(userToken[0])

	_, err = clientset.CoreV1().Secrets(functionsConfig.Namespace).Create(context.Background(),
		&sa, metav1.CreateOptions{})

	return &empty, err

}

func (is *functionsServer) GetRegistries(ctx context.Context, in *igrpc.GetRegistriesRequest) (*igrpc.GetRegistriesResponse, error) {

	resp := &igrpc.GetRegistriesResponse{
		Registries: []*igrpc.Registry{},
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
		resp.Registries = append(resp.Registries, &igrpc.Registry{
			Name: &u,
			Id:   &h,
			User: &user,
		})
	}

	return resp, nil

}

// global
func (is *functionsServer) DeleteGlobalRegistry(ctx context.Context, in *igrpc.DeleteGlobalRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty
	return &resp, kubernetesDeleteGlobalRegistry(ctx, in.GetName(), annotationRegistryTypeGlobalValue)
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

func (is *functionsServer) StoreGlobalRegistry(ctx context.Context, in *igrpc.StoreGlobalRegistryRequest) (*emptypb.Empty, error) {

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

	logger.Debugf("adding global secret %s", in.GetName())

	clientset, err := getClientSet()
	if err != nil {
		return &empty, err
	}

	// make sure it is URL format
	u, err := url.Parse(in.GetName())
	if err != nil {
		return &empty, err
	}

	secretName := fmt.Sprintf("%s-%s", secretsGlobalPrefix, u.Hostname())

	kubernetesDeleteGlobalRegistry(ctx, in.GetName(), annotationRegistryTypeGlobalValue)

	sa := prepareNewRegistrySecret(secretName, in.GetName(), auth)
	sa.Labels[annotationRegistryTypeKey] = annotationRegistryTypeGlobalValue

	sa.Annotations[annotationRegistryObfuscatedUser] = obfuscateUser(userToken[0])

	_, err = clientset.CoreV1().Secrets(functionsConfig.Namespace).Create(context.Background(),
		&sa, metav1.CreateOptions{})

	return &empty, err

}

func (is *functionsServer) GetGlobalRegistries(ctx context.Context, in *emptypb.Empty) (*igrpc.GetRegistriesResponse, error) {

	resp := &igrpc.GetRegistriesResponse{
		Registries: []*igrpc.Registry{},
	}

	clientset, err := getClientSet()
	if err != nil {
		return resp, err
	}

	secrets, err := clientset.CoreV1().Secrets(functionsConfig.Namespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationRegistryTypeKey, annotationRegistryTypeGlobalValue)})
	if err != nil {
		return resp, err
	}

	for _, s := range secrets.Items {
		u := s.Annotations[annotationURL]
		h := s.Annotations[annotationURLHash]
		user := s.Annotations[annotationRegistryObfuscatedUser]
		resp.Registries = append(resp.Registries, &igrpc.Registry{
			Name: &u,
			Id:   &h,
			User: &user,
		})
	}

	return resp, nil

}

// global-private

func (is *functionsServer) DeleteGlobalPrivateRegistry(ctx context.Context, in *igrpc.DeleteGlobalRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty
	return &resp, kubernetesDeleteGlobalRegistry(ctx, in.GetName(), annotationRegistryTypeGlobalPrivateValue)
}

func (is *functionsServer) StoreGlobalPrivateRegistry(ctx context.Context, in *igrpc.StoreGlobalRegistryRequest) (*emptypb.Empty, error) {

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

	logger.Debugf("adding private global secret %s", in.GetName())

	clientset, err := getClientSet()
	if err != nil {
		return &empty, err
	}

	// make sure it is URL format
	u, err := url.Parse(in.GetName())
	if err != nil {
		return &empty, err
	}

	secretName := fmt.Sprintf("%s-%s", secretsGlobalPrivatePrefix, u.Hostname())

	kubernetesDeleteGlobalRegistry(ctx, in.GetName(), annotationRegistryTypeGlobalPrivateValue)

	sa := prepareNewRegistrySecret(secretName, in.GetName(), auth)
	sa.Labels[annotationRegistryTypeKey] = annotationRegistryTypeGlobalPrivateValue

	sa.Annotations[annotationRegistryObfuscatedUser] = obfuscateUser(userToken[0])

	_, err = clientset.CoreV1().Secrets(functionsConfig.Namespace).Create(context.Background(),
		&sa, metav1.CreateOptions{})

	return &empty, err

}

func (is *functionsServer) GetGlobalPrivateRegistries(ctx context.Context, in *emptypb.Empty) (*igrpc.GetRegistriesResponse, error) {

	resp := &igrpc.GetRegistriesResponse{
		Registries: []*igrpc.Registry{},
	}

	clientset, err := getClientSet()
	if err != nil {
		return resp, err
	}

	secrets, err := clientset.CoreV1().Secrets(functionsConfig.Namespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationRegistryTypeKey, annotationRegistryTypeGlobalPrivateValue)})
	if err != nil {
		return resp, err
	}

	for _, s := range secrets.Items {
		u := s.Annotations[annotationURL]
		h := s.Annotations[annotationURLHash]
		user := s.Annotations[annotationRegistryObfuscatedUser]
		resp.Registries = append(resp.Registries, &igrpc.Registry{
			Name: &u,
			Id:   &h,
			User: &user,
		})
	}

	return resp, nil
}

// util

func prepareNewRegistrySecret(name, url, authConfig string) v1.Secret {
	sa := v1.Secret{
		Data: make(map[string][]byte),
	}

	sa.Labels = make(map[string]string)

	h, _ := hash.Hash(fmt.Sprintf("%s", url), hash.FormatV2, nil)
	sa.Labels[annotationURLHash] = fmt.Sprintf("%d", h)

	sa.Annotations = make(map[string]string)
	sa.Annotations[annotationURL] = url
	sa.Annotations[annotationURLHash] = base64.StdEncoding.EncodeToString([]byte(url))

	sa.Name = name
	sa.Data[".dockerconfigjson"] = []byte(authConfig)
	sa.Type = "kubernetes.io/dockerconfigjson"

	return sa
}

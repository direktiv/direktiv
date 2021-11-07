package functions

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	hash "github.com/mitchellh/hashstructure/v2"
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

	// Registry Types
	annotationRegistryTypeKey                = "direktiv.io/registry-type"
	annotationRegistryTypeGlobalValue        = "global"
	annotationRegistryTypeGlobalPrivateValue = "global-private"
	annotationRegistryTypeNamespaceValue     = "namespace"
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

	logger.Debugf("deleting registry %s (%s)", name, namespace)

	clientset, err := getClientSet()
	if err != nil {
		return err
	}

	fo := make(map[string]string)
	fo[annotationNamespace] = namespace
	h, _ := hash.Hash(fmt.Sprintf("%s", name), hash.FormatV2, nil)
	fo[annotationURLHash] = fmt.Sprintf("%d", h)

	lo := metav1.ListOptions{LabelSelector: labels.Set(fo).String()}
	return clientset.CoreV1().Secrets(functionsConfig.Namespace).
		DeleteCollection(context.Background(), metav1.DeleteOptions{}, lo)

}

func listRegistriesNames(namespace string, includeGlobal bool) []string {

	logger.With("includeGlobal", includeGlobal).Debugf("getting registries for namespace %s", namespace)
	var registries []string

	clientset, err := getClientSet()
	if err != nil {
		logger.Errorf("can not get clientset: %v", err)
		return registries
	}

	annotations := map[string]string{
		annotationNamespace: namespace,
	}

	// Add public global registries
	if includeGlobal {
		annotations[annotationRegistryTypeKey] = annotationRegistryTypeGlobalValue
	}

	secrets, err := clientset.CoreV1().Secrets(functionsConfig.Namespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: labels.Set(annotations).String()})
	if err != nil {
		logger.Errorf("can not list secrets: %v", err)
		return registries
	}

	removeDuplicateRegistries(secrets)

	for _, s := range secrets.Items {
		registries = append(registries, s.Name)
	}

	logger.Debugf("registries for namespace: %+v", registries)

	return registries

}

func (is *functionsServer) DeleteRegistry(ctx context.Context, in *igrpc.DeleteRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty
	return &resp, kubernetesDeleteRegistry(in.GetName(), in.GetNamespace())
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

	kubernetesDeleteRegistry(in.GetName(), in.GetNamespace())

	sa := &v1.Secret{
		Data: make(map[string][]byte),
	}

	sa.Labels = make(map[string]string)
	sa.Labels[annotationNamespace] = in.GetNamespace()

	h, _ := hash.Hash(fmt.Sprintf("%s", in.GetName()), hash.FormatV2, nil)
	sa.Labels[annotationURLHash] = fmt.Sprintf("%d", h)

	sa.Annotations = make(map[string]string)
	sa.Annotations[annotationURL] = in.GetName()
	sa.Annotations[annotationURLHash] = base64.StdEncoding.EncodeToString([]byte(in.GetName()))

	sa.Name = secretName
	sa.Data[".dockerconfigjson"] = []byte(auth)
	sa.Type = "kubernetes.io/dockerconfigjson"

	_, err = clientset.CoreV1().Secrets(functionsConfig.Namespace).Create(context.Background(),
		sa, metav1.CreateOptions{})

	return &empty, err

}

func (is *functionsServer) GetRegistries(ctx context.Context, in *igrpc.GetRegistriesRequest) (*igrpc.GetRegistriesResponse, error) {

	resp := &igrpc.GetRegistriesResponse{
		Registries: []*igrpc.GetRegistriesResponse_Registry{},
	}

	clientset, err := getClientSet()
	if err != nil {
		return resp, err
	}

	annotations := map[string]string{
		annotationNamespace: in.GetNamespace(),
	}

	// Add public global registries
	if in.GetIncludeGlobal() {
		annotations[annotationRegistryTypeKey] = annotationRegistryTypeGlobalValue
	}

	secrets, err := clientset.CoreV1().Secrets(functionsConfig.Namespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationNamespace)})
	if err != nil {
		return resp, err
	}

	removeDuplicateRegistries(secrets)

	for _, s := range secrets.Items {
		u := s.Annotations[annotationURL]
		h := s.Annotations[annotationURLHash]
		resp.Registries = append(resp.Registries, &igrpc.GetRegistriesResponse_Registry{
			Name: &u,
			Id:   &h,
		})
	}

	return resp, nil

}

// global

// util
func removeDuplicateRegistries(secretList *v1.SecretList) {

	secretsMap := make(map[string]v1.Secret, 0)

	for i := range secretList.Items {
		secretURL := secretList.Items[i].Annotations[annotationURL]

		if s, ok := secretsMap[secretURL]; ok {
			// Replace Global Public registries as they have lowest priority
			if s.Annotations[annotationRegistryTypeKey] == annotationRegistryTypeGlobalValue {
				secretsMap[secretURL] = secretList.Items[i]
			}
		} else {
			// Add secret to map
			secretsMap[secretURL] = secretList.Items[i]
		}
	}

	secretList.Items = make([]v1.Secret, len(secretsMap))

	// Replace secret items
	for _, secret := range secretsMap {
		secretList.Items = append(secretList.Items, secret)
	}
}

// func listGlobalRegistriesNames() []string {

// 	logger.Debugf("getting global registries")
// 	var registries []string

// 	clientset, err := getClientSet()
// 	if err != nil {
// 		logger.Errorf("can not get clientset: %v", err)
// 		return registries
// 	}

// 	secrets, err := clientset.CoreV1().Secrets(functionsConfig.Namespace).
// 		List(context.Background(),
// 			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationScope, "global")})
// 	if err != nil {
// 		logger.Errorf("can not list secrets: %v", err)
// 		return registries
// 	}

// 	for _, s := range secrets.Items {
// 		registries = append(registries, s.Name)
// 	}

// 	logger.Debugf("registries global namespace: %+v", registries)

// 	return registries

// }

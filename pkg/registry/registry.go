package registry

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"sync"

	"github.com/direktiv/direktiv/pkg/core"
	dReg "github.com/docker/docker/api/types/registry"
	dClient "github.com/docker/docker/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	annotationNamespace    = "direktiv.io/namespace"
	annotationRegistryURL  = "direktiv.io/registry_url"
	annotationRegistryUser = "direktiv.io/registry_user"
)

type kManager struct {
	*kubernetes.Clientset
	K8sNamespace string
}

func (c *kManager) ListRegistries(namespace string) ([]*core.Registry, error) {
	result := []*core.Registry{}

	secrets, err := c.Clientset.CoreV1().Secrets(c.K8sNamespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationNamespace, namespace)})
	if err != nil {
		return nil, err
	}

	for _, s := range secrets.Items {
		u := s.Annotations[annotationRegistryURL]
		user := s.Annotations[annotationRegistryUser]
		result = append(result, &core.Registry{
			Namespace: namespace,
			ID:        s.Name,
			URL:       u,
			User:      user,
			CreatedAt: s.GetCreationTimestamp().Time,
		})
	}

	// Sort registries by CreatedAt (asc order)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result, nil
}

func (c *kManager) DeleteRegistry(namespace string, id string) error {
	secrets, err := c.Clientset.CoreV1().Secrets(c.K8sNamespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationNamespace, namespace)})
	if err != nil {
		return err
	}

	for _, s := range secrets.Items {
		if s.Name == id {
			return c.Clientset.CoreV1().Secrets(c.K8sNamespace).Delete(context.Background(), s.Name, metav1.DeleteOptions{})
		}
	}

	return core.ErrNotFound
}

func (c *kManager) DeleteNamespace(namespace string) error {
	secrets, err := c.Clientset.CoreV1().Secrets(c.K8sNamespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationNamespace, namespace)})
	if err != nil {
		return fmt.Errorf("k8s secrets list: %w", err)
	}
	if len(secrets.Items) == 0 {
		return core.ErrNotFound
	}

	for _, s := range secrets.Items {
		err = c.Clientset.CoreV1().Secrets(c.K8sNamespace).Delete(context.Background(), s.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("k8s secrets delete: %w", err)
		}
	}

	return core.ErrNotFound
}

func (c *kManager) StoreRegistry(registry *core.Registry) (*core.Registry, error) {
	str := fmt.Sprintf("%s-%s", registry.Namespace, registry.URL)
	sh := sha256.Sum256([]byte(str))
	id := fmt.Sprintf("secret-%x", sh[:10])
	registry.ID = id

	// delete the old registry is just a safety measure
	_ = c.DeleteRegistry(registry.Namespace, id)

	s, err := buildSecret(*registry)
	if err != nil {
		return nil, err
	}
	_, err = c.Clientset.CoreV1().Secrets(c.K8sNamespace).Create(context.Background(),
		s, metav1.CreateOptions{})

	registry.Password = ""

	return registry, err
}

func testLogin(registry *core.Registry) error {
	cli, err := dClient.NewClientWithOpts(dClient.WithHost(registry.URL))
	if err != nil {
		return err
	}

	authConfig := dReg.AuthConfig{
		Username:      registry.URL,
		Password:      registry.Password,
		ServerAddress: registry.URL,
	}
	_, err = cli.RegistryLogin(context.Background(), authConfig)

	return err
}

func (c *kManager) TestLogin(registry *core.Registry) error {
	return testLogin(registry)
}

func buildSecret(registry core.Registry) (*v1.Secret, error) {
	_, err := url.Parse(registry.URL)
	if err != nil {
		return nil, err
	}

	auth := fmt.Sprintf(`{
	"auths": {
		"%s": {
			"username": "%s",
			"password": "%s",
			"auth": "%s"
		}
	}
	}`,
		registry.URL,
		registry.User,
		registry.Password,
		base64.StdEncoding.EncodeToString([]byte(registry.User+":"+registry.Password)))

	s := v1.Secret{
		Data: make(map[string][]byte),
	}

	s.Labels = make(map[string]string)
	s.Labels[annotationNamespace] = registry.Namespace

	s.Annotations = make(map[string]string)
	s.Annotations[annotationRegistryURL] = registry.URL
	s.Annotations[annotationRegistryUser] = registry.User

	s.Name = registry.ID
	s.Data[".dockerconfigjson"] = []byte(auth)
	s.Type = "kubernetes.io/dockerconfigjson"

	return &s, nil
}

func NewManager(mocked bool) (core.RegistryManager, error) {
	if mocked {
		return &mockedManager{
			lock: &sync.Mutex{},
			list: make(map[string][]*core.Registry),
		}, nil
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	cSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &kManager{
		K8sNamespace: "direktiv-services-direktiv",
		Clientset:    cSet,
	}, nil
}

var _ core.RegistryManager = &kManager{}

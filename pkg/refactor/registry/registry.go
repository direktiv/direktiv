// nolint
package registry

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"sync"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	dReg "github.com/docker/docker/api/types/registry"
	dClient "github.com/docker/docker/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	annotationNamespace    = "direktiv.io/namespace"
	annotationRegistryURL  = "direktiv.io/registry/url"
	annotationRegistryUser = "direktiv.io/registry/user"
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
		})
	}

	return result, nil
}

func (c *kManager) DeleteRegistry(namespace string, id string) error {
	lo := metav1.ListOptions{LabelSelector: labels.Set(map[string]string{
		annotationNamespace: namespace,
	}).String()}

	secrets, err := c.Clientset.CoreV1().Secrets(c.K8sNamespace).List(context.Background(), lo)
	if err != nil {
		return fmt.Errorf("k8s list secrets: %s", err)
	}

	for _, s := range secrets.Items {
		if s.Name == id {
			return c.Clientset.CoreV1().Secrets(c.K8sNamespace).Delete(context.Background(), secrets.Items[0].Name, metav1.DeleteOptions{})
		}
	}

	return core.ErrNotFound
}

func (c *kManager) StoreRegistry(registry *core.Registry) (*core.Registry, error) {
	str := fmt.Sprintf("%s-%s", registry.Namespace, registry.URL)
	sh := sha256.Sum256([]byte(str))
	id := fmt.Sprintf("secret-%x", sh[:10])

	// delete the old registry is just a safety measure
	_ = c.DeleteRegistry(registry.Namespace, registry.ID)

	r := &core.Registry{
		Namespace: registry.Namespace,
		ID:        id,
		URL:       registry.URL,
		User:      obfuscateUser(registry.User),
		Password:  registry.Password,
	}

	s, err := buildSecret(registry)
	if err != nil {
		return nil, err
	}
	_, err = c.Clientset.CoreV1().Secrets(c.K8sNamespace).Create(context.Background(),
		s, metav1.CreateOptions{})

	return r, err
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

func buildSecret(registry *core.Registry) (*v1.Secret, error) {
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

// nolint
package registry

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"

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

var ErrNotFound = errors.New("ErrNotFound")

type Registry struct {
	Namespace string `json:"namespace"`
	ID        string `json:"id"`
	Url       string `json:"url"`
	User      string `json:"user"`
	Password  string `json:"password,omitempty"`
}

type Manager interface {
	ListRegistries(namespace string) ([]*Registry, error)
	DeleteRegistry(namespace string, id string) error
	StoreRegistry(registry *Registry) (*Registry, error)
}

type kManager struct {
	*kubernetes.Clientset
	K8sNamespace string
}

func (c *kManager) ListRegistries(namespace string) ([]*Registry, error) {
	result := []*Registry{}

	secrets, err := c.Clientset.CoreV1().Secrets(c.K8sNamespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationNamespace, namespace)})
	if err != nil {
		return nil, err
	}

	for _, s := range secrets.Items {
		u := s.Annotations[annotationRegistryURL]
		user := s.Annotations[annotationRegistryUser]
		result = append(result, &Registry{
			Namespace: namespace,
			ID:        s.Name,
			Url:       u,
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

	return ErrNotFound
}

func (c *kManager) StoreRegistry(registry *Registry) (*Registry, error) {
	// delete the old registry is just a safety measure
	_ = c.DeleteRegistry(registry.Namespace, registry.ID)

	str := fmt.Sprintf("%s-%s", registry.Namespace, registry.Url)
	sh := sha256.Sum256([]byte(str))
	id := fmt.Sprintf("secret-%x", sh[:10])

	r := &Registry{
		Namespace: registry.Namespace,
		ID:        id,
		Url:       registry.Url,
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

func buildSecret(registry *Registry) (*v1.Secret, error) {
	_, err := url.Parse(registry.Url)
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
		registry.Url,
		registry.User,
		registry.Password,
		base64.StdEncoding.EncodeToString([]byte(registry.User+":"+registry.Password)))

	s := v1.Secret{
		Data: make(map[string][]byte),
	}

	s.Labels = make(map[string]string)

	s.Labels[annotationNamespace] = registry.Namespace

	s.Annotations = make(map[string]string)
	s.Annotations[annotationRegistryURL] = registry.Url
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

func NewManager() (*kManager, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("error cluster config: %v\n", err)
		return nil, err
	}

	cSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error cluster config: %v\n", err)
		return nil, err
	}

	return &kManager{
		K8sNamespace: "direktiv-services-direktiv",
		Clientset:    cSet,
	}, nil
}

var _ Manager = &kManager{}

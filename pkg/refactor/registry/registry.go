// nolint
package registry

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	hash "github.com/mitchellh/hashstructure/v2"
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

type Registry struct {
	Name string
	ID   string
	User string
}

type Manager interface {
	ListRegistries(namespace string) ([]*Registry, error)
	DeleteRegistry(namespace string, name string) error
	StoreRegistry(namespace string, name string, data string) error
}

type client struct {
	c            *kubernetes.Clientset
	K8sNamespace string
}

func (c *client) ListRegistries(namespace string) ([]*Registry, error) {
	result := []*Registry{}

	secrets, err := c.c.CoreV1().Secrets(c.K8sNamespace).
		List(context.Background(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationNamespace, namespace)})
	if err != nil {
		return nil, err
	}

	for _, s := range secrets.Items {
		u := s.Annotations[annotationURL]
		h := s.Annotations[annotationURLHash]
		user := s.Annotations[annotationRegistryObfuscatedUser]
		result = append(result, &Registry{
			Name: u,
			ID:   h,
			User: user,
		})
	}

	return result, nil
}

func (c *client) DeleteRegistry(namespace string, name string) error {
	fo := make(map[string]string)
	fo[annotationNamespace] = namespace
	h, _ := hash.Hash(name, hash.FormatV2, nil)
	fo[annotationURLHash] = fmt.Sprintf("%d", h)

	lo := metav1.ListOptions{LabelSelector: labels.Set(fo).String()}
	secrets, err := c.c.CoreV1().Secrets(c.K8sNamespace).List(context.Background(), lo)
	if err != nil {
		return fmt.Errorf("k8s list secrets: %s", err)
	}

	if len(secrets.Items) == 0 {
		return fmt.Errorf("registry '%s' does not exist", name)
	}

	return c.c.CoreV1().Secrets(c.K8sNamespace).Delete(context.Background(), secrets.Items[0].Name, metav1.DeleteOptions{})
}

func (c *client) StoreRegistry(namespace string, name string, data string) error {
	// create secret data, needs to be attached to service account
	userToken := strings.SplitN(data, ":", 2)
	if len(userToken) != 2 {
		return fmt.Errorf("invalid username/token format")
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

	auth := fmt.Sprintf(tmpl, name, userToken[0], userToken[1],
		base64.StdEncoding.EncodeToString([]byte(data)))

	// make sure it is URL format
	u, err := url.Parse(name)
	if err != nil {
		return err
	}

	secretName := fmt.Sprintf("%s-%s-%s", secretsPrefix, namespace, u.Hostname())

	// delete the old registry is just a safety measure
	_ = c.DeleteRegistry(namespace, name)

	sa := buildSecret(secretName, name, auth)

	sa.Labels[annotationNamespace] = namespace
	sa.Labels[annotationRegistryTypeKey] = annotationRegistryTypeNamespaceValue

	sa.Annotations[annotationRegistryObfuscatedUser] = obfuscateUser(userToken[0])

	_, err = c.c.CoreV1().Secrets(c.K8sNamespace).Create(context.Background(),
		&sa, metav1.CreateOptions{})

	return err
}

func buildSecret(name, url, authConfig string) v1.Secret {
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

var _ Manager = &client{}

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

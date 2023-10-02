// nolint
package function

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

const (
	httpsProxy = "HTTPS_PROXY"
	httpProxy  = "HTTP_PROXY"
	noProxy    = "NO_PROXY"

	containerUser    = "direktiv-container"
	containerSidecar = "direktiv-sidecar"
)

type client interface {
	createService(cfg *FunctionDefination) error
	updateService(id string, cfg *FunctionDefination) error
	deleteService(id string) error
	listServices() ([]FunctionStatus, error)
}

type ClientConfig struct {
	ServiceAccount string `yaml:"service-account"`
	Namespace      string `yaml:"namespace"`
}

type knClient struct {
	config *ClientConfig

	client versioned.Interface
}

func (c *knClient) createService(cfg *FunctionDefination) error {
	svcDef, err := buildService(c.config, cfg)
	if err != nil {
		return err
	}

	service, err := c.client.ServingV1().Services(c.config.Namespace).Create(context.Background(), svcDef, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("f2: err serving create: %v\n", err)
		return err
	}
	fmt.Printf("f2: serving create output: %v\n", service)
	return nil
}

func (c *knClient) updateService(id string, cfg *FunctionDefination) error {
	// TODO implement me
	panic("implement me updateService\n")
}

func (c *knClient) deleteService(id string) error {
	err := c.client.ServingV1().Services(c.config.Namespace).Delete(context.Background(), id, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("f2: err serving delete: %v\n", err)
		return err
	}
	return err
}

func (c *knClient) listServices() ([]FunctionStatus, error) {
	list, err := c.client.ServingV1().Services(c.config.Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("f2: err serving list: %v\n", err)
		return nil, err
	}

	result := []FunctionStatus{}
	for i := range list.Items {
		result = append(result, &K8sFunctionStatus{&list.Items[i]})
	}

	return result, nil
}

var _ client = &knClient{}

package function

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

type knativeClient struct {
	config *ClientConfig

	client versioned.Interface
}

func (c *knativeClient) createService(cfg *Config) error {
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

func (c *knativeClient) updateService(cfg *Config) error {
	svcDef, err := buildService(c.config, cfg)
	if err != nil {
		return err
	}

	service, err := c.client.ServingV1().Services(c.config.Namespace).Update(context.Background(), svcDef, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("f2: err serving update: %v\n", err)
		return err
	}
	fmt.Printf("f2: serving update output: %v\n", service)
	return nil
}

func (c *knativeClient) deleteService(id string) error {
	err := c.client.ServingV1().Services(c.config.Namespace).Delete(context.Background(), id, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("f2: err serving delete: %v\n", err)
		return err
	}
	return err
}

func (c *knativeClient) listServices() ([]Status, error) {
	list, err := c.client.ServingV1().Services(c.config.Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("f2: err serving list: %v\n", err)
		return nil, err
	}

	result := []Status{}
	for i := range list.Items {
		result = append(result, &knativeStatus{&list.Items[i]})
	}

	return result, nil
}

var _ client = &knativeClient{}

type knativeStatus struct {
	*servingv1.Service
}

func (r *knativeStatus) getConditions() any {
	return r.Status.Conditions
}

func (r *knativeStatus) getId() string {
	return r.Name
}

func (r *knativeStatus) getValueHash() string {
	return r.Annotations["direktiv.io/input_hash"]
}

var _ Status = &knativeStatus{}

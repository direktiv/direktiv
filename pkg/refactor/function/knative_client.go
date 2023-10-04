package function

import (
	"context"
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

	_, err = c.client.ServingV1().Services(c.config.Namespace).Create(context.Background(), svcDef, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *knativeClient) updateService(cfg *Config) error {
	svcDef, err := buildService(c.config, cfg)
	if err != nil {
		return err
	}
	input, err := json.Marshal(&svcDef)
	if err != nil {
		return err
	}
	_, err = c.client.ServingV1().Services(c.config.Namespace).Patch(context.Background(), cfg.getID(), types.MergePatchType, input, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *knativeClient) deleteService(id string) error {
	err := c.client.ServingV1().Services(c.config.Namespace).Delete(context.Background(), id, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *knativeClient) listServices() ([]Status, error) {
	list, err := c.client.ServingV1().Services(c.config.Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
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

func (r *knativeStatus) getID() string {
	return r.Name
}

func (r *knativeStatus) getValueHash() string {
	return r.Annotations["direktiv.io/inputHash"]
}

var _ Status = &knativeStatus{}

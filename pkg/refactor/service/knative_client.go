package service

import (
	"context"
	"encoding/json"
	"io"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

type knativeClient struct {
	config *core.Config

	k8sSet *kubernetes.Clientset

	client versioned.Interface
}

// nolint
func (c *knativeClient) streamServiceLogs(id string, podID string) (io.ReadCloser, error) {
	req := c.k8sSet.CoreV1().Pods(c.config.KnativeNamespace).GetLogs(podID, &v1.PodLogOptions{
		Container: "direktiv-container",
		Follow:    true,
	})

	logsStream, err := req.Stream(context.Background())
	if err != nil {
		return nil, err
	}

	return logsStream, nil
}

func (c *knativeClient) createService(cfg *core.ServiceConfig) error {
	svcDef, err := buildService(c.config, cfg)
	if err != nil {
		return err
	}

	_, err = c.client.ServingV1().Services(c.config.KnativeNamespace).Create(context.Background(), svcDef, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *knativeClient) updateService(cfg *core.ServiceConfig) error {
	svcDef, err := buildService(c.config, cfg)
	if err != nil {
		return err
	}
	input, err := json.Marshal(&svcDef)
	if err != nil {
		return err
	}
	_, err = c.client.ServingV1().Services(c.config.KnativeNamespace).Patch(context.Background(), cfg.GetID(), types.MergePatchType, input, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *knativeClient) deleteService(id string) error {
	err := c.client.ServingV1().Services(c.config.KnativeNamespace).Delete(context.Background(), id, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *knativeClient) listServices() ([]status, error) {
	list, err := c.client.ServingV1().Services(c.config.KnativeNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := []status{}
	for i := range list.Items {
		result = append(result, &knativeStatus{&list.Items[i]})
	}

	return result, nil
}

func (c *knativeClient) listServicePods(id string) (any, error) {
	lo := metav1.ListOptions{}
	l, err := c.k8sSet.CoreV1().Pods(c.config.KnativeNamespace).List(context.Background(), lo)
	if err != nil {
		return nil, err
	}

	type pod struct {
		ID string `json:"id"`
	}

	pods := []*pod{}
	for i := range l.Items {
		if l.Items[i].Labels["serving.knative.dev/service"] != id {
			continue
		}
		pods = append(pods, &pod{
			ID: l.Items[i].Name,
		})
	}

	return pods, nil
}

func (c *knativeClient) killService(id string) error {
	return c.k8sSet.CoreV1().Pods(c.config.KnativeNamespace).Delete(
		context.Background(),
		id,
		metav1.DeleteOptions{})
}

var _ client = &knativeClient{}

type knativeStatus struct {
	*servingv1.Service
}

func (r *knativeStatus) GetConditions() any {
	type condition struct {
		Type    string `json:"type"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	list := []condition{}

	for _, c := range r.Status.Conditions {
		list = append(list, condition{
			Type:    string(c.Type),
			Status:  string(c.Status),
			Message: c.Message,
		})
	}

	return list
}

func (r *knativeStatus) GetID() string {
	return r.Name
}

func (r *knativeStatus) GetValueHash() string {
	return r.Annotations["direktiv.io/inputHash"]
}

var _ status = &knativeStatus{}

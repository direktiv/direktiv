package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

type knativeClient struct {
	config *core.Config

	k8sCli     *kubernetes.Clientset
	knativeCli versioned.Interface
}

func (c *knativeClient) streamServiceLogs(_ string, podID string) (io.ReadCloser, error) {
	req := c.k8sCli.CoreV1().Pods(c.config.KnativeNamespace).GetLogs(podID, &coreV1.PodLogOptions{
		Container: "direktiv-container",
		Follow:    true,
	})

	logsStream, err := req.Stream(context.Background())
	if err != nil {
		return nil, err
	}

	return logsStream, nil
}

func (c *knativeClient) createService(sv *core.ServiceFileData) error {
	if sv.Image == "" {
		return errors.New("image field is empty or not set")
	}

	// Step1: prepare registry secrets
	var registrySecrets []coreV1.LocalObjectReference
	// xKnative
	// secrets, err := c.k8sCli.CoreV1().Secrets(c.config.KnativeNamespace).
	//	List(context.Background(),
	//		metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", annotationNamespace, sv.Namespace)})
	// if err != nil {
	//	return err
	//}
	// for _, s := range secrets.Items {
	//	registrySecrets = append(registrySecrets, v1.LocalObjectReference{
	//		Name: s.Name,
	//	})
	//}

	// Step2: build service object
	depDef, svcDef, err := buildService(c.config, sv, registrySecrets)
	if err != nil {
		return err
	}

	_, err = c.k8sCli.AppsV1().Deployments(c.config.KnativeNamespace).Create(context.Background(), depDef, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	_, err = c.k8sCli.CoreV1().Services(c.config.KnativeNamespace).Create(context.Background(), svcDef, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// xKnative
	// err = c.applyPatch(sv)
	// if err != nil {
	//	return fmt.Errorf("applying patch: %w", err)
	//}

	return nil
}

// xKnative
//
//nolint:unused
func (c *knativeClient) applyPatch(sv *core.ServiceFileData) error {
	pathWhiteList := []string{
		"/spec/template/metadata/labels",
		"/spec/template/metadata/annotations",
		"/spec/template/spec/affinity",
		"/spec/template/spec/securityContext",
		"/spec/template/spec/containers/0",
	}

	// check patch whitelist paths.
	for i := range sv.Patches {
		patch := sv.Patches[i]

		hasAllowedPrefix := false
		for a := range pathWhiteList {
			prefix := pathWhiteList[a]
			if strings.HasPrefix(patch.Path, prefix) {
				hasAllowedPrefix = true

				break
			}
		}
		// if the path is not in the allowed prefix list, return with an error.
		if !hasAllowedPrefix {
			return fmt.Errorf("path %s is not permitted for patches", patch.Path)
		}
	}

	patchBytes, err := json.Marshal(sv.Patches)
	if err != nil {
		return fmt.Errorf("marshalling patch: %w", err)
	}

	_, err = c.knativeCli.ServingV1().Services(c.config.KnativeNamespace).Patch(context.Background(), sv.GetID(), types.JSONPatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("applying patch: %w", err)
	}

	return nil
}

func (c *knativeClient) updateService(sv *core.ServiceFileData) error {
	// Updating knative service is basically done by removing the old one and re-creating it.
	err := c.deleteService(sv.GetID())
	if err != nil {
		return err
	}

	return c.createService(sv)
}

func (c *knativeClient) deleteService(id string) error {
	err := c.k8sCli.AppsV1().Deployments(c.config.KnativeNamespace).Delete(context.Background(), id, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = c.k8sCli.CoreV1().Services(c.config.KnativeNamespace).Delete(context.Background(), id, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *knativeClient) listServices() ([]status, error) {
	list, err := c.k8sCli.AppsV1().Deployments(c.config.KnativeNamespace).List(context.Background(), metav1.ListOptions{})
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
	l, err := c.k8sCli.CoreV1().Pods(c.config.KnativeNamespace).List(context.Background(), lo)
	if err != nil {
		return nil, err
	}

	type pod struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"createdAt"`
	}

	pods := []*pod{}
	for i := range l.Items {
		if l.Items[i].Labels["direktiv-service"] != id {
			continue
		}
		pods = append(pods, &pod{
			ID:        l.Items[i].Name,
			CreatedAt: l.Items[i].CreationTimestamp.Time,
		})
	}

	// Sort by CreatedAt (asc)
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].CreatedAt.Before(pods[j].CreatedAt)
	})

	return pods, nil
}

func (c *knativeClient) rebuildService(id string) error {
	return c.k8sCli.AppsV1().Deployments(c.config.KnativeNamespace).Delete(context.Background(), id,
		metav1.DeleteOptions{})
}

var _ runtimeClient = &knativeClient{}

type knativeStatus struct {
	*appsV1.Deployment
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

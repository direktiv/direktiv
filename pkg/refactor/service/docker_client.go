package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dClient "github.com/docker/docker/client"
)

type dockerClient struct {
	cli *dClient.Client
}

func (c *dockerClient) createService(cfg *core.ServiceConfig) error {
	config := &container.Config{
		Image: cfg.Image,
		Cmd:   []string{cfg.CMD},
		Labels: map[string]string{
			"direktiv.io/inputHash": cfg.GetValueHash(),
			"direktiv.io/id":        cfg.GetID(),
		},
	}
	hostConfig := &container.HostConfig{
		AutoRemove: false,
	}

	// Pull the image (if it doesn't exist locally).
	out, err := c.cli.ImagePull(context.Background(), config.Image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("image pull, err: %w", err)
	}
	defer out.Close()
	_, _ = io.Copy(io.Discard, out)

	// Create a container.
	resp, err := c.cli.ContainerCreate(context.Background(), config, hostConfig, nil, nil, cfg.GetID())
	if err != nil {
		return fmt.Errorf("create container, err: %w", err)
	}

	// Start the container.
	if err := c.cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("start container, err: %w", err)
	}

	return nil
}

func (c *dockerClient) updateService(cfg *core.ServiceConfig) error {
	// Remove the container.
	err := c.deleteService(cfg.GetID())
	if err != nil {
		return err
	}

	return c.createService(cfg)
}

func (c *dockerClient) getContainerBy(id string) (*types.Container, error) {
	containers, err := c.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}
	for i, cntr := range containers {
		if cntr.Labels["direktiv.io/id"] == id {
			return &containers[i], nil
		}
	}

	return nil, core.ErrNotFound
}

func (c *dockerClient) deleteService(id string) error {
	cntr, err := c.getContainerBy(id)
	if err != nil {
		return err
	}

	return c.cli.ContainerRemove(context.Background(), cntr.ID, types.ContainerRemoveOptions{
		Force: true, // Force removal even if the container is running.
	})
}

func (c *dockerClient) listServices() ([]status, error) {
	// List containers.
	containers, err := c.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	list := []status{}
	for i, cnt := range containers {
		if _, ok := cnt.Labels["direktiv.io/inputHash"]; ok {
			list = append(list, &dockerStatus{Container: &containers[i]})
		}
	}

	return list, nil
}

func (c *dockerClient) listServicePods(id string) (any, error) {
	_, err := c.getContainerBy(id)
	if err != nil {
		return nil, err
	}

	type pod struct {
		ID string `json:"id"`
	}

	return []*pod{
		{ID: id},
	}, nil
}

func (c *dockerClient) streamServiceLogs(id string, _ string) (io.ReadCloser, error) {
	cntr, err := c.getContainerBy(id)
	if err != nil {
		return nil, err
	}

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
	}

	// Get the log reader
	logs, err := c.cli.ContainerLogs(context.Background(), cntr.ID, options)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (c *dockerClient) killService(id string) error {
	cntr, err := c.getContainerBy(id)
	if err != nil {
		return err
	}

	options := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   true,
		Force:         true,
	}

	// delete the container
	return c.cli.ContainerRemove(context.Background(), cntr.ID, options)
}

func (c *dockerClient) getServiceURL(id string) string {
	return fmt.Sprintf("http://%s", id)
}

var _ client = &dockerClient{}

type dockerStatus struct {
	*types.Container
}

func (r *dockerStatus) GetConditions() any {
	type condition struct {
		Type    string `json:"type"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	if strings.Contains(r.Status, "Up ") {
		return []condition{
			{Type: "UpAndReady", Status: "True", Message: r.Status},
		}
	}
	if strings.Contains(r.Status, "Exited ") {
		return []condition{
			{Type: "UpAndReady", Status: "False", Message: r.Status},
		}
	}

	return []condition{
		{Type: "UpAndReady", Status: "Unknown", Message: r.Status},
	}
}

func (r *dockerStatus) GetID() string {
	return r.Labels["direktiv.io/id"]
}

func (r *dockerStatus) GetValueHash() string {
	return r.Labels["direktiv.io/inputHash"]
}

var _ status = &dockerStatus{}

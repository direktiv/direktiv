package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	dClient "github.com/docker/docker/client"
)

type dockerClient struct {
	cli *dClient.Client
}

func (c *dockerClient) cleanAll() error {
	containers, err := c.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return err
	}
	for _, cnt := range containers {
		if cnt.Labels["direktiv.io/object-type"] == "container" {
			err := c.cli.ContainerRemove(context.Background(), cnt.ID, types.ContainerRemoveOptions{
				RemoveVolumes: true,
				Force:         true,
			})
			if err != nil {
				return err
			}
		}
	}

	filterArgs := filters.NewArgs()
	filterArgs.Add("label", "direktiv.io/object-type=volume")
	vlms, err := c.cli.VolumeList(context.Background(), volume.ListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return err
	}
	for _, vlm := range vlms.Volumes {
		err := c.cli.VolumeRemove(context.Background(), vlm.Name, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *dockerClient) createService(sv *core.ServiceFileData) error {
	// don't pull any image that has 'local' prefix.
	if !strings.HasPrefix(sv.Image, "local") {
		out, err := c.cli.ImagePull(context.Background(), sv.Image, types.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("image pull, err: %w", err)
		}
		defer out.Close()
		_, _ = io.Copy(io.Discard, out)
	}

	_, err := c.cli.VolumeCreate(context.Background(), volume.CreateOptions{
		Name: sv.GetID(),
		Labels: map[string]string{
			"direktiv.io/object-type": "volume",
		},
	})
	if err != nil {
		return err
	}

	volumeConfig := &mount.Mount{
		Type:   mount.TypeVolume,
		Source: sv.GetID(),
		Target: "/mnt/shared",
	}

	containerConfig := &types.ContainerCreateConfig{
		Name: sv.GetID(),
		Config: &container.Config{
			Image: "direktiv-dev",
			Labels: map[string]string{
				"direktiv.io/object-type":    "container",
				"direktiv.io/container-type": "main",
				"direktiv.io/inputHash":      sv.GetValueHash(),
				"directiv.io/scale":          strconv.Itoa(sv.Scale),
			},
			Env: []string{
				"DIREKTIV_APP=sidecar",
				"DIREKITV_ENABLE_DOCKER=" + os.Getenv("DIREKITV_ENABLE_DOCKER"),
				util.DirektivFlowEndpoint + "=flow",
			},
		},
		HostConfig: &container.HostConfig{
			Mounts:      []mount.Mount{*volumeConfig},
			NetworkMode: "direktiv_default",
			AutoRemove:  false,
		},
		NetworkingConfig: &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"direktiv_default": {
					Aliases: []string{sv.GetID()},
				},
			},
		},
	}

	uContainerConfig := &types.ContainerCreateConfig{
		Name: sv.GetID() + "-user",
		Config: &container.Config{
			Image: sv.Image,
			Labels: map[string]string{
				"direktiv.io/object-type":    "container",
				"direktiv.io/container-type": "user",
			},
		},
		HostConfig: &container.HostConfig{
			Mounts:      []mount.Mount{*volumeConfig},
			NetworkMode: container.NetworkMode("container:" + sv.GetID()),
			AutoRemove:  false,
		},
	}
	if sv.Cmd != "" {
		uContainerConfig.Config.Cmd = []string{sv.Cmd}
	}

	// Create a containers.
	resp, err := c.cli.ContainerCreate(context.Background(),
		containerConfig.Config,
		containerConfig.HostConfig,
		containerConfig.NetworkingConfig, nil,
		containerConfig.Name)
	if err != nil {
		return fmt.Errorf("create main container, err: %w", err)
	}
	uResp, err := c.cli.ContainerCreate(context.Background(),
		uContainerConfig.Config,
		uContainerConfig.HostConfig,
		uContainerConfig.NetworkingConfig, nil,
		uContainerConfig.Name)
	if err != nil {
		return fmt.Errorf("create user container, err: %w", err)
	}

	// Start the container.
	if err := c.cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("start main container, err: %w", err)
	}
	if err := c.cli.ContainerStart(context.Background(), uResp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("start user container, err: %w", err)
	}

	return nil
}

func (c *dockerClient) updateService(sv *core.ServiceFileData) error {
	// Remove the container.
	err := c.deleteService(sv.GetID())
	if err != nil {
		return err
	}

	return c.createService(sv)
}

func (c *dockerClient) getContainerBy(id string) (*types.Container, error) {
	containers, err := c.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	for i, cntr := range containers {
		cntrID := strings.Trim(cntr.Names[0], "/")
		if cntrID == id {
			return &containers[i], nil
		}
	}

	return nil, core.ErrNotFound
}

func (c *dockerClient) deleteService(id string) error {
	err1 := c.deleteContainer(id)
	err2 := c.deleteContainer(id + "-user")

	if err1 != nil {
		return fmt.Errorf("delete container: %w", err1)
	}

	if err2 != nil {
		return fmt.Errorf("delete user container: %w", err2)
	}

	return nil
}

func (c *dockerClient) deleteContainer(id string) error {
	cntr, err := c.getContainerBy(id)
	if err != nil {
		return err
	}

	options := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	// delete the container
	return c.cli.ContainerRemove(context.Background(), cntr.ID, options)
}

func (c *dockerClient) listServices() ([]status, error) {
	// List containers.
	containers, err := c.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	list := []status{}
	for i, cnt := range containers {
		if cnt.Labels["direktiv.io/container-type"] == "main" {
			list = append(list, &dockerStatus{Container: &containers[i]})
		}
	}

	return list, nil
}

// listServicePods returns fake pos objects that refer to the same service container.
func (c *dockerClient) listServicePods(id string) (any, error) {
	ctr, err := c.getContainerBy(id)
	if err != nil {
		return nil, err
	}

	// the number of pods to return should match the same service scale number.
	scaleLabel := ctr.Labels["directiv.io/scale"]
	scale, _ := strconv.Atoi(scaleLabel)
	if scale == 0 {
		scale = 1
	}

	type pod struct {
		ID string `json:"id"`
	}

	result := []*pod{}
	for i := 1; i <= scale; i++ {
		result = append(result, &pod{
			ID: fmt.Sprintf("%s_%d", id, i),
		})
	}

	return result, nil
}

func (c *dockerClient) streamServiceLogs(id string, _ string) (io.ReadCloser, error) {
	cntr, err := c.getContainerBy(id + "-user")
	if err != nil {
		return nil, err
	}

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	}

	// Get the log reader
	logs, err := c.cli.ContainerLogs(context.Background(), cntr.ID, options)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (c *dockerClient) rebuildService(id string) error {
	return c.deleteService(id)
}

var _ runtimeClient = &dockerClient{}

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
	return strings.Trim(r.Names[0], "/")
}

func (r *dockerStatus) GetValueHash() string {
	return r.Labels["direktiv.io/inputHash"]
}

var _ status = &dockerStatus{}

package tsservice

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"
	"gopkg.in/yaml.v3"
)

const (
	composeStaticCmd = "generic"
)

func (fi *FlowInformation) CompileCompose(path, localPath, flowPath, initializer string) (*types.Project, error) {

	// TODO: image from config

	containerVolumes := []types.ServiceVolumeConfig{
		{
			Source: filepath.Join(localPath, "secrets"),
			Target: "/direktiv/secrets",
			Type:   types.VolumeTypeBind,
		},
		{
			Source: localPath,
			Target: "/direktiv/shared",
			Type:   types.VolumeTypeBind,
		},
		{
			Source: filepath.Join(localPath, "instances"),
			Target: "/direktiv/instances",
			Type:   types.VolumeTypeBind,
		},
		{
			Source: "bin",
			Target: "/direktiv/bin",
			Type:   types.VolumeTypeVolume,
		},
	}

	engineEnvs := []string{
		"DIREKTIV_JSENGINE_BASEDIR=/direktiv",
		fmt.Sprintf("DIREKTIV_JSENGINE_FLOWPATH=%s", flowPath),
		fmt.Sprintf("DIREKTIV_JSENGINE_INIT=%s", initializer),
		"DIREKTIV_JSENGINE_SELFCOPY=/direktiv/bin/engine",
	}

	img := "localhost:5000/engine"

	services := types.Services{}

	startPort := 8081

	for k, f := range fi.Functions {

		// only do local functions
		if f.Image != "" {

			envs := []string{
				fmt.Sprintf("DIREKTIV_PORT=%d", startPort),
			}

			for k, v := range f.Envs {
				envs = append(envs, fmt.Sprintf("%s=%s", k, v))
			}

			dependsOn := make(types.DependsOnConfig)
			dependsOn["engine"] = types.ServiceDependency{
				Condition: types.ServiceConditionHealthy,
			}
			sc := types.ServiceConfig{
				Name:        k,
				Image:       f.Image,
				Environment: types.NewMappingWithEquals(envs),
				DependsOn:   dependsOn,
				Volumes:     containerVolumes,
				Restart:     types.RestartPolicyAlways,
			}

			// if set to special command we change the command
			if f.Cmd == composeStaticCmd {
				cmd := make(types.ShellCommand, 2)
				cmd[0] = "/direktiv/bin/engine"
				cmd[1] = "cmd"

				sc.Command = cmd
			}

			services[f.GetID()] = sc
			engineEnvs = append(engineEnvs, fmt.Sprintf("%s=http://%s:%d", k, f.GetID(), startPort))

			startPort = startPort + 1
		}
	}

	var retries uint64 = 10
	var interval types.Duration
	err := interval.DecodeMapstructure("1s")
	if err != nil {

	}
	services["engine"] = types.ServiceConfig{
		Name:        "engine",
		Image:       img,
		Entrypoint:  []string{"/engine", "server"},
		Environment: types.NewMappingWithEquals(engineEnvs),
		Ports: []types.ServicePortConfig{
			{
				Published: "8080",
				Target:    8080,
				Protocol:  "tcp",
			},
		},
		Restart: types.RestartPolicyAlways,
		Volumes: containerVolumes,
		HealthCheck: &types.HealthCheckConfig{
			Test: types.HealthCheckTest{
				"CMD-SHELL",
				"curl -sS --fail http://127.0.0.1:8080/status | jq -e '. | select(.failed == false and .initialized == true) | length > 0' || exit 1",
			},
			Retries:  &retries,
			Interval: &interval,
		},
	}

	volumes := make(types.Volumes)
	volumes["bin"] = types.VolumeConfig{}

	project := types.Project{
		Services: services,
		Volumes:  volumes,
	}

	if path != "" {
		b, err := yaml.Marshal(project)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(path, b, 0755)
		if err != nil {
			return nil, err
		}
	}

	return &project, nil
}

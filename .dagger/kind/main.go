// A generated module for Kind functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/kind/internal/dagger"
	"fmt"

	_ "embed"
)

//go:embed kind-config.yaml
var kindConfig []byte

type Kind struct {
	// +private
	DockerSocket *dagger.Socket
}

const defaultImage = "alpine/k8s:1.32.5"

// Container that contains the kind and k9s binaries
func (k *Kind) Container(ctx context.Context, socket *dagger.Socket) *dagger.Container {

	k.DockerSocket = socket
	id, err := k.DockerSocket.ID(ctx)
	fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! %v %v\n", id, err)

	return dag.Container().
		From(defaultImage).
		WithoutEntrypoint().
		WithUser("root").
		WithWorkdir("/").
		WithExec([]string{"apk", "add", "--no-cache", "docker", "kind", "k9s"}).
		WithUnixSocket("/var/run/docker.sock", k.DockerSocket).
		WithEnvVariable("DOCKER_HOST", "unix:///var/run/docker.sock")
}

// Returns lines that match a pattern in the files of the provided Directory
// func (m *Kind) Cluster(ctx context.Context, directoryArg *dagger.Directory, pattern string) (string, error) {
// 	return dag.Container().
// 		From("alpine:latest").
// 		WithMountedDirectory("/mnt", directoryArg).
// 		WithWorkdir("/mnt").
// 		WithExec([]string{"grep", "-R", pattern, "."}).
// 		Stdout(ctx)
// }

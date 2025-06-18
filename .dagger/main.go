// A generated module for Direktiv functions
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
	"dagger/direktiv/internal/dagger"
	"fmt"
)

// var dag = dagger.Connect()

type Direktiv struct {
	// +private
	Source *dagger.Directory
}

func New(
	// +optional
	// +defaultPath="/"
	source *dagger.Directory) *Direktiv {
	g := &Direktiv{
		Source: source,
	}
	return g
}

// Return a base trivy container
func (m *Direktiv) Base(ctx context.Context) *dagger.Container {
	return dag.Container().
		From("aquasec/trivy").
		WithMountedDirectory("/mnt", m.Source).
		WithWorkdir("/mnt").
		WithMountedCache("/tmp/.cache/trivy", dag.CacheVolume("trivy-db-cache"))
}

// Return a container with a specified directory
func (m *Direktiv) ScanHelm(
	ctx context.Context,
) (string, error) {

	s, err := m.Base(ctx).
		WithExec([]string{"trivy", "config", "--exit-code", "1", "charts"}).
		Stdout(ctx)
	fmt.Println(s)
	fmt.Println(err)

	return s, err
}

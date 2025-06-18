package main

import (
	"context"
	"fmt"

	kindCmd "sigs.k8s.io/kind/pkg/cmd"
)

func (m *Direktiv) Kind(ctx context.Context) {

	kindLogger := kindCmd.NewLogger()
	fmt.Println(kindLogger)
	// return dag.Container().
	// 	From("aquasec/trivy").
	// 	WithMountedDirectory("/mnt", m.Source).
	// 	WithWorkdir("/mnt").
	// 	WithMountedCache("/tmp/.cache/trivy", dag.CacheVolume("trivy-db-cache"))
}

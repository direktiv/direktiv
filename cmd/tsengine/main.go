package tsengine

import (
	"github.com/direktiv/direktiv/pkg/tsengine"
)

// "github.com/direktiv/direktiv/pkg/core"
// "github.com/direktiv/direktiv/pkg/engine"

func RunApplication() {

	loggingCtx := tracing.WithTrack(ns.WithTags(ctx), tracing.BuildNamespaceTrack(ns.Name))

	// namespaceTrackCtx := enginerefactor.WithTrack(loggingCtx, engine.BuildNamespaceTrack(im.instance.Instance.Namespace))
	// slog.Info("Workflow has been triggered", enginerefactor.GetSlogAttributesWithStatus(namespaceTrackCtx, core.LogRunningStatus)...)

	// dir, err := os.MkdirTemp(".", "tsengine")
	// if err != nil {
	// 	panic(err)
	// }

	tsengine.NewServer()
}

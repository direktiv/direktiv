package engine

import (
	"context"

	"github.com/direktiv/direktiv/pkg/core"
)

func (instance *Instance) WithTags(ctx context.Context) context.Context {
	tags, ok := ctx.Value(core.LogTagsKey).(map[string]interface{})
	if !ok {
		tags = make(map[string]interface{}, 0)
	}

	tags["instance"] = instance.Instance.ID
	tags["invoker"] = instance.Instance.Invoker
	callpath := ""
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}
	tags["callpath"] = callpath
	tags["workflow"] = instance.Instance.WorkflowPath

	return context.WithValue(ctx, core.LogTagsKey, tags)
}

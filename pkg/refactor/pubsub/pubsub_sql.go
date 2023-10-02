package pubsub

type Bus interface {
	Publish(channel string, data string) error
	Subscribe(channel string, handler func(data string))
}

var (
	WorkflowCreate = "workflow_create"
	WorkflowDelete = "workflow_delete"
	WorkflowUpdate = "workflow_update"

	FunctionCreate = "function_create"
	FunctionDelete = "function_delete"
	FunctionUpdate = "function_update"

	MirrorSync = "mirror_sync"
)

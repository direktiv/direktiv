package pubsub

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CoreBus interface {
	Publish(channel string, data string) error
	Loop(done <-chan struct{}, logger *zap.SugaredLogger, handler func(channel string, data string))
}

var (
	WorkflowCreate = "workflow_create"
	WorkflowDelete = "workflow_delete"
	WorkflowUpdate = "workflow_update"
	WorkflowRename = "workflow_rename"

	ServiceCreate = "service_create"
	ServiceDelete = "service_delete"
	ServiceUpdate = "service_update"
	ServiceRename = "service_rename"

	EndpointCreate = "endpoint_create"
	EndpointDelete = "endpoint_delete"
	EndpointUpdate = "endpoint_update"
	EndpointRename = "endpoint_rename"

	ConsumerCreate = "consumer_create"
	ConsumerDelete = "consumer_delete"
	ConsumerUpdate = "consumer_update"
	ConsumerRename = "consumer_rename"

	MirrorSync      = "mirror_sync"
	NamespaceDelete = "namespace_delete"
	NamespaceCreate = "namespace_create"

	SecretCreate = "secret_create"
	SecretDelete = "secret_delete"
	SecretUpdate = "secret_update"
)

type FileChangeEvent struct {
	Namespace    string
	NamespaceID  uuid.UUID
	FilePath     string
	OldPath      string
	DeleteFileID uuid.UUID
}

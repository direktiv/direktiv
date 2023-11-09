package pubsub

type Bus interface {
	Publish(channel string, data string)
	Subscribe(channel string, handler func(data string)) string
	Unsubscribe(key string)
}

var (
	WorkflowCreate = "workflow_create"
	WorkflowDelete = "workflow_delete"
	WorkflowUpdate = "workflow_update"

	ServiceCreate = "service_create"
	ServiceDelete = "service_delete"
	ServiceUpdate = "service_update"

	EndpointCreate = "endpoint_create"
	EndpointDelete = "endpoint_delete"
	EndpointUpdate = "endpoint_update"

	MirrorSync   = "mirror_sync"
	MirrorCancel = "mirror_cancel"

	NamespaceDelete                 = "namespace_delete"
	ConfigureRouterCron             = "configure_router_cron"
	TimerDelete                     = "timer_delete"
	TimersInstanceDelete            = "timers_instance_delete"
	EventFilterCacheDelete          = "event_filter_cache_delete"
	EventFilterCacheDeleteNamespace = "event_filter_cache_delete_namespace"
	EventListenersNotify            = "event_listeners_notify"
	EventReceived                   = "event_received"
	EngineCancelInstance            = "engine_cancel_instance"
	InstanceUpdate                  = "instance_update" // TODO: optimize
	LogsNotify                      = "logs_notify"
)

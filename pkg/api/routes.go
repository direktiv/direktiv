package api

// RN = Route Name
const (
	RN_Preflight = "preflight"
	// RN_HealthCheck                  = "healthCheck"
	RN_ListNamespaces       = "listNamespaces"
	RN_AddNamespace         = "addNamespace"
	RN_DeleteNamespace      = "deleteNamespace"
	RN_GetNode              = "getNode"
	RN_CreateDirectory      = "createDirectory"
	RN_CreateWorkflow       = "createWorkflow"
	RN_UpdateWorkflow       = "updateWorkflow"
	RN_SaveWorkflow         = "saveWorkflow"
	RN_DiscardWorkflow      = "discardWorkflow"
	RN_DeleteNode           = "deleteNode"
	RN_GetWorkflowTags      = "getWorkflowTags"
	RN_GetWorkflowRevisions = "getWorkflowRevisions"
	RN_GetWorkflowRefs      = "getWorkflowRefs"
	RN_DeleteRevision       = "deleteRevision"
	RN_Tag                  = "tag"
	RN_Untag                = "untag"
	RN_Retag                = "retag"
	// RN_NamespaceEvent               = "namespaceEvent"
	// RN_ListSecrets                  = "listSecrets"
	// RN_CreateSecret                 = "createSecret"
	// RN_DeleteSecret                 = "deleteSecret"
	// RN_ListRegistries               = "listRegistries"
	// RN_CreateRegistry               = "createRegistry"
	// RN_DeleteRegistry               = "deleteRegistry"
	// RN_GetWorkflowMetrics           = "getWorkflowMetrics"
	// RN_ListWorkflows                = "listWorkflows"
	// RN_GetWorkflow                  = "getWorkflow"
	// RN_UpdateWorkflow               = "updateWorkflow"
	// RN_ToggleWorkflow               = "toggleWorkflow"
	// RN_CreateWorkflow               = "createWorkflow"
	// RN_DeleteWorkflow               = "deleteWorkflow"
	// RN_DownloadWorkflow             = "downloadWorkflow"
	// RN_ExecuteWorkflow              = "executeWorkflow"
	// RN_ListWorkflowInstances        = "listWorkflowInstances"
	// RN_ListInstances                = "listInstances"
	// RN_GetInstance                  = "getInstance"
	// RN_CancelInstance               = "cancelInstance"
	// RN_ListActionTemplateFolders    = "listActionTemplateFolders"
	// RN_ListActionTemplates          = "listActionTemplates"
	// RN_GetActionTemplate            = "getActionTemplate"
	// RN_ListWorkflowTemplateFolders  = "listWorkflowTemplateFolders"
	// RN_ListWorkflowTemplates        = "listWorkflowTemplates"
	// RN_GetWorkflowTemplate          = "getWorkflowTemplate"
	// RN_ListWorkflowVariables        = "listWorkflowVariables"
	// RN_GetWorkflowVariable          = "getWorkflowVariable"
	// RN_SetWorkflowVariable          = "setWorkflowVariable"
	// RN_ListNamespaceVariables       = "listNamespaceVariables"
	// RN_GetNamespaceVariable         = "getNamespaceVariable"
	RN_GetServerLogs    = "getServerLogs"
	RN_GetNamespaceLogs = "getNamespaceLogs"
	RN_GetWorkflowLogs  = "getWorkflowLogs"
	RN_GetInstanceLogs  = "getInstanceLogs"
	// RN_SetNamespaceVariable         = "setNamespaceVariable"
	// RN_JQPlayground                 = "jqPlayground"
	RN_ListServices = "listServices"
	// RN_WatchServices                = "watchServices"
	// RN_WatchInstanceServices        = "watchInstanceServices"
	// RN_WatchNamespaceServices       = "watchNamespaceServices"
	// RN_WatchRevisions               = "watchRevisions"
	// RN_WatchNamespaceRevisions      = "watchNamespaceRevisions"
	// RN_WatchPods                    = "watchPods"
	// RN_WatchLogs                    = "watchLogs"
	// RN_ListPods                     = "listPods"
	RN_DeleteServices = "deleteServices"
	// RN_GetService                   = "getService"
	RN_CreateService = "createService"
	// RN_UpdateService                = "updateService"
	// RN_UpdateServiceTraffic         = "updateServiceTraffic"
	// RN_DeleteService                = "deleteService"
	// RN_DeleteRevision               = "deleteRevision"
	// RN_GetWorkflowFunctions         = "getWorkflowFunctions"
	// RN_namespaceWorkflowsInvoked    = "namespaceWorkflowsInvoked"
	// RN_namespaceWorkflowsSuccessful = "namespaceWorkflowsSuccessful"
	// RN_namespaceWorkflowsFailed     = "namespaceWorkflowsFailed"
	// RN_namespaceWorkflowsMS         = "namespaceWorkflowsMS"
	// RN_metricsWorkflowInvoked       = "metricsWorkflowInvoked"
	// RN_metricsWorkflowSuccessful    = "metricsWorkflowSuccessful"
	// RN_metricsWorkflowFailed        = "metricsWorkflowFailed"
	// RN_metricsWorkflowMS            = "metricsWorkflowMS"
	// RN_metricsStateMS               = "metricsStateMS"
)

var RouteNames = []string{
	RN_Preflight,
	// RN_ListNamespaces,
	// RN_AddNamespace,
	// RN_DeleteNamespace,
	// RN_NamespaceEvent,
	// RN_GetNamespaceLogs,
	// RN_ListSecrets,
	// RN_CreateSecret,
	// RN_DeleteSecret,
	// RN_ListRegistries,
	// RN_CreateRegistry,
	// RN_DeleteRegistry,
	// RN_GetWorkflowMetrics,
	// RN_ListWorkflows,
	// RN_GetWorkflow,
	// RN_WatchLogs,
	// RN_UpdateWorkflow,
	// RN_ToggleWorkflow,
	// RN_CreateWorkflow,
	// RN_DeleteWorkflow,
	// RN_DownloadWorkflow,
	// RN_ExecuteWorkflow,
	// RN_ListWorkflowInstances,
	// RN_ListInstances,
	// RN_GetInstance,
	// RN_CancelInstance,
	// RN_GetInstanceLogs,
	// RN_ListActionTemplateFolders,
	// RN_ListActionTemplates,
	// RN_GetActionTemplate,
	// RN_ListWorkflowTemplateFolders,
	// RN_ListWorkflowTemplates,
	// RN_GetWorkflowTemplate,
	// RN_ListWorkflowVariables,
	// RN_GetWorkflowVariable,
	// RN_SetWorkflowVariable,
	// RN_ListNamespaceVariables,
	// RN_GetNamespaceVariable,
	// RN_SetNamespaceVariable,
	// RN_JQPlayground,
	RN_ListServices,
	// RN_WatchServices,
	RN_DeleteServices,
	// RN_GetService,
	RN_CreateService,
	// RN_UpdateService,
	// RN_UpdateServiceTraffic,
	// RN_DeleteService,
	// RN_DeleteRevision,
	// RN_GetWorkflowFunctions,
	// RN_WatchPods,
	// RN_ListPods,
	// RN_WatchRevisions,
}

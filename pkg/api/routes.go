package api

// RN = Route Name
const (
	RN_Preflight                   = "preflight"
	RN_HealthCheck                 = "healthCheck"
	RN_ListNamespaces              = "listNamespaces"
	RN_AddNamespace                = "addNamespace"
	RN_DeleteNamespace             = "deleteNamespace"
	RN_NamespaceEvent              = "namespaceEvent"
	RN_ListSecrets                 = "listSecrets"
	RN_CreateSecret                = "createSecret"
	RN_DeleteSecret                = "deleteSecret"
	RN_ListRegistries              = "listRegistries"
	RN_CreateRegistry              = "createRegistry"
	RN_DeleteRegistry              = "deleteRegistry"
	RN_GetWorkflowMetrics          = "getWorkflowMetrics"
	RN_ListWorkflows               = "listWorkflows"
	RN_GetWorkflow                 = "getWorkflow"
	RN_UpdateWorkflow              = "updateWorkflow"
	RN_ToggleWorkflow              = "toggleWorkflow"
	RN_CreateWorkflow              = "createWorkflow"
	RN_DeleteWorkflow              = "deleteWorkflow"
	RN_DownloadWorkflow            = "downloadWorkflow"
	RN_ExecuteWorkflow             = "executeWorkflow"
	RN_ListWorkflowInstances       = "listWorkflowInstances"
	RN_ListInstances               = "listInstances"
	RN_GetInstance                 = "getInstance"
	RN_CancelInstance              = "cancelInstance"
	RN_GetInstanceLogs             = "getInstanceLogs"
	RN_ListActionTemplateFolders   = "listActionTemplateFolders"
	RN_ListActionTemplates         = "listActionTemplates"
	RN_GetActionTemplate           = "getActionTemplate"
	RN_ListWorkflowTemplateFolders = "listWorkflowTemplateFolders"
	RN_ListWorkflowTemplates       = "listWorkflowTemplates"
	RN_GetWorkflowTemplate         = "getWorkflowTemplate"
	RN_ListWorkflowVariables       = "listWorkflowVariables"
	RN_GetWorkflowVariable         = "getWorkflowVariable"
	RN_SetWorkflowVariable         = "setWorkflowVariable"
	RN_ListNamespaceVariables      = "listNamespaceVariables"
	RN_GetNamespaceVariable        = "getNamespaceVariable"
	RN_GetNamespaceLogs            = "getNamespaceLogs"
	RN_SetNamespaceVariable        = "setNamespaceVariable"
	RN_JQPlayground                = "jqPlayground"
	RN_ListServices                = "listServices"
	RN_DeleteServices              = "deleteServices"
	RN_GetService                  = "getService"
	RN_CreateService               = "createService"
	RN_UpdateService               = "updateService"
	RN_UpdateServiceTraffic        = "updateServiceTraffic"
	RN_DeleteService               = "deleteService"
	RN_DeleteRevision              = "deleteRevision"
	RN_GetWorkflowFunctions        = "getWorkflowFunctions"
)

var RouteNames = []string{
	RN_Preflight,
	RN_ListNamespaces,
	RN_AddNamespace,
	RN_DeleteNamespace,
	RN_NamespaceEvent,
	RN_GetNamespaceLogs,
	RN_ListSecrets,
	RN_CreateSecret,
	RN_DeleteSecret,
	RN_ListRegistries,
	RN_CreateRegistry,
	RN_DeleteRegistry,
	RN_GetWorkflowMetrics,
	RN_ListWorkflows,
	RN_GetWorkflow,
	RN_UpdateWorkflow,
	RN_ToggleWorkflow,
	RN_CreateWorkflow,
	RN_DeleteWorkflow,
	RN_DownloadWorkflow,
	RN_ExecuteWorkflow,
	RN_ListWorkflowInstances,
	RN_ListInstances,
	RN_GetInstance,
	RN_CancelInstance,
	RN_GetInstanceLogs,
	RN_ListActionTemplateFolders,
	RN_ListActionTemplates,
	RN_GetActionTemplate,
	RN_ListWorkflowTemplateFolders,
	RN_ListWorkflowTemplates,
	RN_GetWorkflowTemplate,
	RN_ListWorkflowVariables,
	RN_GetWorkflowVariable,
	RN_SetWorkflowVariable,
	RN_ListNamespaceVariables,
	RN_GetNamespaceVariable,
	RN_SetNamespaceVariable,
	RN_JQPlayground,
	RN_ListServices,
	RN_DeleteServices,
	RN_GetService,
	RN_CreateService,
	RN_UpdateService,
	RN_UpdateServiceTraffic,
	RN_DeleteService,
	RN_DeleteRevision,
	RN_GetWorkflowFunctions,
}

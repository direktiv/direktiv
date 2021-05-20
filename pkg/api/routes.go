package api

// RN = Route Name
const (
	RN_Preflight                   = "Preflight"
	RN_ListNamespaces              = "ListNamespaces"
	RN_AddNamespace                = "AddNamespace"
	RN_DeleteNamespace             = "DeleteNamespace"
	RN_NamespaceEvent              = "NamespaceEvent"
	RN_ListSecrets                 = "ListSecrets"
	RN_CreateSecret                = "CreateSecret"
	RN_DeleteSecret                = "DeleteSecret"
	RN_ListRegistries              = "ListRegistries"
	RN_CreateRegistry              = "CreateRegistry"
	RN_DeleteRegistry              = "DeleteRegistry"
	RN_GetWorkflowMetrics          = "GetWorkflowMetrics"
	RN_ListWorkflows               = "ListWorkflows"
	RN_GetWorkflow                 = "GetWorkflow"
	RN_UpdateWorkflow              = "UpdateWorkflow"
	RN_ToggleWorkflow              = "ToggleWorkflow"
	RN_CreateWorkflow              = "CreateWorkflow"
	RN_DeleteWorkflow              = "DeleteWorkflow"
	RN_DownloadWorkflow            = "DownloadWorkflow"
	RN_ExecuteWorkflow             = "ExecuteWorkflow"
	RN_ListWorkflowInstances       = "ListWorkflowInstances"
	RN_ListInstances               = "ListInstances"
	RN_GetInstance                 = "GetInstance"
	RN_CancelInstance              = "CancelInstance"
	RN_GetInstanceLogs             = "GetInstanceLogs"
	RN_ListActionTemplateFolders   = "ListActionTemplateFolders"
	RN_ListActionTemplates         = "ListActionTemplates"
	RN_GetActionTemplate           = "GetActionTemplate"
	RN_ListWorkflowTemplateFolders = "ListWorkflowTemplateFolders"
	RN_ListWorkflowTemplates       = "ListWorkflowTemplates"
	RN_GetWorkflowTemplate         = "GetWorkflowTemplate"
	RN_ListWorkflowVariables       = "ListWorkflowVariables"
	RN_GetWorkflowVariable         = "GetWorkflowVariable"
	RN_SetWorkflowVariable         = "SetWorkflowVariable"
	RN_ListNamespaceVariables      = "ListNamespaceVariables"
	RN_GetNamespaceVariable        = "GetNamespaceVariable"
	RN_SetNamespaceVariable        = "SetNamespaceVariable"
	RN_JQPlayground                = "JQPlayground"
)

var RouteNames = []string{
	RN_Preflight,
	RN_ListNamespaces,
	RN_AddNamespace,
	RN_DeleteNamespace,
	RN_NamespaceEvent,
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
}

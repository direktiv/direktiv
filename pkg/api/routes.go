package api

// RN = Route Name
const (
	RN_Preflight                   = "preflight"
	RN_ListNamespaces              = "listNamespaces"
	RN_AddNamespace                = "addNamespace"
	RN_GetNamespaceConfig          = "getNamespaceConfiguration"
	RN_SetNamespaceConfig          = "setNamespaceConfiguration"
	RN_DeleteNamespace             = "deleteNamespace"
	RN_GetNode                     = "getNode"
	RN_CreateDirectory             = "createDirectory"
	RN_CreateWorkflow              = "createWorkflow"
	RN_UpdateWorkflow              = "updateWorkflow"
	RN_SaveWorkflow                = "saveWorkflow"
	RN_DiscardWorkflow             = "discardWorkflow"
	RN_DeleteNode                  = "deleteNode"
	RN_RenameNode                  = "renameNode"
	RN_GetWorkflowTags             = "getWorkflowTags"
	RN_GetWorkflowRevisions        = "getWorkflowRevisions"
	RN_GetWorkflowRefs             = "getWorkflowRefs"
	RN_DeleteRevision              = "deleteRevision"
	RN_Tag                         = "tag"
	RN_Untag                       = "untag"
	RN_Retag                       = "retag"
	RN_GetWorkflowRouter           = "getWorkflowRouter"
	RN_EventListeners              = "eventListeners"
	RN_EventHistory                = "eventHistory"
	RN_EditWorkflowRouter          = "editWorkflowRouter"
	RN_ValidateRef                 = "validateRef"
	RN_ValidateRouter              = "validateRouter"
	RN_NamespaceEvent              = "namespaceEvent"
	RN_ListSecrets                 = "listSecrets"
	RN_SearchSecret                = "searchSecret"
	RN_OverwriteSecret             = "overwriteSecret"
	RN_CreateSecret                = "createSecret"
	RN_DeleteSecret                = "deleteSecret"
	RN_DeleteSecretsFolder         = "deleteSecretsFolder"
	RN_CreateSecretsFolder         = "createSecretsFolder"
	RN_ListRegistries              = "listRegistries"
	RN_CreateRegistry              = "createRegistry"
	RN_DeleteRegistry              = "deleteRegistry"
	RN_ListGlobalRegistries        = "listGlobalRegistries"
	RN_CreateGlobalRegistry        = "createGlobalRegistry"
	RN_DeleteGlobalRegistry        = "deleteGlobalRegistry"
	RN_ListGlobalPrivateRegistries = "listGlobalPrivateRegistries"
	RN_CreateGlobalPrivateRegistry = "createGlobalPrivateRegistry"
	RN_DeleteGlobalPrivateRegistry = "deleteGlobalPrivateRegistry"
	RN_GetNamespaceMetrics         = "getNamespaceMetrics"
	RN_GetWorkflowMetrics          = "getWorkflowMetrics"
	RN_TestRegistry                = "testRegistry"
	// RN_ListWorkflows                = "listWorkflows"
	// RN_GetWorkflow                  = "getWorkflow"
	// RN_ToggleWorkflow               = "toggleWorkflow"
	// RN_DeleteWorkflow               = "deleteWorkflow"
	// RN_DownloadWorkflow             = "downloadWorkflow"
	RN_ExecuteWorkflow = "executeWorkflow"
	// RN_ListWorkflowInstances        = "listWorkflowInstances"
	RN_ListInstances        = "listInstances"
	RN_GetInstance          = "getInstance"
	RN_CancelInstance       = "cancelInstance"
	RN_DeleteNodeAttributes = "deleteNodeAttributes"
	RN_CreateNodeAttributes = "createNodeAttributes"
	// RN_ListActionTemplateFolders    = "listActionTemplateFolders"
	// RN_ListActionTemplates          = "listActionTemplates"
	// RN_GetActionTemplate            = "getActionTemplate"
	// RN_ListWorkflowTemplateFolders  = "listWorkflowTemplateFolders"
	// RN_ListWorkflowTemplates        = "listWorkflowTemplates"
	// RN_GetWorkflowTemplate          = "getWorkflowTemplate"
	RN_ListInstanceVariables  = "listInstanceVariables"
	RN_GetInstanceVariable    = "getInstanceVariable"
	RN_SetInstanceVariable    = "setInstanceVariable"
	RN_ListWorkflowVariables  = "listWorkflowVariables"
	RN_GetWorkflowVariable    = "getWorkflowVariable"
	RN_SetWorkflowVariable    = "setWorkflowVariable"
	RN_ListNamespaceVariables = "listNamespaceVariables"
	RN_GetNamespaceVariable   = "getNamespaceVariable"
	RN_GetServerLogs          = "getServerLogs"
	RN_GetNamespaceLogs       = "getNamespaceLogs"
	RN_GetWorkflowLogs        = "getWorkflowLogs"
	RN_GetInstanceLogs        = "getInstanceLogs"
	RN_SetNamespaceVariable   = "setNamespaceVariable"
	RN_JQPlayground           = "jqPlayground"
	RN_Version                = "version"
	RN_ListServices           = "listServices"
	RN_ListNamespaceServices  = "listNamespacesServices"
	RN_WatchServices          = "watchServices"
	// RN_WatchInstanceServices        = "watchInstanceServices"
	// RN_WatchNamespaceServices       = "watchNamespaceServices"
	RN_WatchRevisions = "watchRevisions"
	// RN_WatchNamespaceRevisions      = "watchNamespaceRevisions"
	RN_WatchPods            = "watchPods"
	RN_WatchLogs            = "watchLogs"
	RN_ListPods             = "listPods"
	RN_DeleteServices       = "deleteServices"
	RN_GetService           = "getService"
	RN_CreateService        = "createService"
	RN_UpdateService        = "updateService"
	RN_UpdateServiceTraffic = "updateServiceTraffic"
	RN_DeleteService        = "deleteService"

	RN_GlobalDependencies     = "globalDependencies"
	RN_NamespacedDependencies = "namespacedDependencies"

	RN_ListNamespacePods             = "listNamespacePods"
	RN_CreateNamespaceService        = "createNamespaceService"
	RN_DeleteNamespaceServices       = "deleteNamespaceService"
	RN_GetNamespaceService           = "getNamespaceService"
	RN_UpdateNamespaceService        = "updateNamespaceService"
	RN_UpdateNamespaceServiceTraffic = "updateNamespaceServiceTraffic"
	RN_DeleteNamespaceRevision       = "deleteNamespaceRevision"

	RN_ListWorkflowServices   = "listWorkflowServices"
	RN_DeleteWorkflowServices = "deleteWorkflowService"
	RN_ListWorkflowPods       = "listWorkflowPods"

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

	RN_UpdateMirror          = "updateMirror"
	RN_LockMirror            = "lockMirror"
	RN_SyncMirror            = "syncMirror"
	RN_GetMirrorInfo         = "getMirrorInfo"
	RN_CancelMirrorActivity  = "cancelMirrorActivity"
	RN_GetMirrorActivityLogs = "getMirrorActivityLogs"
)

// var RouteNames = []string{
// 	RN_Preflight,
// 	// RN_ListNamespaces,
// 	// RN_AddNamespace,
// 	// RN_DeleteNamespace,
// 	// RN_NamespaceEvent,
// 	// RN_GetNamespaceLogs,
// 	// RN_ListSecrets,
// 	// RN_CreateSecret,
// 	// RN_DeleteSecret,
// 	RN_ListRegistries,
// 	RN_CreateRegistry,
// 	RN_DeleteRegistry,
// 	// RN_GetWorkflowMetrics,
// 	// RN_ListWorkflows,
// 	// RN_GetWorkflow,
// 	// RN_WatchLogs,
// 	// RN_UpdateWorkflow,
// 	// RN_ToggleWorkflow,
// 	// RN_CreateWorkflow,
// 	// RN_DeleteWorkflow,
// 	// RN_DownloadWorkflow,
// 	// RN_ExecuteWorkflow,
// 	// RN_ListWorkflowInstances,
// 	// RN_ListInstances,
// 	// RN_GetInstance,
// 	// RN_CancelInstance,
// 	// RN_GetInstanceLogs,
// 	// RN_ListActionTemplateFolders,
// 	// RN_ListActionTemplates,
// 	// RN_GetActionTemplate,
// 	// RN_ListWorkflowTemplateFolders,
// 	// RN_ListWorkflowTemplates,
// 	// RN_GetWorkflowTemplate,
// 	// RN_ListWorkflowVariables,
// 	// RN_GetWorkflowVariable,
// 	// RN_SetWorkflowVariable,
// 	// RN_ListNamespaceVariables,
// 	// RN_GetNamespaceVariable,
// 	// RN_SetNamespaceVariable,
// 	// RN_JQPlayground,
// 	RN_ListServices,
// 	RN_ListNamespaceServices,
// 	RN_WatchServices,
// 	RN_DeleteServices,
// 	RN_GetService,
// 	RN_CreateService,
// 	RN_UpdateService,
// 	RN_UpdateServiceTraffic,
// 	RN_DeleteService,
// 	RN_DeleteRevision,
// 	// RN_GetWorkflowFunctions,
// 	// RN_WatchPods,
// 	RN_ListPods,
// 	RN_ListNamespacePods,
// 	RN_ListNamespacePods,
// 	RN_CreateNamespaceService,
// 	RN_DeleteNamespaceServices,
// 	RN_GetNamespaceService,
// 	RN_UpdateNamespaceService,
// 	RN_UpdateNamespaceServiceTraffic,
// 	RN_DeleteNamespaceRevision,
// 	RN_WatchRevisions,
// }

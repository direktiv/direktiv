package api

// RN = Route Name.
const (

	// admin only routes.
	RN_AddNamespace = "addNamespace"

	// misc route for azure.
	RN_Preflight = "preflight"

	// authenticated only routes.
	RN_JQPlayground = "jqPlayground"
	RN_Version      = "version"

	// permission if any other permission is set in namespace.
	RN_ListNamespaces = "listNamespaces"

	// namespace get routes.
	RN_ListNamespaceVariables = "listNamespaceVariables"
	RN_GetNamespaceVariable   = "getNamespaceVariable"

	// namespace variable set routes.
	RN_SetNamespaceVariable = "setNamespaceVariable"

	// explorer routes.
	RN_GetNode              = "getNode"
	RN_CreateDirectory      = "createDirectory"
	RN_DeleteNode           = "deleteNode"
	RN_RenameNode           = "renameNode"
	RN_DeleteNodeAttributes = "deleteNodeAttributes"
	RN_CreateNodeAttributes = "createNodeAttributes"

	// instance routes.
	RN_ListInstances         = "listInstances"
	RN_GetInstance           = "getInstance"
	RN_GetInstanceVariable   = "getInstanceVariable"
	RN_CancelInstance        = "cancelInstance"
	RN_ListInstanceVariables = "listInstanceVariables"
	RN_SetInstanceVariable   = "setInstanceVariable"
	RN_GetNamespaceMetrics   = "getNamespaceMetrics"

	// workflow routes.
	RN_GetWorkflowTags       = "getWorkflowTags"
	RN_GetWorkflowRefs       = "getWorkflowRefs"
	RN_GetWorkflowRouter     = "getWorkflowRouter"
	RN_GetWorkflowMetrics    = "getWorkflowMetrics"
	RN_ListWorkflowVariables = "listWorkflowVariables"
	RN_GetWorkflowVariable   = "getWorkflowVariable"
	RN_ListWorkflowServices  = "listWorkflowServices"

	RN_CreateWorkflow         = "createWorkflow"
	RN_UpdateWorkflow         = "updateWorkflow"
	RN_SaveWorkflow           = "saveWorkflow"
	RN_DiscardWorkflow        = "discardWorkflow"
	RN_DeleteRevision         = "deleteRevision"
	RN_Tag                    = "tag"
	RN_Untag                  = "untag"
	RN_Retag                  = "retag"
	RN_SetWorkflowVariable    = "setWorkflowVariable"
	RN_EditWorkflowRouter     = "editWorkflowRouter"
	RN_ValidateRef            = "validateRef"
	RN_ValidateRouter         = "validateRouter"
	RN_DeleteWorkflowServices = "deleteWorkflowService"

	RN_ExecuteWorkflow = "executeWorkflow"

	// service and workflow service routes.
	RN_WatchPodLogs = "watchLogs"

	// delete namespace route.
	RN_DeleteNamespace = "deleteNamespace"

	// services get routes.
	RN_ListNamespaceServices = "listNamespacesServices"
	RN_ListNamespacePods     = "listNamespacePods"
	RN_GetNamespaceService   = "getNamespaceService"
	RN_WatchServices         = "watchServices"
	RN_WatchRevisions        = "watchRevisions"

	// services set routes.
	RN_CreateNamespaceService         = "createNamespaceService"
	RN_DeleteNamespaceServices        = "deleteNamespaceService"
	RN_UpdateNamespaceService         = "updateNamespaceService"
	RN_DeleteNamespaceServiceRevision = "deleteNamespaceServiceRevision"

	// events routes.
	RN_EventListeners = "eventListeners"
	RN_EventHistory   = "eventHistory"
	RN_NamespaceEvent = "namespaceEvent"

	// secrets get routes.
	RN_ListSecrets  = "listSecrets"
	RN_SearchSecret = "searchSecret"

	// secrets set routes.
	RN_OverwriteSecret     = "overwriteSecret"
	RN_CreateSecret        = "createSecret"
	RN_CreateSecretsFolder = "createSecretsFolder"

	// secrets delete routes.
	RN_DeleteSecret        = "deleteSecret"
	RN_DeleteSecretsFolder = "deleteSecretsFolder"

	// registries get routes.
	RN_ListRegistries = "listRegistries"
	RN_TestRegistry   = "testRegistry"

	// registries set routes.
	RN_CreateRegistry = "createRegistry"

	// registries delete routes.
	RN_DeleteRegistry = "deleteRegistry"

	// git set routes.
	RN_UpdateMirror         = "updateMirror"
	RN_LockMirror           = "lockMirror"
	RN_SyncMirror           = "syncMirror"
	RN_CancelMirrorActivity = "cancelMirrorActivity"

	// git get routes.
	RN_GetMirrorInfo = "getMirrorInfo"
)

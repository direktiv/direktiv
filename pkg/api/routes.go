package api

// RN = Route Name
const (
	RN_Preflight                      = "preflight"
	RN_ListNamespaces                 = "listNamespaces"
	RN_AddNamespace                   = "addNamespace"
	RN_GetNamespaceConfig             = "getNamespaceConfiguration"
	RN_SetNamespaceConfig             = "setNamespaceConfiguration"
	RN_DeleteNamespace                = "deleteNamespace"
	RN_GetNode                        = "getNode"
	RN_CreateDirectory                = "createDirectory"
	RN_CreateWorkflow                 = "createWorkflow"
	RN_UpdateWorkflow                 = "updateWorkflow"
	RN_SaveWorkflow                   = "saveWorkflow"
	RN_DiscardWorkflow                = "discardWorkflow"
	RN_DeleteNode                     = "deleteNode"
	RN_RenameNode                     = "renameNode"
	RN_GetWorkflowTags                = "getWorkflowTags"
	RN_GetWorkflowRevisions           = "getWorkflowRevisions"
	RN_GetWorkflowRefs                = "getWorkflowRefs"
	RN_DeleteRevision                 = "deleteRevision"
	RN_Tag                            = "tag"
	RN_Untag                          = "untag"
	RN_Retag                          = "retag"
	RN_GetWorkflowRouter              = "getWorkflowRouter"
	RN_GetNamespaceMetrics            = "getNamespaceMetrics"
	RN_GetWorkflowMetrics             = "getWorkflowMetrics"
	RN_ExecuteWorkflow                = "executeWorkflow"
	RN_ListInstances                  = "listInstances"
	RN_GetInstance                    = "getInstance"
	RN_CancelInstance                 = "cancelInstance"
	RN_DeleteNodeAttributes           = "deleteNodeAttributes"
	RN_CreateNodeAttributes           = "createNodeAttributes"
	RN_ListInstanceVariables          = "listInstanceVariables"
	RN_GetInstanceVariable            = "getInstanceVariable"
	RN_SetInstanceVariable            = "setInstanceVariable"
	RN_ListWorkflowVariables          = "listWorkflowVariables"
	RN_GetWorkflowVariable            = "getWorkflowVariable"
	RN_SetWorkflowVariable            = "setWorkflowVariable"
	RN_ListNamespaceVariables         = "listNamespaceVariables"
	RN_GetNamespaceVariable           = "getNamespaceVariable"
	RN_GetServerLogs                  = "getServerLogs"
	RN_GetNamespaceLogs               = "getNamespaceLogs"
	RN_GetWorkflowLogs                = "getWorkflowLogs"
	RN_GetInstanceLogs                = "getInstanceLogs"
	RN_SetNamespaceVariable           = "setNamespaceVariable"
	RN_JQPlayground                   = "jqPlayground"
	RN_Version                        = "version"
	RN_ListServices                   = "listServices"
	RN_ListNamespaceServices          = "listNamespacesServices"
	RN_WatchServices                  = "watchServices"
	RN_WatchRevisions                 = "watchRevisions"
	RN_WatchPods                      = "watchPods"
	RN_WatchPodLogs                   = "watchLogs"
	RN_ListPods                       = "listPods"
	RN_DeleteServices                 = "deleteServices"
	RN_GetService                     = "getService"
	RN_CreateService                  = "createService"
	RN_UpdateService                  = "updateService"
	RN_UpdateServiceTraffic           = "updateServiceTraffic"
	RN_DeleteService                  = "deleteService"
	RN_ListNamespacePods              = "listNamespacePods"
	RN_CreateNamespaceService         = "createNamespaceService"
	RN_DeleteNamespaceServices        = "deleteNamespaceService"
	RN_GetNamespaceService            = "getNamespaceService"
	RN_UpdateNamespaceService         = "updateNamespaceService"
	RN_UpdateNamespaceServiceTraffic  = "updateNamespaceServiceTraffic"
	RN_DeleteNamespaceServiceRevision = "deleteNamespaceServiceRevision"
	RN_ListWorkflowServices           = "listWorkflowServices"
	RN_DeleteWorkflowServices         = "deleteWorkflowService"
	RN_ListWorkflowPods               = "listWorkflowPods"
	RN_PodLogs                        = "podLogs"
	RN_EventListeners                 = "eventListeners"
	RN_EventHistory                   = "eventHistory"
	RN_EditWorkflowRouter             = "editWorkflowRouter"
	RN_ValidateRef                    = "validateRef"
	RN_ValidateRouter                 = "validateRouter"
	RN_NamespaceEvent                 = "namespaceEvent"

	// filter get
	RN_NamespaceEventFilter      = "namespaceEventFilter"
	RN_ListNamespaceEventFilters = "listNamespaceEventFilters"
	RN_GetNamespaceEventFilter   = "getNamespaceEventFilter"

	// filter set
	RN_CreateNamespaceEventFilter = "createNamespaceEventFilter"
	RN_UpdateNamespaceEventFilter = "updateNamespaceEventFilter"

	// filter delete
	RN_DeleteNamespaceEventFilter = "deleteNamespaceEventFilter"

	// secrets get
	RN_ListSecrets  = "listSecrets"
	RN_SearchSecret = "searchSecret"

	// secrets set
	RN_OverwriteSecret     = "overwriteSecret"
	RN_CreateSecret        = "createSecret"
	RN_CreateSecretsFolder = "createSecretsFolder"

	// secrets delete
	RN_DeleteSecret        = "deleteSecret"
	RN_DeleteSecretsFolder = "deleteSecretsFolder"

	// registries get
	RN_ListRegistries = "listRegistries"
	RN_TestRegistry   = "testRegistry"

	// registries set
	RN_CreateRegistry = "createRegistry"

	// registries delete
	RN_DeleteRegistry = "deleteRegistry"

	// git set
	RN_UpdateMirror         = "updateMirror"
	RN_LockMirror           = "lockMirror"
	RN_SyncMirror           = "syncMirror"
	RN_CancelMirrorActivity = "cancelMirrorActivity"

	// git get
	RN_GetMirrorActivityLogs = "getMirrorActivityLogs"
	RN_GetMirrorInfo         = "getMirrorInfo"
)

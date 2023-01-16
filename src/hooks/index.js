import {useDirektivInstances} from './instances'
import {useDirektivJQPlayground} from './jqplayground'
import {useDirektivNamespaces} from './namespaces'
import {useDirektivNamespaceLogs} from './namespaces/logs'
import {useDirektivRegistries} from './registries'
import {useDirektivGlobalRegistries} from './registries/global'
import {useDirektivGlobalPrivateRegistries} from './registries/global-private'
import {useDirektivSecrets} from './secrets'
import {useDirektivNodes} from './nodes'
import { useDirektivWorkflow } from './workflow'
import {useDirektivWorkflowVariables} from './workflow/variables'
import {useDirektivBroadcastConfiguration} from './event-configuration'
import { useDirektivWorkflowLogs } from './workflow/logs'
import {useDirektivInstanceLogs, useDirektivInstance} from './instance'
import {useDirektivEvents} from './events'
import { useDirektivNamespaceMetrics } from './namespaces/metrics'
import { useDirektivNamespaceVariables } from './namespaces/variables'
import { useDirektivGlobalService, useDirektivGlobalServiceRevision, useDirektivGlobalServices} from './services/global'
import { useDirektivNamespaceService, useDirektivNamespaceServiceRevision, useDirektivNamespaceServices } from './services/namespace'
import { useDirektivWorkflowService, useDirektivWorkflowServiceRevision, useDirektivWorkflowServices } from './services/workflow'
import {useDirektivPodLogs} from './services/logs'
import { useDirektivMirror } from './mirror'
import { useDirektivMirrorLogs } from './mirror/logs'
import {HandleError as UtilHandleError} from "./util"


// Services
// Workflow
export const useWorkflowServices = useDirektivWorkflowServices
export const useWorkflowService = useDirektivWorkflowService
export const useWorkflowServiceRevision = useDirektivWorkflowServiceRevision

// Namespace
export const useNamespaceServices = useDirektivNamespaceServices
export const useNamespaceService = useDirektivNamespaceService
export const useNamespaceServiceRevision = useDirektivNamespaceServiceRevision
// Global
export const useGlobalServices = useDirektivGlobalServices
export const useGlobalService = useDirektivGlobalService
export const useGlobalServiceRevision = useDirektivGlobalServiceRevision

// log hooks
export const usePodLogs = useDirektivPodLogs
export const useInstanceLogs = useDirektivInstanceLogs
export const useWorkflowLogs = useDirektivWorkflowLogs
export const useNamespaceLogs = useDirektivNamespaceLogs

// Variables
export const useWorkflowVariables = useDirektivWorkflowVariables
export const useNamespaceVariables = useDirektivNamespaceVariables

// Metrics
export const useNamespaceMetrics = useDirektivNamespaceMetrics

// Eventing
export const useEvents = useDirektivEvents
export const useBroadcastConfiguration = useDirektivBroadcastConfiguration

// Instances
export const useInstance = useDirektivInstance
export const useInstances = useDirektivInstances

// Explorer
export const useWorkflow = useDirektivWorkflow
export const useNodes = useDirektivNodes

// Mirror
export const useMirror = useDirektivMirror
export const useMirrorLogs = useDirektivMirrorLogs

// Misc
export const useJQPlayground = useDirektivJQPlayground

// Namespaces
export const useNamespaces = useDirektivNamespaces

// Registries
export const useRegistries = useDirektivRegistries
export const useGlobalRegistries = useDirektivGlobalRegistries
export const useGlobalPrivateRegistries = useDirektivGlobalPrivateRegistries

// Secrets
export const useSecrets = useDirektivSecrets

export const HandleError = UtilHandleError
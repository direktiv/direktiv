import {
  useDirektivGlobalService,
  useDirektivGlobalServiceRevision,
  useDirektivGlobalServices,
} from "./services/global";
import { useDirektivInstance, useDirektivInstanceLogs } from "./instance";
import {
  useDirektivNamespaceService,
  useDirektivNamespaceServiceRevision,
  useDirektivNamespaceServices,
} from "./services/namespace";
import {
  useDirektivWorkflowService,
  useDirektivWorkflowServiceRevision,
  useDirektivWorkflowServices,
} from "./services/workflow";

import { HandleError as UtilHandleError } from "./util";
import { useDirektivBroadcastConfiguration } from "./event-configuration";
import { useDirektivEvents } from "./events";
import { useDirektivGlobalPrivateRegistries } from "./registries/global-private";
import { useDirektivGlobalRegistries } from "./registries/global";
import { useDirektivInstances } from "./instances";
import { useDirektivJQPlayground } from "./jqplayground";
import { useDirektivMirror } from "./mirror";
import { useDirektivMirrorLogs } from "./mirror/logs";
import { useDirektivNamespaceLogs } from "./namespaces/logs";
import { useDirektivNamespaceMetrics } from "./namespaces/metrics";
import { useDirektivNamespaceVariables } from "./namespaces/variables";
import { useDirektivNamespaces } from "./namespaces";
import { useDirektivNodes } from "./nodes";
import { useDirektivPodLogs } from "./services/logs";
import { useDirektivRegistries } from "./registries";
import { useDirektivSecrets } from "./secrets";
import { useDirektivWorkflow } from "./workflow";
import { useDirektivWorkflowLogs } from "./workflow/logs";
import { useDirektivWorkflowVariables } from "./workflow/variables";

// Services
// Workflow
export const useWorkflowServices = useDirektivWorkflowServices;
export const useWorkflowService = useDirektivWorkflowService;
export const useWorkflowServiceRevision = useDirektivWorkflowServiceRevision;

// Namespace
export const useNamespaceServices = useDirektivNamespaceServices;
export const useNamespaceService = useDirektivNamespaceService;
export const useNamespaceServiceRevision = useDirektivNamespaceServiceRevision;
// Global
export const useGlobalServices = useDirektivGlobalServices;
export const useGlobalService = useDirektivGlobalService;
export const useGlobalServiceRevision = useDirektivGlobalServiceRevision;

// log hooks
export const usePodLogs = useDirektivPodLogs;
export const useInstanceLogs = useDirektivInstanceLogs;
export const useWorkflowLogs = useDirektivWorkflowLogs;
export const useNamespaceLogs = useDirektivNamespaceLogs;

// Variables
export const useWorkflowVariables = useDirektivWorkflowVariables;
export const useNamespaceVariables = useDirektivNamespaceVariables;

// Metrics
export const useNamespaceMetrics = useDirektivNamespaceMetrics;

// Eventing
export const useEvents = useDirektivEvents;
export const useBroadcastConfiguration = useDirektivBroadcastConfiguration;

// Instances
export const useInstance = useDirektivInstance;
export const useInstances = useDirektivInstances;

// Explorer
export const useWorkflow = useDirektivWorkflow;
export const useNodes = useDirektivNodes;

// Mirror
export const useMirror = useDirektivMirror;
export const useMirrorLogs = useDirektivMirrorLogs;

// Misc
export const useJQPlayground = useDirektivJQPlayground;

// Namespaces
export const useNamespaces = useDirektivNamespaces;

// Registries
export const useRegistries = useDirektivRegistries;
export const useGlobalRegistries = useDirektivGlobalRegistries;
export const useGlobalPrivateRegistries = useDirektivGlobalPrivateRegistries;

// Secrets
export const useSecrets = useDirektivSecrets;

export const HandleError = UtilHandleError;

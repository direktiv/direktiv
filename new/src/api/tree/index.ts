import { forceLeadingSlash } from "./utils";

export const treeKeys = {
  nodeContent: (
    namespace: string,
    {
      apiKey,
      path,
      revision,
    }: { apiKey?: string; path?: string; revision?: string }
  ) =>
    [
      {
        scope: "tree-node-content",
        apiKey,
        namespace,
        path: forceLeadingSlash(path ?? "/"),
        revision,
      },
    ] as const,
  revisionsList: (
    namespace: string,
    { apiKey, path }: { apiKey?: string; path?: string }
  ) =>
    [
      {
        scope: "tree-revisions-list",
        apiKey,
        namespace,
        path: forceLeadingSlash(path ?? "/"),
      },
    ] as const,
  tagsList: (
    namespace: string,
    { apiKey, path }: { apiKey?: string; path?: string }
  ) =>
    [
      {
        scope: "tree-tags-list",
        apiKey,
        namespace,
        path: forceLeadingSlash(path ?? ""),
      },
    ] as const,
  router: (
    namespace: string,
    { apiKey, path }: { apiKey?: string; path?: string }
  ) =>
    [
      {
        scope: "tree-router",
        apiKey,
        namespace,
        path: forceLeadingSlash(path ?? ""),
      },
    ] as const,
  workflowVariablesList: (
    namespace: string,
    { apiKey, path }: { apiKey?: string; path: string }
  ) =>
    [
      {
        scope: "workflow-variables-list",
        apiKey,
        namespace,
        path: forceLeadingSlash(path),
      },
    ] as const,
  workflowVariableContent: (
    namespace: string,
    { apiKey, path, name }: { apiKey?: string; path: string; name: string }
  ) =>
    [
      {
        scope: "workflow-variable-content",
        name,
        apiKey,
        namespace,
        path: forceLeadingSlash(path),
      },
    ] as const,
  mirrorInfo: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "mirror-info",
        apiKey,
        namespace,
      },
    ] as const,
  activityLog: (
    namespace: string,
    { activityId, apiKey }: { activityId: string; apiKey?: string }
  ) =>
    [
      {
        scope: "activity-log",
        activityId,
        apiKey,
        namespace,
      },
    ] as const,
  metrics: (
    namespace: string,
    {
      apiKey,
      path,
      type,
    }: { apiKey?: string; path?: string; type: "successful" | "failed" }
  ) =>
    [
      {
        scope: "metrics",
        type,
        apiKey,
        namespace,
        path: forceLeadingSlash(path),
      },
    ] as const,
};

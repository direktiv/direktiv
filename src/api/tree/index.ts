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
};

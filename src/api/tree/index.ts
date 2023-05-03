import { forceLeadingSlash } from "./utils";

export const treeKeys = {
  nodeContent: (
    apiKey: string,
    namespace: string,
    path: string,
    revision: string
  ) =>
    [
      {
        scope: "tree-node-content",
        apiKey,
        namespace,
        path: forceLeadingSlash(path),
        revision: revision,
      },
    ] as const,
  revisionsList: (apiKey: string, namespace: string, path: string) =>
    [
      {
        scope: "tree-revisions-list",
        apiKey,
        namespace,
        path: path ? forceLeadingSlash(path) : "/",
      },
    ] as const,
};

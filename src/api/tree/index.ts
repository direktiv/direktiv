import { forceLeadingSlash } from "./utils";

export const treeKeys = {
  nodeContent: (apiKey: string, namespace: string, path: string) =>
    [
      {
        scope: "tree-node-content",
        apiKey,
        namespace,
        path: path ? forceLeadingSlash(path) : "/",
      },
    ] as const,
};

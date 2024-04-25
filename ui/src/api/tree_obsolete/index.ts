import { forceLeadingSlash } from "../files/utils";

export const treeKeys = {
  nodeContent: (
    namespace: string,
    { apiKey, path }: { apiKey?: string; path?: string }
  ) =>
    [
      {
        scope: "tree-node-content",
        apiKey,
        namespace,
        path: forceLeadingSlash(path ?? "/"),
      },
    ] as const,
};

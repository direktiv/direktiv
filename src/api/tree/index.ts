import { forceLeadingSlash } from "./utils";

export const treeKeys = {
  all: (apiKey: string, namespace: string, path: string) =>
    [
      {
        scope: "tree",
        apiKey,
        namespace,
        path: path ? forceLeadingSlash(path) : "/",
      },
    ] as const,
};

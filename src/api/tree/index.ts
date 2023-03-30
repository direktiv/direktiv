import { forceSlashIfPath } from "./utils";

export const treeKeys = {
  all: (apiKey: string, namespace: string, path: string) =>
    [
      { scope: "tree", apiKey, namespace, path: forceSlashIfPath(path) },
    ] as const,
};

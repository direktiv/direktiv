export const namespaceKeys = {
  all: (apiKey: string, namespace: string, path: string) =>
    [{ scope: "tree", apiKey, namespace, path }] as const,
};

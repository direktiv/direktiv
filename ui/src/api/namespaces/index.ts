export const namespaceKeys = {
  all: (apiKey: string | undefined) =>
    [{ scope: "namespace-list", apiKey }] as const,
  logs: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [{ scope: "namespace-logs", apiKey, namespace }] as const,
};

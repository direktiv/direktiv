export const namespaceKeys = {
  all: (apiKey: string | undefined) =>
    [{ scope: "namespace-list", apiKey }] as const,
  logs: (apiKey: string | undefined) =>
    [{ scope: "namespace-logs", apiKey }] as const,
};

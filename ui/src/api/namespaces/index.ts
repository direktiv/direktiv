export const namespaceKeys = {
  all: (apiKey: string | undefined) =>
    [{ scope: "namespace-list", apiKey }] as const,
};

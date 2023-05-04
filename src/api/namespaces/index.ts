export const namespaceKeys = {
  all: (apiKey: string | undefined) =>
    [{ scope: "namespaces", apiKey }] as const,
};

export const namespaceKeys = {
  all: (apiKey: string) => [{ scope: "namespaces", apiKey }] as const,
};

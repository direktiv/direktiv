export const policyKeys = {
  get: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [{ scope: "policy", apiKey, namespace }] as const,
};

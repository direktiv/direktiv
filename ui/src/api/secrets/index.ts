export const secretKeys = {
  secretsList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "secrets-list",
        apiKey,
        namespace,
      },
    ] as const,
};

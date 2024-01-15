export const secretKeys = {
  secretsList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "secret-list",
        apiKey,
        namespace,
      },
    ] as const,
};

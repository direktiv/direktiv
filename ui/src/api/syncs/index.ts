export const syncKeys = {
  syncsList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "sync-list",
        apiKey,
        namespace,
      },
    ] as const,
};

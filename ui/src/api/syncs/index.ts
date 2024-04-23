export const syncKeys = {
  syncsList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "syncs-list",
        apiKey,
        namespace,
      },
    ] as const,
};

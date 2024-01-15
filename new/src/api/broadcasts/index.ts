export const broadcastKeys = {
  broadcasts: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "broadcasts",
        apiKey,
        namespace,
      },
    ] as const,
};

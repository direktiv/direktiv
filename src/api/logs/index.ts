export const instanceKeys = {
  detail: (
    namespace: string,
    { apiKey, instanceId }: { apiKey?: string; instanceId: string }
  ) =>
    [
      {
        scope: "log-detail",
        apiKey,
        namespace,
        instanceId,
      },
    ] as const,
};

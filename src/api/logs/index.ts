export const logKeys = {
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

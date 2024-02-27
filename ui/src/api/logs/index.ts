export const logKeys = {
  detail: (
    namespace: string,
    { apiKey, instanceId }: { apiKey?: string; instanceId: string; j }
  ) =>
    [
      {
        scope: "logs",
        apiKey,
        namespace,
        instanceId,
      },
    ] as const,
};

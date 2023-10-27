export const serviceKeys = {
  servicesList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "service-list",
        apiKey,
        namespace,
      },
    ] as const,
  servicePods: (
    namespace: string,
    {
      apiKey,
      serviceId,
    }: {
      apiKey?: string;
      serviceId: string;
    }
  ) =>
    [
      {
        scope: "service-pods",
        apiKey,
        namespace,
        serviceId,
      },
    ] as const,
  podLogs: ({
    apiKey,
    namespace,
    serviceId,
    podId,
  }: {
    apiKey?: string;
    namespace: string;
    serviceId: string;
    podId: string;
  }) =>
    [
      {
        scope: "service-pod-logs",
        apiKey,
        namespace,
        serviceId,
        podId,
      },
    ] as const,
};

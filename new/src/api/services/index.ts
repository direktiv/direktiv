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
      service,
    }: {
      apiKey?: string;
      service: string;
    }
  ) =>
    [
      {
        scope: "service-pods",
        apiKey,
        namespace,
        service,
      },
    ] as const,
  podLogs: ({
    apiKey,
    namespace,
    service,
    pod,
  }: {
    apiKey?: string;
    namespace: string;
    service: string;
    pod: string;
  }) =>
    [
      {
        scope: "service-pod-logs",
        apiKey,
        namespace,
        service,
        pod,
      },
    ] as const,
};

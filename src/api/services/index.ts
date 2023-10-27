export const serviceKeys = {
  servicesList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "service-list",
        apiKey,
        namespace,
      },
    ] as const,
  serviceDetail: (
    namespace: string,
    {
      apiKey,
      service,
      workflow,
      version,
    }: { apiKey?: string; service: string; workflow?: string; version?: string }
  ) =>
    [
      {
        scope: "service-detail",
        apiKey,
        namespace,
        service,
        workflow,
        version,
      },
    ] as const,
  serviceRevisionDetail: (
    namespace: string,
    {
      apiKey,
      service,
      revision,
    }: { apiKey?: string; service: string; revision: string }
  ) =>
    [
      {
        scope: "service-detail-revision",
        apiKey,
        namespace,
        service,
        revision,
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
        scope: "service-pods-logs",
        apiKey,
        namespace,
        service,
        pod,
      },
    ] as const,
};

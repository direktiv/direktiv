export const serviceKeys = {
  servicesList: (
    namespace: string,
    { apiKey, workflow }: { apiKey?: string; workflow?: string }
  ) =>
    [
      {
        scope: "service-list",
        apiKey,
        namespace,
        workflow,
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
      service,
      revision,
      workflow,
      version,
    }: {
      apiKey?: string;
      service: string;
      revision: string;
      workflow?: string;
      version?: string;
    }
  ) =>
    [
      {
        scope: "service-pods",
        apiKey,
        namespace,
        service,
        revision,
        workflow,
        version,
      },
    ] as const,
  podLogs: ({
    apiKey,
    name,
    namespace,
  }: {
    apiKey?: string;
    name: string;
    namespace: string;
  }) =>
    [
      {
        scope: "service-pods-logs",
        apiKey,
        name,
        namespace,
      },
    ] as const,
};

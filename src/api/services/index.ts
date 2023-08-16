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
    { apiKey, service }: { apiKey?: string; service: string }
  ) =>
    [
      {
        scope: "service-detail",
        apiKey,
        namespace,
        service,
      },
    ] as const,
};

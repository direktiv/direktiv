export const gatewayKeys = {
  routes: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "gateway-routes",
        apiKey,
        namespace,
      },
    ] as const,
  consumers: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "gateway-consumers",
        apiKey,
        namespace,
      },
    ] as const,
  info: (
    namespace: string,
    { apiKey, expand }: { apiKey?: string; expand?: boolean; server?: string }
  ) =>
    [
      {
        scope: "gateway-info",
        apiKey,
        namespace,
        expand,
      },
    ] as const,
  documentation: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "gateway-documentation",
        apiKey,
        namespace,
      },
    ] as const,
} as const;

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
  docs: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "docs",
        apiKey,
        namespace,
      },
    ] as const,
};

export const gatewayKeys = {
  routes: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "gateway-routes",
        apiKey,
        namespace,
      },
    ] as const,
};

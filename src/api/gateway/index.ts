export const gatewayKeys = {
  routes: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "gateway-routes",
        apiKey,
        namespace,
      },
    ] as const,
  // TODO: REMOVE
  plugins: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "gateway-plugins",
        apiKey,
        namespace,
      },
    ] as const,
};

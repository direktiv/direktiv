export const gatewayKeys = {
  endpoints: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "gateway-endpoints",
        apiKey,
        namespace,
      },
    ] as const,
  plugins: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "gateway-plugins",
        apiKey,
        namespace,
      },
    ] as const,
};

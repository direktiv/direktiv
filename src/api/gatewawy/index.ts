export const gatewayKeys = {
  servicesList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "gateway-list",
        apiKey,
        namespace,
      },
    ] as const,
};

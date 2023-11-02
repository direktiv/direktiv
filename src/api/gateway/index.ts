export const gatewayKeys = {
  endpointList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "endpoint-list",
        apiKey,
        namespace,
      },
    ] as const,
};

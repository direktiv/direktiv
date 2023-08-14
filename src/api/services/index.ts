export const serviceKeys = {
  servicesList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "service-list",
        apiKey,
        namespace,
      },
    ] as const,
};

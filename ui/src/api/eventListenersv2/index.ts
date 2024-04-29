export const eventListenerKeys = {
  eventListenersList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "event-listeners-list",
        apiKey,
        namespace,
      },
    ] as const,
};

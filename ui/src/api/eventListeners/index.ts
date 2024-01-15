export const eventListenerKeys = {
  eventListenersList: (
    namespace: string,
    {
      apiKey,
      limit,
      offset,
    }: { apiKey?: string; limit: number; offset: number }
  ) =>
    [
      {
        scope: "event-listeners-list",
        apiKey,
        namespace,
        limit,
        offset,
      },
    ] as const,
};

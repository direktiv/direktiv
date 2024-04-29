export const eventKeys = {
  eventsList: (
    namespace: string,
    {
      apiKey,
    }: {
      apiKey?: string;
    }
  ) =>
    [
      {
        scope: "events-list",
        apiKey,
        namespace,
      },
    ] as const,
};

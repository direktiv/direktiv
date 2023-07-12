export const instanceKeys = {
  instancesList: (
    namespace: string,
    {
      apiKey,
      limit,
      offset,
      filter,
    }: { apiKey?: string; limit: number; offset: number; filter: string }
  ) =>
    [
      {
        scope: "instance-list",
        apiKey,
        namespace,
        limit,
        offset,
        filter,
      },
    ] as const,
};

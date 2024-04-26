export const instanceKeys = {
  instancesList: (
    namespace: string,
    {
      apiKey,
      limit,
      offset,
    }: { apiKey?: string; limit?: number; offset?: number }
  ) =>
    [
      {
        scope: "instance-list",
        apiKey,
        namespace,
        limit,
        offset,
      },
    ] as const,
};

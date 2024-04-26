export const instanceKeys = {
  instancesList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "instance-list",
        apiKey,
        namespace,
      },
    ] as const,
};

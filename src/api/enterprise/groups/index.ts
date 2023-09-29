export const groupKeys = {
  groupList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "group-list",
        apiKey,
        namespace,
      },
    ] as const,
};

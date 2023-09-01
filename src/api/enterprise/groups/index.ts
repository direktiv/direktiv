export const groupsKeys = {
  groupList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "groups",
        apiKey,
        namespace,
      },
    ] as const,
};

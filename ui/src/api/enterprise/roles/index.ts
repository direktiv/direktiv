export const roleKeys = {
  roleList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "role-list",
        apiKey,
        namespace,
      },
    ] as const,
};

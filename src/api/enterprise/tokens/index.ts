export const tokenKeys = {
  tokenList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "token-list",
        apiKey,
        namespace,
      },
    ] as const,
};

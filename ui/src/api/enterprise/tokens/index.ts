export const tokenKeys = {
  apiTokens: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "api-tokens-list",
        apiKey,
        namespace,
      },
    ] as const,
};

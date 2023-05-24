export const varKeys = {
  varList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "var-list",
        apiKey,
        namespace,
      },
    ] as const,
};

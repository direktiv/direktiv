export const lintingKeys = {
  getLinting: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "linting",
        apiKey,
        namespace,
      },
    ] as const,
};

export const varKeys = {
  varList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "var-list",
        apiKey,
        namespace,
      },
    ] as const,
  varContent: (
    namespace: string,
    {
      apiKey,
      name,
    }: {
      apiKey?: string;
      name: string;
    }
  ) =>
    [
      {
        scope: "var-content",
        apiKey,
        namespace,
        name,
      },
    ] as const,
};

export const varKeys = {
  varList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "variables-list",
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
        scope: "variables-content",
        apiKey,
        namespace,
        name,
      },
    ] as const,
};

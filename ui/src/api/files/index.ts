export const fileKeys = {
  file: (
    namespace: string,
    { apiKey, path }: { apiKey?: string; path?: string }
  ) =>
    [
      {
        scope: "file",
        apiKey,
        namespace,
        path,
      },
    ] as const,
};

export const pathKeys = {
  paths: (
    namespace: string,
    { apiKey, path }: { apiKey?: string; path?: string }
  ) =>
    [
      {
        scope: "paths",
        apiKey,
        namespace,
        path,
      },
    ] as const,
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

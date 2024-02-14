export const nodeKeys = {
  nodesList: (
    namespace: string,
    { apiKey, path }: { apiKey?: string; path?: string }
  ) =>
    [
      {
        scope: "nodes-list",
        apiKey,
        namespace,
        path,
      },
    ] as const,
};

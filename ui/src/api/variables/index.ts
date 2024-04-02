export const varKeys = {
  varList: (
    namespace: string,
    { apiKey, workflowPath }: { apiKey?: string; workflowPath?: string }
  ) =>
    [
      {
        scope: "variables-list",
        apiKey,
        workflowPath,
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

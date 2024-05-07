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
  varDetails: (
    namespace: string,
    {
      apiKey,
      id,
    }: {
      apiKey?: string;
      id: string;
    }
  ) =>
    [
      {
        scope: "variable-details",
        apiKey,
        namespace,
        id,
      },
    ] as const,
};

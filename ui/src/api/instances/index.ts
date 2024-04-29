export const instanceKeys = {
  instancesList: (
    namespace: string,
    {
      apiKey,
      limit,
      offset,
    }: { apiKey?: string; limit?: number; offset?: number }
  ) =>
    [
      {
        scope: "instance-list",
        apiKey,
        namespace,
        limit,
        offset,
      },
    ] as const,
  instancesDetails: (
    namespace: string,
    { apiKey, instanceId }: { apiKey?: string; instanceId: string }
  ) =>
    [
      {
        scope: "instance-details",
        apiKey,
        namespace,
        instanceId,
      },
    ] as const,
  instancesInput: (
    namespace: string,
    { apiKey, instanceId }: { apiKey?: string; instanceId: string }
  ) =>
    [
      {
        scope: "instance-input",
        apiKey,
        namespace,
        instanceId,
      },
    ] as const,
};

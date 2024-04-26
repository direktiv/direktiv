import { FiltersObj } from "./query/get";

export const instanceKeys = {
  instancesList: (
    namespace: string,
    {
      apiKey,
      limit,
      offset,
      filters,
    }: { apiKey?: string; limit: number; offset: number; filters: FiltersObj }
  ) =>
    [
      {
        scope: "instance-list-obsolete",
        apiKey,
        namespace,
        limit,
        offset,
        filters,
      },
    ] as const,
  instanceDetail: (
    namespace: string,
    { apiKey, instanceId }: { apiKey?: string; instanceId: string }
  ) =>
    [
      {
        scope: "instance-detail-obsolete",
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
        scope: "instance-input-obsolete",
        apiKey,
        namespace,
        instanceId,
      },
    ] as const,
  instancesOutput: (
    namespace: string,
    { apiKey, instanceId }: { apiKey?: string; instanceId: string }
  ) =>
    [
      {
        scope: "instance-output-obsolete",
        apiKey,
        namespace,
        instanceId,
      },
    ] as const,
};

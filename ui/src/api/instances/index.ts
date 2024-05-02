import { FiltersObj } from "./query/utils";

export const instanceKeys = {
  instancesList: (
    namespace: string,
    {
      apiKey,
      limit,
      offset,
      filters,
    }: {
      apiKey?: string;
      limit?: number;
      offset?: number;
      filters?: FiltersObj;
    }
  ) =>
    [
      {
        scope: "instance-list",
        apiKey,
        namespace,
        limit,
        offset,
        filters,
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
  instancesOutput: (
    namespace: string,
    { apiKey, instanceId }: { apiKey?: string; instanceId: string }
  ) =>
    [
      {
        scope: "instance-output",
        apiKey,
        namespace,
        instanceId,
      },
    ] as const,
};

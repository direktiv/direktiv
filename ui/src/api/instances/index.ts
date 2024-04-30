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
      filters?: FiltersObj;
      limit?: number;
      offset?: number;
    }
  ) =>
    [
      {
        scope: "instance-list",
        apiKey,
        namespace,
        filters,
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

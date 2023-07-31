import { FiltersObj as FiltersObjDetails } from "./query/details";
import { FiltersObj as FiltersObjList } from "./query/get";

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
      limit: number;
      offset: number;
      filters: FiltersObjList;
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
  instanceDetail: (
    namespace: string,
    {
      apiKey,
      instanceId,
      filters,
    }: { apiKey?: string; instanceId: string; filters: FiltersObjDetails }
  ) =>
    [
      {
        scope: "instance-detail",
        apiKey,
        namespace,
        instanceId,
        filters,
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

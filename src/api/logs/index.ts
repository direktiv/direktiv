import { FiltersObj } from "./query/get";

export const logKeys = {
  detail: (
    namespace: string,
    {
      apiKey,
      instanceId,
      filters,
    }: { apiKey?: string; instanceId: string; filters: FiltersObj }
  ) =>
    [
      {
        scope: "log-detail",
        apiKey,
        namespace,
        instanceId,
        filters,
      },
    ] as const,
};

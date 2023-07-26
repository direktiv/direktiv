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
        scope: "instance-list",
        apiKey,
        namespace,
        limit,
        offset,
        filters,
      },
    ] as const,
};

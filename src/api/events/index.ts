import { FiltersObj } from "./query/get";

export const eventKeys = {
  eventsList: (
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
        scope: "event-list",
        apiKey,
        namespace,
        limit,
        offset,
        filters,
      },
    ] as const,
};

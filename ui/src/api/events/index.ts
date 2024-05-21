import { FiltersSchemaType } from "./schema/filters";

export const eventKeys = {
  eventsList: (
    namespace: string,
    {
      apiKey,
      filters,
    }: {
      apiKey?: string;
      filters?: FiltersSchemaType;
    }
  ) =>
    [
      {
        scope: "events-list",
        apiKey,
        namespace,
        filters,
      },
    ] as const,
};

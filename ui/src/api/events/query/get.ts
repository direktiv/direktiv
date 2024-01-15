import { EventsListSchema, EventsListSchemaType } from "../schema";
import { QueryFunctionContext, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { eventKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";
import { useStreaming } from "~/api/streaming";

/**
 * Filtering events:
 * type = e.g. "com.github.pull.create"
 * text searches through the full JSON representation of the event
 */
export type FiltersObj = {
  TYPE?: { type: "CONTAINS"; value: string };
  TEXT?: { type: "CONTAINS"; value: string };
  AFTER?: { type: "AFTER"; value: Date };
  BEFORE?: { type: "BEFORE"; value: Date };
};

// TODO: this same method is duplicated in several places, extract and import?
export const getFilterQuery = (filters: FiltersObj) => {
  let query = "";
  const filterFields = Object.keys(filters) as Array<keyof FiltersObj>;

  filterFields.forEach((field) => {
    const filterItem = filters[field];

    // Without the guard, TS thinks filterItem may be undefined
    if (!filterItem) {
      return console.error("filterItem is not defined");
    }

    let queryField: string;
    let queryValue: string;

    if (field === "AFTER" || field === "BEFORE") {
      const date = filters[field]?.value;
      if (!date) {
        throw new Error("date is not defined in date filter");
      }
      queryField = "CREATED";
      queryValue = date.toISOString();
    } else {
      const value = filters[field]?.value;
      if (!value) {
        throw new Error("filter value is not defined");
      }
      queryField = field;
      queryValue = value;
    }

    query = query.concat(
      `&filter.field=${queryField}&filter.type=${filterItem.type}&filter.val=${queryValue}`
    );
  });

  return query;
};

// TODO: this is also duplicated, abstract it (url needs to be an argument)
const getUrl = ({
  namespace,
  baseUrl,
  limit,
  offset,
  filters,
}: {
  baseUrl?: string;
  namespace: string;
  limit: number;
  offset: number;
  filters?: FiltersObj;
}) => {
  let url = `${
    baseUrl ?? ""
  }/api/namespaces/${namespace}/events?limit=${limit}&offset=${offset}`;
  if (filters) {
    url = url.concat(getFilterQuery(filters));
  }
  return url;
};

export const getEvents = apiFactory({
  url: getUrl,
  method: "GET",
  schema: EventsListSchema,
});

const fetchEvents = async ({
  queryKey: [{ apiKey, namespace, limit, offset, filters }],
}: QueryFunctionContext<ReturnType<(typeof eventKeys)["eventsList"]>>) =>
  getEvents({
    apiKey,
    urlParams: { namespace, limit, offset, filters },
  });

export const useEvents = ({
  limit,
  offset,
  filters,
}: {
  limit: number;
  offset: number;
  filters: FiltersObj;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: eventKeys.eventsList(namespace, {
      apiKey: apiKey ?? undefined,
      limit,
      offset,
      filters,
    }),
    queryFn: fetchEvents,
    enabled: !!namespace,
  });
};

export const useEventsStream = (
  {
    limit,
    offset,
    filters,
  }: {
    limit: number;
    offset: number;
    filters: FiltersObj;
  },
  { enabled = true }: { enabled?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: getUrl({ namespace, offset, limit, filters }),
    apiKey: apiKey ?? undefined,
    enabled,
    schema: EventsListSchema,
    onMessage: (message) => {
      queryClient.setQueryData<EventsListSchemaType>(
        eventKeys.eventsList(namespace, {
          apiKey: apiKey ?? undefined,
          limit,
          offset,
          filters: filters ?? {},
        }),
        // when useStreaming is invoked with offset and limit, it will always
        // return a full page of results on every update, so it is not
        // necessary to merge old and new data like we do in some other cases.
        message
      );
    },
  });
};

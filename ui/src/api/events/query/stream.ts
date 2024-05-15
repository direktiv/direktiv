import {
  EventListItem,
  EventListItemType,
  EventsListResponseSchemaType,
} from "../schema";

import { FiltersSchemaType } from "../schema/filters";
import { buildSearchParamsString } from "~/api/utils";
import { eventKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useQueryClient } from "@tanstack/react-query";
import { useStreaming } from "~/api/streaming";

type EventsQueryParams = {
  namespace?: string;
  before?: string;
  filters?: FiltersSchemaType;
};

type EventsParams = {
  baseUrl?: string;
  namespace: string;
  useStreaming?: boolean;
} & EventsQueryParams;

const updateCache = (
  oldData: EventsCache | undefined,
  newItem: EventListItemType
): EventsCache | undefined => {
  if (oldData === undefined) return undefined;

  // skip update if the log entry is already in the cache
  if (oldData.data.some((item) => item.event.id === newItem.event.id)) {
    return oldData;
  }

  return {
    ...oldData,
    data: [...oldData.data, newItem],
  };
};

const getUrl = (params: EventsParams) => {
  const { baseUrl, namespace, filters } = params;

  const urlPath = `/api/v2/namespaces/${namespace}/events/history/subscribe`;

  let url = `${baseUrl ?? ""}${urlPath}`;

  if (filters) {
    url = url.concat(buildSearchParamsString(filters));
  }

  return url;
};

export type UseEventsStreamParams = EventsQueryParams & { enabled?: boolean };

type EventsCache = EventsListResponseSchemaType;

export const useEventsStream = ({
  enabled,
  ...params
}: UseEventsStreamParams) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: getUrl({
      useStreaming: true,
      namespace,
      ...params,
    }),
    apiKey: apiKey ?? undefined,
    schema: EventListItem,
    enabled,
    onMessage: (msg) => {
      queryClient.setQueryData<EventsCache>(
        eventKeys.eventsList(namespace, {
          apiKey: apiKey ?? undefined,
          filters: params.filters,
        }),
        (oldData) => updateCache(oldData, msg)
      );
    },
  });
};

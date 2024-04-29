import {
  EventListItem,
  EventListItemType,
  EventsListResponseSchemaType,
} from "../schema";
import { InfiniteData, useQueryClient } from "@tanstack/react-query";

import { eventKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

type EventsQueryParams = {
  namespace?: string;
  before?: string;
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

  const pages = oldData.pages;
  const olderPages = pages.slice(0, -1);
  const newestPage = pages.at(-1);
  if (newestPage === undefined) return undefined;

  const newestPageData = newestPage.data ?? [];

  // skip cache if the log entry is already in the cache
  if (newestPageData.some((item) => item.event.id === newItem.event.id)) {
    return oldData;
  }

  return {
    ...oldData,
    pages: [
      ...olderPages,
      {
        ...newestPage,
        data: [...newestPageData, newItem],
      },
    ],
  };
};

const getUrl = (params: EventsParams) => {
  const { baseUrl, namespace, useStreaming } = params;

  let urlPath = `/api/v2/namespaces/${namespace}/events/history`;

  if (useStreaming) {
    urlPath = `${urlPath}/subscribe`;
  }

  return `${baseUrl ?? ""}${urlPath}`;
};

export type UseEventsStreamParams = EventsQueryParams & { enabled?: boolean };

type EventsCache = InfiniteData<EventsListResponseSchemaType>;

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
        }),
        (oldData) => updateCache(oldData, msg)
      );
    },
  });
};

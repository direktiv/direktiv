import {
  EventListenersListSchema,
  EventListenersListSchemaType,
} from "../schema";
import {
  QueryFunctionContext,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { eventListenerKeys } from "..";
import moment from "moment";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

const updateCache = (
  oldData: EventListenersListSchemaType | undefined,
  message: EventListenersListSchemaType
) => {
  if (!oldData) {
    return message;
  }
  /**
   * TODO: copied from events. Does this behave the same?
   *
   * Dedup logs. The onMessage callback gets called in two different cases:
   *
   * case 1:
   * when the SSE connection is established, the whole set of logs is received
   *
   * case 2:
   * after the connection is established and only some new log entries are received
   *
   * it's also important to note that multiple components can subscribe to the same
   * cache, so we can have case 1 and 2 at the same time, or case 1 after case 2
   */

  const lastCachedItem = oldData?.results[0];
  let newResults: typeof oldData.results = [];

  // there was a previous cache, but with no entries yet
  if (!lastCachedItem) {
    newResults = message.results;
    // there was a previous cache with entries
  } else {
    const newestLogTimeFromCache = moment(lastCachedItem.createdAt);
    // new results are all logs that are newer than the last cached log
    newResults = message.results.filter((entry) =>
      newestLogTimeFromCache.isBefore(entry.createdAt)
    );
  }

  return {
    ...oldData,
    results: [...newResults, ...oldData.results],
  };
};

const getUrl = ({
  namespace,
  baseUrl,
  limit,
  offset,
}: {
  baseUrl?: string;
  namespace: string;
  limit: number;
  offset: number;
}) =>
  `${
    baseUrl ?? ""
  }/api/namespaces/${namespace}/event-listeners?limit=${limit}&offset=${offset}`;

export const getEventListeners = apiFactory({
  url: getUrl,
  method: "GET",
  schema: EventListenersListSchema,
});

const fetchEventListeners = async ({
  queryKey: [{ apiKey, namespace, limit, offset }],
}: QueryFunctionContext<
  ReturnType<(typeof eventListenerKeys)["eventListenersList"]>
>) =>
  getEventListeners({
    apiKey,
    urlParams: { namespace, limit, offset },
  });

export const useEventListeners = ({
  limit,
  offset,
}: {
  limit: number;
  offset: number;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: eventListenerKeys.eventListenersList(namespace, {
      apiKey: apiKey ?? undefined,
      limit,
      offset,
    }),
    queryFn: fetchEventListeners,
    enabled: !!namespace,
  });
};

export const useEventListenersStream = (
  {
    limit,
    offset,
  }: {
    limit: number;
    offset: number;
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
    url: getUrl({ namespace, offset, limit }),
    apiKey: apiKey ?? undefined,
    enabled,
    schema: EventListenersListSchema,
    onMessage: (message) => {
      queryClient.setQueryData<EventListenersListSchemaType>(
        eventListenerKeys.eventListenersList(namespace, {
          apiKey: apiKey ?? undefined,
          limit,
          offset,
        }),
        (oldData) => updateCache(oldData, message)
      );
    },
  });
};

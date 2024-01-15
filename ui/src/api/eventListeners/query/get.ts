import {
  EventListenersListSchema,
  EventListenersListSchemaType,
} from "../schema";
import { QueryFunctionContext, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { eventListenerKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";
import { useStreaming } from "~/api/streaming";

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

  return useQueryWithPermissions({
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
        // when useStreaming is invoked with offset and limit, it will always
        // return a full page of results on every update, so it is not
        // necessary to merge old and new data like we do in some other cases.
        message
      );
    },
  });
};

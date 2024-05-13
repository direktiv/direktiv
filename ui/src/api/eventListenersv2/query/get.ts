import { EventListenerResponseSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { eventListenerKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

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
  }/api/v2/namespaces/${namespace}/events/listeners?limit=${limit}&offset=${offset}`;

export const getEventListeners = apiFactory({
  url: getUrl,
  method: "GET",
  schema: EventListenerResponseSchema,
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

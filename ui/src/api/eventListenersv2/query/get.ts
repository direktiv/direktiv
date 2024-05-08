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
}: {
  baseUrl?: string;
  namespace: string;
}) => `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/events/listeners`;

export const getEventListeners = apiFactory({
  url: getUrl,
  method: "GET",
  schema: EventListenerResponseSchema,
});

const fetchEventListeners = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<
  ReturnType<(typeof eventListenerKeys)["eventListenersList"]>
>) =>
  getEventListeners({
    apiKey,
    urlParams: { namespace },
  });

export const useEventListeners = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: eventListenerKeys.eventListenersList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchEventListeners,
    enabled: !!namespace,
  });
};

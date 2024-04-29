import { EventsListResponseSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { eventKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

const getUrl = ({
  namespace,
  baseUrl,
}: {
  baseUrl?: string;
  namespace: string;
}) => {
  const url = `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/events/history`;
  return url;
};

export const getEvents = apiFactory({
  url: getUrl,
  method: "GET",
  schema: EventsListResponseSchema,
});

const fetchEvents = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof eventKeys)["eventsList"]>>) =>
  getEvents({
    apiKey,
    urlParams: { namespace },
  });

export const useEvents = ({
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
    queryKey: eventKeys.eventsList(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchEvents,
    enabled: !!namespace,
  });
};

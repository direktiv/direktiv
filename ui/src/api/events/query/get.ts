import { EventsListResponseSchema } from "../schema";
import { FiltersSchemaType } from "../schema/filters";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { buildSearchParamsString } from "~/api/utils";
import { eventKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

type EventsListParams = {
  enabled?: boolean;
  filters?: FiltersSchemaType;
};

const getUrl = ({
  namespace,
  baseUrl,
  filters,
}: {
  namespace: string;
  baseUrl?: string;
  filters?: FiltersSchemaType;
}) => {
  let url = `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/events/history`;

  if (filters) {
    url = url.concat(buildSearchParamsString(filters));
  }

  return url;
};

export const getEvents = apiFactory({
  url: getUrl,
  method: "GET",
  schema: EventsListResponseSchema,
});

const fetchEvents = async ({
  queryKey: [{ apiKey, namespace, filters }],
}: QueryFunctionContext<ReturnType<(typeof eventKeys)["eventsList"]>>) =>
  getEvents({
    apiKey,
    urlParams: { namespace, filters },
  });

export const useEvents = ({ enabled = true, filters }: EventsListParams) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  /**
   * The API returns data as an infinite list, using the same pattern as in our logs.
   *
   * This response format could also be consumed using useInfiniteQueryWithPermissions.
   *
   * However, we ignore the infinite list feature for this endpoint,
   * fetching only the first page of results. Users can then rely on filters to narrow in
   * on particular events when too many results are returned.
   */
  return useQueryWithPermissions({
    queryKey: eventKeys.eventsList(namespace, {
      apiKey: apiKey ?? undefined,
      filters,
    }),
    queryFn: fetchEvents,
    enabled: !!namespace && enabled,
    refetchOnWindowFocus: false,
  });
};

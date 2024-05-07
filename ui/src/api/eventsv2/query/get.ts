import { EventListItemType, EventsListResponseSchema } from "../schema";

import { FiltersSchemaType } from "../schema/filters";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { buildSearchParamsString } from "~/api/utils";
import { eventKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useInfiniteQueryWithPermissions from "~/api/useInfiniteQueryWithPermissions";
import { useNamespace } from "~/util/store/namespace";

type EventsListParams = {
  enabled: boolean;
  filters: FiltersSchemaType;
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
   * The API returns data as an infinite list, which means it returns a cursor in form of a timestamp
   * to the previous page of data. The end of the list is not known until the last page is reached and
   * the cursor is null.
   *
   * The API only returns navigation into one direction, which means we always have to start with querying
   * the most recent logs and then navigate to older ones. It is not possible to start at a specific time
   * and then move to more recent logs.
   */
  const queryReturn = useInfiniteQueryWithPermissions({
    queryKey: eventKeys.eventsList(namespace, {
      apiKey: apiKey ?? undefined,
      filters,
    }),
    queryFn: fetchEvents,
    getNextPageParam: () => undefined,
    getPreviousPageParam: (firstPage) =>
      firstPage.meta?.previousPage ?? undefined,
    enabled: !!namespace && enabled,
    initialPageParam: undefined,
    refetchOnWindowFocus: false,
  });

  /**
   * expose a simpler data structure to the consumer of the hook by stripping
   * out the pages and flattening the data into a single array
   */
  let logData: EventListItemType[] | undefined = undefined;
  if (queryReturn.data) {
    const pages = queryReturn.data?.pages.map((page) => page.data ?? []) ?? [];
    logData = pages.flat();
  }

  return {
    ...queryReturn,
    data: logData,
  };
};

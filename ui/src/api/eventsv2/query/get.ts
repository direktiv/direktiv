import { EventListItemType, EventsListResponseSchema } from "../schema";

import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { eventKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useInfiniteQueryWithPermissions from "~/api/useInfiniteQueryWithPermissions";
import { useNamespace } from "~/util/store/namespace";

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

export const useEvents = ({ enabled = true }) => {
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

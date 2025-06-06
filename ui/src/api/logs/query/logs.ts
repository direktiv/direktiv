import {
  InfiniteData,
  QueryFunctionContext,
  useQueryClient,
} from "@tanstack/react-query";
import {
  LogEntrySchema,
  LogEntryType,
  LogsSchema,
  LogsSchemaType,
} from "../schema";

import { apiFactory } from "~/api/apiFactory";
import { buildSearchParamsString } from "~/api/utils";
import { logKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import useInfiniteQueryWithPermissions from "~/api/useInfiniteQueryWithPermissions";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

/**
 * example of a InfiniteData<LogsSchemaType> object. All of these
 * data share one cache key. The pages and pageParams properties
 * are part of useInfiniteQuery hook.
  {
    // the result of every page request is stored here
    "pages": [
      {
        "data": [ ... ]
      },
        "data": [ ... ]
      },
      ...
    ]
    // all page pointers are stored here
    "pageParams": [
      null,
      "FIRST_TIMESTAMP",
      "SECOND_TIMESTAMP",
    ]
  }
*/
type LogsCache = InfiniteData<LogsSchemaType>;

const updateCache = (
  oldData: LogsCache | undefined,
  newLogEntry: LogEntryType
): LogsCache | undefined => {
  if (oldData === undefined) return undefined;

  const pages = oldData.pages;
  const olderPages = pages.slice(1);
  const newestPage = pages[0];
  if (newestPage === undefined) throw Error("Newest page not found");

  const newestPageData = newestPage.data ?? [];

  // skip cache if the log entry is already in the cache
  if (newestPageData.some((logEntry) => logEntry.time === newLogEntry.time)) {
    return oldData;
  }

  return {
    ...oldData,
    pages: [
      {
        ...newestPage,
        data: [...newestPageData, newLogEntry],
      },
      ...olderPages,
    ],
  };
};

type LogsQueryParams = {
  instance?: string;
  route?: string;
  activity?: string;
  after?: string;
  before?: string;
  trace?: string;
  last?: number;
};

type LogsParams = {
  baseUrl?: string;
  namespace: string;
  useStreaming?: boolean;
} & LogsQueryParams;

const getUrl = (params: LogsParams) => {
  const { baseUrl, namespace, useStreaming, ...queryParams } = params;

  let urlPath = `/api/v2/namespaces/${namespace}/logs`;

  if (useStreaming) {
    urlPath = `${urlPath}/subscribe`;

    // Add after param from 3 seconds ago to close a potential
    // gap between the GET request and the beginning of the stream
    const threeSecondsAgo = new Date(Date.now() - 3000);
    queryParams.after = threeSecondsAgo.toISOString();
  }

  if (!useStreaming) {
    queryParams.last = 50;
  }

  const queryParamsString = buildSearchParamsString({
    ...queryParams,
  });

  return `${baseUrl ?? ""}${urlPath}${queryParamsString}`;
};

const getLogs = apiFactory({
  url: getUrl,
  method: "GET",
  schema: LogsSchema,
});

const fetchLogs = async ({
  pageParam,
  queryKey: [{ apiKey, namespace, instance, route, activity, trace, last }],
}: QueryFunctionContext<
  ReturnType<(typeof logKeys)["detail"]>,
  LogsQueryParams["before"]
>) =>
  getLogs({
    apiKey,
    urlParams: {
      namespace,
      instance,
      route,
      activity,
      before: pageParam,
      trace,
      last,
    },
  });

export type UseLogsStreamParams = LogsQueryParams & { enabled?: boolean };

export const useLogsStream = ({ enabled, ...params }: UseLogsStreamParams) => {
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
    schema: LogEntrySchema,
    enabled,
    onMessage: (msg) => {
      queryClient.setQueryData<LogsCache>(
        logKeys.detail(namespace, {
          apiKey: apiKey ?? undefined,
          activity: params.activity,
          instance: params.instance,
          route: params.route,
          trace: params.trace,
          last: params.last,
        }),
        (oldData) => updateCache(oldData, msg)
      );
    },
  });
};

export type UseLogsParams = LogsQueryParams & { enabled?: boolean };

export const useLogs = ({
  instance,
  route,
  activity,
  trace,
  enabled = true,
  last,
}: UseLogsParams = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  /**
   * The API returns data as an infinite list, which means it returns a cursor in form of a timestamp
   * to the next page of data. The end of the list is not known until the last page is reached and
   * the cursor is null.
   *
   * The API only returns navigation into one direction, which means we always have to start with querying
   * the most recent logs and then navigate to older ones. It is not possible to start at a specific time
   * and then move to more recent logs.
   */
  const queryReturn = useInfiniteQueryWithPermissions({
    queryKey: logKeys.detail(namespace, {
      apiKey: apiKey ?? undefined,
      instance,
      route,
      activity,
      trace,
      last,
    }),
    queryFn: fetchLogs,

    getNextPageParam: (_, pages) => {
      const lastPage = pages.at(-1);
      if (lastPage?.data.length === 0) {
        return null;
      }
      const oldestTime = lastPage?.data.at(0)?.time;
      return oldestTime;
    },
    enabled: !!namespace && enabled,
    initialPageParam: undefined,
    refetchOnWindowFocus: false,
  });

  /**
   * expose a simpler data structure to the consumer of the hook by stripping
   * out the pages and flattening the data into a single array
   */
  let logData: LogEntryType[] | undefined = undefined;
  if (queryReturn.data) {
    const pagesReversed = [...queryReturn.data.pages].reverse();
    const pages = pagesReversed.map((page) => page.data ?? []) ?? [];
    logData = pages.flat();

    // DEBUG: Throw error if time codes are not in correct order

    // const timestamps = logData.map((item) => item.time);
    // console.log(timestamps);

    // for (let i = 1; i < timestamps.length; i++) {
    //   if (timestamps[i] < timestamps[i - 1]) {
    //     throw new Error(
    //       `Timestamps are not in ascending order: ${timestamps[i - 1]} > ${timestamps[i]}`
    //     );
    //   }
    // }
  }

  return {
    ...queryReturn,
    data: logData,
  };
};

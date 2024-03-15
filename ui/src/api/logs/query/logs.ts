import {
  InfiniteData,
  QueryFunctionContext,
  useInfiniteQuery,
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
import { memo } from "react";
import { useApiKey } from "~/util/store/apiKey";
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
        "previousPage": "FIRST_TIMESTAMP",
        "data": []
      },
      {
        "previousPage": "SECOND_TIMESTAMP",
        "data": []
      },
      {
        "previousPage": null, // last page
        "data": []
      }
    ]
    // all page pointers that were found in the page request results are stored here
    "pageParams": [
      "FIRST_TIMESTAMP",
      "SECOND_TIMESTAMP",
      null
    ]
  }
*/
type LogsCache = InfiniteData<LogsSchemaType>;

const updateCache = (
  oldData: LogsCache | undefined,
  newLogEntry: LogEntryType
): LogsCache => {
  if (oldData === undefined) {
    // TODO:
    console.error("no old cache");
    return {
      pages: [],
      pageParams: [],
    };
  }

  const pages = oldData.pages;
  const olderPages = pages.slice(0, -1);
  const recentPage = pages.at(-1);

  if (recentPage === undefined) {
    // TODO:
    console.error("no recentPage");
    return {
      pages: [],
      pageParams: [],
    };
  }

  const recentPageData = recentPage.data ?? [];

  if (recentPageData.some((logEntry) => logEntry.id === newLogEntry.id)) {
    console.error(
      `skipping cache update, log entry ${newLogEntry.id} already exists`
    );
    return oldData;
  }

  return {
    ...oldData,
    pages: [
      ...olderPages,
      {
        ...recentPage,
        data: [...recentPageData, newLogEntry],
      },
    ],
  };
};

type LogsQueryParams = {
  instance?: string;
  route?: string;
  activity?: string;
  before?: string;
  trace?: string;
};

type LogsParams = {
  baseUrl?: string;
  namespace: string;
  useStreaming?: boolean;
} & LogsQueryParams;

const getUrl = (params: LogsParams) => {
  const { baseUrl, namespace, useStreaming, ...queryParams } = params;
  /**
   * to avoid a gap in the log entries, we will pass the optional "after"
   * parameter to the streaming url to start streaming the logs from 5
   * seconds ago.
   *
   * This comes with the caveat that we might get log entries that we
   * already have in the cache. So we have to filter duplicates when
   * populating the cache.
   */
  const now = new Date();
  now.setSeconds(now.getSeconds() - 5);
  const after = now.toISOString();
  const queryParamsString = buildSearchParamsString({
    ...queryParams,
    after,
  });

  let urlPath = `/api/v2/namespaces/${namespace}/logs`;
  if (useStreaming) {
    urlPath = `${urlPath}/subscribe`;
  }

  return `${baseUrl ?? ""}${urlPath}${queryParamsString}`;
};

const getLogs = apiFactory({
  url: getUrl,
  method: "GET",
  schema: LogsSchema,
});

const fetchLogs = async ({
  pageParam,
  queryKey: [{ apiKey, namespace, instance, route, activity, trace }],
}: QueryFunctionContext<ReturnType<(typeof logKeys)["detail"]>>) =>
  getLogs({
    apiKey,
    urlParams: {
      namespace,
      instance,
      route,
      activity,
      before: pageParam,
      trace,
    },
  });

export const useLogsStream = (params: LogsQueryParams) => {
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
    onMessage: (msg) => {
      queryClient.setQueryData<LogsCache>(
        logKeys.detail(namespace, {
          apiKey: apiKey ?? undefined,
          activity: params.activity,
          instance: params.instance,
          route: params.route,
          trace: params.trace,
        }),
        (oldData) => updateCache(oldData, msg)
      );
    },
  });
};

export const LogStreamingSubscriber = memo((params: LogsQueryParams) => {
  useLogsStream(params);
  return null;
});

LogStreamingSubscriber.displayName = "LogStreamingSubscriber";

export const useLogs = ({
  instance,
  route,
  activity,
  trace,
}: LogsQueryParams) => {
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
  return useInfiniteQuery({
    queryKey: logKeys.detail(namespace, {
      apiKey: apiKey ?? undefined,
      instance,
      route,
      activity,
      trace,
    }),
    queryFn: fetchLogs,
    getPreviousPageParam: (firstPage) =>
      firstPage.meta?.previousPage ?? undefined,
    enabled: !!namespace,
    refetchOnWindowFocus: false,
  });
};

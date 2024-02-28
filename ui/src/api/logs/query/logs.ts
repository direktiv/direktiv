import { LogEntrySchema, LogsSchema } from "../schema";
import { QueryFunctionContext, useInfiniteQuery } from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { buildSearchParamsString } from "~/api/utils";
import { logKeys } from "..";
import { memo } from "react";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

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

  const queryParamsString = buildSearchParamsString({
    ...queryParams,
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
  // const queryClient = useQueryClient();

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
      console.log("🚀 received a msg", msg);
      // queryClient.setQueryData<LogsSchemaType>(
      //   logKeys.detail(namespace, {
      //     apiKey: apiKey ?? undefined,
      //     instanceId,
      //     filters: filters ?? {},
      //   }),
      //   (oldData) => updateCache(oldData, msg)
      // );
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
   * to the next page of data. The end of the list is not known until the last page is reached and the
   * cursor is null.
   *
   * The API only returns navigation into one direction, which means we always have to start with querying
   * the most recent logs and then navigate to older ones. It is not possible to start at a specific time
   * and then move to newer logs.
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
    getNextPageParam: (firstPage) =>
      firstPage.next_page === "" ? undefined : firstPage.next_page,
    getPreviousPageParam: () => undefined,
    enabled: !!namespace,
  });
};

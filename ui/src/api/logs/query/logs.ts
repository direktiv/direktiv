import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { LogsSchema } from "../schema";
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
} & LogsQueryParams;

const getUrl = (params: LogsParams) => {
  const { baseUrl, namespace, ...queryParams } = params;

  const queryParamsString = buildSearchParamsString({
    ...queryParams,
  });

  return `${
    baseUrl ?? ""
  }/api/v2/namespaces/${namespace}/logs${queryParamsString}`;
};

const getLogs = apiFactory({
  url: getUrl,
  method: "GET",
  schema: LogsSchema,
});

const fetchLogs = async ({
  queryKey: [{ apiKey, namespace, instance, route, activity, before, trace }],
}: QueryFunctionContext<ReturnType<(typeof logKeys)["detail"]>>) =>
  getLogs({
    apiKey,
    urlParams: {
      namespace,
      instance,
      route,
      activity,
      before,
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
      namespace,
      ...params,
    }),
    apiKey: apiKey ?? undefined,
    schema: LogsSchema,
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
  before,
  trace,
}: LogsQueryParams) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: logKeys.detail(namespace, {
      apiKey: apiKey ?? undefined,
      instance,
      route,
      activity,
      before,
      trace,
    }),
    queryFn: fetchLogs,
    enabled: !!namespace,
  });
};

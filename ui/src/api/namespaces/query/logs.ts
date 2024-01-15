import { NamespaceLogListSchema, NamespaceLogListSchemaType } from "../schema";
import { QueryFunctionContext, useQueryClient } from "@tanstack/react-query";

import { apiFactory } from "~/api/apiFactory";
import { memo } from "react";
import moment from "moment";
import { namespaceKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";
import { useStreaming } from "~/api/streaming";

const updateCache = (
  oldData: NamespaceLogListSchemaType | undefined,
  msg: NamespaceLogListSchemaType
) => {
  if (!oldData) {
    return msg;
  }
  /**
   * Dedup logs. The onMessage callback gets called in two different cases:
   *
   * case 1:
   * when the SSE connection is established, the whole set of logs is received
   *
   * case 2:
   * after the connection is established and only some new log entries are received
   *
   * it's also important to note that multiple components can subscribe to the same
   * cache, so we can have case 1 and 2 at the same time, or case 1 after case 2
   */
  const lastCachedLog = oldData.results[oldData.results.length - 1];
  let newResults: typeof oldData.results = [];

  // there was a previous cache, but with no entries yet
  if (!lastCachedLog) {
    newResults = msg.results;
    // there was a previous cache with entries
  } else {
    const newestLogTimeFromCache = moment(lastCachedLog.t);
    // new results are all logs that are newer than the last cached log
    newResults = msg.results.filter((entry) =>
      newestLogTimeFromCache.isBefore(entry.t)
    );
  }

  return {
    ...oldData,
    results: [...oldData.results, ...newResults],
  };
};

export const getInstanceDetails = apiFactory({
  url: ({ namespace, baseUrl }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/namespaces/${namespace}/logs`,
  method: "GET",
  schema: NamespaceLogListSchema,
});

const fetchInstanceDetails = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof namespaceKeys)["logs"]>>) =>
  getInstanceDetails({
    apiKey,
    urlParams: { namespace },
  });

export const useNamespaceLogsStream = ({
  enabled = true,
}: { enabled?: boolean } = {}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: `/api/namespaces/${namespace}/logs`,
    apiKey: apiKey ?? undefined,
    enabled,
    schema: NamespaceLogListSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<NamespaceLogListSchemaType>(
        namespaceKeys.logs(namespace, { apiKey: apiKey ?? undefined }),
        (oldData) => updateCache(oldData, msg)
      );
    },
  });
};

type LogStreamingSubscriberTypeProps = { enabled?: boolean };

export const NamespaceLogsStreamingSubscriber = memo(
  ({ enabled }: LogStreamingSubscriberTypeProps) => {
    useNamespaceLogsStream({ enabled: enabled ?? true });
    return null;
  }
);
NamespaceLogsStreamingSubscriber.displayName =
  "NamespaceLogsStreamingSubscriber";

export const useNamespacelogs = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: namespaceKeys.logs(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchInstanceDetails,
    enabled: !!namespace,
  });
};

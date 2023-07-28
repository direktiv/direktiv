import { LogListSchema, LogListSchemaType } from "../schema";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { logKeys } from "../";
import moment from "moment";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

const getLogs = apiFactory({
  url: ({ namespace, instanceId }: { namespace: string; instanceId: string }) =>
    `/api/namespaces/${namespace}/instances/${instanceId}/logs`,
  method: "GET",
  schema: LogListSchema,
});

const fetchLogs = async ({
  queryKey: [{ apiKey, instanceId, namespace }],
}: QueryFunctionContext<ReturnType<(typeof logKeys)["detail"]>>) =>
  getLogs({
    apiKey,
    urlParams: {
      namespace,
      instanceId,
    },
  });

export const useLogs = (
  { instanceId }: { instanceId: string },
  { streaming }: { streaming?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  useStreaming({
    url: `/api/namespaces/${namespace}/instances/${instanceId}/logs`,
    enabled: !!streaming,
    schema: LogListSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<LogListSchemaType>(
        logKeys.detail(namespace, {
          apiKey: apiKey ?? undefined,
          instanceId,
        }),
        (old) => {
          if (!old) {
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
          const lastCachedLog = old.results[old.results.length - 1];
          const newestIncomingLog = msg.results[0];
          let newResults: typeof old.results = [];
          if (lastCachedLog && newestIncomingLog) {
            const newestLogTimeFromCache = moment(lastCachedLog.t);
            // new results are all logs that are newer than the last cached log
            newResults = msg.results.filter((entry) =>
              newestLogTimeFromCache.isBefore(entry.t)
            );
          }
          return {
            ...old,
            results: [...old.results, ...newResults],
          };
        }
      );
    },
  });

  return useQuery({
    queryKey: logKeys.detail(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchLogs,
    enabled: !streaming,
  });
};

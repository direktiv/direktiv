import { LogListSchema, LogListSchemaType } from "../schema";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { logKeys } from "../";
import moment from "moment";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

export type FiltersObj = {
  QUERY?: { type: "MATCH"; workflowName?: string; stateName?: string };
};

export const getFilterQuery = (filters: FiltersObj) => {
  let query = "";
  const filterFields = Object.keys(filters) as Array<keyof FiltersObj>;

  filterFields.forEach((field) => {
    const filterItem = filters[field];
    // without the guard, TS thinks filterItem may be undefined
    if (!filterItem) {
      return console.error("filterItem is not defined");
    }

    if (field === "QUERY") {
      const workflowName = filterItem?.workflowName ?? "";
      const stateName = filterItem?.stateName ?? "";
      query = query.concat(
        `&filter.field=${field}&filter.type=${filterItem.type}&filter.val=${workflowName}::${stateName}::`
      );
    }
  });

  return query;
};

const getUrl = ({
  namespace,
  baseUrl,
  instanceId,
  filters,
}: {
  baseUrl?: string;
  namespace: string;
  instanceId: string;
  filters?: FiltersObj;
}) => {
  let url = `${
    baseUrl ?? ""
  }/api/namespaces/${namespace}/instances/${instanceId}/logs`;

  if (filters) {
    url = url.concat(`?${filters}`);
  }

  return url;
};

const getLogs = apiFactory({
  url: getUrl,
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

export const useLogsStream = ({
  instanceId,
  enabled = true,
  filters,
}: {
  instanceId: string;
  enabled: boolean;
  filters: FiltersObj;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: `/api/namespaces/${namespace}/instances/${instanceId}/logs`,
    enabled,
    schema: LogListSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<LogListSchemaType>(
        logKeys.detail(namespace, {
          apiKey: apiKey ?? undefined,
          instanceId,
          filters,
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
};

export const useLogs = ({
  instanceId,
  filters,
}: {
  instanceId: string;
  filters: FiltersObj;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: logKeys.detail(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
      filters,
    }),
    queryFn: fetchLogs,
    enabled: !!namespace,
  });
};

import { LogListSchema, LogListSchemaType } from "../schema";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { logKeys } from "../";
import moment from "moment";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

const updateCache = (
  oldData: LogListSchemaType | undefined,
  msg: LogListSchemaType
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

export type FiltersObj = {
  workflowName?: string;
  stateName?: string;
};

export const getFilterQuery = (filters: FiltersObj = {}) => {
  let query = "";
  const workflowName = filters?.workflowName ?? "";
  const stateName = filters?.stateName ?? "";
  if (workflowName || stateName) {
    query = query.concat(
      `&filter.field=QUERY&filter.type=MATCH&filter.val=${workflowName}::${stateName}::`
    );
  }

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
    url = url.concat(`?${getFilterQuery(filters)}`);
  }

  return url;
};

const getLogs = apiFactory({
  url: getUrl,
  method: "GET",
  schema: LogListSchema,
});

const fetchLogs = async ({
  queryKey: [{ apiKey, instanceId, namespace, filters }],
}: QueryFunctionContext<ReturnType<(typeof logKeys)["detail"]>>) =>
  getLogs({
    apiKey,
    urlParams: {
      namespace,
      instanceId,
      filters,
    },
  });

export const useLogsStream = (
  {
    instanceId,
    filters,
  }: {
    instanceId: string;
    filters?: FiltersObj;
  },
  { enabled = true }: { enabled?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: getUrl({ namespace, instanceId, filters }),
    apiKey: apiKey ?? undefined,
    enabled,
    schema: LogListSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<LogListSchemaType>(
        logKeys.detail(namespace, {
          apiKey: apiKey ?? undefined,
          instanceId,
          filters: filters ?? {},
        }),
        (oldData) => updateCache(oldData, msg)
      );
    },
  });
};

export const useLogs = ({
  instanceId,
  filters,
}: {
  instanceId: string;
  filters?: FiltersObj;
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
      filters: filters ?? {},
    }),
    queryFn: fetchLogs,
    enabled: !!namespace,
  });
};

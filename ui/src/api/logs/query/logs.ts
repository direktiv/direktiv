import { QueryFunctionContext, useQuery } from "@tanstack/react-query";

import { LogsSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { buildSearchParamsString } from "~/api/utils";
import { logKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";

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

const getLogs = apiFactory({
  url: ({
    baseUrl,
    namespace,
    instance,
    route,
    activity,
    before,
    trace,
  }: LogsParams) => {
    const queryParamsString = buildSearchParamsString({
      instance,
      route,
      activity,
      before,
      trace,
    });

    return new URL(
      `/api/v2/namespaces/${namespace}/logs${queryParamsString}`,
      baseUrl
    ).toString();
  },
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

// export const useLogsStream = (
//   {
//     instanceId,
//     filters,
//   }: {
//     instanceId: string;
//     filters?: FiltersObj;
//   },
//   { enabled = true }: { enabled?: boolean } = {}
// ) => {
//   const apiKey = useApiKey();
//   const namespace = useNamespace();
//   const queryClient = useQueryClient();

//   if (!namespace) {
//     throw new Error("namespace is undefined");
//   }

//   return useStreaming({
//     url: getUrl({ namespace, instanceId, filters }),
//     apiKey: apiKey ?? undefined,
//     enabled,
//     schema: LogListSchema,
//     onMessage: (msg) => {
//       queryClient.setQueryData<LogListSchemaType>(
//         logKeys.detail(namespace, {
//           apiKey: apiKey ?? undefined,
//           instanceId,
//           filters: filters ?? {},
//         }),
//         (oldData) => updateCache(oldData, msg)
//       );
//     },
//   });
// };

// type LogStreamingSubscriberType = {
//   instanceId: string;
//   filters?: FiltersObj;
//   enabled?: boolean;
// };

// export const LogStreamingSubscriber = memo(
//   ({ instanceId, filters, enabled }: LogStreamingSubscriberType) => {
//     useLogsStream({ instanceId, filters }, { enabled: enabled ?? true });
//     return null;
//   }
// );

// LogStreamingSubscriber.displayName = "LogStreamingSubscriber";

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

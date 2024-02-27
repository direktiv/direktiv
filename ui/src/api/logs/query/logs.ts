import { LogsSchema } from "../schema";
import { apiFactory } from "~/api/apiFactory";
import { buildSearchParamsString } from "~/api/utils";

type GetLogsParams = {
  baseUrl?: string;
  namespace: string;
  route?: string;
  activity?: string;
  trace?: string;
  before?: string;
};

const getLogs = apiFactory({
  url: ({
    baseUrl,
    namespace,
    route,
    activity,
    trace,
    before,
  }: GetLogsParams) => {
    const queryParamsString = buildSearchParamsString({
      route,
      activity,
      trace,
      before,
    });

    return new URL(
      `/api/v2/namespaces/${namespace}/logs${queryParamsString}`,
      baseUrl
    ).toString();
  },
  method: "GET",
  schema: LogsSchema,
});

// const fetchLogs = async ({
//   queryKey: [{ apiKey, instanceId, namespace }],
// }: QueryFunctionContext<ReturnType<(typeof logKeys)["detail"]>>) =>
//   getLogs({
//     apiKey,
//     urlParams: {
//       namespace,
//       instanceId,
//       filters,
//     },
//   });

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

// export const useLogs = ({
//   instanceId,
//   filters,
// }: {
//   instanceId: string;
//   filters?: FiltersObj;
// }) => {
//   const apiKey = useApiKey();
//   const namespace = useNamespace();

//   if (!namespace) {
//     throw new Error("namespace is undefined");
//   }

//   return useQuery({
//     queryKey: logKeys.detail(namespace, {
//       apiKey: apiKey ?? undefined,
//       instanceId,
//       filters: filters ?? {},
//     }),
//     queryFn: fetchLogs,
//     enabled: !!namespace,
//   });
// };

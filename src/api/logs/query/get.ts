import { LogListSchema, LogListSchemaType } from "../schema";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { logKeys } from "../";
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

  const streamingUrl = `/api/namespaces/${namespace}/instances/${instanceId}/logs`;

  useStreaming({
    url: streamingUrl,
    enabled: streaming,
    onMessage: (msg) => {
      if (!msg.data) return null;

      let msgJson = null;
      try {
        // try to parse the response as json
        msgJson = JSON.parse(msg.data);
      } catch (e) {
        console.error(
          `error parsing streaming result (${msg.data}) from ${streamingUrl}}`
        );
        return;
      }

      const parsedResult = LogListSchema.safeParse(msgJson);
      if (parsedResult.success === false) {
        console.error(`error parsing streaming result for ${streamingUrl}`);
        return;
      }

      queryClient.setQueryData<LogListSchemaType>(
        logKeys.detail(namespace, {
          apiKey: apiKey ?? undefined,
          instanceId,
        }),
        () => parsedResult.data
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

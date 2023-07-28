import { LogListSchema } from "../schema";
import type { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { logKeys } from "../";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useQuery } from "@tanstack/react-query";

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

export const useLogs = ({ instanceId }: { instanceId: string }) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery({
    queryKey: logKeys.detail(namespace, {
      apiKey: apiKey ?? undefined,
      instanceId,
    }),
    queryFn: fetchLogs,
    enabled: !!namespace,
  });
};

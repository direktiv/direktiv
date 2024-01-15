import { ConsumersListSchema } from "../schema";
import { QueryFunctionContext } from "@tanstack/react-query";
import { apiFactory } from "~/api/apiFactory";
import { gatewayKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import useQueryWithPermissions from "~/api/useQueryWithPermissions";

export const getConsumers = apiFactory({
  url: ({ baseUrl, namespace }: { baseUrl?: string; namespace: string }) =>
    `${baseUrl ?? ""}/api/v2/namespaces/${namespace}/gateway/consumers`,
  method: "GET",
  schema: ConsumersListSchema,
});

const fetchConsumers = async ({
  queryKey: [{ apiKey, namespace }],
}: QueryFunctionContext<ReturnType<(typeof gatewayKeys)["consumers"]>>) =>
  getConsumers({
    apiKey,
    urlParams: { namespace },
  });

export const useConsumers = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQueryWithPermissions({
    queryKey: gatewayKeys.consumers(namespace, {
      apiKey: apiKey ?? undefined,
    }),
    queryFn: fetchConsumers,
    enabled: !!namespace,
  });
};
